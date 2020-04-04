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

package org.apache.custos.agent.management.interceptors;

import io.grpc.Metadata;
import org.apache.custos.agent.management.service.AgentSearchRequest;
import org.apache.custos.credential.store.client.CredentialStoreServiceClient;
import org.apache.custos.iam.service.*;
import org.apache.custos.identity.client.IdentityClient;
import org.apache.custos.integration.core.exceptions.NotAuthorizedException;
import org.apache.custos.integration.services.commons.interceptors.AuthInterceptor;
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
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
        if (method.equals("enableAgents") || method.equals("configureAgentClient")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            long tenantId = claim.getTenantId();


            return (ReqT) ((AgentClientMetadata) msg).toBuilder()
                    .setTenantId(tenantId)
                    .setAccessToken(token)
                    .setPerformedBy(claim.getPerformedBy())
                    .build();
        } else if (method.equals("registerAndEnableAgent")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            long tenantId = claim.getTenantId();

            return (ReqT) ((RegisterUserRequest) msg).toBuilder()
                    .setTenantId(tenantId)
                    .setAccessToken(token)
                    .setPerformedBy(claim.getPerformedBy())
                    .build();

        } else if (method.equals("getAgent") || method.equals("deleteAgent") || method.equals("disableAgent")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            long tenantId = claim.getTenantId();

            return (ReqT) ((AgentSearchRequest) msg).toBuilder()
                    .setTenantId(tenantId)
                    .setAccessToken(token)
                    .setPerformedBy(claim.getPerformedBy())
                    .build();

        } else if (method.equals("addAgentAttributes")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            long tenantId = claim.getTenantId();
            return (ReqT) ((AddUserAttributesRequest) msg).toBuilder()
                    .setTenantId(tenantId)
                    .setAccessToken(token)
                    .setPerformedBy(claim.getPerformedBy())
                    .build();

        } else if (method.equals("deleteAgentAttributes")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            long tenantId = claim.getTenantId();
            return (ReqT) ((DeleteUserAttributeRequest) msg).toBuilder()
                    .setTenantId(tenantId)
                    .setAccessToken(token)
                    .setPerformedBy(claim.getPerformedBy())
                    .build();

        } else if (method.equals("addRolesToAgent")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            long tenantId = claim.getTenantId();
            return (ReqT) ((AddUserRolesRequest) msg).toBuilder()
                    .setTenantId(tenantId)
                    .setAccessToken(token)
                    .setPerformedBy(claim.getPerformedBy())
                    .build();

        } else if (method.equals("deleteRolesFromAgent")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);

            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            long tenantId = claim.getTenantId();
            return (ReqT) ((DeleteUserRolesRequest) msg).toBuilder()
                    .setTenantId(tenantId)
                    .setAccessToken(token)
                    .setPerformedBy(claim.getPerformedBy())
                    .build();

        }

        return msg;
    }


}
