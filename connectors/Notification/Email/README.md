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
