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
import org.apache.custos.iam.service.*;
import org.apache.custos.integration.core.exceptions.NotAuthorizedException;
import org.apache.custos.integration.services.commons.interceptors.AuthInterceptor;
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.apache.custos.user.management.service.UserProfileRequest;
import org.apache.custos.user.profile.service.UserProfile;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

/**
 * Responsible for validate user specific authorization
 * Methods authenticates users access tokens are implemented here
 */
@Component
public class UserAuthInterceptorImpl extends AuthInterceptor {
    private static final Logger LOGGER = LoggerFactory.getLogger(ClientAuthInterceptorImpl.class);

    @Autowired
    public UserAuthInterceptorImpl(CredentialStoreServiceClient credentialStoreServiceClient, TenantProfileClient tenantProfileClient) {
        super(credentialStoreServiceClient, tenantProfileClient);
    }

    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT msg) {


        if (method.equals("addUserAttributes")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);


            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();

            long tenantId = claim.getTenantId();


            return (ReqT) ((AddUserAttributesRequest) msg).toBuilder()
                    .setClientId(oauthId)
                    .setTenantId(tenantId)
                    .setAccessToken(token)
                    .setPerformedBy(claim.getPerformedBy())
                    .build();

        } else if (method.equals("addRolesToUsers")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);


            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();

            long tenantId = claim.getTenantId();


            return (ReqT) ((AddUserRolesRequest) msg).toBuilder()
                    .setClientId(oauthId)
                    .setTenantId(tenantId)
                    .setAccessToken(token)
                    .setPerformedBy(claim.getPerformedBy())
                    .build();

        } else if (method.equals("registerAndEnableUsers")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();

            long tenantId = claim.getTenantId();
            org.apache.custos.iam.service.RegisterUsersRequest registerUserRequest =
                    ((RegisterUsersRequest) msg).toBuilder()
                            .setTenantId(tenantId)
                            .setClientId(oauthId)
                            .setAccessToken(token)
                            .setPerformedBy(claim.getPerformedBy())
                            .build();
            return (ReqT) registerUserRequest;
        } else if (method.equals("getUser")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            long tenantId = claim.getTenantId();
            UserSearchRequest request = ((UserSearchRequest) msg)
                    .toBuilder()
                    .setAccessToken(token)
                    .setClientId(oauthId)
                    .setTenantId(tenantId)
                    .build();
            return (ReqT) request;

        } else if (method.equals("findUsers")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }
            String oauthId = claim.getIamAuthId();
            long tenantId = claim.getTenantId();

            FindUsersRequest request = ((FindUsersRequest) msg)
                    .toBuilder()
                    .setClientId(oauthId)
                    .setAccessToken(token)
                    .setTenantId(tenantId).build();


            return (ReqT) request;
        } else if (method.equals("deleteUserRoles")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();

            long tenantId = claim.getTenantId();
            DeleteUserRolesRequest operationRequest = ((DeleteUserRolesRequest) msg)
                    .toBuilder()
                    .setClientId(oauthId)
                    .setAccessToken(token)
                    .setTenantId(tenantId)
                    .setPerformedBy(claim.getPerformedBy())
                    .build();

            return (ReqT) operationRequest;

        }
        if (method.equals("updateUserProfile")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();

            long tenantId = claim.getTenantId();
            UserProfile profile = ((UserProfileRequest) msg).getUserProfile().toBuilder().setUpdatedBy(claim.getPerformedBy()).build();

            return (ReqT) ((UserProfileRequest) msg).toBuilder()
                    .setAccessToken(token)
                    .setTenantId(tenantId)
                    .setClientId(oauthId)
                    .setUserProfile(profile)
                    .build();

        } else

            return msg;
    }


}
