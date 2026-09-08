package config

import (
	"os"
	"time"
)

type Config struct {
	AppEnv        string
	HTTPPort      string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBSSLMode     string
	JWTSecret     string
	JWTAccessExp  time.Duration
	JWTRefreshExp time.Duration
	BrapiToken    string
	FinnhubToken  string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		AppEnv:        getEnv("APP_ENV", "development"),
		HTTPPort:      getEnv("HTTP_PORT", "8081"),
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5433"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASS", "postgres123"),
		DBName:        getEnv("DB_NAME", "invest_db"),
		DBSSLMode:     getEnv("DB_SSLMODE", "disable"),
		JWTSecret:     getEnv("JWT_SECRET", "supersecret-jwt-key-for-dev-only-32chars"),
		JWTAccessExp:  15 * time.Minute,
		JWTRefreshExp: 7 * 24 * time.Hour,
		BrapiToken:    getEnv("BRAPI_TOKEN", ""),
		FinnhubToken:  getEnv("FINNHUB_TOKEN", ""),
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
