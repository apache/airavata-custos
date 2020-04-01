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

package org.apache.custos.agent.profile.client;

import io.grpc.ClientInterceptor;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import org.apache.custos.agent.profile.service.Agent;
import org.apache.custos.agent.profile.service.AgentProfileServiceGrpc;
import org.apache.custos.agent.profile.service.AgentRequest;
import org.apache.custos.agent.profile.service.OperationStatus;
import org.apache.custos.integration.core.ServiceCallback;
import org.apache.custos.integration.core.ServiceException;
import org.springframework.beans.factory.annotation.Value;

import java.util.List;

public class AgentProfileClient {

    private ManagedChannel managedChannel;
    private AgentProfileServiceGrpc.AgentProfileServiceStub agentProfileServiceStub;
    private AgentProfileServiceGrpc.AgentProfileServiceBlockingStub agentProfileServiceBlockingStub;


    private final List<ClientInterceptor> clientInterceptorList;


    public AgentProfileClient(List<ClientInterceptor> clientInterceptorList,
                              @Value("${agent.profile.core.service.dns.name}") String serviceHost,
                              @Value("${agent.profile.core.service.port}") int servicePort) {
        this.clientInterceptorList = clientInterceptorList;
        managedChannel = ManagedChannelBuilder.forAddress(
                serviceHost, servicePort).usePlaintext(true).intercept(clientInterceptorList).build();
        agentProfileServiceStub = AgentProfileServiceGrpc.newStub(managedChannel);
        agentProfileServiceBlockingStub = AgentProfileServiceGrpc.newBlockingStub(managedChannel);
    }


    public Agent createAgent(AgentRequest profile) {
        return agentProfileServiceBlockingStub.createAgent(profile);
    }


    public Agent updateAgent(AgentRequest profile) {
        return agentProfileServiceBlockingStub.updateAgent(profile);
    }


    public OperationStatus deleteAgent(AgentRequest request) {
        return agentProfileServiceBlockingStub.deleteAgent(request);
    }


    public Agent getAgent(AgentRequest request) {
        return agentProfileServiceBlockingStub.getAgent(request);
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
