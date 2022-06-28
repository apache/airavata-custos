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
import org.apache.custos.integration.core.exceptions.UnAuthorizedException;
import org.apache.custos.integration.core.utils.Constants;
import org.apache.custos.integration.services.commons.interceptors.MultiTenantAuthInterceptor;
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.logging.service.LogEventRequest;
import org.apache.custos.logging.service.LoggingConfigurationRequest;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.Optional;

/**
 * Responsible for validate user specific authorization
 * Methods authenticates users access tokens are implemented here
 */
@Component
public class LogManagementAuthInterceptorImpl extends MultiTenantAuthInterceptor {
    private static final Logger LOGGER = LoggerFactory.getLogger(LogManagementAuthInterceptorImpl.class);

    @Autowired
    public LogManagementAuthInterceptorImpl(CredentialStoreServiceClient credentialStoreServiceClient, TenantProfileClient tenantProfileClient, IdentityClient identityClient) {
        super(credentialStoreServiceClient, tenantProfileClient, identityClient);
    }

    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT msg) {


        if (method.equals("enable")) {
            LoggingConfigurationRequest loggingConfigRequest = (LoggingConfigurationRequest) msg;
            headers = attachUserToken(headers, loggingConfigRequest.getClientId());
            Optional<AuthClaim> claim = authorize(headers, loggingConfigRequest.getClientId());

            if (claim.isEmpty()) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }

            if (!claim.get().isAdmin()) {
                throw new UnAuthorizedException("Your are not a tenant admin", null);
            }

            long tenantId = claim.get().getTenantId();

            return (ReqT) ((LoggingConfigurationRequest) msg).toBuilder()
                    .setTenantId(tenantId)
                    .build();
        } else if (method.equals("getLogEvents")) {
            LogEventRequest request = (LogEventRequest) msg;
            Optional<AuthClaim> claim = authorize(headers, request.getClientId());
            return claim.map(cl -> {
                String oauthId = cl.getIamAuthId();

                long tenantId = cl.getTenantId();
                return (ReqT) ((LogEventRequest) msg).toBuilder()
                        .setClientId(oauthId)
                        .setTenantId(tenantId)
                        .build();

            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });


        } else if (method.equals("isLogEnabled")) {

            LoggingConfigurationRequest request = (LoggingConfigurationRequest) msg;
            Optional<AuthClaim> claim = authorize(headers, request.getClientId());

            return claim.map(cl -> {
                String oauthId = cl.getIamAuthId();

                long tenantId = cl.getTenantId();
                LoggingConfigurationRequest registerUserRequest =
                        ((LoggingConfigurationRequest) msg).toBuilder()
                                .setTenantId(tenantId)
                                .setClientId(oauthId)
                                .build();

                return (ReqT) registerUserRequest;
            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });

        }
        return msg;
    }


    private Metadata attachUserToken(Metadata headers, String clientId) {
        if (clientId == null || clientId.trim().equals("")) {
            String formattedUserToken = getToken(headers);
            headers.put(Metadata.Key.of(Constants.USER_TOKEN, Metadata.ASCII_STRING_MARSHALLER), formattedUserToken);
            return headers;
        }
        return headers;
    }


}
