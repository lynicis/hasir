CREATE TABLE IF NOT EXISTS email_jobs (
    id VARCHAR(36) PRIMARY KEY,
    invite_id VARCHAR(36) NOT NULL REFERENCES organization_invites(id) ON DELETE CASCADE,
    organization_id VARCHAR(36) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    organization_name VARCHAR(255) NOT NULL,
    invite_token VARCHAR(64) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    CONSTRAINT chk_email_job_status CHECK (status IN ('pending', 'processing', 'completed', 'failed'))
);

CREATE INDEX idx_email_jobs_status ON email_jobs(status);
CREATE INDEX idx_email_jobs_created_at ON email_jobs(created_at);
CREATE INDEX idx_email_jobs_invite_id ON email_jobs(invite_id);
CREATE INDEX idx_email_jobs_organization_id ON email_jobs(organization_id);

-- Index for efficient job selection (pending jobs ordered by creation time)
CREATE INDEX idx_email_jobs_pending ON email_jobs(status, created_at) WHERE status = 'pending';
