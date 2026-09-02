package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv         string
	HTTPPort       string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
	JWTSecret      string
	JWTAccessExp   time.Duration
	JWTRefreshExp  time.Duration
	BrapiToken     string
	FinnhubToken   string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load(".env", "../.env")

	cfg := &Config{
		AppEnv:        getEnv("APP_ENV", "development"),
		HTTPPort:      getEnv("HTTP_PORT", "8080"),
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASS", "postgres123"),
		DBName:        getEnv("DB_NAME", "invest_db"),
		DBSSLMode:     getEnv("DB_SSLMODE", "disable"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       0,
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
