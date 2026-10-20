package main

// Engineering Notes: Markdown files in notes/ published under /notes/.
//
// Drop a .md file into notes/ and it is published on the next build -- no
// other edit needed. The file name becomes the URL (notes/hugo-static-site.md
// -> /notes/hugo-static-site.html), the first "# " heading becomes the title,
// and the first paragraph becomes the summary on the index.
//
// A note may start with an optional front matter block:
//
//	---
//	description: One sentence for search results and link previews.
//	tags: [Hugo, AWS, Static Sites]
//	published: 2026-10-10
//	updated: 2026-10-12
//	---
//
// Dates default to the note's first and latest commit in git history, so
// they normally need no maintenance; set them only to override.
//
// This repository and site are public, so every note is scanned for things
// that look like credentials before it can ship; see secretPatterns.

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

//go:embed notes
var notesFS embed.FS

// Note is one rendered Markdown note.
type Note struct {
	Slug        string
	Title       string
	Summary     string
	Description string
	Tags        []Tag
	Published   time.Time
	Updated     time.Time
	Body        template.HTML
	TOC         []TOCEntry
	Related     []NoteLink
	// Text is the note as plain text, for the search box.
	Text string
}

// Tag is a topic label. Name is shown; Slug is used in URLs and filtering.
type Tag struct {
	Name string
	Slug string
}

// TagGroup is a tag and the notes carrying it, for tag pages and the index.
type TagGroup struct {
	Tag   Tag
	Notes []Note
}

// TOCEntry is one heading in a note's table of contents.
type TOCEntry struct {
	Level int
	ID    string
	Text  string
}

// NoteLink is a lightweight reference to another note.
type NoteLink struct {
	Slug  string
	Title string
}

// minTOC is the fewest headings worth a table of contents.
const minTOC = 3

// maxRelated caps the "Related notes" list.
const maxRelated = 3

// searchTextLimit caps the plain text per note shipped to the search box.
const searchTextLimit = 6000

// noteMD renders notes with GitHub-flavoured Markdown (tables, task lists,
// strikethrough, autolinks) and gives each heading an id so sections can be
// linked. Raw HTML in a note is escaped, not rendered.
var noteMD = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
)

var noteSlug = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// secretPatterns are checked against every note. A match fails the build:
// once pushed, the text is public in git history even if removed later, so
// the only safe place to catch it is before it is committed.
var secretPatterns = map[string]*regexp.Regexp{
	"AWS access key":       regexp.MustCompile(`\b(AKIA|ASIA)[0-9A-Z]{16}\b`),
	"private key":          regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`),
	"GitHub token":         regexp.MustCompile(`\b(ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9]{36}\b|\bgithub_pat_[A-Za-z0-9_]{22,}\b`),
	"Slack token":          regexp.MustCompile(`\bxox[abprs]-[A-Za-z0-9-]{10,}\b`),
	"Stripe secret key":    regexp.MustCompile(`\b(sk|rk)_live_[A-Za-z0-9]{16,}\b`),
	"Google API key":       regexp.MustCompile(`\bAIza[0-9A-Za-z_-]{35}\b`),
	"JWT":                  regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{10,}\.eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}`),
	"password assignment":  regexp.MustCompile(`(?i)\b(password|passwd|pwd|secret|api[_-]?key|token)\s*[:=]\s*["']?[^\s"'<>]{6,}`),
	"credentials in a URL": regexp.MustCompile(`[a-z][a-z0-9+.-]*://[^/\s:@]+:[^/\s@]+@`),
}

// notesSource is notes/ on disk in live dev mode, otherwise the copy
// embedded at compile time.
func notesSource(fromDisk bool) fs.FS {
	if fromDisk {
		return os.DirFS(".")
	}
	return notesFS
}

var (
	embeddedOnce  sync.Once
	embeddedNotes []Note
	embeddedErr   error
)

// loadNotes reads, checks and renders every note, newest first. The embedded
// set never changes while the program runs, so it is loaded once; in live dev
// mode notes are re-read on every call so edits show on refresh.
func loadNotes(fromDisk bool) ([]Note, error) {
	if fromDisk {
		return readNotes(notesSource(true))
	}
	embeddedOnce.Do(func() { embeddedNotes, embeddedErr = readNotes(notesFS) })
	return embeddedNotes, embeddedErr
}

func readNotes(src fs.FS) ([]Note, error) {
	files, err := fs.Glob(src, "notes/*.md")
	if err != nil {
		return nil, err
	}

	var notes []Note
	for _, f := range files {
		n, err := readNote(src, f)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}

	// Newest first; undated notes (not yet committed) sort by title.
	sort.SliceStable(notes, func(i, j int) bool {
		a, b := notes[i].Updated, notes[j].Updated
		if !a.Equal(b) {
			return a.After(b)
		}
		return notes[i].Title < notes[j].Title
	})
	linkRelated(notes)
	return notes, nil
}

func readNote(src fs.FS, f string) (Note, error) {
	slug := strings.TrimSuffix(path.Base(f), ".md")
	if !noteSlug.MatchString(slug) {
		return Note{}, fmt.Errorf("%s: file name must be lowercase words joined by hyphens, like docker-setup.md", f)
	}
	raw, err := fs.ReadFile(src, f)
	if err != nil {
		return Note{}, err
	}
	for name, re := range secretPatterns {
		if loc := re.FindIndex(raw); loc != nil {
			line := bytes.Count(raw[:loc[0]], []byte("\n")) + 1
			return Note{}, fmt.Errorf("%s:%d looks like it contains a secret (%s); remove it before publishing "+
				"(this site and repo are public)", f, line, name)
		}
	}

	fm, body, err := splitFrontMatter(raw)
	if err != nil {
		return Note{}, fmt.Errorf("%s: %w", f, err)
	}

	n := Note{Slug: slug}
	n.Title, n.Summary = noteHeader(body)
	if n.Title == "" {
		return Note{}, fmt.Errorf("%s: needs a \"# Title\" heading as its first line (after any front matter)", f)
	}
	n.Description = fm.Description
	if n.Description == "" {
		n.Description = n.Summary
	}
	if n.Description == "" {
		n.Description = n.Title
	}
	n.Tags = fm.Tags

	n.Published, n.Updated = gitDates(f)
	if !fm.Published.IsZero() {
		n.Published = fm.Published
	}
	if !fm.Updated.IsZero() {
		n.Updated = fm.Updated
	}
	if n.Updated.IsZero() {
		n.Updated = n.Published
	}

	doc := noteMD.Parser().Parse(text.NewReader(body))
	n.TOC = tableOfContents(doc, body)
	n.Text = searchText(doc, body)
	var buf bytes.Buffer
	if err := noteMD.Renderer().Render(&buf, body, doc); err != nil {
		return Note{}, fmt.Errorf("%s: %w", f, err)
	}
	n.Body = template.HTML(buf.String())
	return n, nil
}

// frontMatter is the optional metadata block at the top of a note.
type frontMatter struct {
	Description string
	Tags        []Tag
	Published   time.Time
	Updated     time.Time
}

// splitFrontMatter separates a leading "---" block from the Markdown body.
// Only the four documented keys are accepted, so a typo fails loudly instead
// of being silently ignored.
func splitFrontMatter(raw []byte) (frontMatter, []byte, error) {
	var fm frontMatter
	s := strings.ReplaceAll(string(raw), "\r\n", "\n")
	if !strings.HasPrefix(s, "---\n") {
		return fm, []byte(s), nil
	}
	end := strings.Index(s[4:], "\n---\n")
	if end < 0 {
		return fm, nil, fmt.Errorf("front matter starts with --- but has no closing --- line")
	}
	block, body := s[4:4+end], s[4+end+5:]

	seen := map[string]bool{}
	for i, line := range strings.Split(block, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			return fm, nil, fmt.Errorf("front matter line %d: want \"key: value\", got %q", i+1, line)
		}
		key, val = strings.TrimSpace(key), strings.TrimSpace(val)
		if seen[key] {
			return fm, nil, fmt.Errorf("front matter: %q is set twice", key)
		}
		seen[key] = true

		switch key {
		case "description":
			fm.Description = strings.Trim(val, `"'`)
		case "tags":
			tags, err := parseTags(val)
			if err != nil {
				return fm, nil, err
			}
			fm.Tags = tags
		case "published", "updated":
			t, err := time.Parse("2006-01-02", val)
			if err != nil {
				return fm, nil, fmt.Errorf("front matter: %s must be a date like 2026-10-10, got %q", key, val)
			}
			if key == "published" {
				fm.Published = t
			} else {
				fm.Updated = t
			}
		default:
			return fm, nil, fmt.Errorf("front matter: unknown key %q (want description, tags, published or updated)", key)
		}
	}
	return fm, []byte(body), nil
}

// parseTags accepts "[A, B]" or "A, B".
func parseTags(val string) ([]Tag, error) {
	val = strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(val), "["), "]")
	var tags []Tag
	seen := map[string]bool{}
	for _, name := range strings.Split(val, ",") {
		name = strings.Trim(strings.TrimSpace(name), `"'`)
		if name == "" {
			continue
		}
		slug := tagSlug(name)
		if !noteSlug.MatchString(slug) {
			return nil, fmt.Errorf("front matter: tag %q must use letters, numbers and spaces only", name)
		}
		if !seen[slug] {
			seen[slug] = true
			tags = append(tags, Tag{Name: name, Slug: slug})
		}
	}
	return tags, nil
}

func tagSlug(name string) string {
	return strings.Join(strings.Fields(strings.ToLower(name)), "-")
}

// gitDates returns the first and latest commit times touching file. Outside a
// git checkout, or for a note not yet committed, both are zero.
//
// CI must check out full history (fetch-depth: 0) or every note would show
// the date of the single commit a shallow clone contains.
func gitDates(file string) (first, last time.Time) {
	out, err := exec.Command("git", "log", "--follow", "--format=%cI", "--", file).Output()
	if err != nil {
		return
	}
	lines := strings.Fields(string(out))
	if len(lines) == 0 {
		return
	}
	last, _ = time.Parse(time.RFC3339, lines[0])
	first, _ = time.Parse(time.RFC3339, lines[len(lines)-1])
	return first, last
}

// noteHeader returns the "# " title on the first non-blank line and the first
// paragraph after it, with Markdown emphasis and code marks stripped.
func noteHeader(b []byte) (title, summary string) {
	lines := strings.Split(string(b), "\n")
	i := 0
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	if i == len(lines) || !strings.HasPrefix(lines[i], "# ") {
		return "", ""
	}
	title = strings.TrimSpace(strings.TrimPrefix(lines[i], "# "))

	var para []string
	for _, l := range lines[i+1:] {
		l = strings.TrimSpace(l)
		if l == "" {
			if len(para) > 0 {
				break
			}
			continue
		}
		if strings.HasPrefix(l, "#") || strings.HasPrefix(l, "```") || strings.HasPrefix(l, "- ") || strings.HasPrefix(l, ">") {
			break
		}
		para = append(para, l)
	}
	summary = strings.Join(para, " ")
	summary = mdLink.ReplaceAllString(summary, "$1")
	summary = strings.NewReplacer("**", "", "`", "", "*", "", "_", "").Replace(summary)
	return title, summary
}

var mdLink = regexp.MustCompile(`\[([^\]]+)\]\([^)]*\)`)

// tableOfContents lists the note's level-2 and level-3 headings. Short notes
// get none; a contents box over three lines of text is clutter.
func tableOfContents(doc ast.Node, src []byte) []TOCEntry {
	var toc []TOCEntry
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		h, ok := n.(*ast.Heading)
		if !entering || !ok {
			return ast.WalkContinue, nil
		}
		if h.Level == 2 || h.Level == 3 {
			id, _ := h.AttributeString("id")
			if b, ok := id.([]byte); ok {
				toc = append(toc, TOCEntry{Level: h.Level, ID: string(b), Text: plainText(h, src)})
			}
		}
		return ast.WalkSkipChildren, nil
	})
	if len(toc) < minTOC {
		return nil
	}
	return toc
}

// plainText concatenates the text under n, including inline code.
func plainText(n ast.Node, src []byte) string {
	var b strings.Builder
	ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch t := c.(type) {
		case *ast.Text:
			b.Write(t.Segment.Value(src))
			if t.SoftLineBreak() || t.HardLineBreak() {
				b.WriteByte(' ')
			}
		case *ast.String:
			b.Write(t.Value)
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}

// searchText flattens the whole note, code blocks included, so a search for
// a command or flag finds the note that uses it.
func searchText(doc ast.Node, src []byte) string {
	var b strings.Builder
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n.Kind() {
		case ast.KindFencedCodeBlock, ast.KindCodeBlock:
			lines := n.Lines()
			for i := 0; i < lines.Len(); i++ {
				seg := lines.At(i)
				b.Write(seg.Value(src))
			}
			b.WriteByte(' ')
			return ast.WalkSkipChildren, nil
		case ast.KindParagraph, ast.KindHeading, ast.KindListItem, extast.KindTableCell:
			b.WriteString(plainText(n, src))
			b.WriteByte(' ')
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	s := strings.Join(strings.Fields(b.String()), " ")
	if len(s) > searchTextLimit {
		s = s[:searchTextLimit]
	}
	return s
}

// linkRelated fills each note's Related list with the notes sharing the most
// tags. Notes with no tags in common are never listed.
func linkRelated(notes []Note) {
	for i := range notes {
		mine := map[string]bool{}
		for _, t := range notes[i].Tags {
			mine[t.Slug] = true
		}
		type cand struct {
			link  NoteLink
			score int
		}
		var cands []cand
		for j := range notes {
			if i == j {
				continue
			}
			score := 0
			for _, t := range notes[j].Tags {
				if mine[t.Slug] {
					score++
				}
			}
			if score > 0 {
				cands = append(cands, cand{NoteLink{notes[j].Slug, notes[j].Title}, score})
			}
		}
		sort.SliceStable(cands, func(a, b int) bool { return cands[a].score > cands[b].score })
		for k := 0; k < len(cands) && k < maxRelated; k++ {
			notes[i].Related = append(notes[i].Related, cands[k].link)
		}
	}
}

// tagGroups returns every tag in use, alphabetically, with its notes. The
// first spelling seen is the one displayed.
func tagGroups(notes []Note) []TagGroup {
	bySlug := map[string]*TagGroup{}
	for _, n := range notes {
		for _, t := range n.Tags {
			g, ok := bySlug[t.Slug]
			if !ok {
				g = &TagGroup{Tag: t}
				bySlug[t.Slug] = g
			}
			g.Notes = append(g.Notes, n)
		}
	}
	groups := make([]TagGroup, 0, len(bySlug))
	for _, g := range bySlug {
		groups = append(groups, *g)
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].Tag.Slug < groups[j].Tag.Slug })
	return groups
}

// notePages returns the Engineering Notes index, one page per tag, and one
// page per note.
func notePages(notes []Note) []Page {
	ps := []Page{{
		Path:     "notes/index.html",
		Template: "notes.html",
		Title:    "Engineering Notes — QCS Information Hub",
		Desc:     "Engineering insights on infrastructure, tooling, and software development, drawn from real-world implementation.",
	}}
	for _, g := range tagGroups(notes) {
		ps = append(ps, Page{
			Path:     "notes/tags/" + g.Tag.Slug + ".html",
			Template: "tag.html",
			Title:    g.Tag.Name + " — QCS Engineering Notes",
			Desc:     "Engineering notes on " + g.Tag.Name + ": practical guidance drawn from real-world implementation.",
			Tag:      g.Tag.Slug,
		})
	}
	for _, n := range notes {
		ps = append(ps, Page{
			Path:     "notes/" + n.Slug + ".html",
			Template: "note.html",
			Title:    n.Title + " — QCS Engineering Notes",
			Desc:     n.Description,
			Note:     n.Slug,
		})
	}
	return ps
}

// searchIndex is the JSON the Notes index page embeds for its search box.
func searchIndex(notes []Note) (template.JS, error) {
	type entry struct {
		Slug    string   `json:"slug"`
		Title   string   `json:"title"`
		Summary string   `json:"summary"`
		Tags    []string `json:"tags"`
		Text    string   `json:"text"`
	}
	entries := make([]entry, 0, len(notes))
	for _, n := range notes {
		e := entry{Slug: n.Slug, Title: n.Title, Summary: n.Description, Text: n.Text, Tags: []string{}}
		for _, t := range n.Tags {
			e.Tags = append(e.Tags, t.Name)
		}
		entries = append(entries, e)
	}
	// json.Marshal escapes <, > and & so this is safe inside a <script>.
	b, err := json.Marshal(entries)
	return template.JS(b), err
}

// articleJSONLD is schema.org TechArticle structured data for one note, so
// search engines can show it as a technical article with author and dates.
func articleJSONLD(s Site, a Author, n Note) (template.JS, error) {
	type thing map[string]any
	var author thing
	if a.Name != "" {
		author = thing{"@type": "Person", "name": a.Name, "jobTitle": a.Role, "url": a.URL}
	} else {
		author = thing{"@type": "Organization", "name": s.Name, "url": s.BaseURL}
	}
	doc := thing{
		"@context":         "https://schema.org",
		"@type":            "TechArticle",
		"headline":         n.Title,
		"description":      n.Description,
		"url":              pageURL(s, "notes/"+n.Slug+".html"),
		"mainEntityOfPage": pageURL(s, "notes/"+n.Slug+".html"),
		"image":            s.BaseURL + "/og-image.png",
		"author":           author,
		"publisher":        thing{"@type": "Organization", "name": s.Name, "url": s.BaseURL},
	}
	if !n.Published.IsZero() {
		doc["datePublished"] = n.Published.Format(time.RFC3339)
	}
	if !n.Updated.IsZero() {
		doc["dateModified"] = n.Updated.Format(time.RFC3339)
	}
	if len(n.Tags) > 0 {
		var kw []string
		for _, t := range n.Tags {
			kw = append(kw, t.Name)
		}
		doc["keywords"] = strings.Join(kw, ", ")
	}
	b, err := json.Marshal(doc)
	return template.JS(b), err
}
