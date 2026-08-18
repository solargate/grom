#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SCRIPT="${ROOT}/scripts/changelog_notes.sh"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

assert_eq() {
  local got="$1" want="$2" msg="$3"
  if [[ "${got}" != "${want}" ]]; then
    echo "got:" >&2
    printf '%s\n' "${got}" >&2
    echo "want:" >&2
    printf '%s\n' "${want}" >&2
    fail "${msg}"
  fi
}

assert_fails() {
  local msg="$1"
  shift
  if "$@" >/dev/null 2>"${TMP}/err"; then
    fail "${msg} (expected non-zero exit)"
  fi
}

cat >"${TMP}/CHANGELOG.md" <<'EOF'
# Changelog

## [Unreleased]

### Google Play

- Side menu: Settings, then About, then Sign out

### Changed

- **UI:** Side menu: Settings, then About, then Sign out (logged in); Settings then About when logged out
- **CI:** Release CI uploads the Android AAB to Google Play closed testing (`alpha`) as a draft

### Fixed

- **Server:** bbolt: the newest workout is no longer listed twice
- **Docs:** Install docs: OpenAPI regen target is `make apidoc`

## [1.2.0] - 2026-01-02

### Changed

- **Server:** only a storage tweak

## [1.1.0] - 2026-01-01

### Added

- **UI:** a feature with no Play section

## [1.0.0] - 2025-12-31

### Google Play


### Added

- **Android:** Health Sync fix but empty Play section

## [0.9.0] - 2025-12-01

### Google Play

EOF
# 501 'a' characters (over Play limit)
python3 -c 'print("- " + "a" * 499)' >>"${TMP}/CHANGELOG.md"
cat >>"${TMP}/CHANGELOG.md" <<'EOF'

### Added

- **UI:** too long for Play

[Unreleased]: https://example.com/unreleased
[1.2.0]: https://example.com/1.2.0
[1.1.0]: https://example.com/1.1.0
[1.0.0]: https://example.com/1.0.0
[0.9.0]: https://example.com/0.9.0
EOF

GOT="$("${SCRIPT}" github "${TMP}/CHANGELOG.md" Unreleased)"
WANT="$(cat <<'EOF'
### Changed

- **UI:** Side menu: Settings, then About, then Sign out (logged in); Settings then About when logged out
- **CI:** Release CI uploads the Android AAB to Google Play closed testing (`alpha`) as a draft

### Fixed

- **Server:** bbolt: the newest workout is no longer listed twice
- **Docs:** Install docs: OpenAPI regen target is `make apidoc`
EOF
)"
assert_eq "${GOT}" "${WANT}" "github notes should omit ### Google Play"

GOT="$("${SCRIPT}" play "${TMP}/CHANGELOG.md" Unreleased)"
WANT="- Side menu: Settings, then About, then Sign out"
assert_eq "${GOT}" "${WANT}" "play notes should use ### Google Play"

GOT="$("${SCRIPT}" play "${TMP}/CHANGELOG.md" 1.2.0)"
assert_eq "${GOT}" "Stability and maintenance improvements" "server-only release should use Play stub"

GOT="$("${SCRIPT}" play "${TMP}/CHANGELOG.md" 1.1.0)"
assert_eq "${GOT}" "Stability and maintenance improvements" "UI-only release should use Play stub"

assert_fails "empty ### Google Play with Android tag should fail" \
  "${SCRIPT}" play "${TMP}/CHANGELOG.md" 1.0.0
grep -q 'Android' "${TMP}/err" || fail "expected Android error for 1.0.0"

assert_fails "Play notes over 500 characters should fail" \
  "${SCRIPT}" play "${TMP}/CHANGELOG.md" 0.9.0
grep -q '500' "${TMP}/err" || fail "expected 500-character error for 0.9.0"

assert_fails "missing version should fail" \
  "${SCRIPT}" github "${TMP}/CHANGELOG.md" 9.9.9

cat >"${TMP}/empty.md" <<'EOF'
# Changelog

## [Unreleased]

## [1.0.0] - 2026-01-01

### Added

- **Server:** hello

## [0.8.0] - 2025-01-01

[Unreleased]: https://example.com/unreleased
[1.0.0]: https://example.com/1.0.0
[0.8.0]: https://example.com/0.8.0
EOF

GOT="$("${SCRIPT}" github "${TMP}/empty.md" Unreleased)"
assert_eq "${GOT}" "" "empty Unreleased github notes should be empty"

GOT="$("${SCRIPT}" play "${TMP}/empty.md" Unreleased)"
assert_eq "${GOT}" "Stability and maintenance improvements" "empty Unreleased should use Play stub"

GOT="$("${SCRIPT}" github "${TMP}/empty.md" 1.0.0)"
assert_eq "${GOT}" $'### Added\n\n- **Server:** hello' "versioned section still extracts"

assert_fails "empty versioned section should fail" \
  "${SCRIPT}" github "${TMP}/empty.md" 0.8.0
assert_fails "empty versioned section should fail for play" \
  "${SCRIPT}" play "${TMP}/empty.md" 0.8.0

# Smoke-test the real changelog. Empty Unreleased is allowed (Play stub / empty GitHub notes).
GOT="$("${SCRIPT}" github "${ROOT}/CHANGELOG.md" Unreleased)"
if grep -q '^### Google Play' <<<"${GOT}"; then
  fail "real CHANGELOG GitHub notes should omit ### Google Play"
fi

PLAY_GOT="$("${SCRIPT}" play "${ROOT}/CHANGELOG.md" Unreleased)"
if printf '%s\n' "${GOT}" | grep -E -q -- '^- \*\*Android:\*\*'; then
  if [[ "${PLAY_GOT}" == "Stability and maintenance improvements" ]]; then
    fail "real CHANGELOG Unreleased should not use the Play stub while it has Android entries"
  fi
fi

echo "OK: changelog_notes tests passed"
