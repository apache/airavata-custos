# AMIE Traffic Simulation

k6 load test that generates AMIE test scenarios at varying rates to simulate real-world traffic patterns.

## Prerequisites

```bash
# macOS
brew install k6
```

## Usage

```bash
# Basic run (resets test server automatically)
AMIE_API_KEY=<your-key> k6 run amie-traffic.js

# With Prometheus metrics export
AMIE_API_KEY=<key> k6 run --out experimental-prometheus-rw=http://localhost:9090/api/v1/write amie-traffic.js

# Override AMIE endpoint
AMIE_BASE_URL=https://custom-endpoint AMIE_SITE=MySite AMIE_API_KEY=<key> k6 run amie-traffic.js
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

| Variable | Default | Description |
|----------|---------|-------------|
| `AMIE_API_KEY` | (required) | AMIE API authentication key |
| `AMIE_BASE_URL` | `https://a3mdev.xsede.org/amie-api-test` | AMIE test API base URL |
| `AMIE_SITE` | `GaTech` | Site code for AMIE |
