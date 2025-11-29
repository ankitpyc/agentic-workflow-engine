package store

import (
	"context"
	"database/sql"
	"encoding/json" // Import for json.RawMessage
	"log"
	"os"
	"testing"
	"time"

	"workflow-engine/config"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	err := godotenv.Load("../../.env.test") // Load test environment variables
	if err != nil {
		log.Printf("No .env.test file found, using defaults or system env: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config for tests: %v", err)
	}

	// Use a dedicated test database URL
	// For production readiness, this should ideally use testcontainers-go or similar
	// to spin up an isolated DB for each test run.
	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")
	if testDatabaseURL == "" {
		testDatabaseURL = cfg.DatabaseURL // Fallback to main DB URL, NOT recommended for real tests
		log.Println("WARNING: TEST_DATABASE_URL not set, using main database URL for tests. This is NOT recommended for isolated testing.")
	}

	testDB, err = sql.Open("postgres", testDatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open test database connection: %v", err)
	}
	defer testDB.Close()

	// Set connection pool settings for tests
	testDB.SetMaxOpenConns(5)
	testDB.SetMaxIdleConns(2)
	testDB.SetConnMaxLifetime(1 * time.Minute)

	if err = testDB.Ping(); err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	log.Println("Successfully connected to test PostgreSQL database!")

	// Run tests
	code := m.Run()

	// Teardown (optional, typically done by testcontainers itself)
	// You might want to clear tables here if not using testcontainers for full isolation
	clearTables(testDB)

	os.Exit(code)
}

func clearTables(db *sql.DB) {
	_, err := db.Exec("TRUNCATE TABLE projects, personas, stage_runs RESTART IDENTITY CASCADE;")
	if err != nil {
		log.Printf("Failed to truncate tables: %v", err)
	}
	log.Println("Tables truncated successfully.")
}

func TestProjectStore_CreateAndGetProject(t *testing.T) {
	// Ensure tables are clean before each test
	clearTables(testDB)

	store := NewProjectStore(testDB)
	ctx := context.Background()

	projectName := "Test Project"
	projectDesc := "This is a test project description."

	project := &Project{
		Name:        projectName,
		Description: sql.NullString{String: projectDesc, Valid: true},
	}

	err := store.CreateProject(ctx, project)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	if project.ID == uuid.Nil {
		t.Fatal("Project ID was not generated")
	}

	retrievedProject, err := store.GetProject(ctx, project.ID)
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}
	if retrievedProject == nil {
		t.Fatal("Retrieved project is nil")
	}

	if retrievedProject.Name != projectName {
		t.Errorf("Expected project name %s, got %s", projectName, retrievedProject.Name)
	}
	if retrievedProject.Description.String != projectDesc {
		t.Errorf("Expected project description %s, got %s", projectDesc, retrievedProject.Description.String)
	}
	if retrievedProject.Status != ProjectStatusCreated {
		t.Errorf("Expected project status %s, got %s", ProjectStatusCreated, retrievedProject.Status)
	}
}

func TestProjectStore_UpdateProjectStatus(t *testing.T) {
	clearTables(testDB)

	store := NewProjectStore(testDB)
	ctx := context.Background()

	project := &Project{
		Name: "Updatable Project",
	}
	err := store.CreateProject(ctx, project)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	newStatus := ProjectStatusRunning
	err = store.UpdateProjectStatus(ctx, project.ID, newStatus)
	if err != nil {
		t.Fatalf("UpdateProjectStatus failed: %v", err)
	}

	retrievedProject, err := store.GetProject(ctx, project.ID)
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}
	if retrievedProject.Status != newStatus {
		t.Errorf("Expected status %s, got %s", newStatus, retrievedProject.Status)
	}
	if retrievedProject.UpdatedAt.Equal(project.UpdatedAt) {
		t.Errorf("UpdatedAt was not updated")
	}
}

func TestPersonaStore_CreateAndGetPersona(t *testing.T) {
	clearTables(testDB)

	store := NewPersonaStore(testDB)
	ctx := context.Background()

	personaName := "Test Persona"
	promptTemplate := "You are a {{.Role}} agent."
	modelConfig := json.RawMessage(`{"model_name": "gpt-4", "temperature": 0.7}`)

	persona := &Persona{
		Name:           personaName,
		PromptTemplate: promptTemplate,
		ModelConfig:    modelConfig,
	}

	err := store.CreatePersona(ctx, persona)
	if err != nil {
		t.Fatalf("CreatePersona failed: %v", err)
	}
	if persona.ID == uuid.Nil {
		t.Fatal("Persona ID was not generated")
	}

	retrievedPersona, err := store.GetPersona(ctx, persona.ID)
	if err != nil {
		t.Fatalf("GetPersona failed: %v", err)
	}
	if retrievedPersona == nil {
		t.Fatal("Retrieved persona is nil")
	}

	if retrievedPersona.Name != personaName {
		t.Errorf("Expected persona name %s, got %s", personaName, retrievedPersona.Name)
	}
	if retrievedPersona.PromptTemplate != promptTemplate {
		t.Errorf("Expected prompt template %s, got %s", promptTemplate, retrievedPersona.PromptTemplate)
	}
	if string(retrievedPersona.ModelConfig) != string(modelConfig) {
		t.Errorf("Expected model config %s, got %s", string(modelConfig), string(retrievedPersona.ModelConfig))
	}
}

func TestStageRunStore_CreateAndGetStageRun(t *testing.T) {
	clearTables(testDB)

	// First, create a project to link the stage run to
	projectStore := NewProjectStore(testDB)
	ctx := context.Background()
	project := &Project{Name: "Project for StageRun"}
	err := projectStore.CreateProject(ctx, project)
	if err != nil {
		t.Fatalf("Failed to create project for stage run test: %v", err)
	}

	store := NewStageRunStore(testDB)
	stageName := "initial-setup"
	inputContext := json.RawMessage(`{"param1": "value1"}`)

	stageRun := &StageRun{
		ProjectID:    project.ID,
		StageName:    stageName,
		InputContext: inputContext,
	}

	err = store.CreateStageRun(ctx, stageRun)
	if err != nil {
		t.Fatalf("CreateStageRun failed: %v", err)
	}
	if stageRun.ID == uuid.Nil {
		t.Fatal("StageRun ID was not generated")
	}
	if stageRun.Status != StageRunStatusPending {
		t.Errorf("Expected initial status %s, got %s", StageRunStatusPending, stageRun.Status)
	}

	retrievedStageRun, err := store.GetStageRun(ctx, stageRun.ID)
	if err != nil {
		t.Fatalf("GetStageRun failed: %v", err)
	}
	if retrievedStageRun == nil {
		t.Fatal("Retrieved stage run is nil")
	}

	if retrievedStageRun.ProjectID != project.ID {
		t.Errorf("Expected ProjectID %s, got %s", project.ID, retrievedStageRun.ProjectID)
	}
	if retrievedStageRun.StageName != stageName {
		t.Errorf("Expected StageName %s, got %s", stageName, retrievedStageRun.StageName)
	}
	if string(retrievedStageRun.InputContext) != string(inputContext) {
		t.Errorf("Expected InputContext %s, got %s", string(inputContext), string(retrievedStageRun.InputContext))
	}
}

func TestStageRunStore_UpdateStageRunStatus(t *testing.T) {
	clearTables(testDB)

	projectStore := NewProjectStore(testDB)
	ctx := context.Background()
	project := &Project{Name: "Project for StageRun Update"}
	err := projectStore.CreateProject(ctx, project)
	if err != nil {
		t.Fatalf("Failed to create project for stage run update test: %v", err)
	}

	store := NewStageRunStore(testDB)
	stageName := "update-stage"
	stageRun := &StageRun{
		ProjectID: project.ID,
		StageName: stageName,
	}
	err = store.CreateStageRun(ctx, stageRun)
	if err != nil {
		t.Fatalf("CreateStageRun failed: %v", err)
	}

	newStatus := StageRunStatusRunning
	startedAt := sql.NullTime{Time: time.Now(), Valid: true}
	err = store.UpdateStageRunStatus(ctx, stageRun.ID, newStatus, startedAt, sql.NullTime{})
	if err != nil {
		t.Fatalf("UpdateStageRunStatus failed: %v", err)
	}

	retrievedStageRun, err := store.GetStageRun(ctx, stageRun.ID)
	if err != nil {
		t.Fatalf("GetStageRun failed: %v", err)
	}
	if retrievedStageRun.Status != newStatus {
		t.Errorf("Expected status %s, got %s", newStatus, retrievedStageRun.Status)
	}
	if !retrievedStageRun.StartedAt.Valid || !retrievedStageRun.StartedAt.Time.Equal(startedAt.Time) {
		t.Errorf("StartedAt was not updated correctly")
	}
	if retrievedStageRun.CompletedAt.Valid {
		t.Errorf("CompletedAt should not be valid yet")
	}

	completedStatus := StageRunStatusCompleted
	completedAt := sql.NullTime{Time: time.Now(), Valid: true}
	err = store.UpdateStageRunStatus(ctx, stageRun.ID, completedStatus, retrievedStageRun.StartedAt, completedAt)
	if err != nil {
		t.Fatalf("UpdateStageRunStatus failed to set completed time: %v", err)
	}

	retrievedStageRun, err = store.GetStageRun(ctx, stageRun.ID)
	if err != nil {
		t.Fatalf("GetStageRun after completed status failed: %v", err)
	}
	if retrievedStageRun.Status != completedStatus {
		t.Errorf("Expected final status %s, got %s", completedStatus, retrievedStageRun.Status)
	}
	if !retrievedStageRun.CompletedAt.Valid || !retrievedStageRun.CompletedAt.Time.Equal(completedAt.Time) {
		t.Errorf("CompletedAt was not updated correctly after setting completed status")
	}
}
