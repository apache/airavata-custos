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

package org.apache.custos.sharing.service;

import io.grpc.stub.StreamObserver;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

@GRpcService
public class SharingService extends org.apache.custos.sharing.service.SharingServiceGrpc.SharingServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(SharingService.class);

    @Override
    public void createEntityType(org.apache.custos.sharing.service.EntityTypeRequest request,
                                 StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {

    }

    @Override
    public void updateEntityType(org.apache.custos.sharing.service.EntityTypeRequest request,
                                 StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {

    }

    @Override
    public void deleteEntityType(org.apache.custos.sharing.service.EntityTypeRequest request,
                                 StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {

    }

    @Override
    public void getEntityType(org.apache.custos.sharing.service.EntityTypeRequest request,
                              StreamObserver<org.apache.custos.sharing.service.EntityType> responseObserver) {

    }

    @Override
    public void getEntityTypes(org.apache.custos.sharing.service.SearchRequest request,
                               StreamObserver<org.apache.custos.sharing.service.EntityTypes> responseObserver) {

    }

    @Override
    public void createPermissionType(org.apache.custos.sharing.service.PermissionTypeRequest request,
                                     StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {

    }

    @Override
    public void updatePermissionType(org.apache.custos.sharing.service.PermissionTypeRequest request,
                                     StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {

    }

    @Override
    public void deletePermissionType(org.apache.custos.sharing.service.PermissionTypeRequest request,
                                     StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {

    }

    @Override
    public void getPermissionType(org.apache.custos.sharing.service.PermissionTypeRequest request,
                                  StreamObserver<org.apache.custos.sharing.service.PermissionType> responseObserver) {

    }

    @Override
    public void getPermissionTypes(org.apache.custos.sharing.service.SearchRequest request,
                                   StreamObserver<org.apache.custos.sharing.service.PermissionTypes> responseObserver) {

    }

    @Override
    public void createEntity(org.apache.custos.sharing.service.Entity request,
                             StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {

    }

    @Override
    public void updateEntity(org.apache.custos.sharing.service.Entity request,
                             StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {

    }

    @Override
    public void getEntity(org.apache.custos.sharing.service.EntityRequest request,
                          StreamObserver<org.apache.custos.sharing.service.Entity> responseObserver) {

    }

    @Override
    public void deleteEntity(org.apache.custos.sharing.service.EntityRequest request,
                             StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {

    }

    @Override
    public void searchEntities(org.apache.custos.sharing.service.SearchRequest request,
                               StreamObserver<org.apache.custos.sharing.service.Entities> responseObserver) {

    }

    @Override
    public void getListOfSharedUsers(org.apache.custos.sharing.service.SharingRequest request,
                                     StreamObserver<org.apache.custos.sharing.service.SharedOwners> responseObserver) {

    }

    @Override
    public void getListOfDirectlyShareUsers(org.apache.custos.sharing.service.SharingRequest request,
                                            StreamObserver<org.apache.custos.sharing.service.SharedOwners> responseObserver) {

    }

    @Override
    public void getListOfSharedGroups(org.apache.custos.sharing.service.SharingRequest request,
                                      StreamObserver<org.apache.custos.sharing.service.SharedOwners> responseObserver) {

    }

    @Override
    public void getListOfDirectlySharedGroups(org.apache.custos.sharing.service.SharingRequest request,
                                              StreamObserver<org.apache.custos.sharing.service.SharedOwners> responseObserver) {

    }

    @Override
    public void shareEntityWithUsers(org.apache.custos.sharing.service.SharingRequest request,
                                     StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {

    }

    @Override
    public void shareEntityWithGroups(org.apache.custos.sharing.service.SharingRequest request,
                                      StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {

    }

    @Override
    public void revokeEntitySharingFromUsers(org.apache.custos.sharing.service.SharingRequest request,
                                             StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {

    }

    @Override
    public void revokeEntitySharingFromGroups(org.apache.custos.sharing.service.SharingRequest request,
                                              StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {

    }

    @Override
    public void userHasAccess(org.apache.custos.sharing.service.SharingRequest request,
                              StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {

    }
}
