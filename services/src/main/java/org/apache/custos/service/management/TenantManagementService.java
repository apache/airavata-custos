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
import org.apache.custos.core.credential.store.api.DeleteCredentialRequest;
import org.apache.custos.core.credential.store.api.GetAllCredentialsRequest;
import org.apache.custos.core.credential.store.api.GetAllCredentialsResponse;
import org.apache.custos.core.credential.store.api.GetCredentialRequest;
import org.apache.custos.core.credential.store.api.GetNewCustosCredentialRequest;
import org.apache.custos.core.credential.store.api.Type;
import org.apache.custos.core.federated.authentication.api.CacheManipulationRequest;
import org.apache.custos.core.federated.authentication.api.DeleteClientRequest;
import org.apache.custos.core.federated.authentication.api.GetInstitutionsResponse;
import org.apache.custos.core.iam.api.AddProtocolMapperRequest;
import org.apache.custos.core.iam.api.AddRolesRequest;
import org.apache.custos.core.iam.api.AllRoles;
import org.apache.custos.core.iam.api.DeleteRoleRequest;
import org.apache.custos.core.iam.api.EventPersistenceRequest;
import org.apache.custos.core.iam.api.GetRolesRequest;
import org.apache.custos.core.iam.api.OperationStatus;
import org.apache.custos.core.iam.api.UserAttribute;
import org.apache.custos.core.iam.api.UserRepresentation;
import org.apache.custos.core.iam.api.UserSearchMetadata;
import org.apache.custos.core.iam.api.UserSearchRequest;
import org.apache.custos.core.identity.api.AuthToken;
import org.apache.custos.core.identity.api.GetUserManagementSATokenRequest;
import org.apache.custos.core.task.ServiceCallback;
import org.apache.custos.core.task.ServiceChain;
import org.apache.custos.core.task.ServiceException;
import org.apache.custos.core.tenant.management.api.CreateTenantResponse;
import org.apache.custos.core.tenant.management.api.Credentials;
import org.apache.custos.core.tenant.management.api.DeleteTenantRequest;
import org.apache.custos.core.tenant.management.api.GetTenantRequest;
import org.apache.custos.core.tenant.management.api.TenantValidationRequest;
import org.apache.custos.core.tenant.management.api.UpdateTenantRequest;
import org.apache.custos.core.tenant.profile.api.GetAllTenantsForUserRequest;
import org.apache.custos.core.tenant.profile.api.GetAllTenantsForUserResponse;
import org.apache.custos.core.tenant.profile.api.GetAllTenantsResponse;
import org.apache.custos.core.tenant.profile.api.GetAttributeUpdateAuditTrailResponse;
import org.apache.custos.core.tenant.profile.api.GetAuditTrailRequest;
import org.apache.custos.core.tenant.profile.api.GetStatusUpdateAuditTrailResponse;
import org.apache.custos.core.tenant.profile.api.GetTenantResponse;
import org.apache.custos.core.tenant.profile.api.GetTenantsRequest;
import org.apache.custos.core.tenant.profile.api.Tenant;
import org.apache.custos.core.tenant.profile.api.TenantStatus;
import org.apache.custos.core.tenant.profile.api.UpdateStatusRequest;
import org.apache.custos.core.tenant.profile.api.UpdateStatusResponse;
import org.apache.custos.core.user.profile.api.UserProfile;
import org.apache.custos.core.user.profile.api.UserProfileRequest;
import org.apache.custos.service.credential.store.CredentialStoreService;
import org.apache.custos.service.exceptions.InternalServerException;
import org.apache.custos.service.federated.cilogon.FederatedAuthenticationService;
import org.apache.custos.service.iam.IamAdminService;
import org.apache.custos.service.identity.IdentityService;
import org.apache.custos.service.profile.TenantProfileService;
import org.apache.custos.service.profile.UserProfileService;
import io.grpc.Context;
import jakarta.persistence.EntityNotFoundException;
import org.apache.commons.lang3.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

@Service
public class TenantManagementService {

    private static final Logger LOGGER = LoggerFactory.getLogger(TenantManagementService.class);

    private final TenantProfileService tenantProfileService;

    private final CredentialStoreService credentialStoreService;

    private final IamAdminService iamAdminService;

    private final FederatedAuthenticationService federatedAuthenticationService;

    private final TenantActivationTask<UpdateStatusResponse, UpdateStatusResponse> tenantActivationTask;

    private final UserProfileService userProfileService;

    private final IdentityService identityService;

    @Value("${custos.tenant.base.uri}")
    private String TENANT_BASE_URI;


    public TenantManagementService(TenantProfileService tenantProfileService, CredentialStoreService credentialStoreService,
                                   IamAdminService iamAdminService, FederatedAuthenticationService federatedAuthenticationService,
                                   TenantActivationTask<UpdateStatusResponse, UpdateStatusResponse> tenantActivationTask,
                                   UserProfileService userProfileService, IdentityService identityService) {
        this.tenantProfileService = tenantProfileService;
        this.credentialStoreService = credentialStoreService;
        this.iamAdminService = iamAdminService;
        this.federatedAuthenticationService = federatedAuthenticationService;
        this.tenantActivationTask = tenantActivationTask;
        this.userProfileService = userProfileService;
        this.identityService = identityService;
    }


    public CreateTenantResponse createTenant(Tenant request) {
        try {
            LOGGER.debug("Tenant requested for " + request.getClientName());

            Tenant response = tenantProfileService.addTenant(request);

            long tenantId = response.getTenantId();

            GetNewCustosCredentialRequest req = GetNewCustosCredentialRequest.newBuilder()
                    .setOwnerId(tenantId)
                    .build();

            CredentialMetadata resp = credentialStoreService.getNewCustosCredential(req);

            String message = "Use Base64 encoded clientId:clientSecret as auth token for authorization, " +
                    "Credentials are activated after admin approval";
            boolean isTenantActivated = false;

            if (request.getParentTenantId() > 0) {
                request = request.toBuilder().setTenantId(tenantId).build();
                tenantActivationTask.activateTenant(request, request.getRequesterEmail(), false);
                isTenantActivated = true;
                message = "Credentials are activated";

            } else {
                CredentialMetadata metadata = CredentialMetadata
                        .newBuilder()
                        .setId(request.getAdminUsername())
                        .setSecret(request.getAdminPassword())
                        .setOwnerId(tenantId)
                        .setType(Type.INDIVIDUAL)
                        .build();
                credentialStoreService.putCredential(metadata);
            }

            String tenantBaseURI = TENANT_BASE_URI + "?client_id=" + resp.getId();

            CreateTenantResponse tenantResponse = CreateTenantResponse.newBuilder()
                    .setClientId(resp.getId())
                    .setClientSecret(resp.getSecret())
                    .setClientIdIssuedAt(resp.getClientIdIssuedAt())
                    .setClientSecretExpiresAt(resp.getClientSecretExpiredAt())
                    .setTokenEndpointAuthMethod(Constants.CLIENT_SECRET_BASIC)
                    .setIsActivated(isTenantActivated)
                    .setRegistrationClientUri(tenantBaseURI)
                    .setMsg(message)
                    .build();

            return tenantResponse;

        } catch (Exception ex) {
            String msg = "Error occurred when creating tenant with Id: " + request.getTenantId();
            LOGGER.error(msg, ex);
            throw new IllegalArgumentException(msg, ex);
        }
    }

    public Map<String, String> addClient(long tenantId, String tenantUrl, List<String> redirectUris) {
        GetTenantResponse tenant = tenantProfileService.getTenant(org.apache.custos.core.tenant.profile.api.GetTenantRequest.newBuilder().setTenantId(tenantId).build());

        GetNewCustosCredentialRequest req = GetNewCustosCredentialRequest.newBuilder()
                .setOwnerId(tenant.getTenant().getTenantId())
                .build();

        CredentialMetadata resp = credentialStoreService.getNewCustosCredential(req);
        return iamAdminService.configureClient(tenantId, resp.getId(), tenantUrl, redirectUris);
    }

    public Tenant getTenant(GetTenantRequest request) {
        try {

            Tenant tenant = request.getTenant();

            if (tenant == null) {
                org.apache.custos.core.tenant.profile.api.GetTenantRequest tenantReq = org.apache.custos.core.tenant.profile.api.GetTenantRequest
                        .newBuilder().setTenantId(request.getTenantId()).build();

                org.apache.custos.core.tenant.profile.api.GetTenantResponse response = tenantProfileService.getTenant(tenantReq);
                tenant = response.getTenant();
            }
            if (tenant.getParentTenantId() > 0) {
                GetCredentialRequest cR = GetCredentialRequest.newBuilder()
                        .setOwnerId(tenant.getParentTenantId())
                        .setType(Type.CUSTOS).build();

                CredentialMetadata parentMetadata = credentialStoreService.getCredential(cR);
                tenant = tenant.toBuilder().setParentClientId(parentMetadata.getId()).build();
            }
            GetCredentialRequest credentialRequest = GetCredentialRequest.newBuilder()
                    .setOwnerId(tenant.getTenantId())
                    .setType(Type.CUSTOS).build();

            CredentialMetadata metadata = credentialStoreService.getCredential(credentialRequest);
            return tenant.toBuilder().setClientId(metadata.getId()).build();

        } catch (Exception ex) {
            String msg = "Error occurred when retrieving tenant with the Id:  " + request.getTenantId();
            LOGGER.error(msg);
            throw new IllegalArgumentException(msg, ex);
        }
    }

    public Tenant updateTenant(UpdateTenantRequest request) {

        try {
            Tenant tenant = request.getBody();

            tenant = tenant.toBuilder().setTenantId(request.getTenantId()).build();

            org.apache.custos.core.tenant.profile.api.GetTenantRequest tenantRequest = org.apache.custos.core.tenant.profile.api.GetTenantRequest
                    .newBuilder()
                    .setTenantId(request.getTenantId()).build();

            GetTenantResponse tenantResponse = tenantProfileService.getTenant(tenantRequest);

            if (tenantResponse.getTenant() == null && tenantResponse.getTenant().getTenantId() == 0) {
                String msg = "Cannot find tenant with Tenant name" + tenant.getClientName();
                LOGGER.error(msg);
                throw new EntityNotFoundException(msg);
            }

            Tenant exTenant = tenantResponse.getTenant();
            tenant = tenant.toBuilder().setTenantStatus(exTenant.getTenantStatus()).build();
            GetCredentialRequest passwordRequest = GetCredentialRequest
                    .newBuilder()
                    .setId(tenant.getAdminUsername())
                    .setOwnerId(tenant.getTenantId())
                    .setType(Type.INDIVIDUAL)
                    .build();
            CredentialMetadata metadata = credentialStoreService.getCredential(passwordRequest);

            if (metadata != null && metadata.getSecret() != null) {
                tenant = tenant.toBuilder().setAdminPassword(metadata.getSecret()).build();
            }

            Tenant updateTenant = tenantProfileService.updateTenant(tenant);

            GetCredentialRequest clientIdRequest = GetCredentialRequest.newBuilder()
                    .setOwnerId(tenant.getTenantId())
                    .setType(Type.CUSTOS).build();

            CredentialMetadata idMeta = credentialStoreService.getCredential(clientIdRequest);

            tenant = tenant.toBuilder().setClientId(idMeta.getId()).build();

            if (tenant.getTenantStatus().equals(TenantStatus.ACTIVE)) {

                tenantActivationTask.activateTenant(updateTenant, Constants.GATEWAY_ADMIN, true);
                GetCredentialRequest credentialRequest = GetCredentialRequest.newBuilder()
                        .setOwnerId(tenant.getTenantId())
                        .setId(request.getClientId())
                        .setType(Type.IAM).build();

                CredentialMetadata iamCredential = credentialStoreService.getCredential(credentialRequest);


                GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                        .newBuilder()
                        .setClientId(iamCredential.getId())
                        .setClientSecret(iamCredential.getSecret())
                        .setTenantId(request.getTenantId())
                        .build();
                AuthToken token = identityService.getUserManagementServiceAccountAccessToken(userManagementSATokenRequest);

                if (token != null && token.getAccessToken() != null) {
                    UserSearchMetadata userSearchMetadata = UserSearchMetadata
                            .newBuilder()
                            .setUsername(tenant.getAdminUsername())
                            .build();

                    UserSearchRequest searchRequest = UserSearchRequest.newBuilder().
                            setTenantId(request.getTenantId())
                            .setPerformedBy(Constants.GATEWAY_ADMIN)
                            .setAccessToken(token.getAccessToken())
                            .setUser(userSearchMetadata)
                            .build();

                    UserRepresentation userRepresentation = iamAdminService.getUser(searchRequest);

                    UserProfile profile = convertToProfile(userRepresentation);

                    UserProfileRequest userProfileRequest = UserProfileRequest
                            .newBuilder()
                            .setProfile(profile)
                            .setPerformedBy(Constants.GATEWAY_ADMIN)
                            .setTenantId(request.getTenantId())
                            .build();

                    UserProfile userProfile = userProfileService.getUserProfile(userProfileRequest);

                    if (userProfile == null || StringUtils.isBlank(userProfile.getUsername())) {
                        userProfileService.createUserProfile(userProfileRequest);
                    } else {
                        userProfileService.updateUserProfile(userProfileRequest);
                    }
                }
            }
            return tenant;

        } catch (Exception ex) {
            String msg = "Error occurred at updateTenant with the Id: " + request.getTenantId();
            LOGGER.error(msg, ex);
            throw new IllegalArgumentException(msg, ex);
        }
    }

    public void deleteTenant(DeleteTenantRequest request) {
        try {
            org.apache.custos.core.tenant.profile.api.UpdateStatusRequest updateTenantRequest = org.apache.custos.core.tenant.profile.api.UpdateStatusRequest
                    .newBuilder()
                    .setStatus(TenantStatus.DEACTIVATED)
                    .setTenantId(request.getTenantId())
                    .setUpdatedBy(Constants.GATEWAY_ADMIN)
                    .build();

            tenantProfileService.updateTenantStatus(updateTenantRequest);

            Credentials credentials = request.getCredentials();


            if (!request.getCredentials().getCustosClientId().equals(request.getClientId())) {

                GetAllCredentialsRequest credentialRequest = GetAllCredentialsRequest.newBuilder()
                        .setOwnerId(request.getTenantId())
                        .build();

                GetAllCredentialsResponse response = credentialStoreService.getAllCredentials(credentialRequest);

                if (response.getSecretListCount() > 0) {

                    Credentials.Builder creBuilder = Credentials.newBuilder();
                    response.getSecretListList().forEach(metadata -> {

                                if (metadata.getType() == Type.CUSTOS) {
                                    creBuilder.setCustosClientId(metadata.getId())
                                            .setCustosClientSecret(metadata.getSecret())
                                            .setCustosClientIdIssuedAt(metadata.getClientIdIssuedAt())
                                            .setCustosClientSecretExpiredAt(metadata.getClientSecretExpiredAt());

                                } else if (metadata.getType() == Type.IAM) {
                                    creBuilder.setIamClientId(metadata.getId()).setIamClientSecret(metadata.getSecret());

                                } else if (metadata.getType() == Type.CILOGON) {
                                    creBuilder.setCiLogonClientId(metadata.getId()).setCiLogonClientSecret(metadata.getSecret());
                                }
                            }
                    );
                    credentials = creBuilder.build();
                }
            }

            DeleteClientRequest clientRequest = DeleteClientRequest.newBuilder()
                    .setClientId(credentials.getCiLogonClientId())
                    .setTenantId(request.getTenantId())
                    .setPerformedBy(Constants.GATEWAY_ADMIN)
                    .build();
            federatedAuthenticationService.deleteClient(clientRequest);

            org.apache.custos.core.iam.api.DeleteTenantRequest tenantRequest = org.apache.custos.core.iam.api.DeleteTenantRequest.newBuilder()
                    .setTenantId(request.getTenantId()).build();
            iamAdminService.deleteTenant(tenantRequest);

            DeleteCredentialRequest deleteCredentialRequest = DeleteCredentialRequest.newBuilder()
                    .setOwnerId(request.getTenantId()).build();

            credentialStoreService.deleteCredential(deleteCredentialRequest);

        } catch (Exception ex) {
            String msg = "Error occurred while deleting the tenant with the Id: " + request.getTenantId();
            LOGGER.error(msg);
            throw new IllegalArgumentException(msg, ex);
        }
    }

    public UpdateStatusResponse updateTenantStatus(UpdateStatusRequest request) {
        try {

            GetCredentialRequest credentialRequest = GetCredentialRequest.newBuilder().
                    setId(request.getClientId())
                    .setType(Type.CUSTOS)
                    .build();

            CredentialMetadata metadata = credentialStoreService.getCustosCredentialFromClientId(credentialRequest);

            if (metadata != null) {

                if (request.getSuperTenant()) {
                    metadata = metadata.toBuilder().setSuperTenant(true).build();
                    credentialStoreService.putCredential(metadata);
                }

                request = request.toBuilder().setTenantId(metadata.getOwnerId()).build();
                UpdateStatusResponse response = tenantProfileService.updateTenantStatus(request);

                if (request.getStatus().equals(TenantStatus.ACTIVE)) {

                    Context ctx = Context.current().fork();
                    // Set ctx as the current context within the Runnable
                    UpdateStatusRequest finalRequest = request;
                    CredentialMetadata finalMetadata = metadata;
                    ctx.run(() -> {
                        ServiceCallback callback = new ServiceCallback() {
                            @Override
                            public void onCompleted(Object obj) {
                                org.apache.custos.core.tenant.profile.api.GetTenantRequest tenantRequest = org.apache.custos.core.tenant.profile.api.GetTenantRequest
                                        .newBuilder()
                                        .setTenantId(finalMetadata.getOwnerId())
                                        .build();

                                org.apache.custos.core.tenant.profile.api.GetTenantResponse tenantResponse = tenantProfileService.getTenant(tenantRequest);
                                Tenant savedTenant = tenantResponse.getTenant();

                                GetCredentialRequest credentialRequest = GetCredentialRequest.newBuilder()
                                        .setOwnerId(finalMetadata.getOwnerId())
                                        .setType(Type.IAM)
                                        .build();

                                CredentialMetadata iamMeta = credentialStoreService.getCredential(credentialRequest);

                                GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                                        .newBuilder()
                                        .setClientId(iamMeta.getId())
                                        .setClientSecret(iamMeta.getSecret())
                                        .setTenantId(finalMetadata.getOwnerId())
                                        .build();
                                AuthToken token = identityService.getUserManagementServiceAccountAccessToken(userManagementSATokenRequest);

                                if (token != null && StringUtils.isNotBlank(token.getAccessToken())) {
                                    UserSearchMetadata userSearchMetadata = UserSearchMetadata.newBuilder()
                                            .setUsername(savedTenant.getAdminUsername())
                                            .build();

                                    UserSearchRequest searchRequest = UserSearchRequest.newBuilder()
                                            .setTenantId(savedTenant.getTenantId())
                                            .setPerformedBy(finalRequest.getUpdatedBy())
                                            .setAccessToken(token.getAccessToken())
                                            .setUser(userSearchMetadata)
                                            .build();

                                    UserRepresentation userRepresentation = iamAdminService.getUser(searchRequest);
                                    UserProfile profile = convertToProfile(userRepresentation);
                                    UserProfileRequest userProfileRequest = UserProfileRequest.newBuilder()
                                            .setProfile(profile)
                                            .setPerformedBy(finalRequest.getUpdatedBy())
                                            .setTenantId(finalRequest.getTenantId())
                                            .build();

                                    UserProfile userProfile = userProfileService.getUserProfile(userProfileRequest);

                                    if (userProfile == null || StringUtils.isBlank(userProfile.getUsername())) {
                                        userProfileService.createUserProfile(userProfileRequest);
                                    } else {
                                        userProfileService.updateUserProfile(userProfileRequest);
                                    }

                                } else {
                                    String msg = "Tenant Activation task failed, cannot find IAM server credentials";
                                    LOGGER.error(msg);
                                    throw new RuntimeException(msg);
                                }
                            }

                            @Override
                            public void onError(ServiceException ex) {
                                String msg = "Tenant Activation task failed " + ex;
                                LOGGER.error(msg);
                                org.apache.custos.core.tenant.profile.api.UpdateStatusRequest updateTenantRequest = org.apache.custos.core.tenant.profile.api.UpdateStatusRequest.newBuilder()
                                        .setTenantId(finalRequest.getTenantId())
                                        .setStatus(TenantStatus.CANCELLED)
                                        .setUpdatedBy(Constants.SYSTEM)
                                        .build();
                                tenantProfileService.updateTenantStatus(updateTenantRequest);
                                throw new RuntimeException(msg);
                            }
                        };

                        ServiceChain chain = ServiceChain.newBuilder(tenantActivationTask, callback).build();
                        chain.serve(response);
                    });
                    return response;

                } else {
                    return response;
                }
            } else {
                String msg = "Cannot find a Tenant with given client id " + request.getTenantId();
                LOGGER.error(msg);
                throw new EntityNotFoundException(msg);
            }

        } catch (Exception ex) {
            String msg = "Tenant update task failed for tenant " + request.getTenantId();
            LOGGER.error(msg, ex);
            throw new InternalServerException(msg, ex);
        }
    }

    public OperationStatus validateTenant(TenantValidationRequest request) {
        try {
            GetCredentialRequest credentialRequest = GetCredentialRequest.newBuilder()
                    .setId(request.getClientId()).build();

            CredentialMetadata metadata = credentialStoreService.getCustosCredentialFromClientId(credentialRequest);
            return OperationStatus.newBuilder().setStatus(metadata.getSecret().trim().equals(request.getClientSec().trim())).build();

        } catch (Exception ex) {
            String msg = "Error occurred while validating tenant with Id " + request.getClientId() + " reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new InternalServerException(msg, ex);
        }
    }

    public AllRoles addTenantRoles(AddRolesRequest request) {
        try {
            return iamAdminService.addRolesToTenant(request);
        } catch (Exception ex) {
            String msg = "Error occurred at addTenantRoles " + ex.getMessage();
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    public AllRoles getTenantRoles(GetRolesRequest request) {
        try {
            return iamAdminService.getRolesOfTenant(request);
        } catch (Exception ex) {
            String msg = "Error occurred at getTenantRoles " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new InternalServerException(msg, ex);
        }
    }

    public OperationStatus deleteRole(DeleteRoleRequest request) {
        try {
            return iamAdminService.deleteRole(request);
        } catch (Exception ex) {
            String msg = "Error occurred at deleteRole " + request.getRole() + " of tenant " + request.getTenantId() + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new InternalServerException(msg, ex);
        }
    }

    public OperationStatus addProtocolMapper(AddProtocolMapperRequest request) {
        try {
            return iamAdminService.addProtocolMapper(request);
        } catch (Exception ex) {
            String msg = "Error occurred at addProtocolMapper " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new InternalServerException(msg, ex);
        }
    }

    public OperationStatus configureEventPersistence(EventPersistenceRequest request) {
        try {
            return iamAdminService.configureEventPersistence(request);
        } catch (Exception ex) {
            String msg = "Error occurred at configureEventPersistence " + ex.getMessage();
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    public GetAllTenantsResponse getAllTenants(GetTenantsRequest request) {
        try {
            GetAllTenantsResponse response = tenantProfileService.getAllTenants(request);
            if (response != null && !response.getTenantList().isEmpty()) {
                List<Tenant> tenantList = new ArrayList<>();

                for (Tenant tenant : response.getTenantList()) {
                    GetCredentialRequest credentialRequest = GetCredentialRequest.newBuilder()
                            .setOwnerId(tenant.getTenantId())
                            .setType(Type.CUSTOS)
                            .build();

                    CredentialMetadata metadata = credentialStoreService.getCredential(credentialRequest);

                    if (tenant.getParentTenantId() > 0) {
                        GetCredentialRequest cR = GetCredentialRequest.newBuilder()
                                .setOwnerId(tenant.getParentTenantId())
                                .setType(Type.CUSTOS).build();

                        CredentialMetadata parentMetadata = credentialStoreService.getCredential(cR);
                        tenant = tenant.toBuilder().setParentClientId(parentMetadata.getId()).build();
                    }

                    tenant = tenant.toBuilder().setClientId(metadata.getId()).build();
                    tenantList.add(tenant);
                }

                response = response.toBuilder().clearTenant().addAllTenant(tenantList).build();
            }
            return response;

        } catch (Exception ex) {
            String msg = "Error occurred at getAllTenants " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new InternalServerException(msg, ex);
        }
    }

    public GetAllTenantsResponse getChildTenants(GetTenantsRequest request) {
        try {
            GetAllTenantsResponse response = tenantProfileService.getAllTenants(request);
            if (response != null && !response.getTenantList().isEmpty()) {
                List<Tenant> tenantList = new ArrayList<>();

                for (Tenant tenant : response.getTenantList()) {
                    GetCredentialRequest credentialRequest = GetCredentialRequest.newBuilder()
                            .setOwnerId(tenant.getTenantId())
                            .setType(Type.CUSTOS).build();

                    CredentialMetadata metadata = credentialStoreService.getCredential(credentialRequest);
                    if (tenant.getParentTenantId() > 0) {
                        GetCredentialRequest cR = GetCredentialRequest.newBuilder()
                                .setOwnerId(tenant.getParentTenantId())
                                .setType(Type.CUSTOS)
                                .build();

                        CredentialMetadata parentMetadata = credentialStoreService.getCredential(cR);
                        tenant = tenant.toBuilder().setParentClientId(parentMetadata.getId()).build();
                    }

                    tenant = tenant.toBuilder().setClientId(metadata.getId()).build();
                    tenantList.add(tenant);
                }
                response = response.toBuilder().clearTenant().addAllTenant(tenantList).build();
            }
            return response;

        } catch (Exception ex) {
            String msg = "Error occurred at getChildTenants " + ex.getMessage();
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    public GetAllTenantsForUserResponse getAllTenantsForUser(GetAllTenantsForUserRequest request) {
        try {
            return tenantProfileService.getAllTenantsForUser(request);
        } catch (Exception ex) {
            String msg = "Error occurred at getAllTenantsForUser " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new InternalServerException(msg, ex);
        }
    }

    public org.apache.custos.core.federated.authentication.api.Status addToCache(CacheManipulationRequest request) {
        try {
            LOGGER.debug("Request received to add to cache for tenant  " + request.getTenantId());
            return federatedAuthenticationService.addToCache(request);

        } catch (Exception ex) {
            String msg = "Error occurred calling addToCache method for tenant  " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new InternalServerException(msg, ex);
        }
    }

    public org.apache.custos.core.federated.authentication.api.Status removeFromCache(CacheManipulationRequest request) {
        try {
            LOGGER.debug("Request received to removeFromCache for tenant  " + request.getTenantId());
            return federatedAuthenticationService.removeFromCache(request);

        } catch (Exception ex) {
            String msg = "Error occurred calling removeFromCache method for tenant  " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new InternalServerException(msg, ex);
        }
    }

    public GetInstitutionsResponse getFromCache(CacheManipulationRequest request) {
        try {
            LOGGER.debug("Request received to getFromCache for tenant  " + request.getTenantId());
            return federatedAuthenticationService.getFromCache(request);

        } catch (Exception ex) {
            String msg = "Error occurred calling getFromCache method for tenant  " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new InternalServerException(msg, ex);
        }
    }

    public GetInstitutionsResponse getInstitutions(CacheManipulationRequest request) {
        try {
            LOGGER.debug("Request received to getInstitutions for tenant  " + request.getTenantId());
            return federatedAuthenticationService.getInstitutions(request);

        } catch (Exception ex) {
            String msg = "Error occurred calling getInstitutions method for tenant  " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new InternalServerException(msg, ex);
        }
    }

    public GetStatusUpdateAuditTrailResponse getTenantStatusUpdateAuditTrail(GetAuditTrailRequest request) {
        return tenantProfileService.getTenantStatusUpdateAuditTrail(request);
    }

    public GetAttributeUpdateAuditTrailResponse getTenantAttributeUpdateAuditTrail(GetAuditTrailRequest request) {
        return tenantProfileService.getTenantAttributeUpdateAuditTrail(request);
    }

    private UserProfile convertToProfile(UserRepresentation representation) {
        UserProfile.Builder profileBuilder = UserProfile.newBuilder();

        if (representation.getRealmRolesCount() > 0) {
            profileBuilder.addAllRealmRoles(representation.getRealmRolesList());
        }

        if (representation.getClientRolesCount() > 0) {
            profileBuilder.addAllClientRoles(representation.getClientRolesList());
        }

        if (representation.getAttributesCount() > 0) {
            List<UserAttribute> attributeList = representation.getAttributesList();

            List<org.apache.custos.core.user.profile.api.UserAttribute> userAtrList = new ArrayList<>();
            attributeList.forEach(atr -> {
                org.apache.custos.core.user.profile.api.UserAttribute userAttribute = org.apache.custos.core.user.profile.api.UserAttribute
                        .newBuilder()
                        .setKey(atr.getKey())
                        .addAllValues(atr.getValuesList())
                        .build();
                userAtrList.add(userAttribute);
            });
            profileBuilder.addAllAttributes(userAtrList);
        }

        profileBuilder.setUsername(representation.getUsername().toLowerCase());
        profileBuilder.setFirstName(representation.getFirstName());
        profileBuilder.setLastName(representation.getLastName());
        profileBuilder.setEmail(representation.getEmail());

        return profileBuilder.build();
    }
}
