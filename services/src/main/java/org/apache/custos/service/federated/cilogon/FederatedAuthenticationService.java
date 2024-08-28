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

package org.apache.custos.service.federated.cilogon;

import org.apache.custos.core.commons.StatusUpdater;
import org.apache.custos.core.federated.authentication.api.CacheManipulationRequest;
import org.apache.custos.core.federated.authentication.api.ClientMetadata;
import org.apache.custos.core.federated.authentication.api.DeleteClientRequest;
import org.apache.custos.core.federated.authentication.api.GetClientRequest;
import org.apache.custos.core.federated.authentication.api.GetClientResponse;
import org.apache.custos.core.federated.authentication.api.GetInstitutionsResponse;
import org.apache.custos.core.federated.authentication.api.GetOperationsMetadataRequest;
import org.apache.custos.core.federated.authentication.api.GetOperationsMetadataResponse;
import org.apache.custos.core.federated.authentication.api.Institution;
import org.apache.custos.core.federated.authentication.api.InstitutionCacheType;
import org.apache.custos.core.federated.authentication.api.OperationMetadata;
import org.apache.custos.core.federated.authentication.api.RegisterClientResponse;
import org.apache.custos.core.federated.authentication.api.Status;
import org.apache.custos.core.mapper.federated.InstitutionMapper;
import org.apache.custos.core.model.commons.OperationStatus;
import org.apache.custos.core.model.commons.StatusEntity;
import org.apache.custos.core.model.federated.CILogonInstitution;
import org.apache.custos.core.repo.federated.CiLogonInstitutionCacheRepository;
import org.apache.custos.service.exceptions.FederatedAuthenticationServiceException;
import org.apache.custos.service.exceptions.InternalServerException;
import org.apache.custos.service.federated.client.cilogon.CILogonClient;
import org.apache.custos.service.federated.client.cilogon.CILogonResponse;
import org.apache.commons.lang3.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.dao.DuplicateKeyException;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.atomic.AtomicBoolean;

@Service
public class FederatedAuthenticationService {

    private static final Logger LOGGER = LoggerFactory.getLogger(FederatedAuthenticationService.class);

    private final CILogonClient ciLogonClient;

    private final StatusUpdater statusUpdater;

    private final CiLogonInstitutionCacheRepository institutionRepository;


    public FederatedAuthenticationService(CILogonClient ciLogonClient, StatusUpdater statusUpdater, CiLogonInstitutionCacheRepository institutionRepository) {
        this.ciLogonClient = ciLogonClient;
        this.statusUpdater = statusUpdater;
        this.institutionRepository = institutionRepository;
    }

    public RegisterClientResponse addClient(ClientMetadata request) {
        try {
            LOGGER.debug("Request received to addClient for " + request.getTenantId());

            request.getClientId();
            if (StringUtils.isNotBlank(request.getClientId())) {
                try {
                    ciLogonClient.deleteClient(request.getClientId());
                } catch (Exception ex) {
                    LOGGER.debug("Error occurred while deleting client " + request.getClientId());
                }
            }

            String[] scopes = request.getScopeList().toArray(new String[request.getScopeCount()]);
            String contact = request.getContacts(0);


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

            return RegisterClientResponse.newBuilder()
                    .setClientId(response.getClientId())
                    .setClientSecret(response.getClientSecret())
                    .setClientSecretExpiresAt(response.getClientSecretExpiredAt())
                    .setClientRegistrationUri(response.getRegistrationClientURI())
                    .setClientIdIssuedAt(response.getClientIdIssuedAt()).build();

        } catch (Exception ex) {
            String msg = "Error occurred during addClient" + ex;
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(Operations.ADD_CLIENT.name(), OperationStatus.FAILED, request.getTenantId(), request.getPerformedBy());
            throw new FederatedAuthenticationServiceException(msg, ex);
        }
    }

    public void updateClient(ClientMetadata request) {
        try {
            LOGGER.debug("Request received to updateClient for " + request.getTenantId());
            String[] scopes = request.getScopeList().toArray(new String[request.getScopeCount()]);
            String contact = request.getContacts(0);

            ciLogonClient.updateClient(request.getClientId(), request.getTenantName(),
                    request.getRedirectURIsList().toArray(new String[request.getRedirectURIsCount()]),
                    request.getComment(),
                    scopes,
                    request.getTenantURI(),
                    contact);

            statusUpdater.updateStatus(Operations.UPDATE_CLIENT.name(), OperationStatus.SUCCESS, request.getTenantId(), request.getPerformedBy());

        } catch (Exception ex) {
            String msg = "Error occurred during updateClient" + ex;
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(Operations.UPDATE_CLIENT.name(), OperationStatus.FAILED, request.getTenantId(), request.getPerformedBy());
            throw new FederatedAuthenticationServiceException(msg, ex);
        }
    }

    public GetClientResponse getClient(GetClientRequest request) {
        try {
            LOGGER.debug("Request received to getClient for " + request.getTenantId());

            CILogonResponse response = ciLogonClient.getClient(request.getClientId());
            return GetClientResponse.newBuilder()
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

        } catch (Exception ex) {
            String msg = "Error occurred during getClient" + ex;
            LOGGER.error(msg, ex);
            throw new FederatedAuthenticationServiceException(msg, ex);
        }
    }

    public void deleteClient(DeleteClientRequest request) {
        try {
            LOGGER.debug("Request received to deleteClient for " + request.getTenantId());

            ciLogonClient.deleteClient(request.getClientId());
            statusUpdater.updateStatus(Operations.DELETE_CLIENT.name(), OperationStatus.SUCCESS, request.getTenantId(), request.getPerformedBy());

        } catch (Exception ex) {
            String msg = "Error occurred during deleteClient" + ex;
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(Operations.DELETE_CLIENT.name(), OperationStatus.FAILED, request.getTenantId(), request.getPerformedBy());
            throw new FederatedAuthenticationServiceException(msg, ex);
        }
    }

    public GetOperationsMetadataResponse getOperationMetadata(GetOperationsMetadataRequest request) {
        try {
            LOGGER.debug("Calling getOperationMetadata API for traceId " + request.getTraceId());

            List<OperationMetadata> metadata = null;
            List<StatusEntity> entities = statusUpdater.getOperationStatus(request.getTraceId());

            if (entities != null && !entities.isEmpty()) {
                metadata = entities.stream()
                        .map(this::convertFromEntity)
                        .toList();
            }
            return GetOperationsMetadataResponse.newBuilder().addAllMetadata(metadata).build();

        } catch (Exception ex) {
            LOGGER.error("operation failed for " + request.getTraceId());
            throw new FederatedAuthenticationServiceException("operation failed for " + request.getTraceId(), ex);
        }
    }

    public Status addToCache(CacheManipulationRequest request) {
        try {
            LOGGER.debug("Calling addToCache API for tenantId " + request.getTenantId());

            long tenantId = request.getTenantId();
            List<String> ids = request.getInstitutionIdsList();
            InstitutionCacheType type = request.getType();
            List<CILogonInstitution> ciLogonInstitutions = new ArrayList<>();
            ids.forEach(id -> {
                ciLogonInstitutions.add(InstitutionMapper.convert(tenantId, id, type.name(), request.getPerformedBy()));
            });

            for (CILogonInstitution ciLogonInstitution : ciLogonInstitutions) {
                Optional<CILogonInstitution> optionalCILogonInstitution = institutionRepository.findById(ciLogonInstitution.getId());
                if (optionalCILogonInstitution.isPresent()) {
                    String msg = " Duplicate entry with Id  " + ciLogonInstitution.getInstitutionId();
                    LOGGER.error(msg);
                    throw new DuplicateKeyException(msg);
                }

            }
            institutionRepository.saveAll(ciLogonInstitutions);

            return Status.newBuilder().setStatus(true).build();

        } catch (Exception ex) {
            String msg = " Error at federated authentication core service " + ex;
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    public Status removeFromCache(CacheManipulationRequest request) {
        try {
            LOGGER.debug("Calling removeFromCache API for tenantId" + request.getTenantId());

            long tenantId = request.getTenantId();
            List<String> ids = request.getInstitutionIdsList();
            List<CILogonInstitution> ciLogonInstitutions = new ArrayList<>();
            ids.forEach(id -> {
                String savedId = id + "@" + tenantId;
                Optional<CILogonInstitution> ciLogonList = institutionRepository.findById(savedId);
                ciLogonList.ifPresent(ciLogonInstitutions::add);
            });
            institutionRepository.deleteAll(ciLogonInstitutions);

            return Status.newBuilder().setStatus(true).build();

        } catch (Exception ex) {
            String msg = " Error at federated authentication core service " + ex;
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    public GetInstitutionsResponse getFromCache(CacheManipulationRequest request) {
        try {
            LOGGER.debug("Calling getFromCache API for tenantId " + request.getTenantId());

            long tenant = request.getTenantId();
            String type = request.getType().name();
            List<CILogonInstitution> institutions = institutionRepository.findAllByTenantIdAndType(tenant, type);
            List<Institution> institutionList = new ArrayList<>();

            org.apache.custos.service.federated.client.cilogon.CILogonInstitution[] ciLogonInstitutions = ciLogonClient.getInstitutions();

            if (institutions != null && !institutions.isEmpty()) {
                for (CILogonInstitution institution : institutions) {
                    for (org.apache.custos.service.federated.client.cilogon.CILogonInstitution
                            ciLogonInstitution : ciLogonInstitutions) {
                        if (ciLogonInstitution.getEntityId().equals(institution.getInstitutionId())) {
                            institutionList.add(convertCILogonToInstitution(ciLogonInstitution));
                        }
                    }
                }
            }
            return GetInstitutionsResponse.newBuilder().addAllInstitutions(institutionList).build();

        } catch (Exception ex) {
            String msg = " Error at federated authentication core service " + ex;
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }


    public GetInstitutionsResponse getInstitutions(CacheManipulationRequest request) {
        try {
            LOGGER.debug("Calling getInstitutions API for tenantId " + request.getTenantId());

            long tenant = request.getTenantId();
            List<CILogonInstitution> institutions = institutionRepository.findAllByTenantIdAndType(tenant, InstitutionCacheType.ALLOWLIST.name());
            List<CILogonInstitution> blockedInstitutions = institutionRepository.findAllByTenantIdAndType(tenant, InstitutionCacheType.BLOCKLIST.name());

            org.apache.custos.service.federated.client.cilogon.CILogonInstitution[] ciLogonInstitutions = ciLogonClient.getInstitutions();
            List<org.apache.custos.service.federated.client.cilogon.CILogonInstitution> selectedLists = new ArrayList<>();

            if (institutions.isEmpty() && blockedInstitutions.isEmpty()) {
                selectedLists.addAll(Arrays.asList(ciLogonInstitutions));

            } else if (!institutions.isEmpty()) {
                for (org.apache.custos.service.federated.client.cilogon.CILogonInstitution ciLogonInstitution : ciLogonInstitutions) {
                    institutions.forEach(it -> {
                        if (it.getInstitutionId().equals(ciLogonInstitution.getEntityId()) && it.getType().equals(InstitutionCacheType.ALLOWLIST.name())) {
                            selectedLists.add(ciLogonInstitution);
                        }
                    });
                }

            } else {
                for (org.apache.custos.service.federated.client.cilogon.CILogonInstitution ciLogonInstitution : ciLogonInstitutions) {
                    AtomicBoolean doNotAdd = new AtomicBoolean(false);
                    for (CILogonInstitution it : blockedInstitutions) {
                        if (it.getInstitutionId().equals(ciLogonInstitution.getEntityId())) {
                            doNotAdd.set(true);
                            break;
                        }
                    }

                    if (!doNotAdd.get()) {
                        selectedLists.add(ciLogonInstitution);
                    }

                }
            }

            List<Institution> institutionList = new ArrayList<>();
            selectedLists.forEach(sl -> institutionList.add(convertCILogonToInstitution(sl)));

            return GetInstitutionsResponse.newBuilder().addAllInstitutions(institutionList).build();

        } catch (Exception ex) {
            String msg = " Error at federated authentication core service " + ex;
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    private OperationMetadata convertFromEntity(StatusEntity entity) {
        return OperationMetadata.newBuilder()
                .setEvent(entity.getEvent())
                .setStatus(entity.getState())
                .setPerformedBy(entity.getPerformedBy())
                .setTimeStamp(entity.getTime().toString()).build();
    }

    public Institution convertCILogonToInstitution(org.apache.custos.service.federated.client.cilogon.CILogonInstitution ciLogonInstitution) {
        Institution.Builder msg = Institution.newBuilder();
        msg.setEntityId(ciLogonInstitution.getEntityId());
        msg.setDisplayName(ciLogonInstitution.getDisplayName());
        msg.setOrganizationName(ciLogonInstitution.getOrganizationName());
        msg.setRandS(ciLogonInstitution.isRandS());
        return msg.build();
    }

}
