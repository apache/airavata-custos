#!/bin/bash

# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements. See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership. The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License. You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied. See the License for the
# specific language governing permissions and limitations
# under the License.

# Runs once on first container start (empty data dir). The custos database
# and the admin superuser come from POSTGRES_DB / POSTGRES_USER.
set -e

echo "Creating databases..."

psql -v ON_ERROR_STOP=1 -U "${POSTGRES_USER}" -d postgres <<-EOSQL
    CREATE DATABASE keycloak;
    CREATE DATABASE custos_signer;
    CREATE DATABASE custos_test;
    ALTER DATABASE custos SET timezone = 'UTC';
    ALTER DATABASE keycloak SET timezone = 'UTC';
    ALTER DATABASE custos_signer SET timezone = 'UTC';
    ALTER DATABASE custos_test SET timezone = 'UTC';
EOSQL

echo "Databases created"
