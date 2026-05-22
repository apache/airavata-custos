// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package model

import "time"

// UserDN binds an X.509 distinguished name (e.g. mTLS client cert subject) to
// a Custos user. AMIE delivers DnList fields that span multiple federated
// sites, so DN storage is connector-local rather than a core concern.
// UserID references core users.id by value; no FK enforces it.
type UserDN struct {
	ID        string    `db:"id"         json:"id"`
	UserID    string    `db:"user_id"    json:"user_id"`
	DN        string    `db:"dn"         json:"dn"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
