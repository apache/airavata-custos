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

package org.apache.custos.service.federated.client.keycloak.auth;

import org.apache.commons.lang3.StringUtils;
import org.apache.custos.service.federated.client.keycloak.KeycloakUtils;
import org.apache.http.Consts;
import org.apache.http.HttpHeaders;
import org.apache.http.NameValuePair;
import org.apache.http.client.entity.UrlEncodedFormEntity;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpGet;
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
import java.util.ArrayList;
import java.util.Base64;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Acting as a broker between keycloak server and auth services
 */
@Component
public class KeycloakAuthClient {

    private static final Logger LOGGER = LoggerFactory.getLogger(KeycloakAuthClient.class);

    @Value("${iam.server.url}")
    private String idpServerURL;

    @EventListener(ApplicationReadyEvent.class)
    public void initializeSecurity() throws NoSuchAlgorithmException, KeyStoreException, KeyManagementException {
        try {
            LOGGER.debug("initializing security requirements");
            KeycloakUtils.initializeTrustStoreManager();
        } catch (Exception ex) {
            LOGGER.error("Keycloak Authclient initialization failed ", ex);
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

            Configuration configuration = new Configuration(idpServerURL, realmId, clientId, clientCredentials, httpclient);
            AuthzClient keycloakClient = AuthzClient.create(configuration);

            AccessTokenResponse accessToken = keycloakClient.obtainAccessToken(username, password);

            return accessToken != null ? accessToken.getToken() : null;

        } catch (Exception e) {
            LOGGER.error("Error occurred while authenticating", e);
            throw new RuntimeException("Error occurred while authenticating", e);
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
            String msg = "Error occurred while validating if user: " + username + " is authorized in tenant: " + realmId;
            LOGGER.error(msg, e);
            throw new RuntimeException(msg, e);
        }
    }

    public User getUser(String accessToken, String realmId) {
        try {
            return getUserInfo(realmId, accessToken);
        } catch (Exception e) {
            LOGGER.error("Error occurred while retrieving user info", e);
            throw new RuntimeException("Error occurred while retrieving user info", e);
        }
    }

    public String getUserManagementServiceAccountAccessToken(String clientId, String clientSecret, String realmId) {
        try {
            String tokenURL = getTokenEndpoint(realmId);
            LOGGER.info("token url:" + tokenURL);
            JSONObject clientCredentials = getClientCredentials(tokenURL, clientId, clientSecret);
            return clientCredentials.getString("access_token");
        } catch (Exception e) {
            LOGGER.error("Error occurred while retrieving service account access token", e);
            throw new RuntimeException("Error occurred while retrieving service account access token", e);
        }
    }

    public JSONObject getAccessToken(String clientId, String clientSecret, String realmId, String code,
                                     String redirectUri, String codeVerifier) throws JSONException {
        try {
            String tokenURL = getTokenEndpoint(realmId);
            return getTokenFromOAuthCode(tokenURL, clientId, clientSecret, code, redirectUri, codeVerifier);

        } catch (Exception e) {
            LOGGER.error("Error occurred while retrieving the access token", e);
            throw new RuntimeException("Error occurred while retrieving the access token", e);
        }

    }

    public JSONObject tokenIntrospection(String clientId, String clientSecret, String realmId, String token) {
        try {
            String introspectionURL = getTokenIntrospectionEndpoint(realmId);

            HttpPost httpPost = new HttpPost(introspectionURL);
            String encoded = Base64.getEncoder().encodeToString((clientId + ":" + clientSecret).getBytes(StandardCharsets.UTF_8));
            httpPost.setHeader(HttpHeaders.AUTHORIZATION, "Basic " + encoded);
            List<NameValuePair> formParams = new ArrayList<>();
            formParams.add(new BasicNameValuePair("client_id", clientId));
            formParams.add(new BasicNameValuePair("client_secret", clientSecret));
            formParams.add(new BasicNameValuePair("token", token));
            UrlEncodedFormEntity entity = new UrlEncodedFormEntity(formParams, Consts.UTF_8);
            httpPost.setEntity(entity);

            try (CloseableHttpClient httpClient = HttpClients.createSystem();
                 CloseableHttpResponse response = httpClient.execute(httpPost)) {
                return new JSONObject(EntityUtils.toString(response.getEntity()));

            } catch (IOException | JSONException e) {
                LOGGER.error("Error while extracting the token from the OAuth Code", e);
                throw new RuntimeException("Error while extracting the token from the OAuth Code", e);
            }

        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public JSONObject getAccessTokenFromPasswordGrantType(String clientId, String clientSecret, String realmId,
                                                          String username, String password) {
        try {
            String tokenURL = getTokenEndpoint(realmId);
            return getTokenFromPasswordType(tokenURL, clientId, clientSecret, username, password);

        } catch (Exception e) {
            LOGGER.error("Error occurred while retrieving the access token", e);
            throw new RuntimeException("Error occurred while retrieving the access token", e);
        }
    }

    public JSONObject getAccessTokenFromRefreshTokenGrantType(String clientId, String clientSecret, String realmId, String refreshToken) {
        try {
            String tokenURL = getTokenEndpoint(realmId);
            return getTokenFromRefreshToken(tokenURL, clientId, clientSecret, refreshToken);

        } catch (Exception e) {
            String msg = "Error occurred while retrieving the access token";
            LOGGER.error(msg, e);
            throw new RuntimeException(msg, e);
        }
    }

    public JSONObject getAccessTokenFromClientCredentialsGrantType(String clientId, String clientSecret, String realmId) {
        try {
            String tokenURL = getTokenEndpoint(realmId);
            return getClientCredentials(tokenURL, clientId, clientSecret);

        } catch (Exception e) {
            LOGGER.error("Error occurred while retrieving the access token", e);
            throw new RuntimeException("Error occurred while retrieving the access token", e);
        }
    }

    public JSONObject getJWTVerificationCerts(String clientId, String clientSecret, String realmId) {
        try {
            String tokenURL = getJwksUri(realmId);
            return getJWKSResponse(tokenURL, clientId, clientSecret);

        } catch (Exception e) {
            LOGGER.error("Error occurred while retrieving the access token", e);
            throw new RuntimeException("Error occurred while retrieving the access token", e);
        }

    }

    public boolean revokeRefreshToken(String clientId, String clientSecret, String realmId, String refreshToken) {
        try {
            String tokenURL = getEndSessionEndpoint(realmId);
            endSession(tokenURL, clientId, clientSecret, refreshToken);
            return true;

        } catch (Exception e) {
            LOGGER.error("Error occurred while revoking the refresh token", e);
            throw new RuntimeException("Error occurred while revoking the refresh token", e);
        }
    }

    private String getTokenEndpoint(String realmId) throws Exception {
        String openIdConnectUrl = getOpenIDConfigurationUrl(realmId);
        JSONObject openIdConnectConfig = new JSONObject(getFromUrl(openIdConnectUrl, null));
        return openIdConnectConfig.getString("token_endpoint");
    }

    private String getJwksUri(String realmId) throws Exception {
        String openIdConnectUrl = getOpenIDConfigurationUrl(realmId);
        JSONObject openIdConnectConfig = new JSONObject(getFromUrl(openIdConnectUrl, null));
        return openIdConnectConfig.getString("jwks_uri");
    }

    public String getAuthorizationEndpoint(String realmId) throws Exception {
        String openIdConnectUrl = getOpenIDConfigurationUrl(realmId);
        JSONObject openIdConnectConfig = new JSONObject(getFromUrl(openIdConnectUrl, null));
        return openIdConnectConfig.getString("authorization_endpoint");
    }

    public String getEndSessionEndpoint(String realmId) throws Exception {
        String openIdConnectUrl = getOpenIDConfigurationUrl(realmId);
        JSONObject openIdConnectConfig = new JSONObject(getFromUrl(openIdConnectUrl, null));
        return openIdConnectConfig.getString("end_session_endpoint");
    }

    public String getTokenIntrospectionEndpoint(String realmId) throws Exception {
        String openIdConnectUrl = getOpenIDConfigurationUrl(realmId);
        JSONObject openIdConnectConfig = new JSONObject(getFromUrl(openIdConnectUrl, null));
        return openIdConnectConfig.getString("introspection_endpoint");
    }

    public JSONObject getOIDCConfiguration(String tenantId) throws Exception {
        String openIdConnectUrl = getOpenIDConfigurationUrl(tenantId);
        return new JSONObject(getFromUrl(openIdConnectUrl, null));
    }

    private User getUserInfo(String realmId, String token) throws Exception {
        String openIdConnectUrl = getOpenIDConfigurationUrl(realmId);
        JSONObject openIdConnectConfig = new JSONObject(getFromUrl(openIdConnectUrl, null));
        String userInfoEndPoint = openIdConnectConfig.getString("userinfo_endpoint");
        JSONObject userInfo = new JSONObject(getFromUrl(userInfoEndPoint, token));
        return new User(userInfo.getString("sub"),
                userInfo.has("name") ? userInfo.getString("name") : "",
                userInfo.has("given_name") ? userInfo.getString("given_name") : "",
                userInfo.has("family_name") ? userInfo.getString("family_name") : "",
                userInfo.has("email") ? userInfo.getString("email") : "",
                userInfo.getString("preferred_username"));
    }

    private String getOpenIDConfigurationUrl(String realm) {
        LOGGER.debug("Connecting to " + idpServerURL);
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
        try (BufferedReader rd = new BufferedReader(new InputStreamReader(conn.getInputStream()))) {
            String line;
            while ((line = rd.readLine()) != null) {
                result.append(line);
            }
        }

        return result.toString();
    }


    private JSONObject getClientCredentials(String tokenURL, String clientId, String clientSecret) {
        HttpPost httpPost = new HttpPost(tokenURL);
        String encoded = Base64.getEncoder().encodeToString((clientId + ":" + clientSecret).getBytes(StandardCharsets.UTF_8));
        httpPost.setHeader(HttpHeaders.AUTHORIZATION, "Basic " + encoded);
        List<NameValuePair> formParams = new ArrayList<>();
        formParams.add(new BasicNameValuePair("grant_type", "client_credentials"));
        UrlEncodedFormEntity entity = new UrlEncodedFormEntity(formParams, Consts.UTF_8);
        httpPost.setEntity(entity);

        try (CloseableHttpClient httpClient = HttpClients.createSystem();
             CloseableHttpResponse response = httpClient.execute(httpPost)) {
            String responseBody = EntityUtils.toString(response.getEntity());
            return new JSONObject(responseBody);

        } catch (IOException | JSONException e) {
            LOGGER.error("Error while extracting the Client credentials", e);
            throw new RuntimeException(e);
        }
    }

    private JSONObject getTokenFromOAuthCode(String tokenURL, String clientId, String clientSecret, String code, String redirect_uri, String codeVerifier) {
        HttpPost httpPost = new HttpPost(tokenURL);
        List<NameValuePair> formParams = new ArrayList<>();
        formParams.add(new BasicNameValuePair("grant_type", "authorization_code"));
        formParams.add(new BasicNameValuePair("code", code));
        formParams.add(new BasicNameValuePair("redirect_uri", redirect_uri));
        formParams.add(new BasicNameValuePair("client_id", clientId));

        if (StringUtils.isNotBlank(codeVerifier)) {
            formParams.add(new BasicNameValuePair("code_verifier", codeVerifier));
        } else {
            String encoded = Base64.getEncoder().encodeToString((clientId + ":" + clientSecret).getBytes(StandardCharsets.UTF_8));
            httpPost.setHeader(HttpHeaders.AUTHORIZATION, "Basic " + encoded);

            formParams.add(new BasicNameValuePair("client_secret", clientSecret));
        }

        UrlEncodedFormEntity entity = new UrlEncodedFormEntity(formParams, Consts.UTF_8);
        httpPost.setEntity(entity);

        try (CloseableHttpClient httpClient = HttpClients.createSystem();
             CloseableHttpResponse response = httpClient.execute(httpPost)) {
            String responseBody = EntityUtils.toString(response.getEntity());
            return new JSONObject(responseBody);

        } catch (IOException | JSONException e) {
            LOGGER.error("Error while extracting the token from the OAuth Code", e);
            throw new RuntimeException("Error while extracting the token from the OAuth Code", e);
        }
    }

    private void endSession(String endSessionEndpoint, String clientId, String clientSecret, String refreshToken) {
        HttpPost httpPost = new HttpPost(endSessionEndpoint);
        String encoded = Base64.getEncoder().encodeToString((clientId + ":" + clientSecret).getBytes(StandardCharsets.UTF_8));
        httpPost.setHeader(HttpHeaders.AUTHORIZATION, "Basic " + encoded);
        List<NameValuePair> formParams = new ArrayList<>();
        formParams.add(new BasicNameValuePair("refresh_token", refreshToken));
        formParams.add(new BasicNameValuePair("client_id", clientId));
        formParams.add(new BasicNameValuePair("client_secret", clientSecret));
        UrlEncodedFormEntity entity = new UrlEncodedFormEntity(formParams, Consts.UTF_8);
        httpPost.setEntity(entity);

        try (CloseableHttpClient httpClient = HttpClients.createSystem();
             CloseableHttpResponse response = httpClient.execute(httpPost)) {
            if (response.getStatusLine().getStatusCode() != 204) {
                throw new IllegalStateException("Failed to end session properly: " + EntityUtils.toString(response.getEntity()));
            }
        } catch (Exception e) {
            LOGGER.error("Error while ending the session", e);
            throw new RuntimeException("Error while ending the session", e);
        }
    }

    private JSONObject getJWKSResponse(String jwksUri, String clientId, String clientSecret) {
        HttpGet httpPost = new HttpGet(jwksUri);
        String encoded = Base64.getEncoder().encodeToString((clientId + ":" + clientSecret).getBytes(StandardCharsets.UTF_8));
        httpPost.setHeader(HttpHeaders.AUTHORIZATION, "Basic " + encoded);

        try (CloseableHttpClient httpClient = HttpClients.createSystem();
             CloseableHttpResponse response = httpClient.execute(httpPost)) {

            String responseBody = EntityUtils.toString(response.getEntity());
            return new JSONObject(responseBody);

        } catch (IOException | JSONException e) {
            LOGGER.error("Error while retrieving the JWKS response", e);
            throw new RuntimeException("Error while retrieving the JWKS response", e);
        }
    }


    private JSONObject getTokenFromPasswordType(String tokenURL, String clientId, String clientSecret, String username, String password) {
        HttpPost httpPost = new HttpPost(tokenURL);
        String encoded = Base64.getEncoder().encodeToString((clientId + ":" + clientSecret).getBytes(StandardCharsets.UTF_8));
        httpPost.setHeader(HttpHeaders.AUTHORIZATION, "Basic " + encoded);
        List<NameValuePair> formParams = new ArrayList<>();
        formParams.add(new BasicNameValuePair("grant_type", "password"));
        formParams.add(new BasicNameValuePair("username", username));
        formParams.add(new BasicNameValuePair("password", password));
        formParams.add(new BasicNameValuePair("client_id", clientId));
        formParams.add(new BasicNameValuePair("client_secret", clientSecret));
        formParams.add(new BasicNameValuePair("scope", "openid"));
        UrlEncodedFormEntity entity = new UrlEncodedFormEntity(formParams, Consts.UTF_8);
        httpPost.setEntity(entity);

        try (CloseableHttpClient httpClient = HttpClients.createSystem();
             CloseableHttpResponse response = httpClient.execute(httpPost)) {

            String responseBody = EntityUtils.toString(response.getEntity());
            return new JSONObject(responseBody);

        } catch (IOException | JSONException e) {
            LOGGER.error("Error while extracting the token from the username/password", e);
            throw new RuntimeException("Error while extracting the token from the username/password", e);
        }
    }


    private JSONObject getTokenFromRefreshToken(String tokenURL, String clientId, String clientSecret, String refreshToken) {
        HttpPost httpPost = new HttpPost(tokenURL);
        String encoded = Base64.getEncoder().encodeToString((clientId + ":" + clientSecret).getBytes(StandardCharsets.UTF_8));
        httpPost.setHeader(HttpHeaders.AUTHORIZATION, "Basic " + encoded);
        List<NameValuePair> formParams = new ArrayList<>();
        formParams.add(new BasicNameValuePair("grant_type", "refresh_token"));
        formParams.add(new BasicNameValuePair("refresh_token", refreshToken));
        formParams.add(new BasicNameValuePair("client_id", clientId));
        formParams.add(new BasicNameValuePair("client_secret", clientSecret));
        formParams.add(new BasicNameValuePair("scope", "openid"));
        UrlEncodedFormEntity entity = new UrlEncodedFormEntity(formParams, Consts.UTF_8);
        httpPost.setEntity(entity);
        try (CloseableHttpClient httpClient = HttpClients.createSystem();
             CloseableHttpResponse response = httpClient.execute(httpPost)) {

            String responseBody = EntityUtils.toString(response.getEntity());
            return new JSONObject(responseBody);

        } catch (IOException | JSONException e) {
            LOGGER.error("Error while extracting the token from the refresh token", e);
            throw new RuntimeException("Error while extracting the token from the refresh token", e);
        }
    }
}
