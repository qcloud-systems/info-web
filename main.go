// Command info-web renders and serves the info.qcloud.systems hub.
//
//	go run . serve    # dev server on :8080, re-reads templates each request
//	go run . build    # render the static site into ./public
//
// build produces the directory that GitHub Pages publishes. Everything this
// site emits is public by design; there is no gated content here, and the
// build fails if a file that looks like an agreement reappears.
package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

//go:embed templates
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

var funcs = template.FuncMap{
	"year": func() int { return time.Now().Year() },
}

type data struct {
	Site       Site
	Page       Page
	Canonical  string
	Status     []StatusService
	Roadmap    []RoadmapGroup
	FAQs       []FAQ
	Notices    []Notice
	Notice     Notice
	Categories []CategoryGroup
	Category   CategoryGroup
	Author     Author
	// Search is the Notices index search data; ArticleLD is a notice's
	// structured data. Both are JSON, empty on pages that do not use them.
	Search    template.JS
	ArticleLD template.JS
	Latest    Release
	Built     time.Time
}

// RoadmapGroup is one stage column of the roadmap, in display order.
type RoadmapGroup struct {
	Label string
	Stage string
	Items []RoadmapItem
}

// buildInfo is gathered once per build (or once per dev-server start) rather
// than per page, so every page reports the same build time.
type buildInfo struct {
	Built time.Time
}

func roadmapGroups() []RoadmapGroup {
	groups := []RoadmapGroup{
		{Label: "In progress", Stage: stageInProgress},
		{Label: "Planned", Stage: stagePlanned},
		{Label: "Shipped", Stage: stageShipped},
	}
	for _, it := range roadmap() {
		for i := range groups {
			if groups[i].Stage == it.Stage {
				groups[i].Items = append(groups[i].Items, it)
			}
		}
	}
	return groups
}

// cfToken matches a Cloudflare Web Analytics beacon token.
var cfToken = regexp.MustCompile(`^[0-9a-f]{32}$`)

// forbidden names must never ship from this repo. They are the documents that
// used to sit here behind a client-side password check that did not actually
// restrict anything: the files were served at predictable public URLs and the
// passwords were injected into page JavaScript at deploy time. Agreements are
// no longer published from this repo. This check exists so a copy cannot drift
// back in.
var forbidden = []string{
	"protected.html",
	"service-agreement.html",
	"nda-agreement.html",
	"beta-testing-agreement.html",
}

func main() {
	log.SetFlags(0)

	cmd := "serve"
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		cmd = os.Args[1]
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}

	if err := validate(); err != nil {
		log.Fatalf("content error: %v", err)
	}

	var err error
	switch cmd {
	case "serve":
		err = serve()
	case "build":
		err = build()
	case "check":
		var all []Page
		if all, err = allPages(false); err == nil {
			fmt.Printf("ok: %d page(s), %d document(s)\n", len(all), len(documents()))
		}
	case "-h", "--help", "help":
		fmt.Println("usage: info-web [serve|build|check]")
		return
	default:
		log.Fatalf("unknown command %q (want: serve, build, check)", cmd)
	}
	if err != nil {
		log.Fatal(err)
	}
}

// allPages is pages() plus the Notices index and one page per notice.
func allPages(fromDisk bool) ([]Page, error) {
	notices, err := loadNotices(fromDisk)
	if err != nil {
		return nil, err
	}
	return append(pages(), noticePages(notices)...), nil
}

func validate() error {
	all, err := allPages(false)
	if err != nil {
		return err
	}
	for _, p := range all {
		if _, err := fs.Stat(templateFS, "templates/pages/"+p.Template); err != nil {
			return fmt.Errorf("page %q references missing template %q", p.Path, p.Template)
		}
	}

	for _, d := range documents() {
		if _, err := fs.Stat(staticFS, "static/"+d); err != nil {
			return fmt.Errorf("missing document static/%s", d)
		}
	}

	// GitHub Pages needs the custom domain file to survive every deploy.
	if _, err := fs.Stat(staticFS, "static/CNAME"); err != nil {
		return errors.New("static/CNAME is missing; the custom domain would be dropped on deploy")
	}

	s := site()
	if !strings.HasPrefix(s.BaseURL, "https://") || strings.HasSuffix(s.BaseURL, "/") {
		return fmt.Errorf("site BaseURL must be https with no trailing slash, got %q", s.BaseURL)
	}
	if s.CloudflareAnalytics != "" && !cfToken.MatchString(s.CloudflareAnalytics) {
		return fmt.Errorf("CloudflareAnalytics must be a 32-character hex token, got %q", s.CloudflareAnalytics)
	}

	for _, it := range roadmap() {
		switch it.Stage {
		case stagePlanned, stageInProgress, stageShipped:
		default:
			return fmt.Errorf("roadmap item %q: unknown stage %q (want %q, %q or %q)",
				it.Title, it.Stage, stagePlanned, stageInProgress, stageShipped)
		}
	}
	for _, f := range faqs() {
		if f.Q == "" || f.A == "" {
			return fmt.Errorf("FAQ entry %q needs both a question and an answer", f.Q)
		}
	}
	if _, err := latestRelease(changelogMD); err != nil {
		return err
	}

	// Notices may only use the categories in noticeCategories(). A typo would
	// otherwise create a new category page and split the index silently.
	allowed := map[string]bool{}
	for _, c := range noticeCategories() {
		allowed[categorySlug(c)] = true
	}
	notices, err := loadNotices(false)
	if err != nil {
		return err
	}
	for _, n := range notices {
		if len(n.Categories) == 0 {
			return fmt.Errorf("notice %q needs at least one category (one of %s)",
				n.Slug, strings.Join(noticeCategories(), ", "))
		}
		for _, c := range n.Categories {
			if !allowed[c.Slug] {
				return fmt.Errorf("notice %q: unknown category %q (want one of %s)",
					n.Slug, c.Name, strings.Join(noticeCategories(), ", "))
			}
		}
	}

	// The status strip is fetched from the browser, so every endpoint must be
	// https -- a mixed-content request would be blocked -- and must match an
	// adapter the client script knows how to parse.
	seenRole := map[string]bool{}
	for _, s := range statusServices() {
		if s.Role == "" {
			return fmt.Errorf("status service %q has no Role; the page renders "+
				"Role, not Name, so an empty one shows a blank row", s.Name)
		}
		if seenRole[s.Role] {
			return fmt.Errorf("status service %q reuses Role %q; roles are the "+
				"only label shown, so duplicates are indistinguishable", s.Name, s.Role)
		}
		seenRole[s.Role] = true

		switch s.Kind {
		case kindStatuspage:
			if !strings.HasSuffix(s.API, "/api/v2/status.json") {
				return fmt.Errorf("status service %q: statuspage API must end in "+
					"/api/v2/status.json, got %q", s.Name, s.API)
			}
			if !strings.HasSuffix(s.Incidents, "/api/v2/incidents.json") {
				return fmt.Errorf("status service %q: statuspage incidents must end in "+
					"/api/v2/incidents.json, got %q", s.Name, s.Incidents)
			}
		case kindInstatus:
			if s.Incidents != "" {
				return fmt.Errorf("status service %q: instatus exposes no incident "+
					"history, so Incidents must be empty", s.Name)
			}
		default:
			return fmt.Errorf("status service %q: unknown kind %q (want %q or %q)",
				s.Name, s.Kind, kindStatuspage, kindInstatus)
		}

		for label, u := range map[string]string{"API": s.API, "Page": s.Page} {
			if !strings.HasPrefix(u, "https://") {
				return fmt.Errorf("status service %q: %s must be https, got %q", s.Name, label, u)
			}
		}
		if s.Incidents != "" && !strings.HasPrefix(s.Incidents, "https://") {
			return fmt.Errorf("status service %q: Incidents must be https, got %q", s.Name, s.Incidents)
		}
	}

	// Refuse to ship anything that belongs behind authentication.
	err = fs.WalkDir(staticFS, "static", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		for _, bad := range forbidden {
			if strings.EqualFold(path.Base(p), bad) {
				return fmt.Errorf("%s must not be published from this repo; "+
					"agreements do not belong on this site", p)
			}
		}
		return nil
	})
	return err
}

// parse loads the base layout, every shared partial in templates/partials,
// and the page's own template.
func parse(p Page, fromDisk bool) (*template.Template, error) {
	var src fs.FS = templateFS
	if fromDisk {
		src = os.DirFS(".")
	}
	partials, err := fs.Glob(src, "templates/partials/*.html")
	if err != nil {
		return nil, err
	}
	files := append([]string{"templates/base.html"}, partials...)
	files = append(files, "templates/pages/"+p.Template)
	return template.New("base.html").Funcs(funcs).ParseFS(src, files...)
}

func render(w io.Writer, p Page, fromDisk bool, info buildInfo) error {
	tpl, err := parse(p, fromDisk)
	if err != nil {
		return fmt.Errorf("parse %s: %w", p.Template, err)
	}
	md, err := changelogSource(fromDisk)
	if err != nil {
		return err
	}
	latest, err := latestRelease(md)
	if err != nil {
		return err
	}
	notices, err := loadNotices(fromDisk)
	if err != nil {
		return err
	}
	s := site()
	a := author()
	categories := categoryGroups(notices)

	var notice Notice
	for _, n := range notices {
		if n.Slug == p.Notice {
			notice = n
		}
	}
	if p.Notice != "" && notice.Slug == "" {
		return fmt.Errorf("notice %q no longer exists; restart the dev server", p.Notice)
	}
	var category CategoryGroup
	for _, g := range categories {
		if g.Category.Slug == p.Category {
			category = g
		}
	}
	if p.Category != "" && category.Category.Slug == "" {
		return fmt.Errorf("category %q is no longer used; restart the dev server", p.Category)
	}

	var search, articleLD template.JS
	if p.Template == "notices.html" {
		if search, err = searchIndex(notices); err != nil {
			return err
		}
	}
	if notice.Slug != "" {
		if articleLD, err = articleJSONLD(s, a, notice); err != nil {
			return err
		}
	}
	return tpl.ExecuteTemplate(w, "base.html", data{
		Site:       s,
		Page:       p,
		Canonical:  pageURL(s, p.Path),
		Status:     statusServices(),
		Roadmap:    roadmapGroups(),
		FAQs:       faqs(),
		Notices:    notices,
		Notice:     notice,
		Categories: categories,
		Category:   category,
		Author:     a,
		Search:     search,
		ArticleLD:  articleLD,
		Latest:     latest,
		Built:      info.Built,
	})
}

func serve() error {
	addr := flag.String("addr", ":8080", "listen address")
	live := flag.Bool("live", true, "read templates and static files from disk (dev)")
	flag.Parse()

	// The page list is read once at startup, so a newly added notice needs a
	// restart; edits to an existing notice show on refresh.
	all, err := allPages(*live)
	if err != nil {
		return err
	}
	byPath := map[string]Page{}
	for _, p := range all {
		byPath["/"+p.Path] = p
		if path.Base(p.Path) == "index.html" {
			byPath["/"+strings.TrimSuffix(p.Path, "index.html")] = p
		}
	}

	var assets http.Handler
	if *live {
		assets = http.FileServer(http.Dir("static"))
	} else {
		sub, err := fs.Sub(staticFS, "static")
		if err != nil {
			return err
		}
		assets = http.FileServer(http.FS(sub))
	}

	info := buildInfo{Built: time.Now().UTC()}
	notices, err := loadNotices(*live)
	if err != nil {
		return err
	}
	extra := generated(site(), all, notices, info.Built)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if b, ok := extra[strings.TrimPrefix(r.URL.Path, "/")]; ok {
			ct := "text/plain; charset=utf-8"
			if strings.HasSuffix(r.URL.Path, ".xml") {
				ct = "application/xml; charset=utf-8"
			}
			w.Header().Set("Content-Type", ct)
			w.Write(b)
			return
		}
		if p, ok := byPath[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if err := render(w, p, *live, info); err != nil {
				log.Printf("render %s: %v", p.Path, err)
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
			return
		}
		assets.ServeHTTP(w, r)
	})

	srv := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("info-web listening on http://localhost%s  (live=%v)", *addr, *live)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func build() error {
	out := flag.String("out", "public", "output directory")
	flag.Parse()

	if err := os.RemoveAll(*out); err != nil {
		return fmt.Errorf("clean %s: %w", *out, err)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}

	info := buildInfo{Built: time.Now().UTC()}

	all, err := allPages(false)
	if err != nil {
		return err
	}
	for _, p := range all {
		dst := filepath.Join(*out, filepath.FromSlash(p.Path))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		f, err := os.Create(dst)
		if err != nil {
			return err
		}
		err = render(f, p, false, info)
		closeErr := f.Close()
		if err != nil {
			return fmt.Errorf("render %s: %w", p.Path, err)
		}
		if closeErr != nil {
			return closeErr
		}
	}

	n, err := copyStatic(*out)
	if err != nil {
		return err
	}

	notices, err := loadNotices(false)
	if err != nil {
		return err
	}
	extra := generated(site(), all, notices, info.Built)
	for name, b := range extra {
		dst := filepath.Join(*out, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, b, 0o644); err != nil {
			return err
		}
	}

	// Tell GitHub Pages not to run the content through Jekyll.
	if err := os.WriteFile(filepath.Join(*out, ".nojekyll"), nil, 0o644); err != nil {
		return err
	}

	log.Printf("built %s: %d page(s) + %d static file(s) + %d generated file(s) + .nojekyll",
		*out, len(all), n, len(extra))
	return nil
}

func copyStatic(out string) (int, error) {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return 0, err
	}
	count := 0
	err = fs.WalkDir(sub, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		dst := filepath.Join(out, filepath.FromSlash(p))
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		b, err := fs.ReadFile(sub, p)
		if err != nil {
			return err
		}
		count++
		return os.WriteFile(dst, b, 0o644)
	})
	return count, err
}
