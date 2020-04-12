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
import org.apache.custos.identity.client.IdentityClient;
import org.apache.custos.integration.core.exceptions.NotAuthorizedException;
import org.apache.custos.integration.services.commons.interceptors.AuthInterceptor;
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.apache.custos.user.management.service.LinkUserProfileRequest;
import org.apache.custos.user.management.service.UserProfileRequest;
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
    public UserAuthInterceptorImpl(CredentialStoreServiceClient credentialStoreServiceClient, TenantProfileClient tenantProfileClient, IdentityClient identityClient) {
        super(credentialStoreServiceClient, tenantProfileClient, identityClient);
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

        }  else if (method.equals("deleteUserAttributes")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);


            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();

            long tenantId = claim.getTenantId();


            return (ReqT) ((DeleteUserAttributeRequest) msg).toBuilder()
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
        else if (method.equals("updateUserProfile")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();

            long tenantId = claim.getTenantId();

            return (ReqT) ((UserProfileRequest) msg).toBuilder()
                    .setAccessToken(token)
                    .setTenantId(tenantId)
                    .setClientId(oauthId)
                    .setClientSecret(oauthSec)
                    .setPerformedBy(claim.getPerformedBy())
                    .build();

        } else if (method.equals("deleteUser") || method.equals("grantAdminPrivileges") ||
                method.equals("removeAdminPrivileges")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();

            long tenantId = claim.getTenantId();
            UserSearchRequest operationRequest = ((UserSearchRequest) msg)
                    .toBuilder()
                    .setClientId(oauthId)
                    .setClientSec(oauthId)
                    .setTenantId(tenantId)
                    .setAccessToken(token)
                    .setPerformedBy(claim.getPerformedBy())
                    .build();

            return (ReqT) operationRequest;

        } else if (method.equals("linkUserProfile")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();

            long tenantId = claim.getTenantId();
            LinkUserProfileRequest operationRequest = ((LinkUserProfileRequest) msg)
                    .toBuilder()
                    .setIamClientId(oauthId)
                    .setIamClientSecret(oauthSec)
                    .setTenantId(tenantId)
                    .setAccessToken(token)
                    .setPerformedBy(claim.getPerformedBy())
                    .build();

            return (ReqT) operationRequest;

        } else if (method.equals("createGroups")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            long tenantId = claim.getTenantId();
            GroupsRequest operationRequest = ((GroupsRequest) msg)
                    .toBuilder()
                    .setTenantId(tenantId)
                    .setAccessToken(token)
                    .setClientId(claim.getIamAuthId())
                    .setPerformedBy(claim.getPerformedBy())
                    .build();

            return (ReqT) operationRequest;

        }

            return msg;
    }


}
