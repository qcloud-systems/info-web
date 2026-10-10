package main

import (
	"strings"
	"testing"
	"time"

	"github.com/yuin/goldmark/text"
)

func TestSecretPatterns(t *testing.T) {
	leaks := map[string]string{
		"AWS access key":       "key AKIAABCDEFGHIJKLMNOP",
		"private key":          "-----BEGIN OPENSSH PRIVATE KEY-----",
		"GitHub token":         "ghp_" + strings.Repeat("a", 36),
		"password assignment":  `password = "hunter22"`,
		"credentials in a URL": "postgres://admin:s3cret@db.example.com/app",
	}
	for name, text := range leaks {
		if !secretPatterns[name].MatchString(text) {
			t.Errorf("%s: pattern did not match %q", name, text)
		}
	}

	// Ordinary technical prose must not trip the scanner.
	clean := []string{
		"Run `hugo server -D` and open http://localhost:1313/",
		"Set the password in your password manager.",
		"git submodule add https://github.com/owner/theme.git themes/theme",
	}
	for _, text := range clean {
		for name, re := range secretPatterns {
			if re.MatchString(text) {
				t.Errorf("false positive: %s matched %q", name, text)
			}
		}
	}
}

func TestNoteHeader(t *testing.T) {
	src := "\n# My Note\n\nFirst **bold** line,\nwith `code`.\n\nSecond paragraph.\n"
	title, summary := noteHeader([]byte(src))
	if title != "My Note" {
		t.Errorf("title = %q, want %q", title, "My Note")
	}
	if want := "First bold line, with code."; summary != want {
		t.Errorf("summary = %q, want %q", summary, want)
	}

	if title, _ := noteHeader([]byte("No heading here\n")); title != "" {
		t.Errorf("expected no title without a heading, got %q", title)
	}
}

func TestPageURL(t *testing.T) {
	s := Site{BaseURL: "https://example.com"}
	cases := map[string]string{
		"index.html":       "https://example.com/",
		"notes/index.html": "https://example.com/notes/",
		"security.html":    "https://example.com/security.html",
	}
	for in, want := range cases {
		if got := pageURL(s, in); got != want {
			t.Errorf("pageURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLatestRelease(t *testing.T) {
	src := "# Changelog\n\n## [2.0.0] — 2026-01-02\n### Added\n- One\n  continued\n- Two\n\n## [1.0.0] — 2025-01-01\n- Old\n"
	r, err := latestRelease([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if r.Version != "2.0.0" || r.Date != "2026-01-02" {
		t.Errorf("got %s %s, want 2.0.0 2026-01-02", r.Version, r.Date)
	}
	if len(r.Highlights) != 2 || r.Highlights[0] != "One continued" {
		t.Errorf("highlights = %q", r.Highlights)
	}
}

// TestContentValid runs the same checks as `go run . check` against the real
// content, so a bad note or roadmap entry fails `go test` too.
func TestContentValid(t *testing.T) {
	if err := validate(); err != nil {
		t.Fatal(err)
	}
}

func TestFrontMatter(t *testing.T) {
	src := "---\ndescription: \"A note.\"\ntags: [Hugo, AWS, Static Sites]\nupdated: 2026-10-12\n---\n# Title\n"
	fm, body, err := splitFrontMatter([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if fm.Description != "A note." {
		t.Errorf("description = %q", fm.Description)
	}
	if len(fm.Tags) != 3 || fm.Tags[2] != (Tag{Name: "Static Sites", Slug: "static-sites"}) {
		t.Errorf("tags = %+v", fm.Tags)
	}
	if fm.Updated.Format("2006-01-02") != "2026-10-12" {
		t.Errorf("updated = %v", fm.Updated)
	}
	if string(body) != "# Title\n" {
		t.Errorf("body = %q", body)
	}

	// No front matter: body passes through untouched.
	if _, body, err := splitFrontMatter([]byte("# Plain\n")); err != nil || string(body) != "# Plain\n" {
		t.Errorf("plain note: body %q, err %v", body, err)
	}

	bad := map[string]string{
		"typo key":     "---\ntitel: x\n---\n# T\n",
		"bad date":     "---\nupdated: Oct 12\n---\n# T\n",
		"unclosed":     "---\ntags: [a]\n# T\n",
		"bad tag":      "---\ntags: [C++]\n---\n# T\n",
		"repeated key": "---\ntags: [a]\ntags: [b]\n---\n# T\n",
	}
	for name, src := range bad {
		if _, _, err := splitFrontMatter([]byte(src)); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestTableOfContents(t *testing.T) {
	src := []byte("# Title\n\n## One\n\n### One `a`\n\n## Two\n")
	doc := noteMD.Parser().Parse(text.NewReader(src))
	toc := tableOfContents(doc, src)
	if len(toc) != 3 {
		t.Fatalf("toc = %+v", toc)
	}
	if toc[1] != (TOCEntry{Level: 3, ID: "one-a", Text: "One a"}) {
		t.Errorf("toc[1] = %+v", toc[1])
	}

	short := []byte("# Title\n\n## Only one\n")
	if toc := tableOfContents(noteMD.Parser().Parse(text.NewReader(short)), short); toc != nil {
		t.Errorf("short note should have no TOC, got %+v", toc)
	}
}

func TestLinkRelated(t *testing.T) {
	tag := func(names ...string) []Tag {
		var ts []Tag
		for _, n := range names {
			ts = append(ts, Tag{Name: n, Slug: tagSlug(n)})
		}
		return ts
	}
	notes := []Note{
		{Slug: "a", Title: "A", Tags: tag("Go", "AWS")},
		{Slug: "b", Title: "B", Tags: tag("AWS")},
		{Slug: "c", Title: "C", Tags: tag("Go", "AWS")},
		{Slug: "d", Title: "D", Tags: tag("Docker")},
	}
	linkRelated(notes)
	if got := notes[0].Related; len(got) != 2 || got[0].Slug != "c" || got[1].Slug != "b" {
		t.Errorf("related to a = %+v, want c (2 shared) then b (1 shared)", got)
	}
	if got := notes[3].Related; len(got) != 0 {
		t.Errorf("d shares no tags but got %+v", got)
	}
}

func TestNotesFeed(t *testing.T) {
	s := Site{Name: "QCS", BaseURL: "https://example.com"}
	feed := string(notesFeed(s, []Note{{Slug: "x", Title: "X & Y", Description: "d"}}, time.Now()))
	for _, want := range []string{
		"<rss version=\"2.0\"",
		"<link>https://example.com/notes/x.html</link>",
		"<title>X &amp; Y</title>",
		"rel=\"self\"",
	} {
		if !strings.Contains(feed, want) {
			t.Errorf("feed missing %q:\n%s", want, feed)
		}
	}
}
