#
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
#

output "public_subnet_cidrs" {
  value = var.enable_network ? module.network[0].public_subnet_cidrs : []
}

output "private_subnet_cidrs" {
  value = var.enable_network ? module.network[0].private_subnet_cidrs : local.private_subnet_cidrs
}

output "state_table" {
  value = module.terraform_state_backend.dynamodb_table_name
}

output "state_bucket" {
  value = module.terraform_state_backend.s3_bucket_domain_name
}

output "alb_dns_name" {
  value = module.keycloak.alb_dns_name
}

output "alb_log_bucket" {
  value = module.keycloak.alb_log_bucket
}

output "ecr_repo" {
  value = module.keycloak.ecr_repo
}

output "ecs_cluster" {
  value = module.keycloak.ecs_cluster
}

output "ecs_service" {
  value = module.keycloak.ecs_service
}

output "rds_cluster_endpoint" {
  value = module.keycloak.rds_cluster_endpoint
}

output "rds_cluster_reader_endpoint" {
  value = module.keycloak.rds_cluster_reader_endpoint
}

output "rds_cluster_database_name" {
  value = module.keycloak.rds_cluster_database_name
}

output "rds_cluster_master_username" {
  value     = module.keycloak.rds_cluster_master_username
  sensitive = true
}
