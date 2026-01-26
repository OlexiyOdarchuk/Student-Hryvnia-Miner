package backend

import (
	"encoding/json"
	"net/http"
	"shminer/backend/assets"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var clients = make(map[*websocket.Conn]bool)
var clientsMu sync.Mutex

func BroadcastUpdate() {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	if len(clients) == 0 {
		return
	}

	fullData := GetDashboardData()

	var safeWallets []map[string]interface{}
	for _, wallet := range fullData.Wallets {
		safeWallets = append(safeWallets, map[string]interface{}{
			"address":        wallet.Address,
			"name":           wallet.Name,
			"session_mined":  wallet.SessionMined,
			"total_mined":    wallet.TotalMined,
			"server_balance": wallet.ServerBalance,
			"working":        wallet.Working,
		})
	}

	response := map[string]interface{}{
		"hashrate":      fullData.Hashrate,
		"total_balance": fullData.TotalBalance,
		"uptime":        fullData.Uptime,
		"wallets":       safeWallets,
	}

	msg, err := json.Marshal(response)
	if err != nil {
		return
	}

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}

func StartWebServer() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		clientsMu.Lock()
		clients[conn] = true
		clientsMu.Unlock()
	})

	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		fullData := GetDashboardData()

		var safeWallets []map[string]interface{}
		for _, wallet := range fullData.Wallets {
			safeWallets = append(safeWallets, map[string]interface{}{
				"address":        wallet.Address,
				"name":           wallet.Name,
				"session_mined":  wallet.SessionMined,
				"total_mined":    wallet.TotalMined,
				"server_balance": wallet.ServerBalance,
				"working":        wallet.Working,
			})
		}

		response := map[string]interface{}{
			"hashrate":      fullData.Hashrate,
			"total_balance": fullData.TotalBalance,
			"uptime":        fullData.Uptime,
			"wallets":       safeWallets,
		}

		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/favicon.svg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Write(assets.Favicon)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(assets.WebUI)
	})

	go func() {
		if err := http.ListenAndServe(Config.ServerPort, nil); err != nil {
			PushLog("❌ Web server error: "+err.Error(), "error")
		}
	}()
	PushLog("🌐 API Server running at http://localhost"+Config.ServerPort, "info")
}
