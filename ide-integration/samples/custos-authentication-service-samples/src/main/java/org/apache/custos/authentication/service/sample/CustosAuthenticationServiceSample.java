package org.apache.custos.authentication.service.sample;

import org.apache.custos.authentication.cpi.CustosAuthenticationService;
import org.apache.custos.client.authentication.service.AuthenticationServiceClient;
import org.apache.custos.commons.model.security.AuthzToken;
import org.apache.custos.commons.utils.Constants;
import org.apache.custos.commons.utils.ServerSettings;
import org.apache.samples.common.SamplesCommon;
import org.json.JSONObject;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import java.util.HashMap;

public class CustosAuthenticationServiceSample {

    private static final Logger log = LoggerFactory.getLogger(CustosAuthenticationServiceSample.class);
    //present in keycloak json configuration
    public static String USERNAME = "default-admin";
    public static String GATEWAY_ID = "default";
    public static AuthzToken authzToken;

    public static void main(String[] args){

        try {
            String serverHost = ServerSettings.getAuthenticationServerHost();
            int serverPort = Integer.parseInt(ServerSettings.getAuthenticationServerPort());
            CustosAuthenticationService.Client client = AuthenticationServiceClient.createAuthenticationServiceClient(serverHost, serverPort);
            testUserAuthentication(client);
        }catch (Exception e){
           log.error("Error setting up the authentication server");
        }
    }

    public static void testUserAuthentication(CustosAuthenticationService.Client client){
        try{
            initializeAuthzToken();
            boolean authenticated = client.isUserAuthenticated(authzToken);
            assert (authenticated): "User not authenticated failed. Authentication test failed";
            if(authenticated){
                log.info("User successfully authenticated");
            }
        }catch (Exception e){
            log.error("Test userAuthentication Failed!!", e);
        }
    }
    private static void initializeAuthzToken() throws Exception{
        try{
            authzToken = new AuthzToken();
            HashMap<String, String> map = new HashMap<>();
            map.put(Constants.USER_NAME, USERNAME);
            map.put(Constants.GATEWAY_ID, GATEWAY_ID);
            authzToken.setClaimsMap(map);
            JSONObject json = SamplesCommon.getAccessToken(USERNAME, null, "admin-cli", "password");
            if(json.has("access_token")){
                authzToken.setAccessToken(json.get("access_token").toString());
            }else{
                throw new Exception("Check if the user exists in keycloak");
            }
        }catch (Exception e) {
            throw new Exception("Could not get access token", e);
        }

    }
}
