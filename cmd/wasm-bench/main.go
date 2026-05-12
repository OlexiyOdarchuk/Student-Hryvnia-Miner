//go:build js && wasm

// wasm-bench: SHA-256 hashrate kernel for the ishawyha.dev demo.
//
// Mirrors the real Student-Hryvnia work:
//
//   - Pre-image:    prevHash + nonce + miner + reward + timestamp
//   - Difficulty:   20 leading zero bits  (== first 5 hex chars '0')
//                   exactly what MINERJS.js checks via
//                   hash.startsWith("0".repeat(5)) and what the SHMiner
//                   backend default sets (DefaultDifficulty = 20).
//
// Design note: we *don't* run a long-lived Go goroutine to do the work.
// TinyGo's WASM scheduler is single-threaded and a tight for-loop hogs
// the worker, which keeps JS-side `setInterval` reporting + `benchStop`
// callbacks from ever firing. Instead, we expose `benchBatch(n)` and let
// the JS worker drive the loop in chunks of N — yielding to the event
// loop between batches.  That makes both progress reporting and stop
// responsive without sacrificing throughput meaningfully (the per-batch
// FFI overhead is ~µs for batches of a few thousand hashes).
package main

import (
	"crypto/sha256"
	"strconv"
	"sync/atomic"
	"syscall/js"
	"time"
)

const (
	prevHash = "0000000000000000000000000000000000000000000000000000000000000000"
	miner    = "ishawyha_bench_wallet_pubkey_for_demo_only_xxxxxxxxxxxxxxxx"
	reward   = "1"
)

var (
	count   atomic.Uint64
	tokens  atomic.Uint64
	nonce   atomic.Uint64
	// Pre-allocated scratch — reused across batches.
	scratch []byte
)

// hitsTarget = first two bytes are zero AND top nibble of byte 2 is zero
// (20-bit difficulty == 5 zero hex chars at the start of the digest).
//
//go:inline
func hitsTarget(h *[32]byte) bool {
	return h[0] == 0 && h[1] == 0 && (h[2]&0xF0) == 0
}

func reset(this js.Value, args []js.Value) any {
	count.Store(0)
	tokens.Store(0)
	nonce.Store(0)
	return nil
}

// benchBatch(n) — runs n hashes synchronously, updates counters,
// returns nothing. JS reads counts via benchCount/benchTokens.
func benchBatch(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return nil
	}
	n := uint64(args[0].Int())
	if n == 0 {
		return nil
	}

	if cap(scratch) < 256 {
		scratch = make([]byte, 0, 256)
	}
	buf := scratch[:0]
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)

	startNonce := nonce.Load()
	var localHits uint64
	for i := uint64(0); i < n; i++ {
		buf = buf[:0]
		buf = append(buf, prevHash...)
		buf = strconv.AppendUint(buf, startNonce+i, 10)
		buf = append(buf, miner...)
		buf = append(buf, reward...)
		buf = append(buf, ts...)
		sum := sha256.Sum256(buf)
		if hitsTarget(&sum) {
			localHits++
		}
	}
	scratch = buf

	count.Add(n)
	if localHits > 0 {
		tokens.Add(localHits)
	}
	nonce.Add(n)
	return nil
}

func getCount(this js.Value, args []js.Value) any {
	return js.ValueOf(int(count.Load()))
}

func getTokens(this js.Value, args []js.Value) any {
	return js.ValueOf(int(tokens.Load()))
}

func main() {
	js.Global().Set("benchBatch", js.FuncOf(benchBatch))
	js.Global().Set("benchReset", js.FuncOf(reset))
	js.Global().Set("benchCount", js.FuncOf(getCount))
	js.Global().Set("benchTokens", js.FuncOf(getTokens))
	js.Global().Set("benchReady", js.ValueOf(true))
	select {}
}
