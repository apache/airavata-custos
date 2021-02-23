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

import com.google.protobuf.ByteString;
import com.google.protobuf.Struct;
import io.grpc.ManagedChannel;
import io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.NettyChannelBuilder;
import io.grpc.stub.MetadataUtils;
import org.apache.custos.clients.core.ClientUtils;
import org.apache.custos.identity.service.GetJWKSRequest;
import org.apache.custos.integration.core.utils.ShamirSecretHandler;
import org.apache.custos.resource.secret.management.service.ResourceSecretManagementServiceGrpc;
import org.apache.custos.resource.secret.service.*;

import java.io.Closeable;
import java.io.IOException;
import java.util.List;
import java.util.Map;

/**
 * This class is responsible for managing resource secrets
 */
public class ResourceSecretManagementClient implements Closeable {

    private ManagedChannel managedChannel;

    private ResourceSecretManagementServiceGrpc.ResourceSecretManagementServiceBlockingStub blockingStub;

    private int defaultNumOfShares = 5;
    private int defaultThreshold = 3;


    public ResourceSecretManagementClient(String serviceHost, int servicePort, String clientId,
                                          String clientSecret) throws IOException {

        managedChannel = NettyChannelBuilder.forAddress(serviceHost, servicePort)
                .sslContext(GrpcSslContexts
                        .forClient()
                        .trustManager(ClientUtils.getServerCertificate(serviceHost, clientId, clientSecret)) // public key
                        .build())
                .build();

        blockingStub = ResourceSecretManagementServiceGrpc.newBlockingStub(managedChannel);

        blockingStub = MetadataUtils.attachHeaders(blockingStub,
                ClientUtils.getAuthorizationHeader(clientId, clientSecret));

    }


    /**
     * provides resource secret for given owner type and resource type
     *
     * @param ownerType
     * @param resourceType
     * @return
     */
    public SecretMetadata getSecret(ResourceOwnerType ownerType, ResourceType resourceType) {

        SecretMetadata metadata = SecretMetadata
                .newBuilder()
                .setOwnerType(ownerType)
                .setResourceType(resourceType)
                .build();

        GetSecretRequest request = GetSecretRequest.newBuilder().setMetadata(metadata).build();
        return blockingStub.getSecret(request);
    }

    /**
     * provides JWKS keys for calling tenant
     *
     * @return
     */
    public Struct getJWKS(ResourceSecretManagementServiceGrpc.
                                  ResourceSecretManagementServiceBlockingStub blockingStub) {
        GetJWKSRequest request = GetJWKSRequest.newBuilder().build();
        return blockingStub.getJWKS(request);
    }


    /**
     * Provides metadata object of credentials
     *
     * @param clientId
     * @param token
     * @return SecretMetadata
     */
    public SecretMetadata getResourceCredentialSummary(String clientId, String token) {

        GetResourceCredentialByTokenRequest tokenRequest = GetResourceCredentialByTokenRequest
                .newBuilder()
                .setClientId(clientId)
                .setToken(token)
                .build();

        return blockingStub.getResourceCredentialSummary(tokenRequest);

    }

    /**
     * Provides metadata array of credentials metadata
     *
     * @param clientId
     * @param accessibleTokens
     * @return SecretMetadata[]
     */
    public ResourceCredentialSummaries getAllResourceCredentialSummaries(String clientId, List<String> accessibleTokens) {
        GetResourceCredentialSummariesRequest summariesRequest = GetResourceCredentialSummariesRequest
                .newBuilder()
                .setClientId(clientId)
                .addAllAccessibleTokens(accessibleTokens).build();


        return blockingStub.getAllResourceCredentialSummaries(summariesRequest);

    }

    /**
     * Generate SSH credentials
     *
     * @param clientId
     * @param description
     * @param ownerId
     * @return AddResourceCredentialResponse
     */
    public AddResourceCredentialResponse generateSSHCredential(String clientId, String description, String ownerId) {

        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setClientId(clientId)
                .setDescription(description)
                .setOwnerId(ownerId).build();


        SSHCredential sshCredential = SSHCredential
                .newBuilder()
                .setMetadata(metadata).build();

        return blockingStub.addSSHCredential(sshCredential);

    }

    public AddResourceCredentialResponse addSSHCredential(String token, String passphrase, String privateKey,
                                                          String publicKey, String clientId, String description, String ownerId) {
        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setClientId(clientId)
                .setDescription(description)
                .setOwnerId(ownerId)
                .setToken(token).
                        build();

        SSHCredential sshCredential = SSHCredential
                .newBuilder()
                .setMetadata(metadata)
                .setPassphrase(passphrase)
                .setPrivateKey(privateKey)
                .setPublicKey(publicKey).build();

        return blockingStub.addSSHCredential(sshCredential);

    }


    /**
     * Save password credentials
     *
     * @param clientId
     * @param description
     * @param ownerId
     * @param password
     * @return AddResourceCredentialResponse
     */
    public AddResourceCredentialResponse addPasswordCredential(String clientId, String description, String ownerId, String userId, String password) {
        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setClientId(clientId)
                .setDescription(description)
                .setOwnerId(ownerId).build();


        PasswordCredential sshCredential = PasswordCredential
                .newBuilder()
                .setMetadata(metadata)
                .setPassword(password)
                .build();

        if (userId != null) {
            sshCredential = sshCredential.toBuilder().setUserId(userId).build();
        }

        return blockingStub.addPasswordCredential(sshCredential);

    }

    public AddResourceCredentialResponse addPasswordCredential(String token, String clientId,
                                                               String description, String ownerId, String userId, String password) {
        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setClientId(clientId)
                .setDescription(description)
                .setOwnerId(ownerId)
                .setToken(token)
                .build();


        PasswordCredential sshCredential = PasswordCredential
                .newBuilder()
                .setMetadata(metadata)
                .setPassword(password)
                .build();

        if (userId != null) {
            sshCredential = sshCredential.toBuilder().setUserId(userId).build();
        }

        return blockingStub.addPasswordCredential(sshCredential);

    }


    /**
     * Provides SSHCredential of given token
     *
     * @param clientId
     * @param token
     * @return SSHCredential
     */
    public SSHCredential getSSHCredential(String clientId, String token, boolean useShamirSecret) {

        GetResourceCredentialByTokenRequest tokenRequest = GetResourceCredentialByTokenRequest
                .newBuilder()
                .setClientId(clientId)
                .setToken(token)
                .build();
        if (useShamirSecret) {
            tokenRequest = tokenRequest.toBuilder()
                    .setUseShamirsSecretSharingWithEncryption(true)
                    .setNumOfShares(defaultNumOfShares)
                    .setThreshold(defaultThreshold)
                    .build();
        }

        SSHCredential sshCredential = blockingStub.getSSHCredential(tokenRequest);

        if (useShamirSecret) {
            List<ByteString> shares = sshCredential.getPrivateKeySharesList();
            String secret = ShamirSecretHandler.generateSecret(shares, defaultNumOfShares, defaultThreshold);
            sshCredential = sshCredential.toBuilder().setPrivateKey(secret).build();
        }
        return sshCredential;

    }

    /**
     * provides PasswordCredential of given token
     *
     * @param clientId
     * @param token
     * @return PasswordCredential
     */
    public PasswordCredential getPasswordCredential(String clientId, String token) {

        GetResourceCredentialByTokenRequest tokenRequest = GetResourceCredentialByTokenRequest
                .newBuilder()
                .setClientId(clientId)
                .setToken(token)
                .build();


        return blockingStub.getPasswordCredential(tokenRequest);
    }


    /**
     * Delete SSHCredential of given token
     *
     * @param clientId
     * @param token
     * @return ResourceCredentialOperationStatus
     */
    public ResourceCredentialOperationStatus deleteSSHCredential(String clientId, String token) {
        GetResourceCredentialByTokenRequest tokenRequest = GetResourceCredentialByTokenRequest
                .newBuilder()
                .setClientId(clientId)
                .setToken(token)
                .build();

        return blockingStub.deleteSSHCredential(tokenRequest);
    }


    /**
     * Delete Password Credential of given token
     *
     * @param clientId
     * @param token
     * @return ResourceCredentialOperationStatus
     */
    public ResourceCredentialOperationStatus deletePWDCredential(String clientId, String token) {
        GetResourceCredentialByTokenRequest tokenRequest = GetResourceCredentialByTokenRequest
                .newBuilder()
                .setClientId(clientId)
                .setToken(token)
                .build();

        return blockingStub.deletePWDCredential(tokenRequest);
    }


    /**
     * provides resource secret for given owner type and resource type
     *
     * @param ownerType
     * @param resourceType
     * @return
     */
    public SecretMetadata getSecret(ResourceOwnerType ownerType, ResourceType resourceType, ResourceSecretManagementServiceGrpc.
            ResourceSecretManagementServiceBlockingStub blockingStub) {

        SecretMetadata metadata = SecretMetadata
                .newBuilder()
                .setOwnerType(ownerType)
                .setResourceType(resourceType)
                .build();

        GetSecretRequest request = GetSecretRequest.newBuilder().setMetadata(metadata).build();
        return blockingStub.getSecret(request);
    }

    /**
     * provides JWKS keys for calling tenant
     *
     * @return
     */
    public Struct getJWKS() {
        GetJWKSRequest request = GetJWKSRequest.newBuilder().build();
        return blockingStub.getJWKS(request);
    }


    /**
     * Provides metadata object of credentials
     *
     * @param clientId
     * @param token
     * @return SecretMetadata
     */
    public SecretMetadata getResourceCredentialSummary(String clientId, String token, ResourceSecretManagementServiceGrpc.
            ResourceSecretManagementServiceBlockingStub blockingStub) {

        GetResourceCredentialByTokenRequest tokenRequest = GetResourceCredentialByTokenRequest
                .newBuilder()
                .setClientId(clientId)
                .setToken(token)
                .build();

        return blockingStub.getResourceCredentialSummary(tokenRequest);

    }

    /**
     * Provides metadata array of credentials metadata
     *
     * @param clientId
     * @param accessibleTokens
     * @return SecretMetadata[]
     */
    public ResourceCredentialSummaries getAllResourceCredentialSummaries(String clientId, List<String> accessibleTokens, ResourceSecretManagementServiceGrpc.
            ResourceSecretManagementServiceBlockingStub blockingStub) {
        GetResourceCredentialSummariesRequest summariesRequest = GetResourceCredentialSummariesRequest
                .newBuilder()
                .setClientId(clientId)
                .addAllAccessibleTokens(accessibleTokens).build();


        return blockingStub.getAllResourceCredentialSummaries(summariesRequest);

    }

    /**
     * Generate SSH credentials
     *
     * @param clientId
     * @param description
     * @param ownerId
     * @return AddResourceCredentialResponse
     */
    public AddResourceCredentialResponse generateSSHCredential(String clientId, String description, String ownerId, ResourceSecretManagementServiceGrpc.
            ResourceSecretManagementServiceBlockingStub blockingStub) {

        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setClientId(clientId)
                .setDescription(description)
                .setOwnerId(ownerId).build();


        SSHCredential sshCredential = SSHCredential
                .newBuilder()
                .setMetadata(metadata).build();

        return blockingStub.addSSHCredential(sshCredential);

    }

    public AddResourceCredentialResponse addSSHCredential(String token, String passphrase, String privateKey,
                                                          String publicKey, String clientId, String description, String ownerId,
                                                          ResourceSecretManagementServiceGrpc.
                                                                  ResourceSecretManagementServiceBlockingStub blockingStub) {
        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setClientId(clientId)
                .setDescription(description)
                .setOwnerId(ownerId)
                .setToken(token).
                        build();

        SSHCredential sshCredential = SSHCredential
                .newBuilder()
                .setMetadata(metadata)
                .setPassphrase(passphrase)
                .setPrivateKey(privateKey)
                .setPublicKey(publicKey).build();

        return blockingStub.addSSHCredential(sshCredential);

    }


    /**
     * Save password credentials
     *
     * @param clientId
     * @param description
     * @param ownerId
     * @param password
     * @return AddResourceCredentialResponse
     */
    public AddResourceCredentialResponse addPasswordCredential(String clientId, String description,
                                                               String ownerId, String userId, String password,
                                                               ResourceSecretManagementServiceGrpc.
                                                                       ResourceSecretManagementServiceBlockingStub blockingStub) {
        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setClientId(clientId)
                .setDescription(description)
                .setOwnerId(ownerId).build();

        PasswordCredential sshCredential = PasswordCredential
                .newBuilder()
                .setMetadata(metadata)
                .setPassword(password)
                .build();
        if (userId != null) {
            sshCredential = sshCredential.toBuilder().setUserId(userId).build();
        }

        return blockingStub.addPasswordCredential(sshCredential);
    }


    public AddResourceCredentialResponse addPasswordCredential(String token, String clientId,
                                                               String description, String ownerId,
                                                               String userId, String password,
                                                               ResourceSecretManagementServiceGrpc.
                                                                       ResourceSecretManagementServiceBlockingStub blockingStub) {
        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setClientId(clientId)
                .setDescription(description)
                .setOwnerId(ownerId)
                .setToken(token)
                .build();

        PasswordCredential sshCredential = PasswordCredential
                .newBuilder()
                .setMetadata(metadata)
                .setPassword(password)
                .build();

        if (userId != null) {
            sshCredential = sshCredential.toBuilder().setUserId(userId).build();
        }

        return blockingStub.addPasswordCredential(sshCredential);

    }


    /**
     * Provides SSHCredential of given token
     *
     * @param clientId
     * @param token
     * @return SSHCredential
     */
    public SSHCredential getSSHCredential(String clientId, String token,
                                          boolean useShamirSecret,
                                          ResourceSecretManagementServiceGrpc.
                                                  ResourceSecretManagementServiceBlockingStub blockingStub) {

        GetResourceCredentialByTokenRequest tokenRequest = GetResourceCredentialByTokenRequest
                .newBuilder()
                .setClientId(clientId)
                .setToken(token)
                .build();
        if (useShamirSecret) {
            tokenRequest = tokenRequest.toBuilder()
                    .setUseShamirsSecretSharingWithEncryption(true)
                    .setNumOfShares(defaultNumOfShares)
                    .setThreshold(defaultThreshold)
                    .build();
        }

        SSHCredential sshCredential = blockingStub.getSSHCredential(tokenRequest);

        if (useShamirSecret) {
            List<ByteString> shares = sshCredential.getPrivateKeySharesList();
            String secret = ShamirSecretHandler.generateSecret(shares, defaultNumOfShares, defaultThreshold);
            sshCredential = sshCredential.toBuilder().setPrivateKey(secret).build();
        }
        return sshCredential;


    }

    /**
     * provides PasswordCredential of given token
     *
     * @param clientId
     * @param token
     * @return PasswordCredential
     */
    public PasswordCredential getPasswordCredential(String clientId, String token, ResourceSecretManagementServiceGrpc.
            ResourceSecretManagementServiceBlockingStub blockingStub) {

        GetResourceCredentialByTokenRequest tokenRequest = GetResourceCredentialByTokenRequest
                .newBuilder()
                .setClientId(clientId)
                .setToken(token)
                .build();


        return blockingStub.getPasswordCredential(tokenRequest);
    }


    /**
     * Delete SSHCredential of given token
     *
     * @param clientId
     * @param token
     * @return ResourceCredentialOperationStatus
     */
    public ResourceCredentialOperationStatus deleteSSHCredential(String clientId, String token,
                                                                 ResourceSecretManagementServiceGrpc.
                                                                         ResourceSecretManagementServiceBlockingStub blockingStub) {
        GetResourceCredentialByTokenRequest tokenRequest = GetResourceCredentialByTokenRequest
                .newBuilder()
                .setClientId(clientId)
                .setToken(token)
                .build();

        return blockingStub.deleteSSHCredential(tokenRequest);
    }

    public AddResourceCredentialResponse addCredentialMap(String clientId, String description, String ownerId,
                                                          String token, Map<String, String> credentialMap,
                                                          ResourceSecretManagementServiceGrpc.
                                                                  ResourceSecretManagementServiceBlockingStub blockingStub) {
        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setClientId(clientId)
                .setDescription(description)
                .setOwnerId(ownerId)
                .setToken(token).
                        build();

        CredentialMap creMap = CredentialMap
                .newBuilder()
                .setMetadata(metadata)
                .putAllCredentialMap(credentialMap)
                .build();

        return blockingStub.addCredentialMap(creMap);
    }

    public CredentialMap getCredentialMap(String clientId, String token,
                                          ResourceSecretManagementServiceGrpc.
                                                  ResourceSecretManagementServiceBlockingStub blockingStub) {
        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setClientId(clientId)
                .setToken(token).
                        build();

        CredentialMap tokenRequest = CredentialMap
                .newBuilder()
                .setMetadata(metadata)
                .build();

        return blockingStub.getCredentialMap(tokenRequest);
    }

    public ResourceCredentialOperationStatus deleteCredentialMap(String clientId, String token,
                                                                 ResourceSecretManagementServiceGrpc.
                                                                         ResourceSecretManagementServiceBlockingStub blockingStub) {
        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setClientId(clientId)
                .setToken(token).
                        build();

        CredentialMap tokenRequest = CredentialMap
                .newBuilder()
                .setMetadata(metadata)
                .build();

        return blockingStub.deleteCredentialMap(tokenRequest);
    }


    public ResourceCredentialOperationStatus updateCredentialMap(String clientId, String description, String ownerId,
                                                                 String token, Map<String, String> credentialMap,
                                                                 ResourceSecretManagementServiceGrpc.
                                                                         ResourceSecretManagementServiceBlockingStub blockingStub) {
        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setClientId(clientId)
                .setDescription(description)
                .setOwnerId(ownerId)
                .setToken(token).
                        build();

        CredentialMap creMap = CredentialMap
                .newBuilder()
                .setMetadata(metadata)
                .putAllCredentialMap(credentialMap)
                .build();

        return blockingStub.updateCredentialMap(creMap);
    }


    public AddResourceCredentialResponse addCredentialMap(String clientId, String description, String ownerId,
                                                          String token, Map<String, String> credentialMap) {
        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setClientId(clientId)
                .setDescription(description)
                .setOwnerId(ownerId)
                .setToken(token).
                        build();

        CredentialMap creMap = CredentialMap
                .newBuilder()
                .setMetadata(metadata)
                .putAllCredentialMap(credentialMap)
                .build();

        return blockingStub.addCredentialMap(creMap);
    }

    public CredentialMap getCredentialMap(String clientId, String token) {
        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setClientId(clientId)
                .setToken(token).
                        build();

        CredentialMap tokenRequest = CredentialMap
                .newBuilder()
                .setMetadata(metadata)
                .build();

        return blockingStub.getCredentialMap(tokenRequest);
    }

    public ResourceCredentialOperationStatus deleteCredentialMap(String clientId, String token) {
        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setClientId(clientId)
                .setToken(token).
                        build();

        CredentialMap tokenRequest = CredentialMap
                .newBuilder()
                .setMetadata(metadata)
                .build();

        return blockingStub.deleteCredentialMap(tokenRequest);
    }


    public ResourceCredentialOperationStatus updateCredentialMap(String clientId, String description, String ownerId,
                                                                 String token, Map<String, String> credentialMap) {
        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setClientId(clientId)
                .setDescription(description)
                .setOwnerId(ownerId)
                .setToken(token).
                        build();

        CredentialMap creMap = CredentialMap
                .newBuilder()
                .setMetadata(metadata)
                .putAllCredentialMap(credentialMap)
                .build();

        return blockingStub.updateCredentialMap(creMap);
    }


    /**
     * Delete Password Credential of given token
     *
     * @param clientId
     * @param token
     * @return ResourceCredentialOperationStatus
     */
    public ResourceCredentialOperationStatus deletePWDCredential(String clientId, String token, ResourceSecretManagementServiceGrpc.
            ResourceSecretManagementServiceBlockingStub blockingStub) {
        GetResourceCredentialByTokenRequest tokenRequest = GetResourceCredentialByTokenRequest
                .newBuilder()
                .setClientId(clientId)
                .setToken(token)
                .build();

        return blockingStub.deletePWDCredential(tokenRequest);
    }

    ManagedChannel getManagedChannel() {
        return managedChannel;
    }

    void setManagedChannel(ManagedChannel managedChannel) {
        this.managedChannel = managedChannel;
    }

    public ResourceSecretManagementServiceGrpc.ResourceSecretManagementServiceBlockingStub getBlockingStub() {
        return blockingStub;
    }

    public void setBlockingStub(ResourceSecretManagementServiceGrpc.ResourceSecretManagementServiceBlockingStub blockingStub) {
        this.blockingStub = blockingStub;
    }


    @Override
    public void close() throws IOException {
        if (this.managedChannel != null) {
            this.managedChannel.shutdown();
        }
    }
}
