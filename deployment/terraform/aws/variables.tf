variable "environment" {
  description = "Environment name (development, production, etc)"
  type        = string
}

variable "namespace" {
  description = "Application namespace"
  type        = string
}

variable "region" {
  description = "AWS region to target"
  type        = string
}

variable "enable_network" {
  description = "Use network module. Set to false to use your own network resources"
  type        = bool
  default     = true
}

variable "vpc_id" {
  description = "AWS VPC ID (if not using network module)"
  type        = string
  default     = ""
}

variable "vpc_cidr" {
  description = "RFC1918 CIDR range for VPC"
  type        = string
  default     = ""
}

variable "public_cidr" {
  description = "RFC1918 CIDR range for public subnets (subset of vpc_cidr)"
  type        = string
  default     = ""
}

variable "private_cidr" {
  description = "RFC1918 CIDR range for private subnets (subset of vpc_cidr)"
  type        = string
  default     = ""
}

variable "public_subnet_ids" {
  description = "List of public subnet IDs for deployment if not using network module"
  type        = list(string)
  default     = []
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs for deployment if not using network module"
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Standard tags for all resources"
  type        = map(any)
  default     = {
    ManagedBy = "Terraform"
  }
}

variable "keycloak_alb_certificate_arn" {
  description = "ACM certificate used by Keycloak ALB"
  type        = string
}

variable "alb_destroy_log_bucket" {
  description = "Destroy ALB log bucket on teardown"
  type        = bool
  default     = true
}

variable "container_cpu_units" {
  description = "CPU units to reserve for container (1024 units == 1 CPU)"
  type        = number
}

variable "container_memory_limit" {
  description = "Container memory hard limit"
  type        = number
}

variable "container_memory_reserved" {
  description = "Container memory starting reservation"
  type        = number
}

variable "keycloak_container_port" {
  description = "Keycloak port exposed in container"
  type        = number
  default     = 8080
}

variable "db_backup_retention_days" {
  description = "How long Database backups are retained"
  type        = number
}

variable "db_backup_window" {
  description = "Daily time range during which backups happen"
  type        = string
  default     = "00:00-02:00"
}

variable "db_cluster_family" {
  description = "Family of DB cluster parameter group"
  type        = string
  default     = "aurora-postgresql15"
}

variable "db_cluster_size" {
  description = "Number of RDS cluster instances"
  type        = number
}

variable "db_engine_version" {
  description = "Version of DB engine to use"
  type        = string
  default     = "15.4"
}

variable "db_instance_type" {
  description = "Instance type used for RDS instances"
  type        = string
}

variable "db_maintenance_window" {
  description = "Weekly time range during which system maintenance can occur (UTC)"
  type        = string
  default     = "sat:03:00-sat:04:00"
}

variable "deletion_protection" {
  description = "Protect supporting resources from being deleted (ALB and RDS)"
  type        = bool
  default     = false
}

variable "deployment_maximum_percent" {
  description = "Maximum task instances allowed to run"
  type        = number
}

variable "deployment_minimum_healthy_percent" {
  description = "Minimum percentage of healthy task instances"
  type        = number
}

variable "desired_count" {
  description = "Number of ECS task instances to run"
  type        = number
}

variable "keycloak_dns_name" {
  description = "Keycloak DNS"
  type        = string
}

variable "dns_zone_id" {
  description = "Route53 Zone ID hosting Services"
  type        = string
}

variable "encryption_configuration" {
  type = object({
    encryption_type = string
    kms_key         = any
  })
  description = "ECR encryption configuration"
  default     = {
    encryption_type = "AES256"
    kms_key         = null
  }
}

variable "http_redirect" {
  description = "Controls whether port 80 should redirect to 443 (or not listen)"
  type        = bool
  default     = true
}

variable "http_ingress_cidr_blocks" {
  description = "CIDR ranges allowed to connect to service port 80"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "https_ingress_cidr_blocks" {
  description = "CIDR ranges allowed to connect to service port 443"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "jvm_heap_min" {
  description = "Minimum JVM heap size for application in MB"
  type        = number
}

variable "jvm_heap_max" {
  description = "Maximum JVM heap size for application in MB"
  type        = number
}

variable "jvm_meta_min" {
  description = "Minimum JVM meta space size for application in MB"
  type        = number
}

variable "jvm_meta_max" {
  description = "Maximum JVM meta space size for application in MB"
  type        = number
}

variable "internal" {
  description = "Whether environment should be exposed to Internet (if not using network module)"
  type        = string
  default     = false
}

variable "log_retention_days" {
  description = "Log retention for CloudWatch logs"
  type        = number
}

variable "rds_source_region" {
  description = "Region of primary RDS cluster (needed to support encryption)"
  type        = string
  default     = ""
}

variable "route_table_ids" {
  description = "List of route tables used by s3 VPC endpoint (if not using network module)"
  type        = list(string)
  default     = []
}

variable "stickiness" {
  type = object({
    cookie_duration = number
    enabled         = bool
  })
  description = "Target group sticky configuration"
  default     = {
    cookie_duration = null
    enabled         = false
  }
}

variable "vault_alb_certificate_arn" {
  description = "ACM certificate used by Vault ALB"
  type        = string
}

variable "vault_container_port" {
  description = "Vault port"
  type        = number
  default     = 8200
}

variable "vault_ami" {
  description = "AMI used for Vault"
  type        = string
}

variable "ec2_ssh_key_name" {
  description = "key pair to use for SSH access to instance"
  type        = string
}

variable "vault_instance_type" {
  type        = string
  description = "EC2 instance type"
}

variable "vault_version" {
  type        = string
  description = "Vault version"
}

variable "vault_leader_tls_servername" {
  type        = string
  description = "One of the shared DNS SAN used to create the certs use for mTLS"
}

variable "vault_secrets_manager_arn" {
  type        = string
  description = "Secrets manager ARN"
}

variable "vault_min_nodes" {
  type        = number
  description = "Minimum number of Vault nodes to deploy in ASG"
}

variable "vault_max_nodes" {
  type        = number
  description = "Minimum number of Vault nodes to deploy in ASG"
}
