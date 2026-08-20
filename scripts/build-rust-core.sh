#!/usr/bin/env bash
# Збирає rust_miner_core і кладе статичну бібліотеку туди, де її очікує cgo
# (backend/internal/miner/rust_miner_core/lib/<GOOS>_<GOARCH>/).
#
# Таргет визначається з GOOS/GOARCH або задається явно через RUST_TARGET.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CRATE_DIR="$REPO_ROOT/backend/internal/miner/rust_miner_core"

GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"
RUST_TARGET="${RUST_TARGET:-}"

if [ -z "$RUST_TARGET" ]; then
  case "$GOOS/$GOARCH" in
    linux/amd64)   RUST_TARGET=x86_64-unknown-linux-gnu ;;
    linux/arm64)   RUST_TARGET=aarch64-unknown-linux-gnu ;;
    darwin/amd64)  RUST_TARGET=x86_64-apple-darwin ;;
    darwin/arm64)  RUST_TARGET=aarch64-apple-darwin ;;
    windows/amd64) RUST_TARGET=x86_64-pc-windows-gnu ;;
    *)
      echo "rust-core: невідомий Rust-таргет для $GOOS/$GOARCH" >&2
      echo "rust-core: задайте його явно через RUST_TARGET=..." >&2
      exit 1
      ;;
  esac
fi

OUT_DIR="$CRATE_DIR/lib/${GOOS}_${GOARCH}"

rustup target add "$RUST_TARGET" >/dev/null

cargo build \
  --release \
  --locked \
  --manifest-path "$CRATE_DIR/Cargo.toml" \
  --target "$RUST_TARGET"

mkdir -p "$OUT_DIR"
cp "$CRATE_DIR/target/$RUST_TARGET/release/librust_miner_core.a" "$OUT_DIR/librust_miner_core.a"

echo "rust-core: $RUST_TARGET -> lib/${GOOS}_${GOARCH}/librust_miner_core.a"
