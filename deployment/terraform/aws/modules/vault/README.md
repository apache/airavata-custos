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

# Vault Configurations

### Generating the Self Signed Certificate (Development Purpose Only)
- Navigate to the `airavata-custos/terraform/modules/vault/resources` and run the command
`openssl req -x509 -nodes -newkey rsa:2048 -keyout vault.key -out vault.crt -days 365 -config openssl-vault.cnf -extensions req_ext` 
to generate the `vault.cert` and `vault.key`
- Create a Secret using AWS Secrets Manager and copy the `vault.cert` content into the `vault_cert` and `vault_ca`. The `vault.key` content should go to the key `vault_pk`
Then use the corresponding ARN value for variable `vault_secrets_manager_arn` in `terraform.vars`