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
import org.apache.custos.ssl.certificate.manager.configurations.CustosConfiguration;
import org.shredzone.acme4j.Certificate;

import java.io.IOException;

public class CustosClient {

    CustosConfiguration config;
    ResourceSecretManagementClient resourceSecretManagementClient;

    public CustosClient(CustosConfiguration config) throws IOException {
        this.config = config;
        CustosClientProvider provider = new CustosClientProvider
                .Builder()
                .setServerHost(config.getUrl())
                .setServerPort(config.getPort())
                .setClientId(config.getClientId())
                .setClientSec(config.getClientSecret())
                .build();

        this.resourceSecretManagementClient = provider.getResourceSecretManagementClient();
    }

    public String addCertificate(String privateKey, Certificate certificate) {
        AddResourceCredentialResponse res = resourceSecretManagementClient.addCertificateCredentials(
                this.config.getClientId(),
                this.config.getOwnerId(),
                privateKey,
                certificate.getCertificate().toString()
        );

        return res.getToken();
    }

    public void close() throws IOException {
        resourceSecretManagementClient.close();
    }
}
