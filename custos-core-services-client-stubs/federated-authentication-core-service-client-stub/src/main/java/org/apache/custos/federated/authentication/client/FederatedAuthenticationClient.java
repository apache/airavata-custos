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

package org.apache.custos.federated.authentication.client;

import io.grpc.ClientInterceptor;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import org.apache.custos.federated.authentication.service.*;
import org.apache.custos.integration.core.ServiceCallback;
import org.apache.custos.integration.core.ServiceException;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.List;

/**
 * The client class used to connect IAM admin services
 */
@Component
public class FederatedAuthenticationClient {
    private ManagedChannel managedChannel;
    private FederatedAuthenticationServiceGrpc.FederatedAuthenticationServiceStub federatedAuthenticationServiceStub;
    private FederatedAuthenticationServiceGrpc.FederatedAuthenticationServiceBlockingStub federatedAuthenticationServiceBlockingStub;

    private final List<ClientInterceptor> clientInterceptorList;


    public FederatedAuthenticationClient(List<ClientInterceptor> clientInterceptorList,
                                         @Value("${federated.authentication.service.dns.name}") String serviceHost,
                                         @Value("${federated.authentication.service.port}") int servicePort) {
        this.clientInterceptorList = clientInterceptorList;
        managedChannel = ManagedChannelBuilder.forAddress(
                serviceHost, servicePort).usePlaintext().intercept(clientInterceptorList).build();
        federatedAuthenticationServiceStub = FederatedAuthenticationServiceGrpc.newStub(managedChannel);
        federatedAuthenticationServiceBlockingStub = FederatedAuthenticationServiceGrpc.newBlockingStub(managedChannel);
    }


    public void addClientAsync(ClientMetadata request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "addClient   task failed");
        federatedAuthenticationServiceStub.addClient(request, observer);
    }

    public void updateClientAsync(ClientMetadata request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "updateClient   task failed");
        federatedAuthenticationServiceStub.updateClient(request, observer);
    }

    public void deleteClientAsync(DeleteClientRequest request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "deleteClient   task failed");
        federatedAuthenticationServiceStub.deleteClient(request, observer);
    }

    public void getOperationsMetadataAsync(GetOperationsMetadataRequest request, final ServiceCallback callback) {
        StreamObserver observer = getObserver(callback, "get operations metadata");
        federatedAuthenticationServiceStub.getOperationMetadata(request, observer);
    }

    public void getClientAsync(GetClientRequest request) {
        federatedAuthenticationServiceBlockingStub.getClient(request);
    }

    public RegisterClientResponse addClient(ClientMetadata request) {
        return federatedAuthenticationServiceBlockingStub.addClient(request);
    }

    public Empty updateClient(ClientMetadata request) {
        return federatedAuthenticationServiceBlockingStub.updateClient(request);
    }

    public Empty deleteClient(DeleteClientRequest request) {
        return federatedAuthenticationServiceBlockingStub.deleteClient(request);
    }

    public GetClientResponse getClient(GetClientRequest request) {
        return federatedAuthenticationServiceBlockingStub.getClient(request);
    }

    public GetOperationsMetadataResponse getOperationsMetadata(GetOperationsMetadataRequest request) {
        return federatedAuthenticationServiceBlockingStub.getOperationMetadata(request);
    }

    public Status addToCache(CacheManipulationRequest request) {
        return federatedAuthenticationServiceBlockingStub.addToCache(request);
    }


    public Status removeFromCache(CacheManipulationRequest request) {
        return federatedAuthenticationServiceBlockingStub.removeFromCache(request);
    }

    public GetInstitutionsResponse getFromCache(CacheManipulationRequest request) {
        return federatedAuthenticationServiceBlockingStub.getFromCache(request);
    }

    public GetInstitutionsResponse getInstitutions(CacheManipulationRequest request) {
        return federatedAuthenticationServiceBlockingStub.getInstitutions(request);
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
