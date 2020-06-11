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
                                @Value("${resource.secret.service.dns.name}") String serviceHost,
                                @Value("${resource.secret.service.port}") int servicePort) {
        this.clientInterceptorList = clientInterceptorList;
        managedChannel = ManagedChannelBuilder.forAddress(
                serviceHost, servicePort).usePlaintext(true).intercept(clientInterceptorList).build();
        resourceSecretServiceBlockingStub = ResourceSecretServiceGrpc.newBlockingStub(managedChannel);
    }


    public SecretMetadata getSecretResponse(GetSecretRequest request) {
        return resourceSecretServiceBlockingStub.getSecret(request);

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


}
