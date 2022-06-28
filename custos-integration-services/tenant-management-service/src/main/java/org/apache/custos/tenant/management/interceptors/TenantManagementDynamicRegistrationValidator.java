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
import org.apache.custos.credential.store.service.CredentialMetadata;
import org.apache.custos.identity.client.IdentityClient;
import org.apache.custos.integration.core.exceptions.UnAuthorizedException;
import org.apache.custos.integration.services.commons.interceptors.AuthInterceptor;
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.tenant.management.service.DeleteTenantRequest;
import org.apache.custos.tenant.management.service.GetTenantRequest;
import org.apache.custos.tenant.management.service.UpdateTenantRequest;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.apache.custos.tenant.profile.service.Tenant;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.util.Optional;

/**
 * This class validates the  conditions that should be satisfied by the Dynamic Registration Protocol
 */
@Component
public class TenantManagementDynamicRegistrationValidator extends AuthInterceptor {

    private static final Logger LOGGER = LoggerFactory.getLogger(TenantManagementDynamicRegistrationValidator.class);

    private CredentialStoreServiceClient credentialStoreServiceClient;
    private TenantProfileClient tenantProfileClient;

    public TenantManagementDynamicRegistrationValidator(CredentialStoreServiceClient credentialStoreServiceClient,
                                                        TenantProfileClient tenantProfileClient, IdentityClient identityClient) {
        super(credentialStoreServiceClient, tenantProfileClient, identityClient);
        this.credentialStoreServiceClient = credentialStoreServiceClient;
        this.tenantProfileClient = tenantProfileClient;
    }


    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT msg) {


        if (method.equals("getTenant")) {
            GetTenantRequest tenantRequest = ((GetTenantRequest) msg);

            String clientId = tenantRequest.getClientId();

            CredentialMetadata metadata = getCredentialsFromClientId(clientId);

            Tenant tenant = validateTenantWithUserToken(metadata.getOwnerId(), tenantRequest.getTenantId(), headers);
            return (ReqT) tenantRequest.toBuilder().setTenantId(tenant != null ? tenant.getTenantId() :
                    tenantRequest.getTenantId()).setTenant(tenant).build();

        } else if (method.equals("createTenant")) {

            Tenant tenantRequest = ((Tenant) msg);

            if (tenantRequest.getParentTenantId() > 0) {
                validateTenantWithBasicAuth(tenantRequest.getParentTenantId(), tenantRequest.getParentTenantId(), headers);
            }

        } else if (method.equals("updateTenant")) {


            UpdateTenantRequest tenantRequest = ((UpdateTenantRequest) msg);

            String clientId = tenantRequest.getClientId();

            if (clientId == null || clientId.trim().equals("")) {
                clientId = tenantRequest.getBody().getClientId();
            }

            CredentialMetadata metadata = getCredentialsFromClientId(clientId);


            Tenant tenant = validateTenantWithUserToken(metadata.getOwnerId(), tenantRequest.getTenantId(), headers);

            return (ReqT) tenantRequest.toBuilder().setTenantId(tenant.getTenantId()).setClientId(clientId).build();

        } else if (method.equals("deleteTenant")) {


            DeleteTenantRequest tenantRequest = ((DeleteTenantRequest) msg);

            String clientId = tenantRequest.getClientId();

            CredentialMetadata metadata = getCredentialsFromClientId(clientId);

            Tenant tenant = validateTenantWithUserToken(metadata.getOwnerId(), tenantRequest.getTenantId(), headers);

            return (ReqT) tenantRequest.toBuilder().setTenantId(tenant.getTenantId()).build();

        }
        return msg;

    }


    private Tenant validateTenantWithBasicAuth(long ownerId, long parentTenant, Metadata headers) {

        org.apache.custos.tenant.profile.service.GetTenantRequest tenantReq =
                org.apache.custos.tenant.profile.service.GetTenantRequest
                        .newBuilder().setTenantId(ownerId).build();

        org.apache.custos.tenant.profile.service.GetTenantResponse response =
                tenantProfileClient.getTenant(tenantReq);

        Tenant tenant = response.getTenant();

        Optional<AuthClaim> authClaim = authorize(headers);

        if (authClaim.isPresent() && authClaim.get().isSuperTenant()) {
            return tenant;
        }

        if (tenant == null || (tenant.getParentTenantId() != 0 && tenant.getParentTenantId() != parentTenant)) {
            throw new UnAuthorizedException("Not a valid admin client", null);
        }

        return tenant;
    }

    private Tenant validateTenantWithUserToken(long ownerId, long parentTenant, Metadata headers) {

        org.apache.custos.tenant.profile.service.GetTenantRequest tenantReq =
                org.apache.custos.tenant.profile.service.GetTenantRequest
                        .newBuilder().setTenantId(ownerId).build();

        org.apache.custos.tenant.profile.service.GetTenantResponse response =
                tenantProfileClient.getTenant(tenantReq);

        Tenant tenant = response.getTenant();

        Optional<AuthClaim> authClaim = authorizeUsingUserToken(headers);

        if (authClaim.isPresent() && authClaim.get().isSuperTenant()) {
            return tenant;
        }

        if (tenant == null || (tenant.getParentTenantId() != 0 && tenant.getParentTenantId() != parentTenant)) {
            throw new UnAuthorizedException("Not a valid admin client", null);
        }

        return tenant;
    }


}
