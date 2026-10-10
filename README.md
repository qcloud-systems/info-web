# QCS — Information Hub

![Status](https://img.shields.io/badge/status-active-brightgreen)
![License](https://img.shields.io/badge/license-MIT-blue)
![Version](https://img.shields.io/badge/version-3.1.0-purple)

Source for **info.qcloud.systems**, a public reference hub for QCS projects.
A Go program renders static HTML into `./public`, which GitHub Actions
publishes to GitHub Pages.

**Everything this site serves is public by design.** Agreement templates live on [qcloud.systems](https://qcloud.systems/legal/),
and executed client documents are not published anywhere.

---

## Quick start

```bash
make dev      # live server on http://localhost:8080
make build    # render the site into ./public
make check    # validate content without writing anything
make fmt vet  # gofmt + go vet
```

Without `make`: `go run . serve`, `go run . build`, `go run . check`.

To run this alongside qcloud.systems locally, give one of them a different
port — both default to `:8080`:

```bash
go run . serve -addr :8081
```

After the first deploy of the Go rewrite, work through the post-deploy
checklist in the `qcs-web` repo (`POST-DEPLOY.md`); three of its items apply
to this repo. `ARCHITECTURE.md` in that repo explains the build strategy
shared by both sites and how to extend it.

### Adding a status provider

Add an entry to `statusServices()` in `content.go`. Set `Kind` to
`kindStatuspage` or `kindInstatus` — Statuspage entries need both `/api/v2/`
URLs, Instatus entries must leave `Incidents` empty, and `go run . check`
enforces both. A provider that serves no CORS-enabled JSON cannot be added
without a server-side proxy.

The daily bars are built from each provider's reported incidents, which is
the only history Statuspage exposes. They track disclosure, not measured
uptime.

The only dependency is Go. There is no Node toolchain.

---

## Layout

```
main.go                  serve | build | check, plus the content validator
content.go               site metadata, nav, projects, roadmap, FAQ, pages
notes.go                 notes/*.md -> /notes/ pages: front matter, tags, TOC,
                         git dates, related notes, search data, secret scanner
notes_test.go            unit tests (go test ./... / make test)
notes/                   Markdown notes, one file per page
extras.go                build-time extras: What's New, Medium feed,
                         sitemap.xml, robots.txt, .well-known/security.txt
CHANGELOG.MD             latest entry feeds What's New; full log linked on GitHub
templates/
  base.html              shared chrome: sidebar, nav, meta tags, footer
  pages/index.html       the hub page
  pages/security.html    vulnerability disclosure policy
  pages/accessibility.html  accessibility statement
  pages/404.html         served by GitHub Pages for any missing path
  pages/notes.html       the Engineering Notes index (search + tag filter)
  pages/note.html        a single note (byline, TOC, related, author box)
  pages/tag.html         one page per tag, e.g. /notes/tags/aws.html
  partials/              shared snippets, loaded with every page
static/js/notes.js       copy buttons, heading links, search and filtering
static/og-image.png      1200x630 link-preview image
static/                  copied verbatim into ./public, "static/" stripped
  css/site.css           site styles
  css/doc.css            shared styles for the long documents
  privacy-policy.html    hand-written, kept as authored
  terms-of-use.html      hand-written, kept as authored
  CONTRIBUTING.html      hand-written, kept as authored
  CNAME                  custom domain — must survive every deploy
  favicon.ico
public/                  BUILD OUTPUT — wiped on every build, gitignored
```

`build` also writes `.nojekyll` so Pages serves the output as-is.

### Everyday content edits

| To change…            | Edit                                         |
| --------------------- | -------------------------------------------- |
| What's New, changelog | Add a `## [x.y.z] — date` entry to `CHANGELOG.MD` |
| Roadmap               | `roadmap()` in `content.go`                  |
| FAQ                   | `faqs()` in `content.go`                     |
| Prototypes            | `projects()` in `content.go`                 |
| Articles              | Publish on Medium; the next deploy picks it up |
| Analytics             | Set `CloudflareAnalytics` in `content.go`    |
| Engineering Notes     | Add `notes/your-topic.md` starting with `# Title` |

Notes are published on the next deploy. The file name becomes the URL, the
first `# ` heading the title, and the first paragraph the summary on the
index. An optional front matter block adds metadata:

```markdown
---
description: One sentence for search results and link previews.
tags: [Hugo, AWS, Static Sites]
---
# How to Host a Hugo Site on Amazon S3
```

Published and updated dates come from git history automatically (the deploy
checks out full history for this); `published:` / `updated:` override them.
Each tag gets its own page, notes sharing tags are cross-linked, and every
note is listed in `sitemap.xml` and the RSS feed at `/notes/feed.xml`.

Set `Name` in `author()` in `content.go` to show a byline and author box on
every note and to credit the author in search engines' structured data. Every note is scanned for things that look like credentials (cloud
keys, tokens, private keys, `password = …`, passwords in URLs); a match fails
the build so it never reaches the public site. Raw HTML in a note is escaped.
In `make dev`, edits to a note show on refresh; a new note needs a restart.

Articles are fetched from Medium at build time, so a new post appears on the
next deploy. If Medium is unreachable the build still succeeds and the
section falls back to a link.

`.well-known/security.txt` gets a fresh `Expires` date (one year out) on
every build. Deploy at least once a year so it never expires.

---

## Design

- **Palette**: charcoal ink (`#1a2332`), action blue (`#0066cc`), sky-blue
  accent (`#6cc5e6`) from the qcloud systems wordmark, on grey/white surfaces
- **Typography**: Inter
- **Layout**: fixed sidebar on desktop, stacked on narrow screens
- Matches qcloud.systems so the two sites read as one family

---

## Adding a page

1. Create `templates/pages/yourpage.html` wrapped in `{{define "content"}} … {{end}}`.
2. Add a `Page` entry to `pages()` in `content.go`.

`go run . check` fails on a missing template, a missing document, a project
without a URL, a missing `CNAME`, an unknown roadmap stage, an incomplete FAQ
entry, a changelog without a version heading, or a malformed analytics token.

---

## Deployment

Pushing to `main` triggers `.github/workflows/deploy.yml`: `go vet`,
`go run . build`, then publish `./public` via `peaceiris/actions-gh-pages`.
SSL is managed by GitHub Pages.

No secrets are required beyond the automatic `GITHUB_TOKEN`.

---

## DNS

`static/CNAME` tells Pages which custom domain to serve, and the build copies
it into the output on every deploy — the validator fails the build if it goes
missing. At the registrar, `info` is a CNAME record pointing at
`qcloud-systems.github.io`.

---

## Contributing

See [CONTRIBUTING.html](https://info.qcloud.systems/CONTRIBUTING.html).
Fork, branch, test locally with `make dev`, open a pull request.

---

## Version history

- **3.1.0** — Engineering Notes, roadmap, FAQ, articles, policy pages, SEO
- **3.0.0** — Go static site generator; agreements moved to qcloud.systems
- **2.0.0** — Grey/blue redesign, GitHub Actions deployment
- **1.0.0** — Initial documentation hub

---

## Links

- **Site**: https://info.qcloud.systems
- **Main site**: https://qcloud.systems
- **GitHub**: https://github.com/qcloud-systems

---

## License

MIT. See the LICENSE file.
