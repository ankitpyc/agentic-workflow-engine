package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"workflow-engine/config"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

const (
	ProjectCreatedChannel = "project_created_events"
)

func main() {
	err := godotenv.Load("../../.env") // Load .env for Redis config
	if err != nil {
		log.Printf("No .env file found, using defaults or system env: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Successfully connected to Redis!")

	// Generate a mock project ID and name
	mockProjectID := uuid.New().String()
	mockProjectName := fmt.Sprintf("Auto-generated Project %s", mockProjectID[:8])
	eventPayload := fmt.Sprintf(`{"id": "%s", "name": "%s", "description": "This project was created by a mock event publisher."}`, mockProjectID, mockProjectName)

	err = redisClient.Publish(context.Background(), ProjectCreatedChannel, eventPayload).Err()
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}

	log.Printf("Published message to channel '%s': %s", ProjectCreatedChannel, eventPayload)
	log.Println("Mock event publisher finished.")
}
