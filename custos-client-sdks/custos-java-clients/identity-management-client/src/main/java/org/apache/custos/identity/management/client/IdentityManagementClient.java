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

import com.google.protobuf.Struct;
import io.grpc.ManagedChannel;
import io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.NettyChannelBuilder;
import io.grpc.stub.MetadataUtils;
import org.apache.custos.clients.core.ClientUtils;
import org.apache.custos.credential.store.service.Credentials;
import org.apache.custos.identity.management.service.GetAgentTokenRequest;
import org.apache.custos.identity.management.service.GetCredentialsRequest;
import org.apache.custos.identity.management.service.IdentityManagementServiceGrpc;
import org.apache.custos.identity.service.*;

import java.io.Closeable;
import java.io.IOException;

/**
 * Java client to connect with the Custos Identity Management Service
 */
public class IdentityManagementClient implements Closeable {


    private ManagedChannel managedChannel;

    private IdentityManagementServiceGrpc.IdentityManagementServiceBlockingStub blockingStub;

    private IdentityManagementServiceGrpc.IdentityManagementServiceBlockingStub cleanBlockingStub;


    public IdentityManagementClient(String serviceHost, int servicePort, String clientId,
                                    String clientSecret) throws IOException {

        managedChannel = NettyChannelBuilder.forAddress(serviceHost, servicePort)
                .sslContext(GrpcSslContexts
                        .forClient()
                        .trustManager(ClientUtils.getServerCertificate(serviceHost, clientId, clientSecret)) // public key
                        .build())
                .build();

        blockingStub = IdentityManagementServiceGrpc.newBlockingStub(managedChannel);
        cleanBlockingStub = IdentityManagementServiceGrpc.newBlockingStub(managedChannel);
        blockingStub = MetadataUtils.attachHeaders(blockingStub, ClientUtils.getAuthorizationHeader(clientId, clientSecret));
    }


    /**
     * Get token
     *
     * @param redirectUri
     * @param code
     * @param username
     * @param password
     * @param refreshToken
     * @param grantType
     * @return
     */
    public Struct getToken(String redirectUri, String code, String username, String password, String refreshToken,
                           String grantType) {

        GetTokenRequest.Builder request = GetTokenRequest.newBuilder();

        if (redirectUri != null) {
            request = request.setRedirectUri(redirectUri);
        }

        if (code != null) {
            request = request.setCode(code);
        }
        if (username != null) {
            request = request.setUsername(username);
        }
        if (password != null) {
            request = request.setPassword(password);
        }
        if (refreshToken != null) {
            request = request.setRefreshToken(refreshToken);
        }
        if (grantType != null) {
            request = request.setGrantType(grantType);
        }

        return blockingStub.token(request.build());

    }


    /**
     * Get iam, cilogon credentials of given tenant
     *
     * @param clientId
     * @return
     */
    public Credentials getCredentials(String clientId) {
        GetCredentialsRequest request = GetCredentialsRequest.newBuilder().setClientId(clientId).build();
        return blockingStub.getCredentials(request);
    }


    /**
     * Get OIDC configurations of given client
     *
     * @param clientId
     * @return
     */
    public Struct getOIDCConfiguration(String clientId) {

        GetOIDCConfiguration configuration = GetOIDCConfiguration
                .newBuilder()
                .setClientId(clientId).build();
        return blockingStub.getOIDCConfiguration(configuration);

    }


    /**
     * End user session
     *
     * @param refreshToken
     * @return
     */
    public OperationStatus endUserSession(String refreshToken) {
        EndSessionRequest body = EndSessionRequest
                .newBuilder()
                .setRefreshToken(refreshToken)
                .build();
        org.apache.custos.identity.management.service.EndSessionRequest endSessionRequest =
                org.apache.custos.identity.management.service.EndSessionRequest
                        .newBuilder()
                        .setBody(body)
                        .build();
        return blockingStub.endUserSession(endSessionRequest);
    }


    /**
     * Get agent tokens
     *
     * @param clientId
     * @param grantType
     * @param refreshToken
     * @return
     */
    public Struct getAgentToken(String clientId, String agentId, String agentSec, String grantType, String refreshToken) {

        GetAgentTokenRequest.Builder agentTokenRequest = GetAgentTokenRequest
                .newBuilder()
                .setClientId(clientId)
                .setGrantType(grantType);

        if (refreshToken != null) {
            agentTokenRequest = agentTokenRequest.setRefreshToken(refreshToken);
        }

        cleanBlockingStub = MetadataUtils.attachHeaders(cleanBlockingStub, ClientUtils.getAuthorizationHeader(agentId, agentSec));
        return cleanBlockingStub.getAgentToken(agentTokenRequest.build());

    }

    public User getUser(String clientId, String username, String accessToken) {

        Claim user = Claim.newBuilder().setKey("username").setValue(username).build();
        Claim cli = Claim.newBuilder().setKey("client_id").setValue(clientId).build();

        AuthToken authToken = AuthToken.newBuilder()
                .setAccessToken(accessToken)
                .addClaims(user)
                .addClaims(cli)
                .build();
        return blockingStub.getUser(authToken);

    }

    public boolean isAuthenticated(String accessToken) {
        try {
            AuthToken authToken = AuthToken
                    .newBuilder()
                    .setAccessToken(accessToken)
                    .build();
            IsAuthenticatedResponse authenticatedResponse = blockingStub.isAuthenticated(authToken);
            return authenticatedResponse.getAuthenticated();
        } catch (Exception ex) {
            return false;
        }

    }

    @Override
    public void close() throws IOException {
        this.managedChannel.shutdown();
    }
}
