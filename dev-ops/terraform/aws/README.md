<!--
    Licensed to the Apache Software Foundation (ASF) under one
    or more contributor license agreements.  See the NOTICE file
    distributed with this work for additional information
    regarding copyright ownership.  The ASF licenses this file
    to you under the Apache License, Version 2.0 (the
    "License"); you may not use this file except in compliance
    with the License.  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing,
    software distributed under the License is distributed on an
    "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
    KIND, either express or implied.  See the License for the
    specific language governing permissions and limitations
    under the License.
-->

# Custos AWS deployment

This Terraform configuration deploys the shared services a Custos site runs on AWS. It creates:

- Keycloak on ECS behind an Application Load Balancer, with an Aurora PostgreSQL cluster.
- A HashiCorp Vault cluster on EC2 with Raft storage and KMS auto-unseal.
- The S3 bucket and DynamoDB table that hold its own Terraform state.
- With `enable_network = true`, a VPC with public and private subnets.

It does not deploy the Custos server.

## Before you start

| Item | Where |
|---|---|
| Variable values | `terraform.tfvars`. `variables.tf` describes each. |
| TLS certificates for the load balancers | ACM, referenced by `keycloak_alb_certificate_arn` and `vault_alb_certificate_arn` |
| Vault node TLS material | Secrets Manager, referenced by `vault_secrets_manager_arn`: [Vault module](modules/vault/README.md) |
| Keycloak image | Built from `modules/keycloak/resources/Dockerfile`. Pushed with tag `latest` to the module's ECR repository once the first apply creates it. |

## Create the deployment

Follow the instructions below to create the AWS deployment:

1. Navigate to `airavata-custos/dev-ops/terraform/aws` and run the `terraform init` command to initialize the Terraform
   workspace, which includes downloading the required Terraform modules and providers.
2. Run `terraform plan` to preview the infrastructure changes that Terraform plans to make based on the configuration
   files. This command allows you to review what will be created, updated, or destroyed.
3. Run `terraform apply -auto-approve` to execute the proposed changes and create the resources, including the state
   bucket and DynamoDB locking table. When this command is executed for the first time, the Terraform state will be
   stored locally. To migrate the Terraform state to AWS, move `config/backend.tf` into this directory, then run
   `terraform init -force-copy` only once after the resources are created.

## Destroy the deployment

To safely destroy the deployment, follow these steps:

1. Set the `force_destroy` parameter to `true`, and `terraform_backend_config_file_path` to `""`, in the
   `terraform_state_backend` module within the `main.tf` file. This
   step is necessary to ensure the S3 state bucket can be deleted even if it contains objects.
2. Run `terraform apply -target module.terraform_state_backend -auto-approve` to apply this change, enabling the
   deletion of the S3 state bucket.
3. Delete `backend.tf` from this directory, then run `terraform init -force-copy` to migrate the Terraform state from the S3 bucket back to local storage. This step
   prepares for a complete destruction of the infrastructure.
4. Finally, run `terraform destroy -auto-approve` to remove all resources defined in the Terraform configuration. This
   command deletes the deployment from AWS.
