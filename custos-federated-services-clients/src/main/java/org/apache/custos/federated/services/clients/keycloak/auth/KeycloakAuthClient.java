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

package org.apache.custos.federated.services.clients.keycloak.auth;

import org.apache.custos.federated.services.clients.keycloak.KeycloakUtils;
import org.apache.http.Consts;
import org.apache.http.HttpHeaders;
import org.apache.http.NameValuePair;
import org.apache.http.client.entity.UrlEncodedFormEntity;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.conn.ssl.SSLConnectionSocketFactory;
import org.apache.http.conn.ssl.TrustSelfSignedStrategy;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.message.BasicNameValuePair;
import org.apache.http.ssl.SSLContextBuilder;
import org.apache.http.util.EntityUtils;
import org.json.JSONException;
import org.json.JSONObject;
import org.keycloak.authorization.client.AuthzClient;
import org.keycloak.authorization.client.Configuration;
import org.keycloak.representations.AccessTokenResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.event.EventListener;
import org.springframework.stereotype.Component;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.security.KeyManagementException;
import java.security.KeyStoreException;
import java.security.NoSuchAlgorithmException;
import java.security.cert.CertificateException;
import java.util.*;

/**
 * Acting as a broker between keycloak server and auth services
 */
@Component
public class KeycloakAuthClient {

    @Value("${iam.server.url:https://keycloak.custos.scigap.org:31000/auth/}")
    private String idpServerURL;

    @Value("${iam.server.truststore.path:/home/ubuntu/keystore/keycloak-client-truststore.pkcs12}")
    private String trustStorePath;

    @Value("${iam.server.truststore.password:keycloak}")
    private String trustStorePassword;

    @Value("${introspection.endpoint}")
    private String introEndpoint;

    @Value("${issuer}")
    private String issuer;

    @Value("${authorization.endpoint}")
    private String authorizationEndpoint;

    @Value("${token.endpoint}")
    private String tokenEndpoint;

    @Value("${end.session.endpoint}")
    private String sessionEndpoint;

    @Value("${jwks_uri}")
    private String jwksUri;

    @Value("${registration.endpoint}")
    private String registrationEndpoint;

    @Value("${user.info.endpoint}")
    private String userInfoEndpoint;

    private static final Logger LOGGER = LoggerFactory.getLogger(KeycloakAuthClient.class);

    public KeycloakAuthClient() {

    }

    @EventListener(ApplicationReadyEvent.class)
    public void initializeSecurity() throws CertificateException, NoSuchAlgorithmException,
            KeyStoreException, KeyManagementException, IOException {
        try {
            LOGGER.info("initializing security requirements");
            KeycloakUtils.initializeTrustStoreManager(trustStorePath, trustStorePassword);
        } catch (Exception ex) {
            LOGGER.error("Keycloak Authclient initialization failed " + ex.getMessage());
            throw ex;
        }
    }


    public String authenticate(String clientId, String clientSecret, String realmId, String username, String password) {

        try {
            Map<String, Object> clientCredentials = new HashMap<>();
            clientCredentials.put("secret", clientSecret);
            SSLContextBuilder builder = new SSLContextBuilder();

            builder.loadTrustMaterial(null, new TrustSelfSignedStrategy());


            SSLConnectionSocketFactory sslsf = new SSLConnectionSocketFactory(builder.build());
            CloseableHttpClient httpclient = HttpClients.custom().setSSLSocketFactory(sslsf).build();

            Configuration configuration = new Configuration(idpServerURL,
                    realmId, clientId, clientCredentials, httpclient);
            AuthzClient keycloakClient = AuthzClient.create(configuration);


            AccessTokenResponse accessToken = keycloakClient.obtainAccessToken(username, password);



            if (accessToken != null) {
                return accessToken.getToken();
            }

            return null;

        } catch (Exception e) {
            String msg = "Error occurred while authenticating " + e;
            LOGGER.error(msg);
            throw new RuntimeException(msg, e);
        }

    }


    public boolean isUserAuthenticated(String username, String realmId, String accessToken) {
        try {
            User userInfo = getUserInfo(realmId, accessToken);
            if (!username.equals(userInfo.getUsername())) {
                throw new RuntimeException("Subject name and username for the token doesn't match");
            }
            return true;
        } catch (Exception e) {
            String msg = "Error occurred while checking if user: " + username + " is authorized in gateway: " + realmId;
            LOGGER.error(msg, e);
            throw new RuntimeException(msg, e);
        }

    }


    public User getUser(String accessToken, String realmId) {
        try {
            return getUserInfo(realmId, accessToken);
        } catch (Exception e) {
            String msg = "Error occurred while retrieving user info " + e;
            LOGGER.error(msg);
            throw new RuntimeException(msg, e);
        }

    }

    public String getUserManagementServiceAccountAccessToken(String clientId, String clientSecret,
                                                             String realmId) {
        try {
            String tokenURL = getTokenEndpoint(realmId);
            JSONObject clientCredentials = getClientCredentials(tokenURL, clientId, clientSecret);
            return clientCredentials.getString("access_token");
        } catch (Exception e) {
            String msg = "Error occurred while retrieving service account access token  " + e;
            LOGGER.error(msg);
            throw new RuntimeException(msg, e);
        }
    }

    public JSONObject getAccessToken(String clientId, String clientSecret, String realmId,
                                     String code, String redirectUri) throws JSONException {
        try {
            String tokenURL = getTokenEndpoint(realmId);
            return getTokenFromOAuthCode(tokenURL, clientId, clientSecret, code, redirectUri);

        } catch (Exception e) {
            String msg = "Error occurred while retrieving  access token  " + e;
            LOGGER.error(msg);
            throw new RuntimeException(msg, e);
        }

    }

    public JSONObject getAccessTokenFromPasswordGrantType(String clientId, String clientSecret, String realmId,
                                     String username, String password) throws JSONException {
        try {
            String tokenURL = getTokenEndpoint(realmId);
            return getTokenFromPasswordType(tokenURL, clientId, clientSecret, username, password);

        } catch (Exception e) {
            String msg = "Error occurred while retrieving  access token  " + e;
            LOGGER.error(msg);
            throw new RuntimeException(msg, e);
        }

    }

    public JSONObject getAccessTokenFromRefreshTokenGrantType(String clientId, String clientSecret, String realmId,
                                                          String refreshToken) throws JSONException {
        try {
            String tokenURL = getTokenEndpoint(realmId);
            return getTokenFromRefreshToken(tokenURL, clientId, clientSecret, refreshToken);

        } catch (Exception e) {
            String msg = "Error occurred while retrieving  access token  " + e;
            LOGGER.error(msg);
            throw new RuntimeException(msg, e);
        }

    }



    private String getTokenEndpoint(String gatewayId) throws Exception {
        String openIdConnectUrl = getOpenIDConfigurationUrl(gatewayId);
        JSONObject openIdConnectConfig = new JSONObject(getFromUrl(openIdConnectUrl, null));
        return openIdConnectConfig.getString("token_endpoint");
    }

    public String getAuthorizationEndpoint(String gatewayId) throws Exception {
        String openIdConnectUrl = getOpenIDConfigurationUrl(gatewayId);
        JSONObject openIdConnectConfig = new JSONObject(getFromUrl(openIdConnectUrl, null));
        return openIdConnectConfig.getString("authorization_endpoint");
    }


    public JSONObject getOIDCConfiguration(String tenantId, String clientId) throws Exception {
        String openIdConnectUrl = getOpenIDConfigurationUrl(tenantId);
        JSONObject openIdConnectConfig = new JSONObject(getFromUrl(openIdConnectUrl, null));

       // openIdConnectConfig.put("introspection_endpoint", introEndpoint);
       // openIdConnectConfig.put("issuer", issuer);
        openIdConnectConfig.put("custos_token_endpoint", tokenEndpoint);
      //  openIdConnectConfig.put("end_session_endpoint", sessionEndpoint);
      //  openIdConnectConfig.put("token_introspection_endpoint", introEndpoint);
        openIdConnectConfig.put("userinfo_endpoint", userInfoEndpoint);
       // openIdConnectConfig.put("jwks_uri", jwksUri);
        openIdConnectConfig.put("registration_endpoint", registrationEndpoint);
       // openIdConnectConfig.remove("check_session_iframe");


        return openIdConnectConfig;
    }

    private User getUserInfo(String realmId, String token) throws Exception {
        String openIdConnectUrl = getOpenIDConfigurationUrl(realmId);
        JSONObject openIdConnectConfig = new JSONObject(getFromUrl(openIdConnectUrl, null));
        String userInfoEndPoint = openIdConnectConfig.getString("userinfo_endpoint");
        JSONObject userInfo = new JSONObject(getFromUrl(userInfoEndPoint, token));
        return new User(userInfo.getString("sub"),
                userInfo.getString("name"),
                userInfo.getString("given_name"),
                userInfo.getString("family_name"),
                userInfo.getString("email"),
                userInfo.getString("preferred_username"));
    }


    private String getOpenIDConfigurationUrl(String realm) {
        return idpServerURL + "realms/" + realm + "/.well-known/openid-configuration";
    }

    private String getFromUrl(String urlToRead, String token) throws Exception {
        StringBuilder result = new StringBuilder();
        URL url = new URL(urlToRead);
        HttpURLConnection conn = (HttpURLConnection) url.openConnection();
        conn.setRequestMethod("GET");
        conn.setDoInput(true);
        if (token != null) {
            String bearerAuth = "Bearer " + token;
            conn.setRequestProperty("Authorization", bearerAuth);
        }
        BufferedReader rd = new BufferedReader(new InputStreamReader(conn.getInputStream()));
        try {
            String line;
            while ((line = rd.readLine()) != null) {
                result.append(line);
            }
        } finally {
            rd.close();
        }
        return result.toString();
    }


    private JSONObject getClientCredentials(String tokenURL, String clientId, String clientSecret) {

        CloseableHttpClient httpClient = HttpClients.createSystem();

        HttpPost httpPost = new HttpPost(tokenURL);
        String encoded = Base64.getEncoder().encodeToString((clientId + ":" + clientSecret).getBytes(StandardCharsets.UTF_8));
        httpPost.setHeader(HttpHeaders.AUTHORIZATION, "Basic " + encoded);
        List<NameValuePair> formParams = new ArrayList<>();
        formParams.add(new BasicNameValuePair("grant_type", "client_credentials"));
        UrlEncodedFormEntity entity = new UrlEncodedFormEntity(formParams, Consts.UTF_8);
        httpPost.setEntity(entity);
        try {
            CloseableHttpResponse response = httpClient.execute(httpPost);
            try {
                String responseBody = EntityUtils.toString(response.getEntity());
                JSONObject tokenInfo = new JSONObject(responseBody);
                return tokenInfo;
            } finally {
                response.close();
            }
        } catch (IOException | JSONException e) {
            throw new RuntimeException(e);
        } finally {
            try {
                httpClient.close();
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
        }
    }

    private JSONObject getTokenFromOAuthCode(String tokenURL, String clientId, String clientSecret, String code,
                                             String redirect_uri) {

        CloseableHttpClient httpClient = HttpClients.createSystem();

        HttpPost httpPost = new HttpPost(tokenURL);
        String encoded = Base64.getEncoder().encodeToString((clientId + ":" + clientSecret).getBytes(StandardCharsets.UTF_8));
        httpPost.setHeader(HttpHeaders.AUTHORIZATION, "Basic " + encoded);
        List<NameValuePair> formParams = new ArrayList<>();
        formParams.add(new BasicNameValuePair("grant_type", "authorization_code"));
        formParams.add(new BasicNameValuePair("code", code));
        formParams.add(new BasicNameValuePair("redirect_uri", redirect_uri));
        formParams.add(new BasicNameValuePair("client_id", clientId));
        formParams.add(new BasicNameValuePair("client_secret", clientSecret));
        UrlEncodedFormEntity entity = new UrlEncodedFormEntity(formParams, Consts.UTF_8);
        httpPost.setEntity(entity);
        try {
            CloseableHttpResponse response = httpClient.execute(httpPost);
            try {
                String responseBody = EntityUtils.toString(response.getEntity());
                JSONObject tokenInfo = new JSONObject(responseBody);
                return tokenInfo;
            } finally {
                response.close();
            }
        } catch (IOException | JSONException e) {
            throw new RuntimeException(e);
        } finally {
            try {
                httpClient.close();
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
        }
    }


    private JSONObject getTokenFromPasswordType(String tokenURL, String clientId, String clientSecret, String username,
                                                 String password) {

        CloseableHttpClient httpClient = HttpClients.createSystem();

        HttpPost httpPost = new HttpPost(tokenURL);
        String encoded = Base64.getEncoder().encodeToString((clientId + ":" + clientSecret).getBytes(StandardCharsets.UTF_8));
        httpPost.setHeader(HttpHeaders.AUTHORIZATION, "Basic " + encoded);
        List<NameValuePair> formParams = new ArrayList<>();
        formParams.add(new BasicNameValuePair("grant_type", "password"));
        formParams.add(new BasicNameValuePair("username", username));
        formParams.add(new BasicNameValuePair("password", password));
        formParams.add(new BasicNameValuePair("client_id", clientId));
        formParams.add(new BasicNameValuePair("client_secret", clientSecret));
        UrlEncodedFormEntity entity = new UrlEncodedFormEntity(formParams, Consts.UTF_8);
        httpPost.setEntity(entity);
        try {
            CloseableHttpResponse response = httpClient.execute(httpPost);
            try {
                String responseBody = EntityUtils.toString(response.getEntity());
                JSONObject tokenInfo = new JSONObject(responseBody);
                return tokenInfo;
            } finally {
                response.close();
            }
        } catch (IOException | JSONException e) {
            throw new RuntimeException(e);
        } finally {
            try {
                httpClient.close();
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
        }
    }


    private JSONObject getTokenFromRefreshToken(String tokenURL, String clientId, String clientSecret, String refreshToken) {

        CloseableHttpClient httpClient = HttpClients.createSystem();

        HttpPost httpPost = new HttpPost(tokenURL);
        String encoded = Base64.getEncoder().encodeToString((clientId + ":" + clientSecret).getBytes(StandardCharsets.UTF_8));
        httpPost.setHeader(HttpHeaders.AUTHORIZATION, "Basic " + encoded);
        List<NameValuePair> formParams = new ArrayList<>();
        formParams.add(new BasicNameValuePair("grant_type", "refresh_token"));
        formParams.add(new BasicNameValuePair("refresh_token", refreshToken));
        formParams.add(new BasicNameValuePair("client_id", clientId));
        formParams.add(new BasicNameValuePair("client_secret", clientSecret));
        UrlEncodedFormEntity entity = new UrlEncodedFormEntity(formParams, Consts.UTF_8);
        httpPost.setEntity(entity);
        try {
            CloseableHttpResponse response = httpClient.execute(httpPost);
            try {
                String responseBody = EntityUtils.toString(response.getEntity());
                JSONObject tokenInfo = new JSONObject(responseBody);
                return tokenInfo;
            } finally {
                response.close();
            }
        } catch (IOException | JSONException e) {
            throw new RuntimeException(e);
        } finally {
            try {
                httpClient.close();
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
        }
    }


}
