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

package org.apache.custos.sharing.client;


import io.grpc.ClientInterceptor;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import org.apache.custos.sharing.service.*;
import org.apache.custos.sharing.core.*;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.List;

/**
 * Sharing client for sharing core service
 */
@Component
public class SharingClient {

    private ManagedChannel managedChannel;
    private SharingServiceGrpc.SharingServiceBlockingStub sharingServiceBlockingStub;

    private final List<ClientInterceptor> clientInterceptorList;


    public SharingClient(List<ClientInterceptor> clientInterceptorList,
                         @Value("${core.services.server.hostname:localhost}") String serviceHost,
                         @Value("${core.services.server.port:7070}") int servicePort) {
        this.clientInterceptorList = clientInterceptorList;
        managedChannel = ManagedChannelBuilder.forAddress(
                serviceHost, servicePort).usePlaintext().intercept(clientInterceptorList).build();
        sharingServiceBlockingStub = SharingServiceGrpc.newBlockingStub(managedChannel);

    }


    public Status createEntityType(org.apache.custos.sharing.service.EntityTypeRequest request) {

        return sharingServiceBlockingStub.createEntityType(request);
    }


    public Status updateEntityType(org.apache.custos.sharing.service.EntityTypeRequest request) {
        return sharingServiceBlockingStub.updateEntityType(request);
    }

    public Status deleteEntityType(org.apache.custos.sharing.service.EntityTypeRequest request) {
        return sharingServiceBlockingStub.deleteEntityType(request);

    }


    public EntityType getEntityType(org.apache.custos.sharing.service.EntityTypeRequest request) {
        return sharingServiceBlockingStub.getEntityType(request);

    }


    public EntityTypes getEntityTypes(org.apache.custos.sharing.service.SearchRequest request) {
        return sharingServiceBlockingStub.getEntityTypes(request);
    }


    public Status createPermissionType(org.apache.custos.sharing.service.PermissionTypeRequest request) {

        return sharingServiceBlockingStub.createPermissionType(request);
    }


    public Status updatePermissionType(org.apache.custos.sharing.service.PermissionTypeRequest request) {
        return sharingServiceBlockingStub.updatePermissionType(request);
    }


    public Status deletePermissionType(org.apache.custos.sharing.service.PermissionTypeRequest request) {
        return sharingServiceBlockingStub.deletePermissionType(request);

    }


    public PermissionType getPermissionType(org.apache.custos.sharing.service.PermissionTypeRequest request) {

        return sharingServiceBlockingStub.getPermissionType(request);
    }


    public PermissionTypes getPermissionTypes(org.apache.custos.sharing.service.SearchRequest request) {

        return sharingServiceBlockingStub.getPermissionTypes(request);
    }


    public Status createEntity(org.apache.custos.sharing.service.EntityRequest request) {

        return sharingServiceBlockingStub.createEntity(request);
    }


    public Status updateEntity(org.apache.custos.sharing.service.EntityRequest request) {
        return sharingServiceBlockingStub.updateEntity(request);
    }


    public Status isEntityExists(org.apache.custos.sharing.service.EntityRequest request) {

        return sharingServiceBlockingStub.isEntityExists(request);
    }


    public Entity getEntity(org.apache.custos.sharing.service.EntityRequest request) {
        return sharingServiceBlockingStub.getEntity(request);
    }


    public Status deleteEntity(org.apache.custos.sharing.service.EntityRequest request) {
        return sharingServiceBlockingStub.deleteEntity(request);
    }


    public Entities searchEntities(org.apache.custos.sharing.service.SearchRequest request) {

        return sharingServiceBlockingStub.searchEntities(request);
    }


    public SharedOwners getListOfSharedUsers(org.apache.custos.sharing.service.SharingRequest request) {
        return sharingServiceBlockingStub.getListOfSharedUsers(request);
    }


    public SharedOwners getListOfDirectlySharedUsers(org.apache.custos.sharing.service.SharingRequest request) {
        return sharingServiceBlockingStub.getListOfDirectlySharedUsers(request);
    }


    public SharedOwners getListOfSharedGroups(org.apache.custos.sharing.service.SharingRequest request) {
        return sharingServiceBlockingStub.getListOfSharedGroups(request);

    }


    public SharedOwners getListOfDirectlySharedGroups(org.apache.custos.sharing.service.SharingRequest request) {
        return sharingServiceBlockingStub.getListOfDirectlySharedGroups(request);
    }


    public Status shareEntityWithUsers(org.apache.custos.sharing.service.SharingRequest request) {
        return sharingServiceBlockingStub.shareEntityWithUsers(request);

    }


    public Status shareEntityWithGroups(org.apache.custos.sharing.service.SharingRequest request) {
        return sharingServiceBlockingStub.shareEntityWithGroups(request);
    }


    public Status revokeEntitySharingFromUsers(org.apache.custos.sharing.service.SharingRequest request) {
        return sharingServiceBlockingStub.revokeEntitySharingFromUsers(request);
    }


    public Status revokeEntitySharingFromGroups(org.apache.custos.sharing.service.SharingRequest request) {
        return sharingServiceBlockingStub.revokeEntitySharingFromGroups(request);
    }


    public Status userHasAccess(org.apache.custos.sharing.service.SharingRequest request) {
        return sharingServiceBlockingStub.userHasAccess(request);
    }

    public GetAllDirectSharingsResponse getAllDirectSharings(org.apache.custos.sharing.service.SharingRequest request){
        return sharingServiceBlockingStub.getAllDirectSharings(request);
    }

    public GetAllSharingsResponse getAllSharings(org.apache.custos.sharing.service.SharingRequest request){
        return sharingServiceBlockingStub.getAllSharings(request);
    }



}
