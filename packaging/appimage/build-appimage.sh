#!/bin/bash
set -e

VERSION=${1:-1.1.3}

APPIMAGE_DIR="shminer.AppDir"
APPDIR="$APPIMAGE_DIR"

echo "Building AppImage for v$VERSION..."

mkdir -p "$APPDIR/usr/bin"
mkdir -p "$APPDIR/usr/share/icons/hicolor/256x256/apps"

echo "Downloading SHMiner-linux-amd64..."
curl -sL "https://github.com/OlexiyOdarchuk/sHryvna_miner/releases/download/v${VERSION}/SHMiner-linux-amd64" -o "$APPDIR/usr/bin/shminer"
chmod +x "$APPDIR/usr/bin/shminer"

cat > "$APPDIR/AppRun" << 'APPRUN'
#!/bin/bash
HERE="$(dirname "$(readlink -f "${0}")"
export PATH="$HERE/usr/bin:$PATH"
export LD_LIBRARY_PATH="$HERE/usr/lib:$HERE/usr/lib/x86_64-linux-gnu:$LD_LIBRARY_PATH"
export XDG_DATA_DIRS="$HERE/usr/share:$XDG_DATA_DIRS"
exec "$HERE/usr/bin/shminer" "$@"
APPRUN
chmod +x "$APPDIR/AppRun"

cat > "$APPDIR/shminer.desktop" << 'DESKTOP'
[Desktop Entry]
Name=SHMiner
Exec=shminer
Icon=shminer
Type=Application
Categories=Utility;Finance;
DESKTOP

base64 -d > "$APPDIR/usr/share/icons/hicolor/256x256/apps/shminer.png" << 'PNG'
iVBORw0KGgoAAAANSUhEUgAAAQAAAAEACAYAAABccqhmAAAACXBIWXMAAAsTAAALEwEAmpwYAAAF
8WlUWHRYTUw6Y29tLmFkb2JlLnhtcAAAAAAAPD94cGFja2V0IGJlZ2luPSLvu78iIGlkPSJXNU0w
TXBDZWhpSHpyZVN6TlRjemtjOWQiPz4gPHg6eG1wbWV0YSB4bWxuczp4PSJhZG9iZTpuczpt
ZXRhLyIgeDp4bXB0az0iQWRvYmUgWE1QIENvcmUgNS42LWMxNDUgNzkuMTYzNDk5LCAyMDE4LzA4
LzEzLTE2OjQwOjIyICAgICAgICAiPiA8cmRmOlJERiB4bWxuczpyZGY9Imh0dHA6Ly93d3cu
dzMub3JnLzE5OTkvMDIvMjItcmRmLXN5bnRheC1ucyMiPiA8cmRmOkRlc2NyaXB0aW9uIHJk
ZjphYm91dD0iIi8+IDwvcmRmOlJERj4gPC94OnhtcG1ldGE+IDw/eHBhY2tldCBlbmQ9Inci
Pz7/2wBDAAMCAgMCAgMDAwMEAwMEBQgFBQQEBQoHBwYIDAoMDAsKCwsNDhIQDQ4RDgsLEBYQ
EDA6TCoXFhYRDxEaFB0UHR4fHR8fJyAhIiAhIR8fHy0gKiAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAqICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAqICAgICAgICAgICAgICD/wAARCAAKAAoDASIA
AhEBAxEB/8QAHwAAAQUBAQEBAQEAAAAAAAAAAAECAwQFBgcICQoL/8QAtRAAAgEDAwIEAwUF
BAAAAAABAgMEBREAEiEGMRQiQVFhMhEygZGhscHR8P/EABoBAAICAwEAAAAAAAAAAAAAAAEC
AAEDBQYH/8QANREAAgECBAQFAwQBBQAAAAECAxEEBSExBhIhQRRxE2GBoSIykbHB0f/aAAwD
AQACEQEDEEAf8C0f8HQb+igWCwWCwWCwWCwWCwWCwWCwWCwWCwWCweB4Hg5j/8AQf8AgWCw
WCwWCwWCwWCwWCwWCwWCwWCwWCwWCweB4Hg5j/8AQf8AgWCwWCwWCwWCwWCwaa8PBAgg
ggg0YJI5pNNIJJGJJJJJJJJJJJg0oC8e1o/h4PHtke2R7ZHtke2R7ZHtle2V7ZXtle2V7
ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2
V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2
V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle
2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZX
tle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7
ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2
V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V7ZXtle2V8n/2wBDAqUAACHcKpUqVKpUql
SpVKlSpVKlSpVKlSpVKlSpVKlSpVKlSpVKpAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD/2Q==
PNG

echo "To build AppImage, run:"
echo "  appimagetool $APPIMAGE_DIR shminer-${VERSION}.AppImage"
echo ""
echo "Or use Docker:"
echo "  docker run --rm -v \$(pwd):/output -w /output ghcr.io/ast Sideload/appimage:latest appimagetool $APPIMAGE_DIR shminer-${VERSION}.AppImage"

echo "Done: shminer.AppDir created (run appimagetool to create .AppImage)"