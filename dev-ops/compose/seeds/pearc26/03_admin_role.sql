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

-- PEARC26 seed 03: the workshop approver / bootstrap admin.
--
-- This user doubles as the bootstrap super_admin and the tutorial
-- project PI (seed 04): set CUSTOS_BOOTSTRAP_ADMIN_EMAIL to this email
-- (or edit the email here to your own) and boot grants super_admin
-- idempotently. The access-approver role below covers deployments that
-- keep a separate approver.

SET NAMES utf8mb4;
SET time_zone = '+00:00';

INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status)
VALUES ('pearc26-approver', 'pearc26-org', 'Workshop', 'Approver', '',
        'approver@pearc26.local', 'ACTIVE');

INSERT IGNORE INTO roles (id, name, description, is_system)
VALUES ('44444444-4444-4444-4444-444444444444', 'access-approver',
        'Reviews trial-allocation access requests', 0);

INSERT IGNORE INTO role_privileges (role_id, privilege) VALUES
    ('44444444-4444-4444-4444-444444444444', 'core:access-requests:read'),
    ('44444444-4444-4444-4444-444444444444', 'core:access-requests:write'),
    ('44444444-4444-4444-4444-444444444444', 'core:users:read'),
    ('44444444-4444-4444-4444-444444444444', 'core:allocations:read');

INSERT IGNORE INTO user_roles (user_id, role_id, granted_by, reason)
VALUES ('pearc26-approver', '44444444-4444-4444-4444-444444444444',
        'pearc26-approver', 'pearc26 seed');
