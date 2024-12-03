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

variable "environment" {
  description = "Environment name (development, production, etc)"
  type        = string
}

variable "name" {
  description = "Used by modules to construct labels"
  type        = string
  default     = "vault"
}

variable "tags" {
  description = "Default tags applied to resources"
  type        = map(string)
}

variable "region" {
  description = "AWS region to target"
  type        = string
}

variable "namespace" {
  description = "Application namespace"
  type        = string
}

variable "vpc_id" {
  description = "AWS VPC ID"
  type        = string
}

variable "alb_destroy_log_bucket" {
  description = "Destroy ALB log bucket on teardown"
  type        = bool
}

variable "alb_certificate_arn" {
  description = "ACM certificate ARN used by ALB"
  type        = string
}

variable "deletion_protection" {
  description = "Protect resources from being deleted"
  type        = bool
}

variable "http_ingress_cidr_blocks" {
  description = "CIDR ranges allowed to connect to service port 80"
  type        = list(string)
}

variable "http_redirect" {
  description = "Controls whether port 80 should redirect to 443 (or not listen)"
  type        = bool
}

variable "https_ingress_cidr_blocks" {
  description = "CIDR ranges allowed to connect to service port 443"
  type        = list(string)
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs"
  type        = list(string)
}

variable "container_port" {
  description = "Vault port exposed in container"
  type        = number
}

variable "stickiness" {
  type = object({
    cookie_duration = number
    enabled         = bool
  })
  description = "Target group sticky configuration"
}

variable "alb_health_check_path" {
  type        = string
  description = "Vault health check path"
  default     = "/v1/sys/health"
}

variable "ubuntu_ami" {
  type        = string
  description = "AMI for Ubuntu"
}

variable "instance_type" {
  type        = string
  description = "EC2 instance type"
}

variable "ssh_key_name" {
  type        = string
  description = "key pair to use for SSH access to instance"
}

variable "vault_version" {
  type        = string
  description = "Vault version"
}

variable "leader_tls_servername" {
  type        = string
  description = "One of the shared DNS SAN used to create the certs use for mTLS"
}

variable "secrets_manager_arn" {
  type        = string
  description = "Secrets manager ARN"
}

variable "min_nodes" {
  type        = number
  description = "Minimum number of Vault nodes to deploy in ASG"
}

variable "max_nodes" {
  type        = number
  description = "Minimum number of Vault nodes to deploy in ASG"
}