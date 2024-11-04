# Custos AWS Deployment

### Create the Deployment

Follow the instructions below to create the AWS deployment:

1. Navigate to `airavata-custos/terraform/aws` and run the `terraform init` command to initialize the Terraform
   workspace, which includes downloading the required Terraform modules and providers.
2. Run `terraform plan` to preview the infrastructure changes that Terraform plans to make based on the configuration
   files. This command allows you to review what will be created, updated, or destroyed.
3. Run `terraform apply -auto-approve` to execute the proposed changes and create the resources, including the state
   bucket and DynamoDB locking table. When this command is executed for the first time, the Terraform state will be
   stored locally. To migrate the Terraform state to AWS, run `terraform init -force-copy` only once after the resources
   are created.

### Destroy the Deployment

To safely destroy the deployment, follow these steps:

1. Set the `force_destroy` parameter to `true` in the `terraform_state_backend` module within the `main.tf` file. This
   step is necessary to ensure the S3 state bucket can be deleted even if it contains objects.
2. Run `terraform apply -target module.terraform_state_backend -auto-approve` to apply this change, enabling the
   deletion of the S3 state bucket.
3. Run `terraform init -force-copy` to migrate the Terraform state from the S3 bucket back to local storage. This step
   prepares for a complete destruction of the infrastructure.
4. Finally, run `terraform destroy -auto-approve` to remove all resources defined in the Terraform configuration. This
   command deletes the deployment from AWS.
