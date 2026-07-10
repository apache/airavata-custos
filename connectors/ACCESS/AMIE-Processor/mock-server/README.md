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

# AMIE Traffic Tools

Two scripts you can use together or on their own:

- **`mock-amie-server.py`** — a local AMIE server to generate different packet types (valid, invalid, or a mix).
- **`amie-traffic.js`** — a k6 load test. Use this when you want continuous traffic at varying rates, either against the local server or with the AMIE test endpoint.

## Mock server (different packet types)

Start it:

```bash
python3 mock-amie-server.py
```

```bash
# mix of success and failure packets
curl -X POST 'http://localhost:8180/test/TESTSITE/scenarios?type=mixed'

# only success packets
curl -X POST 'http://localhost:8180/test/TESTSITE/scenarios?type=success_only'

# only failure packets
curl -X POST 'http://localhost:8180/test/TESTSITE/scenarios?type=failures_only'

# heavy batch
curl -X POST 'http://localhost:8180/test/TESTSITE/scenarios?type=heavy'

# dev email across multiple projects with different roles (requires DEV_EMAIL)
curl -X POST 'http://localhost:8180/test/TESTSITE/scenarios?type=dev_email'
```

Point `access-amie` at it with `AMIE_BASE_URL=http://localhost:8180`.

### `dev_email` scenario

Run the `dev_email` scenario to generate AMIE packets placing DEV_EMAIL in multiple projects with different roles.

```bash
DEV_EMAIL=jdoe@etest.org python3 mock-amie-server.py
curl -X POST 'http://localhost:8180/test/TESTSITE/scenarios?type=dev_email'
```

Today the access-amie handlers persist PI and USER memberships only; Co-PI and Allocation Manager positions require handler enhancement to read a role field from the AMIE packet.

## Load test (k6)

Install k6, then run it. If you're hitting the local server, start that one first:

```bash
# macOS
brew install k6

# Against the local server
AMIE_BASE_URL=http://localhost:8180 AMIE_API_KEY=dev AMIE_SITE=TESTSITE k6 run amie-traffic.js

# Against the AMIE test endpoint
AMIE_API_KEY=<your-key> AMIE_SITE=<your-site> k6 run amie-traffic.js
```

## Traffic Profile (~20 minutes)

```
VUs
 8 |                                              *
 5 |         ***************                     * *
 2 |       **               *****   ************     **
 1 | ******                      ***                   ****
   |──────────────────────────────────────────────────────── time
     warm  ramp    peak     cool  steady  quiet  spike  recovery
```

## Environment Variables

| Variable | Default                                  | Description |
|----------|------------------------------------------|-------------|
| `AMIE_API_KEY` | (required)                               | AMIE API authentication key |
| `AMIE_BASE_URL` | `https://a3mdev.xsede.org/amie-api-test` | AMIE test API base URL |
| `AMIE_SITE` | `TESTSITE`                               | Site code for AMIE |
