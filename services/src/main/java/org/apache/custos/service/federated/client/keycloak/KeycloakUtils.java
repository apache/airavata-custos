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

package org.apache.custos.service.federated.client.keycloak;

import org.keycloak.admin.client.Keycloak;
import org.keycloak.admin.client.KeycloakBuilder;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import javax.net.ssl.SSLContext;
import javax.net.ssl.TrustManager;
import javax.net.ssl.TrustManagerFactory;
import javax.ws.rs.client.Client;
import javax.ws.rs.client.ClientBuilder;
import java.security.KeyManagementException;
import java.security.KeyStore;
import java.security.KeyStoreException;
import java.security.NoSuchAlgorithmException;
import java.security.SecureRandom;
import java.util.concurrent.TimeUnit;

public class KeycloakUtils {

    private static final Logger LOGGER = LoggerFactory.getLogger(KeycloakUtils.class);
    private static final String SSL_PROTOCOL = "TLS";

    public static Keycloak getClient(String serverURL, String realm, String accessToken) {
        return KeycloakBuilder.builder()
                .serverUrl(serverURL)
                .realm(realm)
                .authorization(accessToken)
                .resteasyClient(getRestClient())
                .build();
    }

    public static Keycloak getClient(String serverURL, String realm, String loginUsername, String password, String clientId) {
        return KeycloakBuilder.builder()
                .serverUrl(serverURL)
                .realm(realm)
                .username(loginUsername)
                .password(password)
                .clientId(clientId)
                .resteasyClient(getRestClient())
                .build();
    }

    private static Client getRestClient() {
        try {
            TrustManagerFactory tmf = TrustManagerFactory.getInstance(TrustManagerFactory.getDefaultAlgorithm());
            tmf.init((KeyStore) null);

            SSLContext sslContext = SSLContext.getInstance(SSL_PROTOCOL);
            sslContext.init(null, tmf.getTrustManagers(), null);

            return ClientBuilder.newBuilder()
                    .sslContext(sslContext)
                    .connectTimeout(100, TimeUnit.SECONDS)
                    .readTimeout(10, TimeUnit.SECONDS)
                    .build();
        } catch (Exception e) {
            LOGGER.error("Error configuring the rest client", e);
            throw new RuntimeException("Failed to configure the REST client", e);
        }
    }

    public static void initializeTrustStoreManager() throws KeyStoreException, NoSuchAlgorithmException, KeyManagementException {

        TrustManagerFactory trustFactory = TrustManagerFactory.getInstance(TrustManagerFactory.getDefaultAlgorithm());
        trustFactory.init((KeyStore) null);

        // get the trust managers from the factory
        TrustManager[] trustManagers = trustFactory.getTrustManagers();

        // initialize an SSL context to use these managers and set as default
        SSLContext sslContext = SSLContext.getInstance(SSL_PROTOCOL);
        sslContext.init(null, trustManagers, new SecureRandom());
        SSLContext.setDefault(sslContext);
    }

}
