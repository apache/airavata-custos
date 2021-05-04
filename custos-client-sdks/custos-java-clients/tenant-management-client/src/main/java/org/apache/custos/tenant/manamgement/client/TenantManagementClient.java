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

package org.apache.custos.tenant.manamgement.client;

import io.grpc.ManagedChannel;
import io.grpc.Metadata;
import io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.NettyChannelBuilder;
import io.grpc.stub.MetadataUtils;
import org.apache.custos.clients.core.ClientUtils;
import org.apache.custos.iam.service.*;
import org.apache.custos.tenant.management.service.DeleteTenantRequest;
import org.apache.custos.tenant.management.service.GetTenantRequest;
import org.apache.custos.tenant.management.service.GetTenantResponse;
import org.apache.custos.tenant.management.service.*;
import org.apache.custos.tenant.profile.service.*;

import java.io.Closeable;
import java.io.IOException;
import java.util.Arrays;

/**
 * This class contains tenant management operations
 */
public class TenantManagementClient implements Closeable {

    private ManagedChannel managedChannel;

    private TenantManagementServiceGrpc.TenantManagementServiceBlockingStub blockingStub;


    public TenantManagementClient(String serviceHost, int servicePort, String clientId,
                                  String clientSecret) throws IOException {

        managedChannel = NettyChannelBuilder.forAddress(serviceHost, servicePort)
                .sslContext(GrpcSslContexts
                        .forClient()
                        .trustManager(ClientUtils.getServerCertificate(serviceHost, clientId, clientSecret)) // public key
                        .build())
                .build();

        blockingStub = TenantManagementServiceGrpc.newBlockingStub(managedChannel);

        blockingStub = MetadataUtils.attachHeaders(blockingStub,
                ClientUtils.getAuthorizationHeader(clientId, clientSecret));

    }


    /**
     * Register child tenant
     * @param client_name
     * @param requester_email
     * @param admin_frist_name
     * @param admin_last_name
     * @param admin_email
     * @param admin_username
     * @param admin_password
     * @param contacts
     * @param redirect_uris
     * @param client_uri
     * @param scope
     * @param domain
     * @param logo_uri
     * @param comment
     * @return
     */
    public CreateTenantResponse registerTenant(String client_name, String requester_email, String admin_frist_name,
                                               String admin_last_name, String admin_email, String admin_username,
                                               String admin_password, String[] contacts, String[] redirect_uris,
                                               String client_uri, String scope, String domain, String logo_uri,
                                               String comment) {
        Tenant tenant = Tenant.newBuilder()
                .setClientName(client_name)
                .setRequesterEmail(requester_email)
                .setAdminFirstName(admin_frist_name)
                .setAdminLastName(admin_last_name)
                .setAdminEmail(admin_email)
                .setAdminUsername(admin_username)
                .setAdminPassword(admin_password)
                .addAllContacts(Arrays.asList(contacts))
                .addAllRedirectUris(Arrays.asList(redirect_uris))
                .setClientUri(client_uri)
                .setScope(scope)
                .setDomain(domain)
                .setLogoUri(logo_uri)
                .setComment(comment)
                .setApplicationType("web")
                .build();

        return blockingStub.createTenant(tenant);

    }


    /**
     * Update tenant
     * @param clientId
     * @param client_name
     * @param requester_email
     * @param admin_frist_name
     * @param admin_last_name
     * @param admin_email
     * @param admin_username
     * @param admin_password
     * @param contacts
     * @param redirect_uris
     * @param client_uri
     * @param scope
     * @param domain
     * @param logo_uri
     * @param comment
     * @return
     */
    public Tenant updateTenant(String usertoken, String clientId, String client_name, String requester_email, String admin_frist_name,
                                          String admin_last_name, String admin_email, String admin_username,
                                          String admin_password, String[] contacts, String[] redirect_uris,
                                          String client_uri, String scope, String domain, String logo_uri,
                                          String comment) {
        Tenant tenant = Tenant.newBuilder()
                .setClientName(client_name)
                .setRequesterEmail(requester_email)
                .setAdminFirstName(admin_frist_name)
                .setAdminLastName(admin_last_name)
                .setAdminEmail(admin_email)
                .setAdminUsername(admin_username)
                .setAdminPassword(admin_password)
                .addAllContacts(Arrays.asList(contacts))
                .addAllRedirectUris(Arrays.asList(redirect_uris))
                .setClientUri(client_uri)
                .setScope(scope)
                .setDomain(domain)
                .setLogoUri(logo_uri)
                .setComment(comment)
                .setApplicationType("web")
                .build();

        UpdateTenantRequest updateTenantRequest = UpdateTenantRequest.newBuilder()
                .setBody(tenant)
                .setClientId(clientId)
                .build();

        return attachedHeaders(usertoken).updateTenant(updateTenantRequest);

    }


    public Tenant getTenant(String userToken, String clientId) {
        GetTenantRequest tenantRequest = GetTenantRequest
                .newBuilder()
                .setClientId(clientId)
                .build();
        return attachedHeaders(userToken).getTenant(tenantRequest);
    }


    /**
     * delete tenant identified by clientId
     * @param clientId
     */
    public void deleteTenant(String userToken, String clientId) {
        DeleteTenantRequest tenantRequest = DeleteTenantRequest.newBuilder().setClientId(clientId).build();
        attachedHeaders(userToken).deleteTenant(tenantRequest);
    }


    /**
     * Add tenant roles to tenant
     * @param roleRepresentations
     * @param clientLevel
     * @return
     */
    public AllRoles addTenantRoles(RoleRepresentation[] roleRepresentations, boolean clientLevel) {

        AddRolesRequest rolesRequest = AddRolesRequest
                .newBuilder()
                .addAllRoles(Arrays.asList(roleRepresentations))
                .setClientLevel(clientLevel)
                .build();
        return blockingStub.addTenantRoles(rolesRequest);

    }

    public OperationStatus addProtocolMapper(String name, String attributeName, String claimName, String claimType, String mapperType,
                                             boolean addToIdToken, boolean addToAccessToken, boolean addToUserInfo, boolean multiValued,
                                             boolean aggregreteMultiValues) {

        AddProtocolMapperRequest mapperRequest = AddProtocolMapperRequest.newBuilder()
                .setName(name)
                .setAttributeName(attributeName)
                .setClaimName(claimName)
                .setMapperType(MapperTypes.valueOf(mapperType))
                .setClaimType(ClaimJSONTypes.valueOf(claimType))
                .setAddToAccessToken(addToAccessToken)
                .setAddToIdToken(addToIdToken)
                .setAddToUserInfo(addToUserInfo)
                .setMultiValued(multiValued)
                .setAggregateAttributeValues(aggregreteMultiValues)
                .build();
        return blockingStub.addProtocolMapper(mapperRequest);
    }


    /**
     * Get child tenants
     * @param limit
     * @param offset
     * @param status
     * @return
     */
    public GetAllTenantsResponse getChildTenants(int limit, int offset, String status) {

        GetTenantsRequest request = GetTenantsRequest
                .newBuilder()
                .setLimit(limit)
                .setOffset(offset)
                .setStatus(TenantStatus.valueOf(status))
                .build();

        return blockingStub.getChildTenants(request);
    }


    /**
     * provides all tenants requested by given email
     * @param email
     * @return
     */
    public GetAllTenantsForUserResponse getAllTenants(String email) {

        GetAllTenantsForUserRequest request = GetAllTenantsForUserRequest
                .newBuilder()
                .setRequesterEmail(email)
                .build();

        return blockingStub.getAllTenantsForUser(request);
    }

    private TenantManagementServiceGrpc.TenantManagementServiceBlockingStub
    attachedHeaders(String userToken) {
        TenantManagementServiceGrpc.TenantManagementServiceBlockingStub
                blockingStub = TenantManagementServiceGrpc.newBlockingStub(managedChannel);

        Metadata tokenHeader = ClientUtils.getAuthorizationHeader(userToken);
        blockingStub = MetadataUtils.attachHeaders(blockingStub, tokenHeader);
        return blockingStub;
    }


    @Override
    public void close() throws IOException {
        if (this.managedChannel != null) {
            this.managedChannel.shutdown();
        }
    }
}
