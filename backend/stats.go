package backend

import (
	"context"
	"shminer/backend/internal/nodeclient"
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
	logSeq            int64
)

var LogCallback func(LogEntry)

func init() {
	walletDataMap = make(map[string]*WalletStats)
}

func StartSpeedMonitor(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c := atomic.SwapUint64(&hashCount, 0)
			hashPerSec := float64(c) / MegahashDivisor
			globalHashrate.Store(hashPerSec)

			hashrateHistMutex.Lock()
			hashrateHistory[hashrateHistPos%HashrateHistorySize] = hashPerSec
			hashrateHistPos++
			hashrateHistMutex.Unlock()
		}
	}
}

func PushLog(msg string, lType string) {
	id := atomic.AddInt64(&logSeq, 1)
	entry := LogEntry{
		ID:      id,
		Time:    time.Now().Format("15:04:05"),
		Message: msg,
		Type:    lType,
	}

	if LogCallback != nil {
		LogCallback(entry)
	}
}

func StartBalanceUpdater(ctx context.Context) {
	freq := Config.BalanceFreq
	if freq <= 0 {
		freq = DefaultBalanceUpdateFreq
	}

	ticker := time.NewTicker(freq)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var wg sync.WaitGroup
			wallets := GetWallets()

			for _, w := range wallets {
				wg.Add(1)
				go func(wallet string) {
					defer wg.Done()
					updateSingleBalance(wallet)
				}(w)
			}
			wg.Wait()
		}
	}
}

func updateSingleBalance(wallet string) {
	bal := nodeclient.GetBalance(wallet)
	dataMutex.Lock()
	if val, ok := walletDataMap[wallet]; ok {
		val.ServerBalance = bal
	}
	dataMutex.Unlock()
	BroadcastUpdate()
}

func GetDashboardData() DashboardData {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	hashVal := globalHashrate.Load()
	var hash float64
	if hashVal != nil {
		hash = hashVal.(float64)
	}

	totalBal := 0.0
	lifetimeBlocks := 0
	var wStats []WalletStats

	for _, addr := range Wallets {
		if s, ok := walletDataMap[addr]; ok {
			totalBal += s.ServerBalance
			lifetimeBlocks += s.TotalMined
			wStats = append(wStats, *s)
		}
	}

	return DashboardData{
		Hashrate:       hash,
		SessionBlocks:  sessionMined,
		LifetimeBlocks: lifetimeBlocks,
		Uptime:         formatDuration(time.Since(startTime)),
		TotalBalance:   totalBal,
		Wallets:        wStats,
		NewLogs:        []LogEntry{},
	}
}
