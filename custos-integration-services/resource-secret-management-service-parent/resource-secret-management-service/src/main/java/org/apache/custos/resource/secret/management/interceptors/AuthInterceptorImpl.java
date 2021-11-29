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

import java.util.Optional;

/**
 * Responsible for validate confidential client specific authorization.
 * Methods which authenticates based only on client are implemented here.
 */
@Component
public class AuthInterceptorImpl extends MultiTenantAuthInterceptor {
    private static final Logger LOGGER = LoggerFactory.getLogger(AuthInterceptorImpl.class);

    @Autowired
    public AuthInterceptorImpl(CredentialStoreServiceClient credentialStoreServiceClient, TenantProfileClient tenantProfileClient, IdentityClient identityClient) {
        super(credentialStoreServiceClient, tenantProfileClient, identityClient);
    }

    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT reqT) {

        if (method.equals("getSecret")) {
            Optional<AuthClaim> claim = authorize(headers);
            return claim.map(cl -> {
                String oauthId = cl.getIamAuthId();
                String oauthSec = cl.getIamAuthSecret();

                long tenantId = cl.getTenantId();
                return (ReqT) ((GetSecretRequest) reqT).toBuilder()
                        .setClientId(oauthId)
                        .setClientSec(oauthSec)
                        .setTenantId(tenantId)
                        .build();
            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });


        } else if (method.equals("getJWKS")) {
            Optional<AuthClaim> claim = authorize(headers);

            return claim.map(cl -> {
                String oauthId = cl.getIamAuthId();
                String oauthSec = cl.getIamAuthSecret();

                long tenantId = cl.getTenantId();
                return (ReqT) ((GetJWKSRequest) reqT).toBuilder()
                        .setClientId(oauthId)
                        .setClientSecret(oauthSec)
                        .setTenantId(tenantId)
                        .build();
            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });


        } else if (method.equals("getAllResourceCredentialSummaries")) {
            String clientId = ((GetResourceCredentialSummariesRequest) reqT).getClientId();

            Optional<AuthClaim> claim = authorize(headers, clientId);
            return claim.map(cl -> {
                return (ReqT) ((GetResourceCredentialSummariesRequest) reqT).toBuilder()
                        .setTenantId(cl.getTenantId()).build();

            }).orElseThrow(() -> {

                throw new UnAuthorizedException("Request is not authorized", null);
            });

        } else if (method.equals("addSSHCredential")) {
            String clientId = ((SSHCredential) reqT).getMetadata().getClientId();

            Optional<AuthClaim> claim = authorize(headers, clientId);
            return claim.map(cl -> {
                SecretMetadata metadata = ((SSHCredential) reqT).getMetadata().toBuilder()
                        .setTenantId(cl.getTenantId()).build();

                return (ReqT) ((SSHCredential) reqT).toBuilder().setMetadata(metadata).build();

            }).orElseThrow(() -> {

                throw new UnAuthorizedException("Request is not authorized", null);
            });


        } else if (method.equals("addPasswordCredential")) {
            String clientId = ((PasswordCredential) reqT).getMetadata().getClientId();
            Optional<AuthClaim> claim = authorize(headers, clientId);
            return claim.map(cl -> {
                SecretMetadata metadata = ((PasswordCredential) reqT).
                        getMetadata().toBuilder().setTenantId(cl.getTenantId()).build();

                return (ReqT) ((PasswordCredential) reqT).toBuilder().setMetadata(metadata).build();

            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });


        } else if (method.equals("addCertificateCredential")) {
            String clientId = ((CertificateCredential) reqT).getMetadata().getClientId();

            Optional<AuthClaim> claim = authorize(headers, clientId);
            return claim.map(cl -> {
                SecretMetadata metadata = ((CertificateCredential) reqT).getMetadata()
                        .toBuilder().setTenantId(cl.getTenantId()).build();

                return (ReqT) ((CertificateCredential) reqT).toBuilder().setMetadata(metadata).build();

            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });

        } else if (method.equals("updateCertificateCredential")) {
            String clientId = ((CertificateCredential) reqT).getMetadata().getClientId();

            Optional<AuthClaim> claim = authorize(headers, clientId);
            return claim.map(cl -> {
                SecretMetadata metadata = ((CertificateCredential) reqT).getMetadata()
                        .toBuilder().setTenantId(cl.getTenantId()).build();

                return (ReqT) ((CertificateCredential) reqT).toBuilder().setMetadata(metadata).build();

            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });

        } else if (method.equals("getSSHCredential") || method.equals("getPasswordCredential") || method.equals("getCertificateCredential")
                || method.equals("deleteSSHCredential") || method.equals("deletePWDCredential") || method.equals("deleteCertificateCredential")
                || method.equals("getResourceCredentialSummary")) {
            String clientId = ((GetResourceCredentialByTokenRequest) reqT).getClientId();
            Optional<AuthClaim> claim = authorize(headers, clientId);
            return claim.map(cl -> {
                return (ReqT) ((GetResourceCredentialByTokenRequest) reqT).toBuilder().
                        setTenantId(cl.getTenantId()).build();
            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });

        } else if (method.equals("getKVCredential") || method.equals("addKVCredential") || method.equals("updateKVCredential")
                || method.equals("deleteKVCredential")) {
            String clientId = ((KVCredential) reqT).getMetadata().getClientId();
            String username = ((KVCredential) reqT).getMetadata().getOwnerId();

            Optional<AuthClaim> claim = authorize(headers, clientId);
            return claim.map(cl -> {
                SecretMetadata metadata = ((KVCredential) reqT)
                        .getMetadata()
                        .toBuilder().setOwnerId(cl.getUsername() != null &&
                                !cl.getUsername().isEmpty() ? cl.getUsername() : username)
                        .setTenantId(cl.getTenantId()).build();
                return (ReqT) ((KVCredential) reqT).toBuilder().setMetadata(metadata).build();
            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });


        } else if (method.equals("getCredentialMap") || method.equals("addCredentialMap") || method.equals("updateCredentialMap")
                || method.equals("deleteCredentialMap")) {
            String clientId = ((CredentialMap) reqT).getMetadata().getClientId();
            Optional<AuthClaim> claim = authorize(headers, clientId);

            return claim.map(cl -> {
                SecretMetadata metadata = ((CredentialMap) reqT)
                        .getMetadata()
                        .toBuilder().setTenantId(cl.getTenantId()).build();
                return (ReqT) ((CredentialMap) reqT).toBuilder().setMetadata(metadata).build();
            }).orElseThrow(() -> {
                throw new UnAuthorizedException("Request is not authorized", null);
            });

        }
        return reqT;
    }

}
