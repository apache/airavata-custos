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

package org.apache.custos.log.management.interceptors;

import io.grpc.Metadata;
import org.apache.custos.credential.store.client.CredentialStoreServiceClient;
import org.apache.custos.identity.client.IdentityClient;
import org.apache.custos.integration.core.exceptions.NotAuthorizedException;
import org.apache.custos.integration.services.commons.interceptors.MultiTenantAuthInterceptor;
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.logging.service.LogEventRequest;
import org.apache.custos.logging.service.LoggingConfigurationRequest;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

/**
 * Responsible for validate confidential client specific authorization.
 * Methods which authenticates based only on client are implemented here.
 */
@Component
public class ClientAuthInterceptorImpl extends MultiTenantAuthInterceptor {
    private static final Logger LOGGER = LoggerFactory.getLogger(ClientAuthInterceptorImpl.class);

    @Autowired
    public ClientAuthInterceptorImpl(CredentialStoreServiceClient credentialStoreServiceClient, TenantProfileClient tenantProfileClient, IdentityClient identityClient) {
        super(credentialStoreServiceClient, tenantProfileClient, identityClient);
    }

    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT reqT) {


        if (method.equals("getLogEvents")) {
            LogEventRequest request = (LogEventRequest) reqT;
            AuthClaim claim = authorize(headers, request.getClientId());

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();

            long tenantId = claim.getTenantId();
            return (ReqT) ((LogEventRequest) reqT).toBuilder()
                    .setClientId(oauthId)
                    .setTenantId(tenantId)
                    .build();

        } else if (method.equals("isLogEnabled")) {

            LoggingConfigurationRequest request = (LoggingConfigurationRequest) reqT;
            AuthClaim claim = authorize(headers, request.getClientId());

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();

            long tenantId = claim.getTenantId();
            LoggingConfigurationRequest registerUserRequest =
                    ((LoggingConfigurationRequest) reqT).toBuilder()
                            .setTenantId(tenantId)
                            .setClientId(oauthId)
                            .build();

            return (ReqT) registerUserRequest;
        }
        return reqT;
    }

}
