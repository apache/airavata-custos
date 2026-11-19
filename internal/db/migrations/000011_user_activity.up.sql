-- Keep portal-login reads selective while reusing audit_events as the only fact table.
CREATE INDEX IF NOT EXISTS idx_audit_portal_login_user_time
    ON audit_events (entity_id, event_time DESC)
    WHERE event_type = 'PORTAL_LOGIN' AND entity_type = 'user';

-- Existing system roles were seeded before this privilege existed.
INSERT INTO role_privileges (role_id, privilege)
SELECT id, 'core:users:activity:read' FROM roles WHERE name IN ('admin', 'super_admin')
ON CONFLICT (role_id, privilege) DO NOTHING;
