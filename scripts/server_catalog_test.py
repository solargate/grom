#!/usr/bin/env python3
"""Tests for scripts/server_catalog.py."""

from __future__ import annotations

import os
import tempfile
import unittest
from contextlib import redirect_stderr
from io import StringIO
from pathlib import Path
from unittest.mock import patch

import server_catalog as catalog


def _run_main(argv: list[str]) -> tuple[int, str]:
    buf = StringIO()
    with redirect_stderr(buf):
        code = catalog.main(argv)
    return code, buf.getvalue()


def _write_yaml(directory: Path, body: str, name: str = "server-catalog.yaml") -> Path:
    path = directory / name
    path.write_text(body, encoding="utf-8")
    return path


class ServerCatalogTest(unittest.TestCase):
    def test_empty_servers_list_is_valid(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            path = _write_yaml(Path(tmp), "servers: []\n")
            self.assertEqual(catalog.load_catalog(path), [])

    def test_valid_entry_normalizes_path_and_host(self) -> None:
        body = """
servers:
  - url: https://Example.ORG/grom/
    name: Example
    description: A public instance
    email: admin@example.org
"""
        with tempfile.TemporaryDirectory() as tmp:
            path = _write_yaml(Path(tmp), body)
            entries = catalog.load_catalog(path)
        self.assertEqual(len(entries), 1)
        self.assertEqual(entries[0].url, "https://example.org/grom")
        self.assertEqual(entries[0].name, "Example")
        self.assertEqual(entries[0].email, "admin@example.org")

    def test_rejects_http(self) -> None:
        body = """
servers:
  - url: http://example.org
    name: Example
    description: Nope
    email: admin@example.org
"""
        with tempfile.TemporaryDirectory() as tmp:
            path = _write_yaml(Path(tmp), body)
            with self.assertRaisesRegex(catalog.CatalogError, "https://"):
                catalog.load_catalog(path)

    def test_rejects_port(self) -> None:
        body = """
servers:
  - url: https://example.org:8443
    name: Example
    description: Nope
    email: admin@example.org
"""
        with tempfile.TemporaryDirectory() as tmp:
            path = _write_yaml(Path(tmp), body)
            with self.assertRaisesRegex(catalog.CatalogError, "port"):
                catalog.load_catalog(path)

    def test_rejects_explicit_https_port(self) -> None:
        body = """
servers:
  - url: https://example.org:443
    name: Example
    description: Nope
    email: admin@example.org
"""
        with tempfile.TemporaryDirectory() as tmp:
            path = _write_yaml(Path(tmp), body)
            with self.assertRaisesRegex(catalog.CatalogError, "port"):
                catalog.load_catalog(path)

    def test_rejects_ip_address(self) -> None:
        body = """
servers:
  - url: https://127.0.0.1
    name: Example
    description: Nope
    email: admin@example.org
"""
        with tempfile.TemporaryDirectory() as tmp:
            path = _write_yaml(Path(tmp), body)
            with self.assertRaisesRegex(catalog.CatalogError, "IP"):
                catalog.load_catalog(path)

    def test_rejects_duplicate_urls(self) -> None:
        body = """
servers:
  - url: https://example.org/grom/
    name: One
    description: First
    email: a@example.org
  - url: https://example.org/grom
    name: Two
    description: Second
    email: b@example.org
"""
        with tempfile.TemporaryDirectory() as tmp:
            path = _write_yaml(Path(tmp), body)
            with self.assertRaisesRegex(catalog.CatalogError, "duplicate"):
                catalog.load_catalog(path)

    def test_rejects_missing_email_at(self) -> None:
        body = """
servers:
  - url: https://example.org
    name: Example
    description: Public instance
    email: not-an-email
"""
        with tempfile.TemporaryDirectory() as tmp:
            path = _write_yaml(Path(tmp), body)
            with self.assertRaisesRegex(catalog.CatalogError, "email"):
                catalog.load_catalog(path)

    def test_rejects_query_and_userinfo(self) -> None:
        query = """
servers:
  - url: https://example.org/?x=1
    name: Example
    description: Nope
    email: admin@example.org
"""
        userinfo = """
servers:
  - url: https://user:pass@example.org
    name: Example
    description: Nope
    email: admin@example.org
"""
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            with self.assertRaisesRegex(catalog.CatalogError, "query"):
                catalog.load_catalog(_write_yaml(root, query, "query.yaml"))
            with self.assertRaisesRegex(catalog.CatalogError, "userinfo"):
                catalog.load_catalog(_write_yaml(root, userinfo, "userinfo.yaml"))

    def test_generate_dart_escapes_dollar_and_omits_email(self) -> None:
        entries = [
            catalog.CatalogEntry(
                url="https://example.org",
                name="Cost $0",
                description="Public $instance",
                email="admin@example.org",
            )
        ]
        source = catalog.generate_dart(entries)
        self.assertIn(r"Cost \$0", source)
        self.assertIn(r"Public \$instance", source)
        self.assertNotIn("admin@example.org", source)
        self.assertIn("part of '../server_catalog.dart';", source)

    def test_generate_empty_list(self) -> None:
        source = catalog.generate_dart([])
        self.assertIn("const List<CatalogServer> kApprovedServers = [];", source)

    def test_cli_generate_writes_output(self) -> None:
        body = """
servers:
  - url: https://example.org
    name: Example
    description: Public instance
    email: admin@example.org
"""
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            catalog_path = _write_yaml(root, body)
            output = root / "server_catalog.g.dart"
            code, stderr = _run_main(
                [
                    "generate",
                    "--catalog",
                    str(catalog_path),
                    "--output",
                    str(output),
                ]
            )
            self.assertEqual(code, 0)
            self.assertIn(str(output), stderr)
            text = output.read_text(encoding="utf-8")
            self.assertIn("https://example.org", text)
            self.assertNotIn("admin@example.org", text)

    def test_cli_validate_fails_on_http(self) -> None:
        body = """
servers:
  - url: http://example.org
    name: Example
    description: Nope
    email: admin@example.org
"""
        with tempfile.TemporaryDirectory() as tmp:
            catalog_path = _write_yaml(Path(tmp), body)
            with patch.dict(os.environ, {"GITHUB_ACTIONS": ""}):
                code, stderr = _run_main(["validate", "--catalog", str(catalog_path)])
            self.assertEqual(code, 1)
            self.assertIn("https://", stderr)
            self.assertNotIn("::error::", stderr)

    def test_cli_validate_emits_github_error_in_actions(self) -> None:
        body = """
servers:
  - url: http://example.org
    name: Example
    description: Nope
    email: admin@example.org
"""
        with tempfile.TemporaryDirectory() as tmp:
            catalog_path = _write_yaml(Path(tmp), body)
            buf = StringIO()
            with patch.dict(os.environ, {"GITHUB_ACTIONS": "true"}):
                with redirect_stderr(buf):
                    code = catalog.main(["validate", "--catalog", str(catalog_path)])
            self.assertEqual(code, 1)
            self.assertIn("::error::", buf.getvalue())

    def test_repo_catalog_is_valid(self) -> None:
        entries = catalog.load_catalog(catalog.DEFAULT_CATALOG)
        self.assertIsInstance(entries, list)


if __name__ == "__main__":
    unittest.main()
