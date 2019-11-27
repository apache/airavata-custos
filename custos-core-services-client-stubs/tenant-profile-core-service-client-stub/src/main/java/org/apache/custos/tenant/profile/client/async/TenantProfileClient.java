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

package org.apache.custos.tenant.profile.client.async;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import org.apache.custos.integration.core.ServiceCallback;
import org.apache.custos.integration.core.ServiceException;
import org.apache.custos.tenant.profile.service.Gateway;
import org.apache.custos.tenant.profile.service.TenantProfileServiceGrpc;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

/**
 * This class uses gRPC stubs generated for {@link TenantProfileServiceGrpc}
 * and acts as the client
 */
@Component
public class TenantProfileClient {

    private ManagedChannel managedChannel;
    private TenantProfileServiceGrpc.TenantProfileServiceStub tenantProfileServiceBlockingStub;

    @Value("${tenant.profile.core.service.dns.name}")
    private String serviceHost;

    @Value("${tenant.profile.core.service.port}")
    private int servicePort;


    public void addGateway(Gateway gateway, ServiceCallback callback) {


        StreamObserver observer = new StreamObserver() {
            @Override
            public void onNext(Object o) {
                //TODO implement this
            }

            @Override
            public void onError(Throwable throwable) {
                throwable.printStackTrace();
                callback.onCompleted(null, new ServiceException("Add tenant task failed", throwable, null));
            }

            @Override
            public void onCompleted() {
                System.out.println("Add gateway called");
                callback.onCompleted("Completed", null);
            }
        };

        managedChannel = ManagedChannelBuilder.forAddress(
                this.serviceHost, this.servicePort).usePlaintext(true).build();

        tenantProfileServiceBlockingStub = TenantProfileServiceGrpc.newStub(managedChannel);
        tenantProfileServiceBlockingStub.addGateway(gateway, observer);


    }

}
