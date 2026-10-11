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

Emails users about their cluster account: when a request is received,
when the account is ready, and when a request is denied. Settings are in the
`email-notifier` block of `config/custos.yaml`.

## Custom email templates

The emails are the Go templates in `internal/notify/templates/`. To change
them for a deployment, copy the files you want to change into a directory
on the server and set `EMAIL_TEMPLATES_DIR` to it. A file there replaces
the built-in file with the same name. The others stay as they are. The
server does not start if a file in there does not parse.

- `layout.html.tmpl` is the HTML shell around every email. Most styling is here.
- `request-received`, `account-ready` and `request-denied` each have an
  `.html.tmpl` for the body and a `.txt.tmpl` for the subject and plain text.

The fields a template can use are in `Data` in `internal/notify/render.go`.

## Send the emails to your inbox

`internal/notify/live_send_test.go` sends all three emails with sample data
through a real mail server, so you can check how they look in a real inbox.
It only runs with the `livemail` tag and when `EMAIL_TEST_TO` is set.

Set the `EMAIL_*` values in `.env`, plus `EMAIL_TEST_TO` with your address,
then run from the repo root:

```sh
set -a && . ./.env && set +a
go test -tags livemail -run TestLiveSend -v ./connectors/Notification/Email/internal/notify/
```
