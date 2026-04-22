package web_dashboard

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"shminer/backend/internal/web_dashboard/assets"
	"shminer/backend/types"
	"sync"

	"github.com/gorilla/websocket"
)

type dashboardDataGetter interface {
	GetDashboardData() types.DashboardData
}

type wsClient struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (c *wsClient) writeMessage(msg []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, msg)
}

type Server struct {
	port      string
	dashboard dashboardDataGetter
	clients   map[*wsClient]struct{}
	clientsMu sync.Mutex
	upgrader  websocket.Upgrader
}

func NewServer(port string, dashboard dashboardDataGetter) *Server {
	return &Server{
		port:      port,
		dashboard: dashboard,
		clients:   make(map[*wsClient]struct{}),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func buildResponse(data types.DashboardData) map[string]any {
	safeWallets := make([]map[string]any, 0, len(data.Wallets))
	for _, wallet := range data.Wallets {
		safeWallets = append(safeWallets, map[string]any{
			"address":        wallet.Address,
			"name":           wallet.Name,
			"session_mined":  wallet.SessionMined,
			"total_mined":    wallet.TotalMined,
			"server_balance": wallet.ServerBalance,
			"working":        wallet.Working,
		})
	}
	return map[string]any{
		"hashrate":         data.Hashrate,
		"total_balance":    data.TotalBalance,
		"uptime":           data.Uptime,
		"wallets":          safeWallets,
		"session_blocks":   data.SessionBlocks,
		"lifetime_blocks":  data.LifetimeBlocks,
		"session_found":    data.SessionFound,
		"submit_queue_len": data.SubmitQueueLen,
		"blocks_per_min":   data.BlocksPerMin,
		"found_per_min":    data.FoundPerMin,
		"is_mining":        data.IsMining,
	}
}

func (s *Server) BroadcastUpdate() {
	s.clientsMu.Lock()
	if len(s.clients) == 0 {
		s.clientsMu.Unlock()
		return
	}
	snapshot := make([]*wsClient, 0, len(s.clients))
	for c := range s.clients {
		snapshot = append(snapshot, c)
	}
	s.clientsMu.Unlock()

	fullData := s.dashboard.GetDashboardData()
	msg, err := json.Marshal(buildResponse(fullData))
	if err != nil {
		return
	}

	var failed []*wsClient
	for _, client := range snapshot {
		if err := client.writeMessage(msg); err != nil {
			client.conn.Close()
			failed = append(failed, client)
		}
	}

	if len(failed) == 0 {
		return
	}
	s.clientsMu.Lock()
	for _, c := range failed {
		delete(s.clients, c)
	}
	s.clientsMu.Unlock()
}

func (s *Server) StartWebServer() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := &wsClient{conn: conn}
		s.clientsMu.Lock()
		s.clients[client] = struct{}{}
		s.clientsMu.Unlock()
	})

	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(buildResponse(s.dashboard.GetDashboardData()))
	})

	http.HandleFunc("/favicon.svg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Write(assets.Favicon)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(assets.WebUI)
	})

	addr := s.port
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			slog.Error("❌ Web server error: ", "error", err)
		}
	}()
	slog.Info("🌐 API Server listening on LAN", "addr", addr)
}
