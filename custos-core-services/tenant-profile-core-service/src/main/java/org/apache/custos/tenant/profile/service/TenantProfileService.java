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

package org.apache.custos.tenant.profile.service;

import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import org.apache.custos.tenant.profile.mapper.AttributeUpdateMetadataMapper;
import org.apache.custos.tenant.profile.mapper.StatusUpdateMetadataMapper;
import org.apache.custos.tenant.profile.mapper.TenantMapper;
import org.apache.custos.tenant.profile.persistance.model.AttributeUpdateMetadata;
import org.apache.custos.tenant.profile.persistance.model.StatusUpdateMetadata;
import org.apache.custos.tenant.profile.persistance.model.Tenant;
import org.apache.custos.tenant.profile.persistance.respository.*;
import org.apache.custos.tenant.profile.service.TenantProfileServiceGrpc.TenantProfileServiceImplBase;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.Set;

/**
 * This service is responsible for custos gateway management functions
 */
@GRpcService
public class TenantProfileService extends TenantProfileServiceImplBase {
    private static final Logger LOGGER = LoggerFactory.getLogger(TenantProfileService.class);


    @Autowired
    private TenantRepository tenantRepository;

    @Autowired
    private StatusUpdateMetadataRepository statusUpdateMetadataRepository;

    @Autowired
    private AttributeUpdateMetadataRepository attributeUpdateMetadataRepository;

    @Autowired
    private ContactRepository contactRepository;

    @Autowired
    private RedirectURIRepository redirectURIRepository;


    @Override
    public void addTenant(org.apache.custos.tenant.profile.service.Tenant request,
                          StreamObserver<AddTenantResponse> responseObserver) {

        try {
            LOGGER.debug("Add tenant request received for tenant " + TenantMapper.getTenantInfoAsString(request));

            if (!isTenantExist(request.getDomain(), request.getTenantName())) {

                Tenant tenant = TenantMapper.createTenantEntityFromTenant(request);

                tenant.setStatus(TenantStatus.REQUESTED.name());

                Set<StatusUpdateMetadata> metadataSet = StatusUpdateMetadataMapper.
                        createStatusUpdateMetadataEntity(tenant, tenant.getRequesterEmail());

                tenant.setStatusUpdateMetadata(metadataSet);


                Tenant savedTenant = tenantRepository.save(tenant);

                AddTenantResponse response = AddTenantResponse.newBuilder().
                        setTenantId(savedTenant.getId()).build();
                responseObserver.onNext(response);
                responseObserver.onCompleted();
            } else {
                String msg = "Tenant exist with name " + request.getTenantName() + " in domain " + request.getDomain();
                LOGGER.error(msg);
                responseObserver.onError(Status.ALREADY_EXISTS.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Exception occurred while adding the tenant " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }


    }

    @Override
    public void updateTenant(UpdateTenantRequest request, StreamObserver<UpdateTenantResponse> responseObserver) {
        try {
            LOGGER.debug("Update tenant request received for tenant " + TenantMapper.
                    getTenantInfoAsString(request.getTenant()));

            org.apache.custos.tenant.profile.service.Tenant tenant = request.getTenant();

            String updatedBy = request.getUpdatedBy();

            Long tenantId = tenant.getTenantId();


            if (isUpdatable(tenantId, tenant.getDomain(), tenant.getTenantName())) {
                Optional<Tenant> opt = tenantRepository.findById(tenantId);
                Tenant exTenant = opt.get();
                Tenant tenantEntity = TenantMapper.createTenantEntityFromTenant(tenant);

                //Do not update the tenant status

                tenantEntity.setStatus(exTenant.getStatus());


                Set<AttributeUpdateMetadata> metadata = AttributeUpdateMetadataMapper.
                        createAttributeUpdateMetadataEntity(exTenant, tenantEntity, updatedBy);

                tenantEntity.setAttributeUpdateMetadata(metadata);

                contactRepository.deleteAllByTenantId(tenantId);

                redirectURIRepository.deleteAllByTenantId(tenantId);

                tenantRepository.save(tenantEntity);

                UpdateTenantResponse response = UpdateTenantResponse.newBuilder().
                        build().newBuilder().setTenant(tenant).build();
                responseObserver.onNext(response);
                responseObserver.onCompleted();
            } else {
                String msg = "Tenant is not updatable, " +
                        "because tenant may not exist with given Id or there may be a tenant with" +
                        "updated domain and name";
                LOGGER.error(msg);
                responseObserver.onError(Status.ABORTED.withDescription(msg).asRuntimeException());
            }


        } catch (Exception ex) {
            String msg = "Exception occurred while updating the tenant " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getAllTenants(Empty request, StreamObserver<GetAllTenantsResponse> responseObserver) {
        try {
            LOGGER.debug("Get all tenants request received");

            List<Tenant> tenants = tenantRepository.findAll();

            List<org.apache.custos.tenant.profile.service.Tenant> tenantList = new ArrayList<>();

            for (Tenant tenant : tenants) {
                org.apache.custos.tenant.profile.service.Tenant t = TenantMapper.createTenantFromTenantEntity(tenant);
                tenantList.add(t);
            }

            GetAllTenantsResponse response = GetAllTenantsResponse.newBuilder().addAllTenant(tenantList).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Exception occurred while retrieving  tenants " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getAllTenantsForUser(GetAllTenantsForUserRequest request,
                                     StreamObserver<GetAllTenantsForUserResponse> responseObserver) {
        try {
            LOGGER.debug("Get all tenants for user " + request.getRequesterUserName() + " received");

            String username = request.getRequesterUserName();

            List<Tenant> tenants = tenantRepository.findByRequesterEmail(username);

            List<org.apache.custos.tenant.profile.service.Tenant> tenantList = new ArrayList<>();

            for (Tenant tenant : tenants) {
                org.apache.custos.tenant.profile.service.Tenant t = TenantMapper.createTenantFromTenantEntity(tenant);
                tenantList.add(t);
            }

            GetAllTenantsForUserResponse response = GetAllTenantsForUserResponse.newBuilder()
                    .addAllTenant(tenantList).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Exception occurred while retrieving  tenants " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getTenant(GetTenantRequest request, StreamObserver<GetTenantResponse> responseObserver) {
        try {
            LOGGER.debug("Get tenant with Id " + request.getTenantId() + " received");

            Long id = request.getTenantId();

            Optional<Tenant> tenant = tenantRepository.findById(id);
            org.apache.custos.tenant.profile.service.Tenant t = null;
            if (tenant.isPresent()) {
                t = TenantMapper.createTenantFromTenantEntity(tenant.get());
                GetTenantResponse response = GetTenantResponse.newBuilder().setTenant(t).build();
                responseObserver.onNext(response);
                responseObserver.onCompleted();
            } else {
                String msg = "Cannot find the tenant with Id " + t.getTenantId();
                LOGGER.error(msg);
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }


        } catch (Exception ex) {
            String msg = "Exception occurred while retrieving  tenants " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getTenantAttributeUpdateAuditTrail(GetAuditTrailRequest request,
                                                   StreamObserver<GetAttributeUpdateAuditTrailResponse> responseObserver) {
        try {
            LOGGER.debug("Get tenant attribute update audit trail for  " + request.getTenantId());

            Long id = Long.valueOf(request.getTenantId());

            List<AttributeUpdateMetadata> tenantList = attributeUpdateMetadataRepository.findAllByTenantId(id);
            List<TenantAttributeUpdateMetadata> metadata = new ArrayList<>();

            for (AttributeUpdateMetadata attributeUpdateMetadata : tenantList) {

                TenantAttributeUpdateMetadata updatedMetadata = AttributeUpdateMetadataMapper.
                        createAttributeUpdateMetadataFromEntity(attributeUpdateMetadata);
                metadata.add(updatedMetadata);

            }

            GetAttributeUpdateAuditTrailResponse response = GetAttributeUpdateAuditTrailResponse
                    .newBuilder()
                    .addAllMetadata(metadata)
                    .build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Exception occurred while retrieving  attribute status update metadata " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void getTenantStatusUpdateAuditTrail(GetAuditTrailRequest request,
                                                StreamObserver<GetStatusUpdateAuditTrailResponse> responseObserver) {
        try {
            LOGGER.debug("Get tenant attribute update audit trail for  " + request.getTenantId());

            Long id = request.getTenantId();

            List<StatusUpdateMetadata> tenantList = statusUpdateMetadataRepository.findAllByTenantId(id);
            List<TenantStatusUpdateMetadata> metadata = new ArrayList<>();

            for (StatusUpdateMetadata statusUpdateMetadata : tenantList) {

                TenantStatusUpdateMetadata updatedMetadata = StatusUpdateMetadataMapper
                        .createTenantStatusMetadataFrom(statusUpdateMetadata);
                metadata.add(updatedMetadata);

            }

            GetStatusUpdateAuditTrailResponse response = GetStatusUpdateAuditTrailResponse
                    .newBuilder()
                    .addAllMetadata(metadata)
                    .build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Exception occurred while retrieving  status update metadata " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void isTenantExist(IsTenantExistRequest request, StreamObserver<IsTenantExistResponse> responseObserver) {
        try {
            LOGGER.debug("Is tenant exist " + request.getTenantId() + " received");

            Long id = Long.valueOf(request.getTenantId());

            Optional<Tenant> tenant = tenantRepository.findById(id);


            IsTenantExistResponse response = IsTenantExistResponse.newBuilder().setIsExist(tenant.isPresent()).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Exception occurred while retrieving  tenants " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void updateTenantStatus(UpdateStatusRequest request, StreamObserver<UpdateStatusResponse> responseObserver) {
        try {
            LOGGER.debug("Update tenant request status received for " + request.getTenantId() + " received");

            Long id = request.getTenantId();

            String status = request.getStatus().name();

            String updatedBy = request.getUpdatedBy();

            Optional<Tenant> tenant = tenantRepository.findById(id);

            if (tenant.isPresent()) {

                Tenant t = tenant.get();

                t.setStatus(status);

                Set<StatusUpdateMetadata> metadata = StatusUpdateMetadataMapper
                        .createStatusUpdateMetadataEntity(t, updatedBy);
                t.setStatusUpdateMetadata(metadata);

                tenantRepository.save(t);


                UpdateStatusResponse response = UpdateStatusResponse.newBuilder()
                        .setTenantId(id)
                        .setStatus(request.getStatus())
                        .build();
                responseObserver.onNext(response);
                responseObserver.onCompleted();
            }

        } catch (Exception ex) {
            String msg = "Exception occurred while updating tenant status " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    private boolean isTenantExist(String domain, String name) {
        List<Tenant> tenant = tenantRepository.findByDomainAndName(domain, name);
        if (tenant != null && !tenant.isEmpty()) {
            return true;
        }
        return false;
    }

    private boolean isUpdatable(Long tenantId, String domain, String name) {
        List<Tenant> tenantList = tenantRepository.findByDomainAndName(domain, name);

        if (tenantList != null && !tenantList.isEmpty()) {
            Tenant exTeant = tenantList.get(0);
            if (!exTeant.getId().equals(tenantId)) {
                return false;
            }
        }
        Optional<Tenant> opt = tenantRepository.findById(tenantId);
        if (opt.isPresent()) {
            return true;
        } else {
            return false;
        }
    }


}



