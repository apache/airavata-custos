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

-- Dev-only seed: inserts an org, five developer-facing users, and three
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
    -- site staff
    ('dev-admin',      'dev-org', 'Admin',      'Dev', '', 'admin@custos.local',      'ACTIVE'),
    ('dev-operator',   'dev-org', 'Operator',   'Dev', '', 'operator@custos.local',   'ACTIVE'),
    ('dev-auditor',    'dev-org', 'Auditor',    'Dev', '', 'auditor@custos.local',    'ACTIVE'),
    -- researchers
    ('dev-pi',         'dev-org', 'PI',         'Dev', '', 'pi@custos.local',         'ACTIVE'),
    ('dev-researcher', 'dev-org', 'Researcher', 'Dev', '', 'researcher@custos.local', 'ACTIVE');

-- ---------------------------------------------------------------------------
-- Roles (deterministic UUIDs so API examples stay stable across re-applies)
-- ---------------------------------------------------------------------------
-- operator: AMIE + cluster read/write. Day-to-day operations of cluster + AMIE flows.
-- pi:       Manages own projects/allocations + reads AMIE status.
-- auditor:  Read-only across AMIE and cluster surfaces.
INSERT IGNORE INTO roles (id, name, description, is_system)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'operator', 'Day-to-day AMIE and cluster operations (read + write)', 0),
    ('22222222-2222-2222-2222-222222222222', 'auditor',  'Read-only access across AMIE and cluster surfaces', 0),
    ('33333333-3333-3333-3333-333333333333', 'pi',       'Principal Investigator - manages own projects and allocations', 0);

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

-- pi privileges
INSERT IGNORE INTO role_privileges (role_id, privilege) VALUES
    ('33333333-3333-3333-3333-333333333333', 'amie:packets:read'),
    ('33333333-3333-3333-3333-333333333333', 'amie:replies:read'),
    ('33333333-3333-3333-3333-333333333333', 'core:clusters:read'),
    ('33333333-3333-3333-3333-333333333333', 'core:clusters:write');

-- ---------------------------------------------------------------------------
-- Role assignments
-- ---------------------------------------------------------------------------
INSERT IGNORE INTO user_roles (user_id, role_id, granted_by, reason)
VALUES
    -- site staff
    ('dev-operator', '11111111-1111-1111-1111-111111111111', 'dev-admin', 'dev seed'),
    ('dev-auditor',  '22222222-2222-2222-2222-222222222222', 'dev-admin', 'dev seed'),
    -- researchers
    ('dev-pi',       '33333333-3333-3333-3333-333333333333', 'dev-admin', 'dev seed');

-- dev-researcher holds no roles - used to exercise 403 paths from a
-- low-privilege caller.

-- ---------------------------------------------------------------------------
-- OIDC identity link rows - oidc_sub UUIDs match the user `id` field in
-- dev-ops/compose/keycloak/import/custos-realm.json so a Keycloak sign-in
-- resolves to the right backend user via identity_resolver.
-- ---------------------------------------------------------------------------
INSERT IGNORE INTO user_identities (id, user_id, source, external_id, email, oidc_sub)
VALUES
    -- site staff
    ('ui-dev-admin',      'dev-admin',      'keycloak', 'a8f3c52d-7b1e-4d9a-8c3f-2e5b9d8a1c47', 'admin@custos.local',      'a8f3c52d-7b1e-4d9a-8c3f-2e5b9d8a1c47'),
    ('ui-dev-operator',   'dev-operator',   'keycloak', 'b6e2f841-9c5d-4a3b-8f7e-1d6c3b9e8f25', 'operator@custos.local',   'b6e2f841-9c5d-4a3b-8f7e-1d6c3b9e8f25'),
    ('ui-dev-auditor',    'dev-auditor',    'keycloak', 'c4d9e3f7-8b2a-4f5d-9e1c-3a8b7d6e2f14', 'auditor@custos.local',    'c4d9e3f7-8b2a-4f5d-9e1c-3a8b7d6e2f14'),
    -- researchers
    ('ui-dev-pi',         'dev-pi',         'keycloak', 'e7b4c1a2-6d8f-4e9b-8a5c-7d3f9e2b1c08', 'pi@custos.local',         'e7b4c1a2-6d8f-4e9b-8a5c-7d3f9e2b1c08'),
    ('ui-dev-researcher', 'dev-researcher', 'keycloak', 'd5a8b2e6-4f9c-4d1a-9b3e-9f2c6e1b8d54', 'researcher@custos.local', 'd5a8b2e6-4f9c-4d1a-9b3e-9f2c6e1b8d54');
