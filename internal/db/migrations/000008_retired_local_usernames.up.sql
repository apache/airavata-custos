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

-- Usernames of deleted cluster accounts, held for their user so no one else takes them, until an admin releases one.
CREATE TABLE IF NOT EXISTS retired_local_usernames
(
    compute_cluster_id VARCHAR(255)   NOT NULL,
    local_username     VARCHAR(32)    NOT NULL,
    user_id            VARCHAR(255)   NOT NULL,
    retired_at         TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (compute_cluster_id, local_username),
    CONSTRAINT fk_retired_local_usernames_cluster FOREIGN KEY (compute_cluster_id) REFERENCES compute_clusters (id) ON DELETE CASCADE
);
