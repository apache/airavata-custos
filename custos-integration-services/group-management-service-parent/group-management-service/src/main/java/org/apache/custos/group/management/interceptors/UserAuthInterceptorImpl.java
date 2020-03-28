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


        if (method.equals("createGroups")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);


            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();

            long tenantId = claim.getTenantId();


            return (ReqT) ((GroupsRequest) msg).toBuilder()
                    .setClientId(oauthId)
                    .setTenantId(tenantId)
                    .setAccessToken(token)
                    .setPerformedBy(claim.getPerformedBy())
                    .build();

        } else if (method.equals("updateGroup")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);


            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();

            long tenantId = claim.getTenantId();


            return (ReqT) ((GroupRequest) msg).toBuilder()
                    .setClientId(oauthId)
                    .setTenantId(tenantId)
                    .setAccessToken(token)
                    .setPerformedBy(claim.getPerformedBy())
                    .build();

        } else if (method.equals("deleteGroup")) {
            String token = getToken(headers);
            AuthClaim claim = authorizeUsingUserToken(headers);


            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();

            long tenantId = claim.getTenantId();


            return (ReqT) ((GroupRequest) msg).toBuilder()
                    .setClientId(oauthId)
                    .setTenantId(tenantId)
                    .setAccessToken(token)
                    .setPerformedBy(claim.getPerformedBy())
                    .build();

        }

        return msg;
    }


}
