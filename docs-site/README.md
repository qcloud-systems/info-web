# QCS Knowledge Base

Source for **docs.qcloud.systems**: engineering notes, self-hosting guides and product documentation, built with [MkDocs Material](https://squidfunk.github.io/mkdocs-material/).

This folder is self-contained. It shares a repository with info.qcloud.systems, but it has its own toolchain (Python), build and host. Nothing here touches the Go site, and it can move to its own repository unchanged.

## Quick start

```bash
cd docs-site
python3 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
mkdocs serve            # http://127.0.0.1:8000, live reload
mkdocs build --strict   # what CI runs; output in ./site
```

## Layout

```
docs-site/
├── mkdocs.yml              site config and navigation (add every new page here)
├── requirements.txt        pinned MkDocs Material version
├── hooks/secret_scan.py    fails the build on credential-like strings
├── overrides/home.html     branded homepage hero
├── templates/              page templates, shown on /contributing/templates/
└── docs/
    ├── index.md            homepage
    ├── engineering-notes/  infrastructure, security, aws, hugo, github, automation
    ├── self-hosting/       docker, linux, deployments, cloudflare, monitoring
    ├── products/           installation, configuration, administration, troubleshooting
    ├── contributing/       how to write and review a page
    ├── tags.md             tag index (from each page's front matter)
    └── assets/             brand stylesheet and Q mark
```

Brand colours live in `docs/assets/stylesheets/qcs.css` and match info.qcloud.systems.

## Adding a page

See `docs/contributing/index.md`, which is also published at `/contributing/`. In short: copy a template, add the page to `nav` in `mkdocs.yml`, preview, and open a PR.

## Deploying to docs.qcloud.systems (Cloudflare Pages)

One-time setup in the Cloudflare dashboard:

1. **Workers & Pages → Create → Pages → Connect to Git**, then pick `qcloud-systems/info-web`.
2. Build settings:

   | Setting | Value |
   | --- | --- |
   | Production branch | `main` |
   | Framework preset | None |
   | Root directory | `docs-site` |
   | Build command | `pip install -r requirements.txt && mkdocs build` |
   | Build output directory | `site` |
   | Environment variable | `PYTHON_VERSION` = `3.12` |

3. Under **Build watch paths**, include only `docs-site/*`, so changes to the Go site don't trigger a docs build.
4. **Custom domains → Set up a custom domain** → `docs.qcloud.systems`. If qcloud.systems DNS is on Cloudflare, the CNAME is created for you.

Every pull request gets its own preview URL, and merges to `main` go live.

## About MkDocs 2.0

The Material team has announced that MkDocs 2.0 will not support Material themes or plugins. Material 9.7.5+ caps MkDocs below 2.0, so the pinned version keeps building. Material's critical maintenance runs through May 2027. Its successor, [Zensical](https://zensical.org/), reads `mkdocs.yml` and is designed as a drop-in replacement, so moving later should be a config check rather than a rewrite.
