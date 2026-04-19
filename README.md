# QCS — Information Hub

![Status](https://img.shields.io/badge/status-active-brightgreen)  
![License](https://img.shields.io/badge/license-MIT-blue)  
![Version](https://img.shields.io/badge/version-2.0.0-purple)

This repository contains the source code for **info.qcloud.systems**, a collaborative documentation and information hub for QCS projects built with **GitHub Pages**. The site features a professional grey and blue design with public documentation, prototype showcase, and secure document access.

---

## 🚀 Purpose

The goal of this site is to serve as a central hub for public QCS information and collaboration:

- **Project Prototypes** — Showcase active projects and prototypes (e.g., Rizal Center)
- **Public Documentation** — Privacy policies and public-facing agreements
- **Protected Documents** — Secure access to service agreements and internal documents
- **Information Hub** — Centralized reference for QCS infrastructure and development

The site is designed to be collaborative, maintainable, and accessible to both public contributors and team members.

---

## 🎨 Design

- **Color Scheme**: Chrome sleek grey (#2d3e50 to #34495e) with blue accents (#0066cc)
- **Typography**: Inter font family for clean, modern appearance
- **Responsive**: Mobile-friendly design works on all screen sizes
- **Sidebar Navigation**: Fixed sidebar for easy navigation between sections

---

## 🧰 Technology Stack

- **GitHub Pages** — Static hosting with custom domain
- **HTML5 + CSS3** — No build tools or dependencies required
- **GitHub Actions** — Automated password injection for protected documents
- **GitHub Secrets** — Secure password storage (never committed to repo)

---

## 📁 Repository Structure

```
/
├── index.html                    # Main hub page
├── style.css                     # Site styling (grey/blue theme)
├── protected.html                # Password-protected page
├── service-agreement.html        # Service agreement document
├── privacy-policy.html           # QCS privacy policy
├── favicon.ico                   # Site icon
├── .env                          # Local environment variables (NOT committed)
├── .gitignore                    # Excludes .env and sensitive files
├── .github/workflows/deploy.yml  # GitHub Actions deployment workflow
├── README.md                     # This file
└── CNAME                         # Custom domain configuration
```

---

## 🔐 Security & Secrets Management

### Local Development

A local `.env` file stores the protected page password:

```
PROTECTED_PASSWORD=your_password_here
```

**Important**: `.env` is listed in `.gitignore` and will **never** be committed to the repository.

### GitHub Secrets

The GitHub Actions workflow uses `PROTECTED_PASSWORD` secret:

1. Go to your GitHub repo → **Settings** → **Secrets and variables** → **Actions**
2. Create a new repository secret named `PROTECTED_PASSWORD`
3. Enter the secure password value
4. On every push to `main`, the workflow automatically injects the password

### How It Works

- `protected.html` contains: `const CORRECT_PASSWORD = "{{ PROTECTED_PASSWORD }}";`
- GitHub Actions replaces the placeholder with the actual secret
- The password is **never visible** in the public source code
- Only repository maintainers can view/edit the secret in GitHub

---

## 📄 Pages Overview

| Page | File | Access | Purpose |
|------|------|--------|---------|
| Home | `index.html` | Public | Main hub with overview and prototypes |
| Privacy Policy | `privacy-policy.html` | Public | QCS privacy policy documentation |
| Protected Area | `protected.html` | Password | Secure access point |
| Service Agreement | `service-agreement.html` | Protected | Legal service agreement |

---

## 🚀 Getting Started

### Local Development

**Run a local server:**

```bash
cd info-web
python3 -m http.server 8000
```

Then open: `http://localhost:8000`

**Or using Node.js:**

```bash
npx http-server -p 8000
```

### Making Changes

1. Edit HTML/CSS files locally
2. Test at `http://localhost:8000`
3. Commit and push to `main` branch
4. GitHub Actions automatically deploys to GitHub Pages

---

## 🔄 Deployment

The site uses **GitHub Actions** for automated deployment:

1. **Trigger**: Push to `main` branch
2. **Build**: GitHub Actions injects the password from `PROTECTED_PASSWORD` secret
3. **Deploy**: Site automatically deployed to `https://info.qcloud.systems`
4. **SSL**: Automatically managed by GitHub Pages

### Workflow File

See `.github/workflows/deploy.yml` for the complete deployment configuration.

---

## 🌐 DNS Configuration

To use a custom domain with GitHub Pages, you must configure DNS at your domain registrar:

### CNAME Record Setup

1. Go to your domain registrar's DNS settings (e.g., GoDaddy, Namecheap, Route 53)
2. Create a new **CNAME record** with:
   - **Name/Host**: `info` (or `@` for root domain)
   - **Target/Value**: `<github_handle>.github.io`
   
   **Example:**
   ```
   Name: info
   Type: CNAME
   Value: qcloud-systems.github.io
   ```

3. DNS changes can take up to **24 hours** to propagate
4. Once propagated, your site will be accessible at `https://info.qcloud.systems`

### GitHub Pages Configuration

- The `CNAME` file in this repository tells GitHub Pages which custom domain to use
- GitHub Pages automatically manages SSL certificates once DNS is configured
- Always keep the `CNAME` file committed to the repository

---

## 🤝 Contributing

Contributions are welcome! To contribute:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/your-feature`)
3. Make your changes and test locally
4. Commit with clear messages
5. Push and open a pull request

**Guidelines:**
- Keep content accurate and professional
- Test changes locally before pushing
- Update this README if adding new features
- Don't commit `.env` files or secrets

---

## 📋 Version History

- **2.0.0** — Complete redesign with grey/blue theme, protected documents, GitHub Actions secrets management
- **1.0.0** — Initial release with basic documentation hub

---

## 📄 License

This project is released under the **MIT License**.  
See the LICENSE file for details.

---

## 🔗 Links

- **Site**: https://info.qcloud.systems
- **GitHub**: https://github.com/qcloud-systems
- **Privacy Policy**: https://info.qcloud.systems/privacy-policy.html

---

## 📬 Contact

For questions or collaboration inquiries, feel free to reach out through your
preferred channels (email, GitHub profile, etc.).