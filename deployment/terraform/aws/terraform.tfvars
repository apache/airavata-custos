environment = "dev"
namespace   = "custos"
region      = "us-east-2"

vpc_cidr     = "10.20.30.0/24"
public_cidr  = "10.20.30.0/25"
private_cidr = "10.20.30.128/25"


keycloak_alb_certificate_arn = "<KEYCLOAK_ALB_CERT_ARN>"
keycloak_dns_name            = "auth.usecustos.org"
dns_zone_id                  = "<DNS_ZONE_ID>"

container_cpu_units                = 1024
container_memory_limit             = 2048
container_memory_reserved          = 1024
jvm_heap_min                       = 512
jvm_heap_max                       = 1024
jvm_meta_min                       = 128
jvm_meta_max                       = 512
deployment_maximum_percent         = 100
deployment_minimum_healthy_percent = 50
desired_count                      = 1
log_retention_days                 = 5

db_instance_type         = "db.r6g.large"
db_backup_retention_days = 5
db_cluster_size          = 2


vault_alb_certificate_arn   = "<VAULT_ALB_CERT_ARN>"
vault_ami                   = "<AMI_ID>"
ec2_ssh_key_name            = "custos-auth"
vault_instance_type         = "m5.xlarge"
vault_version               = "1.11.0"
vault_leader_tls_servername = "vault.usecustos.org"
vault_secrets_manager_arn   = "<SECRETS_MANAGER_ARN>"
vault_min_nodes             = 1
vault_max_nodes             = 5
