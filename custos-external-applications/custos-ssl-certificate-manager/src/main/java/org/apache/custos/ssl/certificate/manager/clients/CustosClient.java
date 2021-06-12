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

package org.apache.custos.ssl.certificate.manager.clients;

import org.apache.custos.clients.CustosClientProvider;
import org.apache.custos.resource.secret.management.client.ResourceSecretManagementClient;
import org.apache.custos.resource.secret.service.AddResourceCredentialResponse;
import org.apache.custos.resource.secret.service.KVCredential;
import org.apache.custos.resource.secret.service.ResourceCredentialOperationStatus;
import org.apache.custos.ssl.certificate.manager.configurations.CustosConfiguration;

import java.io.IOException;

/**
 * Custos client class to perform custos related operations
 */
public class CustosClient implements AutoCloseable{

    private String clientId;
    private String ownerId;
    private ResourceSecretManagementClient resourceSecretManagementClient;

    public CustosClient(CustosConfiguration config) throws IOException {
        CustosClientProvider provider = new CustosClientProvider
                .Builder()
                .setServerHost(config.getUrl())
                .setServerPort(config.getPort())
                .setClientId(config.getClientId())
                .setClientSec(config.getClientSecret())
                .build();
        this.resourceSecretManagementClient = provider.getResourceSecretManagementClient();
        this.clientId = config.getClientId();
        this.ownerId = config.getOwnerId();
    }


    /**
     * Add ssl certificate to custos
     *
     * @param privateKey  private key
     * @param certificate certificate
     * @return a token received from custos
     */
    public String addCertificate(String privateKey, String certificate) {
        AddResourceCredentialResponse res = resourceSecretManagementClient.addCertificateCredentials(
                clientId, ownerId, privateKey, certificate);
        return res.getToken();
    }


    /**
     * Add key value credential to custos
     *
     * @param key   key
     * @param value value
     * @return status of the operation
     */
    public boolean addKVCredential(String key, String value) {
        ResourceCredentialOperationStatus status =
                resourceSecretManagementClient.addKVCredentials(clientId, ownerId, key, value);
        return status.getStatus();
    }


    /**
     * Retrieves value for key from custos
     *
     * @param key key for the value needs to be retrieved
     * @return value for the key
     */
    public String getKVCredentials(String key) {
        KVCredential kvCredential = resourceSecretManagementClient.getKVCredentials(clientId, ownerId, key);
        return kvCredential.getValue();
    }


    /**
     * Close the resource management client stream
     *
     * @throws IOException if an I/O error occurs
     */
    @Override
    public void close() throws IOException {
        resourceSecretManagementClient.close();
    }
}