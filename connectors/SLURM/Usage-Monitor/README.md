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

# SLURM Usage Monitor

The SLURM Usage Monitor charges finished SLURM jobs to Custos allocations. It reads job records from `slurmrestd`. It
writes one usage row per job, which the portal's analytics and the PI's credit totals read. Enable it on a SLURM
cluster whose allocations Custos manages.

## How it works

The monitor polls every 30 seconds. At startup it resumes from the latest usage time recorded for the cluster, or from
24 hours ago. Each poll re-scans the last `usage_lookback`, so jobs that SLURM records late are still charged.

For each finished job, the monitor:

1. Matches the job's account to an allocation by name, ignoring case.
2. Looks up the job's user as a cluster account on the cluster.
3. Looks up the job's partition as an allocation resource on the cluster.
4. Computes the raw amount as the job's `billing` TRES times its run time in hours.
5. Multiplies it by the resource's rate in effect at the job's end time, giving service units.
6. Writes the usage row. A row for the same allocation and job ID is replaced.

A job whose account matches no allocation is ignored. The monitor logs and skips a job when:

- a user, resource or rate lookup finds nothing
- its run time is not positive
- its `billing` value is zero

A skipped job does not stop the poll.

Set `TRESBillingWeights` and `PriorityFlags=MAX_TRES` on each partition. Without them, `billing` is the CPU count and
the charge is wrong.

## Configuration

Settings are in the `connectors.slurm-usage-monitor` block of [`config/custos.yaml`](../../../config/custos.yaml).

| Key | Sample value | Notes |
|---|---|---|
| `slurm_api.*` | `${SLURM_API}` and so on | The four keys and fallbacks of the [Association-Mapper](../Association-Mapper/README.md#configuration) |
| `cluster_id` | `${CUSTOS_CLUSTER_ID}` | Custos cluster the usage belongs to. Falls back to `SLURM_MONITOR_CLUSTER_ID`, then to `slurm-cluster` with a warning. |
| `usage_lookback` | `15m` | How far back each poll re-scans. Default `15m`. |

## Development

The `slurmrestd` client is `connectors/SLURM/Rest-Client/pkg/client`.

```bash
# From the repository root
go test ./connectors/SLURM/Usage-Monitor/...
```
