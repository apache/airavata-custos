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

CREATE TABLE IF NOT EXISTS user_merges
(
    id                BIGINT       NOT NULL AUTO_INCREMENT,
    retiring_user_id  VARCHAR(255) NOT NULL,
    surviving_user_id VARCHAR(255) NOT NULL,
    reason            TEXT         NULL,
    merged_at         TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_user_merges_retiring (retiring_user_id),
    KEY idx_user_merges_surviving (surviving_user_id),
    CONSTRAINT fk_user_merges_retiring  FOREIGN KEY (retiring_user_id)  REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT fk_user_merges_surviving FOREIGN KEY (surviving_user_id) REFERENCES users (id) ON DELETE RESTRICT
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
