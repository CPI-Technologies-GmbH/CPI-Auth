-- Fix user_roles: allow global (non-org) role assignments
ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS user_roles_pkey;
ALTER TABLE user_roles ALTER COLUMN organization_id DROP NOT NULL;
ALTER TABLE user_roles ALTER COLUMN organization_id SET DEFAULT NULL;
ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS user_roles_organization_id_fkey;
ALTER TABLE user_roles ADD CONSTRAINT user_roles_pkey PRIMARY KEY (user_id, role_id);

-- Assign admin role to the first user (seed admin)
INSERT INTO user_roles (user_id, role_id, organization_id)
SELECT u.id, 'b0000000-0000-0000-0000-000000000001', NULL
FROM users u
ORDER BY u.created_at ASC LIMIT 1
ON CONFLICT DO NOTHING;
