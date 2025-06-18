package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

// Config holds application settings loaded from environment variables
type Config struct {
	AlchemyAPIKey     string
	TelegramBotAPIKey string
	TelegramChatID    string
	WalletsFrom       []string
	WalletsTo         []string
	WindowSeconds     int
	CooldownSeconds   int
	ThresholdETH      float64
}

// Load reads and parses configuration from environment variables
func Load() Config {
	return Config{
		AlchemyAPIKey:     mustEnv("ALCHEMY_API_KEY"),
		TelegramBotAPIKey: mustEnv("TELEGRAM_BOT_API_KEY"),
		TelegramChatID:    mustEnv("TELEGRAM_CHAT_ID"),

		WalletsFrom: getEnvAsSlice("MONITORED_WALLETS_FROM", ","),
		WalletsTo:   getEnvAsSlice("MONITORED_WALLETS_TO", ","),

		WindowSeconds:   getEnvAsInt("AGGREGATION_WINDOW_IN_SECONDS", 300),
		CooldownSeconds: getEnvAsInt("AGGREGATION_NOTIFICATION_COOLDOWN_IN_SECONDS", 30),
		ThresholdETH:    getEnvAsFloat("THRESHOLD_ETH", 0.0),
	}
}

// --- Helpers ---

func mustEnv(key string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		log.Fatalf("Missing required environment variable: %s", key)
	}
	return val
}

func getEnvAsInt(key string, defaultVal int) int {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("Invalid int for %s: %v", key, err)
	}
	return i
}

func getEnvAsFloat(key string, defaultVal float64) float64 {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return defaultVal
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Fatalf("Invalid float for %s: %v", key, err)
	}
	return f
}

func getEnvAsSlice(key, sep string) []string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return nil
	}
	parts := strings.Split(val, sep)
	for i, p := range parts {
		parts[i] = strings.ToLower(strings.TrimSpace(p))
	}
	return parts
}
