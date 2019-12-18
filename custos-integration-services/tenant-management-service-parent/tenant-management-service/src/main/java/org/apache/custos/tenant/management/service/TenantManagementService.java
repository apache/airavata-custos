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
import io.grpc.stub.StreamObserver;
import org.apache.custos.iam.service.User;
import org.apache.custos.integration.core.ServiceCallback;
import org.apache.custos.integration.core.ServiceChain;
import org.apache.custos.tenant.management.service.TenantManagementServiceGrpc.TenantManagementServiceImplBase;
import org.apache.custos.tenant.management.tasks.AddIamAdminUserTask;
import org.apache.custos.tenant.management.tasks.AddTenantTask;
import org.apache.custos.tenant.profile.service.Gateway;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.apache.custos.tenant.management.service.CreateTenantRequest;
import org.apache.custos.tenant.management.service.Tenant;

@GRpcService
public class TenantManagementService extends TenantManagementServiceImplBase {

    @Autowired
    private AddTenantTask<Gateway, User> addTenantTask;

    @Autowired
    private AddIamAdminUserTask<User, User> addIamAdminUserTask;
    private static final Logger LOGGER = LoggerFactory.getLogger(TenantManagementService.class);

    @Override
    public void createTenant(CreateTenantRequest request, StreamObserver<Tenant> responseObserver) {
        LOGGER.info("Tenant requested " + request.getTenant().getTenantName());

        Context ctx = Context.current().fork();
        // Set ctx as the current context within the Runnable
        ctx.run(() -> {
            ServiceCallback callback = (msg, exception) -> System.out.println("Completing create tenant");


            ServiceChain chain = ServiceChain.newBuilder(addTenantTask, callback).
                    nextTask(addIamAdminUserTask).build();

            Gateway gateway = Gateway.newBuilder().setGatewayId("qweqw").setInternalGatewayId("asdasd").build();


            chain.serve(gateway);
        });



        responseObserver.onNext(request.getTenant());
        responseObserver.onCompleted();
    }
}
