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

package org.apache.custos.integration.services.commons.interceptors;

import com.google.protobuf.Descriptors;
import com.google.protobuf.GeneratedMessageV3;
import io.grpc.Metadata;
import org.apache.custos.integration.core.interceptor.IntegrationServiceInterceptor;
import org.apache.custos.integration.core.utils.Constants;
import org.apache.custos.logging.client.LoggingClient;
import org.apache.custos.logging.service.LogEvent;
import org.apache.custos.logging.service.LoggingConfigurationRequest;
import org.apache.custos.logging.service.Status;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.Map;

/**
 * API calling event capturing Interceptor
 */
@Component
public class LoggingInterceptor implements IntegrationServiceInterceptor {

    private static final Logger LOGGER = LoggerFactory.getLogger(LoggingInterceptor.class);

    private static final String X_FORWARDED_FOR = "x-forwarded-for";

    @Autowired
    private LoggingClient loggingClient;


    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT msg) {


        if (msg instanceof GeneratedMessageV3) {
            Map<Descriptors.FieldDescriptor, Object> fieldDescriptors = ((GeneratedMessageV3) msg).getAllFields();

            LogEvent.Builder logEventBuilder = LogEvent.newBuilder();
            LoggingConfigurationRequest.Builder loggingConfigurationBuilder =
                    LoggingConfigurationRequest.newBuilder();
            boolean tenantMatched = false;
            boolean clientMatched = false;

            for (Descriptors.FieldDescriptor descriptor : fieldDescriptors.keySet()) {

                if (descriptor.getName().equals("tenant_id") || descriptor.getName().equals("tenantId")) {
                    logEventBuilder.setTenantId((long) fieldDescriptors.get(descriptor));
                    loggingConfigurationBuilder.setTenantId((long) fieldDescriptors.get(descriptor));
                    tenantMatched = true;
                } else if (descriptor.getName().equals("client_id") || descriptor.getName().equals("clientId")) {
                    logEventBuilder.setClientId((String) fieldDescriptors.get(descriptor));
                    loggingConfigurationBuilder.setClientId((String) fieldDescriptors.get(descriptor));
                    clientMatched = true;
                } else if (descriptor.getName().equals("username") || descriptor.getName().equals("userId")) {
                    logEventBuilder.setUsername((String) fieldDescriptors.get(descriptor));
                }

                String servicename = headers.
                        get(Metadata.Key.of(Constants.SERVICE_NAME, Metadata.ASCII_STRING_MARSHALLER));
                String externaIP = headers.
                        get(Metadata.Key.of(Constants.X_FORWARDED_FOR, Metadata.ASCII_STRING_MARSHALLER));
                logEventBuilder.setServiceName(servicename);
                logEventBuilder.setEventType(method);
                logEventBuilder.setExternalIp(externaIP);
                logEventBuilder.setCreatedTime(System.currentTimeMillis());
            }

            if (tenantMatched && clientMatched) {
                Status status = loggingClient.isLogEnabled(loggingConfigurationBuilder.build());
                if (status.getStatus()) {
                    loggingClient.addLogEventAsync(logEventBuilder.build());
                }
            }
        }
        return msg;
    }
}
