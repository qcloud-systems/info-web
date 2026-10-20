# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Semantic Versioning](https://semver.org/).

---

## [3.2.1] — 2026-10-20
### Changed
- The site, its documentation, and its code are now vendor-neutral: hosting
  and service providers are described by what they do, not by name.
- Upstream status feeds are identified by format rather than product, and
  status providers are labelled by role only.

- The Privacy Policy now covers this site only: what it collects (hosting
  logs only), what visitors' browsers fetch from other services, and visitor
  rights. Health-data and wearable-integration terms belonged to a separate
  project and were removed.
- Brand marketing services are announced as now available on the home page.

### Removed
- The optional web analytics setting, which was never enabled.
- The Rizal Center roadmap item.

---

## [3.2.0] — 2026-10-20
### Added
- **Trust & Status Center** — the site now focuses on service status, notices,
  and policies.
- **Notices** — incident reports, maintenance windows, security advisories, and
  service changes, with categories, search, and an RSS feed.
- **How It's Built** — a page explaining how the Go generator behind this site
  works, from validation to the notices engine.
- **Policies** section grouping the Privacy Policy, Security Policy, Terms of
  Use, and Accessibility Statement.
- Notice categories are checked at build time; an unknown category fails the
  build.

### Changed
- The Markdown engine that powered Engineering Notes now powers Notices.
- Home page, navigation, roadmap, and FAQ reorganized around status and trust.

### Removed
- Engineering Notes, the Articles section, Prototypes, and the general
  Documentation section.
- Links and references to other QCS sites.

---

## [3.1.0] — 2026-10-10
### Added
- **What's New** — the latest release highlights, right on the home page.
- **Roadmap** — what's in progress, what's planned, and what has shipped.
- **Articles** — the latest QCS posts from Medium.
- **FAQ** — answers to common questions about this hub and QCS services.
- **Engineering Notes** — practical how-tos, written in Markdown, starting
  with building a static site with Hugo. Includes search, topic tags with
  their own pages, a table of contents, copy buttons on code, linkable
  headings, related notes, updated dates, and an RSS feed.
- **Security Policy**, **Terms of Use**, and **Accessibility** pages.
- A helpful **404 page** for broken or outdated links.
- `/.well-known/security.txt`, `sitemap.xml`, and `robots.txt`, generated on
  every build.
- Link previews (Open Graph and Twitter cards) with a branded preview image,
  canonical URLs, and structured data for search engines.
- A "last checked" time under the upstream status list.
- An optional web analytics setting, off by default (removed in 3.2.1).
- Unit tests, run on every deploy.

### Changed
- The footer links to every policy page and shows the version and
  last-updated date.
- Sidebar links now work from every page, not just the home page.
- The Projects section was folded into Prototypes & Active Projects.

---

## [3.0.0] — 2026-10-08
### Added
- Go static site generator (`main.go`, `content.go`, `templates/`), replacing
  hand-maintained HTML. `go run . serve` for development, `go run . build` to
  render into `./public`.
- Build-time validation: missing templates, missing documents, projects with
  no URL, and a missing `CNAME` all fail the build.
- Shared `css/doc.css` for the long documents.
- `Makefile` with `dev`, `build`, `check`, `fmt`, `vet`, `clean`.

### Changed
- Agreement templates removed from this site.
- Repalette to the QCS brand colours: charcoal ink `#1a2332`, action blue
  `#0066cc`, sky-blue accent `#6cc5e6` taken from the wordmark.
- Deploy workflow builds with Go and publishes `./public`.
- README rewritten.

### Fixed
- The Contributing Guidelines link now works.

---

## [2.0.0] — 2026-04-19
### Added
- Complete UI redesign with sleek grey and blue color scheme
- Professional sidebar navigation with improved styling
- Privacy Policy document (`privacy-policy.html`) with table of contents
- GitHub Actions workflow for automated deployment
- Documentation section
- Favicon support across all pages
- Enhanced responsive design for mobile and tablet devices
- DNS configuration guide for custom domain setup

### Changed
- Overhauled HTML structure for better semantic organization
- Completely rewrote CSS with modern grey (#2d3e50-#34495e) and blue (#0066cc) theme
- Moved deployment to GitHub Actions
- Improved README with comprehensive sections on deployment and DNS

### Fixed
- Removed incomplete Google Maps integration code
- Cleaned up unused placeholder content and dummy links
- Fixed CSS conflicts from multiple overlapping stylesheets
- Improved workflow reliability with proper git repository handling

---

## [1.0.0] — 2026-02-18
### Added
- Docs-style layout with sidebar navigation.
- Responsive HTML/CSS structure.
- README with technical overview and deployment details.
- Versioning and contribution guidelines.
- Base documentation sections (Introduction, Setup, Architecture, API, Contact).

### Changed
- N/A (initial release)

### Fixed
- N/A (initial release)