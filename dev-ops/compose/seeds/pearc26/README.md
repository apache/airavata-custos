# PEARC26 workshop seed set

Self-contained, ordered seed data for the Access Requests trial-allocation
flow. Duplicates parts of the older seeds on purpose — apply this folder
alone on a fresh DB and the whole flow works.

All cluster facts (name, partition, CPU counts) were read from the real
dev Slurm cluster (`nexus-dev`, login1 / 149.165.159.51) on 2026-07-16.

Apply in order, after core migrations have run:

```
for f in dev-ops/compose/seeds/pearc26/0*.sql; do
  docker exec -i custos_db mariadb -uadmin -padmin custos < "$f"
done
```

| File | Creates |
|---|---|
| 01_org_cluster.sql | PEARC26 Attendees org; the `nexus-dev` cluster row |
| 02_resources_rates.sql | `debug` partition catalog entry + SU rate |
| 03_admin_role.sql | approver/bootstrap-admin user + `access-approver` role |
| 04_project_allocation.sql | project (PI = the seed-03 admin), `pearc26-tutorial` allocation, resource mapping |
| 05_access_event.sql | the PEARC26 event code row |

Notes:

- **The cluster row reuses id `00000000-0000-0000-0000-000000000001`
  (= `CUSTOS_CLUSTER_ID`).** The COmanage subscriber only provisions
  cluster users whose `compute_cluster_id` equals that env value; a
  different id means approvals silently never reach COmanage. The row
  upserts the name to `nexus-dev` if `default_cluster.sql` got there
  first.
- 05 needs the `access_events` migration; apply it after the
  access-requests feature migration has run.
- The seed-03 user is approver, bootstrap admin, AND project PI: set
  `CUSTOS_BOOTSTRAP_ADMIN_EMAIL=approver@pearc26.local` (or edit that
  email to your own) and boot grants super_admin idempotently.
- **Seeded allocations fire no events.** The SLURM Association-Mapper
  and COmanage provisioner are event-driven; SQL inserts bypass them.
  Live runs create the project/allocation/mapping via the API (see the
  goal spec's live gate).
- Everything is `INSERT IGNORE` / upsert — safe to re-apply.
