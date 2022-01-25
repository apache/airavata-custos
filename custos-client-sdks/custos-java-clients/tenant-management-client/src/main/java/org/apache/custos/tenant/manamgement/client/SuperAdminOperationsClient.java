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
import io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.NettyChannelBuilder;
import io.grpc.stub.MetadataUtils;
import org.apache.custos.clients.core.ClientUtils;
import org.apache.custos.tenant.management.service.CreateTenantResponse;
import org.apache.custos.tenant.management.service.TenantManagementServiceGrpc;
import org.apache.custos.tenant.profile.service.*;

import java.io.IOException;
import java.util.Arrays;

/**
 * The class containes operations permitted only by super admin tenant
 */
public class SuperAdminOperationsClient {

    private ManagedChannel managedChannel;

    private TenantManagementServiceGrpc.TenantManagementServiceBlockingStub blockingStubWithoutHeader;
    private TenantManagementServiceGrpc.TenantManagementServiceBlockingStub blockingStub;
    private TenantManagementServiceGrpc.TenantManagementServiceBlockingStub unAuthorizedStub;

    public SuperAdminOperationsClient(String serviceHost, int servicePort, String clientId,
                                      String clientSecret) throws IOException {

        managedChannel = NettyChannelBuilder.forAddress(serviceHost, servicePort)
                .sslContext(GrpcSslContexts
                        .forClient()
                        .trustManager(ClientUtils.getServerCertificate(serviceHost, clientId, clientSecret)) // public key
                        .build())
                .build();

        blockingStubWithoutHeader = TenantManagementServiceGrpc.newBlockingStub(managedChannel);
        blockingStub = TenantManagementServiceGrpc.newBlockingStub(managedChannel);

        blockingStub = MetadataUtils.attachHeaders(blockingStub,
                ClientUtils.getAuthorizationHeader(clientId, clientSecret));
    }


    /**
     * Register admin tenant
     *
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
    public CreateTenantResponse registerAdminTenant(String client_name, String requester_email, String admin_frist_name,
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

        return blockingStubWithoutHeader.createTenant(tenant);

    }


    /**
     * Get all tenants mapping to given status and requested by user
     *
     * @param offset
     * @param limit
     * @param status
     * @param requesterEmail
     * @return
     */
    public GetAllTenantsResponse getAllTenants(int offset, int limit, String status, String requesterEmail) {

        GetTenantsRequest request = GetTenantsRequest
                .newBuilder()
                .setStatus(TenantStatus.valueOf(status))
                .setLimit(limit)
                .setOffset(offset)
                .setRequesterEmail(requesterEmail)
                .build();
        return blockingStub.getAllTenants(request);

    }


    /**
     * Update tenant status
     *
     * @param adminUserToken
     * @param clientId
     * @param status
     * @return
     */
    public UpdateStatusResponse updateTenantStatus(String adminUserToken, String clientId, String status) {

        UpdateStatusRequest request = UpdateStatusRequest
                .newBuilder()
                .setClientId(clientId)
                .setStatus(TenantStatus.valueOf(status))
                .build();

        TenantManagementServiceGrpc.TenantManagementServiceBlockingStub blockingStub =
                MetadataUtils.attachHeaders(this.blockingStub, ClientUtils.getAuthorizationHeader(adminUserToken));
        return blockingStub.updateTenantStatus(request);


    }




}
