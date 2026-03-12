package wallets

import (
	"shminer/backend/types"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestWallets_AddWalletSafe(t *testing.T) {
	tests := []struct {
		name          string
		initialMap    map[string]*types.WalletStats
		addName       string
		addAddr       string
		addPriv       string
		expectError   bool
		expectedCount int
	}{
		{
			name:          "Add new wallet",
			initialMap:    map[string]*types.WalletStats{},
			addName:       "Wallet 1",
			addAddr:       "addr1",
			addPriv:       "priv1",
			expectError:   false,
			expectedCount: 1,
		},
		{
			name: "Add existing address",
			initialMap: map[string]*types.WalletStats{
				"addr1": {Address: "addr1", Name: "Existing"},
			},
			addName:       "Wallet 2",
			addAddr:       "addr1",
			addPriv:       "priv2",
			expectError:   false, // Returns nil if address exists
			expectedCount: 1,
		},
		{
			name: "Add existing name",
			initialMap: map[string]*types.WalletStats{
				"addr1": {Address: "addr1", Name: "Wallet 1"},
			},
			addName:       "Wallet 1",
			addAddr:       "addr2",
			addPriv:       "priv2",
			expectError:   true,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := NewMockStorage(ctrl)
			mockStorage.EXPECT().UpdateWallets(gomock.Any()).AnyTimes()
			mockStorage.EXPECT().GetSessionPassword().Return("pass").AnyTimes()
			mockStorage.EXPECT().GetStorage().Return(types.StorageData{}).AnyTimes()
			mockStorage.EXPECT().SaveStorage(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

			w := New(mockStorage, &sync.RWMutex{}, tt.initialMap)
			for addr := range tt.initialMap {
				w.Wallets = append(w.Wallets, addr)
			}

			err := w.AddWalletSafe(tt.addName, tt.addAddr, tt.addPriv)

			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}

			if len(w.Wallets) != tt.expectedCount {
				t.Errorf("Expected %d wallets, got %d", tt.expectedCount, len(w.Wallets))
			}
		})
	}
}

func TestWallets_DeleteWallet(t *testing.T) {
	tests := []struct {
		name          string
		initialMap    map[string]*types.WalletStats
		delAddr       string
		expectedCount int
	}{
		{
			name: "Delete existing",
			initialMap: map[string]*types.WalletStats{
				"addr1": {Address: "addr1", Name: "W1"},
				"addr2": {Address: "addr2", Name: "W2"},
			},
			delAddr:       "addr1",
			expectedCount: 1,
		},
		{
			name: "Delete non-existing",
			initialMap: map[string]*types.WalletStats{
				"addr1": {Address: "addr1", Name: "W1"},
			},
			delAddr:       "addr2",
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := NewMockStorage(ctrl)
			mockStorage.EXPECT().UpdateWallets(gomock.Any()).AnyTimes()
			mockStorage.EXPECT().GetSessionPassword().Return("pass").AnyTimes()
			mockStorage.EXPECT().GetStorage().Return(types.StorageData{}).AnyTimes()
			mockStorage.EXPECT().SaveStorage(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

			w := New(mockStorage, &sync.RWMutex{}, tt.initialMap)
			for addr := range tt.initialMap {
				w.Wallets = append(w.Wallets, addr)
			}

			err := w.DeleteWallet(tt.delAddr)
			if err != nil {
				t.Errorf("Delete failed: %v", err)
			}

			if len(w.Wallets) != tt.expectedCount {
				t.Errorf("Expected %d wallets after delete, got %d", tt.expectedCount, len(w.Wallets))
			}

			if _, ok := tt.initialMap[tt.delAddr]; ok {
				t.Errorf("Wallet map should not contain deleted address")
			}
		})
	}
}

func TestWallets_ToggleWalletMining(t *testing.T) {
	tests := []struct {
		name          string
		initialMap    map[string]*types.WalletStats
		toggleAddr    string
		expectedState bool
	}{
		{
			name: "Toggle true to false",
			initialMap: map[string]*types.WalletStats{
				"addr1": {Address: "addr1", Working: true},
			},
			toggleAddr:    "addr1",
			expectedState: false,
		},
		{
			name: "Toggle false to true",
			initialMap: map[string]*types.WalletStats{
				"addr1": {Address: "addr1", Working: false},
			},
			toggleAddr:    "addr1",
			expectedState: true,
		},
		{
			name: "Toggle non-existing",
			initialMap: map[string]*types.WalletStats{
				"addr1": {Address: "addr1", Working: true},
			},
			toggleAddr:    "addr2",
			expectedState: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := NewMockStorage(ctrl)
			mockStorage.EXPECT().UpdateWallets(gomock.Any()).AnyTimes()
			mockStorage.EXPECT().GetSessionPassword().Return("pass").AnyTimes()
			mockStorage.EXPECT().GetStorage().Return(types.StorageData{}).AnyTimes()
			mockStorage.EXPECT().SaveStorage(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

			w := New(mockStorage, &sync.RWMutex{}, tt.initialMap)
			for addr := range tt.initialMap {
				w.Wallets = append(w.Wallets, addr)
			}

			state := w.ToggleWalletMining(tt.toggleAddr)
			if state != tt.expectedState {
				t.Errorf("Expected state %t, got %t", tt.expectedState, state)
			}

			if stats, ok := tt.initialMap[tt.toggleAddr]; ok {
				if stats.Working != tt.expectedState {
					t.Errorf("Expected map working state %t, got %t", tt.expectedState, stats.Working)
				}
			}
		})
	}
}

func TestWallets_SetAllMining(t *testing.T) {
	tests := []struct {
		name       string
		initialMap map[string]*types.WalletStats
		setState   bool
	}{
		{
			name: "Set all true",
			initialMap: map[string]*types.WalletStats{
				"addr1": {Address: "addr1", Working: false},
				"addr2": {Address: "addr2", Working: false},
			},
			setState: true,
		},
		{
			name: "Set all false",
			initialMap: map[string]*types.WalletStats{
				"addr1": {Address: "addr1", Working: true},
				"addr2": {Address: "addr2", Working: true},
			},
			setState: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := NewMockStorage(ctrl)
			mockStorage.EXPECT().UpdateWallets(gomock.Any()).AnyTimes()
			mockStorage.EXPECT().GetSessionPassword().Return("pass").AnyTimes()
			mockStorage.EXPECT().GetStorage().Return(types.StorageData{}).AnyTimes()
			mockStorage.EXPECT().SaveStorage(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

			w := New(mockStorage, &sync.RWMutex{}, tt.initialMap)
			for addr := range tt.initialMap {
				w.Wallets = append(w.Wallets, addr)
			}

			err := w.SetAllMining(tt.setState)
			if err != nil {
				t.Errorf("SetAllMining failed: %v", err)
			}

			for _, v := range tt.initialMap {
				if v.Working != tt.setState {
					t.Errorf("Expected all states to be %t", tt.setState)
				}
			}
		})
	}
}
