-- Create a materialized view that unions organizations and repositories for unified search
CREATE MATERIALIZED VIEW IF NOT EXISTS search_items AS
SELECT
    o.id,
    o.name,
    'organization' AS item_type,
    NULL::VARCHAR(36) AS organization_id,
    o.created_at,
    o.deleted_at
FROM organizations o
UNION ALL
SELECT
    r.id,
    r.name,
    'repository' AS item_type,
    r.organization_id,
    r.created_at,
    r.deleted_at
FROM repositories r;

-- Create GIN index for trigram search on the name column
CREATE INDEX IF NOT EXISTS idx_search_items_name_trgm ON search_items USING gin (name gin_trgm_ops);

-- Create regular indexes for filtering
CREATE INDEX IF NOT EXISTS idx_search_items_item_type ON search_items (item_type);
CREATE INDEX IF NOT EXISTS idx_search_items_organization_id ON search_items (organization_id);
CREATE INDEX IF NOT EXISTS idx_search_items_deleted_at ON search_items (deleted_at);
CREATE INDEX IF NOT EXISTS idx_search_items_created_at ON search_items (created_at DESC);

-- Create a unique index to support concurrent refresh
CREATE UNIQUE INDEX IF NOT EXISTS idx_search_items_id_type ON search_items (id, item_type);

-- Create a function to refresh the materialized view
CREATE OR REPLACE FUNCTION refresh_search_items()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY search_items;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger function to refresh the view when organizations or repositories change
CREATE OR REPLACE FUNCTION refresh_search_items_trigger()
RETURNS trigger AS $$
BEGIN
    PERFORM refresh_search_items();
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER organizations_search_refresh
AFTER INSERT OR UPDATE OR DELETE ON organizations
FOR EACH STATEMENT EXECUTE FUNCTION refresh_search_items_trigger();

CREATE TRIGGER repositories_search_refresh
AFTER INSERT OR UPDATE OR DELETE ON repositories
FOR EACH STATEMENT EXECUTE FUNCTION refresh_search_items_trigger();
