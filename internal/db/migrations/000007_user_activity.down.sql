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

DELETE FROM role_privileges WHERE privilege = 'core:users:activity:read';

DROP TABLE IF EXISTS user_login_daily;
DROP TABLE IF EXISTS user_login_events;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS chk_users_login_count_nonnegative,
    DROP CONSTRAINT IF EXISTS chk_users_login_day_count_nonnegative,
    DROP CONSTRAINT IF EXISTS chk_users_login_streak_nonnegative;

ALTER TABLE users
    DROP COLUMN IF EXISTS timezone,
    DROP COLUMN IF EXISTS last_login,
    DROP COLUMN IF EXISTS last_login_local_date,
    DROP COLUMN IF EXISTS login_count,
    DROP COLUMN IF EXISTS login_day_count,
    DROP COLUMN IF EXISTS login_streak;
