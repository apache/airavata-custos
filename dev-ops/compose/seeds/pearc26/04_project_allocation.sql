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

-- PEARC26 seed 04: tutorial project, the allocation every approved
-- attendee joins, and its resource grant. The PI is the bootstrap
-- admin from seed 03, so the super admin operates on the project and
-- allocation directly.
--
-- The allocation name "pearc26-tutorial" is also the Slurm account name
-- and the usage-matching key, so it must stay shell-safe (no spaces).
-- The portal displays it verbatim. 25,000 SU matches the design's
-- decision record.
--
-- LIVE-CLUSTER NOTE: rows inserted here bypass the event bus, so the
-- SLURM Association-Mapper never sees them and no Slurm account is
-- created. Fine for local dev and the mocked test loop. For a live run
-- against nexus-dev, create the project/allocation/mapping through the
-- API instead (the spec's live gate does this) so the events fire.

SET NAMES utf8mb4;
SET time_zone = '+00:00';

INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status)
VALUES ('pearc26-project', 'PEARC26', 'PEARC26 Tutorial Workshop',
        'WORKSHOP', 'pearc26-approver', 'ACTIVE');

INSERT IGNORE INTO project_memberships (project_id, user_id, role)
VALUES ('pearc26-project', 'pearc26-approver', 'PI');

INSERT IGNORE INTO compute_allocations
    (id, project_id, name, status, compute_cluster_id,
     initial_su_amount, start_time, end_time)
VALUES
    ('pearc26-allocation', 'pearc26-project', 'pearc26-tutorial', 'ACTIVE',
     '00000000-0000-0000-0000-000000000001',
     25000, '2026-07-01 00:00:00', '2026-08-31 00:00:00');

-- Grant: all 4 debug-partition CPUs for the allocation window
-- (61 days = 87840 minutes). Native units x minutes, per the mapping
-- convention; the SU budget above is the spend cap.
INSERT IGNORE INTO compute_allocation_resource_mappings
    (id, compute_allocation_id, compute_allocation_resource_id,
     resource_amount, resource_time)
VALUES
    ('pearc26-map-debug-cpu', 'pearc26-allocation', 'pearc26-res-debug-cpu',
     4, 87840);
