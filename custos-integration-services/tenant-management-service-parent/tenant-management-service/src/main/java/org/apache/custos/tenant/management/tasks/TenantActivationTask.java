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

package org.apache.custos.tenant.management.tasks;

import org.apache.custos.credential.store.client.CredentialStoreServiceClient;
import org.apache.custos.credential.store.service.CredentialMetadata;
import org.apache.custos.credential.store.service.GetCredentialRequest;
import org.apache.custos.credential.store.service.Type;
import org.apache.custos.federated.authentication.client.FederatedAuthenticationClient;
import org.apache.custos.federated.authentication.service.ClientMetadata;
import org.apache.custos.federated.authentication.service.RegisterClientResponse;
import org.apache.custos.iam.admin.client.IamAdminServiceClient;
import org.apache.custos.iam.service.SetUpTenantRequest;
import org.apache.custos.iam.service.SetUpTenantResponse;
import org.apache.custos.integration.core.ServiceException;
import org.apache.custos.integration.core.ServiceTaskImpl;
import org.apache.custos.tenant.management.utils.Constants;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.apache.custos.tenant.profile.service.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Component
public class TenantActivationTask<T, U> extends ServiceTaskImpl<T, U> {

    private static final Logger LOGGER = LoggerFactory.getLogger(TenantActivationTask.class);



    @Autowired
    private IamAdminServiceClient iamAdminServiceClient;

    @Autowired
    private FederatedAuthenticationClient federatedAuthenticationClient;

    @Autowired
    private CredentialStoreServiceClient credentialStoreServiceClient;

    @Autowired
    private TenantProfileClient profileClient;


    @Override
    public void invokeService(T data) {
        try {
            if (data instanceof UpdateStatusResponse) {
                long tenantId = ((UpdateStatusResponse) data).getTenantId();
                LOGGER.debug("Invoking tenant activation task for tenant " + tenantId);

                GetTenantRequest tenantRequest = GetTenantRequest
                        .newBuilder()
                        .setTenantId(tenantId)
                        .build();

                GetTenantResponse tenantRes = profileClient.getTenant(tenantRequest);

                Tenant tenant = tenantRes.getTenant();

                if (tenant != null) {

                    GetCredentialRequest request = GetCredentialRequest
                            .newBuilder()
                            .setId(tenant.getAdminUsername())
                            .setOwnerId(tenantId)
                            .setType(Type.INDIVIDUAL)
                            .build();
                    CredentialMetadata metadata = credentialStoreServiceClient.getCredential(request);

                    if (metadata != null && metadata.getSecret() != null) {

                        SetUpTenantRequest setUpTenantRequest = SetUpTenantRequest
                                .newBuilder()
                                .setTenantId(tenantId)
                                .setTenantName(tenant.getTenantName())
                                .setAdminFirstname(tenant.getAdminFirstName())
                                .setAdminLastname(tenant.getAdminLastName())
                                .setAdminEmail(tenant.getAdminEmail())
                                .addAllRedirectURIs(tenant.getRedirectURIsList())
                                .setAdminPassword(metadata.getSecret())
                                .setAdminUsername(tenant.getAdminUsername())
                                .setRequesterEmail(tenant.getRequesterEmail())
                                .setTenantURL(tenant.getTenantURI())
                                .build();

                        SetUpTenantResponse iamResponse = iamAdminServiceClient.setUPTenant(setUpTenantRequest);

                        CredentialMetadata credentialMetadata = CredentialMetadata
                                .newBuilder()
                                .setId(iamResponse.getClientId())
                                .setSecret(iamResponse.getClientSecret())
                                .setOwnerId(tenantId)
                                .setType(Type.IAM)
                                .build();

                        credentialStoreServiceClient.putCredential(credentialMetadata);

                        ClientMetadata clientMetadata = ClientMetadata
                                .newBuilder()
                                .setTenantId(tenantId)
                                .setTenantName(tenant.getTenantName())
                                .setTenantURI(tenant.getTenantURI())
                                .setComment("Created by custos")
                                .addScope(tenant.getScope())
                                .addAllRedirectURIs(tenant.getRedirectURIsList())
                                .addAllContacts(tenant.getContactsList())
                                .setPerformedBy(tenant.getRequesterEmail())
                                .build();

                        RegisterClientResponse registerClientResponse = federatedAuthenticationClient
                                .addClient(clientMetadata);


                        CredentialMetadata credentialMetadataCILogon = CredentialMetadata
                                .newBuilder()
                                .setId(registerClientResponse.getClientId())
                                .setSecret(registerClientResponse.getClientSecret())
                                .setOwnerId(tenantId)
                                .setType(Type.CILOGON)
                                .build();

                        credentialStoreServiceClient.putCredential(credentialMetadataCILogon);

                        org.apache.custos.tenant.profile.service.UpdateStatusRequest updateTenantRequest =
                                org.apache.custos.tenant.profile.service.UpdateStatusRequest.newBuilder()
                                        .setTenantId(tenantId)
                                        .setStatus(TenantStatus.ACTIVE)
                                        .setUpdatedBy(Constants.SYSTEM)
                                        .build();

                        UpdateStatusResponse response = profileClient.updateTenantStatus(updateTenantRequest);

                        invokeNextTask((U) response);

                    } else {
                        String msg = "Admin password not found  for admin  " + tenant.getAdminUsername();
                        LOGGER.error(msg);
                        getServiceCallback().onError(new ServiceException(msg, null, null));
                    }
                } else {
                    String msg = "Tenant not found  for Id  " + tenantId;
                    LOGGER.error(msg);
                    getServiceCallback().onError(new ServiceException(msg, null, null));
                }
            } else {
                String msg = "Invalid payload type ";
                LOGGER.error(msg);
                getServiceCallback().onError(new ServiceException(msg, null, null));
            }
        } catch (Exception ex) {
            String msg = "Error occurred  " + ex;
            LOGGER.error(msg);
            getServiceCallback().onError(new ServiceException(msg, ex, null));
        }
    }
}