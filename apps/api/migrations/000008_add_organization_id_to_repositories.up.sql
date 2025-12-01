-- Add organization_id column if it doesn't already exist
ALTER TABLE repositories
ADD COLUMN IF NOT EXISTS organization_id VARCHAR(36) REFERENCES organizations(id) ON DELETE CASCADE;

-- Create index on organization_id
CREATE INDEX IF NOT EXISTS idx_repositories_organization_id ON repositories(organization_id);

-- Remove the unique constraint on name (repository names are not unique)
ALTER TABLE repositories DROP CONSTRAINT IF EXISTS repositories_name_key;
