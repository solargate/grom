#!/usr/bin/env python3
"""Emit site/privacy.html from MkDocs privacy/index.html with root-relative paths.

MkDocs builds privacy at site/privacy/index.html (links use ../…). A plain copy to
site/privacy.html breaks CSS/JS on GitHub Pages; rewrite paths for the site root.
"""

from __future__ import annotations

from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SRC = ROOT / "site" / "privacy" / "index.html"
DST = ROOT / "site" / "privacy.html"


def main() -> None:
    if not SRC.is_file():
        raise SystemExit(f"missing {SRC} (run mkdocs build first)")

    html = SRC.read_text(encoding="utf-8")
    html = html.replace('href="../', 'href="')
    html = html.replace('src="../', 'src="')
    html = html.replace('href=".."', 'href="."')
    html = html.replace('new URL("..",location)', 'new URL(".",location)')
    html = html.replace('"base": ".."', '"base": "."')
    html = html.replace('"search": "../', '"search": "')
    DST.write_text(html, encoding="utf-8")
    print(f"wrote {DST.relative_to(ROOT)}")


if __name__ == "__main__":
    main()
