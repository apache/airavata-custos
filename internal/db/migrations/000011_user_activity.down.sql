DELETE FROM role_privileges rp USING roles r
WHERE rp.role_id = r.id AND r.name IN ('admin', 'super_admin')
  AND rp.privilege = 'core:users:activity:read';
DROP INDEX IF EXISTS idx_audit_portal_login_user_time;
