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

# Email Notifier

The Email Notifier emails a researcher at each step of a cluster account request: received, approved or denied, and
account ready. It sends through an SMTP server.

## How it works

| Event | Email | Sent when |
|---|---|---|
| `compute_cluster_user::create` | `request-received` | The account is pending approval |
| `compute_cluster_user::approve` | `request-approved` | The user has not signed in yet, so they must sign in to finish setup |
| `compute_cluster_user::ready` | `account-ready` | Core finds the account's identity and storage both done |
| `compute_cluster_user::deny` | `request-denied` | An administrator denied the request. The email carries the review note |

Each email names the project of the user's first allocation on that cluster. That allocation's request created the
cluster account. A successful send writes a `NOTIFICATION_SENT` audit row with the template and recipient.

The [event bus](../../../docs/architecture.md#event-bus) retries a failed send, which resends only that email. A user
with no email address, or a template that fails to render, fails the delivery at once.

## Configuration

Settings are in the `connectors.email-notifier` block of [`config/custos.yaml`](../../../config/custos.yaml). An empty required
setting makes the connector skip loading. It logs `email notifications: required setting not set; skipping` with the
setting's name and reports the gap on `GET /connectors`.

| Key | Variable in `config/custos.yaml` | Required | Purpose |
|---|---|---|---|
| `smtp_host` | `EMAIL_SMTP_HOST` | yes | SMTP server |
| `smtp_port` | | no | SMTP port, default `587` |
| `username`, `password` | `EMAIL_SMTP_USERNAME`, `EMAIL_SMTP_PASSWORD` | no | SMTP login. With a username, TLS is mandatory. Without one, TLS is opportunistic, for a relay that needs no login. |
| `from` | `EMAIL_FROM` | yes | Sender address |
| `support_email` | `EMAIL_SUPPORT` | yes | Contact address shown in the emails |
| `site_name` | `EMAIL_SITE_NAME` | yes | Site name shown in the emails |
| `portal_url` | `EMAIL_PORTAL_URL` | yes | Portal link |
| `cluster_host` | `EMAIL_CLUSTER_HOST` | yes | Host users SSH to |
| `logo_url` | `EMAIL_LOGO_URL` | no | A square PNG. Defaults to `<portal_url>/brand/logo/custos-mark.png`. |
| `templates_dir` | `EMAIL_TEMPLATES_DIR` | no | Directory of custom templates |

### Custom templates

The emails are the Go templates in `internal/notify/templates/`. To customize, copy the files to change into a server
directory and set `templates_dir` to it. A file there replaces the built-in file of the same name. Other built-in files
stay. The server does not start if a file there does not parse.

| File | Contents |
|---|---|
| `layout.html.tmpl` | The HTML shell around each email, with most styling |
| `<name>.html.tmpl` | The HTML body |
| `<name>.txt.tmpl` | The subject and the plain-text body |

`<name>` is `request-received`, `request-approved`, `account-ready` or `request-denied`. `Data` in
`internal/notify/render.go` lists the template fields.

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| No emails; log shows `required setting not set; skipping` | A required setting is empty | Set it, then restart |
| A delivery fails at once with `user ... has no email` | The user record has no email | Add the email, then retry the delivery under `/events/deliveries` |
| Deliveries keep retrying | The SMTP server rejects the connection or login | Check host, port and credentials. A login requires TLS |
| A user who already signed in gets no `request-approved` email | Expected | They get `account-ready` once the account is ready |

## Development

`internal/notify/live_send_test.go` sends the four emails with sample data through a real mail server. It needs the
`livemail` build tag and `EMAIL_TEST_TO`. Set the `EMAIL_*` values and `EMAIL_TEST_TO` (your address) in `.env`, then
run from the repo root:

```sh
set -a && . ./.env && set +a
go test -tags livemail -run TestLiveSend -v ./connectors/Notification/Email/internal/notify/
```

The unit tests need no mail server. The integration test carries the `integration` tag and needs
`CORE_TEST_DATABASE_DSN` ([CONTRIBUTING.md](../../../CONTRIBUTING.md)).

```sh
go test ./connectors/Notification/Email/...
go test -tags integration ./connectors/Notification/Email/...
```
