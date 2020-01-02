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
import org.apache.custos.integration.core.ServiceCallback;
import org.apache.custos.integration.core.ServiceChain;
import org.apache.custos.integration.core.ServiceException;
import org.apache.custos.tenant.management.service.TenantManagementServiceGrpc.TenantManagementServiceImplBase;
import org.apache.custos.tenant.management.tasks.TenantActivationTask;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.apache.custos.tenant.profile.service.*;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;


@GRpcService
public class TenantManagementService extends TenantManagementServiceImplBase {
    private static final Logger LOGGER = LoggerFactory.getLogger(TenantManagementService.class);


    @Autowired
    private TenantProfileClient profileClient;

    @Autowired
    private CredentialStoreServiceClient credentialStoreServiceClient;

    @Autowired
    private TenantActivationTask<UpdateStatusResponse, UpdateStatusResponse> tenantActivationTask;


    @Override
    public void createTenant(CreateTenantRequest request, StreamObserver<CreateTenantResponse> responseObserver) {
        try {
            LOGGER.info("Tenant requested for " + request.getTenant().getTenantName());


            AddTenantResponse response = profileClient.addTenant(request.getTenant());

            LOGGER.info("Admin password :" + request.getTenant().getAdminPassword());
            LOGGER.info("Admin Username :" + request.getTenant().getAdminUsername());

            long tenantId = response.getTenantId();

            GetNewCustosCredentialRequest req = GetNewCustosCredentialRequest.newBuilder().setOwnerId(tenantId).build();

            GetNewCustosCredentialResponse resp = credentialStoreServiceClient.getNewCustosCredentials(req);

            CredentialMetadata metadata = CredentialMetadata
                    .newBuilder()
                    .setId(request.getTenant().getAdminUsername())
                    .setSecret(request.getTenant().getAdminPassword())
                    .setOwnerId(tenantId)
                    .setType(Type.INDIVIDUAL)
                    .build();

            credentialStoreServiceClient.putCredential(metadata);


            CreateTenantResponse tenantResponse = CreateTenantResponse.newBuilder()
                    .setTenantId(response.getTenantId())
                    .setClientId(resp.getClientId())
                    .setClientSecret(resp.getClientSecret())
                    .setMsg("Use Base64 encoded clientId:clientSecret as auth token for authorization, " +
                            "Credentials are activated after admin approval")
                    .build();

            responseObserver.onNext(tenantResponse);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred at createTenant " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void updateTenantStatus(UpdateStatusRequest request, StreamObserver<UpdateStatusResponse> responseObserver) {
        org.apache.custos.tenant.profile.service.UpdateStatusRequest req = request.getStatus();
        UpdateStatusResponse response = profileClient.updateTenantStatus(req);

        Context ctx = Context.current().fork();
        // Set ctx as the current context within the Runnable
        ctx.run(() -> {
            ServiceCallback callback = new ServiceCallback() {
                @Override
                public void onCompleted(Object obj) {
                    LOGGER.info(" Tenant Activate task finished " + obj.toString());
                }

                @Override
                public void onError(ServiceException ex) {
                    LOGGER.error("Tenant Activation task failed " + ex);
                }
            };

            ServiceChain chain = ServiceChain.newBuilder(tenantActivationTask, callback).build();

            chain.serve(response);
        });

        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }


    @Override
    public void getCredentials(GetCredentialsRequest request, StreamObserver<GetCredentialsResponse> responseObserver) {
        try {

            GetTenantRequest req = GetTenantRequest.newBuilder().setTenantId(request.getTenantId()).build();

            GetTenantResponse tenantRes = profileClient.getTenant(req);

            if (tenantRes.getTenant() == null) {
                responseObserver.onError(Status.NOT_FOUND.withDescription("Invalid Request ").asRuntimeException());
            } else if (!tenantRes.getTenant().getTenantStatus().equals(TenantStatus.ACTIVE)) {

                responseObserver.onError(Status.PERMISSION_DENIED.
                        withDescription("Tenant not yet approved or invalidated").asRuntimeException());
            } else {

                GetAllCredentialsRequest allReq = GetAllCredentialsRequest
                        .newBuilder()
                        .setOwnerId(request.getTenantId())
                        .build();

                GetAllCredentialsResponse response = credentialStoreServiceClient.getAllCredentials(allReq);


                if (response != null && response.getSecretListCount() > 0) {
                    GetCredentialsResponse.Builder builder = GetCredentialsResponse.newBuilder();
                    for (CredentialMetadata metadata : response.getSecretListList()) {
                        if (metadata.getType().name().equals(Type.CILOGON.name())) {
                            builder.setCiLogonClientId(metadata.getId());
                            builder.setCiLogonClientSecret(metadata.getSecret());
                        } else if (metadata.getType().name().equals(Type.IAM.name())) {
                            builder.setIamClientId(metadata.getId());
                            builder.setIamClientSecret(metadata.getSecret());
                        }
                    }
                    GetCredentialsResponse res = builder.build();
                    responseObserver.onNext(res);
                    responseObserver.onCompleted();
                } else {
                    responseObserver.onError(Status.NOT_FOUND.withDescription("Cannot find credentials ").asRuntimeException());
                }
            }
        } catch (Exception ex) {
            String msg = "Error occurred while getting credentials " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void updateTenant(UpdateTenantRequest request, StreamObserver<UpdateTenantResponse> responseObserver) {
        UpdateTenantResponse response = profileClient.updateTenant(request.getRequest());
        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }


    @Override
    public void getAllTenants(Empty request, StreamObserver<GetAllTenantsResponse> responseObserver) {
        GetAllTenantsResponse response = profileClient.getAllTenants();
        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    @Override
    public void getAllTenantsForUser(GetAllTenantsForUserRequest request, StreamObserver<GetAllTenantsForUserResponse> responseObserver) {
        GetAllTenantsForUserResponse response = profileClient.getAllTenantsForUser(request);
        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    @Override
    public void getTenant(GetTenantRequest request, StreamObserver<GetTenantResponse> responseObserver) {
        GetTenantResponse response = profileClient.getTenant(request);

        responseObserver.onNext(response);
        responseObserver.onCompleted();
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


}
