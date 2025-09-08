package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                 string
	JWTSecret            string
	OTPExpirationMinutes int
	// ADD THESE TWO LINES
	StorageType string // "inmemory" or "postgres"
	DatabaseURL string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Error loading .env file (might be okay if running in container): %v", err)
	}

	cfg := &Config{
		Port:                 getEnv("PORT", "8080"),
		JWTSecret:            getEnv("JWT_SECRET", "default-jwt-secret"),
		OTPExpirationMinutes: getEnvAsInt("OTP_EXPIRATION_MINUTES", 2),
		// ADD THESE TWO LINES
		StorageType: strings.ToLower(getEnv("STORAGE_TYPE", "inmemory")),
		DatabaseURL: getEnv("DATABASE_URL", ""),
	}

	if cfg.StorageType == "postgres" && cfg.DatabaseURL == "" {
		log.Fatal("FATAL: STORAGE_TYPE is 'postgres' but DATABASE_URL is not set.")
	}

	if cfg.JWTSecret == "default-jwt-secret" {
		log.Println("WARNING: Using default JWT_SECRET. Please set a strong secret in .env or environment variables.")
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, strconv.Itoa(defaultValue))
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
