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

package org.apache.custos.tenant.profile.client.async;

import io.grpc.ClientInterceptor;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import org.apache.custos.integration.core.ServiceCallback;
import org.apache.custos.integration.core.ServiceException;
import org.apache.custos.tenant.profile.service.*;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.List;

/**
 * This class uses gRPC stubs generated for {@link TenantProfileServiceGrpc}
 * and acts as the client
 */
@Component
public class TenantProfileClient {

    private ManagedChannel managedChannel;
    private TenantProfileServiceGrpc.TenantProfileServiceStub profileServiceStub;
    private TenantProfileServiceGrpc.TenantProfileServiceBlockingStub profileServiceBlockingStub;


    private final List<ClientInterceptor> clientInterceptorList;


    public TenantProfileClient(List<ClientInterceptor> clientInterceptorList,
                               @Value("${tenant.profile.core.service.dns.name}") String serviceHost,
                               @Value("${tenant.profile.core.service.port}") int servicePort) {
        this.clientInterceptorList = clientInterceptorList;
        managedChannel = ManagedChannelBuilder.forAddress(
                serviceHost, servicePort).usePlaintext(true).intercept(clientInterceptorList).build();
        profileServiceStub = TenantProfileServiceGrpc.newStub(managedChannel);
        profileServiceBlockingStub = TenantProfileServiceGrpc.newBlockingStub(managedChannel);
    }

    public void addTenantAsync(Tenant tenant, final ServiceCallback callback) {

        StreamObserver observer = this.getObserver(callback, "Add tenant task failed");

        profileServiceStub.addTenant(tenant, observer);
    }


    public AddTenantResponse addTenant(Tenant tenant) {
        return profileServiceBlockingStub.addTenant(tenant);

    }


    public void updateTenantAsync(UpdateTenantRequest request, final ServiceCallback callback) {

        StreamObserver observer = this.getObserver(callback, "Update tenant task failed");

        profileServiceStub.updateTenant(request, observer);
    }


    public UpdateTenantResponse updateTenant(UpdateTenantRequest updateTenantRequest) {
        return profileServiceBlockingStub.updateTenant(updateTenantRequest);


    }


    public void getAllTenantsAsync(final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "Get all tenants task failed");

        Empty empty = Empty.newBuilder().build();
        profileServiceStub.getAllTenants(empty, observer);
    }

    public GetAllTenantsResponse getAllTenants() {
        Empty empty = Empty.newBuilder().build();
        return profileServiceBlockingStub.getAllTenants(empty);
    }


    public void getAllTenantsForUserAsync(GetAllTenantsForUserRequest getAllTenantsForUserRequest, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "Get all tenants for user task failed");

        profileServiceStub.getAllTenantsForUser(getAllTenantsForUserRequest, observer);
    }

    public GetAllTenantsForUserResponse getAllTenantsForUser(GetAllTenantsForUserRequest getAllTenantsForUserRequest) {

        return profileServiceBlockingStub.getAllTenantsForUser(getAllTenantsForUserRequest);
    }


    public void getTenantAsync(GetTenantRequest getTenantRequest, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "Get tenant task failed");

        profileServiceStub.getTenant(getTenantRequest, observer);
    }

    public GetTenantResponse getTenant(GetTenantRequest getTenantRequest) {
        return profileServiceBlockingStub.getTenant(getTenantRequest);
    }

    public void updateTenantStatusAsync(UpdateStatusRequest updateTenantRequest, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "Update tenant status task failed");

        profileServiceStub.updateTenantStatus(updateTenantRequest, observer);
    }

    public UpdateStatusResponse updateTenantStatus(UpdateStatusRequest request) {
        return profileServiceBlockingStub.updateTenantStatus(request);
    }

    public void getAttributeUpdateAuditTrailAsync(
            GetAuditTrailRequest request, final ServiceCallback callback) {

        StreamObserver observer = this.getObserver(callback, "Get attribute update audit trail failed");

        profileServiceStub.getTenantAttributeUpdateAuditTrail(request, observer);
    }

    public GetAttributeUpdateAuditTrailResponse getAttributeUpdateAuditTrail(
            GetAuditTrailRequest request) {

        return profileServiceBlockingStub.getTenantAttributeUpdateAuditTrail(request);
    }

    public void getStatusUpdateAuditTrailAsync(GetAuditTrailRequest request,
                                               final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "Get  status update audit trail task failed");

        profileServiceStub.getTenantStatusUpdateAuditTrail(request, observer);
    }

    public GetStatusUpdateAuditTrailResponse getStatusUpdateAuditTrail(
            GetAuditTrailRequest request) {

        return profileServiceBlockingStub.getTenantStatusUpdateAuditTrail(request);
    }

    public void isTenantExistAsync(IsTenantExistRequest request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "Is tenant exist task failed");

        profileServiceStub.isTenantExist(request, observer);
    }


    public IsTenantExistResponse isTenantExist(IsTenantExistRequest request) {
        return profileServiceBlockingStub.isTenantExist(request);
    }


    private StreamObserver getObserver(ServiceCallback callback, String failureMsg) {
        final Object[] response = new Object[1];
        StreamObserver observer = new StreamObserver() {
            @Override
            public void onNext(Object o) {
                response[0] = o;
            }

            @Override
            public void onError(Throwable throwable) {
                callback.onError(new ServiceException(failureMsg, throwable, null));
            }

            @Override
            public void onCompleted() {
                callback.onCompleted(response[0]);
            }
        };

        return observer;
    }


}
