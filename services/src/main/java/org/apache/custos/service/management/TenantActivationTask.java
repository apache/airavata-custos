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

package org.apache.custos.service.management;

import org.apache.custos.core.constants.Constants;
import org.apache.custos.core.credential.store.api.CredentialMetadata;
import org.apache.custos.core.credential.store.api.GetCredentialRequest;
import org.apache.custos.core.credential.store.api.Type;
import org.apache.custos.core.federated.authentication.api.ClientMetadata;
import org.apache.custos.core.federated.authentication.api.RegisterClientResponse;
import org.apache.custos.core.iam.api.ConfigureFederateIDPRequest;
import org.apache.custos.core.iam.api.FederatedIDPs;
import org.apache.custos.core.iam.api.SetUpTenantRequest;
import org.apache.custos.core.iam.api.SetUpTenantResponse;
import org.apache.custos.core.task.ServiceException;
import org.apache.custos.core.task.ServiceTaskImpl;
import org.apache.custos.core.tenant.profile.api.GetTenantRequest;
import org.apache.custos.core.tenant.profile.api.GetTenantResponse;
import org.apache.custos.core.tenant.profile.api.Tenant;
import org.apache.custos.core.tenant.profile.api.TenantStatus;
import org.apache.custos.core.tenant.profile.api.UpdateStatusResponse;
import org.apache.custos.service.credential.store.CredentialStoreService;
import org.apache.custos.service.federated.cilogon.FederatedAuthenticationService;
import org.apache.custos.service.iam.IamAdminService;
import org.apache.custos.service.profile.TenantProfileService;
import org.apache.commons.lang3.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

@Component
public class TenantActivationTask<T, U> extends ServiceTaskImpl<T, U> {

    private static final Logger LOGGER = LoggerFactory.getLogger(TenantActivationTask.class);

    private final IamAdminService iamAdminService;

    private final FederatedAuthenticationService federatedAuthentication;

    private final CredentialStoreService credentialStoreService;

    private final TenantProfileService tenantProfileService;

    @Value("${spring.profiles.active}")
    private String activeProfile;


    public TenantActivationTask(IamAdminService iamAdminService, FederatedAuthenticationService federatedAuthentication, CredentialStoreService credentialStoreService, TenantProfileService tenantProfileService) {
        this.iamAdminService = iamAdminService;
        this.federatedAuthentication = federatedAuthentication;
        this.credentialStoreService = credentialStoreService;
        this.tenantProfileService = tenantProfileService;
    }


    public void invokeService(T data) {
        try {
            if (data instanceof UpdateStatusResponse) {
                long tenantId = ((UpdateStatusResponse) data).getTenantId();
                LOGGER.debug("Invoking tenant activation task for tenant " + tenantId);

                GetTenantRequest tenantRequest = GetTenantRequest.newBuilder()
                        .setTenantId(tenantId)
                        .build();

                GetTenantResponse tenantRes = tenantProfileService.getTenant(tenantRequest);

                Tenant tenant = tenantRes.getTenant();

                if (tenant != null) {

                    GetCredentialRequest request = GetCredentialRequest
                            .newBuilder()
                            .setId(tenant.getAdminUsername())
                            .setOwnerId(tenantId)
                            .setType(Type.INDIVIDUAL)
                            .build();
                    CredentialMetadata metadata = credentialStoreService.getCredential(request);

                    if (metadata != null && metadata.getSecret() != null) {

                        Tenant newTenant = tenant.toBuilder().setAdminPassword(metadata.getSecret()).build();
                        GetCredentialRequest iamClientRequest = GetCredentialRequest
                                .newBuilder()
                                .setOwnerId(tenantId)
                                .setType(Type.IAM)
                                .build();
                        CredentialMetadata iamMetadata = credentialStoreService.getCredential(iamClientRequest);

                        UpdateStatusResponse response;
                        if (iamMetadata == null || iamMetadata.getId() == null || StringUtils.isBlank(iamMetadata.getId())) {
                            response = this.activateTenant(newTenant, Constants.SYSTEM, false);
                        } else {
                            response = this.activateTenant(newTenant, Constants.SYSTEM, true);
                        }

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
            String msg = "Error occurred  " + ex.getCause();
            LOGGER.error(msg, ex);
            getServiceCallback().onError(new ServiceException(msg, ex.getCause(), null));
        }
    }


    public UpdateStatusResponse activateTenant(Tenant tenant, String performedBy, boolean update) {
        GetCredentialRequest getCreRe = GetCredentialRequest.newBuilder()
                .setOwnerId(tenant.getTenantId())
                .setType(Type.CUSTOS)
                .build();

        CredentialMetadata metadata = credentialStoreService.getCredential(getCreRe);
        SetUpTenantRequest setUpTenantRequest = SetUpTenantRequest
                .newBuilder()
                .setTenantId(tenant.getTenantId())
                .setTenantName(tenant.getClientName())
                .setAdminFirstname(tenant.getAdminFirstName())
                .setAdminLastname(tenant.getAdminLastName())
                .setAdminEmail(tenant.getAdminEmail())
                .addAllRedirectURIs(tenant.getRedirectUrisList())
                .setAdminPassword(tenant.getAdminPassword())
                .setAdminUsername(tenant.getAdminUsername())
                .setRequesterEmail(tenant.getRequesterEmail())
                .setTenantURL(tenant.getClientUri())
                .setCustosClientId(metadata.getId())
                .build();

        SetUpTenantResponse iamResponse;
        if (update) {
            iamResponse = iamAdminService.updateTenant(setUpTenantRequest);
        } else {
            iamResponse = iamAdminService.setUPTenant(setUpTenantRequest);
        }

        CredentialMetadata credentialMetadata = CredentialMetadata
                .newBuilder()
                .setId(iamResponse.getClientId())
                .setSecret(iamResponse.getClientSecret())
                .setOwnerId(tenant.getTenantId())
                .setType(Type.IAM)
                .build();

        credentialStoreService.putCredential(credentialMetadata);

        String comment = (tenant.getComment() == null || tenant.getComment().trim().isEmpty()) ?
                "Created by CUSTOS " : tenant.getComment();


        String[] scopes = tenant.getScope() != null ? tenant.getScope().split(" ") : new String[0];

        GetCredentialRequest credentialRequest = GetCredentialRequest.newBuilder()
                .setOwnerId(tenant.getTenantId())
                .setType(Type.CILOGON).build();

        String ciLogonRedirectURI = iamAdminService.getIamServerURL() +
                "realms" + "/" + tenant.getTenantId() + "/" + "broker" + "/" + "oidc" + "/" + "endpoint";


        List<String> arrayList = new ArrayList<>();
        arrayList.add(ciLogonRedirectURI);

        ClientMetadata.Builder clientMetadataBuilder = ClientMetadata
                .newBuilder()
                .setTenantId(tenant.getTenantId())
                .setTenantName(tenant.getClientName())
                .setTenantURI(tenant.getClientUri())
                .setComment(comment)
                .addAllScope(Arrays.asList(scopes))
                .addAllRedirectURIs(arrayList)
                .addAllContacts(tenant.getContactsList())
                .setPerformedBy(performedBy);

        CredentialMetadata creMeta = credentialStoreService.getCredential(credentialRequest);

        clientMetadataBuilder.setClientId(creMeta.getId());

        if (!update) {
            // skip CILOGON client creation for local development
            if (!activeProfile.equalsIgnoreCase("local")) {
                RegisterClientResponse registerClientResponse = federatedAuthentication.addClient(clientMetadataBuilder.build());

                CredentialMetadata credentialMetadataCILogon = CredentialMetadata
                        .newBuilder()
                        .setId(registerClientResponse.getClientId())
                        .setSecret(registerClientResponse.getClientSecret())
                        .setOwnerId(tenant.getTenantId())
                        .setType(Type.CILOGON)
                        .build();

                credentialStoreService.putCredential(credentialMetadataCILogon);

                ConfigureFederateIDPRequest request = ConfigureFederateIDPRequest
                        .newBuilder()
                        .setTenantId(tenant.getTenantId())
                        .setClientID(registerClientResponse.getClientId())
                        .setClientSec(registerClientResponse.getClientSecret())
                        .setScope(tenant.getScope())
                        .setRequesterEmail(tenant.getRequesterEmail())
                        .setType(FederatedIDPs.CILOGON)
                        .build();
                iamAdminService.configureFederatedIDP(request);
            }
        }

        org.apache.custos.core.tenant.profile.api.UpdateStatusRequest updateTenantRequest = org.apache.custos.core.tenant.profile.api.UpdateStatusRequest.newBuilder()
                .setTenantId(tenant.getTenantId())
                .setStatus(TenantStatus.ACTIVE)
                .setUpdatedBy(Constants.SYSTEM)
                .build();
        return tenantProfileService.updateTenantStatus(updateTenantRequest);
    }

}