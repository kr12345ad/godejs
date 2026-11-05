#!/bin/sh
# GodeJS installer. This is POSIX sh and runs under BusyBox ash on Alpine.
#
# Usage:
#   curl -fsSL https://kr12345ad.github.io/godejs/install.sh | sh
#   curl -fsSL https://kr12345ad.github.io/godejs/install.sh | sh -s -- --source
#   ./install.sh --source
#   ./install.sh --force
#
# Optional environment:
#   GODE_REPO=owner/repo      GitHub repository to read releases from
#   GODE_VERSION=v0.2.1       Install a specific release instead of latest
#   GODE_PREFIX=$HOME/.local  Install under $GODE_PREFIX/bin/gode
#   GODE_FROM_SOURCE=1        Build from source instead of downloading a release
#   GODE_REF=main             Branch or tag used by --source

set -eu

REPO="${GODE_REPO:-kr12345ad/godejs}"
PREFIX="${GODE_PREFIX:-$HOME/.local}"
BIN_DIR="$PREFIX/bin"
RELEASE="${GODE_VERSION:-latest}"
REF="${GODE_REF:-main}"
MODE="release"
FORCE=0
if [ "${GODE_FROM_SOURCE:-0}" = "1" ]; then
  MODE="source"
fi

usage() {
  cat <<'EOF'
Install GodeJS from a GitHub release, or build it from source.

Usage: install.sh [--source] [--force]

  --source    clone the repository and build with the local Go toolchain
  --force     reinstall even when the same release is present, or downgrade
  -h, --help  show this help
EOF
}

for arg in "$@"; do
  case "$arg" in
    --source) MODE="source" ;;
    --force) FORCE=1 ;;
    -h|--help) usage; exit 0 ;;
    *) printf 'gode: unknown option: %s\n' "$arg" >&2; exit 2 ;;
  esac
done

say() {
  printf '%s\n' "$*"
}

die() {
  printf 'gode: %s\n' "$*" >&2
  exit 1
}

need() {
  command -v "$1" >/dev/null 2>&1 || die "$1 is required"
}

is_release_version() {
  case "$1" in
    v[0-9]*.[0-9]*.[0-9]*|[0-9]*.[0-9]*.[0-9]*) return 0 ;;
    *) return 1 ;;
  esac
}

version_compare() {
  awk -v a="$1" -v b="$2" 'BEGIN {
    sub(/^v/, "", a); sub(/^v/, "", b)
    sub(/-.*/, "", a); sub(/-.*/, "", b)
    na = split(a, av, "."); nb = split(b, bv, ".")
    n = na > nb ? na : nb
    for (i = 1; i <= n; i++) {
      x = (i <= na ? av[i] : 0) + 0
      y = (i <= nb ? bv[i] : 0) + 0
      if (x > y) { print "gt"; exit }
      if (x < y) { print "lt"; exit }
    }
    print "eq"
  }'
}

os_id=""
os_like=""
if [ -r /etc/os-release ]; then
  # shellcheck disable=SC1091
  . /etc/os-release
  os_id="${ID:-}"
  os_like="${ID_LIKE:-}"
fi

jsc_runtime_hint() {
  if [ "$(uname -s)" = "Darwin" ]; then
    say "gode needs macOS 12 or later; JavaScriptCore is part of the system."
    return
  fi
  say "JavaScriptCore GTK 4.0 is required to run gode. Install it:"
  case " $os_id $os_like " in
    *" alpine "*)
      say "  sudo apk add webkit2gtk-4.0"
      ;;
    *" debian "*|*" ubuntu "*)
      say "  sudo apt install libjavascriptcoregtk-4.0-18"
      ;;
    *" fedora "*)
      say "  sudo dnf install webkit2gtk4.0"
      ;;
    *)
      say "  Debian/Ubuntu: sudo apt install libjavascriptcoregtk-4.0-18"
      say "  Alpine:        sudo apk add webkit2gtk-4.0"
      ;;
  esac
  say "The binary links libjavascriptcoregtk-4.0.so.18."
}

jsc_dev_hint() {
  if [ "$(uname -s)" = "Darwin" ]; then
    say "Building gode needs the Xcode Command Line Tools:"
    say "  xcode-select --install"
    return
  fi
  say "Building gode needs the JavaScriptCore GTK 4.0 development headers:"
  case " $os_id $os_like " in
    *" alpine "*)
      say "  sudo apk add webkit2gtk-4.0-dev go gcc musl-dev"
      ;;
    *" debian "*|*" ubuntu "*)
      say "  sudo apt install libjavascriptcoregtk-4.0-dev golang gcc"
      ;;
    *" fedora "*)
      say "  sudo dnf install webkit2gtk4.0-devel golang gcc"
      ;;
    *)
      say "  Debian/Ubuntu: sudo apt install libjavascriptcoregtk-4.0-dev golang gcc"
      say "  Alpine:        sudo apk add webkit2gtk-4.0-dev go gcc musl-dev"
      ;;
  esac
}

tmp=""
tmpdir=""
cleanup() {
  [ -z "$tmp" ] || rm -f "$tmp"
  [ -z "$tmpdir" ] || rm -rf "$tmpdir"
}
trap cleanup EXIT HUP INT TERM

install_bin() {
  src="$1"
  label="$2"
  release="$3"
  chmod 755 "$src"
  mkdir -p "$BIN_DIR" || die "could not create $BIN_DIR"
  mv "$src" "$BIN_DIR/gode" || die "could not install to $BIN_DIR/gode"
  chmod 755 "$BIN_DIR/gode"
  printf '%s\n' "$release" > "$BIN_DIR/.gode-version" ||
    die "could not record the installed release"
  say "Installed gode ${label} to ${BIN_DIR}/gode"
  if ! "$BIN_DIR/gode" --version >/dev/null 2>&1; then
    jsc_runtime_hint
  fi
  case ":$PATH:" in
    *":$BIN_DIR:"*) ;;
    *) say "Add it to PATH: export PATH=\"${BIN_DIR}:\$PATH\"" ;;
  esac
}

if [ "$MODE" = "source" ]; then
  need git
  need go
  need cc
  need mktemp
  if [ -e "$BIN_DIR/gode" ]; then
    say "Updating the existing gode install from source (${REF})..."
  else
    say "Installing gode from source (${REF})..."
  fi
  tmpdir=$(mktemp -d "${TMPDIR:-/tmp}/gode-src.XXXXXX")
  say "Cloning ${REPO} (${REF})..."
  git clone --depth 1 --branch "$REF" "https://github.com/${REPO}.git" "$tmpdir/src" ||
    die "could not clone ${REPO}"
  say "Building gode with JavaScriptCore..."
  (cd "$tmpdir/src" && CGO_ENABLED=1 go build -trimpath -o "$tmpdir/gode" ./cmd/gode) || {
    jsc_dev_hint
    die "source build failed"
  }
  install_bin "$tmpdir/gode" "from ${REF}" "source:${REF}"
  rm -rf "$tmpdir"
  tmpdir=""
  exit 0
fi

need curl
need uname
need mktemp
need awk

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *) die "unsupported operating system: $(uname -s)" ;;
esac

case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) die "unsupported architecture: $(uname -m)" ;;
esac

asset="gode-${os}-${arch}"
if [ "$RELEASE" = "latest" ]; then
  latest_url=$(curl --http1.1 --retry 4 --retry-delay 2 -fsSIL -o /dev/null \
    -w '%{url_effective}' "https://github.com/${REPO}/releases/latest") ||
    die "could not determine the latest release"
  case "$latest_url" in
    */tag/*) tag=${latest_url##*/tag/} ;;
    *) die "could not determine the latest release tag" ;;
  esac
else
  tag="$RELEASE"
  case "$tag" in
    v*) ;;
    *) tag="v${tag}" ;;
  esac
fi
base="https://github.com/${REPO}/releases/download/${tag}"
url="${base}/${asset}"
sha_url="${url}.sha256"

if [ -e "$BIN_DIR/gode" ] && [ "$FORCE" = 0 ]; then
  existing_version=""
  if [ -f "$BIN_DIR/.gode-version" ]; then
    IFS= read -r existing_version < "$BIN_DIR/.gode-version" || true
  fi
  case "$existing_version" in
    source:*)
      say "Existing gode was built from source (${existing_version#source:})."
      say "Leaving it unchanged. Use --force to replace it with ${tag}."
      exit 0
      ;;
  esac
  if is_release_version "$existing_version"; then
    case "$(version_compare "$tag" "$existing_version")" in
      gt)
        say "Updating gode ${existing_version} -> ${tag}..."
        ;;
      eq)
        say "gode ${tag} is already installed at ${BIN_DIR}/gode."
        say "Use --force to reinstall it."
        exit 0
        ;;
      lt)
        say "Warning: gode ${existing_version} is newer than ${tag}."
        say "Leaving the existing installation unchanged."
        say "Use --force to downgrade to ${tag}."
        exit 0
        ;;
    esac
  else
    say "Existing gode install found at ${BIN_DIR}/gode (release version unknown)."
    say "Updating it to ${tag}..."
  fi
elif [ -e "$BIN_DIR/gode" ]; then
  say "Replacing the existing gode install with ${tag}..."
else
  say "Installing gode ${tag}..."
fi

tmp=$(mktemp "${TMPDIR:-/tmp}/gode.XXXXXX")
sha_tmp=""
cleanup_release() {
  [ -z "$tmp" ] || rm -f "$tmp"
  [ -z "$sha_tmp" ] || rm -f "$sha_tmp"
}
trap cleanup_release EXIT HUP INT TERM

say "Downloading gode ${tag} (${asset})..."
curl --http1.1 --retry 4 --retry-delay 2 -fsSL "$url" -o "$tmp" || die "could not download ${url}"

if [ -n "$sha_url" ]; then
  sha_tmp=$(mktemp "${TMPDIR:-/tmp}/gode-sha.XXXXXX")
  curl --http1.1 --retry 4 --retry-delay 2 -fsSL "$sha_url" -o "$sha_tmp" || die "could not download ${sha_url}"
  expected=$(awk '{print $1}' "$sha_tmp" | head -n 1)
  if command -v sha256sum >/dev/null 2>&1; then
    actual=$(sha256sum "$tmp" | awk '{print $1}')
  else
    actual=$(shasum -a 256 "$tmp" | awk '{print $1}')
  fi
  [ "$expected" = "$actual" ] || die "checksum mismatch for ${asset}"
fi

install_bin "$tmp" "$tag" "$tag"
tmp=""
if [ -n "$sha_tmp" ]; then
  rm -f "$sha_tmp"
  sha_tmp=""
fi
