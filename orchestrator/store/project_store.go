package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ProjectStatus string

const (
	ProjectStatusCreated   ProjectStatus = "created"
	ProjectStatusRunning   ProjectStatus = "running"
	ProjectStatusCompleted ProjectStatus = "completed"
	ProjectStatusFailed    ProjectStatus = "failed"
)

type Project struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	Status      ProjectStatus  `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type ProjectStore struct {
	db *sql.DB
}

func NewProjectStore(db *sql.DB) *ProjectStore {
	return &ProjectStore{db: db}
}

func (s *ProjectStore) CreateProject(ctx context.Context, project *Project) error {
	project.ID = uuid.New()
	project.Status = ProjectStatusCreated
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()

	query := `
		INSERT INTO projects (project_id, name, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := s.db.ExecContext(ctx, query,
		project.ID,
		project.Name,
		project.Description,
		project.Status,
		project.CreatedAt,
		project.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	return nil
}

func (s *ProjectStore) GetProject(ctx context.Context, id uuid.UUID) (*Project, error) {
	project := &Project{}
	query := `
		SELECT project_id, name, description, status, created_at, updated_at
		FROM projects
		WHERE project_id = $1
	`
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.Status,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Project not found
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return project, nil
}

func (s *ProjectStore) UpdateProjectStatus(ctx context.Context, id uuid.UUID, status ProjectStatus) error {
	query := `
		UPDATE projects
		SET status = $1, updated_at = $2
		WHERE project_id = $3
	`
	_, err := s.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update project status: %w", err)
	}
	return nil
}
