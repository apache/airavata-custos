<!--
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
-->

# HPC Node Enrollment — Account Provisioning

Ansible playbook to enroll RHEL 9/10 nodes for COmanage-provisioned LDAP user resolution and CILogon device-flow SSH authentication.

## Architecture

This setup uses two independent systems that each answer one question:

| System | Question | Technology |
|--------|----------|------------|
| **SSSD/LDAP** | "Does this Unix account exist?" | SSSD → COmanage-provisioned LDAP (posixAccount) |
| **OIDC** | "Is this the right human?" | pam_oauth2_device → CILogon device authorization flow |

**Key design decisions:**

- **Sub→uid mapping via LDAP, not PAM config.** The SSH username is always a Unix username. The PAM module maps the CILogon `sub` to a local `uid` via an LDAP lookup (`voPersonExternalID` → `uid`) to verify the authenticated identity matches the login user. No static mapping or PAM-level account resolution is needed.
- **Authentication then authorization.** The device flow always starts. After the user authenticates via CILogon, the PAM module's LDAP lookup verifies the `sub` maps to the SSH login username. If the user does not exist in LDAP, this lookup fails and login is denied.
- **Revocation via COmanage.** When COmanage sets a person inactive, the LDAP entry is removed, SSSD stops resolving the user, and SSH login fails automatically. No SSH config changes needed.

```
SSH login flow:
  1. ssh jdoe@node
  2. PAM auth → pam_oauth2_device → CILogon device flow
  3. User completes browser auth
  4. PAM module queries LDAP: voPersonExternalID=<sub> → uid
  5. Returned uid matches "jdoe"? No → reject. Yes → continue.
  6. SSSD resolves "jdoe" → LDAP (posixAccount lookup)
  7. pam_mkhomedir creates /home/jdoe if first login
```

## Prerequisites

1. **LDAP server** with COmanage-provisioned entries having both `posixAccount` objectClass and `voPersonExternalID` attribute (from [voPerson schema](https://refeds.org/specifications/voperson)). The PAM module maps CILogon subjects to local usernames via the LDAP filter `(&(objectClass=posixAccount)(voPersonExternalID=%s))`. See [LDAP Schema Requirements](#ldap-schema-requirements) for details on required schemas.
2. **CILogon OIDC client** — must be a **confidential client** (with a client secret), not a public client.
3. **Ansible** 2.14+ on your control node.
4. **RHEL 9/10** target nodes with SSH and sudo access.
5. Target nodes must be able to reach the LDAP server (port 636) and CILogon (port 443).

### LDAP Schema Requirements

If you're setting up a fresh LDAP server for COmanage, these schemas must be imported. COmanage's LDAP provisioning plugin writes entries using these objectClasses.

| Schema | Type | Purpose |
|--------|------|---------|
| **voPerson** | AUXILIARY | COmanage person attributes. Critical: `voPersonExternalID` stores the CILogon subject URI used for OIDC sub→Unix uid mapping |
| **eduMember** | AUXILIARY | Group membership (`isMemberOf`, `hasMember`) |
| **posixAccount** | STRUCTURAL | Unix account attributes (`uidNumber`, `gidNumber`, `homeDirectory`, `loginShell`). Required for SSSD |
| **posixGroup** | STRUCTURAL | Unix group attributes |
| **inetOrgPerson** | STRUCTURAL | Base person object (`cn`, `sn`, `mail`, `uid`) |

## Quick Start

```bash
cd deployment/account-provisioning

# 1. Configure
cp inventory/hosts.example.yml inventory/hosts.yml
cp group_vars/all.yml.example group_vars/all.yml
# Edit both files with your values

# 2. Encrypt secrets
ansible-vault encrypt_string 'your-ldap-password' --name 'vault_ldap_bind_password' >> group_vars/all.yml
ansible-vault encrypt_string 'your-cilogon-secret' --name 'vault_cilogon_client_secret' >> group_vars/all.yml

# 3. Test LDAP connectivity first
ansible-playbook -i inventory/hosts.yml enroll-node.yml --tags prereqs --ask-vault-pass

# 4. Full enrollment
ansible-playbook -i inventory/hosts.yml enroll-node.yml --ask-vault-pass

# 5. Verify
ansible-playbook -i inventory/hosts.yml verify.yml --ask-vault-pass
```

All variables are documented with inline comments in `group_vars/all.yml.example`.

## What the Playbook Does

The playbook has 4 tagged phases that can be run independently with `--tags`:

| Tag | Phase | Key detail |
|-----|-------|------------|
| `prereqs` | Read-only LDAP connectivity check | Fails fast if LDAP is unreachable. No files modified. |
| `sssd` | SSSD + NSS + mkhomedir | Configures `auth_provider = none` — PAM handles authentication, SSSD only resolves identity. Sets SELinux boolean `authlogin_nsswitch_use_ldap`. |
| `pam` | Build pam_oauth2_device, configure PAM stack, SELinux policy | Builds from source (cyber-shuttle fork). Installs SELinux policy allowing `sshd_t` to connect to CILogon on port 443 — without it, auth silently falls through to password-auth. |
| `sshd` | Keyboard-interactive auth config | Drops `99-pam-oauth2-device.conf` into `sshd_config.d/` — the `99-` prefix ensures it overrides RHEL defaults. |

## Verification

```bash
# End-to-end SSH tests (verify.yml checks services/config, these test actual login):
ssh jdoe@<node>                # should show device code prompt
ssh fakeuser@<node>            # device flow starts, but login fails after auth (LDAP lookup finds no match)

# Check effective sshd config:
sudo sshd -T | egrep -i 'usepam|kbdinteractiveauthentication|passwordauthentication|challengeresponseauthentication'
# Expected: all "yes"
```

## Troubleshooting

### SSSD won't start
- Check permissions: `ls -la /etc/sssd/sssd.conf` — must be `0600` owned by root
- Check logs: `journalctl -u sssd --since "5 minutes ago"`
- Clear cache and restart: `rm -rf /var/lib/sss/db/* && systemctl restart sssd`

### `getent passwd <user>` returns nothing
- Verify LDAP connectivity: re-run `--tags prereqs`
- Enable SSSD debug: add `debug_level = 6` under `[domain/...]` in sssd.conf, restart, check `/var/log/sssd/sssd_<domain>.log`

### SSH shows no device code prompt
- Check effective config: `sudo sshd -T | grep kbdinteractive` — if it shows `no`, another file in `/etc/ssh/sshd_config.d/` is overriding it. The `99-` prefix on our config ensures it loads last; verify the file exists and restart sshd.

### SELinux blocking CILogon requests
The PAM module runs inside sshd and makes HTTPS calls to CILogon. SELinux's `sshd_t` context blocks this by default.

**Symptoms:** `journalctl -u sshd` shows `curl failed HTTP POST` or `Authentication failure`, and `sealert -a /var/log/audit/audit.log` shows `SELinux is preventing /usr/sbin/sshd from name_connect access on the tcp_socket port 443`.

**Fix:** Re-run `--tags pam` to reinstall the SELinux policy. For manual diagnosis:
```bash
sudo dnf install -y setroubleshoot-server
sudo sealert -a /var/log/audit/audit.log
```

### CILogon auth fails with missing claims
Your CILogon client is likely a **public** client. See [Prerequisites](#prerequisites) — a confidential client with a client secret is required.

## Operational Notes

- **LDAP outage:** SSSD caches credentials (`cache_credentials = true`), so previously-resolved users can still log in when LDAP is temporarily unreachable.
- **PAM config:** The playbook replaces `/etc/pam.d/sshd` entirely. Back up the existing file before running on nodes with custom PAM configurations.
- **pam_oauth2_device version:** The playbook builds from the latest commit of the configured git repo. Pin to a specific tag by adding `version:` to the git task in `enroll-node.yml` for production stability.
