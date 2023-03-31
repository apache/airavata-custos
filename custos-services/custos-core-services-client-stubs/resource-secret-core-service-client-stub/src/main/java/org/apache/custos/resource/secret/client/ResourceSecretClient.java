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

package org.apache.custos.resource.secret.client;

import io.grpc.ClientInterceptor;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import org.apache.custos.resource.secret.service.*;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.List;

@Component
public class ResourceSecretClient {

    private ManagedChannel managedChannel;
    private ResourceSecretServiceGrpc.ResourceSecretServiceBlockingStub resourceSecretServiceBlockingStub;

    private final List<ClientInterceptor> clientInterceptorList;

    public ResourceSecretClient(List<ClientInterceptor> clientInterceptorList,
                                @Value("${core.services.server.hostname:localhost}") String serviceHost,
                                @Value("${core.services.server.port:7070}") int servicePort) {
        this.clientInterceptorList = clientInterceptorList;
        managedChannel = ManagedChannelBuilder.forAddress(
                serviceHost, servicePort).usePlaintext().intercept(clientInterceptorList).build();
        resourceSecretServiceBlockingStub = ResourceSecretServiceGrpc.newBlockingStub(managedChannel);
    }


    public SecretMetadata getResourceCredentialSummary(GetResourceCredentialByTokenRequest request) {

        return resourceSecretServiceBlockingStub.getResourceCredentialSummary(request);
    }

    public ResourceCredentialSummaries getAllResourceCredentialSummaries(GetResourceCredentialSummariesRequest request) {
        return resourceSecretServiceBlockingStub.getAllResourceCredentialSummaries(request);
    }

    public AddResourceCredentialResponse addSSHCredential(SSHCredential credential) {
        return resourceSecretServiceBlockingStub.addSSHCredential(credential);
    }

    public AddResourceCredentialResponse addPasswordCredential(PasswordCredential credential) {
        return resourceSecretServiceBlockingStub.addPasswordCredential(credential);
    }

    public AddResourceCredentialResponse addCertificateCredential(CertificateCredential certificateCredential) {
        return resourceSecretServiceBlockingStub.addCertificateCredential(certificateCredential);
    }

    public SSHCredential getSSHCredential(GetResourceCredentialByTokenRequest request) {
        return resourceSecretServiceBlockingStub.getSSHCredential(request);
    }

    public PasswordCredential getPasswordCredential(GetResourceCredentialByTokenRequest request) {
        return resourceSecretServiceBlockingStub.getPasswordCredential(request);
    }

    public CertificateCredential getCertificateCredential(GetResourceCredentialByTokenRequest request) {
        return resourceSecretServiceBlockingStub.getCertificateCredential(request);
    }

    public ResourceCredentialOperationStatus deleteSSHCredential(GetResourceCredentialByTokenRequest request) {
        return resourceSecretServiceBlockingStub.deleteSSHCredential(request);
    }

    public ResourceCredentialOperationStatus deletePWDCredential(GetResourceCredentialByTokenRequest request) {
        return resourceSecretServiceBlockingStub.deletePWDCredential(request);
    }

    public ResourceCredentialOperationStatus deleteCertificateCredential(GetResourceCredentialByTokenRequest request) {
        return resourceSecretServiceBlockingStub.deleteCertificateCredential(request);
    }


    public KVCredential getKVCredential(KVCredential request) {
        return resourceSecretServiceBlockingStub.getKVCredential(request);
    }

    public ResourceCredentialOperationStatus setKVCredential(KVCredential request) {
        return resourceSecretServiceBlockingStub.setKVCredential(request);
    }

    public ResourceCredentialOperationStatus updateKVCredential(KVCredential request) {
        return resourceSecretServiceBlockingStub.updateKVCredential(request);
    }

    public ResourceCredentialOperationStatus deleteKVCredential(KVCredential request) {
        return resourceSecretServiceBlockingStub.deleteKVCredential(request);
    }

    public CredentialMap getCredentialMap(CredentialMap request) {
        return resourceSecretServiceBlockingStub.getCredentialMap(request);
    }

    public AddResourceCredentialResponse setCredentialMap(CredentialMap request) {
        return resourceSecretServiceBlockingStub.setCredentialMap(request);
    }

    public ResourceCredentialOperationStatus updateCredentialMap(CredentialMap request) {
        return resourceSecretServiceBlockingStub.updateCredentialMap(request);
    }

    public ResourceCredentialOperationStatus deleteCredentialMap(CredentialMap request) {
        return resourceSecretServiceBlockingStub.deleteCredentialMap(request);
    }


    public ResourceCredentialOperationStatus updateCertificate(CertificateCredential request) {
        return resourceSecretServiceBlockingStub.updateCertificateCredential(request);
    }
}
