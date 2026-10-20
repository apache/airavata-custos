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

-- A deletion of an account or allocation waits here for an admin's approval, which runs the delete.
CREATE TABLE IF NOT EXISTS deletion_requests
(
    entity_type  VARCHAR(32)    NOT NULL CHECK (entity_type IN ('compute_cluster_user', 'compute_allocation')),
    entity_id    VARCHAR(255)   NOT NULL,
    requested_by VARCHAR(255)   NOT NULL,
    requested_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (entity_type, entity_id)
);
