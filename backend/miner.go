package backend

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	hashCount uint64
	found     int32
	startTime time.Time

	diffBytes  int
	diffNibble uint8
)

func compileDifficultyBits(bits int) {
	if bits <= 0 {
		diffBytes = 0
		diffNibble = 0
		return
	}

	if bits >= 256 {
		diffBytes = 32
		diffNibble = 0
		return
	}

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

func MineBlock(prevHash string, wallet string) bool {
	atomic.StoreInt32(&found, 0)

	timestamp := time.Now().UnixMilli()

	minerPart := []byte(wallet)
	rewardPart := []byte("1")
	tsPart := []byte(strconv.FormatInt(timestamp, 10))

	cores := Config.Threads
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
						PushLog(fmt.Sprintf("🔨 Found nonce: %d", nonce), "info")

						if SubmitBlock(prevHash, wallet, nonce, timestamp, hashStr) {
							atomic.StoreInt32(&successFlag, 1)
							PushLog("💰 Блок зараховано! (+1 S-UAH)", "success")
						} else {
							PushLog("❌ Сервер відхилив блок", "error")
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
	if Config.Difficulty < 1 {
		Config.Difficulty = 1
	}

	compileDifficultyBits(Config.Difficulty)
	startTime = time.Now()

	rand.New(rand.NewSource(time.Now().UnixNano()))

	go StartSpeedMonitor(ctx)
	go StartBalanceUpdater(ctx)
	go StartWebServer()

	PushLog("🔨 МАЙНЕР ЗАПУЩЕНО...", "info")

	for {
		select {
		case <-ctx.Done():
			PushLog("🛑 Mining stopped", "info")
			return
		default:
		}

		prevHash := getChainLastHashCached()
		if prevHash == "" {
			PushLog("⚠️ Немає зв'язку з сервером. Рестарт...", "error")
			time.Sleep(2 * time.Second)
			continue
		}

		ws := GetWallets()
		if len(ws) == 0 {
			time.Sleep(2 * time.Second)
			continue
		}

		currentWallet := ws[rand.Intn(len(ws))]

		dataMutex.Lock()
		stats, exists := walletDataMap[currentWallet]
		isWorking := true
		if exists {
			isWorking = stats.Working
		}
		dataMutex.Unlock()

		if !isWorking {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		success := MineBlock(prevHash, currentWallet)

		if success {
			dataMutex.Lock()
			sessionMined++
			if ws, ok := walletDataMap[currentWallet]; ok {
				ws.SessionMined++
				ws.TotalMined++
			}
			dataMutex.Unlock()

			syncStorage()
			SaveStorage(GetSessionPassword(), CurrentStorage)

			go func() {
				updateSingleBalance(currentWallet)
				BroadcastUpdate()
			}()

			BroadcastUpdate()
		}

		time.Sleep(MinerSleepInterval)
	}
}
