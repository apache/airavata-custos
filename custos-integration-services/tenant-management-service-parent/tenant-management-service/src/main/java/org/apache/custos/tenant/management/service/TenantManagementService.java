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

package org.apache.custos.tenant.management.service;

import io.grpc.Context;
import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import org.apache.custos.credential.store.client.CredentialStoreServiceClient;
import org.apache.custos.credential.store.service.*;
import org.apache.custos.federated.authentication.client.FederatedAuthenticationClient;
import org.apache.custos.federated.authentication.service.CacheManipulationRequest;
import org.apache.custos.federated.authentication.service.DeleteClientRequest;
import org.apache.custos.federated.authentication.service.GetInstitutionsIdsAsResponse;
import org.apache.custos.federated.authentication.service.GetInstitutionsResponse;
import org.apache.custos.iam.admin.client.IamAdminServiceClient;
import org.apache.custos.iam.service.OperationStatus;
import org.apache.custos.iam.service.*;
import org.apache.custos.identity.client.IdentityClient;
import org.apache.custos.identity.service.AuthToken;
import org.apache.custos.identity.service.GetUserManagementSATokenRequest;
import org.apache.custos.integration.core.ServiceCallback;
import org.apache.custos.integration.core.ServiceChain;
import org.apache.custos.integration.core.ServiceException;
import org.apache.custos.tenant.management.service.TenantManagementServiceGrpc.TenantManagementServiceImplBase;
import org.apache.custos.tenant.management.tasks.TenantActivationTask;
import org.apache.custos.tenant.management.utils.Constants;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.apache.custos.tenant.profile.service.*;
import org.apache.custos.user.profile.client.UserProfileClient;
import org.apache.custos.user.profile.service.UserProfile;
import org.apache.custos.user.profile.service.UserProfileRequest;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;


@GRpcService
public class TenantManagementService extends TenantManagementServiceImplBase {
    private static final Logger LOGGER = LoggerFactory.getLogger(TenantManagementService.class);


    @Autowired
    private TenantProfileClient profileClient;

    @Autowired
    private CredentialStoreServiceClient credentialStoreServiceClient;

    @Autowired
    private IamAdminServiceClient iamAdminServiceClient;

    @Autowired
    private FederatedAuthenticationClient federatedAuthenticationClient;

    @Autowired
    private TenantActivationTask<UpdateStatusResponse, UpdateStatusResponse> tenantActivationTask;

    @Autowired
    private UserProfileClient userProfileClient;

    @Autowired
    private IdentityClient identityClient;


    @Override
    public void createTenant(Tenant request, StreamObserver<CreateTenantResponse> responseObserver) {
        try {
            LOGGER.debug("Tenant requested for " + request.getClientName());

            Tenant response = profileClient.addTenant(request);

            long tenantId = response.getTenantId();

            GetNewCustosCredentialRequest req = GetNewCustosCredentialRequest.newBuilder().setOwnerId(tenantId).build();

            CredentialMetadata resp = credentialStoreServiceClient.getNewCustosCredentials(req);

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

                credentialStoreServiceClient.putCredential(metadata);

            }

            String tenantBaseURI = Constants.TENANT_BASE_URI + "?client_id=" + resp.getId();


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

            responseObserver.onNext(tenantResponse);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred at createTenant " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INVALID_ARGUMENT.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getTenant(GetTenantRequest request, StreamObserver<GetTenantResponse> responseObserver) {
        try {

            Tenant tenant = request.getTenant(); // retrieved cached tenant from interceptors

            if (tenant == null) {
                org.apache.custos.tenant.profile.service.GetTenantRequest tenantReq =
                        org.apache.custos.tenant.profile.service.GetTenantRequest
                                .newBuilder().setTenantId(request.getTenantId()).build();

                org.apache.custos.tenant.profile.service.GetTenantResponse response =
                        profileClient.getTenant(tenantReq);
                tenant = response.getTenant();
            }

            double clientIdIssuedAt = request.getCredentials().getCustosClientIdIssuedAt();

            if (!request.getCredentials().getCustosClientId().equals(request.getClientId())) {

                GetCredentialRequest credentialRequest = GetCredentialRequest.newBuilder()
                        .setOwnerId(tenant.getTenantId())
                        .setId(request.getClientId())
                        .setType(Type.CUSTOS).build();

                clientIdIssuedAt = credentialStoreServiceClient.
                        getCredential(credentialRequest).getClientIdIssuedAt();
            }


            String[] grantTypes = {Constants.AUTHORIZATION_CODE};

            GetTenantResponse tenantResponse = GetTenantResponse.newBuilder()
                    .setClientId(request.getClientId())
                    .setAdminEmail(tenant.getAdminEmail())
                    .setAdminFirstName(tenant.getAdminFirstName())
                    .setAdminLastName(tenant.getAdminLastName())
                    .setAdminUsername(tenant.getAdminUsername())
                    .setRequesterEmail(tenant.getRequesterEmail())
                    .setApplicationType(tenant.getApplicationType())
                    .setClientName(tenant.getClientName())
                    .setClientUri(tenant.getClientUri())
                    .setComment(tenant.getComment())
                    .setDomain(tenant.getDomain())
                    .setExampleExtensionParameter(tenant.getExampleExtensionParameter())
                    .setJwksUri(tenant.getJwksUri())
                    .addAllContacts(tenant.getContactsList())
                    .addAllRedirectUris(tenant.getRedirectUrisList())
                    .setLogoUri(tenant.getLogoUri())
                    .setPolicyUri(tenant.getPolicyUri())
                    .setTosUri(tenant.getTosUri())
                    .setScope(tenant.getScope())
                    .setSoftwareId(tenant.getSoftwareId())
                    .setSoftwareVersion(tenant.getSoftwareVersion())
                    .addAllGrantTypes(Arrays.asList(grantTypes))
                    .setClientIdIssuedAt(clientIdIssuedAt)
                    .build();
            responseObserver.onNext(tenantResponse);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at getTenant " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INVALID_ARGUMENT.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void updateTenant(UpdateTenantRequest request, StreamObserver<GetTenantResponse> responseObserver) {

        try {
            Tenant tenant = request.getBody();

            tenant = tenant.toBuilder().setTenantId(request.getTenantId()).build();

            Tenant updateTenant = profileClient.updateTenant(tenant);

            tenantActivationTask.activateTenant(updateTenant, Constants.GATEWAY_ADMIN, true);

            double clientIdIssuedAt = request.getCredentials().getCustosClientIdIssuedAt();


            if (!request.getCredentials().getCustosClientId().equals(request.getClientId())) {

                GetCredentialRequest credentialRequest = GetCredentialRequest.newBuilder()
                        .setOwnerId(tenant.getTenantId())
                        .setId(request.getClientId())
                        .setType(Type.CUSTOS).build();

                clientIdIssuedAt = credentialStoreServiceClient.
                        getCredential(credentialRequest).getClientIdIssuedAt();
            }


            String[] grantTypes = {Constants.AUTHORIZATION_CODE,
                    Constants.CLIENT_CREDENTIALS,
                    Constants.PASSWORD_GRANT_TYPE,
                    Constants.REFRESH_TOKEN};

            GetCredentialRequest credentialRequest = GetCredentialRequest.newBuilder()
                    .setOwnerId(tenant.getTenantId())
                    .setId(request.getClientId())
                    .setType(Type.IAM).build();

            CredentialMetadata iamCredential = credentialStoreServiceClient.
                    getCredential(credentialRequest);


            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(iamCredential.getId())
                    .setClientSecret(iamCredential.getSecret())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

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

                UserRepresentation userRepresentation = iamAdminServiceClient.getUser(searchRequest);

                UserProfile profile = convertToProfile(userRepresentation);

                UserProfileRequest userProfileRequest = UserProfileRequest
                        .newBuilder()
                        .setProfile(profile)
                        .setPerformedBy(Constants.GATEWAY_ADMIN)
                        .setTenantId(request.getTenantId())
                        .build();

                UserProfile userProfile = userProfileClient.getUser(userProfileRequest);

                if (userProfile == null || userProfile.getUsername().equals("")) {
                    userProfileClient.createUserProfile(userProfileRequest);
                } else {
                    userProfileClient.updateUserProfile(userProfileRequest);
                }


            }

            GetTenantResponse tenantResponse = GetTenantResponse.newBuilder()
                    .setClientId(request.getClientId())
                    .setAdminEmail(tenant.getAdminEmail())
                    .setAdminFirstName(tenant.getAdminFirstName())
                    .setAdminLastName(tenant.getAdminLastName())
                    .setRequesterEmail(tenant.getRequesterEmail())
                    .setApplicationType(tenant.getApplicationType())
                    .setClientName(tenant.getClientName())
                    .setClientUri(tenant.getClientUri())
                    .setComment(tenant.getComment())
                    .setDomain(tenant.getDomain())
                    .setExampleExtensionParameter(tenant.getExampleExtensionParameter())
                    .setJwksUri(tenant.getJwksUri())
                    .addAllContacts(tenant.getContactsList())
                    .addAllRedirectUris(tenant.getRedirectUrisList())
                    .setLogoUri(tenant.getLogoUri())
                    .setPolicyUri(tenant.getPolicyUri())
                    .setTosUri(tenant.getTosUri())
                    .setScope(tenant.getScope())
                    .setSoftwareId(tenant.getSoftwareId())
                    .setSoftwareVersion(tenant.getSoftwareVersion())
                    .addAllGrantTypes(Arrays.asList(grantTypes))
                    .setClientIdIssuedAt(clientIdIssuedAt)
                    .build();

            responseObserver.onNext(tenantResponse);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at updateTenant " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INVALID_ARGUMENT.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void deleteTenant(DeleteTenantRequest request, StreamObserver<com.google.protobuf.Empty> responseObserver) {
        try {
            org.apache.custos.tenant.profile.service.UpdateStatusRequest updateTenantRequest =
                    org.apache.custos.tenant.profile.service.UpdateStatusRequest
                            .newBuilder()
                            .setStatus(TenantStatus.DEACTIVATED)
                            .setTenantId(request.getTenantId())
                            .setUpdatedBy(Constants.GATEWAY_ADMIN)
                            .build();

            profileClient.updateTenantStatus(updateTenantRequest);

            Credentials credentials = request.getCredentials();


            if (!request.getCredentials().getCustosClientId().equals(request.getClientId())) {

                GetAllCredentialsRequest credentialRequest = GetAllCredentialsRequest.newBuilder()
                        .setOwnerId(request.getTenantId())
                        .build();

                GetAllCredentialsResponse response = credentialStoreServiceClient.getAllCredentials(credentialRequest);

                if (response.getSecretListCount() > 0) {

                    Credentials.Builder creBuilder = Credentials.newBuilder();
                    response.getSecretListList().forEach(metadata -> {

                                if (metadata.getType() == Type.CUSTOS) {
                                    creBuilder.setCustosClientId(metadata.getId())
                                            .setCustosClientSecret(metadata.getSecret())
                                            .setCustosClientIdIssuedAt(metadata.getClientIdIssuedAt())
                                            .setCustosClientSecretExpiredAt(metadata.getClientSecretExpiredAt());
                                } else if (metadata.getType() == Type.IAM) {
                                    creBuilder.setIamClientId(metadata.getId())
                                            .setIamClientSecret(metadata.getSecret());


                                } else if (metadata.getType() == Type.CILOGON) {
                                    creBuilder.setCiLogonClientId(metadata.getId())
                                            .setCiLogonClientSecret(metadata.getSecret());
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
            federatedAuthenticationClient.deleteClient(clientRequest);


            org.apache.custos.iam.service.DeleteTenantRequest tenantRequest = org.apache.custos.iam.service.
                    DeleteTenantRequest.newBuilder().setTenantId(request.getTenantId()).build();
            iamAdminServiceClient.deleteTenant(tenantRequest);

            DeleteCredentialRequest deleteCredentialRequest = DeleteCredentialRequest
                    .newBuilder().setOwnerId(request.getTenantId()).build();

            credentialStoreServiceClient.deleteCredential(deleteCredentialRequest);

            responseObserver.onNext(com.google.protobuf.Empty.newBuilder().build());
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred at deleteTenant " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INVALID_ARGUMENT.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void addTenantRoles(AddRolesRequest request, StreamObserver<AllRoles> responseObserver) {
        try {
            AllRoles allRoles = iamAdminServiceClient.addRolesToTenant(request);

            responseObserver.onNext(allRoles);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at addTenantRoles " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void getTenantRoles(GetRolesRequest request, StreamObserver<AllRoles> responseObserver) {
        try {
            AllRoles allRoles = iamAdminServiceClient.getRolesOfTenant(request);

            responseObserver.onNext(allRoles);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at getTenantRoles " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void addProtocolMapper(AddProtocolMapperRequest request, StreamObserver<OperationStatus> responseObserver) {
        try {
            OperationStatus allRoles = iamAdminServiceClient.addProtocolMapper(request);

            responseObserver.onNext(allRoles);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at addProtocolMapper " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void configureEventPersistence(EventPersistenceRequest request, StreamObserver<OperationStatus> responseObserver) {
        try {
            OperationStatus allRoles = iamAdminServiceClient.configureEventPersistence(request);

            responseObserver.onNext(allRoles);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at configureEventPersistence " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getAllTenants(GetTenantsRequest request, StreamObserver<GetAllTenantsResponse> responseObserver) {
        try {
            GetAllTenantsResponse response = profileClient.getAllTenants(request);

            if (response != null && !response.getTenantList().isEmpty()) {
                List<Tenant> tenantList = new ArrayList<>();

                for (Tenant tenant : response.getTenantList()) {

                    GetCredentialRequest credentialRequest = GetCredentialRequest.newBuilder()
                            .setOwnerId(tenant.getTenantId())
                            .setType(Type.CUSTOS).build();

                    CredentialMetadata metadata = credentialStoreServiceClient.
                            getCredential(credentialRequest);

                    tenant = tenant.toBuilder().setClientId(metadata.getId()).build();
                    tenantList.add(tenant);

                }

                response = response.toBuilder().clearTenant().addAllTenant(tenantList).build();
            }

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred at getAllTenants " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getChildTenants(GetTenantsRequest request, StreamObserver<GetAllTenantsResponse> responseObserver) {
        try {
            GetAllTenantsResponse response = profileClient.getAllTenants(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred at getChildTenants " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void getAllTenantsForUser(GetAllTenantsForUserRequest request, StreamObserver<GetAllTenantsForUserResponse> responseObserver) {
        try {
            GetAllTenantsForUserResponse response = profileClient.getAllTenantsForUser(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred at getAllTenantsForUser " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void updateTenantStatus(UpdateStatusRequest request, StreamObserver<UpdateStatusResponse> responseObserver) {
        try {

            GetCredentialRequest credentialRequest = GetCredentialRequest.newBuilder().
                    setId(request.getClientId())
                    .setType(Type.CUSTOS)
                    .build();

            CredentialMetadata metadata = credentialStoreServiceClient.getCustosCredentialFromClientId(credentialRequest);

            if (metadata != null) {
                request = request.toBuilder().setTenantId(metadata.getOwnerId()).build();
                UpdateStatusResponse response = profileClient.updateTenantStatus(request);

                if (request.getStatus().equals(TenantStatus.ACTIVE)) {

                    Context ctx = Context.current().fork();
                    // Set ctx as the current context within the Runnable
                    UpdateStatusRequest finalRequest = request;
                    ctx.run(() -> {
                        ServiceCallback callback = new ServiceCallback() {
                            @Override
                            public void onCompleted(Object obj) {
                                org.apache.custos.tenant.profile.service.GetTenantRequest tenantRequest =
                                        org.apache.custos.tenant.profile.service.GetTenantRequest
                                                .newBuilder()
                                                .setTenantId(metadata.getOwnerId())
                                                .build();

                                org.apache.custos.tenant.profile.service.GetTenantResponse tenantResponse =
                                        profileClient.getTenant(tenantRequest);
                                Tenant savedTenant = tenantResponse.getTenant();

                                GetCredentialRequest credentialRequest = GetCredentialRequest.newBuilder()
                                        .setOwnerId(metadata.getOwnerId())
                                        .setType(Type.IAM)
                                        .build();

                                CredentialMetadata iamMeta = credentialStoreServiceClient.getCredential(credentialRequest);

                                GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                                        .newBuilder()
                                        .setClientId(iamMeta.getId())
                                        .setClientSecret(iamMeta.getSecret())
                                        .setTenantId(metadata.getOwnerId())
                                        .build();
                                AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

                                if (token != null && token.getAccessToken() != null) {


                                    UserSearchMetadata userSearchMetadata = UserSearchMetadata
                                            .newBuilder()
                                            .setUsername(savedTenant.getAdminUsername())
                                            .build();

                                    UserSearchRequest searchRequest = UserSearchRequest.newBuilder().
                                            setTenantId(savedTenant.getTenantId())
                                            .setPerformedBy(finalRequest.getUpdatedBy())
                                            .setAccessToken(token.getAccessToken())
                                            .setUser(userSearchMetadata)
                                            .build();

                                    UserRepresentation userRepresentation = iamAdminServiceClient.getUser(searchRequest);

                                    UserProfile profile = convertToProfile(userRepresentation);

                                    UserProfileRequest userProfileRequest = UserProfileRequest
                                            .newBuilder()
                                            .setProfile(profile)
                                            .setPerformedBy(finalRequest.getUpdatedBy())
                                            .setTenantId(finalRequest.getTenantId())
                                            .build();


                                    UserProfile userProfile = userProfileClient.getUser(userProfileRequest);

                                    if (userProfile == null || userProfile.getUsername().equals("")) {
                                        userProfileClient.createUserProfile(userProfileRequest);
                                    } else {
                                        userProfileClient.updateUserProfile(userProfileRequest);
                                    }

                                    responseObserver.onNext(response);
                                    responseObserver.onCompleted();
                                } else {
                                    String msg = "Tenant Activation task failed, cannot find IAM server credentials";
                                    LOGGER.error(msg);
                                    responseObserver.onError(Status.CANCELLED.withDescription(msg).asRuntimeException());

                                }
                            }

                            @Override
                            public void onError(ServiceException ex) {
                                String msg = "Tenant Activation task failed " + ex;
                                LOGGER.error(msg);
                                org.apache.custos.tenant.profile.service.UpdateStatusRequest updateTenantRequest =
                                        org.apache.custos.tenant.profile.service.UpdateStatusRequest.newBuilder()
                                                .setTenantId(finalRequest.getTenantId())
                                                .setStatus(TenantStatus.CANCELLED)
                                                .setUpdatedBy(Constants.SYSTEM)
                                                .build();
                                profileClient.updateTenantStatus(updateTenantRequest);
                                responseObserver.onError(Status.CANCELLED.withDescription(msg).asRuntimeException());
                            }
                        };

                        ServiceChain chain = ServiceChain.newBuilder(tenantActivationTask, callback).build();

                        chain.serve(response);
                    });
                } else {
                    responseObserver.onNext(response);
                    responseObserver.onCompleted();
                }
            } else {
                String msg = "Cannot find a Tenant with given client id " + request.getTenantId();
                LOGGER.error(msg);
                responseObserver.onError(Status.NOT_FOUND.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Tenant update task failed for tenant " + request.getTenantId() + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void addToCache(CacheManipulationRequest request,
                           StreamObserver<org.apache.custos.federated.authentication.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to add to cache for tenant  " + request.getTenantId());


            org.apache.custos.federated.authentication.service.Status status = federatedAuthenticationClient.addToCache(request);

            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred calling addToCache method for tenant  " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void removeFromCache(CacheManipulationRequest request,
                                StreamObserver<org.apache.custos.federated.authentication.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to removeFromCache for tenant  " + request.getTenantId());
            org.apache.custos.federated.authentication.service.Status status = federatedAuthenticationClient.removeFromCache(request);

            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred calling removeFromCache method for tenant  " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getFromCache(CacheManipulationRequest request,
                             StreamObserver<GetInstitutionsResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getFromCache for tenant  " + request.getTenantId());

            GetInstitutionsResponse status = federatedAuthenticationClient.getFromCache(request);

            responseObserver.onNext(status);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred calling getFromCache method for tenant  " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getInstitutions(CacheManipulationRequest request,
                                StreamObserver<GetInstitutionsResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getInstitutions for tenant  " + request.getTenantId());
            GetInstitutionsResponse status = federatedAuthenticationClient.getInstitutions(request);

            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred calling getInstitutions method for tenant  " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getTenantStatusUpdateAuditTrail(GetAuditTrailRequest request, StreamObserver<GetStatusUpdateAuditTrailResponse> responseObserver) {

        GetStatusUpdateAuditTrailResponse response = profileClient.getStatusUpdateAuditTrail(request);
        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    @Override
    public void getTenantAttributeUpdateAuditTrail(GetAuditTrailRequest request, StreamObserver<GetAttributeUpdateAuditTrailResponse> responseObserver) {
        GetAttributeUpdateAuditTrailResponse response = profileClient.getAttributeUpdateAuditTrail(request);

        responseObserver.onNext(response);
        responseObserver.onCompleted();
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

            List<org.apache.custos.user.profile.service.UserAttribute> userAtrList = new ArrayList<>();
            attributeList.forEach(atr -> {
                org.apache.custos.user.profile.service.UserAttribute userAttribute =
                        org.apache.custos.user.profile.service.UserAttribute
                                .newBuilder()
                                .setKey(atr.getKey())
                                .addAllValue(atr.getValuesList())
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
