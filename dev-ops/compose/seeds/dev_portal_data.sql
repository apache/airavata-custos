-- Licensed to the Apache Software Foundation (ASF) under one or more
-- contributor license agreements. See the NOTICE file distributed with this
-- work for additional information regarding copyright ownership. Licensed
-- under the Apache License, Version 2.0.
--
-- Dev-only seed: populates rows the AMIE protocol does not produce
-- (resources, rates, mappings, usages, change requests, override). Apply
-- AFTER the AMIE baseline scenario has run so the referenced cluster +
-- allocation UUIDs exist. Safe to re-apply (idempotent).

SET NAMES utf8mb4;
SET
time_zone = '+00:00';

-- ---------------------------------------------------------------------------
-- Extend super_admin with the six operational privileges. No-op until the
-- role is created by the Go bootstrap path; re-apply after first server start.
-- ---------------------------------------------------------------------------
INSERT IGNORE INTO role_privileges (role_id, privilege) SELECT id, 'amie:read'    FROM roles WHERE name = 'super_admin';
INSERT IGNORE INTO role_privileges (role_id, privilege) SELECT id, 'amie:write'   FROM roles WHERE name = 'super_admin';
INSERT IGNORE INTO role_privileges (role_id, privilege) SELECT id, 'hpc:read'     FROM roles WHERE name = 'super_admin';
INSERT IGNORE INTO role_privileges (role_id, privilege) SELECT id, 'hpc:write'    FROM roles WHERE name = 'super_admin';
INSERT IGNORE INTO role_privileges (role_id, privilege) SELECT id, 'signer:read'  FROM roles WHERE name = 'super_admin';
INSERT IGNORE INTO role_privileges (role_id, privilege) SELECT id, 'signer:write' FROM roles WHERE name = 'super_admin';

-- ---------------------------------------------------------------------------
-- Resources (cluster-scoped resource catalog)
-- ---------------------------------------------------------------------------
INSERT INTO compute_allocation_resources (id, name, resource_type, resource_amount, compute_cluster_id)
SELECT 'res-cpu', 'CPU Service Units', 'CPU_HOURS', 1000000, id
FROM compute_clusters
WHERE id = '00000000-0000-0000-0000-000000000001' ON DUPLICATE KEY
UPDATE name =
VALUES (name);

INSERT INTO compute_allocation_resources (id, name, resource_type, resource_amount, compute_cluster_id)
SELECT 'res-gpu', 'GPU Service Units', 'GPU_HOURS', 100000, id
FROM compute_clusters
WHERE id = '00000000-0000-0000-0000-000000000001' ON DUPLICATE KEY
UPDATE name =
VALUES (name);

-- ---------------------------------------------------------------------------
-- Resource → Allocation mappings (one CPU mapping per baseline allocation,
-- and one GPU mapping on BL-001 so the UI shows multi-resource state).
-- ---------------------------------------------------------------------------
INSERT INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id,
                                                  resource_amount, resource_time)
SELECT CONCAT('arm-cpu-', SUBSTRING(ca.id, 1, 8)), ca.id, 'res-cpu', 10000, 3600
FROM compute_allocations ca
WHERE ca.compute_cluster_id = '00000000-0000-0000-0000-000000000001' ON DUPLICATE KEY
UPDATE resource_amount =
VALUES (resource_amount);

INSERT INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id,
                                                  resource_amount, resource_time)
SELECT CONCAT('arm-gpu-', SUBSTRING(ca.id, 1, 8)), ca.id, 'res-gpu', 200, 3600
FROM compute_allocations ca
         JOIN projects p ON p.id = ca.project_id
WHERE p.title = 'Baseline Project 1'
   OR p.originated_id = 'BL-001' ON DUPLICATE KEY
UPDATE resource_amount =
VALUES (resource_amount);

-- ---------------------------------------------------------------------------
-- Resource rates
-- ---------------------------------------------------------------------------
INSERT INTO compute_allocation_resource_rates (id, compute_allocation_resource_id, rate, start_time, end_time)
VALUES ('rate-cpu', 'res-cpu', 1.0, '2026-01-01 00:00:00', '2026-12-31 23:59:59'),
       ('rate-gpu', 'res-gpu', 10.0, '2026-01-01 00:00:00', '2026-12-31 23:59:59') ON DUPLICATE KEY
UPDATE rate =
VALUES (rate);

-- ---------------------------------------------------------------------------
-- Usage record (one sample job per baseline allocation)
-- ---------------------------------------------------------------------------
INSERT INTO compute_allocation_usages
(id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id,
 compute_allocation_resource_id)
SELECT CONCAT('usage-', SUBSTRING(ca.id, 1, 8)),
       ca.id,
       250,
       250,
       '2026-06-10 12:00:00',
       u.id,
       CONCAT('job-', SUBSTRING(ca.id, 1, 8)),
       'res-cpu'
FROM compute_allocations ca
         JOIN compute_cluster_users ccu ON ccu.compute_cluster_id = ca.compute_cluster_id
         JOIN users u ON u.id = ccu.user_id
WHERE ca.compute_cluster_id = '00000000-0000-0000-0000-000000000001' ON DUPLICATE KEY
UPDATE used_raw_amount =
VALUES (used_raw_amount);

-- ---------------------------------------------------------------------------
-- Change request + event (one PENDING request on the supplemented allocation)
-- ---------------------------------------------------------------------------
INSERT INTO compute_allocation_change_requests
(id, compute_allocation_id, requested_su_amount, requested_status, reason, change_status, requester_id, approver_id,
 timestamp)
SELECT 'chreq-bl001',
       ca.id,
       25000,
       'ACTIVE',
       'Additional SUs for end-of-year run',
       'PENDING',
       'dev-researcher',
       '',
       '2026-06-12 09:00:00'
FROM compute_allocations ca
         JOIN projects p ON p.id = ca.project_id
WHERE p.title = 'Baseline Project 1'
   OR p.originated_id = 'BL-001' LIMIT 1
ON DUPLICATE KEY
UPDATE reason =
VALUES (reason);

INSERT INTO compute_allocation_change_request_events
(id, compute_allocation_change_request_id, event_type, description, timestamp)
VALUES ('chreq-bl001-e1', 'chreq-bl001', 'CREATED', 'Change request submitted by researcher.', '2026-06-12 09:00:00'),
       ('chreq-bl001-e2', 'chreq-bl001', 'COMMENTED', 'PI reviewed and queued for approval.',
        '2026-06-13 11:30:00') ON DUPLICATE KEY
UPDATE description =
VALUES (description);

-- ---------------------------------------------------------------------------
-- Membership override (one example: cap a member's GPU resource amount)
-- ---------------------------------------------------------------------------
INSERT INTO compute_allocation_membership_resource_overrides
(id, compute_allocation_membership_id, compute_allocation_resource_id, override_resource_amount, override_resource_time)
SELECT CONCAT('over-', SUBSTRING(m.id, 1, 8)), m.id, 'res-gpu', 50, 3600
FROM compute_allocation_memberships m
         JOIN compute_allocations ca ON ca.id = m.compute_allocation_id
         JOIN projects p ON p.id = ca.project_id
WHERE p.title = 'Baseline Project 1'
   OR p.originated_id = 'BL-001' LIMIT 1
ON DUPLICATE KEY
UPDATE override_resource_amount =
VALUES (override_resource_amount);
