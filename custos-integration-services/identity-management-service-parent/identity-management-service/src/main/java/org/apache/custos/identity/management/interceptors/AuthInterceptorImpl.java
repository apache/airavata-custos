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
import org.apache.custos.identity.management.service.AuthenticationRequest;
import org.apache.custos.identity.management.service.AuthorizationRequest;
import org.apache.custos.identity.service.AuthToken;
import org.apache.custos.identity.service.Claim;
import org.apache.custos.identity.service.GetTokenRequest;
import org.apache.custos.integration.core.exceptions.NotAuthorizedException;
import org.apache.custos.integration.services.commons.interceptors.AuthInterceptor;
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
public class AuthInterceptorImpl extends AuthInterceptor {

    private static final Logger LOGGER = LoggerFactory.getLogger(AuthInterceptorImpl.class);

    @Autowired
    public AuthInterceptorImpl(CredentialStoreServiceClient credentialStoreServiceClient, TenantProfileClient tenantProfileClient) {
        super(credentialStoreServiceClient, tenantProfileClient);
    }

    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT reqT) {

        if (method.equals("getOIDCConfiguration") || method.equals("authorize") || method.equals("authorizetest")) {

            return reqT;
        }

        AuthClaim claim = authorize(headers);
        if (claim == null) {
            throw new NotAuthorizedException("Request is not authorized", null);
        }
        if (method.equals("authenticate")) {

            String userName = ((AuthenticationRequest) reqT).getAuthRequest().getUsername();
            String password = ((AuthenticationRequest) reqT).getAuthRequest().getPassword();
            org.apache.custos.identity.service.AuthenticationRequest reqCore =
                    org.apache.custos.identity.service.AuthenticationRequest
                            .newBuilder()
                            .setTenantId(claim.getTenantId())
                            .setClientId(claim.getIamAuthId())
                            .setClientSecret(claim.getIamAuthSecret())
                            .setUsername(userName)
                            .setPassword(password)
                            .build();

            return (ReqT) AuthenticationRequest.newBuilder().setAuthRequest(reqCore).build();
        } else if (method.equals("isAuthenticated") || method.equals("getUser")) {

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


        } else if (method.equals("getUserManagementServiceAccountAccessToken")) {

            org.apache.custos.identity.service.GetUserManagementSATokenRequest request =
                    org.apache.custos.identity.service.GetUserManagementSATokenRequest
                            .newBuilder()
                            .setTenantId(claim.getTenantId())
                            .setClientId(claim.getIamAuthId())
                            .setClientSecret(claim.getIamAuthSecret())
                            .build();
            return (ReqT) request;

        }  else if (method.equals("token")) {

            GetTokenRequest request = ((GetTokenRequest) reqT).toBuilder()
                    .setTenantId(claim.getTenantId())
                    .setClientId(claim.getIamAuthId())
                    .setClientSecret(claim.getIamAuthSecret())
                    .build();
            return (ReqT) request;

        }
        return reqT;
    }


}
