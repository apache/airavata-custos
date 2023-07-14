/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing,
 *  software distributed under the License is distributed on an
 *  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *  KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

package org.apache.custos.log.management.service;

import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import org.apache.custos.logging.client.LoggingClient;
import org.apache.custos.logging.service.LogEvents;
import org.apache.custos.log.management.service.LogManagementServiceGrpc;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

@GRpcService
public class LogManagementService extends LogManagementServiceGrpc.LogManagementServiceImplBase {
    private static final Logger LOGGER = LoggerFactory.getLogger(LogManagementService.class);

    @Autowired
    private LoggingClient loggingClient;


    @Override
    public void getLogEvents(org.apache.custos.logging.service.LogEventRequest request,
                             StreamObserver<org.apache.custos.logging.service.LogEvents> responseObserver) {
        LOGGER.debug("Request received to getLogEvents of tenant  " + request.getTenantId());
        try {
            LogEvents logEvents = loggingClient.getLogEvents(request);
            responseObserver.onNext(logEvents);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while pulling secretes " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void isLogEnabled(org.apache.custos.logging.service.LoggingConfigurationRequest request,
                             StreamObserver<org.apache.custos.logging.service.Status> responseObserver) {
        LOGGER.debug("Request received to isLogEnabled of tenant  " + request.getTenantId());
        try {

            org.apache.custos.logging.service.Status status = loggingClient.isLogEnabled(request);
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while pulling secretes " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void enable(org.apache.custos.logging.service.LoggingConfigurationRequest request,
                       StreamObserver<org.apache.custos.logging.service.Status> responseObserver) {
        LOGGER.debug("Request received to enable of tenant  " + request.getTenantId());
        try {

            org.apache.custos.logging.service.Status status = loggingClient.enable(request);
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while pulling secretes " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }
}
