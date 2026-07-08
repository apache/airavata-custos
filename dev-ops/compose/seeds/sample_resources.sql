-- Licensed to the Apache Software Foundation (ASF) under one or more
-- contributor license agreements. See the NOTICE file distributed with this
-- work for additional information regarding copyright ownership. Licensed
-- under the Apache License, Version 2.0.
--
-- OPTIONAL sample data: compute clusters, allocation resources, and rate
-- histories for portal demos and manual testing of the resources and rates
-- screens. Self-contained (defines its own clusters and resources) and safe
-- to apply on its own after migrations. Rate windows are chosen so every
-- status renders: expired, superseded (overlapped by a later start), active,
-- scheduled, and one resource with no rates at all.
-- Idempotent via INSERT IGNORE + deterministic IDs.

USE custos;

INSERT IGNORE INTO compute_clusters (id, name) VALUES
  ('smpr-cluster-a', 'ClusterA'),
  ('smpr-cluster-b', 'ClusterB');

INSERT IGNORE INTO compute_allocation_resources (id, name, resource_type, resource_amount, compute_cluster_id) VALUES
  ('smpr-res-a-cpu', 'ClusterA CPU Core Hours', 'CPU_HOURS', 2000000, 'smpr-cluster-a'),
  ('smpr-res-a-storage', 'ClusterA Project Storage', 'STORAGE_TB', 500, 'smpr-cluster-a'),
  ('smpr-res-b-gpu', 'ClusterB GPU Hours', 'GPU_HOURS', 150000, 'smpr-cluster-b'),
  ('smpr-res-b-mem', 'ClusterB Large Memory Hours', 'CPU_HOURS', 300000, 'smpr-cluster-b');

-- ClusterA CPU: full lifecycle. The 2025 window is superseded from
-- 2026-05-01 by the later-starting 1.25 rate (latest start wins); a 2027
-- change is already scheduled.
INSERT IGNORE INTO compute_allocation_resource_rates (id, compute_allocation_resource_id, rate, start_time, end_time) VALUES
  ('smpr-rate-acpu-2024', 'smpr-res-a-cpu', 1.00, '2024-01-01 00:00:00', '2025-01-01 00:00:00'),
  ('smpr-rate-acpu-2025', 'smpr-res-a-cpu', 1.10, '2025-01-01 00:00:00', '2026-12-31 00:00:00'),
  ('smpr-rate-acpu-cur',  'smpr-res-a-cpu', 1.25, '2026-05-01 00:00:00', '2027-05-01 00:00:00'),
  ('smpr-rate-acpu-next', 'smpr-res-a-cpu', 1.40, '2027-05-01 00:00:00', '2028-05-01 00:00:00');

-- ClusterB GPU: one active rate and one scheduled increase.
INSERT IGNORE INTO compute_allocation_resource_rates (id, compute_allocation_resource_id, rate, start_time, end_time) VALUES
  ('smpr-rate-bgpu-cur',  'smpr-res-b-gpu', 8.00, '2026-01-01 00:00:00', '2027-01-01 00:00:00'),
  ('smpr-rate-bgpu-next', 'smpr-res-b-gpu', 9.50, '2027-01-01 00:00:00', '2028-01-01 00:00:00');

-- ClusterB Large Memory: a single expired rate, so the resource currently
-- has no effective rate (the gap case).
INSERT IGNORE INTO compute_allocation_resource_rates (id, compute_allocation_resource_id, rate, start_time, end_time) VALUES
  ('smpr-rate-bmem-old', 'smpr-res-b-mem', 2.00, '2024-06-01 00:00:00', '2025-06-01 00:00:00');

-- ClusterA Project Storage intentionally has no rates (empty-state case).

-- Two sample users with provisioned cluster accounts so the cluster local
-- users view has rows. ClusterB intentionally keeps only one account.
INSERT IGNORE INTO organizations (id, originated_id, name) VALUES
  ('sample-org', 'sample-org', 'Sample Research Consortium');

INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smpr-user-nvasquez', 'sample-org', 'Nadia', 'Vasquez', '', 'nadia.vasquez@sample.example.edu', 'ACTIVE', 'CLUSTER_LOCAL'),
  ('smpr-user-tokada', 'sample-org', 'Takeshi', 'Okada', '', 'takeshi.okada@sample.example.edu', 'ACTIVE', 'CLUSTER_LOCAL');

INSERT IGNORE INTO compute_cluster_users (id, compute_cluster_id, user_id, local_username) VALUES
  ('smpr-ccu-a-1', 'smpr-cluster-a', 'smpr-user-nvasquez', 'nvasquez'),
  ('smpr-ccu-a-2', 'smpr-cluster-a', 'smpr-user-tokada', 'tokada'),
  ('smpr-ccu-b-1', 'smpr-cluster-b', 'smpr-user-nvasquez', 'nvasquez');