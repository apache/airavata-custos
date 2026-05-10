# OIDC Login PAM Module HOWTO

This is a HOWTO document describing how to use the OIDC device flow authenticator PAM module as refactored and augmented
by the [IRIS](www.iris.ac.uk) team.  This work was funded by [STFC](https://www.ukri.org/councils/stfc/), a part of
[UK Research and Innovation (UKRI)](https://www.ukri.org/).

## A Tale of Two (or Three) User Ids

Users who authenticate with this module have two or three user ids.

 * A *remote* user id, picked from the userinfo record published by the IAM.
 * A *local* user id, requested by the user on the ssh command line (or similar login tool)
 * A possibly-suffixed local id, which is used to put a common suffix on all local accounts that users are allowed to
   log into using this module.  See below.

## How to configure sshd to use the module

Your PAM configuration needs to include this module.  A typical example is:

```
#%PAM-1.0
auth       required     pam_sepermit.so
auth       required     pam_oauth2_device.so
account    required     pam_nologin.so
account    required     pam_oauth2_device.so
```
(and it would then continue with session entries).  At the time of writing, the entry for this module in the account
section does nothing (i.e., it always returns success), but this may change in future releases.

This file typically needs to be stored in the directory `/etc/pam.d` with the name of the executable of the ssh daemon
(e.g. `/etc/pam.d/sshd`).

## How to combine this module with other authentication methods

The *bypass* feature allows some users to completely skip the authentication (and authorisation) steps of the module.
Bypass is available in two flavours, one using an LDAP callout and the other using a local username lookup.

In order to use the bypass feature, there must be other authentication/authorisation modules in the PAM configuration.
If not, login will fail.

### Bypass configuration

Currently there are two methods of bypassing the module.

#### Local Users

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

This section of the confiuration would ask the PAM module to skip the OIDC authentication/authorisation for login as
(local) user root and user fred.  There is no remote username in this case, because the OIDC authentication is skipped,
and the username suffix (if defined) is thus ignored for this check.

#### LDAP

If LDAP is configured, the attribute `preauth` can be defined (it is not clear why it is called "preauth" and not
"bypass" but that is what it does).  It should contain an LDAP query in which the first occurrences of `%s` gets the
local username substituted into the query.  The OIDC login is bypassed if and only if the LDAP query is successful
(regardless of what it returns).

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

### Configuration other authentication methods

The bypass feature was designed to selectively let the module say "don't know" about a set of users, rather than "yes"
(authentication successful) or "no" (authentication unsuccessful).

However, this presumes there is a fallback option (if there isn't, authentication will fail.)  There are two
possibilities: have sshd use another method, or configure another method in PAM.

#### Configuring other authentication methods in PAM




#### Configuring other authentication methods in sshd

Assuming the ssh daemon -- and the client -- come from the OpenSSH implementation, the client tries the following
methods:

 - Host based authentication
 - Public key authentication
 - Challenge/response authentication
   - This includes the PAM authentication
 - Password authentication

