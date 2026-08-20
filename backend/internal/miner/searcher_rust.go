//go:build rustcore

package miner

/*
#cgo linux,amd64   LDFLAGS: -L${SRCDIR}/rust_miner_core/lib/linux_amd64
#cgo linux,arm64   LDFLAGS: -L${SRCDIR}/rust_miner_core/lib/linux_arm64
#cgo darwin,amd64  LDFLAGS: -L${SRCDIR}/rust_miner_core/lib/darwin_amd64
#cgo darwin,arm64  LDFLAGS: -L${SRCDIR}/rust_miner_core/lib/darwin_arm64
#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/rust_miner_core/lib/windows_amd64
#cgo windows,arm64 LDFLAGS: -L${SRCDIR}/rust_miner_core/lib/windows_arm64

#cgo linux   LDFLAGS: -lrust_miner_core -lm -ldl -lpthread
#cgo darwin  LDFLAGS: -lrust_miner_core
#cgo windows LDFLAGS: -lrust_miner_core -lws2_32 -luserenv -lbcrypt -ladvapi32 -lntdll

#include <stdint.h>
#include <stddef.h>

uint32_t shminer_core_abi_version(void);

int64_t shminer_mine_chunk(const uint8_t *prefix, size_t prefix_len,
                           const uint8_t *suffix, size_t suffix_len,
                           uint32_t diff_bits, int64_t start_nonce,
                           int64_t step, int64_t chunk_size);
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// coreName — яке ядро перебору вкомпільоване в цю збірку.
const coreName = "rust"

// coreABIVersion має збігатися з CORE_ABI_VERSION у rust_miner_core/src/lib.rs.
// Перевірка ловить ситуацію, коли Go перезібрали, а статичну бібліотеку — ні.
const coreABIVersion = 2

func init() {
	if got := uint32(C.shminer_core_abi_version()); got != coreABIVersion {
		panic(fmt.Sprintf(
			"rust_miner_core: невідповідність ABI (бібліотека %d, потрібно %d) — перезберіть rust_miner_core",
			got, coreABIVersion,
		))
	}
}

// chunkSearcher — перебір у нативному ядрі на Rust.
//
// Слайси передаються в C напряму: вони не містять Go-вказівників, тож правила
// cgo це дозволяють, а на час виклику їх утримує живими цей об'єкт.
type chunkSearcher struct {
	prefix   []byte
	suffix   []byte
	diffBits C.uint32_t
}

func newChunkSearcher(prefix, suffix []byte, diffBits int) *chunkSearcher {
	if diffBits < 0 {
		diffBits = 0
	}

	return &chunkSearcher{
		prefix:   prefix,
		suffix:   suffix,
		diffBits: C.uint32_t(diffBits),
	}
}

// mine перебирає count значень nonce, починаючи зі startNonce, з кроком step.
// Повертає знайдений nonce або -1.
func (s *chunkSearcher) mine(startNonce, step, count int64) int64 {
	return int64(C.shminer_mine_chunk(
		bytePtr(s.prefix), C.size_t(len(s.prefix)),
		bytePtr(s.suffix), C.size_t(len(s.suffix)),
		s.diffBits,
		C.int64_t(startNonce), C.int64_t(step), C.int64_t(count),
	))
}

// bytePtr дає вказівник на дані слайса; для порожнього слайса — nil, який
// ядро коректно обробляє в парі з нульовою довжиною.
func bytePtr(b []byte) *C.uint8_t {
	if len(b) == 0 {
		return nil
	}

	return (*C.uint8_t)(unsafe.Pointer(&b[0]))
}
