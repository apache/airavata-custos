-- Licensed to the Apache Software Foundation (ASF) under one
-- or more contributor license agreements.  See the NOTICE file
-- distributed with this work for additional information
-- regarding copyright ownership.  The ASF licenses this file
-- to you under the Apache License, Version 2.0 (the
-- "License"); you may not use this file except in compliance
-- with the License.  You may obtain a copy of the License at
--
--   http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing,
-- software distributed under the License is distributed on an
-- "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
-- KIND, either express or implied.  See the License for the
-- specific language governing permissions and limitations
-- under the License.

-- oidc_sub is nullable: not every external identity issues an OIDC subject
-- (AMIE binds by external_id only). UNIQUE permits multiple NULLs but blocks
-- collisions on real values across IdPs.
CREATE TABLE IF NOT EXISTS external_identities
(
    id          VARCHAR(255) NOT NULL,
    user_id     VARCHAR(255) NOT NULL,
    source      VARCHAR(64)  NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    oidc_sub    VARCHAR(255) NULL DEFAULT NULL,
    metadata    TEXT         NULL DEFAULT NULL,
    created_at  TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at  TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_external_identities_source_external (source, external_id),
    UNIQUE KEY uq_external_identities_oidc_sub (oidc_sub),
    KEY idx_external_identities_user (user_id),
    CONSTRAINT fk_external_identities_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- A DN is a globally-unique credential. UNIQUE on dn alone subsumes
-- (user_id, dn), so the composite index is omitted.
CREATE TABLE IF NOT EXISTS user_dns
(
    id         VARCHAR(255) NOT NULL,
    user_id    VARCHAR(255) NOT NULL,
    dn         VARCHAR(512) NOT NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_user_dns_dn (dn),
    KEY idx_user_dns_user (user_id),
    CONSTRAINT fk_user_dns_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
