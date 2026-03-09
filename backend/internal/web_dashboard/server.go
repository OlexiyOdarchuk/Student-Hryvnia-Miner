package web_dashboard

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"shminer/backend/internal/types"
	"shminer/backend/internal/web_dashboard/assets"
	"sync"

	"github.com/gorilla/websocket"
)

type dashboardDataGetter interface {
	GetDashboardData() types.DashboardData
}

type Server struct {
	port      string // must begin with ":", for example ":8080"
	dashboard dashboardDataGetter
	clients   map[*websocket.Conn]bool
	clientsMu sync.Mutex
	upgrader  websocket.Upgrader
}

func NewServer(port string, dashboard dashboardDataGetter) *Server {
	return &Server{
		port:      port,
		dashboard: dashboard,
		clients:   make(map[*websocket.Conn]bool),
		clientsMu: sync.Mutex{},
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (s *Server) BroadcastUpdate() {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	if len(s.clients) == 0 {
		return
	}

	fullData := s.dashboard.GetDashboardData()

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

	for client := range s.clients {
		err := client.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			client.Close()
			delete(s.clients, client)
		}
	}
}

func (s *Server) StartWebServer() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		s.clientsMu.Lock()
		s.clients[conn] = true
		s.clientsMu.Unlock()
	})

	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		fullData := s.dashboard.GetDashboardData()

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
		if err := http.ListenAndServe(s.port, nil); err != nil {
			slog.Error("❌ Web server error: ", "error", err)
		}
	}()
	slog.Info("🌐 API Server is running", "port", s.port)
}
