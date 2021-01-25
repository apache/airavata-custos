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

package org.apache.custos.resource.secret.management.client;

import com.google.protobuf.Struct;
import io.grpc.ManagedChannel;
import io.grpc.Metadata;
import io.grpc.stub.MetadataUtils;
import org.apache.custos.clients.core.ClientUtils;
import org.apache.custos.resource.secret.management.service.ResourceSecretManagementServiceGrpc;
import org.apache.custos.resource.secret.service.*;

import java.io.IOException;
import java.util.List;

/**
 * Client for agents or service accounts to manage secrets of Custos
 */
public class ResourceSecretManagementAgentClient extends ResourceSecretManagementClient {

    private ResourceSecretManagementServiceGrpc.ResourceSecretManagementServiceBlockingStub blockingStub;

    public ResourceSecretManagementAgentClient(String serviceHost, int servicePort,
                                               String clientId, String clientSecret) throws IOException {

        super(serviceHost, servicePort, clientId, clientSecret);
        ManagedChannel managedChannel = getManagedChannel();
        blockingStub = ResourceSecretManagementServiceGrpc.newBlockingStub(managedChannel);
        super.setBlockingStub(blockingStub);
    }

    public SecretMetadata getSecret(String userToken, String agentToken,
                                    ResourceOwnerType ownerType, ResourceType resourceType) {
        attachedHeaders(userToken, agentToken);
        return super.getSecret(ownerType, resourceType);
    }


    public Struct getJWKS(String userToken, String agentToken) {
        attachedHeaders(userToken, agentToken);
        return super.getJWKS();
    }

    public SecretMetadata getResourceCredentialSummary(String userToken, String agentToken, String clientId, String token) {
        attachedHeaders(userToken, agentToken);
        return super.getResourceCredentialSummary(clientId, token);
    }


    public ResourceCredentialSummaries getAllResourceCredentialSummaries(String userToken, String agentToken,
                                                                         String clientId, List<String> accessibleTokens) {
        attachedHeaders(userToken, agentToken);
        return super.getAllResourceCredentialSummaries(clientId, accessibleTokens);
    }

    public AddResourceCredentialResponse generateSSHCredential(String userToken, String agentToken,
                                                               String clientId, String description, String ownerId) {
        attachedHeaders(userToken, agentToken);
        return super.generateSSHCredential(clientId, description, ownerId);
    }

    public AddResourceCredentialResponse addSSHCredential(String userToken, String agentToken, String csToken,
                                                          String passphrase, String privateKey, String publicKey,
                                                          String clientId, String description, String ownerId) {
        attachedHeaders(userToken, agentToken);
        return super.addSSHCredential(csToken, passphrase, privateKey, publicKey, clientId, description, ownerId);
    }

    public AddResourceCredentialResponse addPasswordCredential(String userToken, String agentToken,
                                                               String clientId, String description, String ownerId, String password) {
        attachedHeaders(userToken, agentToken);
        return super.addPasswordCredential(clientId, description, ownerId, password);
    }


    public AddResourceCredentialResponse addPasswordCredential(String userToken, String agentToken,
                                                               String token, String clientId, String description,
                                                               String ownerId, String password) {
        attachedHeaders(userToken, agentToken);
        return super.addPasswordCredential(token, clientId, description, ownerId, password);
    }

    public SSHCredential getSSHCredential(String userToken, String agentToken,
                                          String clientId, String token, boolean useShamirSecret) {
        attachedHeaders(userToken, agentToken);
        return super.getSSHCredential(clientId, token, useShamirSecret);
    }


    public PasswordCredential getPasswordCredential(String userToken, String agentToken, String clientId, String token) {
        attachedHeaders(userToken, agentToken);
        return super.getPasswordCredential(clientId, token);
    }

    public ResourceCredentialOperationStatus deleteSSHCredential(String userToken, String agentToken,
                                                                 String clientId, String token) {
        attachedHeaders(userToken, agentToken);
        return super.deleteSSHCredential(clientId, token);
    }

    public ResourceCredentialOperationStatus deletePWDCredential(String userToken, String agentToken,
                                                                 String clientId, String token) {
        attachedHeaders(userToken, agentToken);
        return super.deletePWDCredential(clientId, token);
    }

    private void attachedHeaders(String userToken, String agentToken) {
        Metadata authHeader = ClientUtils.getAuthorizationHeader(agentToken);
        Metadata tokenHeader = ClientUtils.getUserTokenHeader(userToken);
        MetadataUtils.attachHeaders(blockingStub, authHeader);
        MetadataUtils.attachHeaders(blockingStub, tokenHeader);
    }
}
