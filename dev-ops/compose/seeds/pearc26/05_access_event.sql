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

-- PEARC26 seed 05: the event code attendees type on the request form.
-- Requires the access_events migration from the access-requests feature.
--
-- Approving a request with this code creates the account in
-- pearc26-org and a membership on pearc26-allocation.

SET NAMES utf8mb4;
SET time_zone = '+00:00';

INSERT IGNORE INTO access_events (code, compute_allocation_id, organization_id)
VALUES ('PEARC26', 'pearc26-allocation', 'pearc26-org');
