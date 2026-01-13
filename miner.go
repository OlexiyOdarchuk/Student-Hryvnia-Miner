package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	hashCount uint64
	found     int32
	startTime time.Time
)

func checkDifficulty(hashArr [32]byte, difficulty string) bool {
	if len(difficulty) == 0 {
		return true
	}

	for i, c := range difficulty {
		byteIndex := i / 2
		if byteIndex >= 32 {
			return false
		}

		if c == '0' {
			if i%2 == 0 {
				if hashArr[byteIndex] >= 0x10 {
					return false
				}
			} else {
				if hashArr[byteIndex] != 0 {
					return false
				}
			}
		}
	}
	return true
}

func mineBlock(prevHash string, wallet string) bool {
	atomic.StoreInt32(&found, 0)
	timestamp := time.Now().UnixNano() / 1e6

	txPart := []byte(`[{"from":null,"to":"` + wallet + `","amount":1}]`)
	minerPart := []byte(wallet)
	rewardPart := []byte("1")
	tsPart := []byte(strconv.FormatInt(timestamp, 10))

	cores := runtime.NumCPU()
	doneChan := make(chan bool)
	var successFlag int32

	for i := 0; i < cores; i++ {
		go func(workerID int) {
			buffer := make([]byte, 0, 512)
			nonce := workerID

			for atomic.LoadInt32(&found) == 0 {
				buffer = buffer[:0]
				buffer = append(buffer, prevHash...)
				buffer = append(buffer, txPart...)
				buffer = strconv.AppendInt(buffer, int64(nonce), 10)
				buffer = append(buffer, minerPart...)
				buffer = append(buffer, rewardPart...)
				buffer = append(buffer, tsPart...)

				hashArr := sha256.Sum256(buffer)
				atomic.AddUint64(&hashCount, 1)

				if checkDifficulty(hashArr, Config.Difficulty) {
					if atomic.CompareAndSwapInt32(&found, 0, 1) {
						hashStr := hex.EncodeToString(hashArr[:])
						pushLog(fmt.Sprintf("🔨 Found nonce: %d", nonce), "info")

						if submitBlock(prevHash, wallet, nonce, timestamp, hashStr) {
							atomic.StoreInt32(&successFlag, 1)
							pushLog(fmt.Sprintln("💰 Блок зараховано! (+2 S-UAH)"), "success")
						} else {
							pushLog("❌ Сервер відхилив блок", "error")
						}
						doneChan <- true
					}
					return
				}
				nonce += cores
			}
		}(i)
	}

	<-doneChan
	return atomic.LoadInt32(&successFlag) == 1
}
