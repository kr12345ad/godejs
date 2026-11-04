#!/bin/sh
# Install the latest GodeJS release.
#
# Usage:
#   curl -fsSL https://YOUR_PAGES_URL/install.sh | sh
#
# Optional environment:
#   GODE_REPO=owner/repo     GitHub repository to read releases from
#   GODE_VERSION=v0.1.0      Install a specific release instead of latest
#   GODE_PREFIX=$HOME/.local Install under $GODE_PREFIX/bin/gode

set -eu

REPO="${GODE_REPO:-krishnaa/godejs}"
PREFIX="${GODE_PREFIX:-$HOME/.local}"
BIN_DIR="$PREFIX/bin"
VERSION="${GODE_VERSION:-latest}"

say() {
  printf '%s\n' "$*"
}

die() {
  printf 'gode: %s\n' "$*" >&2
  exit 1
}

command -v curl >/dev/null 2>&1 || die "curl is required"
command -v uname >/dev/null 2>&1 || die "uname is required"
command -v mktemp >/dev/null 2>&1 || die "mktemp is required"

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) die "GodeJS release binaries are currently Linux-only" ;;
  *) die "unsupported operating system: $(uname -s)" ;;
esac

case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) die "unsupported architecture: $(uname -m)" ;;
esac

asset="gode-${os}-${arch}"
if [ "$VERSION" = "latest" ]; then
  endpoint="https://api.github.com/repos/${REPO}/releases/latest"
else
  endpoint="https://api.github.com/repos/${REPO}/releases/tags/${VERSION}"
fi

json=$(curl -fsSL -H 'Accept: application/vnd.github+json' "$endpoint") ||
  die "could not query releases for ${REPO}"

tag=""
url=""
sha_url=""
if command -v jq >/dev/null 2>&1; then
  tag=$(printf '%s' "$json" | jq -r '.tag_name // empty')
  url=$(printf '%s' "$json" | jq -r --arg n "$asset" \
    '.assets[]? | select(.name == $n) | .browser_download_url' | head -n 1)
  sha_url=$(printf '%s' "$json" | jq -r --arg n "${asset}.sha256" \
    '.assets[]? | select(.name == $n) | .browser_download_url' | head -n 1)
else
  tag=$(printf '%s' "$json" |
    sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)
  url=$(printf '%s' "$json" | tr ',' '\n' |
    sed -n 's/.*"browser_download_url":[[:space:]]*"\([^"]*\)".*/\1/p' |
    grep -F "/${asset}" | grep -vF '.sha256' | head -n 1 || true)
  sha_url=$(printf '%s' "$json" | tr ',' '\n' |
    sed -n 's/.*"browser_download_url":[[:space:]]*"\([^"]*\)".*/\1/p' |
    grep -F "/${asset}.sha256" | head -n 1 || true)
fi

[ -n "$tag" ] || die "could not determine the latest release tag"
[ -n "$url" ] || die "release ${tag} has no ${asset} asset"

tmp=$(mktemp "${TMPDIR:-/tmp}/gode.XXXXXX")
sha_tmp=""
cleanup() {
  [ -z "$tmp" ] || rm -f "$tmp"
  [ -z "$sha_tmp" ] || rm -f "$sha_tmp"
}
trap cleanup EXIT HUP INT TERM

say "Downloading gode ${tag} (${asset})..."
curl -fsSL "$url" -o "$tmp" || die "could not download ${url}"

if [ -n "$sha_url" ]; then
  sha_tmp=$(mktemp "${TMPDIR:-/tmp}/gode-sha.XXXXXX")
  curl -fsSL "$sha_url" -o "$sha_tmp" || die "could not download ${sha_url}"
  expected=$(awk '{print $1}' "$sha_tmp" | head -n 1)
  if command -v sha256sum >/dev/null 2>&1; then
    actual=$(sha256sum "$tmp" | awk '{print $1}')
  else
    actual=$(shasum -a 256 "$tmp" | awk '{print $1}')
  fi
  [ "$expected" = "$actual" ] || die "checksum mismatch for ${asset}"
fi

chmod 755 "$tmp"
mkdir -p "$BIN_DIR" || die "could not create ${BIN_DIR}"
mv "$tmp" "$BIN_DIR/gode"
tmp=""
chmod 755 "$BIN_DIR/gode"

say "Installed gode ${tag} to ${BIN_DIR}/gode"
if ! "$BIN_DIR/gode" --version >/dev/null 2>&1; then
  say "The binary could not start. Install JavaScriptCore, for example:"
  say "  sudo apt install libjavascriptcoregtk-4.0-18"
fi

case ":$PATH:" in
  *":$BIN_DIR:"*) ;;
  *) say "Add it to PATH: export PATH=\"${BIN_DIR}:\$PATH\"" ;;
esac
