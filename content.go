package main

import "html/template"

// Site-level content for the public information hub.
//
// This site is deliberately public in full. Nothing here is gated, and
// nothing client-specific belongs here -- agreements and customer documents
// live on qcloud.systems. See README.md for the reasoning.

type Site struct {
	Name    string
	Tagline string
	Domain  string
	// BaseURL is the canonical origin, used for canonical links, Open Graph
	// tags, sitemap.xml and robots.txt. No trailing slash.
	BaseURL string
	GitHub  string
	// Repo is this site's own repository; the security policy points
	// reporters at its private advisory form.
	Repo   string
	Medium string
	// MediumFeed is fetched at build time for the Articles section. A failed
	// fetch is logged and the section falls back to a link -- it never fails
	// the build, because Medium being down is not a reason to block a deploy.
	MediumFeed string
	MainURL    string
	// CloudflareAnalytics is a Cloudflare Web Analytics beacon token. Empty
	// means no analytics script is emitted at all. Cloudflare's analytics is
	// free and cookieless; the token is public by design (it ships in the
	// page), so it is safe to keep here.
	CloudflareAnalytics string
	Nav                 []NavLink
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
	// Note is the slug of the note this page renders; empty for other pages.
	Note string
	// Tag is the slug of the tag this page lists; empty for other pages.
	Tag string
}

// Author is credited on every engineering note: a byline at the top, an
// author box at the bottom, and the author in the structured data search
// engines read. Leave Name empty to omit the byline and author box.
type Author struct {
	Name string
	Role string
	URL  string
	Bio  string
}

func author() Author {
	return Author{
		Name: "covertdaemon",
		Role: "Cloud Engineer",
		URL:  "https://github.com/qcloud-systems",
		Bio:  "Infrastructure without lock-in. Software engineering, automation, and multi-cloud platforms.",
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

// Project is one entry in the prototypes grid.
type Project struct {
	Name string
	Desc string
	URL  string
	CTA  string
}

// StatusService is an upstream provider whose health is polled from the
// browser. Only services running Atlassian Statuspage are listed: they expose
// a CORS-enabled summary at /api/v2/status.json, so no proxy is needed.
//
// Azure deliberately has no entry. It publishes no CORS-friendly JSON
// endpoint, so including it would require a server-side proxy -- compute that
// this static site does not have.
type StatusService struct {
	// Role is what the page displays: the function this provider serves,
	// not who it is. A status strip only needs to tell a visitor which part
	// of the system is degraded; naming vendors publishes an inventory of
	// the stack and buys nothing for the reader.
	//
	// Note this is not concealment -- see the comment on statusServices.
	Role string
	// Name is the vendor, for maintainers. Never rendered.
	Name string
	// Kind selects the client-side adapter: "statuspage" or "instatus".
	Kind string
	// API returns the provider's current status.
	API string
	// Incidents returns recent incident history, used to colour the daily
	// bars. Statuspage only; empty for providers that publish no history.
	Incidents string
	// Page is the vendor's own status page, for maintainers. Not linked
	// from the site -- linking it would re-reveal the vendor the Role
	// label is there to keep out of the copy.
	Page string
}

// Supported adapter kinds.
const (
	kindStatuspage = "statuspage"
	kindInstatus   = "instatus"
)

// statusServices lists upstream providers polled from the browser.
//
// Only providers that serve CORS-enabled JSON can be listed; one that does
// not would need a server-side proxy this static site does not have.
//
// On the Role labels: the page shows what each provider does, not who it
// is. That keeps a vendor inventory out of the visible copy, which is worth
// doing -- but it is NOT concealment, and must not be treated as a control.
// The browser does the fetching, so each endpoint is necessarily present in
// the page source as a data-api attribute and visible in the network tab,
// and this repository is public. Hiding the vendors properly would require
// proxying every request server-side. The real defence for the services
// behind these names is their own configuration -- Supabase row-level
// security, keeping service keys off the client -- not label wording.
//
// The daily bars are reconstructed from each provider's reported incidents,
// which is the only history Statuspage exposes publicly. A day is marked
// degraded when a reported incident overlapped it. An outage the provider
// never posted will show green -- these bars track disclosure, not measured
// uptime.
func statusServices() []StatusService {
	return []StatusService{
		{
			Role:      "Database",
			Name:      "Supabase",
			Kind:      kindStatuspage,
			API:       "https://status.supabase.com/api/v2/status.json",
			Incidents: "https://status.supabase.com/api/v2/incidents.json",
			Page:      "https://status.supabase.com",
		},
		{
			Role:      "Source control & CI",
			Name:      "GitHub",
			Kind:      kindStatuspage,
			API:       "https://www.githubstatus.com/api/v2/status.json",
			Incidents: "https://www.githubstatus.com/api/v2/incidents.json",
			Page:      "https://www.githubstatus.com",
		},
		{
			Role:      "CDN & DNS",
			Name:      "Cloudflare",
			Kind:      kindStatuspage,
			API:       "https://www.cloudflarestatus.com/api/v2/status.json",
			Incidents: "https://www.cloudflarestatus.com/api/v2/incidents.json",
			Page:      "https://www.cloudflarestatus.com",
		},
		{
			Role:      "Edge hosting",
			Name:      "Fly.io",
			Kind:      kindStatuspage,
			API:       "https://status.flyio.net/api/v2/status.json",
			Incidents: "https://status.flyio.net/api/v2/incidents.json",
			Page:      "https://status.flyio.net",
		},
		{
			// Railway runs Instatus, which exposes current state only --
			// no incident history, so its bars stay in the "no data" state.
			//
			// The JSON lives on the instatus.com host, not on
			// status.railway.com: that domain serves the rendered status
			// page for every path, so /summary.json there returns HTML and
			// the client-side JSON parse fails.
			Role: "Application backend",
			Name: "Railway",
			Kind: kindInstatus,
			API:  "https://railway.instatus.com/summary.json",
			Page: "https://status.railway.com",
		},
	}
}

func site() Site {
	return Site{
		Name:       "QCS",
		Tagline:    "Information Hub",
		Domain:     "info.qcloud.systems",
		BaseURL:    "https://info.qcloud.systems",
		GitHub:     "https://github.com/qcloud-systems",
		Repo:       "https://github.com/qcloud-systems/info-web",
		Medium:     "https://qcloudsystems.medium.com/",
		MediumFeed: "https://qcloudsystems.medium.com/feed",
		MainURL:    "https://qcloud.systems",
		// Paste a Cloudflare Web Analytics token here to turn analytics on.
		CloudflareAnalytics: "",
		// Hrefs are root-relative ("/#status", not "#status") so the sidebar
		// still works from the secondary pages.
		Nav: []NavLink{
			{Label: "Overview", Href: "/#overview"},
			{Label: "What's New", Href: "/#whats-new"},
			{Label: "Status", Href: "/#status"},
			{Label: "Prototypes", Href: "/#prototypes"},
			{Label: "Roadmap", Href: "/#roadmap"},
			{Label: "Articles", Href: "/#articles"},
			{Label: "Documentation", Href: "/#documentation"},
			{Label: "Engineering Notes", Href: "/notes/"},
			{Label: "FAQ", Href: "/#faq"},
			{Label: "Contact", Href: "/#contact"},
			{Label: "qcloud.systems", Href: "https://qcloud.systems", External: true},
		},
	}
}

func projects() []Project {
	return []Project{
		{
			Name: "Rizal Center",
			Desc: "An interactive platform showcasing community and organizational information.",
			URL:  "https://rizalcenter.app",
			CTA:  "Visit Prototype",
		},
	}
}

// roadmap is shown newest-stage-first on the home page. Add an entry here to
// publish it; nothing else needs to change.
func roadmap() []RoadmapItem {
	return []RoadmapItem{
		{
			Title: "Search engine indexing",
			Desc:  "Register this site with Google Search Console and Bing Webmaster Tools and submit its sitemap.",
			Stage: stagePlanned,
		},
		{
			Title: "Search-friendly note titles",
			Desc:  "Retitle engineering notes to match how people search, such as \"How to Host a Hugo Site on Amazon S3\".",
			Stage: stagePlanned,
		},
		{
			Title: "Engineering notes on Medium",
			Desc:  "Cross-post engineering notes to Medium, credited to this site as the original source.",
			Stage: stagePlanned,
		},
		{
			Title: "QCS service status",
			Desc:  "A status section for QCS's own services, alongside the upstream providers already tracked.",
			Stage: stagePlanned,
		},
		{
			Title: "Expanded technical documentation",
			Desc:  "Architecture guides, API references, and deployment procedures for active projects.",
			Stage: stagePlanned,
		},
		{
			Title: "Rizal Center",
			Desc:  "Interactive platform for community and organizational information.",
			Stage: stageInProgress,
		},
		{
			Title: "Information hub rebuild",
			Desc:  "This site, rebuilt as a Go static site with upstream status and policy pages.",
			Stage: stageShipped,
		},
	}
}

func faqs() []FAQ {
	return []FAQ{
		{
			Q: "Is the status section a status page for QCS services?",
			A: `No. It reports only on the third-party platforms QCS projects depend on, read from each provider's own status feed.`,
		},
		{
			Q: "Where can I find service agreements, NDAs, or beta testing terms?",
			A: `Blank templates are published on <a href="https://qcloud.systems/legal/" target="_blank" rel="noopener">qcloud.systems</a>. Executed, client-specific documents are never posted publicly.`,
		},
		{
			Q: "How is wearable and health data handled?",
			A: `See the <a href="/privacy-policy.html">Privacy Policy</a>, including the section on Garmin Health API compliance.`,
		},
		{
			Q: "How do I report a security vulnerability?",
			A: `Please report it privately &mdash; not in a public issue. The <a href="/security.html">Security Policy</a> explains how.`,
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
			Title:    "QCS — Information Hub",
			Desc:     "A public reference for QCS infrastructure, documentation, prototypes, and active projects.",
		},
		{
			Path:     "security.html",
			Template: "security.html",
			Title:    "Security Policy — QCS Information Hub",
			Desc:     "How to report a security vulnerability to QCS.",
		},
		{
			Path:     "accessibility.html",
			Template: "accessibility.html",
			Title:    "Accessibility — QCS Information Hub",
			Desc:     "QCS's accessibility commitment for this site, known limitations, and how to give feedback.",
		},
		{
			// GitHub Pages serves /404.html for any missing path.
			Path:     "404.html",
			Template: "404.html",
			Title:    "Page Not Found — QCS Information Hub",
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
