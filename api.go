package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

func submitBlock(prev, wallet string, nonce int, ts int64, hash string) bool {
	payload := map[string]interface{}{
		"block": map[string]interface{}{
			"prevHash": prev, "transactions": []map[string]interface{}{{"from": nil, "to": wallet, "amount": 1}},
			"nonce": nonce, "miner": wallet, "reward": 1, "timestamp": ts, "hash": hash,
		},
	}
	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequest("POST", baseURL+"/submit-block", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200 || resp.StatusCode == 201
}

func getChainLastHash() string {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/chain")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var chain []map[string]interface{}
	json.Unmarshal(body, &chain)
	if len(chain) > 0 {
		return chain[len(chain)-1]["hash"].(string)
	}
	return ""
}

func getBalance(addr string) int {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/balance/" + addr)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var data struct {
		Balance int `json:"balance"`
	}
	json.Unmarshal(body, &data)
	return data.Balance
}
