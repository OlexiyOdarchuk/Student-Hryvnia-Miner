package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"sync"
	"time"
)

var httpClient = &http.Client{
	Timeout: DefaultHTTPTimeout,
	Transport: &http.Transport{
		MaxIdleConns:        MaxIdleConnections,
		MaxIdleConnsPerHost: MaxIdleConnsPerHost,
		IdleConnTimeout:     IdleConnTimeout,
	},
}

var (
	cachedHashMutex sync.RWMutex
	cachedHash      string
	cachedHashTime  time.Time
	hashCacheTTL    = DefaultHashCacheTTL
)

func getChainLastHashCached() string {
	cachedHashMutex.RLock()
	if time.Since(cachedHashTime) < hashCacheTTL {
		defer cachedHashMutex.RUnlock()
		return cachedHash
	}
	cachedHashMutex.RUnlock()

	var result string
	err := retryWithBackoff(func() error {
		resp, err := httpClient.Get(Config.BaseURL + "/chain")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		var chain []map[string]interface{}
		json.Unmarshal(body, &chain)
		if len(chain) > 0 {
			result = chain[len(chain)-1]["hash"].(string)
		}
		return nil
	})

	if err == nil {
		cachedHashMutex.Lock()
		cachedHash = result
		cachedHashTime = time.Now()
		cachedHashMutex.Unlock()
	}
	return result
}

func exponentialBackoff(attempt int) time.Duration {
	base := Config.RetryDelay
	delayMs := float64(base.Milliseconds()) * math.Pow(2, float64(attempt))
	if delayMs > ExponentialBackoffMaxMs {
		delayMs = ExponentialBackoffMaxMs
	}
	return time.Duration(delayMs) * time.Millisecond
}

func retryWithBackoff(fn func() error) error {
	var lastErr error
	for attempt := 0; attempt < Config.MaxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err
		if attempt < Config.MaxRetries-1 {
			delay := exponentialBackoff(attempt)
			time.Sleep(delay)
		}
	}
	return lastErr
}

func SubmitBlock(prev, wallet string, nonce int, ts int64, hash string) bool {
	payload := map[string]interface{}{
		"block": map[string]interface{}{
			"prevHash": prev, "transactions": []map[string]interface{}{{"from": nil, "to": wallet, "amount": 1}},
			"nonce": nonce, "miner": wallet, "reward": 1, "timestamp": ts, "hash": hash,
		},
	}
	body, _ := json.Marshal(payload)

	var success bool
	err := retryWithBackoff(func() error {
		req, _ := http.NewRequest("POST", Config.BaseURL+"/submit-block", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 && resp.StatusCode != 201 {
			return fmt.Errorf("status %d", resp.StatusCode)
		}
		success = true
		return nil
	})

	if err != nil {
		PushLog("❌ Помилка при відправці блоку: "+err.Error(), "error")
		return false
	}
	return success
}

func GetBalance(addr string) float64 {
	var balance float64
	err := retryWithBackoff(func() error {
		resp, err := httpClient.Get(Config.BaseURL + "/balance/" + addr)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("status %d", resp.StatusCode)
		}
		body, _ := ioutil.ReadAll(resp.Body)
		var data struct {
			Balance float64 `json:"balance"`
		}
		if err := json.Unmarshal(body, &data); err != nil {
			return err
		}
		balance = data.Balance
		return nil
	})

	if err != nil {
		PushLog("❌ Помилка отримання балансу: "+err.Error(), "error")
		return 0
	}
	return balance
}

func SendTransaction(from, privateKey, to string, amount float64) (string, error) {
	payload := map[string]interface{}{
		"from":       from,
		"to":         to,
		"amount":     amount,
		"privateKey": privateKey,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", Config.BaseURL+"/transaction", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("server returned %d", resp.StatusCode)
	}

	return "Transaction sent", nil
}
