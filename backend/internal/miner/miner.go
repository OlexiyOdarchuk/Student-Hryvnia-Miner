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

// chunkSize — скільки nonce воркер перебирає за один прохід, перш ніж
// перевірити прапор зупинки й оновити лічильник хешів. Достатньо мало, щоб
// скасування блоку відпрацьовувало за частки мілісекунди, і достатньо
// велико, щоб накладні витрати на вхід у ядро були непомітними.
const chunkSize = 10000

// maxNonceDigits — запас під десятковий запис nonce при складанні буфера.
const maxNonceDigits = 20

type Miner struct {
	rewardPart []byte
	nodeClient NodeClient
	hashCount  *atomic.Uint32
	diffBits   int
	diffBytes  int
	threads    int
	diffNibble uint8
}

func InitMiner(hashCount *atomic.Uint32, nodeClient NodeClient, threads int) *Miner {
	slog.Debug("Ініціалізація майнера", "core", coreName, "threads", threads)

	return &Miner{
		hashCount:  hashCount,
		diffBits:   0,
		diffBytes:  0,
		diffNibble: 0,
		nodeClient: nodeClient,
		threads:    threads,
		rewardPart: []byte("1"),
	}
}

// CoreName повідомляє, яке ядро перебору вкомпільоване: "rust" або "go".
func CoreName() string {
	return coreName
}

func (m *Miner) CompileDifficultyBits(bits int) {
	if bits < 0 {
		bits = 0
	}

	m.diffBits = bits
	m.diffBytes, m.diffNibble = difficultyMask(bits)
}

// difficultyMask розкладає складність на кількість повністю нульових байтів
// і маску для наступного, частково нульового байта.
func difficultyMask(bits int) (fullBytes int, mask uint8) {
	if bits < 0 {
		bits = 0
	}

	fullBytes = bits / 8
	if remBits := bits % 8; remBits != 0 {
		mask = 0xFF << (8 - remBits)
	}

	return fullBytes, mask
}

// workerCount визначає, скільки воркерів запускати під поточні налаштування.
func (m *Miner) workerCount() int {
	maxCores := runtime.NumCPU()

	cores := m.threads
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

	return cores
}

func (m *Miner) MineBlock(ctx context.Context, prevHash string, wallet string) (string, int, int64) {
	timestamp := time.Now().UnixMilli()
	tsPart := strconv.FormatInt(timestamp, 10)

	cores := m.workerCount()

	done := make(chan struct{}, 1)
	stopFlag := new(atomic.Bool)
	var blockHash string
	var foundNonce int

	prefix := []byte(prevHash)
	suffix := make([]byte, 0, len(wallet)+len(m.rewardPart)+len(tsPart))
	suffix = append(suffix, wallet...)
	suffix = append(suffix, m.rewardPart...)
	suffix = append(suffix, tsPart...)

	for workerID := range cores {
		go func(workerID int) {
			// Воркер проводить майже весь час у гарячому циклі, тож тримаємо
			// його на одному потоці — так планувальник не перекидає його між
			// ядрами й не збиває кеш.
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()

			search := newChunkSearcher(prefix, suffix, m.diffBits)
			nonce := int64(workerID)
			step := int64(cores)

			for {
				if stopFlag.Load() {
					return
				}

				select {
				case <-ctx.Done():
					stopFlag.Store(true)
					return
				default:
				}

				found := search.mine(nonce, step, chunkSize)
				if found < 0 {
					m.hashCount.Add(chunkSize)
					nonce += step * chunkSize

					continue
				}

				// Рахуємо лише ті хеші, які встигли зробити до знахідки.
				m.hashCount.Add(uint32((found-nonce)/step + 1))

				if stopFlag.CompareAndSwap(false, true) {
					hashStr := hashBlock(prefix, found, suffix)
					slog.Debug("🔨 Found nonce", "nonce", found, "hash", hashStr, "wallet", wallet, "prevHash", prevHash)

					blockHash = hashStr
					foundNonce = int(found)
					done <- struct{}{}
				}

				return
			}
		}(workerID)
	}

	select {
	case <-done:
		stopFlag.Store(true)
		return blockHash, foundNonce, timestamp
	case <-ctx.Done():
		stopFlag.Store(true)
		return "", 0, timestamp
	}
}

// hashBlock перераховує хеш знайденого блоку — потрібен лише один раз, щоб
// віддати його у вигляді рядка.
func hashBlock(prefix []byte, nonce int64, suffix []byte) string {
	buffer := make([]byte, 0, len(prefix)+maxNonceDigits+len(suffix))
	buffer = append(buffer, prefix...)
	buffer = strconv.AppendInt(buffer, nonce, 10)
	buffer = append(buffer, suffix...)

	hashArr := sha256.Sum256(buffer)

	return hex.EncodeToString(hashArr[:])
}

func (m *Miner) checkDifficultyFast(hash [32]byte) bool {
	return hashMeetsDifficulty(hash, m.diffBytes, m.diffNibble)
}

func hashMeetsDifficulty(hash [32]byte, fullBytes int, mask uint8) bool {
	if fullBytes > len(hash) {
		fullBytes = len(hash)
	}

	for i := range fullBytes {
		if hash[i] != 0 {
			return false
		}
	}

	if mask != 0 && fullBytes < len(hash) {
		if hash[fullBytes]&mask != 0 {
			return false
		}
	}

	return true
}
