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

package org.apache.custos.identity.client;

import com.google.protobuf.Struct;
import io.grpc.ClientInterceptor;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import org.apache.custos.identity.service.*;
import org.apache.custos.integration.core.ServiceCallback;
import org.apache.custos.integration.core.ServiceException;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.List;

/**
 * The client class used to connect IAM admin services
 */
@Component
public class IdentityClient {
    private ManagedChannel managedChannel;
    private IdentityServiceGrpc.IdentityServiceStub identityServiceStub;
    private IdentityServiceGrpc.IdentityServiceBlockingStub identityServiceBlockingStub;

    private final List<ClientInterceptor> clientInterceptorList;


    public IdentityClient(List<ClientInterceptor> clientInterceptorList,
                          @Value("${identity.service.dns.name}") String serviceHost,
                          @Value("${identity.service.port}") int servicePort) {
        this.clientInterceptorList = clientInterceptorList;
        managedChannel = ManagedChannelBuilder.forAddress(
                serviceHost, servicePort).usePlaintext(true).intercept(clientInterceptorList).build();
        identityServiceStub = IdentityServiceGrpc.newStub(managedChannel);
        identityServiceBlockingStub = IdentityServiceGrpc.newBlockingStub(managedChannel);
    }


    public void authenticateAsync(AuthenticationRequest request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "Authenticate task failed");
        identityServiceStub.authenticate(request, observer);
    }


    public void isAuthenticatedAsync(AuthToken request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "isAuthenticated task failed");
        identityServiceStub.isAuthenticated(request, observer);
    }


    public void getUserAsync(AuthToken request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "getUser task failed");
        identityServiceStub.getUser(request, observer);
    }


    public void getUserManagementSATokenRequestAsync(GetUserManagementSATokenRequest request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "GetUserManagementSAToken task failed");
        identityServiceStub.getUserManagementServiceAccountAccessToken(request, observer);
    }

    public AuthToken authenticate(AuthenticationRequest request) {
        return identityServiceBlockingStub.authenticate(request);
    }


    public IsAuthenticatedResponse isAuthenticated(AuthToken request) {
        return identityServiceBlockingStub.isAuthenticated(request);
    }


    public User getUser(AuthToken request) {
        return identityServiceBlockingStub.getUser(request);
    }


    public AuthToken getUserManagementSATokenRequest(GetUserManagementSATokenRequest request) {
        return identityServiceBlockingStub.getUserManagementServiceAccountAccessToken(request);
    }

    public Struct getAccessToken(GetTokenRequest request) {
        return identityServiceBlockingStub.getToken(request);
    }

    public AuthorizationResponse getAuthorizationEndpoint(GetAuthorizationEndpointRequest request) {
        return identityServiceBlockingStub.getAuthorizeEndpoint(request);
    }

    public Struct getOIDCConfiguration(GetOIDCConfiguration request) {
        return identityServiceBlockingStub.getOIDCConfiguration(request);
    }

    public Struct getTokenByPasswordGrantType(GetTokenRequest request) {
        return identityServiceBlockingStub.getTokenByPasswordGrantType(request);
    }


    public Struct getTokenByRefreshTokenGrantType(GetTokenRequest request) {
       return identityServiceBlockingStub.getTokenByRefreshTokenGrantType(request);
    }

    public OperationStatus endSession(EndSessionRequest request) {
        return identityServiceBlockingStub.endSession(request);
    }
    public Struct getJWKS(GetJWKSRequest request) {
        return identityServiceBlockingStub.getJWKS(request);
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
