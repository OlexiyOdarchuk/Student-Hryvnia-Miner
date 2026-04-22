#!/bin/bash
set -e

VERSION=${1:-1.1.3}
PACKAGE_DIR="shminer-${VERSION}"

mkdir -p "$PACKAGE_DIR/usr/bin"
mkdir -p "$PACKAGE_DIR/usr/share/doc/shminer"
mkdir -p "$PACKAGE_DIR/usr/share/licenses/shminer"

echo "Downloading SHMiner-linux-amd64..."
curl -sL "https://github.com/OlexiyOdarchuk/Student-Hryvnia-Miner/releases/download/v${VERSION}/SHMiner-linux-amd64" -o "$PACKAGE_DIR/usr/bin/shminer"
chmod +x "$PACKAGE_DIR/usr/share/bin/shminer"

echo "Creating DEB package..."
dpkg-deb --build --compression=gzip "$PACKAGE_DIR" "shminer_${VERSION}_amd64.deb"

rm -rf "$PACKAGE_DIR"

echo "Done: shminer_${VERSION}_amd64.deb"