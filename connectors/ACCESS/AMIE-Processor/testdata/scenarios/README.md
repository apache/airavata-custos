# AMIE scenario fixtures: expected DB state after each mock scenario

Per-scenario test fixtures. Each YAML in this directory describes one mock-server scenario: what gets fired, what should end up in which table, and what the audit log should contain.

## How it's used

The Go integration test at `connectors/ACCESS/AMIE-Processor/pipeline/baseline_integration_test.go` consumes `baseline.yaml`:

1. Stand up the test stack (`make integration-test-amie` does this).
2. POST to the scenario endpoint on the mock server.
3. Wait for `amie_packets` to settle (no `NEW` or in-flight `RUNNING` events).
4. Assert table counts + audit-action breakdown match the YAML's `expectations:` block.

For ad-hoc inspection, the `verification_queries:` block in each YAML gives copy-pasteable SQL you can run against the test DB.

## File format (overview)

```yaml
scenario:
  name: <scenario name>                # matches mock-server's ?type=<name>
  description: <one-line summary>
  determinism: full | partial | random # how much can be asserted exactly

trigger:
  endpoint: POST /test/{site}/scenarios?type=<name>
  preconditions:
    env:
      - VAR_NAME                       # required env vars
    seed:
      - "table: row spec"              # rows that must exist before firing

expectations:
  packets:                             # amie_packets table
    - {type: <packet type>, status: DECODED, count: <N>}
    # ...
  
  tables:                              # domain tables
    <table_name>:
      total_count: <N>                 # or {min, max} for non-deterministic
      rows:                            # ordered or set; see `match:` hint
        - <column>: <value>            # exact value, or pattern object below
          ...
  
  audit_log:
    by_action:
      CREATE_PERSON: <count>
      CREATE_PROJECT: <count>
      # ...
  
  not_expected:                        # tables that should remain empty
    - user_dns
    - user_merges
    # ...

verification_queries:                  # manual SQL for ad-hoc checks
  - name: <human label>
    sql: "SELECT ..."
    expect: <one-line description of the expected output>
```

### Value forms

| Form | Meaning | Example |
|---|---|---|
| literal | exact string / int / bool / null match | `email: hwan@uccs.edu` |
| `${ENV_VAR}` | substituted at test time from process env | `email: ${DEV_EMAIL}` |
| `{regex: "..."}` | column value must match the RE2 regex | `email: {regex: '^user[0-9]+@example\.edu$'}` |
| `{not_null: true}` | column must be non-NULL (any value) | `id: {not_null: true}` |
| `{from: "<table>.<col> where <pred>"}` | runner resolves to that scalar; assertion uses it | `project_pi_id: {from: "users.id where email=${DEV_EMAIL}"}` |

#### Grammar for `${ENV_VAR}`

- **Name**: matches `[A-Z][A-Z0-9_]*`. Anything else is a literal `$`.
- **Substitution context**: runner expands inside YAML scalar values only, NOT inside table or column names.
- **Unset variable**: runner MUST fail the scenario with `env var <name> not set`. Empty-string is treated as set (so `DEV_EMAIL=""` substitutes the empty string).
- **Escaping a literal `$`**: prefix with backslash (`\$NOT_AN_ENV_VAR`).
- **Shell-special characters in the value** (spaces, quotes, semicolons): the runner inserts the value as a SQL parameter (not via string concat), so injection is not a concern. The value flows through bound parameters in the eventual Go assertion.

#### Grammar for `{from: "<table>.<col> where <pred>"}`

- **Form**: `{from: "<table_name>.<column_name> where <predicate>"}`. Single quoted YAML string.
- **Predicate**: SQL-safe `WHERE` fragment. May reference other columns and `${ENV_VAR}` substitutions. Single equality only; joins or sub-selects are not supported.
- **Resolution**: runner executes `SELECT <col> FROM <table> WHERE <pred> LIMIT 2` and:
  - **0 rows** → fail the scenario with `from lookup matched no rows: <expr>`.
  - **1 row** → use the scalar value as the assertion target.
  - **>1 row** → fail the scenario with `from lookup matched multiple rows: <expr>` (the YAML author must constrain the predicate further; ambiguity is always a fixture bug).
- **Order of resolution**: `from` lookups happen AFTER `${ENV_VAR}` substitution. So `from: "users.id where email=${DEV_EMAIL}"` first substitutes the env var, then runs the lookup.
- **Composition**: `from` cannot contain another `from`. Chain by stacking rows in the expected order so each prior row's value is queryable.

### Counts

| Form | Meaning |
|---|---|
| `<N>` | exact count |
| `{min: <N>}` | at least N |
| `{min: <N>, max: <M>}` | range |
| `{>=: <N>}` | shorthand for min |

## Scenario index

| Scenario | YAML | Determinism | Notes |
|---|---|---|---|
| `baseline` | [`baseline.yaml`](baseline.yaml) | Full | **Canonical integration-test fixture.** Hardcoded values covering every handler the mock can drive without needing the Custos-assigned project UUID. |

## What is NOT covered (and why)

Five handlers need the **Custos-assigned project UUID** in the packet body, and the mock can't supply that today (it would need to capture the `notify_project_create` reply and feed the UUID into subsequent packets). These handlers are NOT exercised by `baseline`:

- `request_account_create`
- `request_account_inactivate`
- `request_account_reactivate`
- `request_project_inactivate`
- `request_project_reactivate`

That leaves these tables empty in a mock-only test:
- `compute_cluster_users`
- `compute_allocation_memberships`

And these status-flip code paths unexercised:
- `projects.status`
- `compute_allocations.status`
- `compute_allocation_memberships.membership_status`
- `compute_cluster_users.status`

Closing the gap requires either reply-aware mock sequencing or a Go runner that captures replies between packet posts. Tracked as follow-up.

### What this fixture also does NOT exercise (beyond the 5 untestable handlers)

- **Email-uniqueness collision**: `baseline` never re-creates a user with a clashing email. The `UpdateUser` uniqueness guard is not exercised by this fixture.
- **Intra-packet audit ordering**: the fixture asserts `by_action` counts only, not the temporal order of audits within a single packet. A regression that reordered, say, `CREATE_PROJECT` after `CREATE_ALLOCATION` would not fail this fixture.
- **Retry storms**: every event in `baseline` reaches `SUCCEEDED` on the first attempt. Handlers that succeed only after N retries pass the totals but mask transient bugs. The `processing_events.by_status` block flags any non-SUCCEEDED final state, but does not flag retry count per event.
- **Concurrent re-delivery**: the supplement (packet 3) arrives strictly after packet 1 finishes. Find-or-create TOCTOU races need concurrent fires from the load workload, not this fixture.
- **Read-path race on `oidc_sub`**: packet 3's PI lookup re-uses the bl-pi-001 row created by packet 1, so this fixture covers the read path through `FindBySourceAndExternalID`, but only serially.

## Other mock scenarios (not part of the deterministic fixture suite)

These exist in the mock for ad-hoc load testing but do not have YAML fixtures because their data is randomized per run:

| Scenario | Use |
|---|---|
| `mixed` | 6 success + 4 failure (random GIDs); quick sanity sweep |
| `success_only` | 8 success packets (random) |
| `failures_only` | 8 deliberately malformed packets (random); validation-path testing |
| `heavy` | 15 success + 10 failure (random) |
| `all_handlers` | One of each handler type; pre-creates users for merge + DN handlers |
| `dev_email` | Scripted dev-loop scenario binding `DEV_EMAIL`; partially fails because account_create packets hit the missing-UUID issue |

Validate these scenarios via the k6 load workload + invariant-style checks, not exact-row assertions.

## Out of scope

- Reply-aware mock sequencing would unblock the 5 untestable handlers but adds Python complexity. Tracked as a follow-up.
- Cross-scenario expectations (chaining multiple scenarios in one test run) are explicitly NOT modeled; each scenario assumes a fresh DB.
