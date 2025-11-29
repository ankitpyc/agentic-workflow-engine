package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"encoding/json"

	"github.com/google/uuid"
)

type StageRunStatus string

const (
	StageRunStatusPending   StageRunStatus = "pending"
	StageRunStatusRunning   StageRunStatus = "running"
	StageRunStatusCompleted StageRunStatus = "completed"
	StageRunStatusFailed    StageRunStatus = "failed"
	StageRunStatusApproved  StageRunStatus = "approved"
	StageRunStatusRejected  StageRunStatus = "rejected"
)

type StageRun struct {
	ID            uuid.UUID       `json:"id"`
	ProjectID     uuid.UUID       `json:"project_id"`
	StageName     string          `json:"stage_name"`
	Status        StageRunStatus  `json:"status"`
	InputContext  json.RawMessage `json:"input_context"`  // JSONB type
	OutputContext json.RawMessage `json:"output_context"` // JSONB type
	StartedAt     sql.NullTime    `json:"started_at"`
	CompletedAt   sql.NullTime    `json:"completed_at"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type StageRunStore struct {
	db *sql.DB
}

func NewStageRunStore(db *sql.DB) *StageRunStore {
	return &StageRunStore{db: db}
}

func (s *StageRunStore) CreateStageRun(ctx context.Context, stageRun *StageRun) error {
	stageRun.ID = uuid.New()
	stageRun.Status = StageRunStatusPending
	stageRun.CreatedAt = time.Now()
	stageRun.UpdatedAt = time.Now()

	query := `
		INSERT INTO stage_runs (stage_run_id, project_id, stage_name, status, input_context, output_context, started_at, completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := s.db.ExecContext(ctx, query,
		stageRun.ID,
		stageRun.ProjectID,
		stageRun.StageName,
		stageRun.Status,
		stageRun.InputContext,
		stageRun.OutputContext,
		stageRun.StartedAt,
		stageRun.CompletedAt,
		stageRun.CreatedAt,
		stageRun.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create stage run: %w", err)
	}
	return nil
}

func (s *StageRunStore) GetStageRun(ctx context.Context, id uuid.UUID) (*StageRun, error) {
	stageRun := &StageRun{}
	query := `
		SELECT stage_run_id, project_id, stage_name, status, input_context, output_context, started_at, completed_at, created_at, updated_at
		FROM stage_runs
		WHERE stage_run_id = $1
	`
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&stageRun.ID,
		&stageRun.ProjectID,
		&stageRun.StageName,
		&stageRun.Status,
		&stageRun.InputContext,
		&stageRun.OutputContext,
		&stageRun.StartedAt,
		&stageRun.CompletedAt,
		&stageRun.CreatedAt,
		&stageRun.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Stage run not found
		}
		return nil, fmt.Errorf("failed to get stage run: %w", err)
	}
	return stageRun, nil
}

func (s *StageRunStore) UpdateStageRunStatus(ctx context.Context, id uuid.UUID, status StageRunStatus, startedAt, completedAt sql.NullTime) error {
	query := `
		UPDATE stage_runs
		SET status = $1, started_at = $2, completed_at = $3, updated_at = $4
		WHERE stage_run_id = $5
	`
	_, err := s.db.ExecContext(ctx, query, status, startedAt, completedAt, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update stage run status: %w", err)
	}
	return nil
}
