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

# Signer

The Signer connector is the only caller of
the [SSH certificate signer](../../extensions/SSH-Certificate-Signer/README.md).
It decides who may have a certificate, then calls the signer with one service token. The caller's own token never goes
to the signer. To offer the signer without the rest of Custos, enable only this connector.

## Routes

| Route                                                     | Needs                       | Does                                                          |
|-----------------------------------------------------------|-----------------------------|---------------------------------------------------------------|
| `GET /connectors/signer/certificates`                     | `signer:certificates:read`  | All certificates                                              |
| `GET /connectors/signer/certificates/{serial}`            | `signer:certificates:read`  | Any certificate                                               |
| `POST /connectors/signer/certificates/{serial}/revoke`    | `signer:certificates:write` | Revoke any certificate, recorded under the email of the admin |
| `GET /connectors/signer/me/certificates`                  | Signed in                   | The caller's own certificates                                 |
| `POST /connectors/signer/me/certificates`                 | Signed in                   | Generate a certificate for the caller's account               |
| `GET /connectors/signer/me/certificates/{serial}`         | Signed in                   | One of the caller's own certificates                          |
| `POST /connectors/signer/me/certificates/{serial}/revoke` | Signed in                   | Revoke one of the caller's own, recorded as `self`            |
| `GET /connectors/signer/status`                           | `signer:certificates:read`  | Whether the signer answers its health route, and how fast     |

The certificate lists are keyset-paged: pass `limit` and the last `next_cursor` as `cursor`. Someone else's certificate
on a `me` route is a 404.

To generate, the caller needs an approved and provisioned cluster account on `cluster_id`, and an active membership in
an
active allocation on that cluster. The username comes from the cluster account, and the email and subject from the
caller's verified token, never from the request.

A refusal has an `error` code: `user_not_active`, `no_cluster_account`, `cluster_account_not_ready`,
`no_active_allocation`, `signer_unavailable`, or from the signer `reserved_username`, `invalid_public_key` and
`key_type_not_allowed`.

## App routes

An app signs in with its token and can use only these routes. It never holds the privileges above.

| Route                                                      | Action                            | Does                                                  |
|------------------------------------------------------------|-----------------------------------|-------------------------------------------------------|
| `POST /connectors/signer/app/certificates`                 | `signer:app:generate-certificate` | Generate a certificate for the user in `user_email`   |
| `GET /connectors/signer/app/certificates`                  | `signer:app:generate-certificate` | The certificates this app requested                   |
| `GET /connectors/signer/app/certificates/{serial}`         | `signer:app:generate-certificate` | One certificate this app requested                    |
| `POST /connectors/signer/app/certificates/{serial}/revoke` | `signer:app:revoke-certificate`   | Revoke one this app requested, recorded under the app |

Generate takes `user_email`, `public_key` and an optional `source_address`. Every certificate records `app/<key>` as who
asked. The user must be active, have accepted the current user agreement, have signed in to Custos at least once, have
allowed the app, and pass the same account and allocation checks as above. Besides the codes above, an app can get
`user_not_found`, `consent_required`, `user_not_signed_in` and `app_not_authorized`.

## Configuration

Settings are in the `connectors.signer` block of [`config/custos.yaml`](../../config/custos.yaml). When a required
setting is empty the connector skips loading and reports the gap on `GET /connectors`.

| Key                       | Variable in `config/custos.yaml` | Required | Purpose                                        |
|---------------------------|----------------------------------|----------|------------------------------------------------|
| `base_url`                | `SIGNER_BASE_URL`                | yes      | The signer's URL. Keep it off public networks. |
| `service_token`           | `SIGNER_SERVICE_TOKEN`           | yes      | The same token set on the signer               |
| `cluster_id`              | `CUSTOS_CLUSTER_ID`              | yes      | The cluster certificates log in to             |
| `certificate_ttl_seconds` |                                  | no       | Certificate lifetime, default `28800`          |
| `timeout_seconds`         |                                  | no       | Request timeout, default `10`                  |
