package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	MongoURI      string
	MongoDB       string
	AdminPassword string
	JWTSecret     string
	APIToken      string
	Port          string // public API port
	AdminPort     string // admin API port (local)
	CORSOrigins   string // comma-separated; empty = allow any localhost origin
	// BridgeURL points at the local Telegram bridge service (bridge/server.py).
	// Empty (the default, and in production) disables the "Add from Telegram" UI.
	BridgeURL string
	// KeepAliveKey guards POST /keepalive (the weekly "pin" of manually-added videos).
	// Empty disables the endpoint.
	KeepAliveKey string
}

var C AppConfig

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("[config] no .env file, reading environment variables")
	}
	C = AppConfig{
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:       getEnv("MONGO_DB", "videoscraper"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
		JWTSecret:     getEnv("JWT_SECRET", "changeme"),
		APIToken:      getEnv("API_TOKEN", ""),
		Port:          getEnv("PORT", "8080"),
		AdminPort:     getEnv("ADMIN_PORT", "8081"),
		CORSOrigins:   getEnv("CORS_ORIGINS", ""),
		BridgeURL:     getEnv("BRIDGE_URL", ""),
		KeepAliveKey:  getEnv("KEEPALIVE_KEY", ""),
	}
	log.Printf("[config] loaded — mongo=%s db=%s", C.MongoURI, C.MongoDB)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
