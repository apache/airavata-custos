-- Licensed to the Apache Software Foundation (ASF) under one or more
-- contributor license agreements. See the NOTICE file distributed with this
-- work for additional information regarding copyright ownership. Licensed
-- under the Apache License, Version 2.0.
--
-- OPTIONAL sample data: a rich usage series for the researcher analytics page.
-- Self-contained (defines its own org, users, cluster, resources, project,
-- allocations, and memberships) and safe to apply on its own after migrations.
-- One project with two allocations: one at-risk (about 90% burned, on track to
-- run out before its end date) and one healthy (about 25% burned). Each carries
-- a 35-day daily usage series across GPU, CPU, and storage resources and
-- several members, so the burn-down, usage-over-time, by-resource, and
-- by-member widgets all render with realistic shapes.
--
-- Idempotent via INSERT IGNORE + deterministic IDs (smpu- prefix). To view it
-- signed in as yourself, attach your user to the project once (not committed):
--   INSERT IGNORE INTO project_memberships (project_id, user_id, role, added_time)
--   SELECT 'smpu-project', id, 'PI', NOW(6) FROM users WHERE email = '<your-email>';

USE custos;

INSERT IGNORE INTO organizations (id, originated_id, name) VALUES
  ('smpu-org', 'SMPU-ORG', 'Aurora Research Lab');

INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status) VALUES
  ('smpu-user-pi',   'smpu-org', 'Maria',  'Alvarez', '', 'maria.alvarez@aurora.example.edu',  'ACTIVE'),
  ('smpu-user-mgr',  'smpu-org', 'David',  'Okafor',  '', 'david.okafor@aurora.example.edu',   'ACTIVE'),
  ('smpu-user-mem1', 'smpu-org', 'Lena',   'Petrov',  '', 'lena.petrov@aurora.example.edu',    'ACTIVE'),
  ('smpu-user-mem2', 'smpu-org', 'Sam',    'Nguyen',  '', 'sam.nguyen@aurora.example.edu',     'ACTIVE');

INSERT IGNORE INTO compute_clusters (id, name) VALUES
  ('smpu-cluster', 'Aurora');

INSERT IGNORE INTO compute_allocation_resources (id, name, resource_type, resource_amount, compute_cluster_id) VALUES
  ('smpu-res-gpu',     'B200 GPU Hours',       'GPU_HOURS',  100000, 'smpu-cluster'),
  ('smpu-res-cpu',     'CPU Core Hours',       'CPU_HOURS', 4000000, 'smpu-cluster'),
  ('smpu-res-storage', 'Project Storage',      'STORAGE_TB',    800, 'smpu-cluster');

INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status, created_time) VALUES
  ('smpu-project', 'SMPU-REC-001', 'Aurora Genomics', 'TEST', 'smpu-user-pi', 'ACTIVE', NOW(6));

INSERT IGNORE INTO project_memberships (project_id, user_id, role, added_time) VALUES
  ('smpu-project', 'smpu-user-pi',  'PI',                 NOW(6)),
  ('smpu-project', 'smpu-user-mgr', 'ALLOCATION_MANAGER', NOW(6));

-- At-risk allocation: started 40 days ago, ends 15 days out, ~90% burned.
-- Healthy allocation: started 40 days ago, ends 120 days out, ~25% burned.
INSERT IGNORE INTO compute_allocations
  (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smpu-alloc-atrisk',  'smpu-project', 'Aurora GPU', 'ACTIVE', 'smpu-cluster', 500000,
     DATE_SUB(CURDATE(), INTERVAL 40 DAY), DATE_ADD(CURDATE(), INTERVAL 15 DAY)),
  ('smpu-alloc-healthy', 'smpu-project', 'Aurora Mixed', 'ACTIVE', 'smpu-cluster', 2000000,
     DATE_SUB(CURDATE(), INTERVAL 40 DAY), DATE_ADD(CURDATE(), INTERVAL 120 DAY));

INSERT IGNORE INTO compute_allocation_memberships
  (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smpu-cam-ar-pi',   'smpu-alloc-atrisk',  'smpu-user-pi',   DATE_SUB(CURDATE(), INTERVAL 40 DAY), DATE_ADD(CURDATE(), INTERVAL 15 DAY),  'ACTIVE'),
  ('smpu-cam-ar-mem1', 'smpu-alloc-atrisk',  'smpu-user-mem1', DATE_SUB(CURDATE(), INTERVAL 40 DAY), DATE_ADD(CURDATE(), INTERVAL 15 DAY),  'ACTIVE'),
  ('smpu-cam-ar-mem2', 'smpu-alloc-atrisk',  'smpu-user-mem2', DATE_SUB(CURDATE(), INTERVAL 40 DAY), DATE_ADD(CURDATE(), INTERVAL 15 DAY),  'ACTIVE'),
  ('smpu-cam-he-pi',   'smpu-alloc-healthy', 'smpu-user-pi',   DATE_SUB(CURDATE(), INTERVAL 40 DAY), DATE_ADD(CURDATE(), INTERVAL 120 DAY), 'ACTIVE'),
  ('smpu-cam-he-mem1', 'smpu-alloc-healthy', 'smpu-user-mem1', DATE_SUB(CURDATE(), INTERVAL 40 DAY), DATE_ADD(CURDATE(), INTERVAL 120 DAY), 'ACTIVE');

-- Usage series. Each row below is one (allocation, resource, member, daily
-- rate) stream expanded over 35 days by the recursive day counter. Amounts
-- carry a small deterministic wobble (via the day index) so bars are not flat.
-- su = credits (BIGINT); raw = native units shown in tooltips.
INSERT IGNORE INTO compute_allocation_usages
  (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id)
WITH RECURSIVE days AS (
  SELECT 0 AS n UNION ALL SELECT n + 1 FROM days WHERE n < 34
),
streams AS (
  --      stream_key      alloc                 resource            member            su_base  su_step raw_div
  SELECT 'ar-gpu-pi'  AS k, 'smpu-alloc-atrisk'  AS a, 'smpu-res-gpu'  AS r, 'smpu-user-pi'   AS u, 5400 AS b, 120 AS s, 8  AS rd UNION ALL
  SELECT 'ar-gpu-m1',      'smpu-alloc-atrisk',      'smpu-res-gpu',      'smpu-user-mem1',      2700,     80,      8  UNION ALL
  SELECT 'ar-cpu-m2',      'smpu-alloc-atrisk',      'smpu-res-cpu',      'smpu-user-mem2',      3400,    100,      1  UNION ALL
  SELECT 'he-gpu-pi',      'smpu-alloc-healthy',     'smpu-res-gpu',      'smpu-user-pi',        7600,    160,      8  UNION ALL
  SELECT 'he-cpu-m1',      'smpu-alloc-healthy',     'smpu-res-cpu',      'smpu-user-mem1',      5200,    140,      1  UNION ALL
  SELECT 'he-sto-pi',      'smpu-alloc-healthy',     'smpu-res-storage',  'smpu-user-pi',        1900,     40,     10
)
SELECT
  CONCAT('smpu-use-', st.k, '-', d.n),
  st.a,
  GREATEST(1, (st.b + st.s * ((d.n * 37 + 11) % 9)) DIV st.rd),
  st.b + st.s * ((d.n * 37 + 11) % 9),
  DATE_SUB(CURDATE(), INTERVAL (34 - d.n) DAY) + INTERVAL 12 HOUR,
  st.u,
  CONCAT('job-', st.k, '-', d.n),
  st.r
FROM streams st CROSS JOIN days d;
