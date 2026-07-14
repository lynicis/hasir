CREATE TABLE IF NOT EXISTS sdk_generation_jobs (
    id VARCHAR(36) PRIMARY KEY,
    repository_id VARCHAR(36) NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    commit_hash VARCHAR(40) NOT NULL,
    sdk sdk_type NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 5,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    CONSTRAINT sdk_generation_jobs_status_check CHECK (status IN ('pending', 'processing', 'completed', 'failed'))
);

CREATE INDEX idx_sdk_generation_jobs_repository_id ON sdk_generation_jobs(repository_id);
CREATE INDEX idx_sdk_generation_jobs_status ON sdk_generation_jobs(status);
CREATE INDEX idx_sdk_generation_jobs_commit_hash ON sdk_generation_jobs(commit_hash);
CREATE INDEX idx_sdk_generation_jobs_created_at ON sdk_generation_jobs(created_at);
