package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write([]byte(htmlPage))
}

func handleHashrateHistory(w http.ResponseWriter, r *http.Request) {
	hashrateHistMutex.Lock()
	history := make([]float64, len(hashrateHistory))
	copy(history, hashrateHistory[:])
	pos := hashrateHistPos
	hashrateHistMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"history":  history,
		"position": pos,
	})
}

func handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	clientLogCursor := int64(0)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			response := prepareEventData(&clientLogCursor)

			jsonData, _ := json.Marshal(response)
			_, err := w.Write([]byte("data: "))
			if err != nil {
				return
			}
			_, err = w.Write(jsonData)
			if err != nil {
				return
			}
			_, err = w.Write([]byte("\n\n"))
			if err != nil {
				return
			}

			flusher.Flush()
		}
	}
}

func prepareEventData(clientLogCursor *int64) DashboardData {
	dataMutex.RLock()

	var walletsExport []WalletStats
	totalBal := 0

	for _, wAddr := range getWallets() {
		if stats, ok := walletDataMap[wAddr]; ok {
			totalBal += stats.ServerBalance
			stats.Name = getWalletName(wAddr)
			walletsExport = append(walletsExport, *stats)
		}
	}

	hashRate := float64(0)
	if val := globalHashrate.Load(); val != nil {
		hashRate = val.(float64)
	}

	sessionMined := sessionMined

	dataMutex.RUnlock()

	var newLogs []LogEntry
	logRing.mu.Lock()

	start := 0
	if logRing.pos > 100 {
		start = logRing.pos - 100
	}

	for i := start; i < logRing.pos; i++ {
		log := logRing.data[i%100]
		if log.ID > *clientLogCursor {
			newLogs = append(newLogs, log)
			*clientLogCursor = log.ID
		}
	}

	logRing.mu.Unlock()

	return DashboardData{
		Hashrate:     hashRate,
		TotalBlocks:  sessionMined,
		Uptime:       time.Since(startTime).Round(time.Second).String(),
		TotalBalance: totalBal,
		Wallets:      walletsExport,
		NewLogs:      newLogs,
	}
}

func handleAddWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Address  string `json:"address"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request format",
		})
		return
	}

	req.Address = strings.TrimSpace(req.Address)
	req.Password = strings.TrimSpace(req.Password)

	if len(req.Address) < 20 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Адреса повинна бути щонайменше 20 символів",
		})
		return
	}

	if len(req.Password) < 6 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Пароль повинен бути щонайменше 6 символів",
		})
		return
	}
	existingWallets := getWallets()
	for _, wallet := range existingWallets {
		if wallet == req.Address {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Цей гаманець вже додано",
			})
			return
		}
	}
	if Config.AdminPassword != "" {
		if req.Password != Config.AdminPassword {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Неправильний пароль",
			})
			return
		}
	}
	envPath := ".env"
	envContent, _ := ioutil.ReadFile(envPath)
	envStr := string(envContent)
	lines := strings.Split(envStr, "\n")
	found := false
	var newLines []string

	for _, line := range lines {
		if strings.HasPrefix(line, "WALLETS=") {
			currentWallets := strings.TrimPrefix(line, "WALLETS=")
			currentWallets = strings.Trim(currentWallets, "\"")
			if currentWallets != "" {
				newLines = append(newLines, "WALLETS=\""+currentWallets+","+req.Address+"\"")
			} else {
				newLines = append(newLines, "WALLETS=\""+req.Address+"\"")
			}
			found = true
		} else if line != "" {
			newLines = append(newLines, line)
		}
	}

	if !found {
		newLines = append(newLines, "WALLETS=\""+req.Address+"\"")
	}

	newEnvStr := strings.Join(newLines, "\n")
	if err := ioutil.WriteFile(envPath, []byte(newEnvStr), 0644); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Помилка при збереженні конфіга: " + err.Error(),
		})
		return
	}

	os.Unsetenv("WALLETS")
	godotenv.Load(envPath)
	reloadWallets()

	pushLog("✅ Гаманець додано: "+req.Address[:12]+"...", "success")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Гаманець успішно додано!",
	})
}

func handleRenameWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Address string `json:"address"`
		Name    string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request format",
		})
		return
	}

	req.Address = strings.TrimSpace(req.Address)
	req.Name = strings.TrimSpace(req.Name)

	if req.Name == "" || len(req.Name) > 50 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Назва повинна бути від 1 до 50 символів",
		})
		return
	}

	setWalletName(req.Address, req.Name)
	pushLog("✏️ Гаманець перейменовано: "+req.Name, "success")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Назва оновлена",
	})
}

func handleDeleteWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Address  string `json:"address"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request format",
		})
		return
	}
	if Config.AdminPassword == "" || req.Password != Config.AdminPassword {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Неправильний пароль",
		})
		return
	}
	envPath := ".env"
	envContent, _ := ioutil.ReadFile(envPath)
	envStr := string(envContent)
	lines := strings.Split(envStr, "\n")
	var newLines []string

	for _, line := range lines {
		if strings.HasPrefix(line, "WALLETS=") {
			currentWallets := strings.TrimPrefix(line, "WALLETS=")
			currentWallets = strings.Trim(currentWallets, "\"")
			var walletList []string
			for _, w := range strings.Split(currentWallets, ",") {
				w = strings.TrimSpace(w)
				if w != "" && w != req.Address {
					walletList = append(walletList, w)
				}
			}
			if len(walletList) > 0 {
				newLines = append(newLines, "WALLETS=\""+strings.Join(walletList, ",")+"\"")
			}
		} else if line != "" {
			newLines = append(newLines, line)
		}
	}

	newEnvStr := strings.Join(newLines, "\n")
	if err := ioutil.WriteFile(envPath, []byte(newEnvStr), 0644); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Помилка при збереженні конфіга: " + err.Error(),
		})
		return
	}

	os.Unsetenv("WALLETS")
	godotenv.Load(envPath)
	deleteWalletName(req.Address)
	reloadWallets()

	pushLog("🗑️ Гаманець видалено", "success")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Гаманець видалено",
	})
}

func setupRoutes() {
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/api/hashrate-history", handleHashrateHistory)
	http.HandleFunc("/events", handleEvents)
	http.HandleFunc("/api/add-wallet", handleAddWallet)
	http.HandleFunc("/api/rename-wallet", handleRenameWallet)
	http.HandleFunc("/api/delete-wallet", handleDeleteWallet)
}
