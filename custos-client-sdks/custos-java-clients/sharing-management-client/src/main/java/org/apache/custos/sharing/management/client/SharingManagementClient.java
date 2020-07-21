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

package org.apache.custos.sharing.management.client;

import io.grpc.ManagedChannel;
import io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.NettyChannelBuilder;
import io.grpc.stub.MetadataUtils;
import org.apache.custos.clients.core.ClientUtils;
import org.apache.custos.sharing.management.service.SharingManagementServiceGrpc;
import org.apache.custos.sharing.service.*;

import java.io.IOException;

/**
 * Java client to connect with SharingManagementClient
 */
public class SharingManagementClient {

    private ManagedChannel managedChannel;

    private SharingManagementServiceGrpc.SharingManagementServiceBlockingStub blockingStub;


    public SharingManagementClient(String serviceHost, int servicePort, String clientId,
                                   String clientSecret) throws IOException {
        managedChannel = NettyChannelBuilder.forAddress(serviceHost, servicePort)
                .sslContext(GrpcSslContexts
                        .forClient()
                        .trustManager(ClientUtils.getServerCertificate(serviceHost, clientId, clientSecret)) // public key
                        .build())
                .build();

        blockingStub = SharingManagementServiceGrpc.newBlockingStub(managedChannel);
        blockingStub = MetadataUtils.attachHeaders(blockingStub, ClientUtils.getAuthorizationHeader(clientId, clientSecret));
    }


    public Status createEntityType(String clientId, EntityType entityType) {

        EntityTypeRequest entityTypeRequest = EntityTypeRequest
                .newBuilder()
                .setEntityType(entityType)
                .setClientId(clientId)
                .build();
        return blockingStub.createEntityType(entityTypeRequest);

    }


    public Status updateEntityType(String clientId, EntityType entityType) {
        EntityTypeRequest entityTypeRequest = EntityTypeRequest
                .newBuilder()
                .setEntityType(entityType)
                .setClientId(clientId)
                .build();
        return blockingStub.createEntityType(entityTypeRequest);
    }


    public Status deleteEntityType(String clientId, EntityType entityType) {
        EntityTypeRequest entityTypeRequest = EntityTypeRequest
                .newBuilder()
                .setEntityType(entityType)
                .setClientId(clientId)
                .build();
        return blockingStub.deleteEntityType(entityTypeRequest);
    }


    public EntityType getEntityType(String clientId, EntityType entityType) {
        EntityTypeRequest entityTypeRequest = EntityTypeRequest
                .newBuilder()
                .setEntityType(entityType)
                .setClientId(clientId)
                .build();
        return blockingStub.getEntityType(entityTypeRequest);
    }


    public EntityTypes getEntityTypes(String clientId, SearchRequest request) {
        request = request.toBuilder().setClientId(clientId).build();

        return blockingStub.getEntityTypes(request);
    }


    public Status createPermissionType(String clientId, PermissionType permissionType) {
        PermissionTypeRequest entityTypeRequest = PermissionTypeRequest
                .newBuilder()
                .setPermissionType(permissionType)
                .setClientId(clientId)
                .build();
        return blockingStub.createPermissionType(entityTypeRequest);
    }


    public Status updatePermissionType(String clientId, PermissionType permissionType) {
        PermissionTypeRequest entityTypeRequest = PermissionTypeRequest
                .newBuilder()
                .setPermissionType(permissionType)
                .setClientId(clientId)
                .build();
        return blockingStub.updatePermissionType(entityTypeRequest);
    }


    public Status deletePermissionType(String clientId, PermissionType permissionType) {
        PermissionTypeRequest entityTypeRequest = PermissionTypeRequest
                .newBuilder()
                .setPermissionType(permissionType)
                .setClientId(clientId)
                .build();
        return blockingStub.deletePermissionType(entityTypeRequest);
    }


    public PermissionType getPermissionType(String clientId, PermissionType permissionType) {
        PermissionTypeRequest entityTypeRequest = PermissionTypeRequest
                .newBuilder()
                .setPermissionType(permissionType)
                .setClientId(clientId)
                .build();
        return blockingStub.getPermissionType(entityTypeRequest);
    }


    public PermissionTypes getPermissionTypes(String clientId, SearchRequest request) {
        request = request.toBuilder().setClientId(clientId).build();

        return blockingStub.getPermissionTypes(request);
    }


    public Status createEntity(String clientId, Entity entity) {

        EntityRequest entityRequest = EntityRequest
                .newBuilder()
                .setClientId(clientId)
                .build();
        return blockingStub.createEntity(entityRequest);

    }


    public Status updateEntity(String clientId, Entity entity) {
        EntityRequest entityRequest = EntityRequest
                .newBuilder()
                .setClientId(clientId)
                .build();
        return blockingStub.updateEntity(entityRequest);

    }


    public Status isEntityExists(String clientId, Entity entity) {
        EntityRequest entityRequest = EntityRequest
                .newBuilder()
                .setClientId(clientId)
                .build();
        return blockingStub.isEntityExists(entityRequest);

    }


    public Entity getEntity(String clientId, Entity entity) {
        EntityRequest entityRequest = EntityRequest
                .newBuilder()
                .setClientId(clientId)
                .build();
        return blockingStub.getEntity(entityRequest);

    }


    public Status deleteEntity(String clientId, Entity entity) {

        EntityRequest entityRequest = EntityRequest
                .newBuilder()
                .setClientId(clientId)
                .build();
        return blockingStub.deleteEntity(entityRequest);
    }


    public Entities searchEntities(String clientId, SearchRequest request) {
        request = request.toBuilder().setClientId(clientId).build();
        return blockingStub.searchEntities(request);
    }


    public SharedOwners getListOfSharedUsers(String clientId, SharingRequest request) {

        request = request.toBuilder().setClientId(clientId).build();
        return blockingStub.getListOfSharedUsers(request);


    }


    public SharedOwners getListOfDirectlySharedUsers(String clientId, SharingRequest request) {
        request = request.toBuilder().setClientId(clientId).build();
        return blockingStub.getListOfDirectlySharedUsers(request);

    }


    public SharedOwners getListOfSharedGroups(String clientId, SharingRequest request) {

        request = request.toBuilder().setClientId(clientId).build();
        return blockingStub.getListOfSharedGroups(request);
    }


    public SharedOwners getListOfDirectlySharedGroups(String clientId, SharingRequest request) {
        request = request.toBuilder().setClientId(clientId).build();
        return blockingStub.getListOfDirectlySharedGroups(request);
    }


    public Status shareEntityWithUsers(String clientId, SharingRequest request) {
        request = request.toBuilder().setClientId(clientId).build();
        return blockingStub.shareEntityWithUsers(request);
    }


    public Status shareEntityWithGroups(String clientId, SharingRequest request) {
        request = request.toBuilder().setClientId(clientId).build();
        return blockingStub.shareEntityWithGroups(request);
    }


    public Status revokeEntitySharingFromUsers(String clientId, SharingRequest request) {
        request = request.toBuilder().setClientId(clientId).build();
        return blockingStub.revokeEntitySharingFromUsers(request);
    }


    public Status revokeEntitySharingFromGroups(String clientId, SharingRequest request) {
        request = request.toBuilder().setClientId(clientId).build();
        return blockingStub.revokeEntitySharingFromGroups(request);
    }


    public Status userHasAccess(String clientId, SharingRequest request) {
        request = request.toBuilder().setClientId(clientId).build();
        return blockingStub.userHasAccess(request);
    }


}
