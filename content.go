package main

import "html/template"

// Site-level content for the QCS Trust & Status Center.
//
// This site is deliberately public in full. Nothing here is gated, and
// nothing client-specific belongs here -- agreements and customer documents
// are not published from this repo. See README.md for the reasoning.

type Site struct {
	Name    string
	Tagline string
	Domain  string
	// BaseURL is the canonical origin, used for canonical links, Open Graph
	// categories, sitemap.xml and robots.txt. No trailing slash.
	BaseURL string
	GitHub  string
	// Repo is this site's own repository; the security policy points
	// reporters at its private advisory form.
	Repo   string
	Medium string
	Nav    []NavLink
}

type NavLink struct {
	Label    string
	Href     string
	External bool
}

type Page struct {
	Path     string
	Template string
	Title    string
	Desc     string
	NoIndex  bool
	// Notice is the slug of the notice this page renders; empty for other pages.
	Notice string
	// Category is the slug of the category this page lists; empty for other pages.
	Category string
}

// Author is credited on every notice: a byline at the top, an author box at
// the bottom, and the author in the structured data search engines read.
// Leave Name empty to publish notices in the organization's name instead,
// which is what a trust center should do.
type Author struct {
	Name string
	Role string
	URL  string
	Bio  string
}

func author() Author {
	return Author{
		Name: "", // notices are published as QCS, not an individual
	}
}

// RoadmapItem is one entry on the roadmap. Stage must be one of the
// stage* constants; validate() rejects anything else.
type RoadmapItem struct {
	Title string
	Desc  string
	Stage string
}

const (
	stagePlanned    = "planned"
	stageInProgress = "in-progress"
	stageShipped    = "shipped"
)

// FAQ is one question and answer. Answer is trusted HTML so it can link.
type FAQ struct {
	Q string
	A template.HTML
}

// StatusService is an upstream provider whose health is polled from the
// browser. Only providers whose status feed allows cross-origin requests can
// be listed; one that does not would need a server-side proxy, which this
// static site does not have.
type StatusService struct {
	// Role is what the page displays: the function this provider serves,
	// not who it is. A status strip only needs to tell a visitor which part
	// of the system is degraded; naming vendors publishes an inventory of
	// the stack and buys nothing for the reader.
	//
	// This is not concealment -- see the comment on statusServices.
	Role string
	// Kind selects the client-side adapter for the feed's format.
	Kind string
	// API returns the provider's current status.
	API string
	// Incidents returns recent incident history, used to colour the daily
	// bars. api-v2 feeds only; empty for feeds that publish no history.
	Incidents string
}

// Supported feed formats. Named for their shape, not for any product.
const (
	// kindAPIv2 is a feed with /api/v2/status.json for the current state and
	// /api/v2/incidents.json for recent incident history.
	kindAPIv2 = "api-v2"
	// kindSummary is a single /summary.json with the current state only.
	kindSummary = "summary"
)

// statusServices lists upstream providers polled from the browser.
//
// On the Role labels: the page shows what each provider does, not who it
// is. That keeps a vendor inventory out of the visible copy, which is worth
// doing -- but it is NOT concealment, and must not be treated as a control.
// The browser does the fetching, so each endpoint is necessarily present in
// the page source and visible in the network tab, and this repository is
// public. Hiding the providers properly would require proxying every request
// server-side. The real defence for the services behind these feeds is their
// own configuration -- access rules, keeping service keys off the client --
// not label wording.
//
// The daily bars are reconstructed from each provider's reported incidents,
// which is the only history these feeds expose. A day is marked degraded when
// a reported incident overlapped it. An outage the provider never posted will
// show green -- these bars track disclosure, not measured uptime.
func statusServices() []StatusService {
	return []StatusService{
		{
			Role:      "Database",
			Kind:      kindAPIv2,
			API:       "https://status.supabase.com/api/v2/status.json",
			Incidents: "https://status.supabase.com/api/v2/incidents.json",
		},
		{
			Role:      "Source control & CI",
			Kind:      kindAPIv2,
			API:       "https://www.githubstatus.com/api/v2/status.json",
			Incidents: "https://www.githubstatus.com/api/v2/incidents.json",
		},
		{
			Role:      "CDN & DNS",
			Kind:      kindAPIv2,
			API:       "https://www.cloudflarestatus.com/api/v2/status.json",
			Incidents: "https://www.cloudflarestatus.com/api/v2/incidents.json",
		},
		{
			Role:      "Edge hosting",
			Kind:      kindAPIv2,
			API:       "https://status.flyio.net/api/v2/status.json",
			Incidents: "https://status.flyio.net/api/v2/incidents.json",
		},
		{
			// A summary feed: current state only, no incident history, so
			// its bars stay in the "no data" state.
			//
			// The JSON lives on a different host from the provider's public
			// status page. The public page's domain serves rendered HTML for
			// every path, so /summary.json there would fail to parse.
			Role: "Application backend",
			Kind: kindSummary,
			API:  "https://railway.instatus.com/summary.json",
		},
	}
}

func site() Site {
	return Site{
		Name:    "QCS",
		Tagline: "Trust & Status Center",
		Domain:  "info.qcloud.systems",
		BaseURL: "https://info.qcloud.systems",
		GitHub:  "https://github.com/qcloud-systems",
		Repo:    "https://github.com/qcloud-systems/info-web",
		Medium:  "https://qcloudsystems.medium.com/",
		// Hrefs are root-relative ("/#status", not "#status") so the sidebar
		// still works from the secondary pages.
		Nav: []NavLink{
			{Label: "Overview", Href: "/#overview"},
			{Label: "Status", Href: "/#status"},
			{Label: "Notices", Href: "/notices/"},
			{Label: "What's New", Href: "/#whats-new"},
			{Label: "Roadmap", Href: "/#roadmap"},
			{Label: "Policies", Href: "/#policies"},
			{Label: "FAQ", Href: "/#faq"},
			{Label: "How It's Built", Href: "/how-its-built.html"},
			{Label: "Contact", Href: "/#contact"},
		},
	}
}

// roadmap is shown newest-stage-first on the home page. Add an entry here to
// publish it; nothing else needs to change.
func roadmap() []RoadmapItem {
	return []RoadmapItem{
		{
			Title: "Brand marketing services",
			Desc:  "A future package for businesses establishing their brand online: a domain registered in the business's own name and a branded landing page.",
			Stage: stagePlanned,
		},
		{
			Title: "QCS service status",
			Desc:  "Status for QCS's own services, alongside the upstream providers already tracked.",
			Stage: stagePlanned,
		},
		{
			Title: "Search engine indexing",
			Desc:  "Register this site with search engines and submit its sitemap.",
			Stage: stagePlanned,
		},
		{
			Title: "Trust & Status Center",
			Desc:  "This site, refocused on service status, notices, and policies, with its build documented.",
			Stage: stageShipped,
		},
	}
}

// noticeCategories are the only categories a notice may use; validate()
// rejects anything else. Order is the display order on the Notices index.
func noticeCategories() []string {
	return []string{"Incident", "Maintenance", "Security Advisory", "Service Change"}
}

func faqs() []FAQ {
	return []FAQ{
		{
			Q: "Where are incidents and planned maintenance announced?",
			A: `On the <a href="/notices/">Notices</a> page, which also has an RSS feed you can subscribe to.`,
		},
		{
			Q: "Is the status section a status page for QCS services?",
			A: `No. It reports only on the third-party platforms QCS projects depend on, read from each provider's own status feed.`,
		},
		{
			Q: "Does this site track visitors?",
			A: `No. It sets no cookies and runs no analytics or tracking scripts. The <a href="/privacy-policy.html">Privacy Policy</a> explains the little information that is involved in serving a website.`,
		},
		{
			Q: "How do I report a security vulnerability?",
			A: `Please report it privately &mdash; not in a public issue. The <a href="/security.html">Security Policy</a> explains how.`,
		},
		{
			Q: "How is this site built?",
			A: `It is a small Go program that renders static pages at build time. <a href="/how-its-built.html">How It's Built</a> walks through every part of it.`,
		},
		{
			Q: "How can I suggest a correction or improvement to this site?",
			A: `Open an issue on <a href="https://github.com/qcloud-systems/info-web" target="_blank" rel="noopener">GitHub</a>, or see the <a href="/CONTRIBUTING.html">Contributing Guidelines</a>.`,
		},
	}
}

// pages are rendered through templates/base.html. NoIndex keeps a page out
// of sitemap.xml.
func pages() []Page {
	return []Page{
		{
			Path:     "index.html",
			Template: "index.html",
			Title:    "QCS — Trust & Status Center",
			Desc:     "Service status, notices, and policies for QCS: how QCS keeps services reliable and handles data and security.",
		},
		{
			Path:     "how-its-built.html",
			Template: "how-its-built.html",
			Title:    "How It's Built — QCS Trust & Status Center",
			Desc:     "How this site is built: a Go static site generator with a Markdown notices engine, build-time validation, and a secret scanner.",
		},
		{
			Path:     "security.html",
			Template: "security.html",
			Title:    "Security Policy — QCS Trust & Status Center",
			Desc:     "How to report a security vulnerability to QCS.",
		},
		{
			Path:     "accessibility.html",
			Template: "accessibility.html",
			Title:    "Accessibility — QCS Trust & Status Center",
			Desc:     "QCS's accessibility commitment for this site, known limitations, and how to give feedback.",
		},
		{
			// Static hosts serve /404.html for any missing path.
			Path:     "404.html",
			Template: "404.html",
			Title:    "Page Not Found — QCS Trust & Status Center",
			Desc:     "The page you were looking for does not exist.",
			NoIndex:  true,
		},
	}
}

// documents are pass-through files kept as hand-written HTML because their
// text is policy or guidance that should not be reflowed by a template.
func documents() []string {
	return []string{
		"privacy-policy.html",
		"terms-of-use.html",
		"CONTRIBUTING.html",
	}
}
