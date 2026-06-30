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

# Glossary

Domain terms used across Custos, grouped by where they originate. Connectors
with well-known native vocabulary (e.g. SLURM associations, partitions) are
not duplicated here — refer to the relevant upstream docs.

## Core domain

Concepts that belong to the Custos core.

- **HPC** — High-Performance Computing. The cluster-computing domain Custos serves (e.g. ACCESS-CI sites, university research clusters).
- **Allocation** — A grant of compute time on a cluster, attached to a project. Has a start/end date, a resource (cluster), and rates.
- **Allocation membership** — A user's tie to an allocation (PI, CO_PI, ALLOCATION_MANAGER, or plain member).
- **PI / CO_PI / ALLOCATION_MANAGER** — Project-level governance roles. PI = Principal Investigator (the responsible researcher). CO_PI = Co-Principal Investigator. ALLOCATION_MANAGER = delegated allocation administrator.

## Identity & COmanage

Federated identity stack Custos integrates with.

- **CILogon** — Federated identity service used by US research-computing communities. Issues OIDC tokens and hosts COmanage.
- **COmanage** — Identity registry software from Internet2 used to provision POSIX cluster accounts. Custos talks to a hosted instance via CILogon.
- **DN** — Distinguished Name. LDAP identifier for a person (e.g. an X.509 subject DN).

## ACCESS-CI / AMIE connector

Terms specific to the ACCESS-CI allocation exchange protocol and its
connector (`connectors/ACCESS/AMIE-Processor/`).

- **AMIE** — Account Management Information Exchange. ACCESS-CI's protocol for moving allocation, project, and account state between sites. A REST API plus a packet-and-reply state machine.
- **AMIE packet** — One unit of state in the AMIE protocol. Carries a transaction type (e.g. `request_project_create`, `request_account_create`).
- **Packet families (`request_*`, `data_*`, `notify_*`, `inform_*`)** — AMIE groups packets by intent. `request_*` is "do this" (mutating). `data_*` carries supplementary detail for a prior request (e.g. DNs created in a peer system). `notify_*` is a reply confirming a `request_*` succeeded. `inform_*` carries terminal state, e.g. `inform_transaction_complete` ends a multi-packet exchange.
- **Site code** — Identifier for one ACCESS-CI participating site (e.g. `TESTSITE`, `INDIANA`).
- **Person merge (survivor / retiree)** — When two AMIE person records turn out to be the same human, AMIE sends a `request_person_merge` packet that names a survivor (the record to keep) and a retiree (the record to fold into the survivor).
