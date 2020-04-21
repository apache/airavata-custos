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
                          StreamObserver<org.apache.custos.tenant.profile.service.Tenant> responseObserver) {

        try {
            LOGGER.debug("Add tenant request received for tenant " + TenantMapper.getTenantInfoAsString(request));

            Tenant tenant = TenantMapper.createTenantEntityFromTenant(request);

            tenant.setStatus(TenantStatus.REQUESTED.name());

            Set<StatusUpdateMetadata> metadataSet = StatusUpdateMetadataMapper.
                    createStatusUpdateMetadataEntity(tenant, tenant.getRequesterEmail());

            tenant.setStatusUpdateMetadata(metadataSet);


            Tenant savedTenant = tenantRepository.save(tenant);


            org.apache.custos.tenant.profile.service.Tenant
                    response = request.toBuilder().setTenantId(savedTenant.getId()).build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Exception occurred while adding the tenant " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());

        }


    }

    @Override
    public void updateTenant(org.apache.custos.tenant.profile.service.Tenant tenant, StreamObserver<org.apache.custos.tenant.profile.service.Tenant> responseObserver) {
        try {
            LOGGER.debug("Update tenant request received for tenant " + TenantMapper.
                    getTenantInfoAsString(tenant));


            String updatedBy = "Tenant Admin";

            Long tenantId = tenant.getTenantId();

            if (!isUpdatable(tenantId, tenant.getDomain(), tenant.getClientName())) {
                String msg = "Tenant not exist";
                LOGGER.error(msg);
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
                return;
            }

            Optional<Tenant> opt = tenantRepository.findById(tenantId);
            Tenant exTenant = opt.get();
            tenant = tenant.toBuilder().setParentTenantId(exTenant.getParentId()).build();
            Tenant tenantEntity = TenantMapper.createTenantEntityFromTenant(tenant);

            //Do not update the tenant status

            if (tenantEntity.getParentId() > 0) {

                tenantEntity.setStatus(exTenant.getStatus());
            } else {
                tenantEntity.setStatus(TenantStatus.REQUESTED.name());
            }

            tenantEntity.setCreatedAt(exTenant.getCreatedAt());


            Set<AttributeUpdateMetadata> metadata = AttributeUpdateMetadataMapper.
                    createAttributeUpdateMetadataEntity(exTenant, tenantEntity, updatedBy);

            tenantEntity.setAttributeUpdateMetadata(metadata);

            contactRepository.deleteAllByTenantId(tenantId);

            redirectURIRepository.deleteAllByTenantId(tenantId);

            tenantRepository.save(tenantEntity);

            responseObserver.onNext(tenant);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Exception occurred while updating the tenant " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());

        }
    }

    @Override
    public void getAllTenants(GetTenantsRequest request, StreamObserver<GetAllTenantsResponse> responseObserver) {
        try {
            LOGGER.debug("Get all tenants request received");

            String status = null;

            if (request.getStatus() != null && !request.getStatus().name().equals("")){
                status = request.getStatus().name();
            }

            int offset = request.getOffset();
            int limit = request.getLimit();
            long parentId = request.getParentId();

            String requesterEmail = request.getRequesterEmail();

            List<Tenant> tenants = null;

            if (requesterEmail != null && !requesterEmail.equals("")) {
              tenants = tenantRepository.findByRequesterEmail(requesterEmail);
            } else if (status == null && parentId == 0) {
                tenants = tenantRepository.getAllWithPaginate(limit,offset);
            } else if (status != null && parentId == 0){
                tenants = tenantRepository.findByStatusWithPaginate(status,limit,offset);
            } else if (status == null && parentId > 0) {
                tenants = tenantRepository.getAllChildTenantsWithPaginate(parentId,limit,offset);
            } else if (status != null && parentId >0 ) {
                tenants = tenantRepository.findChildTenantsByStatusWithPaginate(status,parentId,limit,offset);
            }


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
            ex.printStackTrace();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getAllTenantsForUser(GetAllTenantsForUserRequest request,
                                     StreamObserver<GetAllTenantsForUserResponse> responseObserver) {
        try {
            LOGGER.debug("Get all tenants for user " + request.getRequesterEmail() + " received");

            String username = request.getRequesterEmail();

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

//        if (tenantList != null && !tenantList.isEmpty()) {
//            Tenant exTeant = tenantList.get(0);
//            if (!exTeant.getId().equals(tenantId)) {
//                return false;
//            }
//        }
        Optional<Tenant> opt = tenantRepository.findById(tenantId);
        if (opt.isPresent()) {
            return true;
        } else {
            return false;
        }
    }


}



