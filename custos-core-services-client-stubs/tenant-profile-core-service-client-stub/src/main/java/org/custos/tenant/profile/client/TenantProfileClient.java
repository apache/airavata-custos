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

package org.custos.tenant.profile.client;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import org.apache.custos.tenant.profile.service.AddGatewayResponse;
import org.apache.custos.tenant.profile.service.Gateway;
import org.apache.custos.tenant.profile.service.TenantProfileServiceGrpc;
import org.apache.custos.tenant.profile.service.TenantProfileServiceGrpc.TenantProfileServiceBlockingStub;

/**
 * This class uses gRPC stubs generated for {@link org.apache.custos.tenant.profile.service.TenantProfileServiceGrpc}
 * and acts as the client
 */
public class TenantProfileClient {

    private ManagedChannel managedChannel;
    private TenantProfileServiceBlockingStub tenantProfileServiceBlockingStub;

    private String tenantProfileServiceAddress;
    private int port;

    public TenantProfileClient(String tenantProfileServiceAddress, int port) {
        this.tenantProfileServiceAddress = tenantProfileServiceAddress;
        this.port = port;
    }

    public String addGateway(String gatewayId, String gatewayInternalId) {

        Gateway gateway = Gateway.newBuilder().
                setGatewayId(gatewayId).
                setInternalGatewayId(gatewayInternalId).build();

        managedChannel = ManagedChannelBuilder.forAddress(
                this.tenantProfileServiceAddress, this.port).usePlaintext(true).build();

        tenantProfileServiceBlockingStub = TenantProfileServiceGrpc.newBlockingStub(managedChannel);
        AddGatewayResponse response = tenantProfileServiceBlockingStub.addGateway(gateway);
        return response.getCode();

    }

}
