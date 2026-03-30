-- Revert is complex; this is a best-effort rollback
DELETE FROM user_roles WHERE organization_id IS NULL;
