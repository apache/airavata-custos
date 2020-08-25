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
import org.apache.custos.identity.client.IdentityClient;
import org.apache.custos.identity.service.AuthToken;
import org.apache.custos.integration.core.exceptions.NotAuthorizedException;
import org.apache.custos.integration.core.utils.Constants;
import org.apache.custos.integration.services.commons.interceptors.MultiTenantAuthInterceptor;
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
public class ClientAuthInterceptorImpl extends MultiTenantAuthInterceptor {
    private static final Logger LOGGER = LoggerFactory.getLogger(ClientAuthInterceptorImpl.class);

    @Autowired
    public ClientAuthInterceptorImpl(CredentialStoreServiceClient credentialStoreServiceClient, TenantProfileClient tenantProfileClient, IdentityClient identityClient) {
        super(credentialStoreServiceClient, tenantProfileClient, identityClient);
    }

    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT reqT) {


        if (method.equals("deleteUserProfile")) {
            UserProfileRequest request = (UserProfileRequest) reqT;
            AuthClaim claim = authorize(headers, request.getClientId());

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

            RegisterUserRequest request = (RegisterUserRequest) reqT;
            AuthClaim claim = authorize(headers, request.getClientId());

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
        } else if (method.equals("enableUser") || method.equals("disableUser") ||
                method.equals("isUserEnabled") || method.equals("isUsernameAvailable")) {
            UserSearchRequest request = (UserSearchRequest) reqT;
            AuthClaim claim = authorize(headers, request.getClientId());

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

            UserProfileRequest req = (UserProfileRequest) reqT;
            AuthClaim claim = authorize(headers, req.getClientId());

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
            UserProfileRequest req = (UserProfileRequest) reqT;
            AuthClaim claim = authorize(headers, req.getClientId());

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
            ResetUserPassword req = (ResetUserPassword) reqT;
            AuthClaim claim = authorize(headers, req.getClientId());

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
            UserSearchRequest req = (UserSearchRequest) reqT;

            AuthClaim claim = authorize(headers, req.getClientId());

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
            FindUsersRequest req = (FindUsersRequest) reqT;
            AuthClaim claim = authorize(headers, req.getClientId());

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
        } else if (method.equals("updateUserProfile")) {

            UserProfileRequest userProfileRequest = (UserProfileRequest) reqT;

            AuthClaim claim = authorize(headers, userProfileRequest.getClientId());

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();

            long tenantId = claim.getTenantId();

            AuthToken token = getSAToken(claim.getIamAuthId(), claim.getIamAuthSecret(), claim.getTenantId());
            if (token == null || token.getAccessToken() == null) {
                throw new NotAuthorizedException("Request is not authorized SA token is invalid", null);
            }

            return (ReqT) ((UserProfileRequest) reqT).toBuilder()
                    .setAccessToken(token.getAccessToken())
                    .setTenantId(tenantId)
                    .setClientId(oauthId)
                    .setClientSecret(oauthSec)
                    .setPerformedBy(Constants.SYSTEM)
                    .build();

        }
        return reqT;
    }

}
