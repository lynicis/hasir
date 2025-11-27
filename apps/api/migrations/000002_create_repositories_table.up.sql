CREATE TABLE IF NOT EXISTS repositories (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    owner_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    path VARCHAR(512) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_repositories_name ON repositories(name);
CREATE INDEX idx_repositories_owner_id ON repositories(owner_id);
CREATE INDEX idx_repositories_deleted_at ON repositories(deleted_at);

