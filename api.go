package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"time"
)

func exponentialBackoff(attempt int) time.Duration {
	base := Config.RetryDelay
	delayMs := float64(base.Milliseconds()) * math.Pow(2, float64(attempt))
	if delayMs > 30000 {
		delayMs = 30000
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

func submitBlock(prev, wallet string, nonce int, ts int64, hash string) bool {
	payload := map[string]interface{}{
		"block": map[string]interface{}{
			"prevHash": prev, "transactions": []map[string]interface{}{{"from": nil, "to": wallet, "amount": 1}},
			"nonce": nonce, "miner": wallet, "reward": 1, "timestamp": ts, "hash": hash,
		},
	}
	body, _ := json.Marshal(payload)

	var success bool
	err := retryWithBackoff(func() error {
		client := &http.Client{Timeout: Config.HTTPTimeout}
		req, _ := http.NewRequest("POST", Config.BaseURL+"/submit-block", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		success = resp.StatusCode == 200 || resp.StatusCode == 201
		return nil
	})

	if err != nil && !success {
		pushLog("❌ Помилка при відправці блоку: "+err.Error(), "error")
	}
	return success
}

func getChainLastHash() string {
	var result string
	err := retryWithBackoff(func() error {
		client := &http.Client{Timeout: Config.HTTPTimeout}
		resp, err := client.Get(Config.BaseURL + "/chain")
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
		pushLog("❌ Помилка отримання хеша: "+err.Error(), "error")
	}
	return result
}

func getBalance(addr string) int {
	var balance int
	err := retryWithBackoff(func() error {
		client := &http.Client{Timeout: Config.HTTPTimeout}
		resp, err := client.Get(Config.BaseURL + "/balance/" + addr)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		var data struct {
			Balance int `json:"balance"`
		}
		json.Unmarshal(body, &data)
		balance = data.Balance
		return nil
	})

	if err != nil {
		pushLog("❌ Помилка отримання балансу: "+err.Error(), "error")
	}
	return balance
}
