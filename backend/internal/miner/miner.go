package miner

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

//go:generate mockgen -source=miner.go -destination=mocks_test.go -package=miner
type NodeClient interface {
	SubmitBlock(prev, wallet string, nonce int, ts int64, hash string) bool
}

type Miner struct {
	hashCount  *atomic.Uint32
	found      atomic.Bool
	diffBytes  int
	diffNibble uint8
	nodeClient NodeClient
	threads    int
	rewardPart []byte
}

func InitMiner(hashCount *atomic.Uint32, nodeClient NodeClient, threads int) *Miner {
	return &Miner{
		hashCount:  hashCount,
		found:      atomic.Bool{},
		diffBytes:  0,
		diffNibble: 0,
		nodeClient: nodeClient,
		threads:    threads,
		rewardPart: []byte("1"),
	}
}

func (m *Miner) CompileDifficultyBits(bits int) {
	m.diffBytes = bits / 8
	remBits := bits % 8

	if remBits == 0 {
		m.diffNibble = 0
	} else {
		m.diffNibble = 0xFF << (8 - remBits)
	}
}

func (m *Miner) MineBlock(prevHash string, wallet string) bool {
	m.found.Store(false)

	timestamp := time.Now().UnixMilli()

	tsPart := strconv.FormatInt(timestamp, 10)

	cores := m.threads
	maxCores := runtime.NumCPU()

	if cores <= 0 || cores > maxCores {
		cores = maxCores
	}

	if cores < 1 {
		cores = 1
	}
	done := make(chan struct{})
	var successFlag atomic.Bool

	for i := range cores {
		go func(workerID, cores int) {
			buffer := make([]byte, 0, 512)
			nonce := workerID

			for !m.found.Load() {
				buffer = buffer[:0]
				buffer = append(buffer, prevHash...)
				buffer = strconv.AppendInt(buffer, int64(nonce), 10)
				buffer = append(buffer, wallet...)
				buffer = append(buffer, m.rewardPart...)
				buffer = append(buffer, tsPart...)

				hashArr := sha256.Sum256(buffer)
				m.hashCount.Add(1)

				if m.checkDifficultyFast(hashArr) {
					if m.found.CompareAndSwap(false, true) {
						hashStr := hex.EncodeToString(hashArr[:])
						slog.Debug("🔨 Found nonce", "nonce", nonce, "hash", hashStr, "wallet", wallet, "prevHash", prevHash)

						if m.nodeClient.SubmitBlock(prevHash, wallet, nonce, timestamp, hashStr) {
							successFlag.Store(true)
							slog.Info("💰 Block credited! (+1 S-UAH)")
						} else {
							slog.Error("❌ Server rejected block")
						}
						close(done)
					}
					return
				}
				nonce += cores
			}
		}(i, cores)
	}

	<-done
	return successFlag.Load()
}

func (m *Miner) checkDifficultyFast(hash [32]byte) bool {
	for i := 0; i < m.diffBytes; i++ {
		if hash[i] != 0 {
			return false
		}
	}

	if m.diffNibble != 0 && m.diffBytes < 32 {
		if hash[m.diffBytes]&m.diffNibble != 0 {
			return false
		}
	}
	return true
}
