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

-- Persistent monotonic UID allocator per cluster. The row for a given
-- cluster_id holds the NEXT uidNumber to hand out; allocation is a
-- single UPDATE that InnoDB serialises via row locking, so concurrent
-- allocators (in-process or cross-process) cannot pick the same value.
--
-- Rows are never deleted — decommissioning a cluster is handled at a
-- higher level. Never-decrementing counter is what guarantees no UID
-- reuse after an LDAP entry is deleted.
CREATE TABLE ldap_uid_sequence (
    cluster_id VARCHAR(64) NOT NULL PRIMARY KEY,
    next_uid   BIGINT      NOT NULL,
    updated_at TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
