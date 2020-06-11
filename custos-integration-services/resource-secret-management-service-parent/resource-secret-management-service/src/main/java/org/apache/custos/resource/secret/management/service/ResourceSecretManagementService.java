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

package org.apache.custos.resource.secret.management.service;

import com.google.protobuf.Struct;
import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import org.apache.custos.cluster.management.client.ClusterManagementClient;
import org.apache.custos.cluster.management.service.GetServerCertificateRequest;
import org.apache.custos.cluster.management.service.GetServerCertificateResponse;
import org.apache.custos.identity.client.IdentityClient;
import org.apache.custos.identity.service.GetJWKSRequest;
import org.apache.custos.resource.secret.client.ResourceSecretClient;
import org.apache.custos.resource.secret.management.service.ResourceSecretManagementServiceGrpc.ResourceSecretManagementServiceImplBase;
import org.apache.custos.resource.secret.service.*;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

@GRpcService
public class ResourceSecretManagementService extends ResourceSecretManagementServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(ResourceSecretManagementService.class);

    @Autowired
    private ClusterManagementClient clusterManagementClient;

    @Autowired
    private IdentityClient identityClient;

    @Autowired
    private ResourceSecretClient resourceSecretClient;

    @Override
    public void getSecret(GetSecretRequest request,
                          StreamObserver<SecretMetadata> responseObserver) {
        LOGGER.debug("Request received to get secret ");
        try {

            if (request.getMetadata().getOwnerType() == ResourceOwnerType.CUSTOS &&
                    request.getMetadata().getResourceType() == ResourceType.SERVER_CERTIFICATE) {

                GetServerCertificateRequest getServerCertificateRequest = GetServerCertificateRequest.newBuilder().build();
                GetServerCertificateResponse response = clusterManagementClient.getCustosServerCertificate(getServerCertificateRequest);

                SecretMetadata metadata = SecretMetadata.newBuilder().setValue(response.getCertificate()).build();
                responseObserver.onNext(metadata);
                responseObserver.onCompleted();
            } else {

            }

        } catch (Exception ex) {
            String msg = "Error occurred while pulling secretes " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getJWKS(GetJWKSRequest request, StreamObserver<Struct> responseObserver) {
        LOGGER.debug("Request received to get JWKS " + request.getTenantId());
        try {

            Struct struct = identityClient.getJWKS(request);

            responseObserver.onNext(struct);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred while pulling JWKS " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getResourceCredentialSummary(GetResourceCredentialByTokenRequest request, StreamObserver<SecretMetadata> responseObserver) {
        LOGGER.debug("Request received to get ResourceCredentialSummary of " + request.getToken());
        try {

            SecretMetadata metadata = resourceSecretClient.getResourceCredentialSummary(request);
            responseObserver.onNext(metadata);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred while fetching resource credential summary : " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getAllResourceCredentialSummaries(GetResourceCredentialSummariesRequest request, StreamObserver<ResourceCredentialSummaries> responseObserver) {
        LOGGER.debug("Request received to get AllResourceCredentialSummaries in tenant " + request.getTenantId());
        try {

            ResourceCredentialSummaries response = resourceSecretClient.getAllResourceCredentialSummaries(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred while fetching all resource credential summaries : " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void addSSHCredential(SSHCredential request, StreamObserver<AddResourceCredentialResponse> responseObserver) {
        LOGGER.debug("Request received to add SSHCredential ");
        try {

            AddResourceCredentialResponse response = resourceSecretClient.addSSHCredential(request);

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred whiling saving SSH credentials :  " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void addPasswordCredential(PasswordCredential request, StreamObserver<AddResourceCredentialResponse> responseObserver) {
        LOGGER.debug("Request received to add PasswordCredential ");
        try {

            AddResourceCredentialResponse response = resourceSecretClient.addPasswordCredential(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred while  saving password credential : " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void addCertificateCredential(CertificateCredential request, StreamObserver<AddResourceCredentialResponse> responseObserver) {
        LOGGER.debug("Request received to add CertificateCredential ");
        try {

            AddResourceCredentialResponse response = resourceSecretClient.addCertificateCredential(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred while saving  certificate credential : " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }


    }

    @Override
    public void getSSHCredential(GetResourceCredentialByTokenRequest request, StreamObserver<SSHCredential> responseObserver) {
        LOGGER.debug("Request received to get SSHCredential ");
        try {

            SSHCredential response = resourceSecretClient.getSSHCredential(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred while fetching  SSH credentials : " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getPasswordCredential(GetResourceCredentialByTokenRequest request, StreamObserver<PasswordCredential> responseObserver) {
        LOGGER.debug("Request received to get PasswordCredential " + request.getTenantId());
        try {

            PasswordCredential response = resourceSecretClient.getPasswordCredential(request);

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred while  fetching password credentials : " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getCertificateCredential(GetResourceCredentialByTokenRequest request, StreamObserver<CertificateCredential> responseObserver) {
        LOGGER.debug("Request received to get CertificateCredential " + request.getTenantId());
        try {

            CertificateCredential response = resourceSecretClient.getCertificateCredential(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred while fetching  certificate credential : " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteSSHCredential(GetResourceCredentialByTokenRequest request, StreamObserver<ResourceCredentialOperationStatus> responseObserver) {
        LOGGER.debug("Request received to delete SSHCredential " + request.getTenantId());
        try {

            ResourceCredentialOperationStatus response = resourceSecretClient.deleteSSHCredential(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while deleting  SSH credential : " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deletePWDCredential(GetResourceCredentialByTokenRequest request, StreamObserver<ResourceCredentialOperationStatus> responseObserver) {
        LOGGER.debug("Request received to delete PWDCredential " + request.getTenantId());
        try {

            ResourceCredentialOperationStatus response = resourceSecretClient.deletePWDCredential(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred while deleting password credential : " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteCertificateCredential(GetResourceCredentialByTokenRequest request, StreamObserver<ResourceCredentialOperationStatus> responseObserver) {
        LOGGER.debug("Request received to delete CertificateCredential " + request.getTenantId());
        try {
            ResourceCredentialOperationStatus response = resourceSecretClient.deleteCertificateCredential(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while deleting  certificate credential :  " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }
}
