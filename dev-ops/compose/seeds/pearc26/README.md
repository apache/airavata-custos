# PEARC26 workshop setup

Ordered setup for the Access Requests trial-allocation flow. SQL seeds carry
only the rows that fire no events (org, cluster, partition catalog, roles);
everything event-bearing (project, allocation, resource grant) is created
THROUGH THE API by `setup-live.sh` so the SLURM account and grant limits
appear on the cluster automatically.

All cluster facts (name, partition, CPU counts) were read from the real
dev Slurm cluster (`nexus-dev`, login1 / 149.165.159.51) on 2026-07-16.

## From scratch, in order

1. **Start the docker services** (from the repo root):

   ```
   docker compose -f dev-ops/compose/docker-compose.yml up -d
   ```

2. **Run the backend** (applies migrations on startup; needs the SLURM_*,
   COMANAGE_*, OIDC_* values in the repo-root `.env`):

   ```
   set -a && source .env && set +a && go run ./cmd/server
   ```

3. **Seed the event-free data**:

   ```
   for f in dev-ops/compose/seeds/pearc26/0*.sql; do
     docker exec -i custos_db mariadb -uadmin -padmin custos < "$f"
   done
   ```

4. **Create the tutorial allocation through the API** (CILogon device
   auth as the admin; requires the device grant enabled on the client):

   ```
   dev-ops/compose/seeds/pearc26/setup-live.sh
   ```

   Then verify on the cluster:

   ```
   sacctmgr -n show account pearc26-tutorial
   sacctmgr -n show assoc account=pearc26-tutorial
   ```

5. **Run the portal** (`cd web && pnpm dev`) — attendees sign in, land on
   `/no-access?event=PEARC26`, request access; approve from
   Site administration > Access Requests. Approval provisions COmanage
   (person + verified email + login identifier), LDAP/SSSD, and the SLURM
   user association; the attendee can then `ssh nexus-<user>@<cluster>`
   via CILogon device auth and submit jobs.

To reset one attendee (registry person, custos rows, cluster cache):
`scripts/reset-trial-test-user.sh <email> --yes`.

## Files

| File | Creates |
|---|---|
| 01_org_cluster.sql | PEARC26 Attendees org; the `nexus-dev` cluster row |
| 02_resources_rates.sql | `debug` partition catalog entry + SU rate |
| 03_admin_role.sql | approver/bootstrap-admin user + `access-approver` role |
| setup-live.sh | project (PI = seed-03 admin), `pearc26-tutorial` allocation (25000 SU), resource grant, PEARC26 access-event row |

## Notes

- **The cluster row reuses id `00000000-0000-0000-0000-000000000001`
  (= `CUSTOS_CLUSTER_ID`).** The COmanage subscriber only provisions
  cluster users whose `compute_cluster_id` equals that env value; a
  different id means approvals silently never reach COmanage. The row
  upserts the name to `nexus-dev` if `default_cluster.sql` got there
  first.
- The seed-03 user is approver, bootstrap admin, AND project PI: set
  `CUSTOS_BOOTSTRAP_ADMIN_EMAIL=approver@pearc26.local` (or edit that
  email to your own) and boot grants super_admin idempotently.
- **Never create the allocation with SQL.** Seeded rows bypass the event
  bus, so the SLURM account/grant never appear (this bit us once). If an
  allocation must be replaced, tear it down and re-run `setup-live.sh`.
- `setup-live.sh` refuses to run when an ACTIVE `pearc26-tutorial`
  allocation already exists.
- SQL seeds are `INSERT IGNORE` / upsert — safe to re-apply.
- The SLURM API version matters: `SLURM_API_VERSION=40`. On v0.0.38 the
  associations endpoint silently ignores creates (200, empty errors,
  nothing written).
