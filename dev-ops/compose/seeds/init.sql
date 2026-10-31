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

-- What a fresh database needs before anything else: the default cluster and its
-- resource catalog. Apply once after the server has run its migrations:
--
--   docker exec -i custos_db psql -U admin -d custos < dev-ops/compose/seeds/init.sql
--
-- Resource names must match the cluster's SLURM partitions. Usage is matched to a
-- resource by partition, and a job is only charged when a rate covers its end time.

SET TIME ZONE 'UTC';

INSERT INTO compute_clusters (id, name)
VALUES ('00000000-0000-0000-0000-000000000001', 'default-cluster')
ON CONFLICT DO NOTHING;

INSERT INTO compute_allocation_resources (id, name, resource_type, resource_amount, compute_cluster_id) VALUES
  ('default-res-debug', 'debug', 'CPU_HOURS', 1000000, '00000000-0000-0000-0000-000000000001'),
  ('default-res-gpu',   'gpu',   'GPU_HOURS', 100000,  '00000000-0000-0000-0000-000000000001')
ON CONFLICT DO NOTHING;

INSERT INTO compute_allocation_resource_rates (id, compute_allocation_resource_id, rate, start_time, end_time) VALUES
  ('default-rate-debug', 'default-res-debug', 1.00, '2026-01-01 00:00:00', '2028-01-01 00:00:00'),
  ('default-rate-gpu',   'default-res-gpu',   8.00, '2026-01-01 00:00:00', '2028-01-01 00:00:00')
ON CONFLICT DO NOTHING;
