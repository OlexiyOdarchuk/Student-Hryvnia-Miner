package miner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"runtime"
	"shminer/backend/app/config"
	"shminer/backend/internal/nodeclient"
	"shminer/backend/internal/stats"
	"shminer/backend/internal/web_dashboard"
	"strconv"
	"sync/atomic"
	"time"
)

type Miner struct {
	hashCount *atomic.Uint32
	found     atomic.Bool
}

var (
	diffBytes  uint16
	diffNibble uint16
)

func (m *Miner) CompileDifficultyBits(bits uint16) {
	diffBytes = bits / 8
	remBits := bits % 8

	if remBits == 0 {
		diffNibble = 0
	} else {
		diffNibble = 0xFF << (8 - remBits)
	}
}

func checkDifficultyFast(hash [32]byte) bool {
	for i := 0; i < diffBytes; i++ {
		if hash[i] != 0 {
			return false
		}
	}

	if diffNibble != 0 && diffBytes < 32 {
		if hash[diffBytes]&diffNibble != 0 {
			return false
		}
	}
	return true
}

func (m *Miner) MineBlock(prevHash string, wallet string) bool {
	m.found.Store(false)

	timestamp := time.Now().UnixMilli()

	minerPart := []byte(wallet)
	rewardPart := []byte("1")
	tsPart := []byte(strconv.FormatInt(timestamp, 10))

	cores := config.Config.Threads
	maxCores := runtime.NumCPU()

	if cores <= 0 || cores > maxCores {
		cores = maxCores
	}

	if cores < 1 {
		cores = 1
	}

	done := make(chan struct{})
	var successFlag int32

	for i := range cores {
		go func(workerID int) {
			buffer := make([]byte, 0, 512)
			nonce := workerID

			for atomic.LoadInt32(&found) == 0 {
				buffer = buffer[:0]
				buffer = append(buffer, prevHash...)
				buffer = strconv.AppendInt(buffer, int64(nonce), 10)
				buffer = append(buffer, minerPart...)
				buffer = append(buffer, rewardPart...)
				buffer = append(buffer, tsPart...)

				hashArr := sha256.Sum256(buffer)
				atomic.AddUint64(&hashCount, 1)

				if checkDifficultyFast(hashArr) {
					if atomic.CompareAndSwapInt32(&found, 0, 1) {
						hashStr := hex.EncodeToString(hashArr[:])
						stats.PushLog(fmt.Sprintf("🔨 Found nonce: %d", nonce), "info")

						if nodeclient.SubmitBlock(prevHash, wallet, nonce, timestamp, hashStr) {
							atomic.StoreInt32(&successFlag, 1)
							stats.PushLog("💰 Блок зараховано! (+1 S-UAH)", "success")
						} else {
							stats.PushLog("❌ Сервер відхилив блок", "error")
						}
						close(done)
					}
					return
				}
				nonce += cores
			}
		}(i)
	}

	<-done
	return atomic.LoadInt32(&successFlag) == 1
}

func StartMiningLoop(ctx context.Context) {
	if config.Config.Difficulty < 1 {
		config.Config.Difficulty = 1
	}

	compileDifficultyBits(config.Config.Difficulty)
	stats.dataMutex.Lock()
	startTime = time.Now()
	stats.dataMutex.Unlock()

	rand.New(rand.NewSource(time.Now().UnixNano()))

	go stats.StartSpeedMonitor(ctx)
	go stats.StartBalanceUpdater(ctx)
	go web_dashboard.StartWebServer()

	stats.PushLog("🔨 МАЙНЕР ЗАПУЩЕНО...", "info")

	for {
		select {
		case <-ctx.Done():
			stats.PushLog("🛑 Mining stopped", "info")
			return
		default:
		}

		prevHash := nodeclient.GetChainLastHashCached()
		if prevHash == "" {
			stats.PushLog("⚠️ Немає зв'язку з сервером. Рестарт...", "error")
			time.Sleep(2 * time.Second)
			continue
		}

		ws := GetWallets()
		if len(ws) == 0 {
			time.Sleep(2 * time.Second)
			continue
		}

		currentWallet := ws[rand.Intn(len(ws))]

		stats.dataMutex.Lock()
		stats, exists := stats.walletDataMap[currentWallet]
		isWorking := true
		if exists {
			isWorking = stats.Working
		}
		stats.dataMutex.Unlock()

		if !isWorking {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		success := MineBlock(prevHash, currentWallet)

		if success {
			stats.dataMutex.Lock()
			stats.sessionMined++
			if ws, ok := stats.walletDataMap[currentWallet]; ok {
				ws.SessionMined++
				ws.TotalMined++
			}

			syncStorage()
			SaveStorage(sessionPassword, CurrentStorage)
			stats.dataMutex.Unlock()

			go func() {
				stats.updateSingleBalance(currentWallet)
				web_dashboard.BroadcastUpdate()
			}()

			web_dashboard.BroadcastUpdate()
		}

		time.Sleep(MinerSleepInterval)
	}
}
