#!/usr/bin/env bash
# Extract GitHub release notes or Google Play "What's new" from CHANGELOG.md.
set -euo pipefail

usage() {
  echo "usage: $0 <github|play> <changelog> <version>" >&2
  exit 2
}

if [[ $# -ne 3 ]]; then
  usage
fi

MODE="$1"
CHANGELOG="$2"
VERSION="$3"
PLAY_STUB="Stability and maintenance improvements"
PLAY_MAX_CHARS=500

if [[ "${MODE}" != "github" && "${MODE}" != "play" ]]; then
  usage
fi

if [[ ! -f "${CHANGELOG}" ]]; then
  echo "::error::CHANGELOG file not found: ${CHANGELOG}" >&2
  exit 1
fi

section_found() {
  awk -v ver="${VERSION}" '
    /^## \[/ {
      if ($0 ~ "^## \\[" ver "\\]") { found = 1; exit }
    }
    END { exit found ? 0 : 1 }
  ' "${CHANGELOG}"
}

extract_section() {
  awk -v ver="${VERSION}" '
    /^## \[/ {
      if (in_section) exit
      if ($0 ~ "^## \\[" ver "\\]") in_section = 1
      next
    }
    /^\[[^]]+\]: / { if (in_section) exit }
    in_section { print }
  ' "${CHANGELOG}"
}

emit_play_stub() {
  printf '%s\n' "${PLAY_STUB}"
}

is_unreleased() {
  [[ "${VERSION}" == "Unreleased" ]]
}

trim_blank_lines() {
  awk '
    { lines[++n] = $0 }
    END {
      start = 1
      while (start <= n && lines[start] ~ /^[[:space:]]*$/) start++
      end = n
      while (end >= start && lines[end] ~ /^[[:space:]]*$/) end--
      for (i = start; i <= end; i++) print lines[i]
    }
  '
}

strip_google_play() {
  awk '
    /^### Google Play[[:space:]]*$/ { skip = 1; next }
    /^### / { skip = 0 }
    skip { next }
    { print }
  '
}

extract_google_play() {
  awk '
    /^### Google Play[[:space:]]*$/ { in_play = 1; next }
    /^### / { if (in_play) exit }
    in_play { print }
  '
}

has_android() {
  grep -E -q -- '^- \*\*Android:\*\*'
}

char_count() {
  python3 -c 'import sys; print(len(sys.stdin.read()))'
}

if ! section_found; then
  echo "::error::No CHANGELOG section found for version ${VERSION}" >&2
  exit 1
fi

SECTION="$(extract_section | trim_blank_lines)"
if [[ -z "${SECTION}" ]]; then
  if is_unreleased; then
    case "${MODE}" in
      github) exit 0 ;;
      play) emit_play_stub ;;
    esac
    exit 0
  fi
  echo "::error::CHANGELOG section for version ${VERSION} is empty" >&2
  exit 1
fi

case "${MODE}" in
  github)
    NOTES="$(printf '%s\n' "${SECTION}" | strip_google_play | trim_blank_lines)"
    if [[ -z "${NOTES}" ]]; then
      if is_unreleased; then
        exit 0
      fi
      echo "::error::CHANGELOG section for version ${VERSION} is empty after removing ### Google Play" >&2
      exit 1
    fi
    printf '%s\n' "${NOTES}"
    ;;
  play)
    PLAY_BODY="$(printf '%s\n' "${SECTION}" | extract_google_play | trim_blank_lines)"
    if [[ -n "${PLAY_BODY}" ]]; then
      NOTES="${PLAY_BODY}"
    elif printf '%s\n' "${SECTION}" | has_android; then
      echo "::error::CHANGELOG version ${VERSION} has Android entries but no ### Google Play section (or it is empty)" >&2
      exit 1
    else
      NOTES="${PLAY_STUB}"
    fi

    COUNT="$(printf '%s' "${NOTES}" | char_count)"
    if (( COUNT > PLAY_MAX_CHARS )); then
      echo "::error::Play Store what's new is ${COUNT} characters (max ${PLAY_MAX_CHARS})" >&2
      exit 1
    fi
    printf '%s\n' "${NOTES}"
    ;;
esac
