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

package org.apache.custos.cluster.management.service;

import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import io.kubernetes.client.ApiClient;
import io.kubernetes.client.Configuration;
import io.kubernetes.client.apis.CoreV1Api;
import io.kubernetes.client.models.V1Secret;
import io.kubernetes.client.util.ClientBuilder;
import org.apache.custos.cluster.management.service.ClusterManagementServiceGrpc.ClusterManagementServiceImplBase;
import org.apache.custos.core.services.api.commons.StatusUpdater;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;

import javax.net.ssl.HttpsURLConnection;
import javax.net.ssl.SSLSocket;
import javax.net.ssl.SSLSocketFactory;
import java.security.PublicKey;
import java.security.cert.Certificate;
import java.util.Map;

@GRpcService
public class ClusterManagementService extends ClusterManagementServiceImplBase {

    private static Logger LOGGER = LoggerFactory.getLogger(ClusterManagementService.class);

    @Autowired
    private StatusUpdater statusUpdater;

    @Value("${custos.server.host:dev.services.usecustos.org}")
    private String custosServerHostName;

    @Override
    public void getCustosServerCertificate(org.apache.custos.cluster.management.service.GetServerCertificateRequest request,
                                           StreamObserver<org.apache.custos.cluster.management.service.GetServerCertificateResponse> responseObserver) {
        try {

//            ApiClient client = ClientBuilder.cluster().build();
//
//            // set the global default api-client to the in-cluster one from above
//            Configuration.setDefaultApiClient(client);
//
//            // the CoreV1Api loads default api-client from global configuration.
//            CoreV1Api api = new CoreV1Api();
//
//            String namespace = request.getNamespace();
//            String secretName = request.getSecretName();
//
//
//            V1Secret secret = api.readNamespacedSecret(secretName.isEmpty() ? custosServerSecretName : secretName,
//                    namespace.isEmpty() ? custosNameSpace : namespace, null, null, null);
//            Map<String, byte[]> map = secret.getData();
//            byte[] cert = map.get("tls.crt");

            SSLSocketFactory factory = HttpsURLConnection.getDefaultSSLSocketFactory();
            SSLSocket socket = (SSLSocket) factory.createSocket(custosServerHostName, 443);
            socket.startHandshake();
            Certificate[] certs = socket.getSession().getPeerCertificates();
            Certificate cert = certs[0];
            PublicKey key = cert.getPublicKey();

            if (cert == null || cert.getEncoded().length == 0) {
                GetServerCertificateResponse response = GetServerCertificateResponse.newBuilder().build();
                responseObserver.onNext(response);
                responseObserver.onCompleted();
            } else {
                String certificate = new String(cert.getEncoded());
                GetServerCertificateResponse response = GetServerCertificateResponse
                        .newBuilder()
                        .setCertificate(certificate)
                        .build();
                responseObserver.onNext(response);
                responseObserver.onCompleted();
            }

        } catch (Exception ex) {
            String msg = "Exception occurred while fetching custos server certificate" + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }
}
