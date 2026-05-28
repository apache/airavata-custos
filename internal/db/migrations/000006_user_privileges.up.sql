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

-- Privileges held by a user. Only active grants live here; revoke is DELETE.
-- The full grant/revoke history (who, when, why) is recorded in audit_events.
CREATE TABLE IF NOT EXISTS user_privileges
(
    id          VARCHAR(255) NOT NULL,
    user_id     VARCHAR(255) NOT NULL,
    privilege   VARCHAR(64)  NOT NULL,
    granted_by  VARCHAR(255) NULL,
    granted_at  TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    reason      TEXT         NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_user_privileges (user_id, privilege),
    KEY idx_user_privileges_user (user_id),
    KEY idx_user_privileges_priv (privilege),
    CONSTRAINT fk_user_privileges_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_user_privileges_granted_by FOREIGN KEY (granted_by) REFERENCES users (id) ON DELETE SET NULL
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
