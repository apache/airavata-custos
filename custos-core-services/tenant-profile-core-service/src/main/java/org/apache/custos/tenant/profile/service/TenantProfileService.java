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

import io.grpc.stub.StreamObserver;
import org.apache.custos.tenant.profile.persistance.model.Tenant;
import org.apache.custos.tenant.profile.persistance.respository.TenantRepository;
import org.apache.custos.tenant.profile.service.TenantProfileServiceGrpc.TenantProfileServiceImplBase;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

/**
 * This service is responsible for custos gateway management functions
 */
@GRpcService
public class TenantProfileService extends TenantProfileServiceImplBase {
    private static final Logger LOGGER = LoggerFactory.getLogger(TenantProfileService.class);

    @Autowired
    private TenantRepository tenantRepository;


    @Override
    public void addTenant(org.apache.custos.tenant.profile.service.Tenant request, StreamObserver<AddTenantResponse> responseObserver) {
        super.addTenant(request, responseObserver);
    }

    @Override
    public void updateTenant(UpdateTenantRequest request, StreamObserver<UpdateTenantResponse> responseObserver) {
        super.updateTenant(request, responseObserver);
    }

    @Override
    public void getAllTenants(Empty request, StreamObserver<GetAllTenantsResponse> responseObserver) {
        super.getAllTenants(request, responseObserver);
    }

    @Override
    public void getAllTenantsForUser(GetAllTenantsForUserRequest request, StreamObserver<GetAllTenantsForUserResponse> responseObserver) {
        super.getAllTenantsForUser(request, responseObserver);
    }

    @Override
    public void getTenant(GetTenantRequest request, StreamObserver<GetTenantResponse> responseObserver) {
        super.getTenant(request, responseObserver);
    }

    @Override
    public void getTenantAttributeUpdateAuditTrail(GetAuditTrailRequest request, StreamObserver<GetAttributeUpdateAuditTrailResponse> responseObserver) {
        super.getTenantAttributeUpdateAuditTrail(request, responseObserver);
    }


    @Override
    public void getTenantFullAuditTrail(GetAuditTrailRequest request, StreamObserver<GetFullAuditTrailResponse> responseObserver) {
        super.getTenantFullAuditTrail(request, responseObserver);
    }

    @Override
    public void getTenantStatusUpdateAuditTrail(GetAuditTrailRequest request, StreamObserver<GetStatusUpdateAuditTrailResponse> responseObserver) {
        super.getTenantStatusUpdateAuditTrail(request, responseObserver);
    }

    @Override
    public void isTenantExist(IsTenantExistResponse request, StreamObserver<IsTenantExistResponse> responseObserver) {
        super.isTenantExist(request, responseObserver);
    }

    @Override
    public void updateTenantStatus(UpdateStatusRequest request, StreamObserver<UpdateStatusResponse> responseObserver) {
        super.updateTenantStatus(request, responseObserver);
    }


    // @Override
    public void addTenant(Tenant tenant, StreamObserver<AddTenantResponse> responseObserver) {
//        LOGGER.info(gateway.getGatewayId() + " gateway deployed successfully");
//
//        Tenant tenant = new Tenant();
//        tenant.setName("Test");
//        tenant.setAdminEmail("irjanith@gmail.com");
//        tenant.setAdminFirstName("Isuru");
//        tenant.setAdminLastName("Ranawaka");
//        tenant.setRequesterEmail("irjanith@gmail.com");
//        tenant.setRequesterUsername("IsuruR");
//        tenant.setStatus("REQUESTED");
//
//        String[] contacts = {"0714629880","07895678345"};
//        String[] redirectURIs = {"http://wwww.example.com","https://tenant.com"};
//
//        List<Contact> contactList = new ArrayList<>();
//       for(String contact: contacts) {
//           Contact contact1 = new Contact();
//           contact1.setContactInfo(contact);
//           contact1.setTenant(tenant);
//           contactList.add(contact1);
//       }
//
//        List<RedirectURI> redirectURIS = new ArrayList<>();
//        for(String contact: contacts) {
//            RedirectURI contact1 = new RedirectURI();
//            contact1.setRedirectURI(contact);
//            contact1.setTenant(tenant);
//            redirectURIS.add(contact1);
//        }
//
//        tenant.setRedirectURIS(new HashSet<>(redirectURIS));
//        tenant.setContacts(new HashSet<>(contactList));
//
//       Tenant saved = tenantRepository.save(tenant);
//        LOGGER.info("Saved ID "+ saved.getId());
//       Optional<Tenant> readT =  tenantRepository.findById(saved.getId());
//
//
//       if(readT.isPresent()){
//         Tenant read =  readT.get();
//           LOGGER.info("Contacts " + read.getContacts().toArray());
//           LOGGER.info("RedirectURI " + read.getRedirectURIS().toArray());
//       }else {
//           LOGGER.info("NO saved element");
//       }


        // tenantRepository.deleteById(Long.valueOf("10000101"));

//        responseObserver.onNext(AddTenantResponse.newBuilder().setCode("success").build());
//        responseObserver.onCompleted();
    }

}
