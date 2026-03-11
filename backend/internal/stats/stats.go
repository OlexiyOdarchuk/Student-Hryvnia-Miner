package stats

import (
	"context"
	"shminer/backend/types"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const MegahashDivisor = 1000000

//go:generate mockgen -source=stats.go -destination=mocks_test.go -package=stats
type NodeClient interface {
	GetBalance(addr string) float64
}

type Wallets interface {
	GetWallets() []string
}

type WebDashBoard interface {
	BroadcastUpdate()
}
type Stats struct {
	stats         *types.Stats
	walletDataMap map[string]*types.WalletStats
	mu            *sync.RWMutex
	nodeClient    NodeClient
	webDashboard  WebDashBoard
	wallets       Wallets
	BalanceFreqS  time.Duration
}

func InitStats(stats *types.Stats, walletDataMap map[string]*types.WalletStats, mu *sync.RWMutex, node NodeClient, board WebDashBoard, wallets Wallets, balanceFreqS time.Duration) *Stats {
	return &Stats{stats: stats, walletDataMap: walletDataMap, mu: mu, nodeClient: node, webDashboard: board, wallets: wallets, BalanceFreqS: balanceFreqS}
}

func (s *Stats) StartSpeedMonitor(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c := s.stats.HashCount.Swap(0)
			hashPerSec := float64(c) / MegahashDivisor
			s.stats.GlobalHashrate.Store(hashPerSec)

			s.stats.HashrateHistMutex.Lock()
			s.stats.HashrateHistory[s.stats.HashrateHistPos%len(s.stats.HashrateHistory)] = hashPerSec
			s.stats.HashrateHistPos++
			s.stats.HashrateHistMutex.Unlock()
		}
	}
}
func (s *Stats) StartBalanceUpdater(ctx context.Context) {
	freq := s.BalanceFreqS

	ticker := time.NewTicker(freq)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var wg sync.WaitGroup
			wallets := s.wallets.GetWallets()

			for _, w := range wallets {
				wg.Add(1)
				go func(wallet string) {
					defer wg.Done()
					s.UpdateSingleBalance(wallet)
				}(w)
			}
			wg.Wait()
		}
	}
}

func (s *Stats) formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	hour := int(d / time.Hour)
	minutes := int((d % time.Hour) / time.Minute)
	sec := int((d % time.Minute) / time.Second)

	res := make([]byte, 0, 8)

	res = s.appendTwoDigits(res, hour)
	res = append(res, ':')
	res = s.appendTwoDigits(res, minutes)
	res = append(res, ':')
	res = s.appendTwoDigits(res, sec)

	return string(res)
}

func (s *Stats) appendTwoDigits(dst []byte, v int) []byte {
	if v < 10 {
		dst = append(dst, '0')
	}
	return strconv.AppendInt(dst, int64(v), 10)
}

func (s *Stats) UpdateSingleBalance(wallet string) {
	bal := s.nodeClient.GetBalance(wallet)
	s.mu.Lock()
	if val, ok := s.walletDataMap[wallet]; ok {
		val.ServerBalance = bal
	}
	s.mu.Unlock()
	s.webDashboard.BroadcastUpdate()
}

func (s *Stats) GetDashboardData() types.DashboardData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hashVal := s.stats.GlobalHashrate.Load()
	var hash float64
	if hashVal != nil {
		hash = hashVal.(float64)
	}

	totalBal := 0.0
	var lifetimeBlocks uint32
	var wStats []types.WalletStats

	for _, addr := range s.wallets.GetWallets() {
		if ws, ok := s.walletDataMap[addr]; ok {
			totalBal += ws.ServerBalance
			lifetimeBlocks += ws.TotalMined
			wStats = append(wStats, *ws)
		}
	}

	return types.DashboardData{
		Hashrate:       hash,
		SessionBlocks:  s.stats.SessionMined,
		LifetimeBlocks: lifetimeBlocks,
		Uptime:         s.formatDuration(time.Since(s.stats.StartTime)),
		TotalBalance:   totalBal,
		Wallets:        wStats,
		NewLogs:        []types.LogEntry{},
	}
}

func (s *Stats) SessionMinedIncrement() {
	atomic.AddUint32(&s.stats.SessionMined, 1)
}

func (s *Stats) SetWebDashboard(board WebDashBoard) {
	s.webDashboard = board
}

func (s *Stats) HashCountPtr() *atomic.Uint32 {
	return &s.stats.HashCount
}

func (s *Stats) SetNodeClient(node NodeClient) {
	s.nodeClient = node
}

func (s *Stats) SetStartTime(t time.Time) {
	s.stats.StartTime = t
}

func (s *Stats) ResetSessionMined() {
	atomic.StoreUint32(&s.stats.SessionMined, 0)
}

func (s *Stats) SetBalanceFreq(freq time.Duration) {
	s.BalanceFreqS = freq
}
