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

package org.apache.custos.user.management.interceptors;

import io.grpc.Metadata;
import org.apache.custos.credential.store.client.CredentialStoreServiceClient;
import org.apache.custos.iam.service.FindUsersRequest;
import org.apache.custos.iam.service.RegisterUserRequest;
import org.apache.custos.iam.service.ResetUserPassword;
import org.apache.custos.iam.service.UserSearchRequest;
import org.apache.custos.integration.core.exceptions.NotAuthorizedException;
import org.apache.custos.integration.services.commons.interceptors.AuthInterceptor;
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.apache.custos.user.management.service.UserProfileRequest;
import org.apache.custos.user.profile.service.GetUpdateAuditTrailRequest;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

/**
 * Responsible for validate confidential client specific authorization.
 * Methods which authenticates based only on client are implemented here.
 */
@Component
public class ClientAuthInterceptorImpl extends AuthInterceptor {
    private static final Logger LOGGER = LoggerFactory.getLogger(ClientAuthInterceptorImpl.class);

    @Autowired
    public ClientAuthInterceptorImpl(CredentialStoreServiceClient credentialStoreServiceClient, TenantProfileClient tenantProfileClient) {
        super(credentialStoreServiceClient, tenantProfileClient);
    }

    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT reqT) {


        if (method.equals("deleteUserProfile")) {
            AuthClaim claim = authorize(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();

            long tenantId = claim.getTenantId();
            return (ReqT) ((UserProfileRequest) reqT).toBuilder()
                    .setClientId(oauthId)
                    .setClientSecret(oauthSec)
                    .setTenantId(tenantId)
                    .build();

        } else if (method.equals("registerUser")) {
            AuthClaim claim = authorize(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();

            long tenantId = claim.getTenantId();
            org.apache.custos.iam.service.RegisterUserRequest registerUserRequest =
                    ((RegisterUserRequest) reqT).toBuilder()
                            .setTenantId(tenantId)
                            .setClientId(oauthId)
                            .setClientSec(oauthSec)
                            .build();

            return (ReqT) registerUserRequest;
        } else if (method.equals("enableUser") || method.equals("isUserEnabled") || method.equals("isUsernameAvailable")) {
            AuthClaim claim = authorize(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();

            long tenantId = claim.getTenantId();
            UserSearchRequest info = ((UserSearchRequest) reqT)
                    .toBuilder()
                    .setClientId(oauthId)
                    .setClientSec(oauthSec)
                    .setTenantId(tenantId)
                    .build();

            return (ReqT) info;

        } else if (method.equals("getUserProfile")) {
            AuthClaim claim = authorize(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();

            long tenantId = claim.getTenantId();
            UserProfileRequest request = ((UserProfileRequest) reqT)
                    .toBuilder()
                    .setTenantId(tenantId).build();

            return (ReqT) request;
        } else if (method.equals("getAllUserProfilesInTenant")) {
            AuthClaim claim = authorize(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();

            long tenantId = claim.getTenantId();
            UserProfileRequest request = ((UserProfileRequest) reqT)
                    .toBuilder().setTenantId(tenantId).build();

            return (ReqT) request;
        } else if (method.equals("getUserProfileAuditTrails")) {
            AuthClaim claim = authorize(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();

            long tenantId = claim.getTenantId();
            GetUpdateAuditTrailRequest request = ((GetUpdateAuditTrailRequest) reqT)
                    .toBuilder()
                    .setTenantId(tenantId)
                    .build();

            return (ReqT) request;
        } else if (method.equals("resetPassword")) {
            AuthClaim claim = authorize(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();

            long tenantId = claim.getTenantId();

            ResetUserPassword request = ((ResetUserPassword) reqT)
                    .toBuilder()
                    .setClientId(oauthId)
                    .setClientSec(oauthSec)
                    .setTenantId(tenantId)
                    .build();

            return (ReqT) request;
        } else if (method.equals("getUser")) {
            AuthClaim claim = authorize(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();

            long tenantId = claim.getTenantId();
            UserSearchRequest request = ((UserSearchRequest) reqT)
                    .toBuilder()
                    .setClientId(oauthId)
                    .setTenantId(tenantId)
                    .setClientSec(oauthSec)
                    .build();
            return (ReqT) request;

        } else if (method.equals("findUsers")) {
            AuthClaim claim = authorize(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();

            long tenantId = claim.getTenantId();
            FindUsersRequest request = ((FindUsersRequest) reqT)
                    .toBuilder()
                    .setClientId(oauthId)
                    .setClientSec(oauthSec)
                    .setTenantId(tenantId).build();


            return (ReqT) request;
        }
        return reqT;
    }

}
