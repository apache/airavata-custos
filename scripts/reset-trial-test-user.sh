#!/usr/bin/env bash
# Wipe one person from the trial-allocation pipeline so their email can be
# reused as a fresh test user: deletes the registry CoPerson (which pulls the
# LDAP entry), removes every custos dev-DB row for the email, and checks the
# cluster no longer resolves the account.
#
# Usage:
#   scripts/reset-trial-test-user.sh [email] [--yes]
# Default email: lahirujayathilake@gmail.com. --yes skips the confirm prompt.
#
# Requires: repo-root .env with COMANAGE_*, docker (custos_db), ssh access to
# the cluster as exouser. Registry + DB deletes are DESTRUCTIVE.

set -euo pipefail
cd "$(dirname "$0")/.."

EMAIL="${1:-lahirujayathilake@gmail.com}"
CONFIRM="${2:-}"
CLUSTER="exouser@149.165.159.51"
# DB credentials come from DATABASE_DSN when present (deployment hosts
# differ from the local compose defaults).
DB_USER=admin; DB_PASS=admin
if [ -f .env ] && grep -q "^DATABASE_DSN=" .env; then
  _dsn=$(grep "^DATABASE_DSN=" .env | cut -d= -f2- | tr -d '"' | tr -d "'")
  DB_USER=$(echo "$_dsn" | cut -d: -f1)
  DB_PASS=$(echo "$_dsn" | sed -E 's/^[^:]*:([^@]*)@.*/\1/')
fi
DB_EXEC=(docker exec -i custos_db mariadb -u"$DB_USER" -p"$DB_PASS" custos)

set -a; source .env; set +a
: "${COMANAGE_REGISTRY_URL:?missing in .env}" "${COMANAGE_CO_ID:?}" "${COMANAGE_API_USER:?}" "${COMANAGE_API_KEY:?}"

creg() { curl -sf -u "$COMANAGE_API_USER:$COMANAGE_API_KEY" "$@"; }

# free_identifiers frees the login/uid identifiers on a person and its linked
# org identities. Those are unique registry-wide and survive a soft delete,
# so a reused sub/username 403s re-provisioning until they are removed.
free_identifiers() {
  local pid="$1"
  for oid in $(creg "$COMANAGE_REGISTRY_URL/co_org_identity_links.json?copersonid=$pid" \
      | python3 -c "import json,sys; print(' '.join(str(l['OrgIdentityId']) for l in json.load(sys.stdin).get('CoOrgIdentityLinks',[])))" 2>/dev/null); do
    for iid in $(creg "$COMANAGE_REGISTRY_URL/identifiers.json?orgidentityid=$oid" \
        | python3 -c "import json,sys; print(' '.join(str(i['Id']) for i in json.load(sys.stdin).get('Identifiers',[])))" 2>/dev/null); do
      curl -sf -o /dev/null -X DELETE -u "$COMANAGE_API_USER:$COMANAGE_API_KEY" \
        "$COMANAGE_REGISTRY_URL/identifiers/$iid.json" \
        && echo "  freed org identity $oid identifier $iid" \
        || echo "  could not delete identifier $iid on org identity $oid"
    done
  done
  for iid in $(creg "$COMANAGE_REGISTRY_URL/identifiers.json?copersonid=$pid" \
      | python3 -c "import json,sys
free={'oidcsub','sorid','eppn','uid'}
print(' '.join(str(i['Id']) for i in json.load(sys.stdin).get('Identifiers',[]) if i['Type'] in free))" 2>/dev/null); do
    curl -sf -o /dev/null -X DELETE -u "$COMANAGE_API_USER:$COMANAGE_API_KEY" \
      "$COMANAGE_REGISTRY_URL/identifiers/$iid.json" \
      && echo "  freed person $pid identifier $iid" \
      || echo "  could not delete person identifier $iid"
  done
}

echo "== Looking up '$EMAIL' in the registry"
people_json=$(creg "$COMANAGE_REGISTRY_URL/co_people.json?coid=$COMANAGE_CO_ID&search.mail=$EMAIL" || true)
person_ids=$(echo "$people_json" | python3 -c "import json,sys; print(' '.join(str(p['Id']) for p in json.load(sys.stdin).get('CoPeople',[])))" 2>/dev/null || true)

# Soft-deleted matches keep their stale login identifiers; free them now
# regardless of the exact-email gate (they are already dead, so it is safe).
for pid in $(echo "$people_json" | python3 -c "import json,sys; print(' '.join(str(p['Id']) for p in json.load(sys.stdin).get('CoPeople',[]) if p.get('Status')=='Deleted'))" 2>/dev/null); do
  echo "  freeing identifiers on soft-deleted CoPerson $pid"
  free_identifiers "$pid"
done

declare -a targets=()
registry_uids=""
for pid in $person_ids; do
  read -r ident reg_uid < <(creg "$COMANAGE_REGISTRY_URL/identifiers.json?copersonid=$pid" \
    | python3 -c "import json,sys,os
ids=json.load(sys.stdin).get('Identifiers',[])
t=os.environ.get('COMANAGE_PERSON_ID_TYPE','')
print(next((i['Identifier'] for i in ids if i['Type']==t),'-'),
      next((i['Identifier'] for i in ids if i['Type']=='uid'),'-'))")
  [ "$ident" = "-" ] && ident=""
  [ "$reg_uid" != "-" ] && registry_uids="$registry_uids $reg_uid"
  [ -z "$ident" ] && { echo "  CoPerson $pid: no $COMANAGE_PERSON_ID_TYPE identifier, skipping"; continue; }
  # exact-match the email before deleting anything (search.mail is a LIKE)
  exact=$(creg "$COMANAGE_REGISTRY_URL/api/co/$COMANAGE_CO_ID/core/v1/people/$ident" \
    | python3 -c "import json,sys
d=json.load(sys.stdin)
mails=[e.get('mail','').lower() for e in d.get('EmailAddress',[])]
print('yes' if '$EMAIL'.lower() in mails else 'no')")
  if [ "$exact" = "yes" ]; then
    targets+=("$pid:$ident")
    echo "  CoPerson $pid ($ident): exact email match"
  else
    echo "  CoPerson $pid ($ident): LIKE match only, skipping"
  fi
done

echo "== custos dev DB rows for '$EMAIL'"
user_id=$("${DB_EXEC[@]}" -N -e "SELECT id FROM users WHERE email='$EMAIL'" 2>/dev/null || true)
local_usernames=""
if [ -n "$user_id" ]; then
  local_usernames=$("${DB_EXEC[@]}" -N -e "SELECT local_username FROM compute_cluster_users WHERE user_id='$user_id'" 2>/dev/null || true)
  echo "  user: $user_id  cluster usernames: ${local_usernames:-none}"
else
  echo "  no users row"
fi
req_count=$("${DB_EXEC[@]}" -N -e "SELECT COUNT(*) FROM access_requests WHERE email='$EMAIL'" 2>/dev/null || echo 0)
echo "  access requests: $req_count"

if [ ${#targets[@]} -eq 0 ] && [ -z "$user_id" ] && [ "$req_count" = "0" ]; then
  echo "Nothing to delete."; exit 0
fi

if [ "$CONFIRM" != "--yes" ]; then
  read -r -p "Delete all of the above? [y/N] " ans
  [ "$ans" = "y" ] || { echo "Aborted."; exit 1; }
fi

for t in "${targets[@]}"; do
  pid="${t%%:*}"; ident="${t##*:}"
  free_identifiers "$pid"
  echo "== Deleting registry CoPerson $pid ($ident)"
  curl -sf -o /dev/null -X DELETE -u "$COMANAGE_API_USER:$COMANAGE_API_KEY" \
    "$COMANAGE_REGISTRY_URL/api/co/$COMANAGE_CO_ID/core/v1/people/$ident" \
    && echo "  deleted (registry keeps a soft-deleted stub; Expunge in the UI removes it fully - harmless either way, the provisioner skips non-Active people)" || echo "  DELETE failed for $ident (check manually)"
done

echo "== Cleaning custos dev DB"
if [ -n "$user_id" ]; then
  "${DB_EXEC[@]}" <<SQL
DELETE FROM compute_allocation_memberships WHERE user_id='$user_id';
DELETE FROM compute_cluster_users WHERE user_id='$user_id';
DELETE FROM user_identities WHERE user_id='$user_id';
DELETE FROM user_roles WHERE user_id='$user_id';
DELETE FROM user_privileges WHERE user_id='$user_id';
DELETE FROM project_memberships WHERE user_id='$user_id';
UPDATE access_requests SET created_user_id=NULL, approver_id=NULL WHERE created_user_id='$user_id' OR approver_id='$user_id';
DELETE FROM users WHERE id='$user_id';
SQL
fi
"${DB_EXEC[@]}" -e "DELETE FROM access_requests WHERE email='$EMAIL';" 2>/dev/null || true
echo "  done"

echo "== Slurm accounting cleanup"
# slurmctld binds an association to the uid at load time; a recycled
# username returns with a new uid, so stale user associations must go.
if [ -n "${SLURM_API_URL:-}" ] && [ -n "${SLURM_TOKEN:-}" ]; then
  ver="v0.0.${SLURM_API_VERSION:-40}"
  for u in $(printf '%s\n' $local_usernames $registry_uids | sort -u); do
    # Delete the association (not the user record): the /user endpoint
    # 304s without removing the association, leaving a stale uid binding.
    removed=$(curl -s -X DELETE \
      -H "X-SLURM-USER-NAME: ${SLURM_API_USERNAME:-nexus-provisioner}" -H "X-SLURM-USER-TOKEN: $SLURM_TOKEN" \
      "$SLURM_API_URL/slurmdb/$ver/associations?user=$u" \
      | python3 -c "import json,sys; print(len(json.load(sys.stdin).get('removed_associations',[])))" 2>/dev/null || echo 0)
    echo "  slurm associations removed for $u: $removed"
  done
else
  echo "  SLURM_API_URL/SLURM_TOKEN not set; skipped (delete stale associations manually if the username is reused)"
fi

echo "== Cluster checks (read-only + cache flush attempt)"
base="ou=people,o=Nexus,o=CO,dc=nexus,dc=cybershuttle,dc=org"
ldap_left=$(ssh -o ConnectTimeout=10 -o BatchMode=yes "$CLUSTER" \
  "ldapsearch -x -H ldaps://ldap-test.nexus.cybershuttle.org -b '$base' '(mail=$EMAIL)' dn 2>/dev/null | grep -c '^dn:' || true" 2>/dev/null || echo "?")
echo "  LDAP entries remaining for $EMAIL: $ldap_left (registry->LDAP delete can lag a minute)"
for u in $(printf '%s\n' $local_usernames $registry_uids | sort -u); do
  ssh -o ConnectTimeout=10 -o BatchMode=yes "$CLUSTER" "sudo -n sss_cache -u '$u' 2>/dev/null; sudo -n sss_cache -E 2>/dev/null" 2>/dev/null \
    && echo "  sss_cache flushed for $u (full cache invalidated)" \
    || echo "  could not flush sss cache for $u (entry may linger up to the SSSD cache TTL, ~90 min)"
  r=$(ssh -o ConnectTimeout=10 -o BatchMode=yes "$CLUSTER" "id '$u' 2>&1" 2>/dev/null || true)
  echo "  id $u -> $r"
done
for u in $(printf '%s\n' $local_usernames $registry_uids | sort -u); do
  ssh -o ConnectTimeout=10 -o BatchMode=yes "$CLUSTER" "test -d '/home/$u' && ls -ldn '/home/$u'" 2>/dev/null \
    && echo "  NOTE: /home/$u still exists with the OLD uid; a re-provisioned account gets a new uid and cannot use it. Fix on the cluster: sudo rm -rf '/home/$u' (or chown to the new uid after re-provisioning)."
done
echo "Done. Re-run the checks in a few minutes if LDAP had not caught up yet."
