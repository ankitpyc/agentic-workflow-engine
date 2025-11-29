package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"encoding/json"

	"github.com/google/uuid"
)

type Persona struct {
	ID             uuid.UUID       `json:"id"`
	Name           string          `json:"name"`
	Description    sql.NullString  `json:"description"`
	PromptTemplate string          `json:"prompt_template"`
	ModelConfig    json.RawMessage `json:"model_config"` // JSONB type
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type PersonaStore struct {
	db *sql.DB
}

func NewPersonaStore(db *sql.DB) *PersonaStore {
	return &PersonaStore{db: db}
}

func (s *PersonaStore) CreatePersona(ctx context.Context, persona *Persona) error {
	persona.ID = uuid.New()
	persona.CreatedAt = time.Now()
	persona.UpdatedAt = time.Now()

	query := `
		INSERT INTO personas (persona_id, name, description, prompt_template, model_config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := s.db.ExecContext(ctx, query,
		persona.ID,
		persona.Name,
		persona.Description,
		persona.PromptTemplate,
		persona.ModelConfig,
		persona.CreatedAt,
		persona.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create persona: %w", err)
	}
	return nil
}

func (s *PersonaStore) GetPersona(ctx context.Context, id uuid.UUID) (*Persona, error) {
	persona := &Persona{}
	query := `
		SELECT persona_id, name, description, prompt_template, model_config, created_at, updated_at
		FROM personas
		WHERE persona_id = $1
	`
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&persona.ID,
		&persona.Name,
		&persona.Description,
		&persona.PromptTemplate,
		&persona.ModelConfig,
		&persona.CreatedAt,
		&persona.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Persona not found
		}
		return nil, fmt.Errorf("failed to get persona: %w", err)
	}
	return persona, nil
}
