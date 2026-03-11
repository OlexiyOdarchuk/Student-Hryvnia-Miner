package nodeclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetChainLastHashCached_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		serverResponse interface{}
		expectedHash   string
	}{
		{
			name:   "Success",
			status: http.StatusOK,
			serverResponse: []Block{
				{
					ID:           "69af17ca1fdef3f7c2550d7a",
					PrevHash:     "000006c0771cc8366bceae66582bcf2dda7f49cd925be29b42236096df2a6696",
					Transactions: []Transaction{},
					Nonce:        596302,
					Miner:        "04b22cebe3c0085925e016647ba96e54282763dbcbcc149db52baa3aaef1b76826edcc3feee1eb0ac26acc09d6bc4f3f956ab91f14d2caca25c3402bee8712ab61",
					Reward:       1,
					Timestamp:    1773082569970,
					Hash:         "00000ajlfd2hknfrejngfajrsfo324fljafdsafajfjhbe29b422go34hglj43o2",
				},
			},
			expectedHash: "00000ajlfd2hknfrejngfajrsfo324fljafdsafajfjhbe29b422go34hglj43o2",
		},
		{
			name:           "Empty Block",
			status:         http.StatusOK,
			serverResponse: []Block{},
			expectedHash:   "",
		},
		{
			name:           "Err Server error",
			status:         http.StatusInternalServerError,
			serverResponse: nil,
			expectedHash:   "",
		},
		{
			name:           "Err Wrong JSON Structure",
			status:         http.StatusInternalServerError,
			serverResponse: map[string]string{"balance": "not-a-number"},
			expectedHash:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/chain" {
					assert.Error(t, fmt.Errorf("unexpected path: %s", r.URL.Path), "unexpected path")
				}
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			client := NewApiClient(server.URL, http.DefaultClient, 100*time.Millisecond, 200*time.Millisecond, 2)
			hash := client.GetChainLastHashCached()
			assert.Equal(t, tt.expectedHash, hash)
		})
	}

}

func TestGetBalance_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		serverResponse interface{}
		expectedBal    float64
	}{
		{
			name:           "Success",
			status:         http.StatusOK,
			serverResponse: map[string]float64{"balance": 100.5},
			expectedBal:    100.5,
		},
		{
			name:           "Err Server Error",
			status:         http.StatusInternalServerError,
			serverResponse: nil,
			expectedBal:    0,
		},
		{
			name:           "Err Wrong JSON Structure",
			status:         http.StatusOK,
			serverResponse: map[string]string{"balance": "not-a-number"},
			expectedBal:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			client := NewApiClient(server.URL, http.DefaultClient, 1*time.Millisecond, 2*time.Millisecond, 2)
			balance := client.GetBalance("test_wallet")

			assert.Equal(t, tt.expectedBal, balance)
		})
	}
}

func TestRetryLogic(t *testing.T) {
	attempts := 0
	expectedAttempts := 3
	expectedBalance := float64(400.12)
	mockResponse := map[string]float64{"balance": expectedBalance}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < expectedAttempts {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewApiClient(server.URL, http.DefaultClient, 1*time.Millisecond, 5*time.Millisecond, 5)
	balance := client.GetBalance("my-wallet")
	if balance != 400.12 {
		assert.Error(t, fmt.Errorf("expected %f, got %f", expectedBalance, balance))
	}
	if attempts != 3 {
		assert.Error(t, fmt.Errorf("expected %d, got %d", expectedAttempts, attempts), "expected 3, got %d", attempts)
	}
}
