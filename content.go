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
