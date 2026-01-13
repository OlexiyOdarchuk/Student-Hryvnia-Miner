package main

import (
	"sync"
	"sync/atomic"
	"time"
)

var (
	sessionMined   int
	walletDataMap  map[string]*WalletStats
	dataMutex      sync.RWMutex
	globalHashrate float64
)

func speedMonitor() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		c := atomic.SwapUint64(&hashCount, 0)
		dataMutex.Lock()
		globalHashrate = float64(c) / 1000000.0
		dataMutex.Unlock()
	}
}

func pushLog(msg string, lType string) {
	logsMutex.Lock()
	defer logsMutex.Unlock()

	lastLogID++
	entry := LogEntry{
		ID:      lastLogID,
		Time:    time.Now().Format("15:04:05"),
		Message: msg,
		Type:    lType,
	}
	logsBuffer = append(logsBuffer, entry)
}

func balanceUpdater() {
	for {
		for _, w := range getWallets() {
			updateSingleBalance(w)
			time.Sleep(1 * time.Second) // Пауза між запитами, щоб не банили
		}
		time.Sleep(5 * time.Second)
	}
}

func updateSingleBalance(wallet string) {
	bal := getBalance(wallet)
	dataMutex.Lock()
	if val, ok := walletDataMap[wallet]; ok {
		val.ServerBalance = bal
	}
	dataMutex.Unlock()
}
