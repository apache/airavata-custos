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

package org.apache.custos.federated.services.clients.keycloak;

import org.apache.catalina.security.SecurityUtil;
import org.apache.custos.cluster.management.client.ClusterManagementClient;
import org.apache.custos.cluster.management.service.GetServerCertificateRequest;
import org.apache.custos.cluster.management.service.GetServerCertificateResponse;
import org.jboss.resteasy.client.jaxrs.ResteasyClient;
import org.jboss.resteasy.client.jaxrs.ResteasyClientBuilder;
import org.keycloak.admin.client.Keycloak;
import org.keycloak.admin.client.KeycloakBuilder;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import javax.net.ssl.SSLContext;
import javax.net.ssl.TrustManager;
import javax.net.ssl.TrustManagerFactory;
import java.io.*;
import java.security.KeyManagementException;
import java.security.KeyStore;
import java.security.KeyStoreException;
import java.security.NoSuchAlgorithmException;
import java.security.cert.Certificate;
import java.security.cert.CertificateException;
import java.util.concurrent.TimeUnit;
import java.security.cert.CertificateFactory;

public class KeycloakUtils {

    private final static int POOL_SIZE = 10;

    private static final Logger LOGGER = LoggerFactory.getLogger(KeycloakUtils.class);


    public static Keycloak getClient(String serverURL, String realm, String accessToken,
                                     String trustStorePath, String trustorePassword, String profile,
                                     ClusterManagementClient clusterManagementClient) {

        return KeycloakBuilder.builder()
                .serverUrl(serverURL)
                .realm(realm)
                .authorization(accessToken)
                .resteasyClient(getRestClient(trustStorePath, trustorePassword, profile, clusterManagementClient))
                .build();
    }


    public static Keycloak getClient(String serverURL, String realm, String loginUsername,
                                     String password, String clientId, String trustStorePath, String trustorePassword, String profile,
                                     ClusterManagementClient clusterManagementClient) {

        return KeycloakBuilder.builder()
                .serverUrl(serverURL)
                .realm(realm)
                .username(loginUsername)
                .password(password)
                .clientId(clientId)
                .resteasyClient(getRestClient(trustStorePath, trustorePassword, profile, clusterManagementClient))
                .build();
    }


    private static ResteasyClient getRestClient(String trustorePath, String trustorePassword, String profile,
                                                ClusterManagementClient clusterManagementClient) {
        return new ResteasyClientBuilder()
                .establishConnectionTimeout(100, TimeUnit.SECONDS)
                .socketTimeout(10, TimeUnit.SECONDS)
                .connectionPoolSize(POOL_SIZE)
                .trustStore(loadKeyStore(trustorePath, trustorePassword, profile, clusterManagementClient))
                .build();
    }


    private static KeyStore loadKeyStore(String trustStorePath, String trustorePassword,
                                         String profile, ClusterManagementClient clusterManagementClient) {

        InputStream is = null;
        try {

            KeyStore ks = KeyStore.getInstance(KeyStore.getDefaultType());
            LOGGER.info("Profile " + profile);
            if (profile.equals("staging") || profile.equals("prod")) {
                LOGGER.info("Profile inside  " + profile);
                GetServerCertificateRequest getServerCertificateRequest = GetServerCertificateRequest
                        .newBuilder()
                        .setNamespace("keycloak")
                        .setSecretName("tls-keycloak-secret")
                        .build();
                GetServerCertificateResponse response = clusterManagementClient.getCustosServerCertificate(getServerCertificateRequest);
                CertificateFactory cf = CertificateFactory.getInstance("X.509");
                LOGGER.info(response.getCertificate());
                InputStream targetStream = new ByteArrayInputStream(response.getCertificate().getBytes());
                Certificate certs = cf.generateCertificate(targetStream);
                // Add the certificate
                ks.load(null, null);
                ks.setCertificateEntry("custos", certs);

            } else {

                File trustStoreFile = new File(trustStorePath);

                if (trustStoreFile.exists()) {
                    LOGGER.debug("Loading trust store file from path " + trustStorePath);
                    is = new FileInputStream(trustStorePath);
                } else {
                    LOGGER.debug("Trying to load trust store file form class path " + trustStorePath);
                    is = SecurityUtil.class.getClassLoader().getResourceAsStream(trustStorePath);
                    if (is != null) {
                        LOGGER.debug("Trust store file was loaded form class path " + trustStorePath);
                    }
                }

                if (is == null) {
                    throw new RuntimeException("Could not find a trust store file in path " + trustStorePath);
                }


                ks.load(is, trustorePassword.toCharArray());
            }
            return ks;
        } catch (Exception e) {
            throw new RuntimeException("Failed to load trust store KeyStore instance", e);
        } finally {
            if (is != null) {
                try {
                    is.close();
                } catch (IOException e) {
                    LOGGER.error("Failed to close trust store FileInputStream", e);
                }
            }
        }
    }


    public static SSLContext initializeTrustStoreManager(String trustStorePath, String trustStorePassword,
                                                         String profile, ClusterManagementClient clusterManagementClient) throws
            IOException, KeyStoreException, CertificateException, NoSuchAlgorithmException, KeyManagementException {

        KeyStore trustStore = KeyStore.getInstance(KeyStore.getDefaultType());
        if (profile.equals("staging") || profile.equals("prod")) {
            GetServerCertificateRequest getServerCertificateRequest = GetServerCertificateRequest
                    .newBuilder()
                    .setNamespace("keycloak")
                    .setSecretName("tls-keycloak-secret")
                    .build();
            GetServerCertificateResponse response = clusterManagementClient.getCustosServerCertificate(getServerCertificateRequest);
            CertificateFactory cf = CertificateFactory.getInstance("X.509");
            InputStream targetStream = new ByteArrayInputStream(response.getCertificate().getBytes());
            Certificate certs = cf.generateCertificate(targetStream);
            trustStore.load(null, null);
            trustStore.setCertificateEntry("custos", certs);


        } else {
            File trustStoreFile = new File(trustStorePath);
            InputStream is;
            if (trustStoreFile.exists()) {
                LOGGER.debug("Loading trust store file from path " + trustStorePath);
                is = new FileInputStream(trustStorePath);
            } else {
                LOGGER.debug("Trying to load trust store file form class path " + trustStorePath);
                is = SecurityUtil.class.getClassLoader().getResourceAsStream(trustStorePath);
                if (is != null) {
                    LOGGER.debug("Trust store file was loaded form class path " + trustStorePath);
                }
            }

            if (is == null) {
                throw new RuntimeException("Could not find a trust store file in path " + trustStorePath);
            }
            char[] trustPassword = trustStorePassword.toCharArray();
            trustStore.load(is, trustPassword);
        }


        // initialize a trust manager factory
        TrustManagerFactory trustFactory =
                TrustManagerFactory.getInstance(TrustManagerFactory.getDefaultAlgorithm());
        trustFactory.init(trustStore);

        // get the trust managers from the factory
        TrustManager[] trustManagers = trustFactory.getTrustManagers();

        // initialize an ssl context to use these managers and set as default
        SSLContext sslContext = SSLContext.getInstance("SSL");
        sslContext.init(null, trustManagers, null);
        SSLContext.setDefault(sslContext);
        return sslContext;

    }

}
