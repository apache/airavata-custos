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

package org.apache.custos.identity.management.client;

import io.grpc.ManagedChannel;
import io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.NettyChannelBuilder;
import io.grpc.stub.MetadataUtils;
import org.apache.custos.clients.core.ClientUtils;
import org.apache.custos.identity.management.service.AuthorizationRequest;
import org.apache.custos.identity.management.service.AuthorizationResponse;
import org.apache.custos.identity.management.service.IdentityManagementServiceGrpc;
import org.apache.custos.identity.service.GetTokenRequest;
import org.apache.custos.identity.service.TokenResponse;

import javax.net.ssl.SSLException;

/**
 * Java client to connect with the Custos Identity Management Service
 */
public class IdentityManagementClient {


    private ManagedChannel managedChannel;

    private IdentityManagementServiceGrpc.IdentityManagementServiceBlockingStub blockingStub;


    private String clientId;

    private String clientSecret;


    public IdentityManagementClient(String serviceHost, int servicePort, String clientId,
                                    String clientSecret, String certificateFilePath) throws SSLException {

        if (serviceHost == null || certificateFilePath == null || clientId == null || clientSecret == null) {
            throw new NullPointerException("Please provide all the parameters");
        }


        managedChannel = NettyChannelBuilder.forAddress(serviceHost, servicePort)
                .sslContext(GrpcSslContexts
                        .forClient()
                        .trustManager(ClientUtils.getFile(IdentityManagementClient.class, certificateFilePath)) // public key
                        .build())
                .build();

        this.clientId = clientId;
        this.clientSecret = clientSecret;

        blockingStub = IdentityManagementServiceGrpc.newBlockingStub(managedChannel);
        blockingStub = MetadataUtils.attachHeaders(blockingStub, ClientUtils.getAuthorizationHeader(clientId, clientSecret));
    }


    /**
     * Authorization Endpoint
     *
     * @param redirectUri
     * @param responseType
     * @param scope        return AuthorizationResponse
     */
    public AuthorizationResponse authorize(String redirectUri, String responseType, String scope, String state) {

        AuthorizationRequest request = AuthorizationRequest
                .newBuilder()
                .setRedirectUri(redirectUri)
                .setResponseType(responseType)
                .setState(state)
                .setScope(scope).build();
        return blockingStub.authorize(request);
    }

    /**
     * Token Endpoint
     *
     * @param code
     * @param redirectUri
     * @return TokenResponse
     */
    public TokenResponse getToken(String code, String redirectUri) {

        GetTokenRequest request = GetTokenRequest
                .newBuilder()
                .setCode(code)
                .setRedirectUri(redirectUri)
                .build();

        return null;
    }


}
