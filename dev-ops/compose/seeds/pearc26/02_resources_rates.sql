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
-- Read from the live cluster (sinfo / scontrol, 2026-07-16):
--   partition `debug` (default), nodes c[1-2], 2 CPUs + 5500 MB each,
--   4 CPUs total, no GPUs, MaxTime 01:00:00.
-- resource_amount is physical capacity, not a grant; the grant lives in
-- the mapping (seed 04).

SET NAMES utf8mb4;
SET time_zone = '+00:00';

INSERT IGNORE INTO compute_allocation_resources
    (id, name, resource_type, resource_amount, compute_cluster_id)
VALUES
    ('pearc26-res-debug-cpu', 'debug', 'cpu', 4,
     '00000000-0000-0000-0000-000000000001');

-- 1 SU per CPU-hour, effective through the trial horizon.
INSERT IGNORE INTO compute_allocation_resource_rates
    (id, compute_allocation_resource_id, rate, start_time, end_time)
VALUES
    ('pearc26-rate-debug-cpu', 'pearc26-res-debug-cpu', 1.0,
     '2026-07-01 00:00:00', '2027-07-01 00:00:00');
