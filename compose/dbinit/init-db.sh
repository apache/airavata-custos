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

echo "Creating databases and users..."

mariadb -u root -p"${MARIADB_ROOT_PASSWORD}" <<-EOSQL
    CREATE DATABASE IF NOT EXISTS custos;
    CREATE DATABASE IF NOT EXISTS keycloak;
    CREATE DATABASE IF NOT EXISTS access_ci;
    CREATE USER IF NOT EXISTS 'admin'@'%' IDENTIFIED BY 'admin';
    GRANT ALL PRIVILEGES ON custos.* TO 'admin'@'%';
    GRANT ALL PRIVILEGES ON keycloak.* TO 'admin'@'%';
    GRANT ALL PRIVILEGES ON access_ci.* TO 'admin'@'%';
    FLUSH PRIVILEGES;
EOSQL

echo "Databases and users created"