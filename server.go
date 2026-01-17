package main

import (
	"net/http"

	_ "embed"
)

//go:embed index.html
var htmlPage string

func startWebServer() {
	setupRoutes()

	pushLog("🌐 Вебсервер запущено на http://localhost"+Config.ServerPort, "info")
	if err := http.ListenAndServe(Config.ServerPort, nil); err != nil {
		pushLog("❌ Помилка вебсервера: "+err.Error(), "error")
	}
}
