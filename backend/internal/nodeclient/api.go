package nodeclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"math"
	"net/http"
	"time"
)

type ApiClient struct {
	baseUrl      string
	httpClient   *http.Client
	retryDelay   time.Duration
	retryMax     int
	backoffLimit float64
}

func NewApiClient(baseUrl string, httpClient *http.Client, retryDelay, backoffLimit time.Duration, retryMax int) *ApiClient {
	return &ApiClient{
		baseUrl:      baseUrl,
		httpClient:   httpClient,
		retryMax:     retryMax,
		retryDelay:   retryDelay,
		backoffLimit: float64(backoffLimit),
	}
}

func (ac *ApiClient) GetChainLastHashCached() (string, error) {
	var result string
	err := ac.retryWithBackoff(func() error {
		resp, err := ac.httpClient.Get(ac.baseUrl + "/chain")
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
	if err != nil {
		return "", err
	}
	return result, err
}

func (ac *ApiClient) SubmitBlock(prev, wallet string, nonce int, ts int64, hash string) bool {
	payload := map[string]interface{}{
		"block": map[string]interface{}{
			"prevHash": prev, "transactions": []map[string]interface{}{{"from": nil, "to": wallet, "amount": 1}},
			"nonce": nonce, "miner": wallet, "reward": 1, "timestamp": ts, "hash": hash,
		},
	}
	body, _ := json.Marshal(payload)

	var success bool
	err := ac.retryWithBackoff(func() error {
		req, _ := http.NewRequest("POST", ac.baseUrl+"/submit-block", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := ac.httpClient.Do(req)
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
		slog.Error("❌ Помилка при відправці блоку", "error", err) // TODO: оновити логер, щоб slog відповідав за front тоже
		return false
	}
	return success
}

func (ac *ApiClient) GetBalance(addr string) float64 {
	var balance float64
	err := ac.retryWithBackoff(func() error {
		resp, err := ac.httpClient.Get(ac.baseUrl + "/balance/" + addr)
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
		slog.Error("❌ Помилка отримання балансу: ", "error", err)
		return 0
	}
	return balance
}

func (ac *ApiClient) exponentialBackoff(attempt int) time.Duration {
	delayMs := float64(ac.retryDelay.Milliseconds()) * math.Pow(2, float64(attempt))
	if delayMs > ac.backoffLimit {
		delayMs = ac.backoffLimit
	}
	return time.Duration(delayMs) * time.Millisecond
}

func (ac *ApiClient) retryWithBackoff(fn func() error) error {
	var lastErr error
	for attempt := 0; attempt < ac.retryMax; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err
		if attempt < ac.retryMax-1 {
			delay := ac.exponentialBackoff(attempt)
			time.Sleep(delay)
		}
	}
	return lastErr
}
