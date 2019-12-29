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

package org.apache.custos.federated.authentication.service;

import io.grpc.stub.StreamObserver;
import org.apache.custos.federated.authentication.exceptions.FederatedAuthenticationServiceException;
import org.apache.custos.federated.authentication.persistance.model.EventMetadata;
import org.apache.custos.federated.authentication.persistance.repository.EventRepository;
import org.apache.custos.federated.authentication.service.FederatedAuthenticationServiceGrpc.FederatedAuthenticationServiceImplBase;
import org.apache.custos.federated.authentication.utils.OperationStatus;
import org.apache.custos.federated.authentication.utils.Operations;
import org.apache.custos.federated.services.clients.cilogon.CILogonClient;
import org.apache.custos.federated.services.clients.cilogon.CILogonResponse;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import org.apache.custos.federated.authentication.service.ClientMetadata;
import org.apache.custos.federated.authentication.service.GetClientRequest;
import org.apache.custos.federated.authentication.service.DeleteClientRequest;

import org.apache.custos.federated.authentication.service.RegisterClientResponse;
import org.apache.custos.federated.authentication.service.Empty;
import org.apache.custos.federated.authentication.service.GetClientResponse;
import org.springframework.beans.factory.annotation.Autowired;

import java.util.Arrays;

@GRpcService
public class FederatedAuthenticationService extends FederatedAuthenticationServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(FederatedAuthenticationService.class);

    @Autowired
    private CILogonClient ciLogonClient;

    @Autowired
    private EventRepository repository;

    @Override
    public void addClient(ClientMetadata request, StreamObserver<RegisterClientResponse> responseObserver) {
        EventMetadata metadata = new EventMetadata();
        metadata.setEvent(Operations.ADD_CLIENT.name());
        metadata.setTraceId(request.getTenantId());
        metadata.setPerformedBy(request.getPerformedBy());

        try {
            LOGGER.debug("Request received to addClient for " + request.getTenantId());
            String[] scopes = request.getScopeList() != null ?
                           request.getScopeList().toArray(new String[request.getScopeCount()]) : new String[0];
            String  contact = request.getContactsList() != null ? request.getContacts(0): null;


           CILogonResponse response =  ciLogonClient.registerClient(request.getTenantName(),
                     request.getRedirectURIsList().toArray(new String[request.getRedirectURIsCount()]),
                     request.getComment(),
                     scopes,
                     request.getTenantURI(),
                     contact);

           metadata.setState(OperationStatus.SUCCESS.name());
           repository.save(metadata);

           RegisterClientResponse registerClientResponse = RegisterClientResponse.newBuilder()
                                                            .setClientId(response.getClientId())
                                                            .setClientSecret(response.getClientSecret())
                                                            .setClientSecretExpiresAt(response.getClientSecretExpiredAt())
                                                             .setClientIdIssuedAt(response.getClientIdIssuedAt()).build();

           responseObserver.onNext(registerClientResponse);
           responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during addClient" + ex;
            LOGGER.error(msg, ex);
            metadata.setState(OperationStatus.FAILED.name());
            repository.save(metadata);
            FederatedAuthenticationServiceException exception = new FederatedAuthenticationServiceException(msg, ex);
            responseObserver.onError(exception);
        }
    }

    @Override
    public void updateClient(ClientMetadata request, StreamObserver<Empty> responseObserver) {
        EventMetadata metadata = new EventMetadata();
        metadata.setEvent(Operations.UPDATE_CLIENT.name());
        metadata.setTraceId(request.getTenantId());
        metadata.setPerformedBy(request.getPerformedBy());
        try {
            LOGGER.debug("Request received to updateClient for " + request.getTenantId());
            String[] scopes = request.getScopeList() != null ?
                    request.getScopeList().toArray(new String[request.getScopeCount()]) : new String[0];
            String  contact = request.getContactsList() != null ? request.getContacts(0): null;

             ciLogonClient.updateClient(request.getClientId(),request.getTenantName(),
                    request.getRedirectURIsList().toArray(new String[request.getRedirectURIsCount()]),
                    request.getComment(),
                    scopes,
                    request.getTenantURI(),
                    contact);

            metadata.setState(OperationStatus.SUCCESS.name());
            repository.save(metadata);

            Empty empty = Empty.newBuilder().build();

            responseObserver.onNext(empty);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred during updateClient" + ex;
            LOGGER.error(msg, ex);
            metadata.setState(OperationStatus.FAILED.name());
            repository.save(metadata);
            FederatedAuthenticationServiceException exception = new FederatedAuthenticationServiceException(msg, ex);
            responseObserver.onError(exception);
        }
    }

    @Override
    public void getClient(GetClientRequest request, StreamObserver<GetClientResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getClient for " + request.getTenantId());

            CILogonResponse response = ciLogonClient.getClient(request.getClientId());
            GetClientResponse getClientResponse = GetClientResponse.newBuilder()
                                                   .setClientId(response.getClientId())
                                                    .setClientName(response.getClientName())
                                                    .addAllRedirectURIs(Arrays.asList(response.getRedirectURIs()))
                                                     .addAllScope(Arrays.asList(response.getScope()))
                                                      .setComment(response.getComment())
                                                      .setClientIdIssuedAt(response.getClientIdIssuedAt())
                                                    .addAllGrantTypes(Arrays.asList(response.getGrantTypes()))
                                                     .build();

            responseObserver.onNext(getClientResponse);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during getClient" + ex;
            LOGGER.error(msg, ex);
            FederatedAuthenticationServiceException exception = new FederatedAuthenticationServiceException(msg, ex);
            responseObserver.onError(exception);
        }
    }

    @Override
    public void deleteClient(DeleteClientRequest request, StreamObserver<Empty> responseObserver) {
        EventMetadata metadata = new EventMetadata();
        metadata.setEvent(Operations.UPDATE_CLIENT.name());
        metadata.setTraceId(request.getTenantId());
        metadata.setPerformedBy(request.getPerformedBy());
        try {
            LOGGER.debug("Request received to deleteClient for " + request.getTenantId());

             ciLogonClient.deleteClient(request.getClientId());

            metadata.setState(OperationStatus.SUCCESS.name());
            repository.save(metadata);

            Empty empty = Empty.newBuilder().build();

            responseObserver.onNext(empty);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during deleteClient" + ex;
            LOGGER.error(msg, ex);
            metadata.setState(OperationStatus.FAILED.name());
            repository.save(metadata);
            FederatedAuthenticationServiceException exception = new FederatedAuthenticationServiceException(msg, ex);
            responseObserver.onError(exception);
        }
    }
}
