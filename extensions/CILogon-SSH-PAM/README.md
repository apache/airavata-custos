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

# CILogon SSH PAM module

`pam_oauth2_device` lets users log in over SSH with their OpenID Connect identity.
This README is for administrators of cluster login nodes.
[HOWTO.md](HOWTO.md) covers bypass and other authentication methods.

Custos sites use CILogon as the provider.
The COmanage connector provisions an LDAP entry per user.
The entry maps the CILogon subject (`sub`) to the Unix account.

| Property | Detail |
|---|---|
| No cluster password or SSH key needed | The user authenticates to CILogon. |
| Access ends with the LDAP entry | The module queries LDAP at each login. See [revocation](../../README.md#problems-it-addresses) for other methods. |

On RHEL 9 and 10, the [node enrollment playbook](../../dev-ops/account-provisioning/README.md) installs the module and SSSD.
It [keeps other login methods](../../dev-ops/account-provisioning/README.md#other-login-methods).

## How it works

The module uses the OAuth 2.0 device authorization flow.

1. The user runs `ssh jdoe@login.example.org`. PAM passes the local username, `jdoe`, to the module.
2. If the user is [bypassed](HOWTO.md#bypass), the module returns `PAM_IGNORE` and PAM moves on to the next module.
3. The module requests a device code from `device_endpoint` and shows the user a link, and optionally a QR code.
4. The user opens the link and signs in with the provider.
5. The module polls `token_endpoint` until the sign-in completes, then reads the user's claims from `userinfo_endpoint`.
6. It checks whether the identity may use the local account. If so, it returns `PAM_SUCCESS`. Otherwise login fails.

The userinfo response must contain `sub`, `name` and the attribute named by `username_attribute`. A `groups` claim, if
present, feeds the group checks below.

Each user has two names, which can differ:

| Name | Meaning |
|---|---|
| Local username | The name in the login, `fred` in `ssh fred@example.com`. PAM passes it to the module. |
| Remote username | The userinfo attribute chosen by `username_attribute`. Custos sites use `sub`. |

The user lands in the local username's account.

### Authorisation

Custos uses only the LDAP check, number 4.

The module tries these checks in order and allows the login at the first that passes. A section that is absent or
disabled is skipped. If none passes, login is denied.

| # | Section | Passes when |
|---|---|---|
| 1 | `cloud`, with `access: true` | `cloud.username` equals the local username plus `local_username_suffix`, and one of the user's groups is in the list the cloud endpoint returns for this host's project |
| 2 | `group`, with `access: true` | The remote username equals the local username plus `local_username_suffix`, and `group.service_name` is one of the user's groups |
| 3 | `users` | The map lists the local username under the user's remote username. No suffix is used. |
| 4 | `ldap` | The `filter`, with `%s` replaced by the remote username, finds an entry whose `attr` equals the local username. No suffix is used. |

With the filter in the example below and `attr` set to `uid`, the module allows the login when the entry's `uid` is the
local username.

For the `cloud` check, the module reads `project_id` from `cloud.metadata_file`, a JSON file such as
`{"project_id": "fleeps"}`. Without `metadata_file`, it reads the legacy path
`/mnt/context/openstack/latest/meta_data.json`. It then calls `<endpoint>/<project_id>` without client authentication
and expects `{"groups": ["wop", "fap", "users"]}`.

## Build and install

On RHEL, install the build dependencies:

```bash
dnf install -y "@Development Tools" pam-devel libcurl-devel openldap-devel pkgconfig
```

Then build from this directory and install:

```bash
make
install -m 0755 pam_oauth2_device.so /usr/lib64/security/
install -m 0600 -D config_template.json /etc/pam_oauth2_device/config.json
```

On Debian-based systems, install the equivalent development packages and copy the module to
`/lib/x86_64-linux-gnu/security`.
`make install` puts the module in `$(DESTDIR)$(PREFIX)/lib/security`.
It puts the config in `$(DESTDIR)$(PREFIX)/etc/pam_oauth2_device`.

- `make test` runs the unit tests: [test/README.md](test/README.md).
- RPM and DEB builds package upstream releases, not this tree: [packaging/README.md](packaging/README.md).

## Configuration

The module reads the fixed path `/etc/pam_oauth2_device/config.json` on each login.
It ignores PAM-line arguments except `debug`, which must come second or later:
`pam_oauth2_device.so /etc/pam_oauth2_device/config.json debug`.

It rejects invalid JSON only at login. Run `jq . /etc/pam_oauth2_device/config.json` after each edit.

A Custos site's configuration, as the playbook writes it:

```json
{
  "oauth": {
    "client": { "id": "cilogon:/client_id/...", "secret": "..." },
    "device_endpoint":   "https://cilogon.org/oauth2/device_authorization",
    "token_endpoint":    "https://cilogon.org/oauth2/token",
    "userinfo_endpoint": "https://cilogon.org/oauth2/userinfo",
    "scope": "openid profile email org.cilogon.userinfo",
    "username_attribute": "sub",
    "local_username_suffix": ""
  },
  "ldap": {
    "host": "ldaps://ldap.example.org:636",
    "basedn": "ou=people,o=...,o=CO,dc=...",
    "user": "uid=registry_user,ou=system,o=...,o=CO,dc=...",
    "passwd": "...",
    "filter": "(&(objectClass=posixAccount)(voPersonExternalID=%s))",
    "attr": "uid"
  },
  "qr": { "error_correction_level": 0 },
  "prompt": { "hostname": "login.cluster.example.org" },
  "tls": { "ca_bundle": "/etc/pki/tls/certs/ca-bundle.crt" }
}
```

[`config_template.json`](config_template.json) shows all sections. The module ignores keys it does not know.

### Splitting the file

Wherever the module expects a JSON object, a full file path can stand in for it. Split files can split further:

```json
{
  "oauth": {
    "client": "/etc/pam_oauth2_device/client.json",
    "scope": "openid profile"
  }
}
```

with `/etc/pam_oauth2_device/client.json` holding:

```json
{ "id": "abcd", "secret": "meetmeinstlouis" }
```

A string value, such as `scope`, cannot be split off.

### Authentication options

| Section | Entry | Type | Required | Default | Description |
|---|---|---|---|---|---|
| `oauth` | `client` | Object | yes | | `id` and `secret` of a confidential client |
| `oauth` | `scope` | String | yes | | Space-separated OIDC scopes, sent as given |
| `oauth` | `device_endpoint` | String | yes | | Device authorization endpoint |
| `oauth` | `token_endpoint` | String | yes | | Token endpoint |
| `oauth` | `userinfo_endpoint` | String | yes | | Userinfo endpoint |
| `oauth` | `username_attribute` | String | yes | | Userinfo attribute used as the remote username |
| `oauth` | `local_username_suffix` | String | no | `""` | Suffix for the `cloud` and `group` checks |
| `tls` | `ca_bundle` | String | no | | PEM file of concatenated trust anchors. Takes precedence over `ca_path`. |
| `tls` | `ca_path` | String | no | `/etc/grid-security/certificates` | Directory of trust anchors. Works only when curl uses OpenSSL (`curl -V`). |
| `qr` | `error_correction_level` | Int | no | `-1` | QR code: `0` low, `1` medium, `2` high. A negative value, or no `qr` section, turns it off. |
| `prompt` | `hostname` | String | no | `custos-cluster` | System name shown in the sign-in prompt |
| `prompt` | `logo_file` | String | no | | Plain-text file shown in place of the built-in logo art. An unreadable file falls back to the art. |
| (top level) | `client_debug` | Bool | no | `false` | Debug logging |
| (top level) | `http_basic_auth` | Bool | no | `true` | Send client credentials with HTTP Basic. `false` sends them as request parameters (RFC 6749, section 2.3.1). |

The provider's certificate must validate against the trust anchors.
The prompt is plain text, because the SSH client escapes ANSI codes.

### Authorisation options

| Section | Entry | Type | Required in section | Description |
|---|---|---|---|---|
| `ldap` | `host` | String | yes | LDAP URL |
| `ldap` | `basedn` | String | yes | Base DN, least significant RDN first |
| `ldap` | `user`, `passwd` | String | yes | Bind credentials. Empty strings give an anonymous bind. |
| `ldap` | `scope` | String | no | `sub`/`subtree` (default), `one`/`onelevel`, `base`/`baseobject`, or `subordinate`/`children` where the server supports it |
| `ldap` | `filter` | String | yes | Search filter. `%s` becomes the remote username. |
| `ldap` | `attr` | String | yes | Attribute compared with the local username |
| `ldap` | `preauth` | String | no | Bypass query. See [HOWTO.md](HOWTO.md). |
| `group` | `access` | Bool | yes | Turns the check on |
| `group` | `service_name` | String | yes | Group the user must belong to |
| `cloud` | `access` | Bool | yes | Turns the check on |
| `cloud` | `endpoint`, `username` | String | yes | Group-list server and expected local username |
| `cloud` | `metadata_file` | String | no | Project metadata file. Default: the legacy path under [Authorisation](#authorisation). |
| `users` | | Object | | Maps a remote username to a list of local usernames. The key `*bypass*` lists users who skip the module. |

## SSH configuration

You MUST edit the configuration before this module will work!

Make sure the module works correctly before changing your SSH config or you may be locked out!  See Testing below.

Edit `/etc/pam.d/sshd` and comment out the other `auth` sections (unless you need MFA or something else).

```
auth required pam_oauth2_device.so
```

Edit `/etc/ssh/sshd_config` and make sure that the following configuration options are set

```
ChallengeResponseAuthentication yes
KbdInteractiveAuthentication yes
UsePAM yes
```

```
systemctl restart sshd
```

## Testing

You are advised to do this before making changes to your main SSH config.  There are two tests to do which are recommended to do in the order described here.

It is recommended that you create a (hard or sym) link to your `sshd` called (for example) `pamsshd`, e.g. `/usr/local/sbin/pamsshd`.  This means you can have a PAM configuration for `pamsshd` which is different from the normal `sshd`.

In this case you can copy `/etc/pam.d/sshd` to `/etc/pam.d/pamsshd` and edit the latter, leaving the former to log you back into the system if something goes wrong.  Also copy `/etc/ssh/sshd_config` to `/etc/ssh/pamsshd_config` so you can edit the configuration for `pamsshd` separately.

Note that testing *requires* that you install the module in the system location and you have the configuration set up in `/etc/pam_oauth2_device/config.json` and `/etc/pam.d/` and `/etc/ssh/`.

Test 1: install `pamtester`.

Run
```
pamtester -v pamsshd localusername authenticate
```
and follow the onscreen prompts.  Here, `localusername` refers to your local user name so replace it with whatever your name is.

You can check `/var/log/secure` or `/var/log/auth.log` to find what's wrong if there are errors authenticating.  While the module
uses syslog, syslog is normally set up to log PAM stuff into one of these files.

Test 2: while pamtester tests the authenticate section, you should try a proper ssh login from another host.  If you created `pamsshd` as above (and copied configuration as described above), start it with

```
/usr/local/sbin/pamsshd -f /etc/ssh/pamsshd_config -p 2222 -d
```

This should start `sshd` with the name `pamsshd` listening on port 2222.  Now try to log in from another host (bearing in mind the port should be open for incoming tcp).  On the other host, run `ssh -p 2222 localusername@myhost` where `localusername` is the local user name and `myhost` is the host running `pamsshd`.

Again check the logs as in the previous tests.

## Troubleshooting

| Symptom | Check |
|---|---|
| Login fails with no clear reason | The module logs through syslog, usually to `/var/log/secure` or `/var/log/auth.log`. For more detail, set `client_debug: true`, or add `debug` as the second or a later argument on the PAM line. |
| `userinfo lacks 'sub' or 'name'` | The client does not receive the profile claims. Use a confidential client and include `profile` in `scope`. |
| `username_attribute not found in userinfo object` | The attribute is not released for this client. Request the scope that carries it. |
| TLS errors reaching the provider or LDAP | Check the trust anchors with [`util/tls-debug`](util/tls-debug/README.md). |
| The device prompt never appears over SSH | `KbdInteractiveAuthentication` or `UsePAM` is off, which `sshd -T` shows. Otherwise sshd accepted the user's SSH key first, as [HOWTO.md](HOWTO.md#in-sshd) explains. |

## History

The module was first developed at [Masaryk University](https://github.com/ICS-MU/pam_oauth2_device).
[UKRI-STFC](https://github.com/stfc/pam_oauth2_device) refactored it in September 2021 and added bypass.
The work was part of the [IRIS](https://www.iris.ac.uk/) activity.
[STFC](https://www.ukri.org/councils/stfc/), part of [UK Research and Innovation](https://www.ukri.org/), funded it.
Custos added the sign-in prompt options.
A [demo video](https://drive.google.com/file/d/1WzDRL0RFDXfvUgabbXNzBppV-DKXyUN1/view?usp=sharing) shows the upstream module.

| Version | Change |
|---|---|
| v1.03 | Configuration split across files |
| v1.00 (commit `efd50f6376a65319a2834b957a2cf6bbe6b7aa2d`) | First version for a full-scale test |
