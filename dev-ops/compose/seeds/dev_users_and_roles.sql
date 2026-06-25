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

-- Dev-only seed: inserts an org, four developer-facing users, and two
-- non-system roles with sample privilege bundles. Apply after migrations
-- have run; safe to re-apply (everything is idempotent).
--
-- Apply with:
--   docker exec -i custos_db mariadb -uadmin -padmin custos \
--     < dev-ops/compose/seeds/dev_users_and_roles.sql
--
-- The bootstrap super_admin role is NOT created here — set
--   CUSTOS_BOOTSTRAP_ADMIN_EMAIL=admin@custos.local
-- before starting the server and it will be created idempotently on boot
-- and granted to the user inserted below.

SET NAMES utf8mb4;
SET time_zone = '+00:00';

-- ---------------------------------------------------------------------------
-- Organization
-- ---------------------------------------------------------------------------
INSERT IGNORE INTO organizations (id, originated_id, name)
VALUES ('dev-org', 'DEV-ORG', 'Custos Dev Org');

-- ---------------------------------------------------------------------------
-- Users (deterministic IDs so the portal dev-credentials provider can map each
-- dev level to a known backend user; OIDC mode resolves sub via user_identities)
-- ---------------------------------------------------------------------------
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status)
VALUES
    ('dev-admin',      'dev-org', 'Admin',      'Dev', '', 'admin@custos.local',      'ACTIVE'),
    ('dev-operator',   'dev-org', 'Operator',   'Dev', '', 'operator@custos.local',   'ACTIVE'),
    ('dev-auditor',    'dev-org', 'Auditor',    'Dev', '', 'auditor@custos.local',    'ACTIVE'),
    ('dev-researcher', 'dev-org', 'Researcher', 'Dev', '', 'researcher@custos.local', 'ACTIVE');

-- ---------------------------------------------------------------------------
-- Roles (deterministic UUIDs so API examples stay stable across re-applies)
-- ---------------------------------------------------------------------------
-- operator: AMIE + HPC read/write. Day-to-day operations of HPC + AMIE flows.
INSERT IGNORE INTO roles (id, name, description, is_system)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'operator', 'Day-to-day AMIE and HPC operations (read + write)', 0),
    ('22222222-2222-2222-2222-222222222222', 'auditor', 'Read-only access across AMIE and HPC surfaces', 0);

-- operator privileges
INSERT IGNORE INTO role_privileges (role_id, privilege) VALUES
    ('11111111-1111-1111-1111-111111111111', 'amie:packets:read'),
    ('11111111-1111-1111-1111-111111111111', 'amie:packets:write'),
    ('11111111-1111-1111-1111-111111111111', 'amie:replies:read'),
    ('11111111-1111-1111-1111-111111111111', 'amie:replies:write'),
    ('11111111-1111-1111-1111-111111111111', 'amie:unmapped:read'),
    ('11111111-1111-1111-1111-111111111111', 'amie:unmapped:write'),
    ('11111111-1111-1111-1111-111111111111', 'core:clusters:read'),
    ('11111111-1111-1111-1111-111111111111', 'core:clusters:write');

-- auditor privileges
INSERT IGNORE INTO role_privileges (role_id, privilege) VALUES
    ('22222222-2222-2222-2222-222222222222', 'amie:packets:read'),
    ('22222222-2222-2222-2222-222222222222', 'amie:replies:read'),
    ('22222222-2222-2222-2222-222222222222', 'amie:unmapped:read'),
    ('22222222-2222-2222-2222-222222222222', 'core:clusters:read');

-- ---------------------------------------------------------------------------
-- Role assignments
-- ---------------------------------------------------------------------------
INSERT IGNORE INTO user_roles (user_id, role_id, granted_by, reason)
VALUES
    ('dev-operator', '11111111-1111-1111-1111-111111111111', 'dev-admin', 'dev seed'),
    ('dev-auditor',  '22222222-2222-2222-2222-222222222222', 'dev-admin', 'dev seed');

-- dev-researcher holds no roles - used to exercise 403 paths from a
-- low-privilege caller.
