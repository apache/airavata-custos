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

package org.apache.custos.logging.client;

import com.google.common.util.concurrent.ListenableFuture;
import io.grpc.ClientInterceptor;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import org.apache.custos.integration.core.ServiceCallback;
import org.apache.custos.integration.core.ServiceException;
import org.apache.custos.logging.service.*;
import org.apache.custos.logging.service.LogEvent;
import org.apache.custos.logging.service.LogEvents;
import org.apache.custos.logging.service.LoggingConfigurationRequest;
import org.apache.custos.logging.service.LoggingServiceGrpc;
import org.apache.custos.logging.service.Status;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.List;

@Component
public class LoggingClient {
    private ManagedChannel managedChannel;
    private LoggingServiceGrpc.LoggingServiceBlockingStub loggingServiceBlockingStub;
    private LoggingServiceGrpc.LoggingServiceFutureStub loggingServiceFutureStub;

    private final List<ClientInterceptor> clientInterceptorList;


    public LoggingClient(List<ClientInterceptor> clientInterceptorList,
                         @Value("${core.services.server.hostname:localhost}") String serviceHost,
                         @Value("${core.services.server.port:7070}") int servicePort) {
        this.clientInterceptorList = clientInterceptorList;
        managedChannel = ManagedChannelBuilder.forAddress(
                serviceHost, servicePort).usePlaintext().intercept(clientInterceptorList).build();
        loggingServiceBlockingStub = LoggingServiceGrpc.newBlockingStub(managedChannel);
        loggingServiceFutureStub = LoggingServiceGrpc.newFutureStub(managedChannel);


    }


    public ListenableFuture<org.apache.custos.logging.service.Status> addLogEventAsync(LogEvent logEvent) {
        return loggingServiceFutureStub.addLogEvent(logEvent);
    }

    public org.apache.custos.logging.service.Status addLogEvent(LogEvent logEvent) {
        return loggingServiceBlockingStub.addLogEvent(logEvent);
    }

    public org.apache.custos.logging.service.Status enable(LoggingConfigurationRequest loggingConfigurationRequest) {
        return loggingServiceBlockingStub.enable(loggingConfigurationRequest);
    }


    public LogEvents getLogEvents(org.apache.custos.logging.service.LogEventRequest logEventRequest) {
        return loggingServiceBlockingStub.getLogEvents(logEventRequest);
    }

    public Status isLogEnabled(LoggingConfigurationRequest loggingConfigurationRequest) {
        return loggingServiceBlockingStub.isLogEnabled(loggingConfigurationRequest);
    }


    private StreamObserver getObserver(ServiceCallback callback, String failureMsg) {
        final Object[] response = new Object[1];
        StreamObserver observer = new StreamObserver() {
            @Override
            public void onNext(Object o) {
                response[0] = o;
            }

            @Override
            public void onError(Throwable throwable) {
                callback.onError(new ServiceException(failureMsg, throwable, null));
            }

            @Override
            public void onCompleted() {
                callback.onCompleted(response[0]);
            }
        };

        return observer;
    }
}
