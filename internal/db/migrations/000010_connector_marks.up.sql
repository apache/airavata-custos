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
-- Core stamps each connector's completion, so it can wait on the scheduler and storage before readiness.
ALTER TABLE compute_allocations
    ADD COLUMN IF NOT EXISTS scheduler_provisioned_at TIMESTAMPTZ(6) NULL;
ALTER TABLE compute_cluster_users
    ADD COLUMN IF NOT EXISTS storage_provisioned_at TIMESTAMPTZ(6) NULL,
    ADD COLUMN IF NOT EXISTS ready_at TIMESTAMPTZ(6) NULL;
-- A subscriber new to a topic is sent the topic's current state; synced_at records that it was.
ALTER TABLE event_subscriptions
    ADD COLUMN IF NOT EXISTS synced_at TIMESTAMPTZ(6) NULL;
UPDATE event_subscriptions SET synced_at = now();
-- Existing allocations already have their scheduler account, and provisioned accounts their directories and ready email.
UPDATE compute_allocations SET scheduler_provisioned_at = now();
UPDATE compute_cluster_users
SET storage_provisioned_at = provisioned_at,
    ready_at               = provisioned_at
WHERE provisioned_at IS NOT NULL;
