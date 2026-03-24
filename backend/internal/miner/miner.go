package miner

import (
	"encoding/hex"
	"log/slog"
	"runtime"
	"sync/atomic"
	"time"
)

//go:generate mockgen -source=miner.go -destination=mocks_test.go -package=miner
type NodeClient interface {
	SubmitBlock(prev, wallet string, nonce int, ts int64, hash string) bool
}

type Miner struct {
	hashCount      *atomic.Uint32
	difficultyBits int
	nodeClient     NodeClient
	threads        int
}

func InitMiner(hashCount *atomic.Uint32, nodeClient NodeClient, threads int) *Miner {
	return &Miner{
		hashCount:  hashCount,
		nodeClient: nodeClient,
		threads:    threads,
	}
}

func (m *Miner) CompileDifficultyBits(bits int) {
	if bits < 0 {
		bits = 0
	}
	m.difficultyBits = bits
}

func (m *Miner) MineBlock(prevHash string, wallet string) bool {
	timestamp := time.Now().UnixMilli()
	cores := m.threads
	maxCores := runtime.NumCPU()

	if cores <= 0 || cores > maxCores {
		cores = maxCores
	}
	if cores < 1 {
		cores = 1
	}

	sessionID, ok := createMiningSession(prevHash, wallet, m.difficultyBits, cores, timestamp)
	if !ok || sessionID == 0 {
		slog.Error("🛑 Failed to spawn Rust mining session")
		return false
	}

	result, found := pollMiningSession(sessionID)
	stopMiningSession(sessionID)

	if m.hashCount != nil {
		m.hashCount.Add(uint32(result.HashCount))
	}

	if !found {
		return false
	}

	hashHex := hex.EncodeToString(result.Hash[:])
	nonce := result.Nonce

	slog.Debug("🔨 Found nonce", "nonce", nonce, "hash", hashHex, "wallet", wallet, "prevHash", prevHash)

	if m.nodeClient.SubmitBlock(prevHash, wallet, int(nonce), timestamp, hashHex) {
		slog.Info("💰 Block credited! (+1 S-UAH)")
		return true
	}

	slog.Error("❌ Server rejected block")
	return false
}
