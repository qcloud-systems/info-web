# QCS — Information Hub

![Status](https://img.shields.io/badge/status-active-brightgreen)
![License](https://img.shields.io/badge/license-MIT-blue)
![Version](https://img.shields.io/badge/version-3.0.0-purple)

Source for **info.qcloud.systems**, a public reference hub for QCS projects.
A Go program renders static HTML into `./public`, which GitHub Actions
publishes to GitHub Pages.

**Everything this site serves is public by design.** There is no gated area.
Agreement templates live on [qcloud.systems](https://qcloud.systems/legal/),
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
content.go               site metadata, nav, project list
templates/
  base.html              shared chrome: sidebar, nav, footer
  pages/index.html       the hub page
static/                  copied verbatim into ./public, "static/" stripped
  css/site.css           site styles
  css/doc.css            shared styles for the long documents
  privacy-policy.html    hand-written, kept as authored
  CONTRIBUTING.html      hand-written, kept as authored
  CNAME                  custom domain — must survive every deploy
  favicon.ico
public/                  BUILD OUTPUT — wiped on every build, gitignored
```

`build` also writes `.nojekyll` so Pages serves the output as-is.

---

## Why the protected area was removed

Earlier versions of this site had a `protected.html` page that asked for a
password before showing agreements. **It did not restrict anything**, for two
independent reasons:

1. The passwords were injected into client-side JavaScript at deploy time, so
   anyone could read them with View Source.
2. The agreement files were published as ordinary static files at predictable
   URLs. `/service-agreement.html` loaded directly, with no password involved.
   The password form only pointed an iframe at that already-public URL.

A static host has no server-side compute, so it cannot enforce access control.
Any "protection" implemented in the browser is advisory at best, and here it
was not even that.

The page and those documents have been removed, and `go run . build` now
**fails** if a file named like one of them reappears — see `forbidden` in
`main.go`. Blank templates belong on qcloud.systems; executed documents are
delivered privately and never published.

If you still have `PROTECTED_PASSWORD_SA`, `PROTECTED_PASSWORD_NDA` or
`PROTECTED_PASSWORD_BTA` configured as repository secrets, delete them. They
are unused, and they were public while the old gate was live.

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
without a URL, or a missing `CNAME`.

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

- **3.0.0** — Go static site generator; removed the non-functional password
  gate and moved agreements to qcloud.systems
- **2.0.0** — Grey/blue redesign, GitHub Actions secret injection
- **1.0.0** — Initial documentation hub

---

## Links

- **Site**: https://info.qcloud.systems
- **Main site**: https://qcloud.systems
- **GitHub**: https://github.com/qcloud-systems

---

## License

MIT. See the LICENSE file.
