package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppPort           string
	DatabaseURL       string
	RedisURL          string
	APIAccessToken    string
	MetaVerifyToken   string
	MetaAppSecret     string
	MetaAPIVersion    string
	MetaAccessToken   string
	MetaPhoneNumberID string
	DefaultAccountID  int
	DefaultInboxID    int
	LogLevel          string
	WorkerConcurrency int
}

func Load() *Config {
	return &Config{
		AppPort:           getEnv("APP_PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/chatwoot_development?sslmode=disable"),
		RedisURL:          getEnv("REDIS_URL", "redis://localhost:6379/0"),
		APIAccessToken:    getEnv("API_ACCESS_TOKEN", ""),
		MetaVerifyToken:   getEnv("META_VERIFY_TOKEN", "whatsapp_gateway_verify_token"),
		MetaAppSecret:     getEnv("META_APP_SECRET", ""),
		MetaAPIVersion:    getEnv("META_API_VERSION", "v19.0"),
		MetaAccessToken:   getEnv("META_ACCESS_TOKEN", ""),
		MetaPhoneNumberID: getEnv("META_PHONE_NUMBER_ID", ""),
		DefaultAccountID:  getEnvAsInt("DEFAULT_ACCOUNT_ID", 1),
		DefaultInboxID:    getEnvAsInt("DEFAULT_INBOX_ID", 1),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		WorkerConcurrency: getEnvAsInt("WORKER_CONCURRENCY", 5),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}
