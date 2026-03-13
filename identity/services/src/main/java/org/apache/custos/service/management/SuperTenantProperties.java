/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

package org.apache.custos.service.management;

import org.springframework.boot.context.properties.ConfigurationProperties;

@ConfigurationProperties(prefix = "custos.super-tenant")
public record SuperTenantProperties(
        Admin admin,
        String redirectUris,
        String clientUri,
        String scope,
        String domain
) {
    public SuperTenantProperties {
        if (admin == null) admin = new Admin("", "", "");
        if (redirectUris == null) redirectUris = "";
        if (clientUri == null) clientUri = "";
        if (scope == null) scope = "openid email profile cilogon";
        if (domain == null) domain = "localhost";
    }

    public record Admin(String username, String password, String email) {
        public Admin {
            if (username == null) username = "";
            if (password == null) password = "";
            if (email == null) email = "";
        }
    }
}
