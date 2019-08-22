package org.apache.custos.authentication.service.sample;

import org.apache.custos.authentication.cpi.CustosAuthenticationService;
import org.apache.custos.client.authentication.service.AuthenticationServiceClientFactory;
import org.apache.custos.commons.exceptions.CustosSecurityException;
import org.apache.custos.commons.model.security.AuthzToken;
import org.apache.custos.commons.utils.Constants;
import org.apache.custos.security.manager.KeyCloakSecurityManager;
import org.apache.http.Consts;
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

import java.io.IOException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;

public class CustosAuthenticationServiceSample {

    private static final Logger log = LoggerFactory.getLogger(CustosAuthenticationServiceSample.class);
    //present in keycloak json configuration
    public static String USERNAME = "default-admin";
    public static String GATEWAY_ID = "default";

    public static void main(String[] args) throws Exception{

        try {
            //change this based on the port on which the server is started
            String serverHost = "localhost";
            int serverPort = 9091;
            CustosAuthenticationService.Client client = AuthenticationServiceClientFactory.createAuthenticationServiceClient(serverHost, serverPort);
            testUserAuthentication(client);
        }catch (TException e) {
            throw new Exception(e);
        }catch (Exception e){
            throw new Exception("Error setting up the authentication server", e);
        }
    }

    public static void testUserAuthentication(CustosAuthenticationService.Client client) throws TException, CustosSecurityException {
        boolean authenticated = false;
        new KeyCloakSecurityManager().initializeSecurityInfra();
        AuthzToken authzToken = new AuthzToken();
        HashMap<String, String> map_ = new HashMap<>();
        map_.put(Constants.GATEWAY_ID, GATEWAY_ID);
        map_.put(Constants.USER_NAME, USERNAME);
        authzToken.setClaimsMap(map_);
        JSONObject json = getAccessToken();
        if(json.has("access_token")){
            String access_token = getAccessToken().get("access_token").toString();
            authzToken.setAccessToken(access_token);
            authenticated = client.isUserAuthenticated(authzToken);
        }

        assert (authenticated): "User authenticating failed";
        if(authenticated){
            log.info("User successfully authenticated");
        }

    }

    private static JSONObject getAccessToken() {
        String password = "123456";
        String tokenEndPoint = "https://airavata.host:8443/auth/realms/default/protocol/openid-connect/token";
        String grant_type = "password";
        String client_id = "admin-cli";

        CloseableHttpClient httpClient = HttpClients.createSystem();
        HttpPost httpPost = new HttpPost(tokenEndPoint);
        ;
        List<NameValuePair> formParams = new ArrayList<>();
        formParams.add(new BasicNameValuePair("grant_type", "password"));
        formParams.add(new BasicNameValuePair("client_id", client_id));
        formParams.add(new BasicNameValuePair("username", USERNAME));
        formParams.add(new BasicNameValuePair("password", password));
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
}
