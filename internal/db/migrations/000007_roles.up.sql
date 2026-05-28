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

SET NAMES utf8mb4;
SET time_zone = '+00:00';

-- Roles are named bundles of privileges. Granting a role to a user is
-- equivalent to granting each of the role's privileges.
CREATE TABLE IF NOT EXISTS roles
(
    id          VARCHAR(255) NOT NULL,
    name        VARCHAR(64)  NOT NULL,
    description TEXT         NULL,
    is_system   TINYINT(1)   NOT NULL DEFAULT 0,
    created_at  TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_roles_name (name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- Privileges that a role bundles. Holding the role implies holding every
-- listed privilege.
CREATE TABLE IF NOT EXISTS role_privileges
(
    role_id   VARCHAR(255) NOT NULL,
    privilege VARCHAR(64)  NOT NULL,
    added_at  TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (role_id, privilege),
    CONSTRAINT fk_role_privileges_role FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- Active role assignments. Revoke is DELETE; full history lives in audit_events.
-- A user may hold any number of roles simultaneously.
CREATE TABLE IF NOT EXISTS user_roles
(
    user_id    VARCHAR(255) NOT NULL,
    role_id    VARCHAR(255) NOT NULL,
    granted_by VARCHAR(255) NULL,
    granted_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    reason     TEXT         NULL,
    PRIMARY KEY (user_id, role_id),
    KEY idx_user_roles_role (role_id),
    CONSTRAINT fk_user_roles_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_user_roles_role FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE,
    CONSTRAINT fk_user_roles_granted_by FOREIGN KEY (granted_by) REFERENCES users (id) ON DELETE SET NULL
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
