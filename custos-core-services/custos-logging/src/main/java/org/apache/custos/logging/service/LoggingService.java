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

package org.apache.custos.logging.service;

import io.grpc.stub.StreamObserver;
import org.apache.custos.logging.mapper.LogEventMapper;
import org.apache.custos.logging.persistance.model.LoggingEnabledStatus;
import org.apache.custos.logging.persistance.repository.LogEventRepository;
import org.apache.custos.logging.persistance.repository.LoggingEnabledStatusRepository;
import org.apache.custos.logging.service.LoggingServiceGrpc.LoggingServiceImplBase;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

@GRpcService
public class LoggingService extends LoggingServiceImplBase {
    private static final Logger LOGGER = LoggerFactory.getLogger(LoggingService.class);

    @Autowired
    private LoggingEnabledStatusRepository loggingEnabledStatusRepository;

    @Autowired
    private LogEventRepository logEventRepository;


    @Override
    public void addLogEvent(org.apache.custos.logging.service.LogEvent request,
                            StreamObserver<org.apache.custos.logging.service.Status> responseObserver) {
        try {

            logEventRepository.save(LogEventMapper.transform(request));

            org.apache.custos.logging.service.Status status =
                    org.apache.custos.logging.service.Status.newBuilder().setStatus(true).build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Exception occurred while adding logging event " + ex;
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void getLogEvents(org.apache.custos.logging.service.LogEventRequest request,
                             StreamObserver<org.apache.custos.logging.service.LogEvents> responseObserver) {
        try {

            List<org.apache.custos.logging.persistance.model.LogEvent> logEvents =
                    logEventRepository.searchEvents(request);

            List<org.apache.custos.logging.service.LogEvent> logEventList = new ArrayList<>();

            if (logEvents != null && logEvents.size() > 0) {
                for (org.apache.custos.logging.persistance.model.LogEvent logEvent : logEvents) {
                    logEventList.add(LogEventMapper.transform(logEvent));
                }
            }

        } catch (Exception ex) {
            String msg = "Exception occurred while fetching log events " + ex;
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void isLogEnabled(org.apache.custos.logging.service.LoggingConfigurationRequest request,
                             StreamObserver<org.apache.custos.logging.service.Status> responseObserver) {
        try {
            Optional<LoggingEnabledStatus> loggingEnabledStatus = loggingEnabledStatusRepository.
                    findById(request.getTenantId());

            boolean status = false;

            if (loggingEnabledStatus.isPresent()) {
                LoggingEnabledStatus stat = loggingEnabledStatus.get();
                status = stat.isStatus();
            }

            org.apache.custos.logging.service.Status res = org.apache.custos.logging.service.Status
                    .newBuilder()
                    .setStatus(status)
                    .build();

            responseObserver.onNext(res);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Exception occurred while checking logging is enabled " + ex;
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void enable(org.apache.custos.logging.service.LoggingConfigurationRequest request,
                       StreamObserver<org.apache.custos.logging.service.Status> responseObserver) {
        try {

            Optional<LoggingEnabledStatus> loggingEnabledStatus = loggingEnabledStatusRepository.
                    findById(request.getTenantId());

            if (loggingEnabledStatus.isEmpty()) {
                LoggingEnabledStatus status = new LoggingEnabledStatus();
                status.setStatus(true);
                status.setTenantId(request.getTenantId());
                loggingEnabledStatusRepository.save(status);
            }

            org.apache.custos.logging.service.Status status = org.apache.custos.logging.service.Status
                    .newBuilder()
                    .setStatus(true)
                    .build();

            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Exception occurred while logs are enabled " + ex;
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }
}
