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
import org.apache.custos.core.services.commons.StatusUpdater;
import org.apache.custos.core.services.commons.persistance.model.OperationStatus;
import org.apache.custos.core.services.commons.persistance.model.StatusEntity;
import org.apache.custos.federated.authentication.exceptions.FederatedAuthenticationServiceException;
import org.apache.custos.federated.authentication.mapper.ModelMapper;
import org.apache.custos.federated.authentication.persistance.model.CILogonInstitution;
import org.apache.custos.federated.authentication.persistance.repository.CiLogonInstitutionCacheRepository;
import org.apache.custos.federated.authentication.service.FederatedAuthenticationServiceGrpc.FederatedAuthenticationServiceImplBase;
import org.apache.custos.federated.authentication.utils.Operations;
import org.apache.custos.federated.services.clients.cilogon.CILogonClient;
import org.apache.custos.federated.services.clients.cilogon.CILogonResponse;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.atomic.AtomicBoolean;

@GRpcService
public class FederatedAuthenticationService extends FederatedAuthenticationServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(FederatedAuthenticationService.class);

    @Autowired
    private CILogonClient ciLogonClient;

    @Autowired
    private StatusUpdater statusUpdater;

    @Autowired
    private CiLogonInstitutionCacheRepository institutionRespository;

    @Override
    public void addClient(ClientMetadata request, StreamObserver<RegisterClientResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to addClient for " + request.getTenantId());

            if (request.getClientId() != null && !request.getClientId().trim().equals("")) {

                try {
                    ciLogonClient.deleteClient(request.getClientId());
                } catch (Exception ex) {
                    LOGGER.debug("Error occurred while deleting client " + request.getClientId());
                }
            }

            String[] scopes = request.getScopeList() != null ?
                    request.getScopeList().toArray(new String[request.getScopeCount()]) : new String[0];
            String contact = request.getContactsList() != null ? request.getContacts(0) : null;


            CILogonResponse response = ciLogonClient.registerClient(request.getTenantName(),
                    request.getRedirectURIsList().toArray(new String[request.getRedirectURIsCount()]),
                    request.getComment(),
                    scopes,
                    request.getTenantURI(),
                    contact);

            statusUpdater.updateStatus(Operations.ADD_CLIENT.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(),
                    request.getPerformedBy());

            RegisterClientResponse registerClientResponse = RegisterClientResponse.newBuilder()
                    .setClientId(response.getClientId())
                    .setClientSecret(response.getClientSecret())
                    .setClientSecretExpiresAt(response.getClientSecretExpiredAt())
                    .setClientRegistrationUri(response.getRegistrationClientURI())
                    .setClientIdIssuedAt(response.getClientIdIssuedAt()).build();

            responseObserver.onNext(registerClientResponse);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during addClient" + ex;
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(Operations.ADD_CLIENT.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());
            FederatedAuthenticationServiceException exception = new FederatedAuthenticationServiceException(msg, ex);
            responseObserver.onError(exception);
        }
    }

    @Override
    public void updateClient(ClientMetadata request, StreamObserver<Empty> responseObserver) {
        try {
            LOGGER.debug("Request received to updateClient for " + request.getTenantId());
            String[] scopes = request.getScopeList() != null ?
                    request.getScopeList().toArray(new String[request.getScopeCount()]) : new String[0];
            String contact = request.getContactsList() != null ? request.getContacts(0) : null;

            ciLogonClient.updateClient(request.getClientId(), request.getTenantName(),
                    request.getRedirectURIsList().toArray(new String[request.getRedirectURIsCount()]),
                    request.getComment(),
                    scopes,
                    request.getTenantURI(),
                    contact);

            statusUpdater.updateStatus(Operations.UPDATE_CLIENT.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(),
                    request.getPerformedBy());

            Empty empty = Empty.newBuilder().build();

            responseObserver.onNext(empty);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred during updateClient" + ex;
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(Operations.UPDATE_CLIENT.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());
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
                    .setClientRegistrationUri(response.getRegistrationClientURI())
                    .setClientSecretExpiresAt(response.getClientSecretExpiredAt())
                    .setClientSecret(response.getClientSecret())
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

        try {
            LOGGER.debug("Request received to deleteClient for " + request.getTenantId());

            ciLogonClient.deleteClient(request.getClientId());

            statusUpdater.updateStatus(Operations.DELETE_CLIENT.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(),
                    request.getPerformedBy());

            Empty empty = Empty.newBuilder().build();

            responseObserver.onNext(empty);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during deleteClient" + ex;
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(Operations.DELETE_CLIENT.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());
            FederatedAuthenticationServiceException exception = new FederatedAuthenticationServiceException(msg, ex);
            responseObserver.onError(exception);
        }
    }

    @Override
    public void getOperationMetadata(GetOperationsMetadataRequest request, StreamObserver<GetOperationsMetadataResponse> responseObserver) {
        try {
            LOGGER.debug("Calling getOperationMetadata API for traceId " + request.getTraceId());

            List<OperationMetadata> metadata = new ArrayList<>();
            List<StatusEntity> entities = statusUpdater.getOperationStatus(request.getTraceId());
            if (entities == null || entities.size() > 0) {

                for (StatusEntity statusEntity : entities) {
                    OperationMetadata data = convertFromEntity(statusEntity);
                    metadata.add(data);
                }
            }

            GetOperationsMetadataResponse response = GetOperationsMetadataResponse
                    .newBuilder()
                    .addAllMetadata(metadata)
                    .build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            FederatedAuthenticationServiceException failedException = new FederatedAuthenticationServiceException(" operation failed for "
                    + request.getTraceId(), ex);
            responseObserver.onError(failedException);
        }
    }


    @Override
    public void addToCache(InstitutionOperationRequest request, StreamObserver<Status> responseObserver) {
        try {
            LOGGER.debug("Calling addToCache API for tenantId " + request.getTenantId());


            long tenantId = request.getTenantId();

            List<String> ids = request.getInstitutionIdList();

            InstitutionCacheType type = request.getType();

            List<CILogonInstitution> ciLogonInstitutions = new ArrayList<>();

            ids.forEach(id -> {
                ciLogonInstitutions.add(ModelMapper.convert(tenantId, id, type.name(), request.getPerformedBy()));

            });

            institutionRespository.saveAll(ciLogonInstitutions);

            Status status = Status.newBuilder().setStatus(true).build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            FederatedAuthenticationServiceException failedException =
                    new FederatedAuthenticationServiceException(" operation failed for "
                            + request.getTenantId(), ex);
            responseObserver.onError(failedException);
        }
    }

    @Override
    public void removeFromCache(InstitutionOperationRequest request, StreamObserver<Status> responseObserver) {
        try {
            LOGGER.debug("Calling removeFromCache API for tenantId" + request.getTenantId());
            long tenantId = request.getTenantId();

            List<String> ids = request.getInstitutionIdList();

            List<CILogonInstitution> ciLogonInstitutions = new ArrayList<>();

            ids.forEach(id -> {

                String savedId = id + "@" + tenantId;

                Optional<CILogonInstitution> ciLogonList = institutionRespository.findById(savedId);

                if (ciLogonList.isPresent()) {
                    ciLogonInstitutions.add(ciLogonList.get());
                }


            });

            institutionRespository.deleteAll(ciLogonInstitutions);

            Status status = Status.newBuilder().setStatus(true).build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            FederatedAuthenticationServiceException failedException =
                    new FederatedAuthenticationServiceException(" operation failed for "
                            + request.getTenantId(), ex);
            responseObserver.onError(failedException);
        }
    }

    @Override
    public void getFromCache(InstitutionOperationRequest request, StreamObserver<GetInstitutionsResponse> responseObserver) {
        try {
            LOGGER.debug("Calling getFromCache API for tenantId " + request.getTenantId());


            long tenant = request.getTenantId();

            String type = request.getType().name();

            List<CILogonInstitution> institutions = institutionRespository.findAllByTenantIdAndType(tenant, type);

            List<Institution> institutionList = new ArrayList<>();

            if (institutions != null && !institutions.isEmpty()) {

                for (CILogonInstitution institution : institutions) {
                    institutionList.add(ModelMapper.convert(institution));
                }
            }


            GetInstitutionsResponse response = GetInstitutionsResponse.
                    newBuilder().addAllInstitutions(institutionList).build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            FederatedAuthenticationServiceException failedException =
                    new FederatedAuthenticationServiceException(" operation failed for "
                            + request.getTenantId(), ex);
            responseObserver.onError(failedException);
        }
    }


    @Override
    public void getInstitutions(InstitutionOperationRequest request, StreamObserver<GetInstitutionsResponse> responseObserver) {
        try {
            LOGGER.debug("Calling getInstitutions API for tenantId " + request.getTenantId());

            long tenant = request.getTenantId();

            String type = request.getType().name();

            List<CILogonInstitution> institutions = institutionRespository.findAllByTenantIdAndType(tenant, type);


            org.apache.custos.federated.services.clients.cilogon.CILogonInstitution[] ciLogonInstitutions =
                    ciLogonClient.getInstitutions();

            List<org.apache.custos.federated.services.clients.cilogon.CILogonInstitution> selectedLists = new ArrayList<>();

            if (type.equals(InstitutionCacheType.WHITELIST.name())) {

                for (org.apache.custos.federated.services.clients.cilogon.CILogonInstitution ciLogonInstitution : ciLogonInstitutions) {

                    institutions.forEach(it -> {

                        if (it.getInstitutionId().equals(ciLogonInstitution.getEntityId()) &&
                                it.getType().equals(InstitutionCacheType.WHITELIST.name())) {
                            selectedLists.add(ciLogonInstitution);
                        }
                    });

                }

            } else {

                for (org.apache.custos.federated.services.clients.cilogon.CILogonInstitution ciLogonInstitution : ciLogonInstitutions) {

                    AtomicBoolean doNotAdd = new AtomicBoolean(false);
                    institutions.forEach(it -> {

                        if (it.getInstitutionId().equals(ciLogonInstitution.getEntityId())) {
                            doNotAdd.set(true);
                        }
                    });

                    if (!doNotAdd.get()) {
                        selectedLists.add(ciLogonInstitution);
                    }

                }
            }

            List<Institution> institutionList = new ArrayList<>();

            selectedLists.forEach(sl -> {

                institutionList.add(ModelMapper.convert(sl));

            });

            GetInstitutionsResponse response = GetInstitutionsResponse
                    .newBuilder().addAllInstitutions(institutionList).build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            FederatedAuthenticationServiceException failedException =
                    new FederatedAuthenticationServiceException(" operation failed for "
                            + request.getTenantId(), ex);
            responseObserver.onError(failedException);
        }
    }

    private OperationMetadata convertFromEntity(StatusEntity entity) {
        return OperationMetadata.newBuilder()
                .setEvent(entity.getEvent())
                .setStatus(entity.getState())
                .setPerformedBy(entity.getPerformedBy())
                .setTimeStamp(entity.getTime().toString()).build();
    }

}
