#!/bin/bash

echo "Creating databases and users..."

mysql -u root -p"$MYSQL_ROOT_PASSWORD" <<-EOSQL
    CREATE DATABASE IF NOT EXISTS custos;
    CREATE DATABASE IF NOT EXISTS keycloak;
    CREATE USER IF NOT EXISTS 'admin'@'%' IDENTIFIED BY 'admin';
    GRANT ALL PRIVILEGES ON custos.* TO 'admin'@'%';
    GRANT ALL PRIVILEGES ON keycloak.* TO 'admin'@'%';
    FLUSH PRIVILEGES;
EOSQL

echo "Databases and users created"