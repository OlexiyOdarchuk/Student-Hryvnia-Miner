package miner

import (
	"encoding/hex"
	"sync/atomic"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestCheckDifficultyFast(t *testing.T) {
	tests := []struct {
		Name  string
		Hash  string
		Valid bool
	}{
		{
			Name:  "valid hash",
			Hash:  "0000018dcf196dcfefe8c7bf6168075d68a518579b4d274b2c42b8a3637de605",
			Valid: true,
		},
		{
			Name:  "invalid hash",
			Hash:  "00000b042d7781c7fec896159ad0337c2c9b5898e8285",
			Valid: true,
		},
		{
			Name:  "not valid hash",
			Hash:  "00100741360f68d0f9cec60e2b11015438efb63bb0b5a76af801245b8eefb4a0",
			Valid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			node := NewMockNodeClient(ctrl)
			var hashCount atomic.Uint32
			miner := InitMiner(&hashCount, node, 3)
			miner.CompileDifficultyBits(20)

			arr := HexHash(tt.Hash)
			valid := miner.checkDifficultyFast(arr)
			if valid != tt.Valid {
				t.Errorf("Want valid %v, but got %v", tt.Valid, valid)
			}
		})
	}
}

func HexHash(s string) [32]byte {
	var h [32]byte
	b, _ := hex.DecodeString(s)
	copy(h[:], b)
	return h
}
