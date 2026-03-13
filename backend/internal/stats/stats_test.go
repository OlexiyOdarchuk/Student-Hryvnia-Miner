package stats

import (
	"shminer/backend/types"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func TestStats_FormatDuration(t *testing.T) {
	s := &Stats{}

	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"With hours", time.Hour + 2*time.Minute + 3*time.Second, "01:02:03"},
		{"More than 9 hours", 10*time.Hour + 15*time.Minute + 30*time.Second, "10:15:30"},
		{"Only seconds", 5 * time.Second, "00:00:05"},
		{"Zero duration", 0, "00:00:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := s.formatDuration(tt.duration)
			if res != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, res)
			}
		})
	}
}

func TestStats_SessionMinedIncrement(t *testing.T) {
	tests := []struct {
		name           string
		initialMined   uint32
		incrementCount int
		expectedMined  uint32
	}{
		{"Increment once", 0, 1, 1},
		{"Increment multiple times", 0, 5, 5},
		{"Increment from existing", 10, 2, 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := InitStats(&types.Stats{SessionMined: tt.initialMined}, nil, &sync.RWMutex{}, nil, nil, nil, 0)

			for i := 0; i < tt.incrementCount; i++ {
				s.SessionMinedIncrement()
			}

			if s.stats.SessionMined != tt.expectedMined {
				t.Errorf("Expected %d, got %d", tt.expectedMined, s.stats.SessionMined)
			}
		})
	}
}

func TestStats_UpdateSingleBalance(t *testing.T) {
	tests := []struct {
		name         string
		wallet       string
		initialBal   float64
		fetchedBal   float64
		walletExists bool
		expectedBal  float64
	}{
		{"Update existing wallet", "wallet1", 0.0, 15.5, true, 15.5},
		{"Update non-existent wallet", "wallet_missing", 0.0, 10.0, false, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockNodeClient := NewMockNodeClient(ctrl)
			mockDashboard := NewMockWebDashBoard(ctrl)

			walletMap := make(map[string]*types.WalletStats)
			if tt.walletExists {
				walletMap[tt.wallet] = &types.WalletStats{ServerBalance: tt.initialBal}
			}

			s := InitStats(&types.Stats{}, walletMap, &sync.RWMutex{}, mockNodeClient, mockDashboard, nil, 0)

			mockNodeClient.EXPECT().GetBalance(tt.wallet).Return(tt.fetchedBal)
			mockDashboard.EXPECT().BroadcastUpdate()

			s.UpdateSingleBalance(tt.wallet)

			if tt.walletExists {
				if walletMap[tt.wallet].ServerBalance != tt.expectedBal {
					t.Errorf("Expected balance %f, got %f", tt.expectedBal, walletMap[tt.wallet].ServerBalance)
				}
			}
		})
	}
}

func TestStats_GetDashboardData(t *testing.T) {
	tests := []struct {
		name           string
		wallets        []string
		walletMap      map[string]*types.WalletStats
		globalHash     float64
		sessionMined   uint32
		expectedHash   float64
		expectedBal    float64
		expectedLTB    uint32
		expectedWCount int
	}{
		{
			name:    "Multiple wallets",
			wallets: []string{"wallet1", "wallet2"},
			walletMap: map[string]*types.WalletStats{
				"wallet1": {Address: "wallet1", ServerBalance: 10.0, TotalMined: 5},
				"wallet2": {Address: "wallet2", ServerBalance: 20.0, TotalMined: 15},
			},
			globalHash:     5.5,
			sessionMined:   2,
			expectedHash:   5.5,
			expectedBal:    30.0,
			expectedLTB:    20,
			expectedWCount: 2,
		},
		{
			name:           "No wallets",
			wallets:        []string{},
			walletMap:      map[string]*types.WalletStats{},
			globalHash:     0.0,
			sessionMined:   0,
			expectedHash:   0.0,
			expectedBal:    0.0,
			expectedLTB:    0,
			expectedWCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockWallets := NewMockWallets(ctrl)

			st := &types.Stats{}
			st.GlobalHashrate.Store(tt.globalHash)
			st.SessionMined = tt.sessionMined
			st.StartTime = time.Now().Add(-1 * time.Hour)

			s := InitStats(st, tt.walletMap, &sync.RWMutex{}, nil, nil, mockWallets, 0)

			mockWallets.EXPECT().GetWallets().Return(tt.wallets)

			data := s.GetDashboardData()

			if data.Hashrate != tt.expectedHash {
				t.Errorf("Expected hashrate %f, got %f", tt.expectedHash, data.Hashrate)
			}
			if data.TotalBalance != tt.expectedBal {
				t.Errorf("Expected total balance %f, got %f", tt.expectedBal, data.TotalBalance)
			}
			if data.SessionBlocks != tt.sessionMined {
				t.Errorf("Expected session blocks %d, got %d", tt.sessionMined, data.SessionBlocks)
			}
			if data.LifetimeBlocks != tt.expectedLTB {
				t.Errorf("Expected lifetime blocks %d, got %d", tt.expectedLTB, data.LifetimeBlocks)
			}
			if len(data.Wallets) != tt.expectedWCount {
				t.Errorf("Expected %d wallets in dashboard data, got %d", tt.expectedWCount, len(data.Wallets))
			}
		})
	}
}
