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
import org.apache.custos.identity.service.AuthToken;
import org.apache.custos.integration.core.exceptions.NotAuthorizedException;
import org.apache.custos.integration.core.utils.Constants;
import org.apache.custos.integration.services.commons.interceptors.MultiTenantAuthInterceptor;
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
public class UserAuthInterceptorImpl extends MultiTenantAuthInterceptor {
    private static final Logger LOGGER = LoggerFactory.getLogger(ClientAuthInterceptorImpl.class);

    @Autowired
    public UserAuthInterceptorImpl(CredentialStoreServiceClient credentialStoreServiceClient, TenantProfileClient tenantProfileClient, IdentityClient identityClient) {
        super(credentialStoreServiceClient, tenantProfileClient, identityClient);
    }

    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT msg) {


        if (method.equals("addUserAttributes")) {
            AddUserAttributesRequest userAttributesRequest = (AddUserAttributesRequest) msg;
            headers =  attachUserToken(headers, userAttributesRequest.getClientId());
            AuthClaim claim = authorize(headers, userAttributesRequest.getClientId());

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();

            long tenantId = claim.getTenantId();

            AuthToken token = getSAToken(claim.getIamAuthId(), claim.getIamAuthSecret(), claim.getTenantId());
            if (token == null || token.getAccessToken() == null) {
                throw new NotAuthorizedException("Request is not authorized SA token is invalid", null);
            }


            return (ReqT) ((AddUserAttributesRequest) msg).toBuilder()
                    .setClientId(oauthId)
                    .setTenantId(tenantId)
                    .setAccessToken(token.getAccessToken())
                    .setPerformedBy(claim.getPerformedBy())
                    .build();

        } else if (method.equals("deleteUserAttributes")) {

            DeleteUserAttributeRequest userAttributesRequest = (DeleteUserAttributeRequest) msg;
            headers =  attachUserToken(headers, userAttributesRequest.getClientId());
            AuthClaim claim = authorize(headers, userAttributesRequest.getClientId());


            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();

            long tenantId = claim.getTenantId();

            AuthToken token = getSAToken(claim.getIamAuthId(), claim.getIamAuthSecret(), claim.getTenantId());
            if (token == null || token.getAccessToken() == null) {
                throw new NotAuthorizedException("Request is not authorized SA token is invalid", null);
            }

            return (ReqT) ((DeleteUserAttributeRequest) msg).toBuilder()
                    .setClientId(oauthId)
                    .setTenantId(tenantId)
                    .setAccessToken(token.getAccessToken())
                    .setPerformedBy(claim.getPerformedBy())
                    .build();

        } else if (method.equals("addRolesToUsers")) {

            AddUserRolesRequest userAttributesRequest = (AddUserRolesRequest) msg;
            headers =  attachUserToken(headers, userAttributesRequest.getClientId());
            AuthClaim claim = authorize(headers, userAttributesRequest.getClientId());


            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();

            long tenantId = claim.getTenantId();

            String userToken = getUserTokenFromUserTokenHeader(headers);

            if (userToken == null || userToken.trim().equals("")) {
                userToken = getToken(headers);
            }


            return (ReqT) ((AddUserRolesRequest) msg).toBuilder()
                    .setClientId(oauthId)
                    .setTenantId(tenantId)
                    .setAccessToken(userToken)
                    .setPerformedBy(claim.getPerformedBy())
                    .build();

        } else if (method.equals("registerAndEnableUsers")) {

            RegisterUsersRequest registerUsersRequest = (RegisterUsersRequest) msg;
            headers =  attachUserToken(headers, registerUsersRequest.getClientId());
            AuthClaim claim = authorize(headers, registerUsersRequest.getClientId());

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();
            String userToken = getUserTokenFromUserTokenHeader(headers);

            if (userToken == null || userToken.trim().equals("")) {
                userToken = getToken(headers);
            }

            long tenantId = claim.getTenantId();
            org.apache.custos.iam.service.RegisterUsersRequest registerUserRequest =
                    ((RegisterUsersRequest) msg).toBuilder()
                            .setTenantId(tenantId)
                            .setClientId(oauthId)
                            .setAccessToken(userToken)
                            .setPerformedBy(claim.getPerformedBy())
                            .build();
            return (ReqT) registerUserRequest;
        } else if (method.equals("deleteUserRoles")) {

            DeleteUserRolesRequest deleteUserRolesRequest = (DeleteUserRolesRequest) msg;
            headers =  attachUserToken(headers, deleteUserRolesRequest.getClientId());
            AuthClaim claim = authorize(headers, deleteUserRolesRequest.getClientId());

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();
            String userToken = getUserTokenFromUserTokenHeader(headers);

            if (userToken == null || userToken.trim().equals("")) {
                userToken = getToken(headers);
            }

            long tenantId = claim.getTenantId();
            DeleteUserRolesRequest operationRequest = ((DeleteUserRolesRequest) msg)
                    .toBuilder()
                    .setClientId(oauthId)
                    .setAccessToken(userToken)
                    .setTenantId(tenantId)
                    .setPerformedBy(claim.getPerformedBy())
                    .build();

            return (ReqT) operationRequest;

        }  else if (method.equals("deleteUser") || method.equals("grantAdminPrivileges") ||
                method.equals("removeAdminPrivileges")) {

            UserSearchRequest userSearchRequest = (UserSearchRequest) msg;
            headers =  attachUserToken(headers, userSearchRequest.getClientId());
            AuthClaim claim = authorize(headers, userSearchRequest.getClientId());

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();

            String userToken = getUserTokenFromUserTokenHeader(headers);

            if (userToken == null || userToken.trim().equals("")) {
                userToken = getToken(headers);
            }


            long tenantId = claim.getTenantId();
            UserSearchRequest operationRequest = ((UserSearchRequest) msg)
                    .toBuilder()
                    .setClientId(oauthId)
                    .setClientSec(oauthId)
                    .setTenantId(tenantId)
                    .setAccessToken(userToken)
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

        }

        return msg;
    }



    private Metadata attachUserToken(Metadata headers, String clientId) {
        if (clientId == null || clientId.trim().equals("")) {
           String formattedUserToken =  getToken(headers);
           headers.put(Metadata.Key.of(Constants.USER_TOKEN,Metadata.ASCII_STRING_MARSHALLER), formattedUserToken);
           return headers;
        }
        return headers;
    }


}
