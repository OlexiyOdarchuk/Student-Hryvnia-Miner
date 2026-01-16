package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env не знайдено, використовую system env")
	}

	LoadConfig()
	compileDifficultyBits(Config.Difficulty)

	startTime = time.Now()
	walletDataMap = make(map[string]*WalletStats)

	wallets = loadWalletsFromEnv()
	reloadWallets()

	setupGracefulShutdown()

	go watchEnvFile()
	go startWebServer()
	go speedMonitor()
	go balanceUpdater()

	fmt.Println("==================================================")
	fmt.Printf("🌐 ВЕБІНТФЕЙС: http://localhost%s\n", Config.ServerPort)
	fmt.Println("🔨 МАЙНЕР ЗАПУЩЕНО...")
	fmt.Println("==================================================")

	rand.Seed(time.Now().UnixNano())

	for {
		prevHash := getChainLastHash()
		if prevHash == "" {
			pushLog("⚠️ Немає зв'язку з сервером. Рестарт...", "error")
			time.Sleep(2 * time.Second)
			continue
		}

		ws := getWallets()
		if len(ws) == 0 {
			pushLog("⚠️ Немає гаманців, очікую .env", "error")
			time.Sleep(1 * time.Second)
			continue
		}
		currentWallet := ws[rand.Intn(len(ws))]

		success := mineBlock(prevHash, currentWallet)

		if success {
			dataMutex.Lock()
			sessionMined++
			if ws, ok := walletDataMap[currentWallet]; ok {
				ws.SessionMined++
			}
			dataMutex.Unlock()

			go updateSingleBalance(currentWallet)
		}

		time.Sleep(MinerSleepInterval)
	}
}

func setupGracefulShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		pushLog(fmt.Sprintf("🛑 Сигнал: %v. Завершення майнера...", sig), "info")
		fmt.Println("\n==================================================")
		fmt.Println("🛑 МАЙНЕР ЗУПИНЯЄТЬСЯ...")
		fmt.Println("==================================================")
		os.Exit(0)
	}()
}
