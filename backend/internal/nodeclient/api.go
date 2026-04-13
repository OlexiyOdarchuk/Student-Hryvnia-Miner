package nodeclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"log/slog"
	"math"
	"net/http"
	"shminer/backend/types"
	"strconv"
	"time"
)

type NodeClient interface {
	GetChainLastHashCached() string
	SubmitBlock(prev, wallet string, nonce int, ts int64, hash string) bool
	GetBalance(addr string) float64
	SendTransaction(tx types.TxPayload) error
}

type ApiClient struct {
	baseUrl      string
	httpClient   *http.Client
	retryDelay   time.Duration
	retryMax     int
	backoffLimit float64
}

type Transaction struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int    `json:"amount"`
}

type Block struct {
	ID           string        `json:"_id"`
	PrevHash     string        `json:"prevHash"`
	Transactions []Transaction `json:"transactions"`
	Nonce        int           `json:"nonce"`
	Miner        string        `json:"miner"`
	Reward       int           `json:"reward"`
	Timestamp    int64         `json:"timestamp"`
	Hash         string        `json:"hash"`
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

func (ac *ApiClient) GetChainLastHashCached() string {

	var result string
	err := ac.retryWithBackoff(func() error {
		resp, err := ac.httpClient.Get(ac.baseUrl + "/chain")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		var chain []Block
		if err = json.NewDecoder(resp.Body).Decode(&chain); err != nil {
			return err
		}
		if len(chain) == 0 {
			return errors.New("chain response empty")
		}
		io.Copy(io.Discard, resp.Body)
		result = chain[len(chain)-1].Hash
		return nil
	})
	if err != nil {
		slog.Error("❌ Помилка при отриманні останнього блоку", "err", err)
	}
	return result
}

func (ac *ApiClient) SubmitBlock(prev, wallet string, nonce int, ts int64, hash string) bool {
	body := buildBody(prev, wallet, nonce, ts, hash)

	req, _ := http.NewRequest("POST", ac.baseUrl+"/submit-block", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return false
	}

	return true
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
			return errors.New("stats " + strconv.Itoa(resp.StatusCode))
		}
		body, _ := io.ReadAll(resp.Body)
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

func (ac *ApiClient) SendTransaction(tx types.TxPayload) error {
	finalPayload, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	return ac.retryWithBackoff(func() error {
		req, err := http.NewRequest("POST", ac.baseUrl+"/transaction", bytes.NewBuffer(finalPayload))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := ac.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			var errResp struct {
				Message string `json:"message"`
				Error   string `json:"error"`
			}
			if json.Unmarshal(body, &errResp) == nil {
				if errResp.Message != "" {
					return errors.New(errResp.Message)
				}
				if errResp.Error != "" {
					return errors.New(errResp.Error)
				}
			}
			return errors.New("Server rejected the transaction: " + string(body))
		}
		io.Copy(io.Discard, resp.Body)
		return nil
	})
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

func buildBody(prev, wallet string, nonce int, ts int64, hash string) []byte {
	buf := make([]byte, 0, 256)

	buf = append(buf, `{"block":{`...)

	buf = append(buf, `"prevHash":"`...)
	buf = append(buf, prev...)
	buf = append(buf, `",`...)

	buf = append(buf, `"transactions":[{"from":"","to":"`...)
	buf = append(buf, wallet...)
	buf = append(buf, `","amount":1}],`...)

	buf = append(buf, `"nonce":`...)
	buf = strconv.AppendInt(buf, int64(nonce), 10)
	buf = append(buf, `,`...)

	buf = append(buf, `"miner":"`...)
	buf = append(buf, wallet...)
	buf = append(buf, `",`...)

	buf = append(buf, `"reward":1,`...)

	buf = append(buf, `"timestamp":`...)
	buf = strconv.AppendInt(buf, ts, 10)
	buf = append(buf, `,`...)

	buf = append(buf, `"hash":"`...)
	buf = append(buf, hash...)
	buf = append(buf, `"}}`...)

	return buf
}
