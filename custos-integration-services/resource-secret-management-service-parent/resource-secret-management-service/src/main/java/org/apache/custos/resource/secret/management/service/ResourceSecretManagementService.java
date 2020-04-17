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
import org.apache.custos.resource.secret.management.service.ResourceSecretManagementServiceGrpc.ResourceSecretManagementServiceImplBase;
import org.apache.custos.resource.secret.service.GetSecretRequest;
import org.apache.custos.resource.secret.service.ResourceOwnerType;
import org.apache.custos.resource.secret.service.ResourceType;
import org.apache.custos.resource.secret.service.SecretMetadata;
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

    @Override
    public void getSecret(GetSecretRequest request,
                          StreamObserver<SecretMetadata> responseObserver) {
        LOGGER.debug("Request received to getSecret ");
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
        LOGGER.debug("Request received to getJWKS " + request.getTenantId());
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
}
