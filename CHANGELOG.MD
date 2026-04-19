# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Semantic Versioning](https://semver.org/).

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