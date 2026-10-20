# QCS — Information Hub

![Status](https://img.shields.io/badge/status-active-brightgreen)
![License](https://img.shields.io/badge/license-MIT-blue)
![Version](https://img.shields.io/badge/version-3.1.0-purple)

Source for **info.qcloud.systems**, a public reference hub for QCS projects.
A Go program renders static HTML into `./public`, which GitHub Actions
publishes to managed static hosting.

**Everything this site serves is public by design.** Agreements and executed
client documents are not published from this repo.

---

## Quick start

```bash
make dev      # live server on http://localhost:8080
make build    # render the site into ./public
make check    # validate content without writing anything
make fmt vet  # gofmt + go vet
```

Without `make`: `go run . serve`, `go run . build`, `go run . check`.

To run this alongside another local site that also uses `:8080`, give one of
them a different port:

```bash
go run . serve -addr :8081
```

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
  pages/404.html         served by the host for any missing path
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
- QCS brand palette and type

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
`go test`, `go run . build`, then publish `./public` to the static host, which
also provides HTTPS. Deployment needs no stored credentials.

---

## Custom domain

`static/CNAME` holds the site's custom domain. The build copies it into the
output on every deploy, and the validator fails the build if it goes missing,
because losing it would silently drop the domain.

---

## Working in GitHub Codespaces

### Day-to-day workflow

`main` only accepts changes through pull requests, so every change goes on a
branch:

```bash
git switch main && git pull           # start from the latest main
git switch -c YYMMDDShortName          # e.g. 261014PrivacyRepair
# ...edit, then run `make check` (and `make test` where available)...
git add -A
git commit -m "Short summary" -m "Longer explanation of what and why."
git push -u origin YYMMDDShortName
```

Then open the pull request and merge it, either on github.com (use the
**Compare & pull request** banner, then **Merge pull request**) or with
`gh pr create --fill` and `gh pr merge --merge`. Merging to `main` deploys.

Afterwards, bring your local copy up to date and delete the merged branch:

```bash
git switch main && git pull && git branch -d YYMMDDShortName
```

### Working on another repository in the same codespace

A codespace's built-in token can only access the repository the codespace was
created from. Cloning or pushing any other **private** repository fails with
`403 Write access to repository not granted`, and SSH URLs fail because a
codespace has no SSH key. Use HTTPS with your personal GitHub login instead.

Sign in once per codespace, if `gh` is not already logged in with your account:

```bash
env -u GITHUB_TOKEN gh auth login      # GitHub.com → HTTPS → browser login
```

Clone with credentials stored in the new clone, so later `pull` and `push`
work without extra steps:

```bash
cd /workspaces
git clone \
  -c credential.helper= \
  -c 'credential.helper=!env -u GITHUB_TOKEN gh auth git-credential' \
  https://github.com/OWNER/REPO.git
```

For a repository that was already cloned and now refuses to pull or push, run
this once inside it:

```bash
git config credential.helper ''
git config --add credential.helper '!env -u GITHUB_TOKEN gh auth git-credential'
```

In those repositories, prefix `gh` commands the same way, for example
`env -u GITHUB_TOKEN gh pr create --fill`.

Optional shortcut, saved to your shell profile in this codespace:

```bash
cat >> ~/.bashrc <<'BASHRC'
ghclone() {
  cd /workspaces && git clone \
    -c credential.helper= \
    -c 'credential.helper=!env -u GITHUB_TOKEN gh auth git-credential' \
    "https://github.com/$1.git" && cd "$(basename "$1")"
}
BASHRC
source ~/.bashrc
ghclone OWNER/REPO
```

To see several repositories side by side in the Explorer, use **File → Add
Folder to Workspace…**, or save a `.code-workspace` file in `/workspaces`
(outside every repository) and open it with **File → Open Workspace from
File…**.

### Codespace usage and billing

**Typical usage** (measured 2026-10-14 on a 2-core codespace):

| Item | Usage | Notes |
|---|---|---|
| A cloned repository | a few MB each | Clones are tiny next to everything else |
| Tools and caches in `~` | about 1.6 GB | Go module cache, CLI tools, editor settings |
| Codespace disk overall | about 12 GB of 32 GB | Mostly the base image: OS, languages, tools |
| Memory (RAM) | about 2–3 GB of 8 GB | Comes with the machine size; not billed separately |

**Free monthly allowance** on a GitHub Free personal account:

| | Included | Counts when |
|---|---|---|
| Compute | 120 core-hours (60 hours on a 2-core machine) | the codespace is **running** |
| Storage | 15 GB-month | the codespace **exists**, running or stopped |

One codespace kept for a whole month uses most of the storage allowance; a
second codespace kept alongside it is what pushes usage over, not extra clones.

**When GitHub charges:** only if a Codespaces spending limit above $0 is set.
With the default $0 limit, codespaces stop when the allowance runs out and
can be used again when it resets at the start of the next month. Nothing is
billed and files are kept.

**Check actual usage:** github.com/settings/billing, under Codespaces. The
allowances above can change; GitHub's documentation page "About billing for
GitHub Codespaces" has the current figures.

**Keeping usage low:**

- **Stop** the codespace when you finish (Command Palette → *Codespaces: Stop
  Current Codespace*) rather than closing the tab; idle codespaces keep
  running until they time out.
- Shorten the idle timeout to 10–15 minutes at github.com/settings/codespaces.
- Stay on a 2-core machine; larger machines use the allowance faster.
- Keep one codespace and clone extra repositories into it, rather than
  creating a codespace per repository.
- Push your work before **deleting** a codespace; anything not pushed is lost.
- Delete clones you no longer need once their work is pushed.

---

## Contributing

See [CONTRIBUTING.html](https://info.qcloud.systems/CONTRIBUTING.html).
Fork, branch, test locally with `make dev`, open a pull request.

---

## Version history

- **3.1.0** — Engineering Notes, roadmap, FAQ, articles, policy pages, SEO
- **3.0.0** — Go static site generator; agreements removed from this site
- **2.0.0** — Grey/blue redesign, GitHub Actions deployment
- **1.0.0** — Initial documentation hub

---

## Links

- **Site**: https://info.qcloud.systems
- **GitHub**: https://github.com/qcloud-systems

---

## License

MIT. See the LICENSE file.
