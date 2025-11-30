-- Restore original unique constraint on name
ALTER TABLE repositories ADD CONSTRAINT repositories_name_key UNIQUE (name);

-- Drop index and column
DROP INDEX IF EXISTS idx_repositories_organization_id;
ALTER TABLE repositories DROP COLUMN IF EXISTS organization_id;

