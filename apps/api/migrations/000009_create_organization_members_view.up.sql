-- Create view for organization members with user details
CREATE OR REPLACE VIEW organization_members_view AS
SELECT
    om.id,
    om.organization_id,
    om.user_id,
    om.role,
    om.joined_at,
    u.username,
    u.email
FROM organization_members om
INNER JOIN users u ON om.user_id = u.id
WHERE u.deleted_at IS NULL;
