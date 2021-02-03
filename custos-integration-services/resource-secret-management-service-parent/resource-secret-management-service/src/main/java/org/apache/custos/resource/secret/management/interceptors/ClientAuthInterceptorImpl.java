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

package org.apache.custos.resource.secret.management.interceptors;

import io.grpc.Metadata;
import org.apache.custos.credential.store.client.CredentialStoreServiceClient;
import org.apache.custos.identity.client.IdentityClient;
import org.apache.custos.identity.service.GetJWKSRequest;
import org.apache.custos.integration.core.exceptions.UnAuthorizedException;
import org.apache.custos.integration.services.commons.interceptors.MultiTenantAuthInterceptor;
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.resource.secret.service.*;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

/**
 * Responsible for validate confidential client specific authorization.
 * Methods which authenticates based only on client are implemented here.
 */
@Component
public class ClientAuthInterceptorImpl extends MultiTenantAuthInterceptor {
    private static final Logger LOGGER = LoggerFactory.getLogger(ClientAuthInterceptorImpl.class);

    @Autowired
    public ClientAuthInterceptorImpl(CredentialStoreServiceClient credentialStoreServiceClient, TenantProfileClient tenantProfileClient, IdentityClient identityClient) {
        super(credentialStoreServiceClient, tenantProfileClient, identityClient);
    }

    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT reqT) {

        if (method.equals("getSecret")) {
            AuthClaim claim = authorize(headers);

            if (claim == null) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();

            long tenantId = claim.getTenantId();
            return (ReqT) ((GetSecretRequest) reqT).toBuilder()
                    .setClientId(oauthId)
                    .setClientSec(oauthSec)
                    .setTenantId(tenantId)
                    .build();

        } else if (method.equals("getJWKS")) {
            AuthClaim claim = authorize(headers);

            if (claim == null) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }

            String oauthId = claim.getIamAuthId();
            String oauthSec = claim.getIamAuthSecret();

            long tenantId = claim.getTenantId();
            return (ReqT) ((GetJWKSRequest) reqT).toBuilder()
                    .setClientId(oauthId)
                    .setClientSecret(oauthSec)
                    .setTenantId(tenantId)
                    .build();

        } else if (method.equals("getAllResourceCredentialSummaries")) {
            String clientId = ((GetResourceCredentialSummariesRequest) reqT).getClientId();

            AuthClaim claim = authorize(headers, clientId);
            if (claim == null) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }
            return (ReqT) ((GetResourceCredentialSummariesRequest) reqT).toBuilder().setTenantId(claim.getTenantId()).build();


        } else if (method.equals("addSSHCredential")) {
            String clientId = ((SSHCredential) reqT).getMetadata().getClientId();

            AuthClaim claim = authorize(headers, clientId);
            if (claim == null) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }
            SecretMetadata metadata = ((SSHCredential) reqT).getMetadata().toBuilder().setTenantId(claim.getTenantId()).build();

            return (ReqT) ((SSHCredential) reqT).toBuilder().setMetadata(metadata).build();


        } else if (method.equals("addPasswordCredential")) {
            String clientId = ((PasswordCredential) reqT).getMetadata().getClientId();

            AuthClaim claim = authorize(headers, clientId);
            if (claim == null) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }
            SecretMetadata metadata = ((PasswordCredential) reqT).getMetadata().toBuilder().setTenantId(claim.getTenantId()).build();

            return (ReqT) ((PasswordCredential) reqT).toBuilder().setMetadata(metadata).build();

        } else if (method.equals("addCertificateCredential")) {
            String clientId = ((CertificateCredential) reqT).getMetadata().getClientId();

            AuthClaim claim = authorize(headers, clientId);
            if (claim == null) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }
            SecretMetadata metadata = ((CertificateCredential) reqT).getMetadata().toBuilder().setTenantId(claim.getTenantId()).build();

            return (ReqT) ((CertificateCredential) reqT).toBuilder().setMetadata(metadata).build();

        } else if (method.equals("getSSHCredential") || method.equals("getPasswordCredential") || method.equals("getCertificateCredential")
                || method.equals("deleteSSHCredential") || method.equals("deletePWDCredential") || method.equals("deleteCertificateCredential")
                || method.equals("getResourceCredentialSummary")) {
            String clientId = ((GetResourceCredentialByTokenRequest) reqT).getClientId();

            AuthClaim claim = authorize(headers, clientId);
            if (claim == null) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }
            return (ReqT) ((GetResourceCredentialByTokenRequest) reqT).toBuilder().setTenantId(claim.getTenantId()).build();

        } else if (method.equals("getKVCredential") || method.equals("addKVCredential") || method.equals("updateKVCredential")
                || method.equals("deleteKVCredential")) {
            String clientId = ((KVCredential) reqT).getMetadata().getClientId();

            AuthClaim claim = authorize(headers, clientId);
            if (claim == null) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }
            SecretMetadata metadata = ((KVCredential) reqT)
                    .getMetadata()
                    .toBuilder().setOwnerId(claim.getUsername()).setTenantId(claim.getTenantId()).build();
            return (ReqT) ((KVCredential) reqT).toBuilder().setMetadata(metadata).build();

        }
        return reqT;
    }

}
