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

import io.grpc.stub.StreamObserver;
import org.apache.custos.iam.service.User;
import org.apache.custos.tenant.management.service.TenantManagementServiceGrpc.TenantManagementServiceImplBase;
import org.apache.custos.tenant.management.tasks.AddIamAdminUserTask;
import org.apache.custos.tenant.management.tasks.AddTenantTask;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.apache.custos.tenant.profile.service.*;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;


@GRpcService
public class TenantManagementService extends TenantManagementServiceImplBase {

    @Autowired
    private AddTenantTask<Tenant, User> addTenantTask;

    @Autowired
    private TenantProfileClient profileClient;

    @Autowired
    private AddIamAdminUserTask<User, User> addIamAdminUserTask;
    private static final Logger LOGGER = LoggerFactory.getLogger(TenantManagementService.class);

    @Override
    public void createTenant(CreateTenantRequest request, StreamObserver<CreateTenantResponse> responseObserver) {
        LOGGER.info("Tenant requested " + request.getTenant().getTenantName());

        AddTenantResponse response = profileClient.addTenant(request.getTenant());

//        Context ctx = Context.current().fork();
//        // Set ctx as the current context within the Runnable
//        ctx.run(() -> {
//            ServiceCallback callback = (msg, exception) -> System.out.println("Completing create tenant");
//
//
//            ServiceChain chain = ServiceChain.newBuilder(addTenantTask, callback).
//                    nextTask(addIamAdminUserTask).build();
//
//           Tenant tenant =   request.getTenant();
//
//            chain.serve(tenant);
//        });

        CreateTenantResponse tenantResponse = CreateTenantResponse.newBuilder()
                .setTenantId(response.getTenantId())
                .setMsg("Tenant Requested Successfully")
                .build();

        responseObserver.onNext(tenantResponse);
        responseObserver.onCompleted();
    }


    @Override
    public void updateTenant(UpdateTenantRequest request, StreamObserver<UpdateTenantResponse> responseObserver) {
        UpdateTenantResponse response = profileClient.updateTenant(request.getRequest());
        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    @Override
    public void updateTenantStatus(UpdateStatusRequest request, StreamObserver<UpdateStatusResponse> responseObserver) {
        org.apache.custos.tenant.profile.service.UpdateStatusRequest req = request.getStatus();
        UpdateStatusResponse response = profileClient.updateTenantStatus(req);
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
