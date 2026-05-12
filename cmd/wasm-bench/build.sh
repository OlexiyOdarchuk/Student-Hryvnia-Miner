#!/usr/bin/env bash
# Build the SHA-256 bench WASM artifact for the ishawyha.dev live demo.
#
#   Output:
#     dist/bench-sha256.wasm     — TinyGo, -opt=z, wasm-opt -Oz
#     dist/wasm_exec_tinygo.js   — TinyGo's glue (shared with piton-runner;
#                                  ship one copy, not both)
#
# Usage:  nix develop -c bash cmd/wasm-bench/build.sh [OUT_DIR]
#
# Note: TinyGo 0.40 caps at Go 1.25; this repo tracks Go 1.26.  We temporarily
# downshift go.mod's go directive for the tinygo step only, then restore it
# on exit (success, failure, or interrupt).

set -euo pipefail
OUT="${1:-dist}"
mkdir -p "$OUT"

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

GOMOD_BACKUP="$(mktemp)"
cp go.mod "$GOMOD_BACKUP"
trap 'mv "$GOMOD_BACKUP" go.mod' EXIT INT TERM
sed -i -E 's/^go 1\.(2[6-9]|[3-9][0-9]).*/go 1.25/' go.mod

echo "==> Building cmd/wasm-bench with TinyGo (-opt=z, no debug)"
tinygo build \
    -target=wasm \
    -no-debug \
    -opt=z \
    -o "$OUT/bench-sha256.wasm" \
    ./cmd/wasm-bench

mv "$GOMOD_BACKUP" go.mod
trap - EXIT INT TERM

echo "==> Optimising with wasm-opt"
wasm-opt -Oz --enable-bulk-memory --enable-nontrapping-float-to-int --enable-sign-ext \
    "$OUT/bench-sha256.wasm" -o "$OUT/bench-sha256.wasm.tmp"
mv "$OUT/bench-sha256.wasm.tmp" "$OUT/bench-sha256.wasm"

echo "==> Copying TinyGo wasm_exec.js"
cp "$(tinygo env TINYGOROOT)/targets/wasm_exec.js" "$OUT/wasm_exec_tinygo.js"

echo
echo "==> Done:"
ls -la "$OUT/"
if command -v brotli >/dev/null 2>&1; then
    sz=$(brotli -q 11 -c "$OUT/bench-sha256.wasm" | wc -c)
    printf "==> Brotli projection: %d bytes\n" "$sz"
fi
