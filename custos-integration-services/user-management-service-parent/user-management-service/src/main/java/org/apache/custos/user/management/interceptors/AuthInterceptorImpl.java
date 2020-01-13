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
import org.apache.custos.integration.core.exceptions.NotAuthorizedException;
import org.apache.custos.integration.services.commons.interceptors.AuthInterceptor;
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.apache.custos.user.management.service.*;
import org.apache.custos.user.profile.service.*;
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
        LOGGER.info("In authinterceptorImpl");

        AuthClaim claim = authorize(headers);

        String oauthId = claim.getIamAuthId();
        String oauthSec = claim.getIamAuthSecret();

        long tenantId = claim.getTenantId();

        if (claim == null) {
            throw new NotAuthorizedException("Request is not authorized", null);
        }
        if (method.equals("updateUserProfile")) {
            UserProfile pr = ((UserProfileRequest) reqT).getUserProfile().toBuilder().setTenantId(tenantId).build();

            return (ReqT) ((UserProfileRequest) reqT).toBuilder()
                    .setIamClientId(oauthId).setUserProfile(pr).setIamClientSecret(oauthSec).build();

        } else if (method.equals("deleteUserProfile")) {
            DeleteUserProfileRequest pr = ((DeleteProfileRequest) reqT).getDeleteRequest().toBuilder().setTenantId(tenantId).build();
            return (ReqT) ((DeleteProfileRequest) reqT).toBuilder()
                    .setIamClientId(oauthId).setDeleteRequest(pr).setIamClientSecret(oauthSec).build();

        } else if (method.equals("registerUser")) {
            org.apache.custos.iam.service.RegisterUserRequest registerUserRequest =
                    ((RegisterUserRequest) reqT).getUserRequest().toBuilder().setTenantId(tenantId).build();


            RegisterUserRequest pr = ((RegisterUserRequest) reqT).toBuilder()
                    .setIamClientId(oauthId)
                    .setIamClientSecret(oauthSec)
                    .setUserRequest(registerUserRequest)
                    .build();

            return (ReqT) pr;
        } else if (method.equals("enableUser") || method.equals("deleteUser")) {

            UserIdentificationRequest info = ((UserIdentificationRequest) reqT)
                    .toBuilder()
                    .setInfo(((UserIdentificationRequest) reqT).getInfo().toBuilder().setTenantId(tenantId).build())
                    .setIamClientId(oauthId)
                    .setIamClientSecret(oauthSec)
                    .build();

            return (ReqT) info;
        } else if (method.equals("getUser") || method.equals("isUserEnabled")) {

            GetUserRequest info = ((GetUserRequest) reqT)
                    .toBuilder()
                    .setIamClientId(oauthId)
                    .setIamClientSecret(oauthSec)
                    .setTenantId(tenantId)
                    .build();

            return (ReqT) info;


        } else if (method.equals("getUsers")) {

            GetUsersRequest request = ((GetUsersRequest) reqT)
                    .toBuilder()
                    .setIamClientId(oauthId)
                    .setIamClientSecret(oauthSec)
                    .setTenantId(tenantId).build();


            return (ReqT) request;
        } else if (method.equals("addRoleToUser") || method.equals("deleteRoleFromUser")) {

            RoleOperationRequest operationRequest = ((RoleOperationRequest) reqT)
                    .toBuilder()
                    .setRole(((RoleOperationRequest) reqT)
                            .getRole().toBuilder().setTenantId(tenantId).build())
                    .build();

            return (ReqT) operationRequest;

        } else if (method.equals("getUserProfile")) {

            GetUserProfileRequest request = ((GetUserProfileRequest) reqT)
                    .toBuilder()
                    .setTenantId(tenantId).build();

            return (ReqT) request;
        } else if (method.equals("getAllUserProfilesInTenant")) {

            GetAllUserProfilesRequest request = ((GetAllUserProfilesRequest) reqT)
                    .toBuilder().setTenantId(tenantId).build();

            return (ReqT) request;
        } else if (method.equals("getUserProfileAuditTrails")) {

            GetUpdateAuditTrailRequest request = ((GetUpdateAuditTrailRequest) reqT)
                    .toBuilder()
                    .setTenantId(tenantId)
                    .build();

            return (ReqT) request;
        }

        return reqT;
    }

}
