CREATE TYPE stage_run_status AS ENUM ('pending', 'running', 'completed', 'failed', 'approved', 'rejected');

CREATE TABLE stage_runs (
    stage_run_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    stage_name TEXT NOT NULL,
    status stage_run_status NOT NULL DEFAULT 'pending',
    input_context JSONB,
    output_context JSONB,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_stage_runs_project_id ON stage_runs (project_id);
CREATE INDEX idx_stage_runs_status ON stage_runs (status);
