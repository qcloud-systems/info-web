package main

// Site-level content for the public information hub.
//
// This site is deliberately public in full. Nothing here is gated, and
// nothing client-specific belongs here -- agreements and customer documents
// live on qcloud.systems. See README.md for the reasoning.

type Site struct {
	Name    string
	Tagline string
	Domain  string
	GitHub  string
	Medium  string
	MainURL string
	Nav     []NavLink
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
	Name string
	// Kind selects the client-side adapter: "statuspage" or "instatus".
	Kind string
	// API returns the provider's current status.
	API string
	// Incidents returns recent incident history, used to colour the daily
	// bars. Statuspage only; empty for providers that publish no history.
	Incidents string
	// Page is the human-readable status page to link to.
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
// The daily bars are reconstructed from each provider's reported incidents,
// which is the only history Statuspage exposes publicly. A day is marked
// degraded when a reported incident overlapped it. An outage the provider
// never posted will show green -- these bars track disclosure, not measured
// uptime.
func statusServices() []StatusService {
	return []StatusService{
		{
			Name:      "Supabase",
			Kind:      kindStatuspage,
			API:       "https://status.supabase.com/api/v2/status.json",
			Incidents: "https://status.supabase.com/api/v2/incidents.json",
			Page:      "https://status.supabase.com",
		},
		{
			Name:      "GitHub",
			Kind:      kindStatuspage,
			API:       "https://www.githubstatus.com/api/v2/status.json",
			Incidents: "https://www.githubstatus.com/api/v2/incidents.json",
			Page:      "https://www.githubstatus.com",
		},
		{
			Name:      "Cloudflare",
			Kind:      kindStatuspage,
			API:       "https://www.cloudflarestatus.com/api/v2/status.json",
			Incidents: "https://www.cloudflarestatus.com/api/v2/incidents.json",
			Page:      "https://www.cloudflarestatus.com",
		},
		{
			// Fly.io runs Statuspage at status.flyio.net. Verify in the
			// browser console on first deploy -- see POST-DEPLOY.md.
			Name:      "Fly.io",
			Kind:      kindStatuspage,
			API:       "https://status.flyio.net/api/v2/status.json",
			Incidents: "https://status.flyio.net/api/v2/incidents.json",
			Page:      "https://status.flyio.net",
		},
		{
			// Railway runs Instatus, which exposes current state only --
			// no incident history, so its bars stay in the "no data" state.
			// Verify the host in the browser on first deploy.
			Name: "Railway",
			Kind: kindInstatus,
			API:  "https://status.railway.com/summary.json",
			Page: "https://status.railway.com",
		},
	}
}

func site() Site {
	return Site{
		Name:    "QCS",
		Tagline: "Information Hub",
		Domain:  "info.qcloud.systems",
		GitHub:  "https://github.com/qcloud-systems",
		Medium:  "https://qcloudsystems.medium.com/",
		MainURL: "https://qcloud.systems",
		Nav: []NavLink{
			{Label: "Overview", Href: "#overview"},
			{Label: "Status", Href: "#status"},
			{Label: "Prototypes", Href: "#prototypes"},
			{Label: "Projects", Href: "#projects"},
			{Label: "Documentation", Href: "#documentation"},
			{Label: "Contact", Href: "#contact"},
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

func pages() []Page {
	return []Page{
		{
			Path:     "index.html",
			Template: "index.html",
			Title:    "QCS — Information Hub",
			Desc:     "A public reference for QCS infrastructure, documentation, prototypes, and active projects.",
		},
	}
}

// documents are pass-through files kept as hand-written HTML because their
// text is policy or guidance that should not be reflowed by a template.
func documents() []string {
	return []string{
		"privacy-policy.html",
		"CONTRIBUTING.html",
	}
}
