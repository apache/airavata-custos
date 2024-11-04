# Vault Configurations

### Generating the Self Signed Certificate (Development Purpose Only)
- Navigate to the `airavata-custos/terraform/modules/vault/resources` and run the command
`openssl req -x509 -nodes -newkey rsa:2048 -keyout vault.key -out vault.crt -days 365 -config openssl-vault.cnf -extensions req_ext` 
to generate the `vault.cert` and `vault.key`
- Create a Secret using AWS Secrets Manager and copy the `vault.cert` content into the `vault_cert` and `vault_ca`. The `vault.key` content should go to the key `vault_pk`
Then use the corresponding ARN value for variable `vault_secrets_manager_arn` in `terraform.vars`