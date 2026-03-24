package miner

// #cgo CFLAGS: -I${SRCDIR}/../../SHMiner_RUST/include
// #cgo LDFLAGS: -L${SRCDIR}/../../SHMiner_RUST/target/release -lshminer
// #include <stdlib.h>
// #include "shminer.h"
import "C"

import "unsafe"

type RustMineResult struct {
	Nonce     uint64
	Hash      [32]byte
	HashCount uint64
}

func createMiningSession(prevHash, wallet string, difficulty, threads int, timestamp int64) (uint32, bool) {
	prevPtr := C.CString(prevHash)
	walletPtr := C.CString(wallet)
	defer C.free(unsafe.Pointer(prevPtr))
	defer C.free(unsafe.Pointer(walletPtr))

	prevLen := C.size_t(len(prevHash))
	walletLen := C.size_t(len(wallet))

	var req C.ShMinerWorkRequest
	req.prev_hash = prevPtr
	req.prev_hash_len = prevLen
	req.wallet = walletPtr
	req.wallet_len = walletLen
	if difficulty < 0 {
		difficulty = 0
	}
	req.difficulty_bits = C.uint32_t(uint32(difficulty))
	if threads < 1 {
		threads = 1
	}
	req.threads = C.uint32_t(uint32(threads))
	req.timestamp = C.int64_t(timestamp)

	id := C.create_session(&req)
	if id == 0 {
		return 0, false
	}
	return uint32(id), true
}

func pollMiningSession(sessionID uint32) (RustMineResult, bool) {
	var out C.ShMinerResult
	ok := C.step_session(C.uint32_t(sessionID), &out)

	var result RustMineResult
	result.Nonce = uint64(out.nonce)
	result.HashCount = uint64(out.hash_count)
	for i := range result.Hash {
		result.Hash[i] = byte(out.hash[i])
	}

	return result, bool(ok)
}

func stopMiningSession(sessionID uint32) {
	C.stop_session(C.uint32_t(sessionID))
}
