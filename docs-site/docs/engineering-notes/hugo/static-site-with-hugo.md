---
description: Create, preview, and build a static website with Hugo, then host it privately on Amazon S3 behind CloudFront.
tags:
  - Hugo
  - AWS
  - Static Sites
---
# Building a Static Site with Hugo

How to create, preview, and build a static website with [Hugo](https://gohugo.io/), a fast static site generator written in Go, and how to host the result on Amazon S3.

## Prerequisites

- Hugo installed. Check with `hugo version`. Many themes need the **extended** edition.
- Git, for version control and for installing a theme.

## 1. Create the site

```bash
hugo new site my-site
cd my-site
git init
```

This creates the standard layout: `archetypes/`, `content/`, `data/`, `layouts/`, `static/`, `themes/`, and a config file — `hugo.toml` in current versions, `config.toml` in older ones.

## 2. Add a theme

Install the theme as a Git submodule so it stays pinned to a specific commit:

```bash
git submodule add https://github.com/<owner>/<theme>.git themes/<theme>
echo "theme = '<theme>'" >> hugo.toml
```

Browse themes at [themes.gohugo.io](https://themes.gohugo.io/). Prefer one that is actively maintained and supports your Hugo version — an abandoned theme is the most common reason an old site stops building.

!!! warning "Themes are third-party code"
    A theme is third-party code that runs in your build. Read what it pulls in, and review its changes before updating the submodule.

## 3. Preview locally

```bash
hugo server -D
```

| Flag | What it does |
| --- | --- |
| `-D` | Include draft pages |
| `--disableFastRender` | Rebuild everything on each change, not just the changed page |

The site is served from memory at <http://localhost:1313/> and reloads as you edit content, layouts, static files, or config. It binds to `127.0.0.1`, so it is only reachable from your own machine. Press `Ctrl+C` to stop.

There is no need to run Hugo with `sudo` — it only reads and writes inside the project folder.

## 4. Build

```bash
hugo --minify
```

The finished site is written to `public/`. Every file there is plain HTML, CSS, and JavaScript that any static host can serve. Add `public/` to `.gitignore` if you build in CI.

## 5. Host on Amazon S3

The recommended setup keeps the bucket **private** and serves it through CloudFront:

1. Create an S3 bucket and leave **Block Public Access** turned on.
2. Create a CloudFront distribution with the bucket as its origin, using **Origin Access Control** so only CloudFront can read the bucket.
3. Attach a free TLS certificate from AWS Certificate Manager (it must be issued in `us-east-1` for CloudFront) and point your domain at the distribution.
4. Upload the site:

   ```bash
   aws s3 sync public/ s3://<bucket-name> --delete
   aws cloudfront create-invalidation --distribution-id <id> --paths "/*"
   ```

Hugo can also do the upload itself with `hugo deploy`, configured in a `[deployment]` section of `hugo.toml`.

### Retiring a site safely

!!! danger "Subdomain takeover"
    When a site is shut down, **delete its DNS record before deleting the bucket or distribution**. A DNS name left pointing at a deleted bucket can be claimed by anyone who creates a bucket with the same name, letting them serve content on your domain.

## Lessons learned

- **Themes age faster than Hugo does.** The `Page.URL` deprecation warning showed up years before it became a hard failure. Treat warnings from a theme as early notice: pin the theme, read Hugo's release notes before upgrading, and prefer a maintained theme — or a small set of your own layouts — over a large abandoned one.
- **Elevated privileges create their own problems.** Running the build with `sudo` leaves root-owned files in `public/` and `resources/`, and the next normal build then fails with permission errors. Build as your own user from the start.
- **Copying files by hand does not scale.** Moving the built site between WSL and Windows by hand worked for a first look, but it is neither repeatable nor reviewable. A scripted deploy (`aws s3 sync`, `hugo deploy`, or a CI pipeline) produces the same result every time.
- **Retire DNS with the site.** When an experimental subdomain is shut down, its DNS record goes first. Nothing left pointing at deleted infrastructure means nothing left to take over.

## Troubleshooting

**`WARN … Page.URL is deprecated`** — the theme uses a template variable that newer Hugo versions have removed, which turns the warning into a build failure. Replace `.URL` with `.Permalink` (full URL) or `.RelPermalink` (path only) in the theme's layouts, or switch to a maintained theme.

## Further reading

- [Hugo documentation](https://gohugo.io/documentation/)
- [Using S3 and Hugo to Create a Hosting Static Website](https://capgemini.github.io/development/Using-S3-and-Hugo-to-Create-Hosting-Static-Website/) — Capgemini Engineering
