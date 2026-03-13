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

# Apache Airavata Custos

[![License](http://img.shields.io/badge/license-Apache--2-blue.svg?style=flat)](https://apache.org/licenses/LICENSE-2.0)
[![GitHub closed pull requests](https://img.shields.io/github/issues-pr-closed/apache/airavata-custos)](https://github.com/apache/airavata-custos/pulls?q=is%3Apr+is%3Aclosed)
[![Build Status](https://travis-ci.org/apache/airavata-custos.png?branch=develop)](https://travis-ci.org/github/apache/airavata-custos)

Custos is a multi-tenant security middleware for science gateways, developed under the [Apache Airavata](https://airavata.apache.org/) umbrella. It provides identity and access management, credential storage, federated authentication, and resource allocation services to science gateway frameworks through a language-independent API. Custos is designed as a set of composable product components that can be deployed independently or together, built on a scalable architecture to deliver highly available, fault-tolerant operations.

**[Project website](https://airavata.apache.org/custos/)**

## Components

### Identity Server (`identity/`)

The Identity Server is the core IAM component of Custos. It handles user identity and access management, tenant profile management, resource secrets management, and groups and sharing management. Built with Java 17 and Spring Boot, it integrates with Keycloak for federated authentication, HashiCorp Vault for secrets management, and MariaDB for persistence.

| Module | Description |
|--------|-------------|
| `identity/core` | Domain entities, repositories, protobuf definitions, mappers |
| `identity/services` | Business logic, Keycloak and Vault integrations |
| `identity/api` | REST API controllers |
| `identity/application` | Spring Boot entry point |

See [`identity/README.md`](identity/README.md) for setup and development instructions.

### Allocations (`allocations/`)

The Allocations component provides meta-allocation authority services for HPC and cloud resources. It acts as a bridge between Custos-managed tenants and external resource allocation providers.

| Module | Description |
|--------|-------------|
| `allocations/access-ci-service` | ACCESS CI AMIE packet adapter |

Additional allocation adapters for other resource providers are planned. See `allocations/README.md` for details as they become available.

## Repository Layout

```
airavata-custos/
├── identity/          # Identity Server
├── allocations/       # Allocation management and usage
├── compose/           # Docker Compose for local development
├── deployment/        # Terraform configurations (AWS)
├── legacy/            # Archived modules (not actively maintained)
└── pom.xml            # Root Maven reactor
```

## Prerequisites

* Java 17
* Maven 3.6+
* Docker and Docker Compose

## Quick Start

Clone the repository:

```sh
git clone https://github.com/apache/airavata-custos.git
cd airavata-custos
```

Start the backing services (Keycloak, MariaDB, Vault, Adminer):

```sh
cd compose
docker compose up -d
```

Build all components:

```sh
mvn clean install
```

Refer to each component's README for detailed configuration and run instructions.

## Questions or Need Help?

* Open a [GitHub issue](https://github.com/apache/airavata-custos/issues)
* Subscribe to the Custos mailing list: `custos-subscribe@airavata.apache.org`

## Publications

```
@inproceedings{10.1145/3311790.3396635,
author = {Ranawaka, Isuru and Marru, Suresh and Graham, Juleen and Bisht, Aarushi and Basney, Jim and Fleury, Terry and Gaynor, Jeff and Wannipurage, Dimuthu and Christie, Marcus and Mahmoud, Alexandru and Afgan, Enis and Pierce, Marlon},
title = {Custos: Security Middleware for Science Gateways},
year = {2020},
isbn = {9781450366892},
publisher = {Association for Computing Machinery},
address = {New York, NY, USA},
url = {https://doi.org/10.1145/3311790.3396635},
doi = {10.1145/3311790.3396635},
booktitle = {Practice and Experience in Advanced Research Computing},
pages = {278–284},
numpages = {7},
location = {Portland, OR, USA},
series = {PEARC '20}
}
```

```
@inproceedings{10.1145/3491418.3535177,
author = {Ranawaka, Isuru and Goonasekara, Nuwan and Afgan, Enis and Basney, Jim and Marru, Suresh and Pierce, Marlon},
title = {Custos Secrets: A Service for Managing User-Provided Resource Credential Secrets for Science Gateways},
year = {2022},
isbn = {9781450391610},
publisher = {Association for Computing Machinery},
address = {New York, NY, USA},
url = {https://doi.org/10.1145/3491418.3535177},
doi = {10.1145/3491418.3535177},
booktitle = {Practice and Experience in Advanced Research Computing},
articleno = {40},
numpages = {4},
location = {Boston, MA, USA},
series = {PEARC '22}
}
```

## Acknowledgment

This project is funded by the National Science Foundation (NSF).

We are grateful to [Trusted CI](https://www.trustedci.org/) for conducting the [First Principles Vulnerability Assessment (FPVA)](https://dl.acm.org/doi/10.1145/1866835.1866852) for this software and providing security architecture guidance and improvements.
