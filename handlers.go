package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
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
		Address    string `json:"address"`
		PrivateKey string `json:"private_key"` // Необов'язкове поле
		Password   string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request format",
		})
		return
	}

	req.Address = strings.TrimSpace(req.Address)
	req.PrivateKey = strings.TrimSpace(req.PrivateKey)
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
	dataMutex.Lock()
	if existing, exists := walletDataMap[req.Address]; exists {
		if req.PrivateKey != "" && existing.PrivateKey == "" {
			existing.PrivateKey = req.PrivateKey
			dataMutex.Unlock()
			saveWallets()
			pushLog("🔑 Ключі оновлено для: "+req.Address[:8]+"...", "success")
			json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Приватний ключ додано до існуючого гаманця"})
			return
		}

		dataMutex.Unlock()
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Вже існує"})
		return
	}

	newWallet := &WalletStats{
		Address:       req.Address,
		PrivateKey:    req.PrivateKey,
		Name:          "Worker",
		Working:       true,
		SessionMined:  0,
		ServerBalance: 0,
	}
	walletDataMap[req.Address] = newWallet
	wallets = append(wallets, req.Address)
	dataMutex.Unlock()

	saveWallets()

	msg := "➕ Додано (Тільки для майнингу): "
	if req.PrivateKey != "" {
		msg = "➕ Додано (Повний доступ): "
	}
	pushLog(msg+req.Address[:8]+"...", "success")

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
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

	dataMutex.Lock()
	stats, exists := walletDataMap[req.Address]
	if exists {
		stats.Name = req.Name
	}
	dataMutex.Unlock()

	if !exists {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Гаманець не знайдено",
		})
		return
	}

	saveWallets()

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
	dataMutex.Lock()
	if _, exists := walletDataMap[req.Address]; !exists {
		dataMutex.Unlock()
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Не знайдено"})
		return
	}

	delete(walletDataMap, req.Address)

	newWalletsList := []string{}
	for _, addr := range wallets {
		if addr != req.Address {
			newWalletsList = append(newWalletsList, addr)
		}
	}
	wallets = newWalletsList
	dataMutex.Unlock()

	saveWallets()

	pushLog("🗑️ Видалено: "+req.Address[:8]+"...", "info")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func handleToggleWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Address string `json:"address"`
		Working bool   `json:"working"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request format",
		})
		return
	}

	req.Address = strings.TrimSpace(req.Address)

	dataMutex.Lock()
	stats, ok := walletDataMap[req.Address]
	if ok {
		stats.Working = req.Working
	}
	dataMutex.Unlock()

	if !ok {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Гаманець не знайдено",
		})
		return
	}

	statusIcon := "▶️"
	statusText := "відновлено"
	if !req.Working {
		statusIcon = "⏸️"
		statusText = "призупинено"
	}
	saveWallets()
	pushLog(fmt.Sprintf("%s Майнінг %s для: %s...", statusIcon, statusText, req.Address[:12]), "info")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Статус оновлено",
		"working": req.Working,
	})
}

func setupRoutes() {
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/api/hashrate-history", handleHashrateHistory)
	http.HandleFunc("/events", handleEvents)
	http.HandleFunc("/api/add-wallet", handleAddWallet)
	http.HandleFunc("/api/rename-wallet", handleRenameWallet)
	http.HandleFunc("/api/delete-wallet", handleDeleteWallet)
	http.HandleFunc("/api/toggle-wallet", handleToggleWallet)
}
