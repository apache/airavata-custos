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

-- PEARC26 seed 01: organization + compute cluster.
--
-- Every attendee approved through the PEARC26 event lands in this
-- organization; their typed institution stays on the access request as
-- display text only.

SET NAMES utf8mb4;
SET time_zone = '+00:00';

INSERT IGNORE INTO organizations (id, originated_id, name)
VALUES ('pearc26-org', 'PEARC26', 'PEARC26 Attendees');

-- The id must equal CUSTOS_CLUSTER_ID: the COmanage subscriber only
-- provisions cluster users created on this cluster. Upsert corrects the
-- placeholder name from default_cluster.sql to the real cluster name.
INSERT INTO compute_clusters (id, name)
VALUES ('00000000-0000-0000-0000-000000000001', 'nexus-dev')
ON DUPLICATE KEY UPDATE name = 'nexus-dev';
