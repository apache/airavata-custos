-- Licensed to the Apache Software Foundation (ASF) under one
-- or more contributor license agreements.  See the NOTICE file
-- distributed with this work for additional information
-- regarding copyright ownership.  The ASF licenses this file
-- to you under the Apache License, Version 2.0 (the
-- "License"); you may not use this file except in compliance
-- with the License.  You may obtain a copy of the License at
--
--   http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing,
-- software distributed under the License is distributed on an
-- "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
-- KIND, either express or implied.  See the License for the
-- specific language governing permissions and limitations
-- under the License.

-- A deactivated cluster account turns INACTIVE, still holding its username, until an admin deletes it.
ALTER TABLE compute_cluster_users
    ADD COLUMN IF NOT EXISTS status VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    ADD CONSTRAINT chk_compute_cluster_users_status CHECK (status IN ('ACTIVE', 'INACTIVE'));

-- Deactivating and deleting were gated by the write privilege, so whoever holds it keeps both.
WITH grants (privilege, granted_from) AS (VALUES
    ('core:clusters:deactivate', 'core:clusters:write'),
    ('core:clusters:delete', 'core:clusters:write'),
    ('core:allocations:deactivate', 'core:allocations:write'),
    ('core:allocations:delete', 'core:allocations:write'),
    ('core:users:deactivate', 'core:users:write'),
    ('core:users:delete', 'core:users:write')),
roles AS (
    INSERT INTO role_privileges (role_id, privilege)
    SELECT rp.role_id, g.privilege FROM role_privileges rp JOIN grants g ON g.granted_from = rp.privilege
    ON CONFLICT DO NOTHING)
INSERT INTO user_privileges (id, user_id, privilege, granted_by, reason)
SELECT gen_random_uuid()::text, up.user_id, g.privilege, up.granted_by, up.reason
  FROM user_privileges up JOIN grants g ON g.granted_from = up.privilege
ON CONFLICT DO NOTHING;
