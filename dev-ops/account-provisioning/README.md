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

# HPC node enrollment

This Ansible playbook enrolls RHEL 9 and 10 login nodes. Enrolled nodes let users provisioned through COmanage SSH in
with their CILogon identity. It is for site administrators. Run it on each login node.

It does two things:

- Configures SSSD to resolve accounts from the COmanage-provisioned LDAP.
- Builds and installs the [CILogon SSH PAM module](../../extensions/CILogon-SSH-PAM/README.md), which
  authenticates users through the OAuth 2.0 device flow.

Password and SSH key logins keep working: see [Other login methods](#other-login-methods).

## How it works

Two independent systems each answer one question:

| System | Question | Technology |
|--------|----------|------------|
| **SSSD/LDAP** | "Does this Unix account exist?" | SSSD → COmanage-provisioned LDAP (posixAccount) |
| **OIDC** | "Is this the right human?" | pam_oauth2_device → CILogon device authorization flow |

```
SSH login flow:
  1. ssh jdoe@node
  2. sshd resolves "jdoe" through SSSD (posixAccount lookup)
  3. PAM auth → pam_oauth2_device → CILogon device flow
  4. User completes browser auth
  5. PAM module queries LDAP: voPersonExternalID=<CILogon sub> → uid
  6. Returned uid matches "jdoe"? No → reject. Yes → continue.
  7. pam_oddjob_mkhomedir creates the home directory on first login
```

The SSH username is the Unix username. The node needs no static mapping.

Deleting the user's LDAP entry ends device-flow logins. Other methods: [revocation](../../README.md#problems-it-addresses).

### Other login methods

The playbook adds the device flow without removing other methods:

| Method | After enrollment |
|---|---|
| Device flow | `pam_oauth2_device` is `sufficient` in `/etc/pam.d/sshd`, so a completed device flow logs the user in. |
| Password | The standard `password-auth` stack follows the module. The sshd override sets `PasswordAuthentication yes`. It works only for accounts with a local password, because SSSD does not authenticate. |
| SSH key | sshd handles it. The playbook does not change it. |

To allow only the device flow, disable the other methods in the site's own sshd and PAM configuration.

## Prerequisites

1. **LDAP server** with COmanage-provisioned entries having both `posixAccount` objectClass and `voPersonExternalID` attribute (from [voPerson schema](https://refeds.org/specifications/voperson)). The PAM module maps CILogon subjects to local usernames via the LDAP filter `(&(objectClass=posixAccount)(voPersonExternalID=%s))`. See the schema table below for details on required schemas.
2. **CILogon OIDC client** — must be a **confidential client** (with a client secret), not a public client.
3. **Ansible** 2.14+ on your control node.
4. **RHEL 9/10** target nodes with SSH and sudo access.
5. Target nodes must be able to reach the LDAP server (port 636) and CILogon (port 443).

If you're setting up a fresh LDAP server for COmanage, these schemas must be imported. COmanage's LDAP provisioning plugin writes entries using these objectClasses.

| Schema | Type | Purpose |
|--------|------|---------|
| **voPerson** | AUXILIARY | COmanage person attributes. Critical: `voPersonExternalID` stores the CILogon subject URI used for OIDC sub→Unix uid mapping |
| **eduMember** | AUXILIARY | Group membership (`isMemberOf`, `hasMember`) |
| **posixAccount** | AUXILIARY | Unix account attributes (`uidNumber`, `gidNumber`, `homeDirectory`, `loginShell`). Required for SSSD |
| **posixGroup** | STRUCTURAL | Unix group attributes |
| **inetOrgPerson** | STRUCTURAL | Base person object (`cn`, `sn`, `mail`, `uid`) |

## Setup

```bash
cd dev-ops/account-provisioning

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

All variables are documented with inline comments in `group_vars/all.yml.example`. Set `test_username` to a user in LDAP, which the `sssd` phase and
`verify.yml` resolve. `pam_prompt_hostname` and `pam_logo_file_src` set the system name and logo in the sign-in prompt.

## What the playbook does

The playbook has 4 tagged phases that can be run independently with `--tags`:

| Tag | Phase | Key detail |
|-----|-------|------------|
| `prereqs` | LDAP check | Installs `openldap-clients`, then fails fast if LDAP is unreachable or the bind fails. Changes no configuration. |
| `sssd` | SSSD + NSS + mkhomedir | Configures `auth_provider = none` — PAM handles authentication, SSSD only resolves identity. Sets SELinux boolean `authlogin_nsswitch_use_ldap`. |
| `pam` | Module, PAM stack and SELinux policy | Builds `extensions/CILogon-SSH-PAM` from `pam_oauth2_device_repo` at `pam_oauth2_device_version`. The SELinux policy lets sshd reach CILogon on port 443 and LDAP on port 636. Without it, authentication falls through to `password-auth` silently. |
| `sshd` | Keyboard-interactive authentication | Writes `99-pam-oauth2-device.conf` to `sshd_config.d/`. Its settings sit in a `Match all` block, which overrides global values set earlier, such as RHEL's `50-redhat.conf`. |

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

**SSSD won't start**
- Check permissions: `ls -la /etc/sssd/sssd.conf` — must be `0600` owned by root
- Check logs: `journalctl -u sssd --since "5 minutes ago"`
- Clear cache and restart: `rm -rf /var/lib/sss/db/* && systemctl restart sssd`

**`getent passwd <user>` returns nothing**
- Verify LDAP connectivity: re-run `--tags prereqs`
- Enable SSSD debug: add `debug_level = 6` under `[domain/...]` in sssd.conf, restart, check `/var/log/sssd/sssd_<domain>.log`

**SSH shows no device code prompt**
- Check effective config: `sudo sshd -T | grep kbdinteractive` — if it shows `no`, another file in `/etc/ssh/sshd_config.d/` is overriding it. Verify `99-pam-oauth2-device.conf` exists and restart sshd.

**SELinux blocking CILogon requests**

The PAM module runs inside sshd and makes HTTPS calls to CILogon. SELinux's `sshd_t` context blocks this by default.

**Symptoms:** `journalctl -u sshd` shows `curl failed HTTP POST` or `Authentication failure`, and `sealert -a /var/log/audit/audit.log` shows `SELinux is preventing /usr/sbin/sshd from name_connect access on the tcp_socket port 443`.

**Fix:** Re-run `--tags pam` to reinstall the SELinux policy. For manual diagnosis:
```bash
sudo dnf install -y setroubleshoot-server
sudo sealert -a /var/log/audit/audit.log
```

**CILogon auth fails with missing claims**

Your CILogon client is likely a **public** client. See [Prerequisites](#prerequisites) — a confidential client with a client secret is required.

## Operational notes

- **LDAP outage.** The device flow queries LDAP at each login, so it fails while LDAP is unreachable. SSSD serves
  cached entries while offline, so SSH key logins of users it has resolved before still work.
- **PAM config:** The playbook replaces `/etc/pam.d/sshd` entirely. Back up the existing file before running on nodes with custom PAM configurations.
- **Module version.** The `pam` phase builds `pam_oauth2_device_version`, `master` by default. Pin it to a tag or commit
  for production.
