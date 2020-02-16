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

package org.apache.custos.credential.store.client;

import io.grpc.ClientInterceptor;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import org.apache.custos.credential.store.service.*;
import org.apache.custos.integration.core.ServiceCallback;
import org.apache.custos.integration.core.ServiceException;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.List;

/**
 * The client class used to connect with credential admin services
 */
@Component
public class CredentialStoreServiceClient {

    private ManagedChannel managedChannel;
    private CredentialStoreServiceGrpc.CredentialStoreServiceStub credentialStoreServiceStub;
    private CredentialStoreServiceGrpc.CredentialStoreServiceBlockingStub credentialStoreServiceBlockingStub;

    private final List<ClientInterceptor> clientInterceptorList;


    public CredentialStoreServiceClient(List<ClientInterceptor> clientInterceptorList,
                                        @Value("${credential.store.service.dns.name}") String serviceHost,
                                        @Value("${credential.store.service.port}") int servicePort) {
        this.clientInterceptorList = clientInterceptorList;
        managedChannel = ManagedChannelBuilder.forAddress(
                serviceHost, servicePort).usePlaintext(true).intercept(clientInterceptorList).build();
        credentialStoreServiceStub = CredentialStoreServiceGrpc.newStub(managedChannel);
        credentialStoreServiceBlockingStub = CredentialStoreServiceGrpc.newBlockingStub(managedChannel);
    }

    public void putCredentialAsync(CredentialMetadata request, final ServiceCallback callback) {
        StreamObserver observer = getObserver(callback, " PutCredentialAsync task failed");
        credentialStoreServiceStub.putCredential(request, observer);

    }


    public void getCredentialAsync(GetCredentialRequest request, final ServiceCallback callback) {
        StreamObserver observer = getObserver(callback, " getCredentialAsync task failed");
        credentialStoreServiceStub.getCredential(request, observer);
    }


    public void getAllCredentialsAsync(GetAllCredentialsRequest request, final ServiceCallback callback) {
        StreamObserver observer = getObserver(callback, " getAllCredentialsAsync task failed");
        credentialStoreServiceStub.getAllCredentials(request, observer);
    }


    public void deleteCredentialAsync(DeleteCredentialRequest request, final ServiceCallback callback) {
        StreamObserver observer = getObserver(callback, " deleteCredentialAsync task failed");
        credentialStoreServiceStub.deleteCredential(request, observer);
    }

    public void getOperationsMetadataAsync(GetOperationsMetadataRequest request, final ServiceCallback callback) {
        StreamObserver observer = getObserver(callback, "get operations metadata");
         credentialStoreServiceStub.getOperationMetadata(request,observer);
    }


    public void getNewCustosCredentialsAsync(GetNewCustosCredentialRequest request, final  ServiceCallback callback) {
        StreamObserver observer = getObserver(callback, "get new custos credentials metadata");
        credentialStoreServiceStub.getNewCustosCredential(request,observer);

    }


    public void getOwnerIdFormToken( TokenRequest request, final  ServiceCallback callback) {
        StreamObserver observer = getObserver(callback, "get new custos credentials metadata");
        credentialStoreServiceStub.getOwnerIdFromToken(request,observer);

    }





    public OperationStatus putCredential(CredentialMetadata request) {
        return credentialStoreServiceBlockingStub.putCredential(request);

    }


    public CredentialMetadata getCredential(GetCredentialRequest request) {
        return credentialStoreServiceBlockingStub.getCredential(request);
    }


    public GetAllCredentialsResponse getAllCredentials(GetAllCredentialsRequest request) {
        return credentialStoreServiceBlockingStub.getAllCredentials(request);
    }


    public OperationStatus deleteCredential(DeleteCredentialRequest request) {
        return credentialStoreServiceBlockingStub.deleteCredential(request);
    }


    public GetOperationsMetadataResponse getOperationsMetadata(GetOperationsMetadataRequest request) {
        return credentialStoreServiceBlockingStub.getOperationMetadata(request);
    }

    public CredentialMetadata getNewCustosCredentials(GetNewCustosCredentialRequest request) {
        return  credentialStoreServiceBlockingStub.getNewCustosCredential(request);
    }

    public GetOwnerIdResponse getOwnerIdFormToken( TokenRequest request) {
       return credentialStoreServiceBlockingStub.getOwnerIdFromToken(request);
    }

    public CredentialMetadata getCustosCredentialFromToken( TokenRequest request) {
        return credentialStoreServiceBlockingStub.getCustosCredentialFromToken(request);
    }

    public CredentialMetadata getCustosCredentialFromClientId (GetCredentialRequest request) {
        return credentialStoreServiceBlockingStub.getCustosCredentialFromClientId(request);
    }
    public GetAllCredentialsResponse getAllCredentialFromToken (TokenRequest request) {
        return credentialStoreServiceBlockingStub.getAllCredentialsFromToken(request);
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
