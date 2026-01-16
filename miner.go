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

func mineBlock(prevHash string, wallet string) bool {
	atomic.StoreInt32(&found, 0)

	timestamp := time.Now().UnixMilli()

	//txPart := []byte(`[{"from":null,"to":"` + wallet + `","amount":1}]`)
	minerPart := []byte(wallet)
	rewardPart := []byte("1")
	tsPart := []byte(strconv.FormatInt(timestamp, 10))

	cores := runtime.NumCPU()
	done := make(chan struct{})
	var successFlag int32

	for i := range cores {
		go func(workerID int) {
			buffer := make([]byte, 0, 512)
			nonce := workerID

			for atomic.LoadInt32(&found) == 0 {
				buffer = buffer[:0]
				buffer = append(buffer, prevHash...)
				//buffer = append(buffer, txPart...)
				buffer = strconv.AppendInt(buffer, int64(nonce), 10)
				buffer = append(buffer, minerPart...)
				buffer = append(buffer, rewardPart...)
				buffer = append(buffer, tsPart...)

				hashArr := sha256.Sum256(buffer)
				atomic.AddUint64(&hashCount, 1)

				if checkDifficultyFast(hashArr) {
					if atomic.CompareAndSwapInt32(&found, 0, 1) {
						hashStr := hex.EncodeToString(hashArr[:])
						pushLog(fmt.Sprintf("🔨 Found nonce: %d", nonce), "info")

						if submitBlock(prevHash, wallet, nonce, timestamp, hashStr) {
							atomic.StoreInt32(&successFlag, 1)
							pushLog("💰 Блок зараховано! (+1 S-UAH)", "success")
						} else {
							pushLog("❌ Сервер відхилив блок", "error")
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
