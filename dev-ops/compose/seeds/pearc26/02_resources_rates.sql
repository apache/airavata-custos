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

-- PEARC26 seed 02: partition catalog + SU rate for nexus-dev.
--
-- Read from the live cluster (sinfo / scontrol):
--   partition `debug` (default), nodes c[1-2], 4 CPUs total, no GPUs, 1h.
--   partition `gpu`, node g1 (g3.medium: 8 CPUs, 30 GB, 1/4 A100-10GB).
-- resource_amount is physical capacity, not a grant; the grant lives in
-- the mapping (seed 04).
--
-- The resource `name` is the Slurm partition name: the association mapper
-- creates one per-user association per granted resource, scoped to that
-- partition. So a user needs the account to grant BOTH resources to submit
-- on both partitions.
--
-- gpu is catalogued as cpu here on purpose: the mapper sends resource_type
-- straight through as the Slurm TRES type, and Slurm's GPU TRES is
-- `gres/gpu`, not `gpu`. cpu limits are valid today and g3.medium gpu jobs
-- are cpu-bound (DefCpuPerGPU=4). Proper gpu-hour tracking needs a connector
-- change (tracked in docs/drafts/followups.md).

SET NAMES utf8mb4;
SET time_zone = '+00:00';

INSERT IGNORE INTO compute_allocation_resources
    (id, name, resource_type, resource_amount, compute_cluster_id)
VALUES
    ('pearc26-res-debug-cpu', 'debug', 'cpu', 4,
     '00000000-0000-0000-0000-000000000001'),
    ('pearc26-res-gpu-cpu', 'gpu', 'cpu', 8,
     '00000000-0000-0000-0000-000000000001');

-- 1 SU per CPU-hour, effective through the trial horizon.
INSERT IGNORE INTO compute_allocation_resource_rates
    (id, compute_allocation_resource_id, rate, start_time, end_time)
VALUES
    ('pearc26-rate-debug-cpu', 'pearc26-res-debug-cpu', 1.0,
     '2026-07-01 00:00:00', '2027-07-01 00:00:00'),
    ('pearc26-rate-gpu-cpu', 'pearc26-res-gpu-cpu', 1.0,
     '2026-07-01 00:00:00', '2027-07-01 00:00:00');
