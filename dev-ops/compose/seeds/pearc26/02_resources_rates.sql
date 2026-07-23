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
-- gpu is catalogued as gres/gpu, the Slurm TRES for a GPU. The mapper splits
-- it into type gres, name gpu when writing limits, and the usage monitor reads
-- the same pair back. The g1 node (g3.medium) exposes 1 GPU, so capacity is 1.
-- This needs the cluster's gpu node to advertise gres/gpu; until it does, gpu
-- jobs record no gpu usage.

SET NAMES utf8mb4;
SET time_zone = '+00:00';

INSERT IGNORE INTO compute_allocation_resources
    (id, name, resource_type, resource_amount, compute_cluster_id)
VALUES
    ('pearc26-res-debug-cpu', 'debug', 'cpu', 4,
     '00000000-0000-0000-0000-000000000001'),
    ('pearc26-res-gpu', 'gpu', 'gres/gpu', 1,
     '00000000-0000-0000-0000-000000000001');

-- 1 SU per CPU-hour and per GPU-hour, effective through the trial horizon.
INSERT IGNORE INTO compute_allocation_resource_rates
    (id, compute_allocation_resource_id, rate, start_time, end_time)
VALUES
    ('pearc26-rate-debug-cpu', 'pearc26-res-debug-cpu', 1.0,
     '2026-07-01 00:00:00', '2027-07-01 00:00:00'),
    ('pearc26-rate-gpu', 'pearc26-res-gpu', 1.0,
     '2026-07-01 00:00:00', '2027-07-01 00:00:00');
