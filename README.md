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

Custos is a security middleware for science gateways and HPC research computing, developed under the [Apache Airavata](https://airavata.apache.org/) umbrella. It provides identity and access management, credential storage, federated authentication, and resource allocation services through a language-independent API.

The project is currently being rebuilt around an HPC allocation management focus.

**[Project website](https://airavata.apache.org/custos/)**

## Repository Layout

Custos is composed of pluggable pieces a deployment site mixes and matches.

```
airavata-custos/
├── core/          # Shared contracts and domain models
├── connectors/    # Adapters to external allocation systems (ACCESS-CI, SLURM, ...)
├── extensions/    # Node-side components a site may opt into (PAM, SSH cert signer)
└── dev-ops/       # Local compose stack, Terraform, Ansible
```

| Area | Purpose | Examples |
|------|---------|----------|
| `core/` | Go interfaces and shared domain types that connectors and extensions depend on | `accountprovisioning.Provisioner` |
| `connectors/` | Protocol adapters that bring external state into Custos | `ACCESS/AMIE-Processor`, `SLURM/Association-Mapper` |
| `extensions/` | Independent services that run alongside Custos to extend HPC node behavior | `CILogon-SSH-PAM`, `SSH-Certificate-Signer` |
| `dev-ops/` | Local dev stack and deployment automation | `compose/`, `terraform/`, `account-provisioning/` |

For runtime topology and the audit conventions every component follows, see [docs/architecture.md](docs/architecture.md). New connector authors should start with [docs/contributing/writing-a-connector.md](docs/contributing/writing-a-connector.md).

## Prerequisites

* Go 1.24+
* Docker and Docker Compose

## Quick Start

```sh
git clone https://github.com/apache/airavata-custos.git
cd airavata-custos
```

See [INSTALL.md](INSTALL.md) to bring up the dev stack and run the server. See each connector's and extension's README for run and configuration details.

## Documentation

- [INSTALL.md](INSTALL.md): run the server locally against the dev compose stack
- [CONTRIBUTING.md](CONTRIBUTING.md): coding conventions, build, and test workflow
- [docs/architecture.md](docs/architecture.md): runtime topology and audit conventions
- [docs/contributing/writing-a-connector.md](docs/contributing/writing-a-connector.md): authoring guide for new connectors
- [docs/glossary.md](docs/glossary.md): domain terms (HPC, AMIE, COmanage, PI, DN, site code, etc.)
- [docs/API-Docs.md](docs/API-Docs.md): REST API reference
- [docs/Allocation-Data-Models.md](docs/Allocation-Data-Models.md): domain model overview
- [docs/ACCESS-HPC-Reference.md](docs/ACCESS-HPC-Reference.md): ACCESS-CI integration reference

## Questions or Need Help?

* Open a [GitHub issue](https://github.com/apache/airavata-custos/issues)
* Join the [Airavata dev mailing list](https://airavata.apache.org/mailing-list.html)

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
