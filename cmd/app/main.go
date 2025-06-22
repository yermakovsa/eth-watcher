package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	"github.com/yermakovsa/alchemyws"
	"github.com/yermakovsa/eth-watcher/internal/aggregator"
	"github.com/yermakovsa/eth-watcher/internal/config"
	"github.com/yermakovsa/eth-watcher/internal/notifier"
	"github.com/yermakovsa/eth-watcher/internal/watcher"
)

func main() {
	// Load .env file (optional, non-fatal)
	_ = godotenv.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Load application config
	cfg := config.Load()

	// Initialize services
	bot := mustInitTelegramBot(cfg.TelegramBotAPIKey)
	chatID := mustParseChatID(cfg.TelegramChatID)
	notif := notifier.NewTelegramNotifier(bot, chatID)

	agg := aggregator.NewAggregator(
		ctx,
		notif,
		cfg.ThresholdETH,
		time.Duration(cfg.WindowSeconds)*time.Second,
		time.Duration(cfg.CooldownSeconds)*time.Second,
	)

	client, err := alchemyws.NewAlchemyClient(cfg.AlchemyAPIKey, nil)
	if err != nil {
		log.Fatalf("[Main] Failed to initialize Alchemy client: %v", err)
	}

	w := watcher.NewWatcher(ctx, client, cfg.WalletsFrom, cfg.WalletsTo, agg)

	log.Println("[Main] Starting transaction watcher...")
	if err := w.Start(); err != nil {
		log.Fatalf("[Main] Watcher failed to start: %v", err)
	}

	// Wait for termination signal
	<-sigChan
	log.Println("[Main] Shutdown signal received. Cleaning up...")
	w.Stop()
}

// mustInitTelegramBot initializes the Telegram bot or exits on failure.
func mustInitTelegramBot(apiKey string) *telego.Bot {
	bot, err := telego.NewBot(apiKey, telego.WithDiscardLogger())
	if err != nil {
		log.Fatalf("[Telegram] Failed to initialize bot: %v", err)
	}
	return bot
}

// mustParseChatID converts Telegram chat ID from string to int64 or exits on failure.
func mustParseChatID(chatIDStr string) int64 {
	id, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		log.Fatalf("[Telegram] Invalid chat ID '%s': %v", chatIDStr, err)
	}
	return id
}
