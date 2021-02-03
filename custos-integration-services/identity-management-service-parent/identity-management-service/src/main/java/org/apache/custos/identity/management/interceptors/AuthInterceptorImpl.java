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
import org.apache.custos.identity.management.utils.Constants;
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

/**
 * Responsible for managing auth flow
 */
@Component
public class AuthInterceptorImpl extends MultiTenantAuthInterceptor {

    private static final Logger LOGGER = LoggerFactory.getLogger(AuthInterceptorImpl.class);

    @Autowired
    public AuthInterceptorImpl(CredentialStoreServiceClient credentialStoreServiceClient, TenantProfileClient tenantProfileClient, IdentityClient identityClient) {
        super(credentialStoreServiceClient, tenantProfileClient, identityClient);
    }

    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT reqT) {

        if (method.equals("authorize") || method.equals("getAgentToken") || method.equals("endAgentSession")) {

            return reqT;
        }


        if (method.equals("authenticate")) {
            AuthClaim claim = authorize(headers);
            if (claim == null) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }
            AuthenticationRequest reqCore =
                    ((AuthenticationRequest) reqT).toBuilder()
                            .setTenantId(claim.getTenantId())
                            .setClientId(claim.getIamAuthId())
                            .setClientSecret(claim.getIamAuthSecret())
                            .build();

            return (ReqT) reqCore;
        } else if (method.equals("isAuthenticated")) {
            AuthClaim claim = authorize(headers);
            if (claim == null) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }

            String accessToken = ((AuthToken) reqT).getAccessToken();

            AuthToken.Builder authzBuilder = AuthToken.newBuilder()
                    .setAccessToken(accessToken);

            String username = null;

            for (Claim claims : ((AuthToken) reqT).getClaimsList()) {
                if (claims.getKey().equals("username")) {
                    username = claims.getValue();
                }
            }
            Claim userClaim = Claim.newBuilder().setKey("username").setValue(username).build();

            Claim tenantClaim = Claim.newBuilder().setKey("tenantId").setValue(String.valueOf(claim.getTenantId())).build();

            authzBuilder.addClaims(userClaim);
            authzBuilder.addClaims(tenantClaim);

            return (ReqT) authzBuilder.build();


        } else if (method.equals("getUser")) {

            AuthToken authToken = ((AuthToken) reqT);
            String clientId = null;
            for (Claim claim : authToken.getClaimsList()) {
                if (claim.getKey().equals(Constants.CLIENT_ID)) {
                    clientId = claim.getValue();
                }
            }

            String username = null;
            for (Claim claims : ((AuthToken) reqT).getClaimsList()) {
                if (claims.getKey().equals("username")) {
                    username = claims.getValue();
                }
            }

            String accessToken = ((AuthToken) reqT).getAccessToken();
            AuthToken.Builder authzBuilder = AuthToken.newBuilder()
                    .setAccessToken(accessToken);
            AuthClaim claim = authorize(headers, clientId);
            if (claim == null) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }

            long tenantId = claim.getTenantId();

            Claim userClaim = Claim.newBuilder().setKey("username").setValue(username).build();
            Claim tenantClaim = Claim.newBuilder().setKey("tenantId").setValue(String.valueOf(tenantId)).build();
            authzBuilder.addClaims(userClaim);
            authzBuilder.addClaims(tenantClaim);

            return (ReqT) authzBuilder.build();
        } else if (method.equals("getUserManagementServiceAccountAccessToken")) {
            AuthClaim claim = authorize(headers);
            if (claim == null) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }

            org.apache.custos.identity.service.GetUserManagementSATokenRequest request =
                    org.apache.custos.identity.service.GetUserManagementSATokenRequest
                            .newBuilder()
                            .setTenantId(claim.getTenantId())
                            .setClientId(claim.getIamAuthId())
                            .setClientSecret(claim.getIamAuthSecret())
                            .build();
            return (ReqT) request;

        } else if (method.equals("token")) {
            AuthClaim claim = authorize(headers);
            if (claim == null) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }

            GetTokenRequest request = ((GetTokenRequest) reqT).toBuilder()
                    .setTenantId(claim.getTenantId())
                    .setClientId(claim.getIamAuthId())
                    .setClientSecret(claim.getIamAuthSecret())
                    .build();
            return (ReqT) request;

        } else if (method.equals("getCredentials")) {
            AuthClaim claim = authorize(headers);
            if (claim == null) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }

            Credentials credentials = Credentials.newBuilder()
                    .setCustosClientId(claim.getCustosId())
                    .setCustosClientSecret(claim.getCustosSecret())
                    .setCustosClientIdIssuedAt(claim.getCustosIdIssuedAt())
                    .setCustosClientSecretExpiredAt(claim.getCustosSecretExpiredAt())
                    .setCiLogonClientId(claim.getCiLogonId())
                    .setCiLogonClientSecret(claim.getCiLogonSecret())
                    .setIamClientId(claim.getIamAuthId())
                    .setIamClientSecret(claim.getIamAuthSecret())
                    .build();

            return (ReqT) ((GetCredentialsRequest) reqT).toBuilder().setCredentials(credentials).build();

        } else if (method.equals("endUserSession")) {
            AuthClaim claim = authorize(headers);
            if (claim == null) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }
            org.apache.custos.identity.service.EndSessionRequest endSessionRequest =
                    ((EndSessionRequest) reqT).getBody().toBuilder()
                            .setClientId(claim.getIamAuthId())
                            .setClientSecret(claim.getIamAuthSecret())
                            .setTenantId(claim.getTenantId())
                            .build();
            return (ReqT) ((EndSessionRequest) reqT).toBuilder().setBody(endSessionRequest).build();

        }

        return reqT;
    }
}
