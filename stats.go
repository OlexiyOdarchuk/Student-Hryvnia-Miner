package main

import (
	"sync"
	"sync/atomic"
	"time"
)

var (
	sessionMined      int
	walletDataMap     map[string]*WalletStats
	dataMutex         sync.RWMutex
	globalHashrate    atomic.Value
	hashrateHistory   [60]float64
	hashrateHistPos   int
	hashrateHistMutex sync.Mutex
)

type LogRing struct {
	data [100]LogEntry
	pos  int
	mu   sync.Mutex
}

var logRing LogRing

func speedMonitor() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		c := atomic.SwapUint64(&hashCount, 0)
		hashPerSec := float64(c) / MegahashDivisor
		globalHashrate.Store(hashPerSec)

		hashrateHistMutex.Lock()
		hashrateHistory[hashrateHistPos%HashrateHistorySize] = hashPerSec
		hashrateHistPos++
		hashrateHistMutex.Unlock()
	}
}

func pushLog(msg string, lType string) {
	logRing.mu.Lock()
	defer logRing.mu.Unlock()

	entry := LogEntry{
		ID:      int64(logRing.pos),
		Time:    time.Now().Format("15:04:05"),
		Message: msg,
		Type:    lType,
	}
	logRing.data[logRing.pos%LogRingBufferSize] = entry
	logRing.pos++
}

func balanceUpdater() {
	for {
		var wg sync.WaitGroup
		wallets := getWallets()

		for _, w := range wallets {
			wg.Add(1)
			go func(wallet string) {
				defer wg.Done()
				updateSingleBalance(wallet)
			}(w)
		}

		wg.Wait()
		time.Sleep(Config.BalanceFreq)
	}
}

func updateSingleBalance(wallet string) {
	bal := getBalance(wallet)
	dataMutex.Lock()
	if val, ok := walletDataMap[wallet]; ok {
		val.ServerBalance = bal
		val.Status = "✓"
	}
	dataMutex.Unlock()
}
