/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing,
 *  software distributed under the License is distributed on an
 *  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *  KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

package org.apache.custos.integration.services.commons.interceptors;

import io.grpc.Metadata;
import org.apache.custos.credential.store.client.CredentialStoreServiceClient;
import org.apache.custos.identity.client.IdentityClient;
import org.apache.custos.integration.core.utils.Constants;
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Responsible for authorization  of  multi tenant middleware requests
 */
public abstract class MultiTenantAuthInterceptor extends AuthInterceptor {

    private static final Logger LOGGER = LoggerFactory.getLogger(AuthInterceptor.class);

    private CredentialStoreServiceClient credentialStoreServiceClient;

    private TenantProfileClient tenantProfileClient;

    private IdentityClient identityClient;

    public MultiTenantAuthInterceptor(CredentialStoreServiceClient credentialStoreServiceClient,
                                      TenantProfileClient tenantProfileClient, IdentityClient identityClient) {
        super(credentialStoreServiceClient, tenantProfileClient, identityClient);
        this.credentialStoreServiceClient = credentialStoreServiceClient;
        this.tenantProfileClient = tenantProfileClient;
        this.identityClient = identityClient;

    }


    public AuthClaim authorize(Metadata headers, String clientId) {

        if (clientId != null && clientId.trim().equals("")) {
            clientId = null;
        }

        String userToken = headers.get(Metadata.Key.of(Constants.USER_TOKEN, Metadata.ASCII_STRING_MARSHALLER));

        if (clientId == null && userToken == null) {
            return authorize(headers);
        } else if (clientId != null && userToken == null) {
            return authorizeWithParentChildTenantValidationByBasicAuth(headers, clientId);
        } else if (clientId != null && userToken != null) {
            return authorizeWithParentChildTenantValidationByBasicAuthAndUserTokenValidation(headers, clientId, userToken);
        } else {
            return authorizeUsingUserToken(headers);
        }
    }
}
