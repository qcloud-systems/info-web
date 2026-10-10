package main

// Build-time extras: the What's New summary, the Medium articles feed, and the
// generated machine-readable files (robots.txt, sitemap.xml, security.txt).
//
// All of this runs when the site is built, never in the visitor's browser
// and never on a server -- the published output stays plain static files on
// GitHub Pages, which is what keeps hosting free.

import (
	"bytes"
	_ "embed"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/yuin/goldmark"
)

//go:embed CHANGELOG.MD
var changelogMD []byte

// changelogSource returns CHANGELOG.MD, from disk in live dev mode so edits
// show on refresh, otherwise from the copy embedded at compile time.
func changelogSource(fromDisk bool) ([]byte, error) {
	if fromDisk {
		return os.ReadFile("CHANGELOG.MD")
	}
	return changelogMD, nil
}

func markdown(src []byte) (template.HTML, error) {
	var buf bytes.Buffer
	// goldmark escapes raw HTML by default, so Markdown sources cannot
	// inject markup into the page.
	if err := goldmark.Convert(src, &buf); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

// Release is the newest changelog entry, summarised for the home page.
type Release struct {
	Version    string
	Date       string
	Highlights []template.HTML
}

var releaseHeading = regexp.MustCompile(`^## \[([^\]]+)\]\s*(?:—|-)\s*(.+)$`)

// maxHighlights caps how many bullets of the latest release the home page
// shows; the full entry is in CHANGELOG.MD on GitHub.
const maxHighlights = 5

// latestRelease reads the first "## [x.y.z] — date" section and returns its
// top-level bullets. Continuation lines are folded into their bullet.
func latestRelease(src []byte) (Release, error) {
	var r Release
	var bullets []string
	inRelease := false

	for _, line := range strings.Split(string(src), "\n") {
		if m := releaseHeading.FindStringSubmatch(line); m != nil {
			if inRelease {
				break
			}
			inRelease = true
			r.Version, r.Date = m[1], strings.TrimSpace(m[2])
			continue
		}
		if !inRelease {
			continue
		}
		switch {
		case strings.HasPrefix(line, "- "):
			bullets = append(bullets, strings.TrimPrefix(line, "- "))
		case strings.HasPrefix(line, "  ") && len(bullets) > 0 && strings.TrimSpace(line) != "":
			bullets[len(bullets)-1] += " " + strings.TrimSpace(line)
		}
	}
	if r.Version == "" {
		return r, fmt.Errorf("CHANGELOG.MD has no \"## [version] — date\" heading")
	}

	for i, b := range bullets {
		if i == maxHighlights {
			break
		}
		h, err := markdown([]byte(b))
		if err != nil {
			return r, err
		}
		// A single line renders as one <p>; unwrap it so it sits in an <li>.
		s := strings.TrimSpace(string(h))
		s = strings.TrimSuffix(strings.TrimPrefix(s, "<p>"), "</p>")
		r.Highlights = append(r.Highlights, template.HTML(s))
	}
	return r, nil
}

// Post is one Medium article.
type Post struct {
	Title string
	URL   string
	Date  time.Time
}

const maxPosts = 3

// fetchPosts reads the newest articles from the Medium RSS feed. Any failure
// returns no posts and a logged warning: the Articles section then shows a
// plain link to Medium instead.
func fetchPosts(feedURL string) []Post {
	if feedURL == "" {
		return nil
	}
	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Get(feedURL)
	if err != nil {
		log.Printf("warning: articles feed: %v (section will link to Medium)", err)
		return nil
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Printf("warning: articles feed: HTTP %d (section will link to Medium)", res.StatusCode)
		return nil
	}

	var feed struct {
		Items []struct {
			Title   string `xml:"title"`
			Link    string `xml:"link"`
			PubDate string `xml:"pubDate"`
		} `xml:"channel>item"`
	}
	if err := xml.NewDecoder(io.LimitReader(res.Body, 4<<20)).Decode(&feed); err != nil {
		log.Printf("warning: articles feed: %v (section will link to Medium)", err)
		return nil
	}

	var posts []Post
	for _, it := range feed.Items {
		if len(posts) == maxPosts {
			break
		}
		// Only link out over https; anything else in the feed is skipped.
		if it.Title == "" || !strings.HasPrefix(it.Link, "https://") {
			continue
		}
		p := Post{Title: strings.TrimSpace(it.Title), URL: it.Link}
		if t, err := time.Parse(time.RFC1123, it.PubDate); err == nil {
			p.Date = t
		} else if t, err := time.Parse(time.RFC1123Z, it.PubDate); err == nil {
			p.Date = t
		}
		posts = append(posts, p)
	}
	return posts
}

// generated returns the machine-readable files written alongside the pages,
// keyed by output path.
func generated(s Site, all []Page, notes []Note, built time.Time) map[string][]byte {
	var sitemap strings.Builder
	sitemap.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	sitemap.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	lastmod := built.Format("2006-01-02")
	for _, p := range all {
		if p.NoIndex {
			continue
		}
		fmt.Fprintf(&sitemap, "  <url><loc>%s</loc><lastmod>%s</lastmod></url>\n", pageURL(s, p.Path), lastmod)
	}
	for _, d := range documents() {
		fmt.Fprintf(&sitemap, "  <url><loc>%s</loc><lastmod>%s</lastmod></url>\n", pageURL(s, d), lastmod)
	}
	sitemap.WriteString("</urlset>\n")

	robots := "User-agent: *\nAllow: /\n\nSitemap: " + s.BaseURL + "/sitemap.xml\n"

	// RFC 9116. Expires is required and should be under a year out; it is
	// regenerated on every deploy, so redeploy at least once a year or the
	// file goes stale.
	security := strings.Join([]string{
		"Contact: " + s.Repo + "/security/advisories/new",
		"Expires: " + built.AddDate(1, 0, -1).Format(time.RFC3339),
		"Preferred-Languages: en",
		"Canonical: " + s.BaseURL + "/.well-known/security.txt",
		"Policy: " + s.BaseURL + "/security.html",
	}, "\n") + "\n"

	return map[string][]byte{
		"notes/feed.xml":           notesFeed(s, notes, built),
		"sitemap.xml":              []byte(sitemap.String()),
		"robots.txt":               []byte(robots),
		".well-known/security.txt": []byte(security),
	}
}

// pageURL is the canonical URL for an output path; a directory's index.html
// maps to the directory itself ("notes/index.html" -> "/notes/").
func pageURL(s Site, p string) string {
	return s.BaseURL + "/" + strings.TrimSuffix(p, "index.html")
}

// notesFeed is an RSS 2.0 feed of the engineering notes, newest first, for
// feed readers and developer news aggregators.
func notesFeed(s Site, notes []Note, built time.Time) []byte {
	type item struct {
		Title       string   `xml:"title"`
		Link        string   `xml:"link"`
		GUID        string   `xml:"guid"`
		PubDate     string   `xml:"pubDate,omitempty"`
		Description string   `xml:"description"`
		Categories  []string `xml:"category"`
	}
	type atomLink struct {
		Href string `xml:"href,attr"`
		Rel  string `xml:"rel,attr"`
		Type string `xml:"type,attr"`
	}
	type channel struct {
		Title       string   `xml:"title"`
		Link        string   `xml:"link"`
		Description string   `xml:"description"`
		Language    string   `xml:"language"`
		BuildDate   string   `xml:"lastBuildDate"`
		Self        atomLink `xml:"atom:link"`
		Items       []item   `xml:"item"`
	}
	type rss struct {
		XMLName xml.Name `xml:"rss"`
		Version string   `xml:"version,attr"`
		Atom    string   `xml:"xmlns:atom,attr"`
		Channel channel  `xml:"channel"`
	}

	feedURL := s.BaseURL + "/notes/feed.xml"
	ch := channel{
		Title:       s.Name + " Engineering Notes",
		Link:        pageURL(s, "notes/index.html"),
		Description: "Engineering insights on infrastructure, tooling, and software development.",
		Language:    "en",
		BuildDate:   built.Format(time.RFC1123Z),
		Self:        atomLink{Href: feedURL, Rel: "self", Type: "application/rss+xml"},
	}
	for _, n := range notes {
		u := pageURL(s, "notes/"+n.Slug+".html")
		it := item{Title: n.Title, Link: u, GUID: u, Description: n.Description}
		if !n.Updated.IsZero() {
			it.PubDate = n.Updated.Format(time.RFC1123Z)
		}
		for _, t := range n.Tags {
			it.Categories = append(it.Categories, t.Name)
		}
		ch.Items = append(ch.Items, it)
	}

	b, err := xml.MarshalIndent(rss{Version: "2.0", Atom: "http://www.w3.org/2005/Atom", Channel: ch}, "", "  ")
	if err != nil {
		// Only fixed struct types are marshalled, so this cannot happen.
		panic(err)
	}
	return append([]byte(xml.Header), append(b, '\n')...)
}
