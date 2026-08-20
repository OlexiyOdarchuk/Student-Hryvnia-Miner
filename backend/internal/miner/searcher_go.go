//go:build !rustcore

package miner

import (
	"crypto/sha256"
	"strconv"
)

// coreName — яке ядро перебору вкомпільоване в цю збірку.
const coreName = "go"

// chunkSearcher — перебір на чистому Go. Використовується у збірках без тега
// `rustcore`: у розробці та на платформах, під які статична бібліотека ядра
// не збирається.
type chunkSearcher struct {
	prefix    []byte
	suffix    []byte
	buffer    []byte
	fullBytes int
	mask      uint8
}

func newChunkSearcher(prefix, suffix []byte, diffBits int) *chunkSearcher {
	fullBytes, mask := difficultyMask(diffBits)

	return &chunkSearcher{
		prefix:    prefix,
		suffix:    suffix,
		buffer:    make([]byte, 0, len(prefix)+maxNonceDigits+len(suffix)),
		fullBytes: fullBytes,
		mask:      mask,
	}
}

// mine перебирає count значень nonce, починаючи зі startNonce, з кроком step.
// Повертає знайдений nonce або -1.
func (s *chunkSearcher) mine(startNonce, step, count int64) int64 {
	if startNonce < 0 || step <= 0 || count <= 0 {
		return -1
	}

	nonce := startNonce
	for range count {
		s.buffer = s.buffer[:0]
		s.buffer = append(s.buffer, s.prefix...)
		s.buffer = strconv.AppendInt(s.buffer, nonce, 10)
		s.buffer = append(s.buffer, s.suffix...)

		if hashMeetsDifficulty(sha256.Sum256(s.buffer), s.fullBytes, s.mask) {
			return nonce
		}

		nonce += step
	}

	return -1
}
