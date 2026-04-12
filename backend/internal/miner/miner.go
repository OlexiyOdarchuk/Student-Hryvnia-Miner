package miner

import (
	"context"
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
	rewardPart []byte
	nodeClient NodeClient
	hashCount  *atomic.Uint32
	diffBytes  int
	threads    int
	diffNibble uint8
}

func InitMiner(hashCount *atomic.Uint32, nodeClient NodeClient, threads int) *Miner {
	return &Miner{
		hashCount:  hashCount,
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

func (m *Miner) MineBlock(ctx context.Context, prevHash string, wallet string) (string, int, int64) {
	timestamp := time.Now().UnixMilli()
	tsPart := strconv.FormatInt(timestamp, 10)

	cores := m.threads
	maxCores := runtime.NumCPU()
	if cores <= 0 {
		if maxCores > 2 {
			cores = maxCores - 1
		} else {
			cores = maxCores
		}
	} else if cores > maxCores {
		cores = maxCores
	}
	if cores < 1 {
		cores = 1
	}

	done := make(chan struct{}, 1)
	stopFlag := new(atomic.Bool)
	var blockHash string
	var foundNonce int

	prefix := []byte(prevHash)
	suffix := make([]byte, 0, len(wallet)+len(m.rewardPart)+len(tsPart))
	suffix = append(suffix, wallet...)
	suffix = append(suffix, m.rewardPart...)
	suffix = append(suffix, tsPart...)

	for i := range cores {
		go func(workerID, cores int) {
			buffer := make([]byte, 0, 512)
			nonce := workerID
			var localCount uint32 = 0

			for {
				if localCount >= 2048 {
					m.hashCount.Add(localCount)
					localCount = 0
					if stopFlag.Load() {
						return
					}
					select {
					case <-ctx.Done():
						stopFlag.Store(true)
						return
					default:
					}
				}
				localCount++

				buffer = buffer[:0]
				buffer = append(buffer, prefix...)
				buffer = strconv.AppendInt(buffer, int64(nonce), 10)
				buffer = append(buffer, suffix...)

				hashArr := sha256.Sum256(buffer)

				if m.checkDifficultyFast(hashArr) {
					if stopFlag.CompareAndSwap(false, true) {
						m.hashCount.Add(localCount)
						hashStr := hex.EncodeToString(hashArr[:])
						slog.Debug("🔨 Found nonce", "nonce", nonce, "hash", hashStr, "wallet", wallet, "prevHash", prevHash)
						blockHash = hashStr
						foundNonce = nonce
						done <- struct{}{}
					}
					return
				}
				nonce += cores
			}
		}(i, cores)
	}

	select {
	case <-done:
	case <-ctx.Done():
	}
	stopFlag.Store(true)
	return blockHash, foundNonce, timestamp
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
