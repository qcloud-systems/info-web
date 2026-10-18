"""Fail the build if a page looks like it contains a credential.

This site and its repository are public. Once a secret is pushed it stays in
git history even if the line is deleted later, so the only safe place to catch
it is before it ships. The patterns mirror the scanner in info-web's notes.go.

To publish an example that trips a pattern on purpose, use an obviously fake
placeholder such as <your-api-key> instead of a realistic value.
"""

import re

from mkdocs.exceptions import PluginError

SECRET_PATTERNS = {
    "AWS access key": re.compile(r"\b(AKIA|ASIA)[0-9A-Z]{16}\b"),
    "private key": re.compile(r"-----BEGIN [A-Z ]*PRIVATE KEY-----"),
    "GitHub token": re.compile(
        r"\b(ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9]{36}\b|\bgithub_pat_[A-Za-z0-9_]{22,}\b"
    ),
    "Slack token": re.compile(r"\bxox[abprs]-[A-Za-z0-9-]{10,}\b"),
    "Stripe secret key": re.compile(r"\b(sk|rk)_live_[A-Za-z0-9]{16,}\b"),
    "Google API key": re.compile(r"\bAIza[0-9A-Za-z_-]{35}\b"),
    "Cloudflare API token": re.compile(r"(?i)\bCF_API_TOKEN\s*[:=]\s*[\"']?[A-Za-z0-9_-]{40}\b"),
    "JWT": re.compile(r"\beyJ[A-Za-z0-9_-]{10,}\.eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}"),
    "password assignment": re.compile(
        r"(?i)\b(password|passwd|pwd|secret|api[_-]?key|token)\s*[:=]\s*[\"']?(?![<$%{])[^\s\"'<>]{6,}"
    ),
    "credentials in a URL": re.compile(r"[a-z][a-z0-9+.-]*://[^/\s:@<]+:[^/\s@<]+@"),
}


def on_page_markdown(markdown, page, config, files, **kwargs):
    for name, pattern in SECRET_PATTERNS.items():
        match = pattern.search(markdown)
        if match:
            line = markdown.count("\n", 0, match.start()) + 1
            raise PluginError(
                f"{page.file.src_uri}:{line} looks like it contains a secret ({name}). "
                "Remove it before publishing: this site and its repo are public."
            )
    return markdown
