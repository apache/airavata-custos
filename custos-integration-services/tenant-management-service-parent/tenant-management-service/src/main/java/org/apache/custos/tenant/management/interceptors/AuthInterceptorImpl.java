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

package org.apache.custos.tenant.management.interceptors;

import io.grpc.Metadata;
import org.apache.custos.credential.store.client.CredentialStoreServiceClient;
import org.apache.custos.integration.core.exceptions.NotAuthorizedException;
import org.apache.custos.integration.services.commons.interceptors.AuthInterceptor;
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.tenant.management.service.*;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.apache.custos.tenant.profile.service.Tenant;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

/**
 * This validates custos credentials
 */
@Component
public class AuthInterceptorImpl extends AuthInterceptor {

    private static final Logger LOGGER = LoggerFactory.getLogger(AuthInterceptorImpl.class);

    private CredentialStoreServiceClient credentialStoreServiceClient;

    public AuthInterceptorImpl(CredentialStoreServiceClient credentialStoreServiceClient, TenantProfileClient tenantProfileClient) {
        super(credentialStoreServiceClient, tenantProfileClient);
        this.credentialStoreServiceClient = credentialStoreServiceClient;
    }


    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT msg) {


        if (method.equals("getCredentials")) {
            AuthClaim claim = authorize(headers);
            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }
            long tenantId = claim.getTenantId();


            return (ReqT) GetCredentialsRequest.newBuilder().setTenantId(tenantId).build();
        } else if (method.equals("createTenant")) {
            AuthClaim claim = authorize(headers);
            if (claim != null) {

                return (ReqT) ((Tenant) msg).toBuilder().setParentTenantId(claim.getTenantId()).build();
            }
        } else if (method.equals("getTenant")) {

            AuthClaim claim = authorize(headers);
            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            GetTenantRequest tenantRequest = ((GetTenantRequest) msg);

            Credentials credentials = getCredentials(claim);


            return (ReqT) tenantRequest.toBuilder()
                    .setTenantId(claim.getTenantId()).setCredentials(credentials).build();
        } else if (method.equals("updateTenant")) {


            AuthClaim claim = authorize(headers);
            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            UpdateTenantRequest tenantRequest = ((UpdateTenantRequest) msg);

            Credentials credentials = getCredentials(claim);

            return (ReqT) tenantRequest.toBuilder()
                    .setTenantId(claim.getTenantId()).setCredentials(credentials).build();
        } else if (method.equals("deleteTenant")) {

            AuthClaim claim = authorize(headers);
            if (claim == null) {
                throw new NotAuthorizedException("Request is not authorized", null);
            }

            DeleteTenantRequest tenantRequest = ((DeleteTenantRequest) msg);

            Credentials credentials = getCredentials(claim);


            return (ReqT) tenantRequest.toBuilder()
                    .setTenantId(claim.getTenantId()).setCredentials(credentials).build();
        }

        return msg;
    }


    private Credentials getCredentials(AuthClaim claim) {
        return Credentials.newBuilder()
                .setCustosClientId(claim.getCustosId())
                .setCustosClientSecret(claim.getCustosSecret())
                .setCustosClientIdIssuedAt(claim.getCustosIdIssuedAt())
                .setCustosClientSecretExpiredAt(claim.getCustosSecretExpiredAt())
                .setIamClientId(claim.getIamAuthId())
                .setIamClientSecret(claim.getIamAuthSecret())
                .setCiLogonClientId(claim.getCiLogonId())
                .setCiLogonClientSecret(claim.getCiLogonSecret()).build();

    }
}
