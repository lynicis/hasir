-- Drop triggers if they exist
DROP TRIGGER IF EXISTS repositories_search_refresh ON repositories;
DROP TRIGGER IF EXISTS organizations_search_refresh ON organizations;

-- Drop trigger functions
DROP FUNCTION IF EXISTS refresh_search_items_trigger();
DROP FUNCTION IF EXISTS refresh_search_items();

-- Drop materialized view
DROP MATERIALIZED VIEW IF EXISTS search_items;
