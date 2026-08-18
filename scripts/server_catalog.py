#!/usr/bin/env python3
"""Validate server-catalog.yaml and generate the Flutter catalog Dart file."""

from __future__ import annotations

import argparse
import ipaddress
import json
import sys
from dataclasses import dataclass
from pathlib import Path
from urllib.parse import urlparse

try:
    import yaml
except ImportError:
    sys.stderr.write(
        "PyYAML is required. Run `make venv` (or `pip install -r requirements.txt`) "
        "and invoke this script with that interpreter.\n"
    )
    sys.exit(2)

REPO_ROOT = Path(__file__).resolve().parent.parent
DEFAULT_CATALOG = REPO_ROOT / "server-catalog.yaml"
DEFAULT_OUTPUT = REPO_ROOT / "ui" / "grom" / "lib" / "generated" / "server_catalog.g.dart"

MAX_NAME_LEN = 80
MAX_DESCRIPTION_LEN = 280
ALLOWED_KEYS = frozenset({"url", "name", "description", "email"})


class CatalogError(Exception):
    pass


@dataclass(frozen=True)
class CatalogEntry:
    url: str
    name: str
    description: str
    email: str


def dart_string(value: str) -> str:
    """JSON-quote a string and escape `$` so Dart does not interpolate it."""
    return json.dumps(value, ensure_ascii=False).replace("$", r"\$")


def _require_str(value: object, field: str, index: int) -> str:
    if not isinstance(value, str):
        raise CatalogError(f"servers[{index}].{field} must be a string")
    stripped = value.strip()
    if not stripped:
        raise CatalogError(f"servers[{index}].{field} must be a non-empty string")
    return stripped


def normalize_catalog_url(url: str) -> str:
    parsed = urlparse(url.strip())
    host = (parsed.hostname or "").rstrip(".").lower()
    path = parsed.path.rstrip("/")
    return f"https://{host}{path}"


def validate_url(url: str, index: int) -> str:
    if not url.startswith("https://"):
        raise CatalogError(
            f"servers[{index}].url must start with https:// (got {url!r})"
        )

    parsed = urlparse(url)
    if parsed.scheme != "https":
        raise CatalogError(
            f"servers[{index}].url must use the https scheme (got {parsed.scheme!r})"
        )
    if parsed.username is not None or parsed.password is not None:
        raise CatalogError(f"servers[{index}].url must not include userinfo")
    if parsed.query or parsed.fragment:
        raise CatalogError(f"servers[{index}].url must not include a query or fragment")
    if parsed.port is not None:
        raise CatalogError(
            f"servers[{index}].url must not include a port (got {parsed.port})"
        )

    host = (parsed.hostname or "").rstrip(".")
    if not host:
        raise CatalogError(f"servers[{index}].url must include a hostname")
    try:
        ipaddress.ip_address(host)
        raise CatalogError(
            f"servers[{index}].url hostname must be a DNS name, not an IP address"
        )
    except ValueError:
        pass
    if "." not in host:
        raise CatalogError(
            f"servers[{index}].url hostname must be a DNS name with a dot"
        )

    return normalize_catalog_url(url)


def load_catalog(path: Path) -> list[CatalogEntry]:
    if not path.is_file():
        raise CatalogError(f"catalog file not found: {path}")

    try:
        raw = yaml.safe_load(path.read_text(encoding="utf-8"))
    except yaml.YAMLError as exc:
        raise CatalogError(f"invalid YAML: {exc}") from exc

    if raw is None:
        raise CatalogError("catalog is empty")
    if not isinstance(raw, dict):
        raise CatalogError("catalog root must be a mapping with a servers key")

    extra_root = set(raw) - {"servers"}
    if extra_root:
        keys = ", ".join(sorted(extra_root))
        raise CatalogError(f"unexpected top-level keys: {keys}")
    if "servers" not in raw:
        raise CatalogError("catalog must have a servers key")

    servers = raw["servers"]
    if servers is None:
        raise CatalogError("servers must be a list (use [] when empty)")
    if not isinstance(servers, list):
        raise CatalogError("servers must be a list")

    entries: list[CatalogEntry] = []
    seen_urls: dict[str, int] = {}
    for index, item in enumerate(servers):
        if not isinstance(item, dict):
            raise CatalogError(f"servers[{index}] must be a mapping")
        extra = set(item) - ALLOWED_KEYS
        if extra:
            keys = ", ".join(sorted(str(k) for k in extra))
            raise CatalogError(f"servers[{index}] has unexpected keys: {keys}")
        missing = ALLOWED_KEYS - set(item)
        if missing:
            keys = ", ".join(sorted(missing))
            raise CatalogError(f"servers[{index}] missing keys: {keys}")

        url = validate_url(_require_str(item.get("url"), "url", index), index)
        name = _require_str(item.get("name"), "name", index)
        description = _require_str(item.get("description"), "description", index)
        email = _require_str(item.get("email"), "email", index)

        if len(name) > MAX_NAME_LEN:
            raise CatalogError(
                f"servers[{index}].name must be at most {MAX_NAME_LEN} characters"
            )
        if len(description) > MAX_DESCRIPTION_LEN:
            raise CatalogError(
                f"servers[{index}].description must be at most {MAX_DESCRIPTION_LEN} characters"
            )
        if "@" not in email:
            raise CatalogError(f"servers[{index}].email must contain @")

        previous = seen_urls.get(url)
        if previous is not None:
            raise CatalogError(
                f"duplicate url {url!r} at servers[{index}] (same as servers[{previous}])"
            )
        seen_urls[url] = index
        entries.append(
            CatalogEntry(url=url, name=name, description=description, email=email)
        )

    return entries


def generate_dart(entries: list[CatalogEntry]) -> str:
    lines = [
        "// Code generated by scripts/server_catalog.py. DO NOT EDIT.",
        "// Source: server-catalog.yaml",
        "",
        "part of '../server_catalog.dart';",
        "",
    ]
    if not entries:
        lines.append("const List<CatalogServer> kApprovedServers = [];")
    else:
        lines.append("const List<CatalogServer> kApprovedServers = [")
        for entry in entries:
            lines.extend(
                [
                    "  CatalogServer(",
                    f"    url: {dart_string(entry.url)},",
                    f"    name: {dart_string(entry.name)},",
                    f"    description: {dart_string(entry.description)},",
                    "  ),",
                ]
            )
        lines.append("];")
    lines.append("")
    return "\n".join(lines)


def write_dart(path: Path, source: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(source, encoding="utf-8")


def _emit_error(message: str) -> None:
    print(f"::error::{message}", file=sys.stderr)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(
        description="Validate server-catalog.yaml and generate Flutter Dart."
    )
    parser.add_argument(
        "command",
        choices=("validate", "generate"),
        help="validate the YAML, or validate and write the Dart catalog",
    )
    parser.add_argument(
        "--catalog",
        type=Path,
        default=DEFAULT_CATALOG,
        help="path to server-catalog.yaml",
    )
    parser.add_argument(
        "--output",
        type=Path,
        default=DEFAULT_OUTPUT,
        help="path to generated server_catalog.g.dart",
    )
    args = parser.parse_args(argv)

    try:
        entries = load_catalog(args.catalog)
    except CatalogError as exc:
        _emit_error(str(exc))
        return 1

    if args.command == "generate":
        write_dart(args.output, generate_dart(entries))
        print(f"Wrote {args.output} ({len(entries)} server(s))", file=sys.stderr)

    return 0


if __name__ == "__main__":
    sys.exit(main())
