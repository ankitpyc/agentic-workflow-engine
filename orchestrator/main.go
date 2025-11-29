package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"workflow-engine/config"
	"workflow-engine/store"

	"github.com/go-redis/redis/v8" // Import Redis client
)

const (
	ProjectCreatedChannel = "project_created_events"
)

type Orchestrator struct {
	cfg         *config.Config
	dbStore     *store.Store
	redisClient *redis.Client
}

func NewOrchestrator(cfg *config.Config, dbStore *store.Store, redisClient *redis.Client) *Orchestrator {
	return &Orchestrator{
		cfg:         cfg,
		dbStore:     dbStore,
		redisClient: redisClient,
	}
}

func (o *Orchestrator) Run() {
	log.Println("Orchestrator service starting...")

	// Subscribe to project_created events
	go o.subscribeToProjectCreatedEvents()

	// Keep the service running until an interrupt signal is received
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan
	log.Println("Orchestrator service shutting down gracefully...")
}

func (o *Orchestrator) subscribeToProjectCreatedEvents() {
	pubsub := o.redisClient.Subscribe(context.Background(), ProjectCreatedChannel)
	defer pubsub.Close()

	log.Printf("Subscribed to Redis channel: %s", ProjectCreatedChannel)

	for msg := range pubsub.Channel() {
		log.Printf("Received message on channel %s: %s", msg.Channel, msg.Payload)
		// Process the event in a goroutine to avoid blocking the subscriber
		go o.handleProjectCreatedEvent(context.Background(), msg.Payload)
	}
}

func (o *Orchestrator) handleProjectCreatedEvent(ctx context.Context, payload string) {
	// Task 3.6: Implement project_created event handler
	log.Printf("Handling project_created event: %s", payload)

	// Placeholder for actual event parsing and database storage
	// This will be fully implemented in Task 3.6
	project := &store.Project{
		Name:        fmt.Sprintf("Project from Event: %s", payload),
		Description: sql.NullString{String: "Created via event bus", Valid: true},
	}
	err := o.dbStore.Projects.CreateProject(ctx, project)
	if err != nil {
		log.Printf("Error creating project from event payload %s: %v", payload, err)
	} else {
		log.Printf("Project %s (%s) created successfully from event.", project.Name, project.ID)
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	dbStore, err := store.NewStore(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database store: %v", err)
	}
	defer func() {
		if err := dbStore.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

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

	orchestrator := NewOrchestrator(cfg, dbStore, redisClient)
	orchestrator.Run()
}
