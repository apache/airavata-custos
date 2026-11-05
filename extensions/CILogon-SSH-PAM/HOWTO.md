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

# CILogon SSH PAM module HOWTO

This is a HOWTO document describing how to use the OIDC device flow authenticator PAM module as refactored and augmented
by the [IRIS](https://www.iris.ac.uk/) team.  This work was funded by [STFC](https://www.ukri.org/councils/stfc/), a part of
[UK Research and Innovation (UKRI)](https://www.ukri.org/).

The [README](README.md) covers the flow, usernames and configuration.

## Configure sshd to use the module

Your PAM configuration needs to include this module.  A typical example is:

```
#%PAM-1.0
auth       required     pam_sepermit.so
auth       required     pam_oauth2_device.so
account    required     pam_nologin.so
account    required     pam_oauth2_device.so
```
(and it would then continue with session entries).  The account entry and `pam_sm_setcred` return success, or
`PAM_IGNORE` for a bypassed user.

This file typically needs to be stored in the directory `/etc/pam.d` with the name of the executable of the ssh daemon
(e.g. `/etc/pam.d/sshd`).

## Bypass

The *bypass* feature allows some users to completely skip the authentication (and authorisation) steps of the module.
Bypass is available in two flavours, one using an LDAP callout and the other using a local username lookup.
A bypassed user gets `PAM_IGNORE`, and PAM moves to the next module.

In order to use the bypass feature, there must be other authentication/authorisation modules in the PAM configuration.
If not, login will fail.

Currently there are two methods of bypassing the module.

### Local users

The `users` section of the configuration file can have a special entry called `*bypass*` where normally the remote
username would be specified.  So the section would be

```
"users": {
  "*bypass*":
  [
    "root",
	"fred"
  ],
  "remote1@example.com":
  [
    "znap"
  ]
}
```

This section of the configuration would ask the PAM module to skip the OIDC authentication/authorisation for login as
(local) user root and user fred.  There is no remote username in this case, because the OIDC authentication is skipped,
and the username suffix (if defined) is thus ignored for this check.

### LDAP

If LDAP is configured, the attribute `preauth` can be defined (it is not clear why it is called "preauth" and not
"bypass" but that is what it does).  It should contain an LDAP query in which the first occurrences of `%s` gets the
local username substituted into the query.  The OIDC login is bypassed if the query returns one or more entries; an
empty result or an LDAP error does not bypass it.

```
"ldap": {
  "host": "ldaps://ldap.example.com:636",
  "basedn": "ou=users,dc=example,dc=com",
  "user": "system",
  "passwd": "abc123@321cba",
  "filter": "(&(objectclass=user)(uid=%s))",
  "preauth": "(&(objectclass=user)(cn=%s)(department=scd))",
  "attr": "cn"
}
```

This is for a system where a username and password are required to query the LDAP server.  We assume that user records:

 * Are identified with `(objectclass=user)`
 * Have a `department` entry
 * Have a `uid` entry which should match the **remote** user id (i.e. the id that IAM expects you to have)
 * Have a `cn` (commonName) entry which matches the **local** user id

Here, users have records that are searched with the (default subtree) search; the filter would select user records
(assuming the cn (commonName) is the string users would use as their local id).  Users who additionally have a
department attribute with a value of "scd" would be bypassed.

For users who are *not* bypassed, the value of the cn attribute in the user object is looked up and compared to the id
requested by the user.

Once again,

 * The filter uses the **remote** id to look up a record
 * The preauth (bypass) feature uses the **local** id because there is no remote id prior to the module running

## Combine the module with other methods

The bypass feature was designed to selectively let the module say "don't know" about a set of users, rather than "yes"
(authentication successful) or "no" (authentication unsuccessful).

However, this presumes there is a fallback option (if there isn't, authentication will fail.)  There are two
possibilities: have sshd use another method, or configure another method in PAM.

### In PAM

Mark the module `sufficient` and add another method after it. Success ends the `auth` phase. Bypass or failure continues.
The [node enrollment playbook](../../dev-ops/account-provisioning/README.md) writes this stack on RHEL.
The listing omits its `password` and `session` lines and the config-path argument, which the module ignores.

```
#%PAM-1.0
auth       required     pam_sepermit.so
auth       sufficient   pam_oauth2_device.so
auth       substack     password-auth
auth       include      postlogin
account    required     pam_sepermit.so
account    required     pam_nologin.so
account    include      password-auth
```

### In sshd

Assuming the ssh daemon -- and the client -- come from the OpenSSH implementation, the client tries the following
methods:

1. GSSAPI authentication
2. Host-based authentication
3. Public key authentication
4. Challenge-response (keyboard-interactive) authentication, which includes PAM
5. Password authentication

By default, a user whose SSH key is accepted never reaches the module.
To require the device login, turn off the earlier methods or set `AuthenticationMethods` in `sshd_config`.
