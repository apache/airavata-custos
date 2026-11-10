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

-- A row means the user has allowed the app. Removing it is recorded in audit_events.
CREATE TABLE IF NOT EXISTS app_grants
(
    app_id     VARCHAR(255)   NOT NULL,
    user_id    VARCHAR(255)   NOT NULL,
    granted_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (app_id, user_id),
    CONSTRAINT fk_app_grants_app FOREIGN KEY (app_id) REFERENCES apps (id) ON DELETE RESTRICT,
    CONSTRAINT fk_app_grants_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT
);
