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
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Optional;

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


    public Optional<AuthClaim> authorize(Metadata headers, String clientId) {

        try {

            if (clientId != null && clientId.trim().isEmpty()) {
                clientId = null;
            }

            boolean agentAuthenticationEnabled = isAgentAuthenticationEnabled(headers);

            if (agentAuthenticationEnabled) {
                return authorizeUsingAgentAndUserJWTTokens(headers);
            }
            Optional<String> userToken = getUserTokenFromUserTokenHeader(headers);
            boolean isBasicAuth = isBasicAuth(headers);

            if (clientId == null && userToken.isEmpty() && isBasicAuth) {
                return authorize(headers);
            } else if (clientId != null && userToken.isEmpty() && isBasicAuth) {
                return authorizeParentChildTenantValidationWithBasicAuth(headers, clientId);
            } else if (clientId != null && userToken.isPresent()) {
                return authorizeParentChildTenantWithBasicAuthAndUserTokenValidation(headers, clientId, userToken.get());
            } else if (clientId != null && isUserToken(headers)) {
                return authorizeParentChildTenantWithUserTokenValidation(headers, clientId);
            } else {
                return authorizeUsingUserToken(headers);
            }

        } catch (Exception ex) {
            LOGGER.error(ex.getMessage(),ex);
            clearUserTokenFromHeader(headers);
            throw ex;
        }
    }


}
