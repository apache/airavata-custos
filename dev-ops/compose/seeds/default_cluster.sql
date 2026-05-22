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

-- Dev-only seed: inserts a single default compute_clusters row so a
-- freshly-wiped dev DB can immediately exercise ClusterAccount creation,
-- AMIE handler flows, and any future feature that requires a cluster
-- reference. NOT applied automatically. Apply after core has run its
-- migrations (which create the compute_clusters table):
--
--   docker exec -i custos_db mariadb -u admin -padmin custos \
--     < dev-ops/compose/seeds/default_cluster.sql
--

INSERT IGNORE INTO compute_clusters (id, name)
VALUES ('00000000-0000-0000-0000-000000000001', 'default-cluster');
