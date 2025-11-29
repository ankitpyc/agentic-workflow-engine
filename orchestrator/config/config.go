package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	RedisAddr   string
	RedisPass   string
	RedisDB     int
}

func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	dbHost := os.Getenv("DB_HOST")
	dbPortStr := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Fallback to docker-compose.yml defaults if .env not present
	if dbHost == "" {
		dbHost = "localhost" // Host is localhost from the perspective of the Orchestrator for direct connection
	}
	if dbPortStr == "" {
		dbPortStr = "5433" // Use the reconfigured port
	}
	if dbUser == "" {
		dbUser = "user"
	}
	if dbPassword == "" {
		dbPassword = "password"
	}
	if dbName == "" {
		dbName = "workflow_engine_db"
	}

	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	databaseURL := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default Redis address
	}
	redisPass := os.Getenv("REDIS_PASSWORD")
	redisDBStr := os.Getenv("REDIS_DB")
	redisDB := 0
	if redisDBStr != "" {
		redisDB, err = strconv.Atoi(redisDBStr)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
		}
	}

	return &Config{
		DatabaseURL: databaseURL,
		RedisAddr:   redisAddr,
		RedisPass:   redisPass,
		RedisDB:     redisDB,
	}, nil
}
