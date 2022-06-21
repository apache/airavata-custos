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

package org.apache.custos.identity.management.interceptors;

import io.grpc.Metadata;
import org.apache.custos.credential.store.client.CredentialStoreServiceClient;
import org.apache.custos.credential.store.service.Credentials;
import org.apache.custos.identity.client.IdentityClient;
import org.apache.custos.identity.management.service.EndSessionRequest;
import org.apache.custos.identity.management.service.GetCredentialsRequest;
import org.apache.custos.identity.service.AuthToken;
import org.apache.custos.identity.service.AuthenticationRequest;
import org.apache.custos.identity.service.Claim;
import org.apache.custos.identity.service.GetTokenRequest;
import org.apache.custos.integration.core.exceptions.UnAuthorizedException;
import org.apache.custos.integration.services.commons.interceptors.MultiTenantAuthInterceptor;
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.Optional;

/**
 * Responsible for managing auth flow
 */
@Component
public class IdentityManagementAuthInterceptorImpl extends MultiTenantAuthInterceptor {

    private static final Logger LOGGER = LoggerFactory.getLogger(IdentityManagementAuthInterceptorImpl.class);

    @Autowired
    public IdentityManagementAuthInterceptorImpl(CredentialStoreServiceClient credentialStoreServiceClient,
                                                 TenantProfileClient tenantProfileClient, IdentityClient identityClient) {
        super(credentialStoreServiceClient, tenantProfileClient, identityClient);
    }

    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT reqT) {

        LOGGER.info("request received ....." + method);
        if (method.equals("authorize") || method.equals("getAgentToken") || method.equals("endAgentSession")) {
            return reqT;
        }

        if (method.equals("authenticate")) {
            Optional<AuthClaim> claim = authorize(headers);
            return claim.map(cl -> {
                AuthenticationRequest reqCore =
                        ((AuthenticationRequest) reqT).toBuilder()
                                .setTenantId(cl.getTenantId())
                                .setClientId(cl.getIamAuthId())
                                .setClientSecret(cl.getIamAuthSecret())
                                .build();

                return (ReqT) reqCore;
            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });

        } else if (method.equals("isAuthenticated") || method.equals("getUser")) {
            Optional<AuthClaim> claim = authorize(headers);
            if (claim.isEmpty()) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }

            String accessToken = ((AuthToken) reqT).getAccessToken();

            Optional<AuthClaim> authClaim = authorizeUsingUserToken(accessToken);

            AuthToken.Builder authzBuilder = AuthToken.newBuilder()
                    .setAccessToken(accessToken);

            if (authClaim.isPresent()) {
                Claim userClaim = Claim.newBuilder().setKey("username").setValue(authClaim.get()
                        .getUsername()).build();

                Claim tenantClaim = Claim.newBuilder().setKey("tenantId").setValue(String.valueOf(authClaim.get()
                        .getTenantId())).build();
                Claim clientClaim = Claim.newBuilder().setKey("clientId").setValue(String.valueOf(authClaim.get()
                        .getCustosId())).build();
                authzBuilder.addClaims(userClaim);
                authzBuilder.addClaims(tenantClaim);
                authzBuilder.addClaims(clientClaim);
            } else {
                throw new UnAuthorizedException("Request is not authorized, User token not found", null);
            }
            return (ReqT) authzBuilder.build();


        } else if (method.equals("getUserManagementServiceAccountAccessToken")) {
            Optional<AuthClaim> claim = authorize(headers);
            return claim.map(cl -> {
                org.apache.custos.identity.service.GetUserManagementSATokenRequest request =
                        org.apache.custos.identity.service.GetUserManagementSATokenRequest
                                .newBuilder()
                                .setTenantId(cl.getTenantId())
                                .setClientId(cl.getIamAuthId())
                                .setClientSecret(cl.getIamAuthSecret())
                                .build();
                return (ReqT) request;
            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });

        } else if (method.equals("token")) {
            Optional<AuthClaim> claim = authorize(headers);
            return claim.map(cl -> {
                GetTokenRequest request = ((GetTokenRequest) reqT).toBuilder()
                        .setTenantId(cl.getTenantId())
                        .setClientId(cl.getIamAuthId())
                        .setClientSecret(cl.getIamAuthSecret())
                        .build();
                return (ReqT) request;
            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });

        } else if (method.equals("getCredentials")) {
            String clientId = ((GetCredentialsRequest) reqT).getClientId();
            Optional<AuthClaim> claim = authorize(headers, clientId);
            return claim.map(cl -> {
                Credentials credentials = Credentials.newBuilder()
                        .setCustosClientId(cl.getCustosId())
                        .setCustosClientSecret(cl.getCustosSecret())
                        .setCustosClientIdIssuedAt(cl.getCustosIdIssuedAt())
                        .setCustosClientSecretExpiredAt(cl.getCustosSecretExpiredAt())
                        .setCiLogonClientId(cl.getCiLogonId())
                        .setCiLogonClientSecret(cl.getCiLogonSecret())
                        .setIamClientId(cl.getIamAuthId())
                        .setIamClientSecret(cl.getIamAuthSecret())
                        .build();

                return (ReqT) ((GetCredentialsRequest) reqT).toBuilder().setCredentials(credentials).build();
            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });

        } else if (method.equals("endUserSession")) {
            Optional<AuthClaim> claim = authorize(headers);
            return claim.map(cl -> {
                org.apache.custos.identity.service.EndSessionRequest endSessionRequest =
                        ((EndSessionRequest) reqT).getBody().toBuilder()
                                .setClientId(cl.getIamAuthId())
                                .setClientSecret(cl.getIamAuthSecret())
                                .setTenantId(cl.getTenantId())
                                .build();
                return (ReqT) ((EndSessionRequest) reqT).toBuilder().setBody(endSessionRequest).build();

            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });
        }

        return reqT;
    }
}
