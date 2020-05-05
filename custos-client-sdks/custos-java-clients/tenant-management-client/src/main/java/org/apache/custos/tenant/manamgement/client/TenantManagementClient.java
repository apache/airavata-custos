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
import org.apache.custos.tenant.management.service.*;
import org.apache.custos.tenant.profile.service.Tenant;

import javax.net.ssl.SSLException;

/**
 * Java client to managed registered Custos Tenants
 */
public class TenantManagementClient {

    private ManagedChannel managedChannel;

    private TenantManagementServiceGrpc.TenantManagementServiceBlockingStub blockingStub;


    public TenantManagementClient(String serviceHost, int servicePort, String certificateFilePath, String adminClientId,
                                  String adminClientSecret) throws SSLException {

        if (serviceHost == null || certificateFilePath == null || adminClientId == null || adminClientSecret == null) {
            throw new NullPointerException("Please provide all the parameters");
        }

        managedChannel = NettyChannelBuilder.forAddress(serviceHost, servicePort)
                .sslContext(GrpcSslContexts
                        .forClient()
                        .trustManager(ClientUtils.getFile(TenantManagementClient.class, certificateFilePath)) // public key
                        .build())
                .build();


        blockingStub = TenantManagementServiceGrpc.newBlockingStub(managedChannel);
        blockingStub = MetadataUtils.attachHeaders(blockingStub,
                ClientUtils.getAuthorizationHeader(adminClientId, adminClientSecret));

    }

    /**
     * Only accessible by admin tenants and provides information related  to given Client Id.
     *
     * @param clientId
     * @return {
     * <p>
     * string client_id
     * string client_name
     * string requester_email
     * string admin_first_name
     * string admin_last_name
     * string admin_email
     * repeated string contacts
     * repeated string redirect_uris
     * repeated string grant_types
     * double client_id_issued_at
     * string client_uri
     * string scope
     * string domain
     * string comment
     * string logo_uri
     * string application_type
     * string jwks_uri ;
     * string example_extension_parameter
     * string tos_uri
     * string policy_uri
     * map<string, string> jwks
     * string software_id
     * string software_version
     * }
     */
    public GetTenantResponse getTenant(String clientId) {

        if (clientId == null) {
            throw new NullPointerException("Client Id is null");
        }

        GetTenantRequest tenantRequest = GetTenantRequest.newBuilder().setClientId(clientId).build();
        return blockingStub.getTenant(tenantRequest);

    }


    /**
     * Only accessible by admin tenants and updates and get information related to given Client Id
     *
     * @param tenant
     * @param clientId
     * @return {
     * <p>
     * string client_id
     * string client_name
     * string requester_email
     * string admin_first_name
     * string admin_last_name
     * string admin_email
     * repeated string contacts
     * repeated string redirect_uris
     * repeated string grant_types
     * double client_id_issued_at
     * string client_uri
     * string scope
     * string domain
     * string comment
     * string logo_uri
     * string application_type
     * string jwks_uri
     * string example_extension_parameter
     * string tos_uri
     * string policy_uri
     * map<string, string> jwks
     * string software_id
     * string software_version
     * }
     */
    public GetTenantResponse updateAndGetTenant(Tenant tenant, String clientId) {

        if (tenant == null || clientId == null) {
            throw new NullPointerException(" Tenant and Client Id should not be null");
        }

        UpdateTenantRequest updateTenantRequest = UpdateTenantRequest.newBuilder()
                .setClientId(clientId).setBody(tenant).build();

        return blockingStub.updateTenant(updateTenantRequest);

    }

    /**
     * Only accessible by admin tenants and deletes the  tenant  attached to given Client Id
     *
     * @param clientId
     */
    public void deleteTenant(String clientId) {
        if (clientId == null) {
            throw new NullPointerException("Client Id should not be null");
        }

        DeleteTenantRequest tenantRequest = DeleteTenantRequest.newBuilder().setClientId(clientId).build();
        blockingStub.deleteTenant(tenantRequest);

    }


}
