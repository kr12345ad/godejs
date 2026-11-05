#!/bin/sh
# Build an unsigned macOS installer for GodeJS. Run this on macOS.
#
# Usage:
#   packaging/macos/build-pkg.sh <binary> <version> <arch> [output-dir]
#
# <arch> is arm64 or amd64. The pkg contains /usr/local/bin/gode and the
# project licenses under /usr/local/share/doc/godejs.

set -eu

if [ "$#" -lt 3 ]; then
  printf 'usage: %s <binary> <version> <arm64|amd64> [output-dir]\n' "$0" >&2
  exit 2
fi

BIN=$1
VERSION=${2#v}
ARCH=$3
OUTDIR=${4:-dist}
ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)

[ -f "$BIN" ] || { printf 'gode: binary not found: %s\n' "$BIN" >&2; exit 1; }
[ -n "$VERSION" ] || VERSION=0.0.0

case "$ARCH" in
  arm64) HOST_ARCH=arm64 ;;
  amd64|x86_64) ARCH=amd64; HOST_ARCH=x86_64 ;;
  *) printf 'gode: unsupported architecture: %s\n' "$ARCH" >&2; exit 2 ;;
esac

for tool in pkgbuild productbuild shasum; do
  command -v "$tool" >/dev/null 2>&1 || {
    printf 'gode: %s is required (run this on macOS)\n' "$tool" >&2
    exit 1
  }
done

mkdir -p "$OUTDIR"
OUTDIR=$(CDPATH= cd -- "$OUTDIR" && pwd)
NAME="gode-darwin-${ARCH}.pkg"
WORK=$(mktemp -d "${TMPDIR:-/tmp}/gode-pkg.XXXXXX")
trap 'rm -rf "$WORK"' EXIT HUP INT TERM

ROOT="$WORK/root"
SCRIPTS="$WORK/scripts"
RESOURCES="$WORK/resources"
mkdir -p "$ROOT/usr/local/bin" "$ROOT/usr/local/share/doc/godejs" "$SCRIPTS" "$RESOURCES"

install -m 0755 "$BIN" "$ROOT/usr/local/bin/gode"
[ -f "$ROOT_DIR/LICENSE" ] && install -m 0644 "$ROOT_DIR/LICENSE" "$ROOT/usr/local/share/doc/godejs/LICENSE"
[ -f "$ROOT_DIR/NOTICE" ] && install -m 0644 "$ROOT_DIR/NOTICE" "$ROOT/usr/local/share/doc/godejs/NOTICE"

cat > "$SCRIPTS/postinstall" <<'EOF'
#!/bin/sh
set -eu
chmod 755 /usr/local/bin/gode
exit 0
EOF
chmod 755 "$SCRIPTS/postinstall"

cat > "$RESOURCES/welcome.html" <<EOF
<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><style>
body{font:14px/1.5 -apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;color:#1f2937;margin:18px}
h1{font-size:24px;margin:0 0 8px}code{background:#f3f4f6;padding:2px 5px;border-radius:4px}
p{margin:8px 0}.muted{color:#6b7280}
</style></head>
<body>
<h1>GodeJS ${VERSION}</h1>
<p>This installer places the <code>gode</code> runtime in <code>/usr/local/bin</code>.</p>
<p class="muted">Requires macOS 12 or later. JavaScriptCore is provided by macOS.</p>
</body>
</html>
EOF

cat > "$RESOURCES/conclusion.html" <<'EOF'
<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><style>
body{font:14px/1.5 -apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;color:#1f2937;margin:18px}
h1{font-size:24px;margin:0 0 8px}code{background:#f3f4f6;padding:2px 5px;border-radius:4px}
p{margin:8px 0}
</style></head>
<body>
<h1>GodeJS is installed</h1>
<p>Open Terminal and run <code>gode --version</code> to verify the install.</p>
<p>If your shell does not find it, add <code>/usr/local/bin</code> to <code>PATH</code>.</p>
</body>
</html>
EOF

if [ -n "${GODE_CODESIGN_IDENTITY:-}" ] && command -v codesign >/dev/null 2>&1; then
  codesign --force --options runtime --timestamp \
    --sign "$GODE_CODESIGN_IDENTITY" "$ROOT/usr/local/bin/gode"
fi

pkgbuild \
  --root "$ROOT" \
  --identifier "dev.godejs.gode" \
  --version "$VERSION" \
  --install-location "/" \
  --ownership recommended \
  --scripts "$SCRIPTS" \
  "$WORK/gode.pkg"

cat > "$WORK/distribution.xml" <<EOF
<?xml version="1.0" encoding="utf-8"?>
<installer-gui-script minSpecVersion="2">
  <title>GodeJS ${VERSION}</title>
  <organization>dev.godejs</organization>
  <options customize="never" require-scripts="false" hostArchitectures="${HOST_ARCH}"/>
  <welcome file="welcome.html"/>
  <conclusion file="conclusion.html"/>
  <pkg-ref id="dev.godejs.gode"/>
  <choices-outline>
    <line choice="default"><line choice="gode"/></line>
  </choices-outline>
  <choice id="default"/>
  <choice id="gode" visible="false">
    <pkg-ref id="dev.godejs.gode"/>
  </choice>
  <pkg-ref id="dev.godejs.gode" version="${VERSION}" onConclusion="none">gode.pkg</pkg-ref>
</installer-gui-script>
EOF

if [ -n "${GODE_INSTALLER_SIGN_IDENTITY:-}" ]; then
  productbuild \
    --distribution "$WORK/distribution.xml" \
    --resources "$RESOURCES" \
    --package-path "$WORK" \
    --sign "$GODE_INSTALLER_SIGN_IDENTITY" \
    --timestamp \
    "$OUTDIR/$NAME"
else
  productbuild \
    --distribution "$WORK/distribution.xml" \
    --resources "$RESOURCES" \
    --package-path "$WORK" \
    "$OUTDIR/$NAME"
fi

shasum -a 256 "$OUTDIR/$NAME" > "$OUTDIR/$NAME.sha256"
printf 'Built %s\n' "$OUTDIR/$NAME"
