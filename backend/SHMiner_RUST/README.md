# Rust Miner

Lightweight Rust translation of the backend miner. It keeps the SHA-256 loop, difficulty check, wallet rotation, and block submission logic but runs as a standalone CLI so you can benchmark Rust performance against the existing Go implementation.

## Building

```bash
cargo build --release --manifest-path rust_miner/Cargo.toml
```

## Running

### Basic usage

```bash
SH_MINER_SERVER=https://s-hryvnia-1.onrender.com \
SH_MINER_WALLETS="wallet1,wallet2" \
cargo run --release --manifest-path rust_miner/Cargo.toml
```

The miner will:

1. Pull the latest block hash from `SH_MINER_SERVER/chain`.
2. Spin up one worker per `SH_MINER_THREADS` (or CPU count when unset).
3. Hash `prevHash + nonce + wallet + "1" + timestamp` until the difficulty bits (default: 20) are satisfied.
4. Submit the winning block to `SH_MINER_SERVER/submit-block`.

### Environment variables

| Variable | Default | Description |
| --- | --- | --- |
| `SH_MINER_SERVER` | `https://s-hryvnia-1.onrender.com` | Base URL for `/chain` and `/submit-block`. |
| `SH_MINER_DIFFICULTY` | `20` | Number of leading zero bits required by the proof-of-work. |
| `SH_MINER_THREADS` | CPU count | Worker threads for hashing. |
| `SH_MINER_WALLETS` | (default wallet in code) | Comma-separated list of wallets to rotate through. |
| `SH_MINER_HTTP_TIMEOUT` | `5` | Timeout (seconds) for HTTP calls. |

## Testing

Currently the Rust miner runs in a tight loop and exits only manually (Ctrl+C). Use the same environment variables to control the run. The code is intentionally minimal so you can profile raw hashing throughput before wrapping it in a UI (e.g., Tauri).

## Benchmark mode

Set `SH_MINER_BENCH_SECONDS` (and optionally `SH_MINER_THREADS`, `SH_MINER_DIFFICULTY`) to run only the hash loop for a fixed duration. The binary emits a single JSON line describing the hash count, threads, difficulty, elapsed time, and MH/s so the Go `monitor` tool can parse it:

```bash
SH_MINER_BENCH_SECONDS=10 SH_MINER_THREADS=0 SH_MINER_DIFFICULTY=20 \
  ./target/release/rust_miner
```
