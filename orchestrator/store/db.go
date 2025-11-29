package store

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"workflow-engine/config"

	_ "github.com/lib/pq" // PostgreSQL driver
)

type Store struct {
	db        *sql.DB
	Projects  *ProjectStore
	Personas  *PersonaStore
	StageRuns *StageRunStore
}

func NewStore(cfg *config.Config) (*Store, error) {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)                 // Max number of open connections
	db.SetMaxIdleConns(10)                 // Max number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Max lifetime of a connection

	// Ping the database to verify connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database!")

	return &Store{
		db:        db,
		Projects:  NewProjectStore(db),
		Personas:  NewPersonaStore(db),
		StageRuns: NewStageRunStore(db),
	}, nil
}

func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
