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

package org.apache.custos.group.management.interceptors;

import io.grpc.Metadata;
import org.apache.custos.credential.store.client.CredentialStoreServiceClient;
import org.apache.custos.iam.service.GroupRequest;
import org.apache.custos.iam.service.GroupsRequest;
import org.apache.custos.iam.service.UserGroupMappingRequest;
import org.apache.custos.identity.client.IdentityClient;
import org.apache.custos.identity.service.AuthToken;
import org.apache.custos.integration.core.exceptions.NotAuthorizedException;
import org.apache.custos.integration.core.utils.Constants;
import org.apache.custos.integration.services.commons.interceptors.MultiTenantAuthInterceptor;
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.apache.custos.user.profile.service.GroupMembership;
import org.apache.custos.user.profile.service.GroupToGroupMembership;
import org.apache.custos.user.profile.service.UserProfileRequest;
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


        if (method.equals("findGroup") || method.equals("getAllGroups")
                || method.equals("updateGroup") || method.equals("deleteGroup")) {
            GroupRequest request = (GroupRequest) reqT;
            AuthClaim claim = authorize(headers, request.getClientId());

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


            return (ReqT) ((GroupRequest) reqT).toBuilder()
                    .setClientId(oauthId)
                    .setTenantId(tenantId)
                    .setClientSec(oauthSec)
                    .setAccessToken(token.getAccessToken())
                    .setPerformedBy(claim.getPerformedBy() != null ? claim.getPerformedBy(): Constants.SYSTEM)
                    .build();

        } else if (method.equals("createGroups")) {
            GroupsRequest request = (GroupsRequest) reqT;
            AuthClaim claim = authorize(headers, request.getClientId());

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


            return (ReqT) ((GroupsRequest) reqT).toBuilder()
                    .setClientId(oauthId)
                    .setTenantId(tenantId)
                    .setClientSec(oauthSec)
                    .setAccessToken(token.getAccessToken())
                    .setPerformedBy(claim.getPerformedBy() != null ? claim.getPerformedBy(): Constants.SYSTEM)
                    .build();

        } else if (method.equals("addUserToGroup") || method.equals("removeUserFromGroup")) {
            UserGroupMappingRequest request = (UserGroupMappingRequest) reqT;
            AuthClaim claim = authorize(headers, request.getClientId());


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


            return (ReqT) ((UserGroupMappingRequest) reqT).toBuilder()
                    .setClientId(oauthId)
                    .setClientSec(oauthSec)
                    .setTenantId(tenantId)
                    .setAccessToken(token.getAccessToken())
                    .setPerformedBy(claim.getPerformedBy() != null ? claim.getPerformedBy(): Constants.SYSTEM)
                    .build();

        } else if (method.equals("addChildGroupToParentGroup") || method.equals("removeChildGroupFromParentGroup")) {
            GroupToGroupMembership groupToGroupMembership = (GroupToGroupMembership)reqT;
            AuthClaim claim = authorize(headers, groupToGroupMembership.getClientId());


            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }
            long tenantId = claim.getTenantId();


            return (ReqT) ((GroupToGroupMembership) reqT).toBuilder()
                    .setTenantId(tenantId)
                    .build();


        } else if (method.equals("getAllGroupsOfUser")) {
            UserProfileRequest request = (UserProfileRequest)reqT;
            AuthClaim claim = authorize(headers,request.getClientId());


            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }
            long tenantId = claim.getTenantId();

            return (ReqT) ((UserProfileRequest) reqT).toBuilder()
                    .setTenantId(tenantId)
                    .build();


        } else if (method.equals("getAllChildUsers") || method.equals("getAllChildGroups")
                || method.equals("getAllParentGroupsOfGroup")) {

            org.apache.custos.user.profile.service.GroupRequest request =
                    (org.apache.custos.user.profile.service.GroupRequest) reqT;
            AuthClaim claim = authorize(headers, request.getClientId());


            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }
            long tenantId = claim.getTenantId();


            return (ReqT) ((org.apache.custos.user.profile.service.GroupRequest) reqT).toBuilder()
                    .setTenantId(tenantId)
                    .build();
        } else if (method.equals("changeUserMembershipType") || method.equals("hasAccess")) {
            GroupMembership request =
                    (GroupMembership) reqT;
            AuthClaim claim = authorize(headers, request.getClientId());


            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }
            long tenantId = claim.getTenantId();

            return (ReqT) ((GroupMembership) reqT).toBuilder()
                    .setTenantId(tenantId)
                    .build();
        }
        return reqT;
    }


}
