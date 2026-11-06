-- Licensed to the Apache Software Foundation (ASF) under one
-- or more contributor license agreements. See the NOTICE file
-- distributed with this work for additional information
-- regarding copyright ownership. The ASF licenses this file
-- to you under the Apache License, Version 2.0 (the
-- "License"); you may not use this file except in compliance
-- with the License. You may obtain a copy of the License at
--
-- http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing,
-- software distributed under the License is distributed on an
-- "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
-- KIND, either express or implied. See the License for the
-- specific language governing permissions and limitations
-- under the License.

-- Dev client for the compose stack. Secret (bcrypt below): bench-secret
INSERT INTO client_ssh_configs
    (client_id, client_secret, target_host, target_port,
     max_ttl_seconds, allowed_key_types, enabled)
VALUES
    ('bench',
     '$2y$10$2K2FooZeA95Z7ilfrZwB4OZOAKyBKwrT6CePe8q57psu/p3k1g1aa',
     'localhost', 2222, 86400, '["ed25519","rsa","ecdsa"]', TRUE)
ON CONFLICT (client_id) DO UPDATE SET
    client_secret = EXCLUDED.client_secret,
    enabled = EXCLUDED.enabled;
