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

output "alb_dns_name" {
  value = module.alb.alb_dns_name
}

output "alb_log_bucket" {
  value = module.alb.access_logs_bucket_id
}

output "ecr_repo" {
  value = module.ecr.repository_url
}

output "ecs_cluster" {
  value = aws_ecs_cluster.keycloak.name
}

output "ecs_service" {
  value = module.ecs.service_name
}

output "rds_cluster_endpoint" {
  value = module.rds_cluster.endpoint
}

output "rds_cluster_reader_endpoint" {
  value = module.rds_cluster.reader_endpoint
}

output "rds_cluster_database_name" {
  value = module.rds_cluster.database_name
}

output "rds_cluster_master_username" {
  value = module.rds_cluster.master_username
}
