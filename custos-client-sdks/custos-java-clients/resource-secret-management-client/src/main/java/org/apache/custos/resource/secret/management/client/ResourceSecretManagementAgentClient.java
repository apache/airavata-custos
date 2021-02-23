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
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.util.List;
import java.util.Map;

/**
 * Client for agents or service accounts to manage secrets of Custos
 */
public class ResourceSecretManagementAgentClient extends ResourceSecretManagementClient  {

    private static final Logger LOGGER = LoggerFactory.getLogger(ResourceSecretManagementAgentClient.class);


    ManagedChannel managedChannel = null;

    public ResourceSecretManagementAgentClient(String serviceHost, int servicePort,
                                               String clientId, String clientSecret) throws IOException {

        super(serviceHost, servicePort, clientId, clientSecret);
        managedChannel = getManagedChannel();
    }

    public SecretMetadata getSecret(String userToken, String agentToken,
                                    ResourceOwnerType ownerType, ResourceType resourceType) {

        return super.getSecret(ownerType, resourceType, attachedHeaders(userToken, agentToken));
    }


    public Struct getJWKS(String userToken, String agentToken) {
        attachedHeaders(userToken, agentToken);
        return super.getJWKS();
    }

    public SecretMetadata getResourceCredentialSummary(String userToken, String agentToken, String clientId, String token) {
        return super.getResourceCredentialSummary(clientId, token, attachedHeaders(userToken, agentToken));
    }


    public ResourceCredentialSummaries getAllResourceCredentialSummaries(String userToken, String agentToken,
                                                                         String clientId, List<String> accessibleTokens) {
        return super.getAllResourceCredentialSummaries(clientId, accessibleTokens, attachedHeaders(userToken, agentToken));
    }

    public AddResourceCredentialResponse generateSSHCredential(String userToken, String agentToken,
                                                               String clientId, String description, String ownerId) {
        return super.generateSSHCredential(clientId, description, ownerId, attachedHeaders(userToken, agentToken));
    }

    public AddResourceCredentialResponse addSSHCredential(String userToken, String agentToken, String csToken,
                                                          String passphrase, String privateKey, String publicKey,
                                                          String clientId, String description, String ownerId) {
        return super.addSSHCredential(csToken, passphrase, privateKey,
                publicKey, clientId, description, ownerId, attachedHeaders(userToken, agentToken));
    }

    public AddResourceCredentialResponse addPasswordCredential(String userToken, String agentToken,
                                                               String clientId, String description, String ownerId,  String userId,String password) {
        return super.addPasswordCredential(clientId,
                description, ownerId, userId, password,attachedHeaders(userToken, agentToken));
    }


    public AddResourceCredentialResponse addPasswordCredential(String userToken, String agentToken,
                                                               String token, String clientId, String description,
                                                               String ownerId,  String userId, String password) {
        return super.addPasswordCredential(token, clientId,
                description, ownerId, userId,password, attachedHeaders(userToken, agentToken));
    }

    public SSHCredential getSSHCredential(String userToken, String agentToken, String csToken, boolean useShamirSecret) {
        return super.getSSHCredential("",
                csToken, useShamirSecret, attachedHeaders(userToken, agentToken));
    }


    public PasswordCredential getPasswordCredential(String userToken, String agentToken,
                                                    String clientId, String token) {
        return super.getPasswordCredential(clientId, token, attachedHeaders(userToken, agentToken));
    }

    public ResourceCredentialOperationStatus deleteSSHCredential(String userToken, String agentToken,
                                                                 String clientId, String token) {
        return super.deleteSSHCredential(clientId, token, attachedHeaders(userToken, agentToken));
    }

    public ResourceCredentialOperationStatus deletePWDCredential(String userToken, String agentToken,
                                                                 String clientId, String token) {
        return super.deletePWDCredential(clientId, token, attachedHeaders(userToken, agentToken));
    }

    public AddResourceCredentialResponse addCredentialMap(String userToken, String agentToken,
                                                              String clientId, String description, String ownerId,
                                                              String token, Map<String, String> credentialMap) {
        return super.addCredentialMap(clientId, description, ownerId,
                token, credentialMap, attachedHeaders(userToken, agentToken));
    }

    public CredentialMap getCredentialMap(String userToken, String agentToken, String clientId, String token) {
        return super.getCredentialMap(clientId, token, attachedHeaders(userToken, agentToken));
    }

    public ResourceCredentialOperationStatus deleteCredentialMap(String userToken,
                                                                 String agentToken, String clientId, String token) {
        return super.deleteCredentialMap(clientId, token, attachedHeaders(userToken, agentToken));
    }


    public ResourceCredentialOperationStatus updateCredentialMap(String userToken, String agentToken,
                                                                 String clientId, String description, String ownerId,
                                                                 String token, Map<String, String> credentialMap) {
        return super.updateCredentialMap(clientId, description, ownerId, token, credentialMap,
                attachedHeaders(userToken, agentToken));
    }


    private ResourceSecretManagementServiceGrpc.ResourceSecretManagementServiceBlockingStub
    attachedHeaders(String userToken, String agentToken) {
        ResourceSecretManagementServiceGrpc.ResourceSecretManagementServiceBlockingStub
                blockingStub = ResourceSecretManagementServiceGrpc.newBlockingStub(managedChannel);

        Metadata authHeader = ClientUtils.getAuthorizationHeader(agentToken);
        Metadata tokenHeader = ClientUtils.getUserTokenHeader(userToken);
        Metadata agentEnablingHeaders = ClientUtils.getAgentEnablingHeader();
        blockingStub = MetadataUtils.attachHeaders(blockingStub, authHeader);
        blockingStub = MetadataUtils.attachHeaders(blockingStub, tokenHeader);
        blockingStub = MetadataUtils.attachHeaders(blockingStub, agentEnablingHeaders);
        return blockingStub;
    }

    @Override
    public void close() throws IOException {
        super.close();
        if (this.managedChannel != null) {
            this.managedChannel.shutdown();
        }
    }
}
