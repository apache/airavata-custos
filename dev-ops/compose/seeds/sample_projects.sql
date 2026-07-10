-- Licensed to the Apache Software Foundation (ASF) under one or more
-- contributor license agreements. See the NOTICE file distributed with this
-- work for additional information regarding copyright ownership. Licensed
-- under the Apache License, Version 2.0.
--
-- OPTIONAL sample data: ~20 research projects with allocations, members,
-- usage history, and change requests, for portal demos and manual testing.
-- Not required for a working deployment. Apply after migrations and
-- default_cluster.sql. Idempotent via INSERT IGNORE + deterministic IDs.

USE custos;

INSERT IGNORE INTO organizations (id, originated_id, name) VALUES
  ('sample-org', 'sample-org', 'Sample Research Consortium');

INSERT IGNORE INTO compute_allocation_resources (id, name, resource_type, resource_amount, compute_cluster_id) VALUES
  ('res-cpu', 'CPU Service Units', 'CPU_HOURS', 1000000, '00000000-0000-0000-0000-000000000001'),
  ('res-gpu', 'GPU Service Units', 'GPU_HOURS', 100000, '00000000-0000-0000-0000-000000000001');

INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-344b252c9dd481944db0', 'sample-org', 'Elena', 'Vasquez', '', 'elena.vasquez@sample.example.edu', 'PENDING', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-c8e6b563ac8f348d6069', 'sample-org', 'Marcus', 'Chen', '', 'marcus.chen@sample.example.edu', 'ACTIVE', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-2fd62ff24cea6be8de4c', 'sample-org', 'Priya', 'Raghavan', '', 'priya.raghavan@sample.example.edu', 'ACTIVE', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-bc08d052551918510023', 'sample-org', 'Tomás', 'Silva', '', 'tomas.silva@sample.example.edu', 'PENDING', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-b800fa9f20c414f38e44', 'sample-org', 'Aisha', 'Okafor', '', 'aisha.okafor@sample.example.edu', 'ACTIVE', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-3cd6d104b01ef129b379', 'sample-org', 'David', 'Kim', '', 'david.kim@sample.example.edu', 'ACTIVE', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-90f9c3a872e74ceb9b54', 'sample-org', 'Sarah', 'Lindqvist', '', 'sarah.lindqvist@sample.example.edu', 'PENDING', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-38b35e0e4b22d0d6460a', 'sample-org', 'Rajesh', 'Patel', '', 'rajesh.patel@sample.example.edu', 'ACTIVE', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-26420ea0545ca72a32b3', 'sample-org', 'Mei', 'Zhang', '', 'mei.zhang@sample.example.edu', 'ACTIVE', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-3b164a9395896373810e', 'sample-org', 'James', 'O''Brien', '', 'james.obrien@sample.example.edu', 'PENDING', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-626aede3f65aeb969c74', 'sample-org', 'Fatima', 'Hassan', '', 'fatima.hassan@sample.example.edu', 'ACTIVE', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-4907af7ab30bd84e6deb', 'sample-org', 'Lukas', 'Weber', '', 'lukas.weber@sample.example.edu', 'ACTIVE', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-ae4c14e3cd98f58c2363', 'sample-org', 'Ana', 'Ferreira', '', 'ana.ferreira@sample.example.edu', 'PENDING', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-343175c9b7007d1ff0f8', 'sample-org', 'Hiro', 'Tanaka', '', 'hiro.tanaka@sample.example.edu', 'ACTIVE', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-f908bd32e57fd5c2896d', 'sample-org', 'Nadia', 'Petrova', '', 'nadia.petrova@sample.example.edu', 'ACTIVE', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-f6343c23c6a7028f8ed7', 'sample-org', 'Carlos', 'Mendoza', '', 'carlos.mendoza@sample.example.edu', 'PENDING', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-fa5b4cdf77ce78bd27e0', 'sample-org', 'Ingrid', 'Johansson', '', 'ingrid.johansson@sample.example.edu', 'ACTIVE', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-598aba3fd25a5b74a2cf', 'sample-org', 'Kwame', 'Asante', '', 'kwame.asante@sample.example.edu', 'ACTIVE', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-35b28b0dfed3e7dfcd36', 'sample-org', 'Leila', 'Nasser', '', 'leila.nasser@sample.example.edu', 'PENDING', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-02b1cd6cafbd27dd7201', 'sample-org', 'Viktor', 'Novak', '', 'viktor.novak@sample.example.edu', 'ACTIVE', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-b8e82843b10d1a15b995', 'sample-org', 'Grace', 'Adeyemi', '', 'grace.adeyemi@sample.example.edu', 'ACTIVE', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-318427dda738e0504620', 'sample-org', 'Sven', 'Eriksson', '', 'sven.eriksson@sample.example.edu', 'PENDING', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-52fae2d4d9ff23760057', 'sample-org', 'Rosa', 'Delgado', '', 'rosa.delgado@sample.example.edu', 'ACTIVE', 'CLUSTER_LOCAL');
INSERT IGNORE INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES
  ('smp-8d36f8ec62cc1ab62fd5', 'sample-org', 'Arjun', 'Menon', '', 'arjun.menon@sample.example.edu', 'ACTIVE', 'CLUSTER_LOCAL');

-- CHM240012: Ab Initio Screening of Transition-Metal Catalysts
INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES
  ('smp-4c8ab475ba01ce29fb67', 'CHM240012', 'Ab Initio Screening of Transition-Metal Catalysts', 'access', 'smp-344b252c9dd481944db0', 'INACTIVE');
INSERT IGNORE INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smp-e874cc69f25dfd4b8d9b', 'smp-4c8ab475ba01ce29fb67', 'CHM240012 startup allocation', 'INACTIVE', '00000000-0000-0000-0000-000000000001', 500000, '2026-01-01 00:00:00', '2026-12-31 00:00:00');
INSERT IGNORE INTO project_memberships (project_id, user_id, role) VALUES
  ('smp-4c8ab475ba01ce29fb67', 'smp-344b252c9dd481944db0', 'PI'),
  ('smp-4c8ab475ba01ce29fb67', 'smp-3cd6d104b01ef129b379', 'CO_PI'),
  ('smp-4c8ab475ba01ce29fb67', 'smp-4907af7ab30bd84e6deb', 'ALLOCATION_MANAGER');
INSERT IGNORE INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smp-03f67ac7ec854e58822e', 'smp-e874cc69f25dfd4b8d9b', 'smp-344b252c9dd481944db0', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'INACTIVE'),
  ('smp-00ecfa5a2f215d673736', 'smp-e874cc69f25dfd4b8d9b', 'smp-3cd6d104b01ef129b379', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-4fe04464197b2f2fa95c', 'smp-e874cc69f25dfd4b8d9b', 'smp-2fd62ff24cea6be8de4c', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-5e607b81210b9bed2acd', 'smp-e874cc69f25dfd4b8d9b', 'smp-38b35e0e4b22d0d6460a', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-71cf0b7ab64664364795', 'smp-e874cc69f25dfd4b8d9b', 'smp-343175c9b7007d1ff0f8', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE');
INSERT IGNORE INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time) VALUES
  ('smp-cb33ed122c36262d7ce3', 'smp-e874cc69f25dfd4b8d9b', 'res-cpu', 5000, 3600),
  ('smp-7baff5fb5af1b937d1c2', 'smp-e874cc69f25dfd4b8d9b', 'res-gpu', 250, 3600);
INSERT IGNORE INTO compute_allocation_usages (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id) VALUES
  ('smp-530ad0d71ddc54f4b616', 'smp-e874cc69f25dfd4b8d9b', 25122, 25122, '2026-02-15 12:00:00', 'smp-344b252c9dd481944db0', 'job-chm240012-02', 'res-cpu'),
  ('smp-105c2ba6727a17d131e0', 'smp-e874cc69f25dfd4b8d9b', 25244, 25244, '2026-04-15 12:00:00', 'smp-344b252c9dd481944db0', 'job-chm240012-04', 'res-cpu'),
  ('smp-d66812cbdfc4d5f9754a', 'smp-e874cc69f25dfd4b8d9b', 25366, 25366, '2026-06-15 12:00:00', 'smp-344b252c9dd481944db0', 'job-chm240012-06', 'res-cpu'),
  ('smp-5222a7b3a5ed5aa06b6b', 'smp-e874cc69f25dfd4b8d9b', 25259, 25259, '2026-02-15 12:00:00', 'smp-2fd62ff24cea6be8de4c', 'job-chm240012-12', 'res-cpu'),
  ('smp-ae21cf27ab83fab933e9', 'smp-e874cc69f25dfd4b8d9b', 25381, 25381, '2026-04-15 12:00:00', 'smp-2fd62ff24cea6be8de4c', 'job-chm240012-14', 'res-cpu'),
  ('smp-6826a5eafd15dcd7cbd1', 'smp-e874cc69f25dfd4b8d9b', 25503, 25503, '2026-06-15 12:00:00', 'smp-2fd62ff24cea6be8de4c', 'job-chm240012-16', 'res-cpu'),
  ('smp-6a4ce2db6e7755a2f0f0', 'smp-e874cc69f25dfd4b8d9b', 25396, 25396, '2026-02-15 12:00:00', 'smp-38b35e0e4b22d0d6460a', 'job-chm240012-22', 'res-cpu'),
  ('smp-de8cfe7a9daa94af63dd', 'smp-e874cc69f25dfd4b8d9b', 25518, 25518, '2026-04-15 12:00:00', 'smp-38b35e0e4b22d0d6460a', 'job-chm240012-24', 'res-cpu'),
  ('smp-bab3e2728afe00548e12', 'smp-e874cc69f25dfd4b8d9b', 25640, 25640, '2026-06-15 12:00:00', 'smp-38b35e0e4b22d0d6460a', 'job-chm240012-26', 'res-cpu');
INSERT IGNORE INTO compute_allocation_change_requests (id, compute_allocation_id, requested_su_amount, requested_status, reason, change_status, requester_id, approver_id, timestamp) VALUES
  ('smp-bee39eb5dcf4d798dadd', 'smp-e874cc69f25dfd4b8d9b', 250000, 'ACTIVE', 'Supplement for CHM240012 production campaign', 'PENDING', 'smp-344b252c9dd481944db0', '', '2026-06-20 09:00:00');
INSERT IGNORE INTO compute_allocation_change_request_events (id, compute_allocation_change_request_id, event_type, description, timestamp) VALUES
  ('smp-bee39eb5dcf4d798dadd-e1', 'smp-bee39eb5dcf4d798dadd', 'CREATED', 'Change request submitted by PI.', '2026-06-20 09:00:00');

-- PHY240034: Lattice QCD at the Physical Point
INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES
  ('smp-eedc3ed54172fa1539ee', 'PHY240034', 'Lattice QCD at the Physical Point', 'access', 'smp-c8e6b563ac8f348d6069', 'ACTIVE');
INSERT IGNORE INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smp-d29064eb686184d12cbe', 'smp-eedc3ed54172fa1539ee', 'PHY240034 startup allocation', 'ACTIVE', '00000000-0000-0000-0000-000000000001', 250000, '2026-01-01 00:00:00', '2026-12-31 00:00:00');
INSERT IGNORE INTO project_memberships (project_id, user_id, role) VALUES
  ('smp-eedc3ed54172fa1539ee', 'smp-c8e6b563ac8f348d6069', 'PI'),
  ('smp-eedc3ed54172fa1539ee', 'smp-90f9c3a872e74ceb9b54', 'CO_PI'),
  ('smp-eedc3ed54172fa1539ee', 'smp-ae4c14e3cd98f58c2363', 'ALLOCATION_MANAGER');
INSERT IGNORE INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smp-dff1bea8f3c9fc390dc2', 'smp-d29064eb686184d12cbe', 'smp-c8e6b563ac8f348d6069', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-7e3a683d1b228d3fd397', 'smp-d29064eb686184d12cbe', 'smp-90f9c3a872e74ceb9b54', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-781a6da2739d158a9381', 'smp-d29064eb686184d12cbe', 'smp-bc08d052551918510023', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-11ef2821527195380666', 'smp-d29064eb686184d12cbe', 'smp-26420ea0545ca72a32b3', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-8f07e38746a089ac45ce', 'smp-d29064eb686184d12cbe', 'smp-f908bd32e57fd5c2896d', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'INACTIVE');
INSERT IGNORE INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time) VALUES
  ('smp-81ce9b9915d4330c9618', 'smp-d29064eb686184d12cbe', 'res-cpu', 2500, 3600);
INSERT IGNORE INTO compute_allocation_usages (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id) VALUES
  ('smp-93a9297f42034c4f6c54', 'smp-d29064eb686184d12cbe', 12622, 12622, '2026-02-15 12:00:00', 'smp-c8e6b563ac8f348d6069', 'job-phy240034-02', 'res-cpu'),
  ('smp-182f0db7f97cb88e60ca', 'smp-d29064eb686184d12cbe', 12744, 12744, '2026-04-15 12:00:00', 'smp-c8e6b563ac8f348d6069', 'job-phy240034-04', 'res-cpu'),
  ('smp-de38d27ef1afee55bc46', 'smp-d29064eb686184d12cbe', 12866, 12866, '2026-06-15 12:00:00', 'smp-c8e6b563ac8f348d6069', 'job-phy240034-06', 'res-cpu'),
  ('smp-50909688e38a289c775e', 'smp-d29064eb686184d12cbe', 12759, 12759, '2026-02-15 12:00:00', 'smp-bc08d052551918510023', 'job-phy240034-12', 'res-cpu'),
  ('smp-1a7cf4b898e66519bbfa', 'smp-d29064eb686184d12cbe', 12881, 12881, '2026-04-15 12:00:00', 'smp-bc08d052551918510023', 'job-phy240034-14', 'res-cpu'),
  ('smp-eb632ef53e053b61be18', 'smp-d29064eb686184d12cbe', 13003, 13003, '2026-06-15 12:00:00', 'smp-bc08d052551918510023', 'job-phy240034-16', 'res-cpu'),
  ('smp-fe7ed21602138bb5d6bd', 'smp-d29064eb686184d12cbe', 12896, 12896, '2026-02-15 12:00:00', 'smp-26420ea0545ca72a32b3', 'job-phy240034-22', 'res-cpu'),
  ('smp-fcb629755d1e9143f58f', 'smp-d29064eb686184d12cbe', 13018, 13018, '2026-04-15 12:00:00', 'smp-26420ea0545ca72a32b3', 'job-phy240034-24', 'res-cpu'),
  ('smp-a528a8ce9397f2f2c14e', 'smp-d29064eb686184d12cbe', 13140, 13140, '2026-06-15 12:00:00', 'smp-26420ea0545ca72a32b3', 'job-phy240034-26', 'res-cpu');

-- BIO240041: Cryo-EM Reconstruction of Membrane Transporters
INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES
  ('smp-9a75381c02101891703d', 'BIO240041', 'Cryo-EM Reconstruction of Membrane Transporters', 'access', 'smp-2fd62ff24cea6be8de4c', 'ACTIVE');
INSERT IGNORE INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smp-4d0e44e3216d12f4d479', 'smp-9a75381c02101891703d', 'BIO240041 startup allocation', 'ACTIVE', '00000000-0000-0000-0000-000000000001', 1200000, '2026-01-01 00:00:00', '2026-12-31 00:00:00');
INSERT IGNORE INTO project_memberships (project_id, user_id, role) VALUES
  ('smp-9a75381c02101891703d', 'smp-2fd62ff24cea6be8de4c', 'PI'),
  ('smp-9a75381c02101891703d', 'smp-38b35e0e4b22d0d6460a', 'CO_PI'),
  ('smp-9a75381c02101891703d', 'smp-343175c9b7007d1ff0f8', 'ALLOCATION_MANAGER');
INSERT IGNORE INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smp-a3ab9b4d9c40788c1903', 'smp-4d0e44e3216d12f4d479', 'smp-2fd62ff24cea6be8de4c', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-3267f51d3012259d259a', 'smp-4d0e44e3216d12f4d479', 'smp-38b35e0e4b22d0d6460a', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-629ec6cc62ffb28c951a', 'smp-4d0e44e3216d12f4d479', 'smp-b800fa9f20c414f38e44', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-df43769412be97bff936', 'smp-4d0e44e3216d12f4d479', 'smp-3b164a9395896373810e', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'INACTIVE'),
  ('smp-ceeef3865b4d552302af', 'smp-4d0e44e3216d12f4d479', 'smp-f6343c23c6a7028f8ed7', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE');
INSERT IGNORE INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time) VALUES
  ('smp-3adf9e4196cfe653ed10', 'smp-4d0e44e3216d12f4d479', 'res-cpu', 12000, 3600);
INSERT IGNORE INTO compute_allocation_usages (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id) VALUES
  ('smp-f9b40c25d80a709c32e4', 'smp-4d0e44e3216d12f4d479', 60122, 60122, '2026-02-15 12:00:00', 'smp-2fd62ff24cea6be8de4c', 'job-bio240041-02', 'res-cpu'),
  ('smp-ee6c9eb02449247cee7e', 'smp-4d0e44e3216d12f4d479', 60244, 60244, '2026-04-15 12:00:00', 'smp-2fd62ff24cea6be8de4c', 'job-bio240041-04', 'res-cpu'),
  ('smp-a8db1b3ea4e59139a759', 'smp-4d0e44e3216d12f4d479', 60366, 60366, '2026-06-15 12:00:00', 'smp-2fd62ff24cea6be8de4c', 'job-bio240041-06', 'res-cpu'),
  ('smp-409930454eb546fe69f1', 'smp-4d0e44e3216d12f4d479', 60259, 60259, '2026-02-15 12:00:00', 'smp-b800fa9f20c414f38e44', 'job-bio240041-12', 'res-cpu'),
  ('smp-a2868ad098bf8dca568e', 'smp-4d0e44e3216d12f4d479', 60381, 60381, '2026-04-15 12:00:00', 'smp-b800fa9f20c414f38e44', 'job-bio240041-14', 'res-cpu'),
  ('smp-6dc600c77b58f77802c9', 'smp-4d0e44e3216d12f4d479', 60503, 60503, '2026-06-15 12:00:00', 'smp-b800fa9f20c414f38e44', 'job-bio240041-16', 'res-cpu'),
  ('smp-f932d2e828db398abb2c', 'smp-4d0e44e3216d12f4d479', 60396, 60396, '2026-02-15 12:00:00', 'smp-3b164a9395896373810e', 'job-bio240041-22', 'res-cpu'),
  ('smp-dd39cc232b646d084e0b', 'smp-4d0e44e3216d12f4d479', 60518, 60518, '2026-04-15 12:00:00', 'smp-3b164a9395896373810e', 'job-bio240041-24', 'res-cpu'),
  ('smp-c822b6cfa2e56941eaaf', 'smp-4d0e44e3216d12f4d479', 60640, 60640, '2026-06-15 12:00:00', 'smp-3b164a9395896373810e', 'job-bio240041-26', 'res-cpu');

-- ATM240007: Convection-Permitting Regional Climate Downscaling
INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES
  ('smp-739a563efab75e5ab41a', 'ATM240007', 'Convection-Permitting Regional Climate Downscaling', 'access', 'smp-bc08d052551918510023', 'ACTIVE');
INSERT IGNORE INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smp-2551d58902d6316834f7', 'smp-739a563efab75e5ab41a', 'ATM240007 startup allocation', 'ACTIVE', '00000000-0000-0000-0000-000000000001', 100000, '2026-01-01 00:00:00', '2026-12-31 00:00:00');
INSERT IGNORE INTO project_memberships (project_id, user_id, role) VALUES
  ('smp-739a563efab75e5ab41a', 'smp-bc08d052551918510023', 'PI'),
  ('smp-739a563efab75e5ab41a', 'smp-26420ea0545ca72a32b3', 'CO_PI'),
  ('smp-739a563efab75e5ab41a', 'smp-f908bd32e57fd5c2896d', 'ALLOCATION_MANAGER');
INSERT IGNORE INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smp-f646fe054930300ae250', 'smp-2551d58902d6316834f7', 'smp-bc08d052551918510023', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-83a967a84b8f97fcc7f8', 'smp-2551d58902d6316834f7', 'smp-26420ea0545ca72a32b3', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-f6b5af8ead74db587789', 'smp-2551d58902d6316834f7', 'smp-3cd6d104b01ef129b379', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'INACTIVE'),
  ('smp-8605e7b86b9d576f483b', 'smp-2551d58902d6316834f7', 'smp-626aede3f65aeb969c74', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-fd74a21ed701b3fde48c', 'smp-2551d58902d6316834f7', 'smp-fa5b4cdf77ce78bd27e0', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE');
INSERT IGNORE INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time) VALUES
  ('smp-cd993b04e4f2c02b88f8', 'smp-2551d58902d6316834f7', 'res-cpu', 1000, 3600),
  ('smp-32db965752251c364028', 'smp-2551d58902d6316834f7', 'res-gpu', 50, 3600);
INSERT IGNORE INTO compute_allocation_usages (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id) VALUES
  ('smp-5c988efc4a580a59b89d', 'smp-2551d58902d6316834f7', 5122, 5122, '2026-02-15 12:00:00', 'smp-bc08d052551918510023', 'job-atm240007-02', 'res-cpu'),
  ('smp-acb54b7dbe6529188a0c', 'smp-2551d58902d6316834f7', 5244, 5244, '2026-04-15 12:00:00', 'smp-bc08d052551918510023', 'job-atm240007-04', 'res-cpu'),
  ('smp-84b534468be86ac704e3', 'smp-2551d58902d6316834f7', 5366, 5366, '2026-06-15 12:00:00', 'smp-bc08d052551918510023', 'job-atm240007-06', 'res-cpu'),
  ('smp-a5a772ba09c1faa5fd80', 'smp-2551d58902d6316834f7', 5259, 5259, '2026-02-15 12:00:00', 'smp-3cd6d104b01ef129b379', 'job-atm240007-12', 'res-cpu'),
  ('smp-310899e8eb00b1bbb13b', 'smp-2551d58902d6316834f7', 5381, 5381, '2026-04-15 12:00:00', 'smp-3cd6d104b01ef129b379', 'job-atm240007-14', 'res-cpu'),
  ('smp-228ad1f5a429f1b97327', 'smp-2551d58902d6316834f7', 5503, 5503, '2026-06-15 12:00:00', 'smp-3cd6d104b01ef129b379', 'job-atm240007-16', 'res-cpu'),
  ('smp-e1a05cb0fa2fed877d41', 'smp-2551d58902d6316834f7', 5396, 5396, '2026-02-15 12:00:00', 'smp-626aede3f65aeb969c74', 'job-atm240007-22', 'res-cpu'),
  ('smp-b8bd552c0e4ccee94b45', 'smp-2551d58902d6316834f7', 5518, 5518, '2026-04-15 12:00:00', 'smp-626aede3f65aeb969c74', 'job-atm240007-24', 'res-cpu'),
  ('smp-a656878e65a300e21644', 'smp-2551d58902d6316834f7', 5640, 5640, '2026-06-15 12:00:00', 'smp-626aede3f65aeb969c74', 'job-atm240007-26', 'res-cpu');

-- MAT240019: Phase-Field Modeling of Additive Manufacturing
INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES
  ('smp-5591ed41e4391631471d', 'MAT240019', 'Phase-Field Modeling of Additive Manufacturing', 'access', 'smp-b800fa9f20c414f38e44', 'ACTIVE');
INSERT IGNORE INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smp-e1d2a3f5f134af253407', 'smp-5591ed41e4391631471d', 'MAT240019 startup allocation', 'ACTIVE', '00000000-0000-0000-0000-000000000001', 750000, '2026-01-01 00:00:00', '2026-12-31 00:00:00');
INSERT IGNORE INTO project_memberships (project_id, user_id, role) VALUES
  ('smp-5591ed41e4391631471d', 'smp-b800fa9f20c414f38e44', 'PI'),
  ('smp-5591ed41e4391631471d', 'smp-3b164a9395896373810e', 'CO_PI'),
  ('smp-5591ed41e4391631471d', 'smp-f6343c23c6a7028f8ed7', 'ALLOCATION_MANAGER');
INSERT IGNORE INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smp-e3e592f0cf08e6810c4a', 'smp-e1d2a3f5f134af253407', 'smp-b800fa9f20c414f38e44', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-c1175f5f185e611386bc', 'smp-e1d2a3f5f134af253407', 'smp-3b164a9395896373810e', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'INACTIVE'),
  ('smp-6fbefdb58caf23d4ec9a', 'smp-e1d2a3f5f134af253407', 'smp-90f9c3a872e74ceb9b54', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-846394a530902c49b71f', 'smp-e1d2a3f5f134af253407', 'smp-4907af7ab30bd84e6deb', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-3d456ef04b41df4b6b21', 'smp-e1d2a3f5f134af253407', 'smp-598aba3fd25a5b74a2cf', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE');
INSERT IGNORE INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time) VALUES
  ('smp-78e31b4a2a87e668ede9', 'smp-e1d2a3f5f134af253407', 'res-cpu', 7500, 3600);
INSERT IGNORE INTO compute_allocation_usages (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id) VALUES
  ('smp-94012f10b2837f8e70ac', 'smp-e1d2a3f5f134af253407', 37622, 37622, '2026-02-15 12:00:00', 'smp-b800fa9f20c414f38e44', 'job-mat240019-02', 'res-cpu'),
  ('smp-4bf63c33ab7ad750bbcb', 'smp-e1d2a3f5f134af253407', 37744, 37744, '2026-04-15 12:00:00', 'smp-b800fa9f20c414f38e44', 'job-mat240019-04', 'res-cpu'),
  ('smp-1f45c13ef3fb9504c686', 'smp-e1d2a3f5f134af253407', 37866, 37866, '2026-06-15 12:00:00', 'smp-b800fa9f20c414f38e44', 'job-mat240019-06', 'res-cpu'),
  ('smp-d6a9533a80eb12507803', 'smp-e1d2a3f5f134af253407', 37759, 37759, '2026-02-15 12:00:00', 'smp-90f9c3a872e74ceb9b54', 'job-mat240019-12', 'res-cpu'),
  ('smp-35f5a1dd5e0c6768d31e', 'smp-e1d2a3f5f134af253407', 37881, 37881, '2026-04-15 12:00:00', 'smp-90f9c3a872e74ceb9b54', 'job-mat240019-14', 'res-cpu'),
  ('smp-4c4f22d12ad43f31a1eb', 'smp-e1d2a3f5f134af253407', 38003, 38003, '2026-06-15 12:00:00', 'smp-90f9c3a872e74ceb9b54', 'job-mat240019-16', 'res-cpu'),
  ('smp-c0202dbaa96535442ec2', 'smp-e1d2a3f5f134af253407', 37896, 37896, '2026-02-15 12:00:00', 'smp-4907af7ab30bd84e6deb', 'job-mat240019-22', 'res-cpu'),
  ('smp-7ff22c79168831d62ee2', 'smp-e1d2a3f5f134af253407', 38018, 38018, '2026-04-15 12:00:00', 'smp-4907af7ab30bd84e6deb', 'job-mat240019-24', 'res-cpu'),
  ('smp-4892ec0a6d21e1cdd04c', 'smp-e1d2a3f5f134af253407', 38140, 38140, '2026-06-15 12:00:00', 'smp-4907af7ab30bd84e6deb', 'job-mat240019-26', 'res-cpu');
INSERT IGNORE INTO compute_allocation_change_requests (id, compute_allocation_id, requested_su_amount, requested_status, reason, change_status, requester_id, approver_id, timestamp) VALUES
  ('smp-254f9b0991e8ba676b9e', 'smp-e1d2a3f5f134af253407', 375000, 'ACTIVE', 'Supplement for MAT240019 production campaign', 'APPROVED', 'smp-b800fa9f20c414f38e44', '', '2026-06-24 09:00:00');
INSERT IGNORE INTO compute_allocation_change_request_events (id, compute_allocation_change_request_id, event_type, description, timestamp) VALUES
  ('smp-254f9b0991e8ba676b9e-e1', 'smp-254f9b0991e8ba676b9e', 'CREATED', 'Change request submitted by PI.', '2026-06-24 09:00:00');

-- AST240023: Galaxy Formation in Cosmological Zoom Simulations
INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES
  ('smp-2acc6fcc75db53616416', 'AST240023', 'Galaxy Formation in Cosmological Zoom Simulations', 'access', 'smp-3cd6d104b01ef129b379', 'ACTIVE');
INSERT IGNORE INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smp-603a9403849f5bf1779d', 'smp-2acc6fcc75db53616416', 'AST240023 startup allocation', 'ACTIVE', '00000000-0000-0000-0000-000000000001', 300000, '2026-01-01 00:00:00', '2026-12-31 00:00:00');
INSERT IGNORE INTO project_memberships (project_id, user_id, role) VALUES
  ('smp-2acc6fcc75db53616416', 'smp-3cd6d104b01ef129b379', 'PI'),
  ('smp-2acc6fcc75db53616416', 'smp-626aede3f65aeb969c74', 'CO_PI'),
  ('smp-2acc6fcc75db53616416', 'smp-fa5b4cdf77ce78bd27e0', 'ALLOCATION_MANAGER');
INSERT IGNORE INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smp-2a7e4efefba9ce923ebe', 'smp-603a9403849f5bf1779d', 'smp-3cd6d104b01ef129b379', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'INACTIVE'),
  ('smp-2caa26d94f937425d1a4', 'smp-603a9403849f5bf1779d', 'smp-626aede3f65aeb969c74', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-c41e4253e1dc415f9d0f', 'smp-603a9403849f5bf1779d', 'smp-38b35e0e4b22d0d6460a', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-4b581a2c13cf595a4883', 'smp-603a9403849f5bf1779d', 'smp-ae4c14e3cd98f58c2363', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-f065f61b096e0c240128', 'smp-603a9403849f5bf1779d', 'smp-35b28b0dfed3e7dfcd36', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE');
INSERT IGNORE INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time) VALUES
  ('smp-2a8b2229da7f623bb18a', 'smp-603a9403849f5bf1779d', 'res-cpu', 3000, 3600);
INSERT IGNORE INTO compute_allocation_usages (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id) VALUES
  ('smp-6119d8a1f8d04129f3f8', 'smp-603a9403849f5bf1779d', 15122, 15122, '2026-02-15 12:00:00', 'smp-3cd6d104b01ef129b379', 'job-ast240023-02', 'res-cpu'),
  ('smp-647181201e12d96b2b62', 'smp-603a9403849f5bf1779d', 15244, 15244, '2026-04-15 12:00:00', 'smp-3cd6d104b01ef129b379', 'job-ast240023-04', 'res-cpu'),
  ('smp-28df35315256eb87cb4f', 'smp-603a9403849f5bf1779d', 15366, 15366, '2026-06-15 12:00:00', 'smp-3cd6d104b01ef129b379', 'job-ast240023-06', 'res-cpu'),
  ('smp-f4034b96c8c2ceb41e9f', 'smp-603a9403849f5bf1779d', 15259, 15259, '2026-02-15 12:00:00', 'smp-38b35e0e4b22d0d6460a', 'job-ast240023-12', 'res-cpu'),
  ('smp-a340f3315559c8acea83', 'smp-603a9403849f5bf1779d', 15381, 15381, '2026-04-15 12:00:00', 'smp-38b35e0e4b22d0d6460a', 'job-ast240023-14', 'res-cpu'),
  ('smp-1cf9d1e335c97e2b8952', 'smp-603a9403849f5bf1779d', 15503, 15503, '2026-06-15 12:00:00', 'smp-38b35e0e4b22d0d6460a', 'job-ast240023-16', 'res-cpu'),
  ('smp-3466fea28ce65b64f3f3', 'smp-603a9403849f5bf1779d', 15396, 15396, '2026-02-15 12:00:00', 'smp-ae4c14e3cd98f58c2363', 'job-ast240023-22', 'res-cpu'),
  ('smp-9bfa80727d3c8049409b', 'smp-603a9403849f5bf1779d', 15518, 15518, '2026-04-15 12:00:00', 'smp-ae4c14e3cd98f58c2363', 'job-ast240023-24', 'res-cpu'),
  ('smp-70407e7274fdb84e6670', 'smp-603a9403849f5bf1779d', 15640, 15640, '2026-06-15 12:00:00', 'smp-ae4c14e3cd98f58c2363', 'job-ast240023-26', 'res-cpu');

-- CHE240055: Coarse-Grained Dynamics of Polymer Electrolytes
INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES
  ('smp-b7545fb17c505974fac7', 'CHE240055', 'Coarse-Grained Dynamics of Polymer Electrolytes', 'access', 'smp-90f9c3a872e74ceb9b54', 'ACTIVE');
INSERT IGNORE INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smp-50a930bb1fb569beb058', 'smp-b7545fb17c505974fac7', 'CHE240055 startup allocation', 'INACTIVE', '00000000-0000-0000-0000-000000000001', 500000, '2026-01-01 00:00:00', '2026-12-31 00:00:00');
INSERT IGNORE INTO project_memberships (project_id, user_id, role) VALUES
  ('smp-b7545fb17c505974fac7', 'smp-90f9c3a872e74ceb9b54', 'PI'),
  ('smp-b7545fb17c505974fac7', 'smp-4907af7ab30bd84e6deb', 'CO_PI'),
  ('smp-b7545fb17c505974fac7', 'smp-598aba3fd25a5b74a2cf', 'ALLOCATION_MANAGER');
INSERT IGNORE INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smp-df1b9ec3ffc9f799f8c7', 'smp-50a930bb1fb569beb058', 'smp-90f9c3a872e74ceb9b54', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-d9b23eb43e735256ff09', 'smp-50a930bb1fb569beb058', 'smp-4907af7ab30bd84e6deb', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-0f69ddeda7cf2638c12c', 'smp-50a930bb1fb569beb058', 'smp-26420ea0545ca72a32b3', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-4e05771035c6043e8ce7', 'smp-50a930bb1fb569beb058', 'smp-343175c9b7007d1ff0f8', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-015a0efad9c581a798c3', 'smp-50a930bb1fb569beb058', 'smp-02b1cd6cafbd27dd7201', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'INACTIVE');
INSERT IGNORE INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time) VALUES
  ('smp-585418a12468f375fb17', 'smp-50a930bb1fb569beb058', 'res-cpu', 5000, 3600),
  ('smp-22ebd1572fb944db1fef', 'smp-50a930bb1fb569beb058', 'res-gpu', 250, 3600);
INSERT IGNORE INTO compute_allocation_usages (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id) VALUES
  ('smp-7303d44888b9449db6a1', 'smp-50a930bb1fb569beb058', 25122, 25122, '2026-02-15 12:00:00', 'smp-90f9c3a872e74ceb9b54', 'job-che240055-02', 'res-cpu'),
  ('smp-6833b38aa98debeb791c', 'smp-50a930bb1fb569beb058', 25244, 25244, '2026-04-15 12:00:00', 'smp-90f9c3a872e74ceb9b54', 'job-che240055-04', 'res-cpu'),
  ('smp-b53af282aaef9a521e26', 'smp-50a930bb1fb569beb058', 25366, 25366, '2026-06-15 12:00:00', 'smp-90f9c3a872e74ceb9b54', 'job-che240055-06', 'res-cpu'),
  ('smp-06af0ab1514506801fe0', 'smp-50a930bb1fb569beb058', 25259, 25259, '2026-02-15 12:00:00', 'smp-26420ea0545ca72a32b3', 'job-che240055-12', 'res-cpu'),
  ('smp-acd31c1c8e35266a6744', 'smp-50a930bb1fb569beb058', 25381, 25381, '2026-04-15 12:00:00', 'smp-26420ea0545ca72a32b3', 'job-che240055-14', 'res-cpu'),
  ('smp-65746ddfa778564f94cf', 'smp-50a930bb1fb569beb058', 25503, 25503, '2026-06-15 12:00:00', 'smp-26420ea0545ca72a32b3', 'job-che240055-16', 'res-cpu'),
  ('smp-3cea6aa6fef3a2a2c02c', 'smp-50a930bb1fb569beb058', 25396, 25396, '2026-02-15 12:00:00', 'smp-343175c9b7007d1ff0f8', 'job-che240055-22', 'res-cpu'),
  ('smp-beb62c91dc7e783233e6', 'smp-50a930bb1fb569beb058', 25518, 25518, '2026-04-15 12:00:00', 'smp-343175c9b7007d1ff0f8', 'job-che240055-24', 'res-cpu'),
  ('smp-6bdbddae9f82d1257645', 'smp-50a930bb1fb569beb058', 25640, 25640, '2026-06-15 12:00:00', 'smp-343175c9b7007d1ff0f8', 'job-che240055-26', 'res-cpu');

-- GEO240011: Full-Waveform Inversion for Cascadia Subduction
INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES
  ('smp-3df742c724053f0bbfcc', 'GEO240011', 'Full-Waveform Inversion for Cascadia Subduction', 'access', 'smp-38b35e0e4b22d0d6460a', 'INACTIVE');
INSERT IGNORE INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smp-903e6a1acd2c67f92e00', 'smp-3df742c724053f0bbfcc', 'GEO240011 startup allocation', 'ACTIVE', '00000000-0000-0000-0000-000000000001', 250000, '2026-01-01 00:00:00', '2026-12-31 00:00:00');
INSERT IGNORE INTO project_memberships (project_id, user_id, role) VALUES
  ('smp-3df742c724053f0bbfcc', 'smp-38b35e0e4b22d0d6460a', 'PI'),
  ('smp-3df742c724053f0bbfcc', 'smp-ae4c14e3cd98f58c2363', 'CO_PI'),
  ('smp-3df742c724053f0bbfcc', 'smp-35b28b0dfed3e7dfcd36', 'ALLOCATION_MANAGER');
INSERT IGNORE INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smp-6824f6e16aaacea2c844', 'smp-903e6a1acd2c67f92e00', 'smp-38b35e0e4b22d0d6460a', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-7bab7f463ea43025f80e', 'smp-903e6a1acd2c67f92e00', 'smp-ae4c14e3cd98f58c2363', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-21fa0da14c6bed4666d1', 'smp-903e6a1acd2c67f92e00', 'smp-3b164a9395896373810e', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-c3110fe149f18a80ded9', 'smp-903e6a1acd2c67f92e00', 'smp-f908bd32e57fd5c2896d', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'INACTIVE'),
  ('smp-d4fee66ef35b31194264', 'smp-903e6a1acd2c67f92e00', 'smp-b8e82843b10d1a15b995', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE');
INSERT IGNORE INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time) VALUES
  ('smp-b2d422b17c67c38949ba', 'smp-903e6a1acd2c67f92e00', 'res-cpu', 2500, 3600);
INSERT IGNORE INTO compute_allocation_usages (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id) VALUES
  ('smp-45ad6b15a89566275d44', 'smp-903e6a1acd2c67f92e00', 12622, 12622, '2026-02-15 12:00:00', 'smp-38b35e0e4b22d0d6460a', 'job-geo240011-02', 'res-cpu'),
  ('smp-a0098b972e16b7584386', 'smp-903e6a1acd2c67f92e00', 12744, 12744, '2026-04-15 12:00:00', 'smp-38b35e0e4b22d0d6460a', 'job-geo240011-04', 'res-cpu'),
  ('smp-4b6185e17858dac52571', 'smp-903e6a1acd2c67f92e00', 12866, 12866, '2026-06-15 12:00:00', 'smp-38b35e0e4b22d0d6460a', 'job-geo240011-06', 'res-cpu'),
  ('smp-c977356bdb43d99edd22', 'smp-903e6a1acd2c67f92e00', 12759, 12759, '2026-02-15 12:00:00', 'smp-3b164a9395896373810e', 'job-geo240011-12', 'res-cpu'),
  ('smp-af233ec0d4444a9edfa5', 'smp-903e6a1acd2c67f92e00', 12881, 12881, '2026-04-15 12:00:00', 'smp-3b164a9395896373810e', 'job-geo240011-14', 'res-cpu'),
  ('smp-b84ddb9faa6307a5a95e', 'smp-903e6a1acd2c67f92e00', 13003, 13003, '2026-06-15 12:00:00', 'smp-3b164a9395896373810e', 'job-geo240011-16', 'res-cpu'),
  ('smp-0e4ae5caa92f5e0d82a8', 'smp-903e6a1acd2c67f92e00', 12896, 12896, '2026-02-15 12:00:00', 'smp-f908bd32e57fd5c2896d', 'job-geo240011-22', 'res-cpu'),
  ('smp-e04a55b89c5261a53754', 'smp-903e6a1acd2c67f92e00', 13018, 13018, '2026-04-15 12:00:00', 'smp-f908bd32e57fd5c2896d', 'job-geo240011-24', 'res-cpu'),
  ('smp-d9a25ca35552f21d7be9', 'smp-903e6a1acd2c67f92e00', 13140, 13140, '2026-06-15 12:00:00', 'smp-f908bd32e57fd5c2896d', 'job-geo240011-26', 'res-cpu');

-- CIS240068: Distributed Training of Sparse Mixture Models
INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES
  ('smp-46566060f3d3d2eb9420', 'CIS240068', 'Distributed Training of Sparse Mixture Models', 'access', 'smp-26420ea0545ca72a32b3', 'ACTIVE');
INSERT IGNORE INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smp-f6d592517cf3b690a68b', 'smp-46566060f3d3d2eb9420', 'CIS240068 startup allocation', 'ACTIVE', '00000000-0000-0000-0000-000000000001', 1200000, '2026-01-01 00:00:00', '2026-12-31 00:00:00');
INSERT IGNORE INTO project_memberships (project_id, user_id, role) VALUES
  ('smp-46566060f3d3d2eb9420', 'smp-26420ea0545ca72a32b3', 'PI'),
  ('smp-46566060f3d3d2eb9420', 'smp-343175c9b7007d1ff0f8', 'CO_PI'),
  ('smp-46566060f3d3d2eb9420', 'smp-02b1cd6cafbd27dd7201', 'ALLOCATION_MANAGER');
INSERT IGNORE INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smp-64d81cddbe82cffe8c4c', 'smp-f6d592517cf3b690a68b', 'smp-26420ea0545ca72a32b3', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-7bccab59588549eca7fe', 'smp-f6d592517cf3b690a68b', 'smp-343175c9b7007d1ff0f8', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-b692ed3def842f0502b3', 'smp-f6d592517cf3b690a68b', 'smp-626aede3f65aeb969c74', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'INACTIVE'),
  ('smp-2e22c2973dd09d97a3d2', 'smp-f6d592517cf3b690a68b', 'smp-f6343c23c6a7028f8ed7', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-7c7f4f97cb391e2c7bbc', 'smp-f6d592517cf3b690a68b', 'smp-318427dda738e0504620', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE');
INSERT IGNORE INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time) VALUES
  ('smp-2ee3008445252911c454', 'smp-f6d592517cf3b690a68b', 'res-cpu', 12000, 3600);
INSERT IGNORE INTO compute_allocation_usages (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id) VALUES
  ('smp-ed290c19f19b9711e6de', 'smp-f6d592517cf3b690a68b', 60122, 60122, '2026-02-15 12:00:00', 'smp-26420ea0545ca72a32b3', 'job-cis240068-02', 'res-cpu'),
  ('smp-b92b66eeb33cb669dc2d', 'smp-f6d592517cf3b690a68b', 60244, 60244, '2026-04-15 12:00:00', 'smp-26420ea0545ca72a32b3', 'job-cis240068-04', 'res-cpu'),
  ('smp-883764679b1ad666aaa2', 'smp-f6d592517cf3b690a68b', 60366, 60366, '2026-06-15 12:00:00', 'smp-26420ea0545ca72a32b3', 'job-cis240068-06', 'res-cpu'),
  ('smp-b41963c27d6a8873a912', 'smp-f6d592517cf3b690a68b', 60259, 60259, '2026-02-15 12:00:00', 'smp-626aede3f65aeb969c74', 'job-cis240068-12', 'res-cpu'),
  ('smp-e1d28528d1a7b2b1286c', 'smp-f6d592517cf3b690a68b', 60381, 60381, '2026-04-15 12:00:00', 'smp-626aede3f65aeb969c74', 'job-cis240068-14', 'res-cpu'),
  ('smp-d4e85aad8a2a10fbe185', 'smp-f6d592517cf3b690a68b', 60503, 60503, '2026-06-15 12:00:00', 'smp-626aede3f65aeb969c74', 'job-cis240068-16', 'res-cpu'),
  ('smp-30ed7b95adcdbcc8432c', 'smp-f6d592517cf3b690a68b', 60396, 60396, '2026-02-15 12:00:00', 'smp-f6343c23c6a7028f8ed7', 'job-cis240068-22', 'res-cpu'),
  ('smp-3bf95a6ff91f95a7bf45', 'smp-f6d592517cf3b690a68b', 60518, 60518, '2026-04-15 12:00:00', 'smp-f6343c23c6a7028f8ed7', 'job-cis240068-24', 'res-cpu'),
  ('smp-c47aec92d29b9aa71a2c', 'smp-f6d592517cf3b690a68b', 60640, 60640, '2026-06-15 12:00:00', 'smp-f6343c23c6a7028f8ed7', 'job-cis240068-26', 'res-cpu');
INSERT IGNORE INTO compute_allocation_change_requests (id, compute_allocation_id, requested_su_amount, requested_status, reason, change_status, requester_id, approver_id, timestamp) VALUES
  ('smp-a362284a709cb8db8365', 'smp-f6d592517cf3b690a68b', 600000, 'ACTIVE', 'Supplement for CIS240068 production campaign', 'REJECTED', 'smp-26420ea0545ca72a32b3', '', '2026-06-20 09:00:00');
INSERT IGNORE INTO compute_allocation_change_request_events (id, compute_allocation_change_request_id, event_type, description, timestamp) VALUES
  ('smp-a362284a709cb8db8365-e1', 'smp-a362284a709cb8db8365', 'CREATED', 'Change request submitted by PI.', '2026-06-20 09:00:00');

-- MCB240032: Molecular Dynamics of Intrinsically Disordered Proteins
INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES
  ('smp-2ef2fd65182345f961cf', 'MCB240032', 'Molecular Dynamics of Intrinsically Disordered Proteins', 'access', 'smp-3b164a9395896373810e', 'ACTIVE');
INSERT IGNORE INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smp-c9853aa7a5047707793f', 'smp-2ef2fd65182345f961cf', 'MCB240032 startup allocation', 'ACTIVE', '00000000-0000-0000-0000-000000000001', 100000, '2026-01-01 00:00:00', '2026-12-31 00:00:00');
INSERT IGNORE INTO project_memberships (project_id, user_id, role) VALUES
  ('smp-2ef2fd65182345f961cf', 'smp-3b164a9395896373810e', 'PI'),
  ('smp-2ef2fd65182345f961cf', 'smp-f908bd32e57fd5c2896d', 'CO_PI'),
  ('smp-2ef2fd65182345f961cf', 'smp-b8e82843b10d1a15b995', 'ALLOCATION_MANAGER');
INSERT IGNORE INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smp-9476c4f1e5bf7533ef75', 'smp-c9853aa7a5047707793f', 'smp-3b164a9395896373810e', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-6f19adb01050b92b28ad', 'smp-c9853aa7a5047707793f', 'smp-f908bd32e57fd5c2896d', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'INACTIVE'),
  ('smp-f6693a2b1c7120ffe29c', 'smp-c9853aa7a5047707793f', 'smp-4907af7ab30bd84e6deb', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-caa25cae214c2173de67', 'smp-c9853aa7a5047707793f', 'smp-fa5b4cdf77ce78bd27e0', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-07bffefcd0482952db55', 'smp-c9853aa7a5047707793f', 'smp-52fae2d4d9ff23760057', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE');
INSERT IGNORE INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time) VALUES
  ('smp-04145f1728c174284d80', 'smp-c9853aa7a5047707793f', 'res-cpu', 1000, 3600),
  ('smp-86cca597b1b4dfe7f75f', 'smp-c9853aa7a5047707793f', 'res-gpu', 50, 3600);
INSERT IGNORE INTO compute_allocation_usages (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id) VALUES
  ('smp-0545c6e6175b54feda98', 'smp-c9853aa7a5047707793f', 5122, 5122, '2026-02-15 12:00:00', 'smp-3b164a9395896373810e', 'job-mcb240032-02', 'res-cpu'),
  ('smp-da639b4db3786bbb4839', 'smp-c9853aa7a5047707793f', 5244, 5244, '2026-04-15 12:00:00', 'smp-3b164a9395896373810e', 'job-mcb240032-04', 'res-cpu'),
  ('smp-2ff12ffb0c267e76a30a', 'smp-c9853aa7a5047707793f', 5366, 5366, '2026-06-15 12:00:00', 'smp-3b164a9395896373810e', 'job-mcb240032-06', 'res-cpu'),
  ('smp-b6b494f6fea4c318d2d6', 'smp-c9853aa7a5047707793f', 5259, 5259, '2026-02-15 12:00:00', 'smp-4907af7ab30bd84e6deb', 'job-mcb240032-12', 'res-cpu'),
  ('smp-7db21ec71651cc4242c8', 'smp-c9853aa7a5047707793f', 5381, 5381, '2026-04-15 12:00:00', 'smp-4907af7ab30bd84e6deb', 'job-mcb240032-14', 'res-cpu'),
  ('smp-a494340a98c6f8a8d1f6', 'smp-c9853aa7a5047707793f', 5503, 5503, '2026-06-15 12:00:00', 'smp-4907af7ab30bd84e6deb', 'job-mcb240032-16', 'res-cpu'),
  ('smp-9660252e76ac549b0a47', 'smp-c9853aa7a5047707793f', 5396, 5396, '2026-02-15 12:00:00', 'smp-fa5b4cdf77ce78bd27e0', 'job-mcb240032-22', 'res-cpu'),
  ('smp-878859e4c47f6b6cc517', 'smp-c9853aa7a5047707793f', 5518, 5518, '2026-04-15 12:00:00', 'smp-fa5b4cdf77ce78bd27e0', 'job-mcb240032-24', 'res-cpu'),
  ('smp-552d8c8f929c5fd05719', 'smp-c9853aa7a5047707793f', 5640, 5640, '2026-06-15 12:00:00', 'smp-fa5b4cdf77ce78bd27e0', 'job-mcb240032-26', 'res-cpu');

-- ENG240046: Wall-Resolved LES of Turbine Film Cooling
INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES
  ('smp-0948a283a07e0b73d2f7', 'ENG240046', 'Wall-Resolved LES of Turbine Film Cooling', 'access', 'smp-626aede3f65aeb969c74', 'ACTIVE');
INSERT IGNORE INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smp-c4071741a17cbea31f46', 'smp-0948a283a07e0b73d2f7', 'ENG240046 startup allocation', 'ACTIVE', '00000000-0000-0000-0000-000000000001', 750000, '2026-01-01 00:00:00', '2026-12-31 00:00:00');
INSERT IGNORE INTO project_memberships (project_id, user_id, role) VALUES
  ('smp-0948a283a07e0b73d2f7', 'smp-626aede3f65aeb969c74', 'PI'),
  ('smp-0948a283a07e0b73d2f7', 'smp-f6343c23c6a7028f8ed7', 'CO_PI'),
  ('smp-0948a283a07e0b73d2f7', 'smp-318427dda738e0504620', 'ALLOCATION_MANAGER');
INSERT IGNORE INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smp-ceeab8b8d03eb169950a', 'smp-c4071741a17cbea31f46', 'smp-626aede3f65aeb969c74', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'INACTIVE'),
  ('smp-d3f4c3b41e0e54c34bed', 'smp-c4071741a17cbea31f46', 'smp-f6343c23c6a7028f8ed7', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-9d0325840ec68ace6a7d', 'smp-c4071741a17cbea31f46', 'smp-ae4c14e3cd98f58c2363', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-95c353026d3979df0010', 'smp-c4071741a17cbea31f46', 'smp-598aba3fd25a5b74a2cf', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-3f7b090718cb5d3bffda', 'smp-c4071741a17cbea31f46', 'smp-8d36f8ec62cc1ab62fd5', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE');
INSERT IGNORE INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time) VALUES
  ('smp-290e16e3bc2f7ef0e0ff', 'smp-c4071741a17cbea31f46', 'res-cpu', 7500, 3600);
INSERT IGNORE INTO compute_allocation_usages (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id) VALUES
  ('smp-1b8f689a0cffe1d52ca8', 'smp-c4071741a17cbea31f46', 37622, 37622, '2026-02-15 12:00:00', 'smp-626aede3f65aeb969c74', 'job-eng240046-02', 'res-cpu'),
  ('smp-9ed2a39afce2ac00f3e5', 'smp-c4071741a17cbea31f46', 37744, 37744, '2026-04-15 12:00:00', 'smp-626aede3f65aeb969c74', 'job-eng240046-04', 'res-cpu'),
  ('smp-cfecf6fd8b56831fabed', 'smp-c4071741a17cbea31f46', 37866, 37866, '2026-06-15 12:00:00', 'smp-626aede3f65aeb969c74', 'job-eng240046-06', 'res-cpu'),
  ('smp-8ce1615d93d0f31f13fa', 'smp-c4071741a17cbea31f46', 37759, 37759, '2026-02-15 12:00:00', 'smp-ae4c14e3cd98f58c2363', 'job-eng240046-12', 'res-cpu'),
  ('smp-ab60b36df0db3dc6c392', 'smp-c4071741a17cbea31f46', 37881, 37881, '2026-04-15 12:00:00', 'smp-ae4c14e3cd98f58c2363', 'job-eng240046-14', 'res-cpu'),
  ('smp-e9c9bfbc8282a2058866', 'smp-c4071741a17cbea31f46', 38003, 38003, '2026-06-15 12:00:00', 'smp-ae4c14e3cd98f58c2363', 'job-eng240046-16', 'res-cpu'),
  ('smp-4177f7b4cfc4945eafe2', 'smp-c4071741a17cbea31f46', 37896, 37896, '2026-02-15 12:00:00', 'smp-598aba3fd25a5b74a2cf', 'job-eng240046-22', 'res-cpu'),
  ('smp-aa9d56523363c1d988ea', 'smp-c4071741a17cbea31f46', 38018, 38018, '2026-04-15 12:00:00', 'smp-598aba3fd25a5b74a2cf', 'job-eng240046-24', 'res-cpu'),
  ('smp-91b516fb1a3cd25627bb', 'smp-c4071741a17cbea31f46', 38140, 38140, '2026-06-15 12:00:00', 'smp-598aba3fd25a5b74a2cf', 'job-eng240046-26', 'res-cpu');

-- OCE240015: Mesoscale Eddy Tracking in Global Ocean Reanalysis
INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES
  ('smp-c642dcb393db75aa5526', 'OCE240015', 'Mesoscale Eddy Tracking in Global Ocean Reanalysis', 'access', 'smp-4907af7ab30bd84e6deb', 'ACTIVE');
INSERT IGNORE INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smp-de5487db358f6fb7d6fe', 'smp-c642dcb393db75aa5526', 'OCE240015 startup allocation', 'ACTIVE', '00000000-0000-0000-0000-000000000001', 300000, '2026-01-01 00:00:00', '2026-12-31 00:00:00');
INSERT IGNORE INTO project_memberships (project_id, user_id, role) VALUES
  ('smp-c642dcb393db75aa5526', 'smp-4907af7ab30bd84e6deb', 'PI'),
  ('smp-c642dcb393db75aa5526', 'smp-fa5b4cdf77ce78bd27e0', 'CO_PI'),
  ('smp-c642dcb393db75aa5526', 'smp-52fae2d4d9ff23760057', 'ALLOCATION_MANAGER');
INSERT IGNORE INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smp-adc667d00a365ef25633', 'smp-de5487db358f6fb7d6fe', 'smp-4907af7ab30bd84e6deb', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-48470b978fb2146882a6', 'smp-de5487db358f6fb7d6fe', 'smp-fa5b4cdf77ce78bd27e0', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-036435c793e419eb68e5', 'smp-de5487db358f6fb7d6fe', 'smp-343175c9b7007d1ff0f8', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-cab86841b6858594eb99', 'smp-de5487db358f6fb7d6fe', 'smp-35b28b0dfed3e7dfcd36', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-6b247b1e68a043d0386b', 'smp-de5487db358f6fb7d6fe', 'smp-344b252c9dd481944db0', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'INACTIVE');
INSERT IGNORE INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time) VALUES
  ('smp-352960efaeab23d059b5', 'smp-de5487db358f6fb7d6fe', 'res-cpu', 3000, 3600);
INSERT IGNORE INTO compute_allocation_usages (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id) VALUES
  ('smp-5f737417545f4ef958b9', 'smp-de5487db358f6fb7d6fe', 15122, 15122, '2026-02-15 12:00:00', 'smp-4907af7ab30bd84e6deb', 'job-oce240015-02', 'res-cpu'),
  ('smp-47166fc6b0dd678f1bd5', 'smp-de5487db358f6fb7d6fe', 15244, 15244, '2026-04-15 12:00:00', 'smp-4907af7ab30bd84e6deb', 'job-oce240015-04', 'res-cpu'),
  ('smp-23f3f687943779bef143', 'smp-de5487db358f6fb7d6fe', 15366, 15366, '2026-06-15 12:00:00', 'smp-4907af7ab30bd84e6deb', 'job-oce240015-06', 'res-cpu'),
  ('smp-764478c5e1a63403d465', 'smp-de5487db358f6fb7d6fe', 15259, 15259, '2026-02-15 12:00:00', 'smp-343175c9b7007d1ff0f8', 'job-oce240015-12', 'res-cpu'),
  ('smp-53be17fc6ec1b4bd2f26', 'smp-de5487db358f6fb7d6fe', 15381, 15381, '2026-04-15 12:00:00', 'smp-343175c9b7007d1ff0f8', 'job-oce240015-14', 'res-cpu'),
  ('smp-6baa75f9201df61b025f', 'smp-de5487db358f6fb7d6fe', 15503, 15503, '2026-06-15 12:00:00', 'smp-343175c9b7007d1ff0f8', 'job-oce240015-16', 'res-cpu'),
  ('smp-ba1a17e1891bf497e4a3', 'smp-de5487db358f6fb7d6fe', 15396, 15396, '2026-02-15 12:00:00', 'smp-35b28b0dfed3e7dfcd36', 'job-oce240015-22', 'res-cpu'),
  ('smp-fb89291e07eabaf74b00', 'smp-de5487db358f6fb7d6fe', 15518, 15518, '2026-04-15 12:00:00', 'smp-35b28b0dfed3e7dfcd36', 'job-oce240015-24', 'res-cpu'),
  ('smp-a107d2c5d41c9a8015c0', 'smp-de5487db358f6fb7d6fe', 15640, 15640, '2026-06-15 12:00:00', 'smp-35b28b0dfed3e7dfcd36', 'job-oce240015-26', 'res-cpu');

-- DMS240027: High-Order Solvers for Kinetic Plasma Equations
INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES
  ('smp-22445a61084f5764042b', 'DMS240027', 'High-Order Solvers for Kinetic Plasma Equations', 'access', 'smp-ae4c14e3cd98f58c2363', 'ACTIVE');
INSERT IGNORE INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smp-ce03e036de1da6863393', 'smp-22445a61084f5764042b', 'DMS240027 startup allocation', 'INACTIVE', '00000000-0000-0000-0000-000000000001', 500000, '2026-01-01 00:00:00', '2026-12-31 00:00:00');
INSERT IGNORE INTO project_memberships (project_id, user_id, role) VALUES
  ('smp-22445a61084f5764042b', 'smp-ae4c14e3cd98f58c2363', 'PI'),
  ('smp-22445a61084f5764042b', 'smp-598aba3fd25a5b74a2cf', 'CO_PI'),
  ('smp-22445a61084f5764042b', 'smp-8d36f8ec62cc1ab62fd5', 'ALLOCATION_MANAGER');
INSERT IGNORE INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smp-5f8a62442cc1be68f10d', 'smp-ce03e036de1da6863393', 'smp-ae4c14e3cd98f58c2363', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-7d6d016ed837e198c526', 'smp-ce03e036de1da6863393', 'smp-598aba3fd25a5b74a2cf', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-43c41d94073b21d4db5f', 'smp-ce03e036de1da6863393', 'smp-f908bd32e57fd5c2896d', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-f5fa7ced0d73482111ba', 'smp-ce03e036de1da6863393', 'smp-02b1cd6cafbd27dd7201', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'INACTIVE'),
  ('smp-2b3ebce9395cb23a040e', 'smp-ce03e036de1da6863393', 'smp-c8e6b563ac8f348d6069', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE');
INSERT IGNORE INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time) VALUES
  ('smp-01b687d9274b62c9f6b5', 'smp-ce03e036de1da6863393', 'res-cpu', 5000, 3600),
  ('smp-f9c7a3aa26187e49d26e', 'smp-ce03e036de1da6863393', 'res-gpu', 250, 3600);
INSERT IGNORE INTO compute_allocation_usages (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id) VALUES
  ('smp-dba38111b0c14e71fd7f', 'smp-ce03e036de1da6863393', 25122, 25122, '2026-02-15 12:00:00', 'smp-ae4c14e3cd98f58c2363', 'job-dms240027-02', 'res-cpu'),
  ('smp-2a90ead06577735119e1', 'smp-ce03e036de1da6863393', 25244, 25244, '2026-04-15 12:00:00', 'smp-ae4c14e3cd98f58c2363', 'job-dms240027-04', 'res-cpu'),
  ('smp-7072c544a0eec959aef0', 'smp-ce03e036de1da6863393', 25366, 25366, '2026-06-15 12:00:00', 'smp-ae4c14e3cd98f58c2363', 'job-dms240027-06', 'res-cpu'),
  ('smp-a414051ec8de4520b9b6', 'smp-ce03e036de1da6863393', 25259, 25259, '2026-02-15 12:00:00', 'smp-f908bd32e57fd5c2896d', 'job-dms240027-12', 'res-cpu'),
  ('smp-b614b9f0d21a56d80103', 'smp-ce03e036de1da6863393', 25381, 25381, '2026-04-15 12:00:00', 'smp-f908bd32e57fd5c2896d', 'job-dms240027-14', 'res-cpu'),
  ('smp-c4278139fad25c9719e3', 'smp-ce03e036de1da6863393', 25503, 25503, '2026-06-15 12:00:00', 'smp-f908bd32e57fd5c2896d', 'job-dms240027-16', 'res-cpu'),
  ('smp-20b22d60cdb397c05c42', 'smp-ce03e036de1da6863393', 25396, 25396, '2026-02-15 12:00:00', 'smp-02b1cd6cafbd27dd7201', 'job-dms240027-22', 'res-cpu'),
  ('smp-8147ad5d8acf9a477192', 'smp-ce03e036de1da6863393', 25518, 25518, '2026-04-15 12:00:00', 'smp-02b1cd6cafbd27dd7201', 'job-dms240027-24', 'res-cpu'),
  ('smp-c462c852980fe4eaaef5', 'smp-ce03e036de1da6863393', 25640, 25640, '2026-06-15 12:00:00', 'smp-02b1cd6cafbd27dd7201', 'job-dms240027-26', 'res-cpu');
INSERT IGNORE INTO compute_allocation_change_requests (id, compute_allocation_id, requested_su_amount, requested_status, reason, change_status, requester_id, approver_id, timestamp) VALUES
  ('smp-cba9d22751366aa10b94', 'smp-ce03e036de1da6863393', 250000, 'ACTIVE', 'Supplement for DMS240027 production campaign', 'PENDING', 'smp-ae4c14e3cd98f58c2363', '', '2026-06-24 09:00:00');
INSERT IGNORE INTO compute_allocation_change_request_events (id, compute_allocation_change_request_id, event_type, description, timestamp) VALUES
  ('smp-cba9d22751366aa10b94-e1', 'smp-cba9d22751366aa10b94', 'CREATED', 'Change request submitted by PI.', '2026-06-24 09:00:00');

-- EAR240038: Mantle Convection with Realistic Rheology
INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES
  ('smp-6f71b0a457dc2d9383c7', 'EAR240038', 'Mantle Convection with Realistic Rheology', 'access', 'smp-343175c9b7007d1ff0f8', 'ACTIVE');
INSERT IGNORE INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smp-cac362e9bc8a5b73dc56', 'smp-6f71b0a457dc2d9383c7', 'EAR240038 startup allocation', 'ACTIVE', '00000000-0000-0000-0000-000000000001', 250000, '2026-01-01 00:00:00', '2026-12-31 00:00:00');
INSERT IGNORE INTO project_memberships (project_id, user_id, role) VALUES
  ('smp-6f71b0a457dc2d9383c7', 'smp-343175c9b7007d1ff0f8', 'PI'),
  ('smp-6f71b0a457dc2d9383c7', 'smp-35b28b0dfed3e7dfcd36', 'CO_PI'),
  ('smp-6f71b0a457dc2d9383c7', 'smp-344b252c9dd481944db0', 'ALLOCATION_MANAGER');
INSERT IGNORE INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smp-7433b8a0532af1994f82', 'smp-cac362e9bc8a5b73dc56', 'smp-343175c9b7007d1ff0f8', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-8ab28e66c8069430b34a', 'smp-cac362e9bc8a5b73dc56', 'smp-35b28b0dfed3e7dfcd36', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-7ed90896509e0411de77', 'smp-cac362e9bc8a5b73dc56', 'smp-f6343c23c6a7028f8ed7', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'INACTIVE'),
  ('smp-ee268f40f73c25a878df', 'smp-cac362e9bc8a5b73dc56', 'smp-b8e82843b10d1a15b995', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-5e3817eadd4645ed02d4', 'smp-cac362e9bc8a5b73dc56', 'smp-2fd62ff24cea6be8de4c', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE');
INSERT IGNORE INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time) VALUES
  ('smp-81f50c2c9a64f4e65068', 'smp-cac362e9bc8a5b73dc56', 'res-cpu', 2500, 3600);
INSERT IGNORE INTO compute_allocation_usages (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id) VALUES
  ('smp-c88c10bdb0c05befa0f2', 'smp-cac362e9bc8a5b73dc56', 12622, 12622, '2026-02-15 12:00:00', 'smp-343175c9b7007d1ff0f8', 'job-ear240038-02', 'res-cpu'),
  ('smp-324ff31c36ac2b276a59', 'smp-cac362e9bc8a5b73dc56', 12744, 12744, '2026-04-15 12:00:00', 'smp-343175c9b7007d1ff0f8', 'job-ear240038-04', 'res-cpu'),
  ('smp-dc60c880beda5edcc869', 'smp-cac362e9bc8a5b73dc56', 12866, 12866, '2026-06-15 12:00:00', 'smp-343175c9b7007d1ff0f8', 'job-ear240038-06', 'res-cpu'),
  ('smp-ac4f17b1f4d9a2653164', 'smp-cac362e9bc8a5b73dc56', 12759, 12759, '2026-02-15 12:00:00', 'smp-f6343c23c6a7028f8ed7', 'job-ear240038-12', 'res-cpu'),
  ('smp-18f43de9e2b2ce99dc97', 'smp-cac362e9bc8a5b73dc56', 12881, 12881, '2026-04-15 12:00:00', 'smp-f6343c23c6a7028f8ed7', 'job-ear240038-14', 'res-cpu'),
  ('smp-e4ea2a77728d3faa087a', 'smp-cac362e9bc8a5b73dc56', 13003, 13003, '2026-06-15 12:00:00', 'smp-f6343c23c6a7028f8ed7', 'job-ear240038-16', 'res-cpu'),
  ('smp-a1cd91aae6e29ba33065', 'smp-cac362e9bc8a5b73dc56', 12896, 12896, '2026-02-15 12:00:00', 'smp-b8e82843b10d1a15b995', 'job-ear240038-22', 'res-cpu'),
  ('smp-35a3fb4f232de26e49a6', 'smp-cac362e9bc8a5b73dc56', 13018, 13018, '2026-04-15 12:00:00', 'smp-b8e82843b10d1a15b995', 'job-ear240038-24', 'res-cpu'),
  ('smp-ad766f924c56521625d2', 'smp-cac362e9bc8a5b73dc56', 13140, 13140, '2026-06-15 12:00:00', 'smp-b8e82843b10d1a15b995', 'job-ear240038-26', 'res-cpu');

-- NEU240052: Whole-Brain Spiking Network Simulation
INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES
  ('smp-f5e84bbacec87c3deb6d', 'NEU240052', 'Whole-Brain Spiking Network Simulation', 'access', 'smp-f908bd32e57fd5c2896d', 'INACTIVE');
INSERT IGNORE INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smp-b1145214bc21940084fa', 'smp-f5e84bbacec87c3deb6d', 'NEU240052 startup allocation', 'ACTIVE', '00000000-0000-0000-0000-000000000001', 1200000, '2026-01-01 00:00:00', '2026-12-31 00:00:00');
INSERT IGNORE INTO project_memberships (project_id, user_id, role) VALUES
  ('smp-f5e84bbacec87c3deb6d', 'smp-f908bd32e57fd5c2896d', 'PI'),
  ('smp-f5e84bbacec87c3deb6d', 'smp-02b1cd6cafbd27dd7201', 'CO_PI'),
  ('smp-f5e84bbacec87c3deb6d', 'smp-c8e6b563ac8f348d6069', 'ALLOCATION_MANAGER');
INSERT IGNORE INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smp-ef7b3cf471a18f724bfe', 'smp-b1145214bc21940084fa', 'smp-f908bd32e57fd5c2896d', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-dd291d9acb1a3cf3a34d', 'smp-b1145214bc21940084fa', 'smp-02b1cd6cafbd27dd7201', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'INACTIVE'),
  ('smp-8172646ef19cf2aba13d', 'smp-b1145214bc21940084fa', 'smp-fa5b4cdf77ce78bd27e0', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-62e2175da6c8cd7bc527', 'smp-b1145214bc21940084fa', 'smp-318427dda738e0504620', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-5ee1e3afbdb762fc82b9', 'smp-b1145214bc21940084fa', 'smp-bc08d052551918510023', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE');
INSERT IGNORE INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time) VALUES
  ('smp-272f3c504c041203dbb6', 'smp-b1145214bc21940084fa', 'res-cpu', 12000, 3600);
INSERT IGNORE INTO compute_allocation_usages (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id) VALUES
  ('smp-bbce515125ea9a07b292', 'smp-b1145214bc21940084fa', 60122, 60122, '2026-02-15 12:00:00', 'smp-f908bd32e57fd5c2896d', 'job-neu240052-02', 'res-cpu'),
  ('smp-22f88a7c4c6b3d6fe05d', 'smp-b1145214bc21940084fa', 60244, 60244, '2026-04-15 12:00:00', 'smp-f908bd32e57fd5c2896d', 'job-neu240052-04', 'res-cpu'),
  ('smp-01fffd2e411603a60bbd', 'smp-b1145214bc21940084fa', 60366, 60366, '2026-06-15 12:00:00', 'smp-f908bd32e57fd5c2896d', 'job-neu240052-06', 'res-cpu'),
  ('smp-fe4ae389920eee9d080b', 'smp-b1145214bc21940084fa', 60259, 60259, '2026-02-15 12:00:00', 'smp-fa5b4cdf77ce78bd27e0', 'job-neu240052-12', 'res-cpu'),
  ('smp-c9efd8c32a16059fb7e6', 'smp-b1145214bc21940084fa', 60381, 60381, '2026-04-15 12:00:00', 'smp-fa5b4cdf77ce78bd27e0', 'job-neu240052-14', 'res-cpu'),
  ('smp-99f3926b6c00e9c5d335', 'smp-b1145214bc21940084fa', 60503, 60503, '2026-06-15 12:00:00', 'smp-fa5b4cdf77ce78bd27e0', 'job-neu240052-16', 'res-cpu'),
  ('smp-8b8ef509006b1f3c9baa', 'smp-b1145214bc21940084fa', 60396, 60396, '2026-02-15 12:00:00', 'smp-318427dda738e0504620', 'job-neu240052-22', 'res-cpu'),
  ('smp-ec7a532ca3a2d62e3556', 'smp-b1145214bc21940084fa', 60518, 60518, '2026-04-15 12:00:00', 'smp-318427dda738e0504620', 'job-neu240052-24', 'res-cpu'),
  ('smp-7bb7b64c3a488df7b353', 'smp-b1145214bc21940084fa', 60640, 60640, '2026-06-15 12:00:00', 'smp-318427dda738e0504620', 'job-neu240052-26', 'res-cpu');

-- ECO240009: Agent-Based Modeling of Land-Use Change
INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES
  ('smp-e483b73b4458eae0fa14', 'ECO240009', 'Agent-Based Modeling of Land-Use Change', 'access', 'smp-f6343c23c6a7028f8ed7', 'ACTIVE');
INSERT IGNORE INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smp-7512b54990a4efa50cf6', 'smp-e483b73b4458eae0fa14', 'ECO240009 startup allocation', 'ACTIVE', '00000000-0000-0000-0000-000000000001', 100000, '2026-01-01 00:00:00', '2026-12-31 00:00:00');
INSERT IGNORE INTO project_memberships (project_id, user_id, role) VALUES
  ('smp-e483b73b4458eae0fa14', 'smp-f6343c23c6a7028f8ed7', 'PI'),
  ('smp-e483b73b4458eae0fa14', 'smp-b8e82843b10d1a15b995', 'CO_PI'),
  ('smp-e483b73b4458eae0fa14', 'smp-2fd62ff24cea6be8de4c', 'ALLOCATION_MANAGER');
INSERT IGNORE INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smp-9d0742abbe4931843e57', 'smp-7512b54990a4efa50cf6', 'smp-f6343c23c6a7028f8ed7', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'INACTIVE'),
  ('smp-849058dc7fcecbf7b6b5', 'smp-7512b54990a4efa50cf6', 'smp-b8e82843b10d1a15b995', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-a934e4f3f489392b9a05', 'smp-7512b54990a4efa50cf6', 'smp-598aba3fd25a5b74a2cf', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-432ed414e8e53550f302', 'smp-7512b54990a4efa50cf6', 'smp-52fae2d4d9ff23760057', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-5cbd30c58d82e19d150a', 'smp-7512b54990a4efa50cf6', 'smp-b800fa9f20c414f38e44', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE');
INSERT IGNORE INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time) VALUES
  ('smp-5a5aec09c83d141ffac9', 'smp-7512b54990a4efa50cf6', 'res-cpu', 1000, 3600),
  ('smp-a96c3041058a5feb48be', 'smp-7512b54990a4efa50cf6', 'res-gpu', 50, 3600);
INSERT IGNORE INTO compute_allocation_usages (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id) VALUES
  ('smp-03881086768fe03766f4', 'smp-7512b54990a4efa50cf6', 5122, 5122, '2026-02-15 12:00:00', 'smp-f6343c23c6a7028f8ed7', 'job-eco240009-02', 'res-cpu'),
  ('smp-6a8c06b1577be4f478a6', 'smp-7512b54990a4efa50cf6', 5244, 5244, '2026-04-15 12:00:00', 'smp-f6343c23c6a7028f8ed7', 'job-eco240009-04', 'res-cpu'),
  ('smp-a47aeaa1650e33338ea2', 'smp-7512b54990a4efa50cf6', 5366, 5366, '2026-06-15 12:00:00', 'smp-f6343c23c6a7028f8ed7', 'job-eco240009-06', 'res-cpu'),
  ('smp-3bb282631adbeef4f91c', 'smp-7512b54990a4efa50cf6', 5259, 5259, '2026-02-15 12:00:00', 'smp-598aba3fd25a5b74a2cf', 'job-eco240009-12', 'res-cpu'),
  ('smp-0b7d4354f5fa6d91b12a', 'smp-7512b54990a4efa50cf6', 5381, 5381, '2026-04-15 12:00:00', 'smp-598aba3fd25a5b74a2cf', 'job-eco240009-14', 'res-cpu'),
  ('smp-2fd5abb579af4df3c888', 'smp-7512b54990a4efa50cf6', 5503, 5503, '2026-06-15 12:00:00', 'smp-598aba3fd25a5b74a2cf', 'job-eco240009-16', 'res-cpu'),
  ('smp-cb736460f484f94be590', 'smp-7512b54990a4efa50cf6', 5396, 5396, '2026-02-15 12:00:00', 'smp-52fae2d4d9ff23760057', 'job-eco240009-22', 'res-cpu'),
  ('smp-1fcffefc2bf002f5599c', 'smp-7512b54990a4efa50cf6', 5518, 5518, '2026-04-15 12:00:00', 'smp-52fae2d4d9ff23760057', 'job-eco240009-24', 'res-cpu'),
  ('smp-72ae86ab1c594bf42f50', 'smp-7512b54990a4efa50cf6', 5640, 5640, '2026-06-15 12:00:00', 'smp-52fae2d4d9ff23760057', 'job-eco240009-26', 'res-cpu');

-- MED240061: Virtual Screening Against Antibiotic Resistance Targets
INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES
  ('smp-980231459f6b23c93e99', 'MED240061', 'Virtual Screening Against Antibiotic Resistance Targets', 'access', 'smp-fa5b4cdf77ce78bd27e0', 'ACTIVE');
INSERT IGNORE INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smp-bdb96ba1e6a5d37513db', 'smp-980231459f6b23c93e99', 'MED240061 startup allocation', 'ACTIVE', '00000000-0000-0000-0000-000000000001', 750000, '2026-01-01 00:00:00', '2026-12-31 00:00:00');
INSERT IGNORE INTO project_memberships (project_id, user_id, role) VALUES
  ('smp-980231459f6b23c93e99', 'smp-fa5b4cdf77ce78bd27e0', 'PI'),
  ('smp-980231459f6b23c93e99', 'smp-318427dda738e0504620', 'CO_PI'),
  ('smp-980231459f6b23c93e99', 'smp-bc08d052551918510023', 'ALLOCATION_MANAGER');
INSERT IGNORE INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smp-7ffab5fb09b6212c8741', 'smp-bdb96ba1e6a5d37513db', 'smp-fa5b4cdf77ce78bd27e0', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-99a0dad83547931b15ad', 'smp-bdb96ba1e6a5d37513db', 'smp-318427dda738e0504620', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-fd315fada5b30c7a1b15', 'smp-bdb96ba1e6a5d37513db', 'smp-35b28b0dfed3e7dfcd36', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-b485267a18604bb920b4', 'smp-bdb96ba1e6a5d37513db', 'smp-8d36f8ec62cc1ab62fd5', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-c28380a7fcc0a214c012', 'smp-bdb96ba1e6a5d37513db', 'smp-3cd6d104b01ef129b379', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'INACTIVE');
INSERT IGNORE INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time) VALUES
  ('smp-7d2a9327eed6552b9597', 'smp-bdb96ba1e6a5d37513db', 'res-cpu', 7500, 3600);
INSERT IGNORE INTO compute_allocation_usages (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id) VALUES
  ('smp-fa6d5c9f7535e88b9fe3', 'smp-bdb96ba1e6a5d37513db', 37622, 37622, '2026-02-15 12:00:00', 'smp-fa5b4cdf77ce78bd27e0', 'job-med240061-02', 'res-cpu'),
  ('smp-7c000be8f80e3ca352db', 'smp-bdb96ba1e6a5d37513db', 37744, 37744, '2026-04-15 12:00:00', 'smp-fa5b4cdf77ce78bd27e0', 'job-med240061-04', 'res-cpu'),
  ('smp-fb4aaee12a343fe01586', 'smp-bdb96ba1e6a5d37513db', 37866, 37866, '2026-06-15 12:00:00', 'smp-fa5b4cdf77ce78bd27e0', 'job-med240061-06', 'res-cpu'),
  ('smp-32193d634be270c82d67', 'smp-bdb96ba1e6a5d37513db', 37759, 37759, '2026-02-15 12:00:00', 'smp-35b28b0dfed3e7dfcd36', 'job-med240061-12', 'res-cpu'),
  ('smp-31a909ee65bae384a14e', 'smp-bdb96ba1e6a5d37513db', 37881, 37881, '2026-04-15 12:00:00', 'smp-35b28b0dfed3e7dfcd36', 'job-med240061-14', 'res-cpu'),
  ('smp-13e67e5444e2ec360cbd', 'smp-bdb96ba1e6a5d37513db', 38003, 38003, '2026-06-15 12:00:00', 'smp-35b28b0dfed3e7dfcd36', 'job-med240061-16', 'res-cpu'),
  ('smp-c7929bce2db2faeecc55', 'smp-bdb96ba1e6a5d37513db', 37896, 37896, '2026-02-15 12:00:00', 'smp-8d36f8ec62cc1ab62fd5', 'job-med240061-22', 'res-cpu'),
  ('smp-b03fe4d7ef1395d89bde', 'smp-bdb96ba1e6a5d37513db', 38018, 38018, '2026-04-15 12:00:00', 'smp-8d36f8ec62cc1ab62fd5', 'job-med240061-24', 'res-cpu'),
  ('smp-ecb89e0a260d7720850c', 'smp-bdb96ba1e6a5d37513db', 38140, 38140, '2026-06-15 12:00:00', 'smp-8d36f8ec62cc1ab62fd5', 'job-med240061-26', 'res-cpu');
INSERT IGNORE INTO compute_allocation_change_requests (id, compute_allocation_id, requested_su_amount, requested_status, reason, change_status, requester_id, approver_id, timestamp) VALUES
  ('smp-6ce228100e9a0587de58', 'smp-bdb96ba1e6a5d37513db', 375000, 'ACTIVE', 'Supplement for MED240061 production campaign', 'PENDING', 'smp-fa5b4cdf77ce78bd27e0', '', '2026-06-20 09:00:00');
INSERT IGNORE INTO compute_allocation_change_request_events (id, compute_allocation_change_request_id, event_type, description, timestamp) VALUES
  ('smp-6ce228100e9a0587de58-e1', 'smp-6ce228100e9a0587de58', 'CREATED', 'Change request submitted by PI.', '2026-06-20 09:00:00');

-- AER240044: Hypersonic Boundary-Layer Transition Prediction
INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES
  ('smp-ac4de8764ba7aa98e886', 'AER240044', 'Hypersonic Boundary-Layer Transition Prediction', 'access', 'smp-598aba3fd25a5b74a2cf', 'ACTIVE');
INSERT IGNORE INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smp-fd5e6684bf9182850ebd', 'smp-ac4de8764ba7aa98e886', 'AER240044 startup allocation', 'ACTIVE', '00000000-0000-0000-0000-000000000001', 300000, '2026-01-01 00:00:00', '2026-12-31 00:00:00');
INSERT IGNORE INTO project_memberships (project_id, user_id, role) VALUES
  ('smp-ac4de8764ba7aa98e886', 'smp-598aba3fd25a5b74a2cf', 'PI'),
  ('smp-ac4de8764ba7aa98e886', 'smp-52fae2d4d9ff23760057', 'CO_PI'),
  ('smp-ac4de8764ba7aa98e886', 'smp-b800fa9f20c414f38e44', 'ALLOCATION_MANAGER');
INSERT IGNORE INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smp-9ac2e29e9bca669bedac', 'smp-fd5e6684bf9182850ebd', 'smp-598aba3fd25a5b74a2cf', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-276a2a13e89431d65dd1', 'smp-fd5e6684bf9182850ebd', 'smp-52fae2d4d9ff23760057', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-9f80e1cb60b7fed5519d', 'smp-fd5e6684bf9182850ebd', 'smp-02b1cd6cafbd27dd7201', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-23ee3cd03d7dd5895364', 'smp-fd5e6684bf9182850ebd', 'smp-344b252c9dd481944db0', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'INACTIVE'),
  ('smp-5f60915d151ee7c64632', 'smp-fd5e6684bf9182850ebd', 'smp-90f9c3a872e74ceb9b54', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE');
INSERT IGNORE INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time) VALUES
  ('smp-785ef1f0e1c1269098dd', 'smp-fd5e6684bf9182850ebd', 'res-cpu', 3000, 3600);
INSERT IGNORE INTO compute_allocation_usages (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id) VALUES
  ('smp-2840160fb65ec4f4d7ce', 'smp-fd5e6684bf9182850ebd', 15122, 15122, '2026-02-15 12:00:00', 'smp-598aba3fd25a5b74a2cf', 'job-aer240044-02', 'res-cpu'),
  ('smp-894388502b8d4c479e50', 'smp-fd5e6684bf9182850ebd', 15244, 15244, '2026-04-15 12:00:00', 'smp-598aba3fd25a5b74a2cf', 'job-aer240044-04', 'res-cpu'),
  ('smp-fec37bbc32f5d233a7f6', 'smp-fd5e6684bf9182850ebd', 15366, 15366, '2026-06-15 12:00:00', 'smp-598aba3fd25a5b74a2cf', 'job-aer240044-06', 'res-cpu'),
  ('smp-02f68a751de4c2ce6483', 'smp-fd5e6684bf9182850ebd', 15259, 15259, '2026-02-15 12:00:00', 'smp-02b1cd6cafbd27dd7201', 'job-aer240044-12', 'res-cpu'),
  ('smp-775794c0f0e4e295f8d1', 'smp-fd5e6684bf9182850ebd', 15381, 15381, '2026-04-15 12:00:00', 'smp-02b1cd6cafbd27dd7201', 'job-aer240044-14', 'res-cpu'),
  ('smp-bbef762e52a9d5327f23', 'smp-fd5e6684bf9182850ebd', 15503, 15503, '2026-06-15 12:00:00', 'smp-02b1cd6cafbd27dd7201', 'job-aer240044-16', 'res-cpu'),
  ('smp-0cdc260bcd08b694ecef', 'smp-fd5e6684bf9182850ebd', 15396, 15396, '2026-02-15 12:00:00', 'smp-344b252c9dd481944db0', 'job-aer240044-22', 'res-cpu'),
  ('smp-594643d1848f1c5317a0', 'smp-fd5e6684bf9182850ebd', 15518, 15518, '2026-04-15 12:00:00', 'smp-344b252c9dd481944db0', 'job-aer240044-24', 'res-cpu'),
  ('smp-1b6c4033a2c532111c26', 'smp-fd5e6684bf9182850ebd', 15640, 15640, '2026-06-15 12:00:00', 'smp-344b252c9dd481944db0', 'job-aer240044-26', 'res-cpu');

-- QCN240017: Tensor-Network Simulation of Quantum Circuits
INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES
  ('smp-0b70f6fc04a8a1f03514', 'QCN240017', 'Tensor-Network Simulation of Quantum Circuits', 'access', 'smp-35b28b0dfed3e7dfcd36', 'ACTIVE');
INSERT IGNORE INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smp-81d26521f4dc52bb030b', 'smp-0b70f6fc04a8a1f03514', 'QCN240017 startup allocation', 'INACTIVE', '00000000-0000-0000-0000-000000000001', 500000, '2026-01-01 00:00:00', '2026-12-31 00:00:00');
INSERT IGNORE INTO project_memberships (project_id, user_id, role) VALUES
  ('smp-0b70f6fc04a8a1f03514', 'smp-35b28b0dfed3e7dfcd36', 'PI'),
  ('smp-0b70f6fc04a8a1f03514', 'smp-8d36f8ec62cc1ab62fd5', 'CO_PI'),
  ('smp-0b70f6fc04a8a1f03514', 'smp-3cd6d104b01ef129b379', 'ALLOCATION_MANAGER');
INSERT IGNORE INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smp-10a5eafb6445f979c71d', 'smp-81d26521f4dc52bb030b', 'smp-35b28b0dfed3e7dfcd36', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-2065f6b4a0c81045f9a6', 'smp-81d26521f4dc52bb030b', 'smp-8d36f8ec62cc1ab62fd5', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-77b9d327bcbd501ea39e', 'smp-81d26521f4dc52bb030b', 'smp-b8e82843b10d1a15b995', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'INACTIVE'),
  ('smp-9957caef1d18c43f12b3', 'smp-81d26521f4dc52bb030b', 'smp-c8e6b563ac8f348d6069', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-e7da29a49c0b1e06670a', 'smp-81d26521f4dc52bb030b', 'smp-38b35e0e4b22d0d6460a', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE');
INSERT IGNORE INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time) VALUES
  ('smp-f39914bda842d3029073', 'smp-81d26521f4dc52bb030b', 'res-cpu', 5000, 3600),
  ('smp-48fe1aeed1dee260a1cd', 'smp-81d26521f4dc52bb030b', 'res-gpu', 250, 3600);
INSERT IGNORE INTO compute_allocation_usages (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id) VALUES
  ('smp-8aba1e6527fb6978d840', 'smp-81d26521f4dc52bb030b', 25122, 25122, '2026-02-15 12:00:00', 'smp-35b28b0dfed3e7dfcd36', 'job-qcn240017-02', 'res-cpu'),
  ('smp-6a627f8a8045fc5d0914', 'smp-81d26521f4dc52bb030b', 25244, 25244, '2026-04-15 12:00:00', 'smp-35b28b0dfed3e7dfcd36', 'job-qcn240017-04', 'res-cpu'),
  ('smp-e011e4b777f76f1850da', 'smp-81d26521f4dc52bb030b', 25366, 25366, '2026-06-15 12:00:00', 'smp-35b28b0dfed3e7dfcd36', 'job-qcn240017-06', 'res-cpu'),
  ('smp-e16f0134dc556cd3a7c1', 'smp-81d26521f4dc52bb030b', 25259, 25259, '2026-02-15 12:00:00', 'smp-b8e82843b10d1a15b995', 'job-qcn240017-12', 'res-cpu'),
  ('smp-8bea4370c2aedb5041f0', 'smp-81d26521f4dc52bb030b', 25381, 25381, '2026-04-15 12:00:00', 'smp-b8e82843b10d1a15b995', 'job-qcn240017-14', 'res-cpu'),
  ('smp-6f5f0597c0fcc780cf39', 'smp-81d26521f4dc52bb030b', 25503, 25503, '2026-06-15 12:00:00', 'smp-b8e82843b10d1a15b995', 'job-qcn240017-16', 'res-cpu'),
  ('smp-5699a40460fb16a2423e', 'smp-81d26521f4dc52bb030b', 25396, 25396, '2026-02-15 12:00:00', 'smp-c8e6b563ac8f348d6069', 'job-qcn240017-22', 'res-cpu'),
  ('smp-9ef9920430d5a62d1381', 'smp-81d26521f4dc52bb030b', 25518, 25518, '2026-04-15 12:00:00', 'smp-c8e6b563ac8f348d6069', 'job-qcn240017-24', 'res-cpu'),
  ('smp-3a66778317d3640f3405', 'smp-81d26521f4dc52bb030b', 25640, 25640, '2026-06-15 12:00:00', 'smp-c8e6b563ac8f348d6069', 'job-qcn240017-26', 'res-cpu');

-- SES240029: Urban Microclimate Simulation for Heat Resilience
INSERT IGNORE INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES
  ('smp-30a2eabc73a65a860667', 'SES240029', 'Urban Microclimate Simulation for Heat Resilience', 'access', 'smp-02b1cd6cafbd27dd7201', 'ACTIVE');
INSERT IGNORE INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES
  ('smp-0384acb90162ba62c422', 'smp-30a2eabc73a65a860667', 'SES240029 startup allocation', 'ACTIVE', '00000000-0000-0000-0000-000000000001', 250000, '2026-01-01 00:00:00', '2026-12-31 00:00:00');
INSERT IGNORE INTO project_memberships (project_id, user_id, role) VALUES
  ('smp-30a2eabc73a65a860667', 'smp-02b1cd6cafbd27dd7201', 'PI'),
  ('smp-30a2eabc73a65a860667', 'smp-344b252c9dd481944db0', 'CO_PI'),
  ('smp-30a2eabc73a65a860667', 'smp-90f9c3a872e74ceb9b54', 'ALLOCATION_MANAGER');
INSERT IGNORE INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
  ('smp-08f7c147f068caa62854', 'smp-0384acb90162ba62c422', 'smp-02b1cd6cafbd27dd7201', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-acee8d6c86e4d40da1e0', 'smp-0384acb90162ba62c422', 'smp-344b252c9dd481944db0', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'INACTIVE'),
  ('smp-ee0408bdf7f2155c00ac', 'smp-0384acb90162ba62c422', 'smp-318427dda738e0504620', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-a5b622ae224e8db0569e', 'smp-0384acb90162ba62c422', 'smp-2fd62ff24cea6be8de4c', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
  ('smp-40a595e58ab783c22456', 'smp-0384acb90162ba62c422', 'smp-26420ea0545ca72a32b3', '2026-01-01 00:00:00', '2026-12-31 00:00:00', 'ACTIVE');
INSERT IGNORE INTO compute_allocation_resource_mappings (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time) VALUES
  ('smp-71ee2172bc07537ec447', 'smp-0384acb90162ba62c422', 'res-cpu', 2500, 3600);
INSERT IGNORE INTO compute_allocation_usages (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id) VALUES
  ('smp-8973a187de40708db0da', 'smp-0384acb90162ba62c422', 12622, 12622, '2026-02-15 12:00:00', 'smp-02b1cd6cafbd27dd7201', 'job-ses240029-02', 'res-cpu'),
  ('smp-33251ff628cd501dc185', 'smp-0384acb90162ba62c422', 12744, 12744, '2026-04-15 12:00:00', 'smp-02b1cd6cafbd27dd7201', 'job-ses240029-04', 'res-cpu'),
  ('smp-34541172ae248bfde92a', 'smp-0384acb90162ba62c422', 12866, 12866, '2026-06-15 12:00:00', 'smp-02b1cd6cafbd27dd7201', 'job-ses240029-06', 'res-cpu'),
  ('smp-520674936ddbc4a3ca07', 'smp-0384acb90162ba62c422', 12759, 12759, '2026-02-15 12:00:00', 'smp-318427dda738e0504620', 'job-ses240029-12', 'res-cpu'),
  ('smp-be63b53db20750905ba4', 'smp-0384acb90162ba62c422', 12881, 12881, '2026-04-15 12:00:00', 'smp-318427dda738e0504620', 'job-ses240029-14', 'res-cpu'),
  ('smp-5f848d4163e3d823d89f', 'smp-0384acb90162ba62c422', 13003, 13003, '2026-06-15 12:00:00', 'smp-318427dda738e0504620', 'job-ses240029-16', 'res-cpu'),
  ('smp-87de20f2ff9c078f1bad', 'smp-0384acb90162ba62c422', 12896, 12896, '2026-02-15 12:00:00', 'smp-2fd62ff24cea6be8de4c', 'job-ses240029-22', 'res-cpu'),
  ('smp-63be3f2e1dc38e8cc659', 'smp-0384acb90162ba62c422', 13018, 13018, '2026-04-15 12:00:00', 'smp-2fd62ff24cea6be8de4c', 'job-ses240029-24', 'res-cpu'),
  ('smp-c90ca83a563957d78e84', 'smp-0384acb90162ba62c422', 13140, 13140, '2026-06-15 12:00:00', 'smp-2fd62ff24cea6be8de4c', 'job-ses240029-26', 'res-cpu');

-- Allocation history entries so the allocation History tab has rows.
INSERT IGNORE INTO compute_allocation_diffs (id, compute_allocation_id, diff_type, new_su_amount, status, timestamp, description) VALUES
  ('smp-diff-1', 'smp-c4071741a17cbea31f46', 'ALLOCATION_STATUS_CHANGE', 500000, 'ACTIVE', '2026-01-15 09:00:00', 'Allocation activated after project approval'),
  ('smp-diff-2', 'smp-c4071741a17cbea31f46', 'USAGE_UPDATE', 500000, 'ACTIVE', '2026-04-01 02:00:00', 'SU usage updated from quarterly accounting run'),
  ('smp-diff-3', 'smp-c4071741a17cbea31f46', 'ALLOCATION_INCREASE', 650000, 'ACTIVE', '2026-06-20 14:30:00', 'Supplement approved via change request'),
  ('smp-diff-4', 'smp-81d26521f4dc52bb030b', 'ALLOCATION_STATUS_CHANGE', 250000, 'INACTIVE', '2026-05-30 00:00:00', 'Allocation end date reached');
