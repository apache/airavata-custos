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

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS timezone VARCHAR(64) NULL DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS last_login TIMESTAMPTZ(6) NULL DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS last_login_local_date DATE NULL DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS login_count BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS login_day_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS login_streak INTEGER NOT NULL DEFAULT 0;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS chk_users_login_count_nonnegative,
    DROP CONSTRAINT IF EXISTS chk_users_login_day_count_nonnegative,
    DROP CONSTRAINT IF EXISTS chk_users_login_streak_nonnegative;
ALTER TABLE users
    ADD CONSTRAINT chk_users_login_count_nonnegative CHECK (login_count >= 0),
    ADD CONSTRAINT chk_users_login_day_count_nonnegative CHECK (login_day_count >= 0),
    ADD CONSTRAINT chk_users_login_streak_nonnegative CHECK (login_streak >= 0);

CREATE TABLE IF NOT EXISTS user_login_events
(
    id          VARCHAR(255) NOT NULL,
    event_key   VARCHAR(255) NOT NULL,
    user_id     VARCHAR(255) NOT NULL,
    occurred_at TIMESTAMPTZ(6) NOT NULL,
    local_date  DATE         NOT NULL,
    timezone    VARCHAR(64)  NOT NULL,
    provider    VARCHAR(64)  NULL,
    session_id  VARCHAR(255) NULL,
    created_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    CONSTRAINT uq_user_login_event_key UNIQUE (event_key),
    CONSTRAINT fk_user_login_event_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_user_login_time ON user_login_events (user_id, occurred_at);
CREATE INDEX IF NOT EXISTS idx_user_login_date ON user_login_events (user_id, local_date);

CREATE TABLE IF NOT EXISTS user_login_daily
(
    user_id        VARCHAR(255) NOT NULL,
    local_date     DATE         NOT NULL,
    timezone       VARCHAR(64)  NOT NULL,
    login_count    INTEGER      NOT NULL DEFAULT 0,
    first_login_at TIMESTAMPTZ(6) NOT NULL,
    last_login_at  TIMESTAMPTZ(6) NOT NULL,
    PRIMARY KEY (user_id, local_date),
    CONSTRAINT chk_user_login_daily_count_nonnegative CHECK (login_count >= 0),
    CONSTRAINT fk_user_login_daily_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_login_daily_date ON user_login_daily (local_date);

-- Existing system roles were created from the privilege registry at the time.
-- Backfill the new read privilege for upgrades; newly created roles receive it
-- through the normal KnownPrivileges path.
INSERT INTO role_privileges (role_id, privilege)
SELECT id, 'core:users:activity:read'
FROM roles
WHERE is_system = TRUE AND name IN ('admin', 'super_admin')
ON CONFLICT DO NOTHING;
