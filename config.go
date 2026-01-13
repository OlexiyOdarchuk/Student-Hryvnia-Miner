package main

import "sync"

var (
	wallets      []string
	walletsMutex sync.RWMutex
)

const (
	baseURL    = "https://s-hryvnia-1.onrender.com"
	difficulty = "00000"
	serverPort = ":8090"
)
