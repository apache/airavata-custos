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

# Apache Airavata Custos Security

[![License](http://img.shields.io/badge/license-Apache--2-blue.svg?style=flat)](https://apache.org/licenses/LICENSE-2.0)
[![GitHub closed pull requests](https://img.shields.io/github/issues-pr-closed/apache/airavata-custos)](https://github.com/apache/airavata-custos/pulls?q=is%3Apr+is%3Aclosed)
[![Build Status](https://travis-ci.org/apache/airavata-custos.png?branch=develop)](https://travis-ci.org/github/apache/airavata-custos)

Science gateways represent potential targets for cybersecurity threats to users, scientific research, and scientific resources. Custos is a software framework that provides common security operations for science gateways, including user identity and access management, gateway tenant profile management, resource secrets management, and groups and sharing management. The goals of the Custos project are to provide these services to a wide range of science gateway frameworks, providing the community with an open-source, transparent, and reviewed code base for common security operations; and to operate trustworthy security services for the science gateway community using this software base. To accomplish these goals, we implement Custos using a scalable microservice architecture that can provide highly available, fault-tolerant operations. Custos exposes these services through a language-independent Application Programming Interface that encapsulates science gateway usage scenarios.

**To find out more, please check out the [Custos website](https://airavata.apache.org/custos/) and the [Custos wiki](https://cwiki.apache.org/confluence/display/CUSTOS/Home).**

## Quickstart

## Installation Instructions
### Deploy Custos on remote server
### Setup Custos for local development

#### Prerequisites

* Java 17

* Docker installed on local environment 

#### Clone the repository
  ```
    git clone -b develop https://github.com/apache/airavata-custos.git
    
  ```

#### Build source code
  
  Following  command builds the Custos source code and create two docker images of custos_core_server and custos_integration_server
  
  ```
    cd airavata-custos
    mvn clean install
  ```
  
#### Run Custos on docker
  
    Following command starts Custos main services and its depend services
  - Dependent Services
    * Keycloak
    * MySQL
    * HashiCorp Vault
    * CILogon
    
  - Custos Services
    * Custos Core Service
    * Custos Integration Service
    * Custos Rest Proxy
    
  ```
     cd custos-utilities/ide-integration/src/main/containers
     docker-compose up
  ```

#### Bootstrapping Custos  Super Tenant
  
If all services were successfully ran. Custos bootstrap service needs to be run to create a  Super tenant to launch Custos Portal
   ```
    cd custos-utilities/custos-bootstrap-service/
    mvn spring-boot:run
   ```
The above command should create the super tenant and it outputs super tenant credentials. Copy those credentials to configure
Custos Portal.



#### Install Custos Portal Locally

Following the following link to access portal deployment instructions

[custos portal](https://github.com/apache/airavata-custos-portal/blob/master/README.md)

## Custos Integration With External Applications
Custos can be integrated with external applications using Custos REST Endpoints, Python SDK, or Java SDK.

### Integrate Using Java SDK
In order to perform this operation you need to have a already activated tenant in either Custos Managed Services or Your own deployment.
Following instructions are given for locally deployed custos setup which can be extended to any deployment,

####Initializing Custos Java SDK

* Add maven dependency to your project
```
<dependency>
   <groupId>org.apache.custos</groupId>
   <artifactId>custos-java-sdk</artifactId>
   <version>1.1-SNAPSHOT</version>
</dependency>

```
* Initialize Custos Client Provider in your application
```
 CustosClientProvider custosClientProvider = new CustosClientProvider.Builder().setServerHost("localhost")
                    .setServerPort(7000)
                    .setClientId(CUSTOS CLIENT ID) // client Id generated from above step or any active tenant id
                    .setClientSec(CUSTOS CLIENT SECRET)  
                    .usePlainText(true) // Don't use this in production setup
                    .build();
```
Once above step is done, you can use custos available methods for  authentication and authorization purposes
* Sample client code to register and enable a User

```
 UserManagementClient userManagementClient =  custosClientProvider.getUserManagementClient();
 userManagementClient.registerUser("Jhon","Smith","testpassword","smith@1",
                    "jhon@email.com",false);
 userManagementClient.enableUser("Jhon");
 OperationStatus status =  userManagementClient.isUserEnabled("Jhon");
```
##### 

## Roadmap

## Contributing

## Questions or need help?

