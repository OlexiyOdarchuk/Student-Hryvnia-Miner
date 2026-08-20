package miner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"runtime"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

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

func BenchmarkMiner(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	node := NewMockNodeClient(ctrl)
	node.EXPECT().SubmitBlock(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
	var hashCount atomic.Uint32
	miner := InitMiner(&hashCount, node, 0)
	miner.CompileDifficultyBits(20)
	for b.Loop() {
		miner.MineBlock(context.Background(), "00000741360f68d0f9cec60e2b11015438efb63bb0b5a76af801245b8eefb4a0", "04b22cebe3c0085925e016647ba96e54282763dbcbcc149db52baa3aaef1b76826edcc3feee1eb0ac26acc09d6bc4f3f956ab91f14d2caca25c3402bee8712ab61")
	}
}

func HexHash(s string) [32]byte {
	var h [32]byte
	b, _ := hex.DecodeString(s)
	copy(h[:], b)
	return h
}

// TestMineBlockFindsValidBlock ганяє повний шлях MineBlock (включно з нативним
// ядром, якщо збірка з тегом rustcore) і перевіряє, що повернений блок справді
// хешується в те, що заявлено, і задовольняє складність.
func TestMineBlockFindsValidBlock(t *testing.T) {
	const (
		prevHash = "00000741360f68d0f9cec60e2b11015438efb63bb0b5a76af801245b8eefb4a0"
		wallet   = "04b22cebe3c0085925e016647ba96e54282763dbcbcc149db52baa3aaef1b768"
		bits     = 16
	)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var hashCount atomic.Uint32
	m := InitMiner(&hashCount, NewMockNodeClient(ctrl), 2)
	m.CompileDifficultyBits(bits)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	hash, nonce, ts := m.MineBlock(ctx, prevHash, wallet)
	if hash == "" {
		t.Fatalf("MineBlock не знайшов блок (ядро %q)", CoreName())
	}

	payload := prevHash + strconv.Itoa(nonce) + wallet + "1" + strconv.FormatInt(ts, 10)
	sum := sha256.Sum256([]byte(payload))

	if got := hex.EncodeToString(sum[:]); got != hash {
		t.Errorf("хеш не збігається з перерахованим:\n want %s\n got  %s", got, hash)
	}

	if !m.checkDifficultyFast(sum) {
		t.Errorf("знайдений хеш %s не задовольняє складність %d біт", hash, bits)
	}

	if hashCount.Load() == 0 {
		t.Error("лічильник хешів не оновився")
	}
}

// TestMineBlockRespectsContext перевіряє, що скасування контексту зупиняє
// перебір, а не залишає воркерів крутитися.
func TestMineBlockRespectsContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var hashCount atomic.Uint32
	m := InitMiner(&hashCount, NewMockNodeClient(ctrl), 2)
	// Складність, яку за час тесту не підібрати.
	m.CompileDifficultyBits(120)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	before := runtime.NumGoroutine()

	hash, nonce, _ := m.MineBlock(ctx, "0", "wallet")
	if hash != "" || nonce != 0 {
		t.Fatalf("очікували порожній результат після скасування, отримали %q / %d", hash, nonce)
	}

	// Воркери виходять на межі чанка, тож даємо їм трохи часу.
	deadline := time.Now().Add(5 * time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if leaked := runtime.NumGoroutine() - before; leaked > 0 {
		t.Errorf("після скасування залишилось %d зайвих горутин", leaked)
	}
}

// TestDifficultyMask фіксує розкладку складності на байти й маску.
func TestDifficultyMask(t *testing.T) {
	tests := []struct {
		Bits      int
		FullBytes int
		Mask      uint8
	}{
		{Bits: 0, FullBytes: 0, Mask: 0x00},
		{Bits: 4, FullBytes: 0, Mask: 0xF0},
		{Bits: 8, FullBytes: 1, Mask: 0x00},
		{Bits: 20, FullBytes: 2, Mask: 0xF0},
		{Bits: 23, FullBytes: 2, Mask: 0xFE},
		{Bits: -5, FullBytes: 0, Mask: 0x00},
	}

	for _, tt := range tests {
		fullBytes, mask := difficultyMask(tt.Bits)
		if fullBytes != tt.FullBytes || mask != tt.Mask {
			t.Errorf("difficultyMask(%d) = (%d, %#x), want (%d, %#x)",
				tt.Bits, fullBytes, mask, tt.FullBytes, tt.Mask)
		}
	}
}

// TestHashMeetsDifficultyOutOfRange перевіряє, що позамежна складність не
// призводить до виходу за межі масиву.
func TestHashMeetsDifficultyOutOfRange(t *testing.T) {
	var zero [32]byte
	if !hashMeetsDifficulty(zero, 64, 0xFF) {
		t.Error("нульовий хеш має задовольняти будь-яку складність")
	}

	var one [32]byte
	one[31] = 1
	if hashMeetsDifficulty(one, 64, 0xFF) {
		t.Error("ненульовий хеш не має проходити складність у 512 біт")
	}
}
