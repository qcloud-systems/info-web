---
description: How a page goes from private notes to the published knowledge base.
---
# Contributing

Every page starts as Markdown in `docs-site/docs/`. It gets reviewed in a pull request and goes live on docs.qcloud.systems once merged.

## From notes to a published page

1. **Pick the area.** Where does the page belong?

    | If the page… | It goes in |
    | --- | --- |
    | explains how something works or what was learned | `engineering-notes/<topic>/` |
    | walks someone through running a service themselves | `self-hosting/<topic>/` |
    | documents a QCS product for customers | `products/<stage>/` |

2. **Start from a [template](templates.md).** Copy the matching template into the folder and name the file in `kebab-case.md`.
3. **Add it to the navigation** in `mkdocs.yml`, under its section. Then update that section's `index.md`: move the topic from *Planned pages* to *Published*.
4. **Preview it.**

    ```bash
    cd docs-site
    pip install -r requirements.txt
    mkdocs serve        # http://127.0.0.1:8000, reloads on save
    ```

5. **Open a pull request.** CI runs a strict build. Broken links, missing pages and anything that looks like a credential fail the check before the page can ship.

## Review checklist

Each page is checked for:

- [ ] **Accuracy.** Commands run as written, versions are current, and nothing relies on a deprecated flag or API.
- [ ] **Safety.** Placeholders like `<bucket-name>` and `<your-api-key>` instead of real values, with no account IDs, hostnames or internal IPs.
- [ ] **Placement.** It sits in the right area and links forward along the learning path.
- [ ] **Clarity.** A description in the front matter, tags, and headings that make sense in the table of contents.

!!! danger "This site is public"
    The repository and the site are both public, and git history keeps everything that was ever pushed. Strip credentials, client names and internal details from notes **before** they are committed. The build blocks common secret formats, but it cannot catch everything.

## Writing style

- Write for someone smart who hasn't done this before.
- Lead with the command, then explain it.
- Keep *Lessons learned* honest: what broke is often the most useful part.
- Prefer one topic per page. Split a page once it needs more than one goal.
