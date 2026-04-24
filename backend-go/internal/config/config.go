package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                    string
	DatabasePath            string
	SecretKey               string
	CORSOrigins             string
	FrontendURL             string
	EmailAPIKey             string
	EmailFrom               string
	AccessTokenExpireMin    int
	RefreshTokenExpireDays  int
	RateLimitEnabled        bool
}

func Load() *Config {
	return &Config{
		Port:                    getEnv("PORT", "8080"),
		DatabasePath:            getEnv("DATABASE_PATH", "./data/alaya-archive.db"),
		SecretKey:               getEnv("SECRET_KEY", "change-me-in-production"),
		CORSOrigins:             getEnv("CORS_ORIGINS", "http://localhost:5173"),
		FrontendURL:             getEnv("FRONTEND_URL", "http://localhost:5173"),
		EmailAPIKey:             getEnv("EMAIL_API_KEY", ""),
		EmailFrom:               getEnv("EMAIL_FROM", ""),
		AccessTokenExpireMin:    getEnvInt("ACCESS_TOKEN_EXPIRE_MINUTES", 15),
		RefreshTokenExpireDays:  getEnvInt("REFRESH_TOKEN_EXPIRE_DAYS", 30),
		RateLimitEnabled:        getEnv("RATE_LIMIT_ENABLED", "true") == "true",
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}
