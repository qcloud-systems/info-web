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
	Site     Site
	Page     Page
	Projects []Project
	Status   []StatusService
}

// forbidden names must never ship from this repo. They are the documents that
// used to sit here behind a client-side password check that did not actually
// restrict anything: the files were served at predictable public URLs and the
// passwords were injected into page JavaScript at deploy time. Agreements now
// live on qcloud.systems. This check exists so a copy cannot drift back in.
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
		fmt.Printf("ok: %d page(s), %d document(s)\n", len(pages()), len(documents()))
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

func validate() error {
	for _, p := range pages() {
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

	for _, p := range projects() {
		if p.URL == "" {
			return fmt.Errorf("project %q has no URL", p.Name)
		}
	}

	// The status strip is fetched from the browser, so every endpoint must be
	// https -- a mixed-content request would be blocked -- and must match an
	// adapter the client script knows how to parse.
	for _, s := range statusServices() {
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
	err := fs.WalkDir(staticFS, "static", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		for _, bad := range forbidden {
			if strings.EqualFold(path.Base(p), bad) {
				return fmt.Errorf("%s must not be published from this repo; "+
					"agreements belong on qcloud.systems", p)
			}
		}
		return nil
	})
	return err
}

func parse(p Page, fromDisk bool) (*template.Template, error) {
	files := []string{"templates/base.html", "templates/pages/" + p.Template}
	if fromDisk {
		return template.New("base.html").Funcs(funcs).ParseFiles(files...)
	}
	return template.New("base.html").Funcs(funcs).ParseFS(templateFS, files...)
}

func render(w io.Writer, p Page, fromDisk bool) error {
	tpl, err := parse(p, fromDisk)
	if err != nil {
		return fmt.Errorf("parse %s: %w", p.Template, err)
	}
	return tpl.ExecuteTemplate(w, "base.html", data{
		Site:     site(),
		Page:     p,
		Projects: projects(),
		Status:   statusServices(),
	})
}

func serve() error {
	addr := flag.String("addr", ":8080", "listen address")
	live := flag.Bool("live", true, "read templates and static files from disk (dev)")
	flag.Parse()

	byPath := map[string]Page{}
	for _, p := range pages() {
		byPath["/"+p.Path] = p
		if path.Base(p.Path) == "index.html" && path.Dir(p.Path) == "." {
			byPath["/"] = p
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

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if p, ok := byPath[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if err := render(w, p, *live); err != nil {
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

	for _, p := range pages() {
		dst := filepath.Join(*out, filepath.FromSlash(p.Path))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		f, err := os.Create(dst)
		if err != nil {
			return err
		}
		err = render(f, p, false)
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

	// Tell GitHub Pages not to run the content through Jekyll.
	if err := os.WriteFile(filepath.Join(*out, ".nojekyll"), nil, 0o644); err != nil {
		return err
	}

	log.Printf("built %s: %d page(s) + %d static file(s) + .nojekyll", *out, len(pages()), n)
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
