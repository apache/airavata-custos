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

package org.apache.custos.tenant.profile.service;

import io.grpc.stub.StreamObserver;
import org.apache.custos.tenant.profile.service.TenantProfileServiceGrpc.TenantProfileServiceImplBase;
import org.lognet.springboot.grpc.GRpcService;
import org.apache.custos.tenant.profile.service.Gateway;
import org.apache.custos.tenant.profile.service.AddGatewayResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * This service is responsible for custos gateway management functions
 */
@GRpcService
public class TenantProfileService extends TenantProfileServiceImplBase {
    private static final Logger LOGGER = LoggerFactory.getLogger(TenantProfileService.class);

    @Override
    public void addGateway(Gateway gateway, StreamObserver<AddGatewayResponse> responseObserver) {
        LOGGER.info(gateway.getGatewayId() +" gateway deployed successfully");
        responseObserver.onNext(AddGatewayResponse.newBuilder().setCode("success").build());
        responseObserver.onCompleted();
    }

}
