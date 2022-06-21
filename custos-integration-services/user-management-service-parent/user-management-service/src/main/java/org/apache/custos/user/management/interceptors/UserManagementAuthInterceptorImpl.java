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
import org.apache.custos.integration.core.exceptions.UnAuthorizedException;
import org.apache.custos.integration.core.utils.Constants;
import org.apache.custos.integration.services.commons.interceptors.MultiTenantAuthInterceptor;
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.apache.custos.user.management.service.LinkUserProfileRequest;
import org.apache.custos.user.management.service.UserProfileRequest;
import org.apache.custos.user.profile.service.GetUpdateAuditTrailRequest;
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
public class UserManagementAuthInterceptorImpl extends MultiTenantAuthInterceptor {
    private static final Logger LOGGER = LoggerFactory.getLogger(UserManagementAuthInterceptorImpl.class);

    private CredentialStoreServiceClient credentialStoreServiceClient;

    @Autowired
    public UserManagementAuthInterceptorImpl(CredentialStoreServiceClient credentialStoreServiceClient, TenantProfileClient tenantProfileClient, IdentityClient identityClient) {
        super(credentialStoreServiceClient, tenantProfileClient, identityClient);
        this.credentialStoreServiceClient = credentialStoreServiceClient;
    }

    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT msg) {


        if (method.equals("addUserAttributes")) {
            AddUserAttributesRequest userAttributesRequest = (AddUserAttributesRequest) msg;
            headers = attachUserToken(headers, userAttributesRequest.getClientId());
            Optional<AuthClaim> claim = authorize(headers, userAttributesRequest.getClientId());

            if (claim.isEmpty()) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.get().getIamAuthId();
            long tenantId = claim.get().getTenantId();
            AuthToken token = getSAToken(claim.get().getIamAuthId(),
                    claim.get().getIamAuthSecret(), claim.get().getTenantId());
            if (token == null || token.getAccessToken() == null) {
                throw new UnAuthorizedException("Request is not authorized SA token is invalid", null);
            }

            return (ReqT) ((AddUserAttributesRequest) msg).toBuilder()
                    .setClientId(oauthId)
                    .setTenantId(tenantId)
                    .setAccessToken(token.getAccessToken())
                    .setPerformedBy(claim.get().getPerformedBy())
                    .build();

        } else if (method.equals("deleteUserAttributes")) {

            DeleteUserAttributeRequest userAttributesRequest = (DeleteUserAttributeRequest) msg;
            headers = attachUserToken(headers, userAttributesRequest.getClientId());
            Optional<AuthClaim> claim = authorize(headers, userAttributesRequest.getClientId());


            if (claim.isEmpty()) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.get().getIamAuthId();

            long tenantId = claim.get().getTenantId();

            AuthToken token = getSAToken(claim.get().getIamAuthId(), claim.get().getIamAuthSecret(),
                    claim.get().getTenantId());
            if (token == null || token.getAccessToken() == null) {
                throw new UnAuthorizedException("Request is not authorized SA token is invalid", null);
            }

            return (ReqT) ((DeleteUserAttributeRequest) msg).toBuilder()
                    .setClientId(oauthId)
                    .setTenantId(tenantId)
                    .setAccessToken(token.getAccessToken())
                    .setPerformedBy(claim.get().getPerformedBy())
                    .build();

        } else if (method.equals("addRolesToUsers")) {

            AddUserRolesRequest userAttributesRequest = (AddUserRolesRequest) msg;
            headers = attachUserToken(headers, userAttributesRequest.getClientId());
            Optional<AuthClaim> claim = authorize(headers, userAttributesRequest.getClientId());
            if (claim.isEmpty()) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }


            String oauthId = claim.get().getIamAuthId();

            long tenantId = claim.get().getTenantId();

            AuthToken token = getSAToken(claim.get().getIamAuthId(), claim.get().getIamAuthSecret(),
                    claim.get().getTenantId());
            if (token == null || token.getAccessToken() == null) {
                throw new UnAuthorizedException("Request is not authorized SA token is invalid", null);
            }

            return (ReqT) ((AddUserRolesRequest) msg).toBuilder()
                    .setClientId(oauthId)
                    .setTenantId(tenantId)
                    .setAccessToken(token.getAccessToken())
                    .setPerformedBy(claim.get().getPerformedBy())
                    .build();

        } else if (method.equals("registerAndEnableUsers")) {

            RegisterUsersRequest registerUsersRequest = (RegisterUsersRequest) msg;
            headers = attachUserToken(headers, registerUsersRequest.getClientId());
            Optional<AuthClaim> claim = authorize(headers, registerUsersRequest.getClientId());

            if (claim.isEmpty()) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.get().getIamAuthId();
            String oauthSec = claim.get().getIamAuthSecret();
            Optional<String> userTokenOp = getUserTokenFromUserTokenHeader(headers);

            String userToken = null;

            if (userTokenOp.isEmpty()) {
                userToken = getToken(headers);
            } else {
                userToken = userTokenOp.get();
            }

            long tenantId = claim.get().getTenantId();
            org.apache.custos.iam.service.RegisterUsersRequest registerUserRequest =
                    ((RegisterUsersRequest) msg).toBuilder()
                            .setTenantId(tenantId)
                            .setClientId(oauthId)
                            .setAccessToken(userToken)
                            .setPerformedBy(claim.get().getPerformedBy())
                            .build();
            return (ReqT) registerUserRequest;
        } else if (method.equals("deleteUserRoles")) {

            DeleteUserRolesRequest deleteUserRolesRequest = (DeleteUserRolesRequest) msg;
            headers = attachUserToken(headers, deleteUserRolesRequest.getClientId());
            Optional<AuthClaim> claim = authorize(headers, deleteUserRolesRequest.getClientId());

            if (claim.isEmpty()) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }


            String oauthId = claim.get().getIamAuthId();

            long tenantId = claim.get().getTenantId();

            AuthToken token = getSAToken(claim.get().getIamAuthId(), claim.get().getIamAuthSecret(),
                    claim.get().getTenantId());
            if (token == null || token.getAccessToken() == null) {
                throw new UnAuthorizedException("Request is not authorized SA token is invalid", null);
            }

            DeleteUserRolesRequest operationRequest = ((DeleteUserRolesRequest) msg)
                    .toBuilder()
                    .setClientId(oauthId)
                    .setAccessToken(token.getAccessToken())
                    .setTenantId(tenantId)
                    .setPerformedBy(claim.get().getPerformedBy().isEmpty() ? Constants.SYSTEM : claim.get().getPerformedBy())
                    .build();

            return (ReqT) operationRequest;

        } else if (method.equals("deleteUser") || method.equals("grantAdminPrivileges") ||
                method.equals("removeAdminPrivileges")) {

            UserSearchRequest userSearchRequest = (UserSearchRequest) msg;
            headers = attachUserToken(headers, userSearchRequest.getClientId());
            Optional<AuthClaim> claim =
                    validateRoleManagementAuthorizations(headers, userSearchRequest.getClientId());
            String oauthId = claim.get().getIamAuthId();
            String oauthSec = claim.get().getIamAuthSecret();

            AuthToken token = getSAToken(claim.get().getIamAuthId(), claim.get().getIamAuthSecret(), claim.get().getTenantId());
            if (token == null || token.getAccessToken() == null) {
                throw new UnAuthorizedException("Request is not authorized SA token is invalid", null);
            }


            long tenantId = claim.get().getTenantId();
            UserSearchRequest operationRequest = ((UserSearchRequest) msg)
                    .toBuilder()
                    .setClientId(oauthId)
                    .setClientSec(oauthSec)
                    .setTenantId(tenantId)
                    .setAccessToken(token.getAccessToken())
                    .setPerformedBy(Constants.SYSTEM)
                    .build();

            return (ReqT) operationRequest;

        } else if (method.equals("linkUserProfile")) {
            String token = getToken(headers);
            Optional<AuthClaim> claim = authorizeUsingUserToken(headers);

            if (claim.isEmpty()) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.get().getIamAuthId();
            String oauthSec = claim.get().getIamAuthSecret();

            long tenantId = claim.get().getTenantId();
            LinkUserProfileRequest operationRequest = ((LinkUserProfileRequest) msg)
                    .toBuilder()
                    .setIamClientId(oauthId)
                    .setIamClientSecret(oauthSec)
                    .setTenantId(tenantId)
                    .setAccessToken(token)
                    .setPerformedBy(claim.get().getPerformedBy())
                    .build();

            return (ReqT) operationRequest;

        } else if (method.equals("deleteUserProfile")) {
            UserProfileRequest request = (UserProfileRequest) msg;
            Optional<AuthClaim> claim = authorize(headers, request.getClientId());
            return claim.map(cl -> {
                String oauthId = cl.getIamAuthId();
                String oauthSec = cl.getIamAuthSecret();

                long tenantId = cl.getTenantId();
                return (ReqT) ((UserProfileRequest) msg).toBuilder()
                        .setClientId(oauthId)
                        .setClientSecret(oauthSec)
                        .setTenantId(tenantId)
                        .build();
            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });

        } else if (method.equals("registerUser")) {

            RegisterUserRequest request = (RegisterUserRequest) msg;
            Optional<AuthClaim> claim = authorize(headers, request.getClientId());
            return claim.map(cl -> {
                String oauthId = cl.getIamAuthId();
                String oauthSec = cl.getIamAuthSecret();

                long tenantId = cl.getTenantId();
                org.apache.custos.iam.service.RegisterUserRequest registerUserRequest =
                        ((RegisterUserRequest) msg).toBuilder()
                                .setTenantId(tenantId)
                                .setClientId(oauthId)
                                .setClientSec(oauthSec)
                                .build();
                return (ReqT) registerUserRequest;
            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });

        } else if (method.equals("enableUser") || method.equals("disableUser") ||
                method.equals("isUserEnabled") || method.equals("isUsernameAvailable")) {
            UserSearchRequest request = (UserSearchRequest) msg;
            Optional<AuthClaim> claim = authorize(headers, request.getClientId());
            return claim.map(cl -> {
                String oauthId = cl.getIamAuthId();
                String oauthSec = cl.getIamAuthSecret();

                long tenantId = cl.getTenantId();
                UserSearchRequest info = ((UserSearchRequest) msg)
                        .toBuilder()
                        .setClientId(oauthId)
                        .setClientSec(oauthSec)
                        .setTenantId(tenantId)
                        .build();

                return (ReqT) info;
            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });

        } else if (method.equals("getUserProfile")) {

            UserProfileRequest req = (UserProfileRequest) msg;
            Optional<AuthClaim> claim = authorize(headers, req.getClientId());
            return claim.map(cl -> {
                long tenantId = cl.getTenantId();
                UserProfileRequest request = ((UserProfileRequest) msg)
                        .toBuilder()
                        .setTenantId(tenantId).build();

                return (ReqT) request;
            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });
        } else if (method.equals("getAllUserProfilesInTenant")) {
            UserProfileRequest req = (UserProfileRequest) msg;
            Optional<AuthClaim> claim = authorize(headers, req.getClientId());
            return claim.map(cl -> {
                long tenantId = cl.getTenantId();
                UserProfileRequest request = ((UserProfileRequest) msg)
                        .toBuilder().setTenantId(tenantId).build();

                return (ReqT) request;

            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });

        } else if (method.equals("getUserProfileAuditTrails")) {
            Optional<AuthClaim> claim = authorize(headers);
            return claim.map(cl -> {
                long tenantId = cl.getTenantId();
                GetUpdateAuditTrailRequest request = ((GetUpdateAuditTrailRequest) msg)
                        .toBuilder()
                        .setTenantId(tenantId)
                        .build();

                return (ReqT) request;

            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });

        } else if (method.equals("resetPassword")) {
            ResetUserPassword req = (ResetUserPassword) msg;
            Optional<AuthClaim> claim = authorize(headers, req.getClientId());
            return claim.map(cl -> {
                String oauthId = cl.getIamAuthId();
                String oauthSec = cl.getIamAuthSecret();

                long tenantId = cl.getTenantId();

                ResetUserPassword request = ((ResetUserPassword) msg)
                        .toBuilder()
                        .setClientId(oauthId)
                        .setClientSec(oauthSec)
                        .setTenantId(tenantId)
                        .build();

                return (ReqT) request;

            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });

        } else if (method.equals("getUser")) {
            UserSearchRequest req = (UserSearchRequest) msg;
            Optional<AuthClaim> claim = authorize(headers, req.getClientId());
            return claim.map(cl -> {
                String oauthId = cl.getIamAuthId();
                String oauthSec = cl.getIamAuthSecret();

                long tenantId = cl.getTenantId();
                UserSearchRequest request = ((UserSearchRequest) msg)
                        .toBuilder()
                        .setClientId(oauthId)
                        .setTenantId(tenantId)
                        .setClientSec(oauthSec)
                        .build();
                return (ReqT) request;
            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });


        } else if (method.equals("findUsers")) {
            FindUsersRequest req = (FindUsersRequest) msg;
            Optional<AuthClaim> claim = authorize(headers, req.getClientId());

            return claim.map(cl -> {
                String oauthId = cl.getIamAuthId();
                String oauthSec = cl.getIamAuthSecret();

                long tenantId = cl.getTenantId();
                FindUsersRequest request = ((FindUsersRequest) msg)
                        .toBuilder()
                        .setClientId(oauthId)
                        .setClientSec(oauthSec)
                        .setTenantId(tenantId).build();

                return (ReqT) request;

            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });

        } else if (method.equals("updateUserProfile")) {

            UserProfileRequest userProfileRequest = (UserProfileRequest) msg;

            Optional<AuthClaim> claim = authorize(headers, userProfileRequest.getClientId());

            if (claim.isEmpty()) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.get().getIamAuthId();
            String oauthSec = claim.get().getIamAuthSecret();

            long tenantId = claim.get().getTenantId();

            AuthToken token = getSAToken(claim.get().getIamAuthId(), claim.get().getIamAuthSecret(), claim.get().getTenantId());
            if (token == null || token.getAccessToken() == null) {
                throw new UnAuthorizedException("Request is not authorized SA token is invalid", null);
            }

            return (ReqT) ((UserProfileRequest) msg).toBuilder()
                    .setAccessToken(token.getAccessToken())
                    .setTenantId(tenantId)
                    .setClientId(oauthId)
                    .setClientSecret(oauthSec)
                    .setPerformedBy(Constants.SYSTEM)
                    .build();

        } else if (method.equals("deleteExternalIDPsOfUsers")) {
            DeleteExternalIDPsRequest deleteExternalIDPsRequest = (DeleteExternalIDPsRequest) msg;

            Optional<AuthClaim> claim = authorize(headers, deleteExternalIDPsRequest.getClientId());

            if (claim.isEmpty()) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }
            String oauthId = claim.get().getIamAuthId();
            long tenantId = claim.get().getTenantId();

            return (ReqT) ((DeleteExternalIDPsRequest) msg).toBuilder()
                    .setTenantId(tenantId)
                    .setClientId(oauthId)
                    .build();
        } else if (method.equals("getExternalIDPsOfUsers")) {
            GetExternalIDPsRequest getExternalIDPsRequest = (GetExternalIDPsRequest) msg;

            Optional<AuthClaim> claim = authorize(headers, getExternalIDPsRequest.getClientId());

            if (claim.isEmpty()) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }
            String oauthId = claim.get().getIamAuthId();
            long tenantId = claim.get().getTenantId();

            return (ReqT) ((GetExternalIDPsRequest) msg).toBuilder()
                    .setTenantId(tenantId)
                    .setClientId(oauthId)
                    .build();
        } else if (method.equals("addExternalIDPsOfUsers")) {
            AddExternalIDPLinksRequest getExternalIDPsRequest = (AddExternalIDPLinksRequest) msg;

            Optional<AuthClaim> claim = authorize(headers, getExternalIDPsRequest.getClientId());

            if (claim.isEmpty()) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }
            String oauthId = claim.get().getIamAuthId();
            long tenantId = claim.get().getTenantId();

            return (ReqT) ((AddExternalIDPLinksRequest) msg).toBuilder()
                    .setTenantId(tenantId)
                    .setClientId(oauthId)
                    .build();
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

    private Optional<AuthClaim> validateRoleManagementAuthorizations(Metadata headers, String clientId) {
//        Optional<AuthClaim> parentClaim = authorizeUsingUserToken(headers);
        Optional<AuthClaim> claim = authorize(headers, clientId);

//        if (claim.isEmpty() || parentClaim.isEmpty() || !parentClaim.get().isAdmin()) {
//            throw new UnAuthorizedException("Request is not authorized", null);
//        }
        return claim;
    }


}
