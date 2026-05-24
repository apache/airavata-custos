# local-amie

Isolated Docker stack for AMIE connector integration tests. Runs on offset
ports so it coexists with the main dev stack in `dev-ops/compose/`.

## Services

| Service     | Container               | Host port |
|-------------|-------------------------|-----------|
| MariaDB     | `custos_amie_test_db`   | `3307`    |
| mock-amie   | `custos_amie_test_mock` | `8181`    |

## Usage

```bash
make up      # build + start, blocks until both are healthy
make down    # stop + remove + drop volumes
make logs    # tail both services
make ps      # status
```

The integration test runner at `scripts/run-amie-integration-tests.sh`
handles up/down for you.
