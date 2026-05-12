//go:build js && wasm

// wasm-bench is a minimal SHA-256 hashrate bench exposed to the browser.
//
// It replicates the *semantic* work of the Student Hryvnia official JS miner
// (`calculateBlockHash`: prevHash + nonce + miner + reward + timestamp →
// SHA-256 → lower-case hex) so the comparison on ishawyha.dev is honest.
// The whole point of this binary is to show how much overhead the original
// miner pays for going through `await window.crypto.subtle.digest` and a
// cooperative `sleep(1)` every 1000 iterations.
package main

import (
	"crypto/sha256"
	"strconv"
	"sync/atomic"
	"syscall/js"
	"time"
)

var (
	running atomic.Bool
	count   atomic.Uint64
)

// Fixed inputs that match the shape of an Student-Hryvnia block. Length
// matters (one extra byte per nonce digit changes how the SHA-256 message
// schedule fills) but content does not — both sides hash the same bytes.
const (
	prevHash = "0000000000000000000000000000000000000000000000000000000000000000"
	miner    = "ishawyha_bench_wallet_pubkey_for_demo_only_xxxxxxxxxxxxxxxx"
	reward   = "1"
)

func start(this js.Value, args []js.Value) any {
	if !running.CompareAndSwap(false, true) {
		return nil
	}
	count.Store(0)

	go func() {
		// Pre-allocate a single buffer; rebuild only the volatile parts each
		// iteration. The JS baseline can't dodge this allocation because
		// TextEncoder.encode always returns a fresh Uint8Array.
		base := []byte(prevHash)
		minerBytes := []byte(miner)
		rewardBytes := []byte(reward)

		buf := make([]byte, 0, len(base)+24+len(minerBytes)+len(rewardBytes)+24)
		hex := make([]byte, sha256.Size*2)
		_ = hex

		ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
		tsBytes := []byte(ts)

		var nonce uint64
		for running.Load() {
			// Refresh timestamp roughly once every 1024 hashes — keeps the
			// pre-image churning the way the original does, without paying
			// for time.Now on every call (microbenchmark contamination).
			if nonce&0x3FF == 0 {
				ts = strconv.FormatInt(time.Now().UnixMilli(), 10)
				tsBytes = tsBytes[:0]
				tsBytes = append(tsBytes, ts...)
			}

			buf = buf[:0]
			buf = append(buf, base...)
			buf = strconv.AppendUint(buf, nonce, 10)
			buf = append(buf, minerBytes...)
			buf = append(buf, rewardBytes...)
			buf = append(buf, tsBytes...)

			sum := sha256.Sum256(buf)
			// Hex-encode the digest to mirror what the JS impl returns.
			// We don't actually use the string — discard so the compiler can't
			// strip the work — but we make sure encode runs.
			_ = encodeHexFast(hex, sum[:])

			nonce++
			count.Add(1)
		}
	}()
	return nil
}

func encodeHexFast(dst, src []byte) []byte {
	const table = "0123456789abcdef"
	for i, b := range src {
		dst[i*2] = table[b>>4]
		dst[i*2+1] = table[b&0x0f]
	}
	return dst
}

func stop(this js.Value, args []js.Value) any {
	running.Store(false)
	return nil
}

func getCount(this js.Value, args []js.Value) any {
	// JS numbers are 53-bit safe; our counts won't approach 2^53 in a
	// realistic browser session, so a plain int conversion is fine.
	return js.ValueOf(int(count.Load()))
}

// Single-call helper used by the browser to do a fixed number of hashes
// synchronously (handy for a "first burst" warm-up). Returns the elapsed
// wall-clock time in nanoseconds so JS can compute hashes/sec without
// involving the timer infrastructure of the host page.
func runFixed(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return js.ValueOf(0)
	}
	n := uint64(args[0].Int())
	if n == 0 {
		return js.ValueOf(0)
	}
	buf := make([]byte, 0, 256)
	hex := make([]byte, sha256.Size*2)
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	t0 := time.Now()
	for nonce := uint64(0); nonce < n; nonce++ {
		buf = buf[:0]
		buf = append(buf, prevHash...)
		buf = strconv.AppendUint(buf, nonce, 10)
		buf = append(buf, miner...)
		buf = append(buf, reward...)
		buf = append(buf, ts...)
		sum := sha256.Sum256(buf)
		_ = encodeHexFast(hex, sum[:])
	}
	return js.ValueOf(time.Since(t0).Nanoseconds())
}

func main() {
	js.Global().Set("benchStart", js.FuncOf(start))
	js.Global().Set("benchStop", js.FuncOf(stop))
	js.Global().Set("benchCount", js.FuncOf(getCount))
	js.Global().Set("benchRunFixed", js.FuncOf(runFixed))
	js.Global().Set("benchReady", js.ValueOf(true))
	// Park the runtime so the registered FuncOf callbacks stay alive.
	select {}
}
