package main

// Notices: Markdown files in notices/ published under /notices/ -- incident
// reports, maintenance windows, security advisories and service changes.
//
// Drop a .md file into notices/ and it is published on the next build -- no
// other edit needed. The file name becomes the URL (notices/api-outage.md
// -> /notices/api-outage.html), the first "# " heading becomes the title,
// and the first paragraph becomes the summary on the index.
//
// A notice may start with an optional front matter block:
//
//	---
//	description: One sentence for search results and link previews.
//	categories: [Incident, Security Advisory]
//	published: 2026-10-10
//	updated: 2026-10-12
//	---
//
// Dates default to the notice's first and latest commit in git history, so
// they normally need no maintenance; set them only to override.
//
// This repository and site are public, so every notice is scanned for things
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

//go:embed notices
var noticesFS embed.FS

// Notice is one rendered Markdown notice.
type Notice struct {
	Slug        string
	Title       string
	Summary     string
	Description string
	Categories  []Category
	Published   time.Time
	Updated     time.Time
	Body        template.HTML
	TOC         []TOCEntry
	Related     []NoticeLink
	// Text is the notice as plain text, for the search box.
	Text string
}

// Category is a topic label. Name is shown; Slug is used in URLs and filtering.
type Category struct {
	Name string
	Slug string
}

// CategoryGroup is a category and the notices carrying it, for category pages and the index.
type CategoryGroup struct {
	Category Category
	Notices  []Notice
}

// TOCEntry is one heading in a notice's table of contents.
type TOCEntry struct {
	Level int
	ID    string
	Text  string
}

// NoticeLink is a lightweight reference to another notice.
type NoticeLink struct {
	Slug  string
	Title string
}

// minTOC is the fewest headings worth a table of contents.
const minTOC = 3

// maxRelated caps the "Related notices" list.
const maxRelated = 3

// searchTextLimit caps the plain text per notice shipped to the search box.
const searchTextLimit = 6000

// noticeMD renders notices with GitHub-flavoured Markdown (tables, task lists,
// strikethrough, autolinks) and gives each heading an id so sections can be
// linked. Raw HTML in a notice is escaped, not rendered.
var noticeMD = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
)

var noticeSlug = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// secretPatterns are checked against every notice. A match fails the build:
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

// noticesSource is notices/ on disk in live dev mode, otherwise the copy
// embedded at compile time.
func noticesSource(fromDisk bool) fs.FS {
	if fromDisk {
		return os.DirFS(".")
	}
	return noticesFS
}

var (
	embeddedOnce    sync.Once
	embeddedNotices []Notice
	embeddedErr     error
)

// loadNotices reads, checks and renders every notice, newest first. The embedded
// set never changes while the program runs, so it is loaded once; in live dev
// mode notices are re-read on every call so edits show on refresh.
func loadNotices(fromDisk bool) ([]Notice, error) {
	if fromDisk {
		return readNotices(noticesSource(true))
	}
	embeddedOnce.Do(func() { embeddedNotices, embeddedErr = readNotices(noticesFS) })
	return embeddedNotices, embeddedErr
}

func readNotices(src fs.FS) ([]Notice, error) {
	files, err := fs.Glob(src, "notices/*.md")
	if err != nil {
		return nil, err
	}

	var notices []Notice
	for _, f := range files {
		n, err := readNotice(src, f)
		if err != nil {
			return nil, err
		}
		notices = append(notices, n)
	}

	// Newest first; undated notices (not yet committed) sort by title.
	sort.SliceStable(notices, func(i, j int) bool {
		a, b := notices[i].Updated, notices[j].Updated
		if !a.Equal(b) {
			return a.After(b)
		}
		return notices[i].Title < notices[j].Title
	})
	linkRelated(notices)
	return notices, nil
}

func readNotice(src fs.FS, f string) (Notice, error) {
	slug := strings.TrimSuffix(path.Base(f), ".md")
	if !noticeSlug.MatchString(slug) {
		return Notice{}, fmt.Errorf("%s: file name must be lowercase words joined by hyphens, like api-outage.md", f)
	}
	raw, err := fs.ReadFile(src, f)
	if err != nil {
		return Notice{}, err
	}
	for name, re := range secretPatterns {
		if loc := re.FindIndex(raw); loc != nil {
			line := bytes.Count(raw[:loc[0]], []byte("\n")) + 1
			return Notice{}, fmt.Errorf("%s:%d looks like it contains a secret (%s); remove it before publishing "+
				"(this site and repo are public)", f, line, name)
		}
	}

	fm, body, err := splitFrontMatter(raw)
	if err != nil {
		return Notice{}, fmt.Errorf("%s: %w", f, err)
	}

	n := Notice{Slug: slug}
	n.Title, n.Summary = noticeHeader(body)
	if n.Title == "" {
		return Notice{}, fmt.Errorf("%s: needs a \"# Title\" heading as its first line (after any front matter)", f)
	}
	n.Description = fm.Description
	if n.Description == "" {
		n.Description = n.Summary
	}
	if n.Description == "" {
		n.Description = n.Title
	}
	n.Categories = fm.Categories

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

	doc := noticeMD.Parser().Parse(text.NewReader(body))
	n.TOC = tableOfContents(doc, body)
	n.Text = searchText(doc, body)
	var buf bytes.Buffer
	if err := noticeMD.Renderer().Render(&buf, body, doc); err != nil {
		return Notice{}, fmt.Errorf("%s: %w", f, err)
	}
	n.Body = template.HTML(buf.String())
	return n, nil
}

// frontMatter is the optional metadata block at the top of a notice.
type frontMatter struct {
	Description string
	Categories  []Category
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
		case "categories":
			categories, err := parseCategories(val)
			if err != nil {
				return fm, nil, err
			}
			fm.Categories = categories
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
			return fm, nil, fmt.Errorf("front matter: unknown key %q (want description, categories, published or updated)", key)
		}
	}
	return fm, []byte(body), nil
}

// parseCategories accepts "[A, B]" or "A, B".
func parseCategories(val string) ([]Category, error) {
	val = strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(val), "["), "]")
	var categories []Category
	seen := map[string]bool{}
	for _, name := range strings.Split(val, ",") {
		name = strings.Trim(strings.TrimSpace(name), `"'`)
		if name == "" {
			continue
		}
		slug := categorySlug(name)
		if !noticeSlug.MatchString(slug) {
			return nil, fmt.Errorf("front matter: category %q must use letters, numbers and spaces only", name)
		}
		if !seen[slug] {
			seen[slug] = true
			categories = append(categories, Category{Name: name, Slug: slug})
		}
	}
	return categories, nil
}

func categorySlug(name string) string {
	return strings.Join(strings.Fields(strings.ToLower(name)), "-")
}

// gitDates returns the first and latest commit times touching file. Outside a
// git checkout, or for a notice not yet committed, both are zero.
//
// CI must check out full history (fetch-depth: 0) or every notice would show
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

// noticeHeader returns the "# " title on the first non-blank line and the first
// paragraph after it, with Markdown emphasis and code marks stripped.
func noticeHeader(b []byte) (title, summary string) {
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

// tableOfContents lists the notice's level-2 and level-3 headings. Short notices
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

// searchText flattens the whole notice, code blocks included, so a search for
// a command or flag finds the notice that uses it.
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

// linkRelated fills each notice's Related list with the notices sharing the most
// categories. Notices with no categories in common are never listed.
func linkRelated(notices []Notice) {
	for i := range notices {
		mine := map[string]bool{}
		for _, t := range notices[i].Categories {
			mine[t.Slug] = true
		}
		type cand struct {
			link  NoticeLink
			score int
		}
		var cands []cand
		for j := range notices {
			if i == j {
				continue
			}
			score := 0
			for _, t := range notices[j].Categories {
				if mine[t.Slug] {
					score++
				}
			}
			if score > 0 {
				cands = append(cands, cand{NoticeLink{notices[j].Slug, notices[j].Title}, score})
			}
		}
		sort.SliceStable(cands, func(a, b int) bool { return cands[a].score > cands[b].score })
		for k := 0; k < len(cands) && k < maxRelated; k++ {
			notices[i].Related = append(notices[i].Related, cands[k].link)
		}
	}
}

// categoryGroups returns every category in use, alphabetically, with its notices. The
// first spelling seen is the one displayed.
func categoryGroups(notices []Notice) []CategoryGroup {
	bySlug := map[string]*CategoryGroup{}
	for _, n := range notices {
		for _, t := range n.Categories {
			g, ok := bySlug[t.Slug]
			if !ok {
				g = &CategoryGroup{Category: t}
				bySlug[t.Slug] = g
			}
			g.Notices = append(g.Notices, n)
		}
	}
	groups := make([]CategoryGroup, 0, len(bySlug))
	for _, g := range bySlug {
		groups = append(groups, *g)
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].Category.Slug < groups[j].Category.Slug })
	return groups
}

// noticePages returns the Notices index, one page per category, and one
// page per notice.
func noticePages(notices []Notice) []Page {
	ps := []Page{{
		Path:     "notices/index.html",
		Template: "notices.html",
		Title:    "Notices — QCS Trust & Status Center",
		Desc:     "Incident reports, maintenance windows, security advisories, and service changes from QCS.",
	}}
	for _, g := range categoryGroups(notices) {
		ps = append(ps, Page{
			Path:     "notices/categories/" + g.Category.Slug + ".html",
			Template: "category.html",
			Title:    g.Category.Name + " — QCS Notices",
			Desc:     "QCS notices in the " + g.Category.Name + " category.",
			Category: g.Category.Slug,
		})
	}
	for _, n := range notices {
		ps = append(ps, Page{
			Path:     "notices/" + n.Slug + ".html",
			Template: "notice.html",
			Title:    n.Title + " — QCS Notices",
			Desc:     n.Description,
			Notice:   n.Slug,
		})
	}
	return ps
}

// searchIndex is the JSON the Notices index page embeds for its search box.
func searchIndex(notices []Notice) (template.JS, error) {
	type entry struct {
		Slug       string   `json:"slug"`
		Title      string   `json:"title"`
		Summary    string   `json:"summary"`
		Categories []string `json:"categories"`
		Text       string   `json:"text"`
	}
	entries := make([]entry, 0, len(notices))
	for _, n := range notices {
		e := entry{Slug: n.Slug, Title: n.Title, Summary: n.Description, Text: n.Text, Categories: []string{}}
		for _, t := range n.Categories {
			e.Categories = append(e.Categories, t.Name)
		}
		entries = append(entries, e)
	}
	// json.Marshal escapes <, > and & so this is safe inside a <script>.
	b, err := json.Marshal(entries)
	return template.JS(b), err
}

// articleJSONLD is schema.org Article structured data for one notice, so
// search engines can show it as a technical article with author and dates.
func articleJSONLD(s Site, a Author, n Notice) (template.JS, error) {
	type thing map[string]any
	var author thing
	if a.Name != "" {
		author = thing{"@type": "Person", "name": a.Name, "jobTitle": a.Role, "url": a.URL}
	} else {
		author = thing{"@type": "Organization", "name": s.Name, "url": s.BaseURL}
	}
	doc := thing{
		"@context":         "https://schema.org",
		"@type":            "Article",
		"headline":         n.Title,
		"description":      n.Description,
		"url":              pageURL(s, "notices/"+n.Slug+".html"),
		"mainEntityOfPage": pageURL(s, "notices/"+n.Slug+".html"),
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
	if len(n.Categories) > 0 {
		var kw []string
		for _, t := range n.Categories {
			kw = append(kw, t.Name)
		}
		doc["keywords"] = strings.Join(kw, ", ")
	}
	b, err := json.Marshal(doc)
	return template.JS(b), err
}
