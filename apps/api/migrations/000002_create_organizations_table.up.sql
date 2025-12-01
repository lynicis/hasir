DO $$ BEGIN
    CREATE TYPE visibility AS ENUM ('private', 'public');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS organizations (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    visibility visibility NOT NULL DEFAULT 'private',
    created_by VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_organizations_name ON organizations(name);
CREATE INDEX idx_organizations_created_by ON organizations(created_by);
CREATE INDEX idx_organizations_visibility ON organizations(visibility);
CREATE INDEX idx_organizations_deleted_at ON organizations(deleted_at);
