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
-- Core sets ready_at while it has told connectors a membership grants access.
ALTER TABLE compute_allocation_memberships
    ADD COLUMN IF NOT EXISTS ready_at TIMESTAMPTZ(6) NULL;
-- Readies the memberships the scheduler sweep already grants.
UPDATE compute_allocation_memberships m
SET ready_at = now()
FROM compute_allocations a
         JOIN compute_cluster_users cu ON cu.compute_cluster_id = a.compute_cluster_id
WHERE a.id = m.compute_allocation_id
  AND cu.user_id = m.user_id
  AND cu.provisioned_at IS NOT NULL
  AND m.membership_status = 'ACTIVE'
  AND a.status = 'ACTIVE';
-- Core stamps each connector's completion, so it can wait on the scheduler and storage before readiness.
ALTER TABLE compute_allocations
    ADD COLUMN IF NOT EXISTS scheduler_provisioned_at TIMESTAMPTZ(6) NULL;
ALTER TABLE compute_cluster_users
    ADD COLUMN IF NOT EXISTS storage_provisioned_at TIMESTAMPTZ(6) NULL,
    ADD COLUMN IF NOT EXISTS ready_at TIMESTAMPTZ(6) NULL;
-- Existing allocations already have their scheduler account, and provisioned accounts their directories and ready email.
UPDATE compute_allocations SET scheduler_provisioned_at = now();
UPDATE compute_cluster_users
SET storage_provisioned_at = provisioned_at,
    ready_at               = provisioned_at
WHERE provisioned_at IS NOT NULL;
-- Exhausted-request scans match later events for the same entity by payload id.
CREATE INDEX IF NOT EXISTS idx_events_type_entity ON events (event_type, (payload ->> 'id'));
