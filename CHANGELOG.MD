# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Semantic Versioning](https://semver.org/).

---

## [3.0.0] — 2026-10-08
### Removed
- **The password-protected area, which never actually protected anything.**
  Passwords were injected into client-side JavaScript at deploy time and were
  readable with View Source; separately, the agreement files were published at
  predictable public URLs and loaded without any password at all. The form
  only pointed an iframe at an already-public file.
- `protected.html`, `service-agreement.html`, `nda-agreement.html` and
  `beta-testing-agreement.html`. Blank templates now live on qcloud.systems;
  executed client documents are not published anywhere.
- The password-injection step and the `PROTECTED_PASSWORD_*` secrets, along
  with the workflow step that echoed the injected line into the build log.

### Added
- Go static site generator (`main.go`, `content.go`, `templates/`), replacing
  hand-maintained HTML. `go run . serve` for development, `go run . build` to
  render into `./public`.
- Build-time validation: missing templates, missing documents, projects with
  no URL, and a missing `CNAME` all fail the build.
- A guard that **fails the build** if a file named like one of the removed
  agreements reappears in `static/`.
- Shared `css/doc.css` for the long documents, replacing ~280 lines of
  inline CSS duplicated across them.
- `Makefile` with `dev`, `build`, `check`, `fmt`, `vet`, `clean`.
- `.nojekyll` written on every build.

### Changed
- Repalette to match qcloud.systems: charcoal ink `#1a2332`, action blue
  `#0066cc`, sky-blue accent `#6cc5e6` taken from the wordmark.
- Deploy workflow builds with Go and publishes `./public`.
- README rewritten; the previous version documented the password scheme as
  secure, including the claim that passwords were "never visible in the
  public source code," which was not true.

### Fixed
- `CONTRIBUTING.html` was linked from the hub but never copied into the
  published output, so the link 404'd.
- The beta testing agreement's password placeholder was never substituted and
  its file was never published, leaving a card that could not work.

---

## [2.0.0] — 2026-04-19
### Added
- Complete UI redesign with sleek grey and blue color scheme
- Professional sidebar navigation with improved styling
- Password-protected area for more sensitive documents
- Service Agreement document (`service-agreement.html`) with signature blocks
- Privacy Policy document (`privacy-policy.html`) with table of contents
- GitHub Actions workflow for automated deployment with password injection
- GitHub Secrets integration for secure password storage
- Local `.env` file for development (excluded from git)
- `.gitignore` configuration to prevent committing sensitive files
- Documentation section with public and sensitive documents
- Favicon support across all pages
- Enhanced responsive design for mobile and tablet devices
- Security documentation in README explaining protection limitations
- DNS configuration guide for custom domain setup

### Changed
- Overhauled HTML structure for better semantic organization
- Completely rewrote CSS with modern grey (#2d3e50-#34495e) and blue (#0066cc) theme
- Updated deployment process to use GitHub Actions instead of basic GitHub Pages
- Improved README with comprehensive sections on security, DNS, and deployment
- Updated version to 2.0.0 to reflect major redesign

### Fixed
- Removed incomplete Google Maps integration code
- Cleaned up unused placeholder content and dummy links
- Removed temporary `commands.out` file
- Fixed CSS conflicts from multiple overlapping stylesheets
- Improved workflow reliability with proper git repository handling

### Security
- Passwords now stored securely in GitHub Secrets instead of source code
- Added `.gitignore` to prevent accidental commits of sensitive files
- Passwords injected at build time via GitHub Actions
- Added security limitations documentation to README

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