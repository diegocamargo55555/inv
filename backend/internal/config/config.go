package config

import (
	"bufio"
	"os"
	"strings"
	"time"
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

func loadDotEnv(paths ...string) {
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
				if os.Getenv(key) == "" {
					_ = os.Setenv(key, val)
				}
			}
		}
		f.Close()
	}
}

func LoadConfig() (*Config, error) {
	loadDotEnv(".env", "../.env")

	cfg := &Config{
		AppEnv:        getEnv("APP_ENV", "development"),
		HTTPPort:      getEnv("HTTP_PORT", "8081"),
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5433"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASS", "postgres123"),
		DBName:        getEnv("DB_NAME", "invest_db"),
		DBSSLMode:     getEnv("DB_SSLMODE", "disable"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6380"),
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
