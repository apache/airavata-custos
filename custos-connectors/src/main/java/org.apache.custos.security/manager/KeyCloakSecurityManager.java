/*
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
*/
package org.apache.custos.security.manager;

import org.apache.custos.commons.exceptions.ApplicationSettingsException;
import org.apache.custos.commons.exceptions.CustosSecurityException;
import org.apache.custos.commons.model.security.AuthzToken;
import org.apache.custos.commons.model.security.UserInfo;
import org.apache.custos.commons.utils.Constants;
import org.apache.custos.commons.utils.ServerSettings;
import org.apache.custos.commons.utils.ThriftUtils;
import org.apache.custos.profile.client.ProfileServiceClientFactory;
import org.apache.custos.profile.model.workspace.Gateway;
import org.apache.custos.profile.tenant.cpi.TenantProfileService;
import org.apache.custos.profile.tenant.cpi.exception.TenantProfileServiceException;
import org.apache.custos.security.authzcache.*;
import org.apache.custos.security.utils.TrustStoreManager;
import org.apache.http.Consts;
import org.apache.http.HttpHeaders;
import org.apache.http.NameValuePair;
import org.apache.http.client.entity.UrlEncodedFormEntity;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.message.BasicNameValuePair;
import org.apache.http.util.EntityUtils;
import org.apache.thrift.TException;
import org.json.JSONObject;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Base64;
import java.util.List;

public class KeyCloakSecurityManager implements CustosSecurityManager {
    private final static Logger logger = LoggerFactory.getLogger(KeyCloakSecurityManager.class);
    private TenantProfileService.Client tenantProfileClient = null;
    public KeyCloakSecurityManager() throws CustosSecurityException {
        initializeSecurityInfra();
    }

    /**
     * Implement this method in your SecurityManager to perform necessary initializations at the server startup.
     *
     * @throws CustosSecurityException
     */
    @Override
    public void initializeSecurityInfra() throws CustosSecurityException {
        try {
            //initialize SSL context with the trust store that contains the public cert of WSO2 Identity Server.
            TrustStoreManager trustStoreManager = new TrustStoreManager();
            trustStoreManager.initializeTrustStoreManager(ServerSettings.getTrustStorePath(),
                    ServerSettings.getTrustStorePassword());
        } catch (Exception e) {
            throw new CustosSecurityException(e.getMessage(), e);
        }

    }

    /**
     * Implement this method with the user authentication logic in your SecurityManager.
     *
     * @param authzToken : this includes OAuth token and user's claims
     * @return
     * @throws CustosSecurityException
     */
    @Override
    public boolean isUserAuthenticated(AuthzToken authzToken) throws CustosSecurityException {
        String subject = authzToken.getClaimsMap().get(Constants.USER_NAME);
        String accessToken = authzToken.getAccessToken();
        String gatewayId = authzToken.getClaimsMap().get(Constants.GATEWAY_ID);
        try {
            if (!ServerSettings.isAPISecured()) {
                return true;
            }
            initServiceClients();

            if (ServerSettings.isAuthzCacheEnabled()) {
                //obtain an instance of AuthzCacheManager implementation.
                AuthzCacheManager authzCacheManager = AuthzCacheManagerFactory.getAuthzCacheManager();

                //check in the cache
                AuthzCachedStatus authzCachedStatus = authzCacheManager.getAuthzCachedStatus(
                        new AuthzCacheIndex(subject, gatewayId, accessToken));

                if (AuthzCachedStatus.AUTHORIZED.equals(authzCachedStatus)) {
                    logger.debug("Authz decision for: (" + subject + ", " + accessToken + ") is retrieved from cache.");
                    return true;
                } else if (AuthzCachedStatus.NOT_AUTHORIZED.equals(authzCachedStatus)) {
                    logger.debug("Authz decision for: (" + subject + ", " + accessToken + ") is retrieved from cache.");
                    return false;
                } else if (AuthzCachedStatus.NOT_CACHED.equals(authzCachedStatus)) {
                    logger.debug("Authz decision for: (" + subject + ", " + accessToken + ") is not in the cache. " +
                            "Generating decision based on group membership.");
//                    GatewayGroupMembership gatewayGroupMembership = getGatewayGroupMembership(subject, accessToken, gatewayId);
//                    boolean authorizationDecision = hasPermission(gatewayGroupMembership, action);
                    boolean authorizationDecision = validateToken(subject, accessToken, gatewayId);
                    //cache the authorization decision
                    long currentTime = System.currentTimeMillis();
                    //TODO get the actual token expiration time
                    authzCacheManager.addToAuthzCache(new AuthzCacheIndex(subject, gatewayId, accessToken),
                            new AuthzCacheEntry(authorizationDecision, currentTime + 1000 * 60 * 60, currentTime));
                    return authorizationDecision;
                } else {
                    //undefined status returned from the authz cache manager
                    throw new CustosSecurityException("Error in reading from the authorization cache.");
                }
            } else {
//                GatewayGroupMembership gatewayGroupMembership = getGatewayGroupMembership(subject, accessToken, gatewayId);
//                return hasPermission(gatewayGroupMembership, action);
                return validateToken(subject, accessToken, gatewayId);
            }

        } catch (ApplicationSettingsException e) {
            logger.error("Missing or invalid application setting.", e);
            throw new CustosSecurityException(e.getMessage(), e);
        } catch (Exception e) {
            logger.error("Error occurred while checking if user: " + subject + " is authorized in gateway: " + gatewayId, e);
            throw new CustosSecurityException(e.getMessage(), e);
        }
//        } finally {
//            closeServiceClients();
//        }
    }

    @Override
    public AuthzToken getUserManagementServiceAccountAuthzToken(AuthzToken authzToken, String gatewayId) throws CustosSecurityException {
        try {
            initServiceClients();
            Gateway gateway = tenantProfileClient.getGatewayUsingGatewayId(authzToken,gatewayId);
            String tokenURL = getTokenEndpoint(gatewayId);
            JSONObject clientCredentials = getClientCredentials(tokenURL, gateway.getOauthClientId(), gateway.getOauthClientSecret());
            String accessToken = clientCredentials.getString("access_token");
            AuthzToken userManagementServiceAccountAuthzToken = new AuthzToken(accessToken);
            userManagementServiceAccountAuthzToken.putToClaimsMap(Constants.GATEWAY_ID, gatewayId);
            userManagementServiceAccountAuthzToken.putToClaimsMap(Constants.USER_NAME, gateway.getOauthClientId());
            return userManagementServiceAccountAuthzToken;
        } catch (Exception e) {
            throw new CustosSecurityException(e);
        } finally {
            closeServiceClients();
        }
    }

    @Override
    public UserInfo getUserInfoFromAuthzToken(AuthzToken authzToken) throws CustosSecurityException {
        try {
//            initServiceClients();
            final String gatewayId = authzToken.getClaimsMap().get(Constants.GATEWAY_ID);
            final String token = authzToken.getAccessToken();
            return getUserInfo(gatewayId, token);
        } catch (Exception e) {
            throw new CustosSecurityException(e);
        }
//        } finally {
//            closeServiceClients();
//        }
    }

    private UserInfo getUserInfo(String gatewayId, String token) throws Exception {
//        GatewayResourceProfile gwrp = registryServiceClient.getGatewayResourceProfile(gatewayId);
//        String identityServerRealm = gwrp.getIdentityServerTenant();
        //TODO: Confirm the difference between gatewayId and IdentityServerTenant, using gatewayId as of now
        String openIdConnectUrl = getOpenIDConfigurationUrl(gatewayId);
        JSONObject openIdConnectConfig = new JSONObject(getFromUrl(openIdConnectUrl, null));
        String userInfoEndPoint = openIdConnectConfig.getString("userinfo_endpoint");
        JSONObject userInfo = new JSONObject(getFromUrl(userInfoEndPoint, token));
        return new UserInfo()
                .setSub(userInfo.getString("sub"))
                .setFullName(userInfo.getString("name"))
                .setFirstName(userInfo.getString("given_name"))
                .setLastName(userInfo.getString("family_name"))
                .setEmailAddress(userInfo.getString("email"))
                .setUsername(userInfo.getString("preferred_username"));
    }

//    private GatewayGroupMembership getGatewayGroupMembership(String username, String token, String gatewayId) throws Exception {
//        validateToken(username, token, gatewayId);
//        GatewayGroups gatewayGroups = getGatewayGroups(gatewayId);
//        List<UserGroup> userGroups = sharingRegistryServiceClient.getAllMemberGroupsForUser(gatewayId, username + "@" + gatewayId);
//        List<String> userGroupIds = userGroups.stream().map(g -> g.getGroupId()).collect(Collectors.toList());
//        GatewayGroupMembership gatewayGroupMembership = new GatewayGroupMembership();
//        gatewayGroupMembership.setInAdminsGroup(userGroupIds.contains(gatewayGroups.getAdminsGroupId()));
//        gatewayGroupMembership.setInReadOnlyAdminsGroup(userGroupIds.contains(gatewayGroups.getReadOnlyAdminsGroupId()));
//        return gatewayGroupMembership;
//    }

//    private GatewayGroups getGatewayGroups(String gatewayId) throws Exception {
//        if (registryServiceClient.isGatewayGroupsExists(gatewayId)) {
//            return registryServiceClient.getGatewayGroups(gatewayId);
//        } else {
//            return GatewayGroupsInitializer.initializeGatewayGroups(gatewayId);
//        }
//    }

    private boolean validateToken(String username, String token, String gatewayId) throws Exception {
        UserInfo userInfo = getUserInfo(gatewayId, token);
        if (!username.equals(userInfo.getUsername())) {
            throw new CustosSecurityException("Subject name and username for the token doesn't match");
        }
        return true;
    }

//    private String[] getUserRolesFromOAuthToken(String username, String token, String gatewayId) throws Exception {
//        GatewayResourceProfile gwrp = getRegistryServiceClient().getGatewayResourceProfile(gatewayId);
//        String identityServerRealm = gwrp.getIdentityServerTenant();
//        String openIdConnectUrl = getOpenIDConfigurationUrl(identityServerRealm);
//        JSONObject openIdConnectConfig = new JSONObject(getFromUrl(openIdConnectUrl, token));
//        String userInfoEndPoint = openIdConnectConfig.getString("userinfo_endpoint");
//        JSONObject userInfo = new JSONObject(getFromUrl(userInfoEndPoint, token));
//        if (!username.equals(userInfo.get("preferred_username"))) {
//            throw new CustosSecurityException("Subject name and username for the token doesn't match");
//        }
//        String userId = userInfo.getString("sub");
//
//        String userRoleMappingUrl = ServerSettings.getRemoteIDPServiceUrl() + "/admin/realms/"
//                + identityServerRealm + "/users/"
//                + userId + "/role-mappings/realm";
//        JSONArray roleMappings = new JSONArray(getFromUrl(userRoleMappingUrl, getAdminAccessToken(gatewayId)));
//        String[] roles = new String[roleMappings.length()];
//        for (int i = 0; i < roleMappings.length(); i++) {
//            roles[i] = (new JSONObject(roleMappings.get(i).toString())).get("name").toString();
//        }
//
//        return roles;
//    }

    private String getOpenIDConfigurationUrl(String realm) throws ApplicationSettingsException {
        return ServerSettings.getRemoteIDPServiceUrl() + "/realms/" + realm + "/.well-known/openid-configuration";
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
        String line;
        while ((line = rd.readLine()) != null) {
            result.append(line);
        }
        rd.close();
        return result.toString();
    }

//    private String getAdminAccessToken(String gatewayId) throws Exception {
//        CredentialStoreService.Client csClient = getCredentialStoreServiceClient();
//        GatewayResourceProfile gwrp = getRegistryServiceClient().getGatewayResourceProfile(gatewayId);
//        String identityServerRealm = gwrp.getIdentityServerTenant();
//        String openIdConnectUrl = getOpenIDConfigurationUrl(identityServerRealm);
//        JSONObject openIdConnectConfig = new JSONObject(getFromUrl(openIdConnectUrl, null));
//        PasswordCredential credential = csClient.getPasswordCredential(gwrp.getIdentityServerPwdCredToken(), gwrp.getGatewayID());
//        String username = credential.getLoginUserName();
//        String password = credential.getPassword();
//        String urlString = openIdConnectConfig.getString("token_endpoint");
//        StringBuilder result = new StringBuilder();
//        URL url = new URL(urlString);
//        HttpURLConnection conn = (HttpURLConnection) url.openConnection();
//        conn.setRequestMethod("POST");
//        conn.setDoOutput(true);
//        String postFields = "client_id=admin-cli&username=" + username + "&password=" + password + "&grant_type=password";
//        conn.getOutputStream().write(postFields.getBytes());
//        BufferedReader rd = new BufferedReader(new InputStreamReader(conn.getInputStream()));
//        String line;
//        while ((line = rd.readLine()) != null) {
//            result.append(line);
//        }
//        rd.close();
//        JSONObject tokenInfo = new JSONObject(result.toString());
//        return tokenInfo.get("access_token").toString();
//    }

    private String getTokenEndpoint(String gatewayId) throws Exception {
        String openIdConnectUrl = getOpenIDConfigurationUrl(gatewayId);
        JSONObject openIdConnectConfig = new JSONObject(getFromUrl(openIdConnectUrl, null));
        return openIdConnectConfig.getString("token_endpoint");
    }

    private JSONObject getClientCredentials(String tokenURL, String clientId, String clientSecret) throws ApplicationSettingsException, CustosSecurityException {

        CloseableHttpClient httpClient = HttpClients.createSystem();

        HttpPost httpPost = new HttpPost(tokenURL);
        String encoded = Base64.getEncoder().encodeToString((clientId+":"+clientSecret).getBytes(StandardCharsets.UTF_8));
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
        } catch (IOException e) {
            throw new RuntimeException(e);
        } finally {
            try {
                httpClient.close();
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
        }
    }

//    private boolean hasPermission(GatewayGroupMembership gatewayGroupMembership, String apiMethod) {
//
//        // Note: as a stopgap solution, until all resources are secured with group-based authorization, map the Admins
//        // and Read Only Admins groups to the corresponding roles
//        final String role;
//        if (gatewayGroupMembership.isInAdminsGroup()) {
//            return true;
//        } else if (gatewayGroupMembership.isInReadOnlyAdminsGroup()) {
//            role = "admin-read-only";
//        } else {
//            // If not in Admins or Read Only Admins groups, treat as a gateway-user
//            role = "gateway-user";
//        }
//        Pattern pattern = Pattern.compile(this.rolePermissionConfig.get(role));
//        Matcher matcher = pattern.matcher(apiMethod);
//        return matcher.matches();
//    }
//
//    private boolean hasPermission(String[] roles, String apiMethod) {
//        for (int i = 0; i < roles.length; i++) {
//            String role = roles[i];
//            if (this.rolePermissionConfig.keySet().contains(role)) {
//                Pattern pattern = Pattern.compile(this.rolePermissionConfig.get(role));
//                Matcher matcher = pattern.matcher(apiMethod);
//                if (matcher.matches())
//                    return true;
//            }
//        }
//        return false;
//    }
//
    private void initServiceClients() throws TException, ApplicationSettingsException {
        tenantProfileClient = getTenantProfileServiceClient();
    }

    private void closeServiceClients() {
        if (tenantProfileClient != null) {
            ThriftUtils.close(tenantProfileClient);
        }
    }

//    private RegistryService.Client getRegistryServiceClient() throws TException, ApplicationSettingsException {
//        final int serverPort = Integer.parseInt(ServerSettings.getRegistryServerPort());
//        final String serverHost = ServerSettings.getRegistryServerHost();
//        try {
//            return RegistryServiceClientFactory.createRegistryClient(serverHost, serverPort);
//        } catch (RegistryServiceException e) {
//            throw new TException("Unable to create registry client...", e);
//        }
//    }
//
//    private CredentialStoreService.Client getCredentialStoreServiceClient() throws TException, ApplicationSettingsException {
//        final int serverPort = Integer.parseInt(ServerSettings.getCredentialStoreServerPort());
//        final String serverHost = ServerSettings.getCredentialStoreServerHost();
//        try {
//            return CredentialStoreClientFactory.createAiravataCSClient(serverHost, serverPort);
//        } catch (CredentialStoreException e) {
//            throw new TException("Unable to create credential store client...", e);
//        }
//    }
//
//    private SharingRegistryService.Client getSharingRegistryServiceClient() throws TException, ApplicationSettingsException {
//        final int serverPort = Integer.parseInt(ServerSettings.getSharingRegistryPort());
//        final String serverHost = ServerSettings.getSharingRegistryHost();
//        try {
//            return SharingRegistryServiceClientFactory.createSharingRegistryClient(serverHost, serverPort);
//        } catch (SharingRegistryException e) {
//            throw new TException("Unable to create sharing registry client...", e);
//        }
//    }

//    public static void main(String[] args) throws CustosSecurityException, ApplicationSettingsException {
//        ServerSettings.setSetting("trust.store", "./modules/configuration/server/src/main/resources/client_truststore.jks");
//        ServerSettings.setSetting("trust.store.password", "airavata");
//        KeyCloakSecurityManager keyCloakSecurityManager = new KeyCloakSecurityManager();
//        final String tokenURL = "...";
//        final String clientId = "...";
//        final String clientSecret = "...";
//        JSONObject jsonObject = keyCloakSecurityManager.getClientCredentials(tokenURL, clientId, clientSecret);
//        System.out.println("access_token=" + jsonObject.getString("access_token"));
//    }
    private TenantProfileService.Client getTenantProfileServiceClient() throws ApplicationSettingsException , TenantProfileServiceException, TException {
        try {
            TenantProfileService.Client tenantProfileClient = ProfileServiceClientFactory.createTenantProfileServiceClient(ServerSettings.getProfileServiceServerHost(), Integer.parseInt(ServerSettings.getProfileServiceServerPort()));
            return tenantProfileClient;
        } catch (TenantProfileServiceException e) {
            throw new TenantProfileServiceException("Unable to create tenant profile client...");
        }
    }
}
