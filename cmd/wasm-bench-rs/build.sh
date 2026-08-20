#!/usr/bin/env bash
# Збирає bench-ядро у WASM для живого демо на ishawyha.dev.
#
#   Вихід:
#     dist/bench-core.wasm  — те саме ядро перебору, що й у десктопному
#                             клієнті, зібране під wasm32 з +simd128
#
# Використання:  cmd/wasm-bench-rs/build.sh [OUT_DIR]
#
# Замінило попередню версію на TinyGo (cmd/wasm-bench, див. історію):
# на тому ж залізі, один потік — 0.78 -> 3.88 MH/s, файл 147 KB -> 23 KB.
set -euo pipefail

OUT="${1:-dist}"
CRATE="$(cd "$(dirname "$0")" && pwd)"

mkdir -p "$OUT"
rustup target add wasm32-unknown-unknown >/dev/null

# simd128 треба вмикати на етапі компіляції: у WASM немає рантайм-детекції
# можливостей CPU, як SHA-NI на десктопі.
RUSTFLAGS="-C target-feature=+simd128" cargo build \
    --release \
    --locked \
    --manifest-path "$CRATE/Cargo.toml" \
    --target wasm32-unknown-unknown

cp "$CRATE/target/wasm32-unknown-unknown/release/wasm_bench.wasm" "$OUT/bench-core.wasm"

if command -v wasm-opt >/dev/null 2>&1; then
    echo "==> wasm-opt"
    wasm-opt -O3 --enable-simd "$OUT/bench-core.wasm" -o "$OUT/bench-core.wasm.tmp"
    mv "$OUT/bench-core.wasm.tmp" "$OUT/bench-core.wasm"
else
    echo "==> wasm-opt не знайдено, пропускаю (не обов'язковий)"
fi

ls -la "$OUT/bench-core.wasm"
