-- CPI Auth IAM Platform - Seed Data Rollback
-- 002_seed_data.down.sql
-- Removes all seed data in reverse dependency order using fixed UUIDs

-- Remove audit log entry for seed
DELETE FROM audit_logs
    WHERE action = 'system.seed_data_applied'
    AND tenant_id = 'a0000000-0000-0000-0000-000000000001';

-- Remove JWKS key placeholder
DELETE FROM jwks_keys
    WHERE id = 'aa000000-0000-0000-0000-000000000001';

-- Remove branding config
DELETE FROM branding_configs
    WHERE id = 'f0000000-0000-0000-0000-000000000001';

-- Remove email templates
DELETE FROM email_templates
    WHERE id IN (
        'e0000000-0000-0000-0000-000000000001',
        'e0000000-0000-0000-0000-000000000002',
        'e0000000-0000-0000-0000-000000000003',
        'e0000000-0000-0000-0000-000000000004',
        'e0000000-0000-0000-0000-000000000005',
        'e0000000-0000-0000-0000-000000000006'
    );

-- Remove default application
DELETE FROM applications
    WHERE id = 'd0000000-0000-0000-0000-000000000001';

-- Remove user-role assignments for admin
DELETE FROM user_roles
    WHERE user_id = 'c0000000-0000-0000-0000-000000000001'
    AND role_id = 'b0000000-0000-0000-0000-000000000001';

-- Remove admin user
DELETE FROM users
    WHERE id = 'c0000000-0000-0000-0000-000000000001';

-- Remove roles (children first due to parent_role_id references)
DELETE FROM roles
    WHERE id = 'b0000000-0000-0000-0000-000000000004'; -- viewer
DELETE FROM roles
    WHERE id = 'b0000000-0000-0000-0000-000000000003'; -- editor
DELETE FROM roles
    WHERE id = 'b0000000-0000-0000-0000-000000000002'; -- manager
DELETE FROM roles
    WHERE id = 'b0000000-0000-0000-0000-000000000001'; -- admin

-- Remove default tenant
DELETE FROM tenants
    WHERE id = 'a0000000-0000-0000-0000-000000000001';
