package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port string
	Env  string

	// Database
	DatabaseURL string

	// JWT
	JWTSecret          string
	JWTExpiryHours     int
	RefreshExpiryHours int

	// App
	UploadDir    string
	MaxFileSize  int64
	FrontendURL  string
}

var App *Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	App = &Config{
		Port:               getEnv("PORT", "8080"),
		Env:                getEnv("ENV", "development"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/gogame?sslmode=disable"),
		JWTSecret:          getEnv("JWT_SECRET", "change-this-secret-in-production"),
		JWTExpiryHours:     getEnvInt("JWT_EXPIRY_HOURS", 24),
		RefreshExpiryHours: getEnvInt("REFRESH_EXPIRY_HOURS", 168),
		UploadDir:          getEnv("UPLOAD_DIR", "./uploads"),
		MaxFileSize:        int64(getEnvInt("MAX_FILE_SIZE_MB", 10)) * 1024 * 1024,
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:5173"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
