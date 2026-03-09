package stats

import (
	"context"
	"shminer/backend"
	"shminer/backend/internal/nodeclient"
	"shminer/backend/internal/web_dashboard"
	"shminer/backend/types"
	"sync"
	"sync/atomic"
	"time"
)

var (
	sessionMined      int
	walletDataMap     map[string]*types.WalletStats
	dataMutex         sync.RWMutex
	globalHashrate    atomic.Value
	hashrateHistory   [60]float64
	hashrateHistPos   int
	hashrateHistMutex sync.Mutex
	logSeq            int64
)

func init() {
	walletDataMap = make(map[string]*types.WalletStats)
}

func StartSpeedMonitor(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c := atomic.SwapUint64(&backend.hashCount, 0)
			hashPerSec := float64(c) / backend.MegahashDivisor
			globalHashrate.Store(hashPerSec)

			hashrateHistMutex.Lock()
			hashrateHistory[hashrateHistPos%backend.HashrateHistorySize] = hashPerSec
			hashrateHistPos++
			hashrateHistMutex.Unlock()
		}
	}
}

func StartBalanceUpdater(ctx context.Context) {
	freq := backend.Config.BalanceFreq
	if freq <= 0 {
		freq = backend.DefaultBalanceUpdateFreq
	}

	ticker := time.NewTicker(freq)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var wg sync.WaitGroup
			wallets := backend.GetWallets()

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
	web_dashboard.BroadcastUpdate()
}

func GetDashboardData() types.DashboardData {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	hashVal := globalHashrate.Load()
	var hash float64
	if hashVal != nil {
		hash = hashVal.(float64)
	}

	totalBal := 0.0
	lifetimeBlocks := 0
	var wStats []types.WalletStats

	for _, addr := range backend.Wallets {
		if s, ok := walletDataMap[addr]; ok {
			totalBal += s.ServerBalance
			lifetimeBlocks += s.TotalMined
			wStats = append(wStats, *s)
		}
	}

	return types.DashboardData{
		Hashrate:       hash,
		SessionBlocks:  sessionMined,
		LifetimeBlocks: lifetimeBlocks,
		Uptime:         backend.formatDuration(time.Since(backend.startTime)),
		TotalBalance:   totalBal,
		Wallets:        wStats,
		NewLogs:        []types.LogEntry{},
	}
}
