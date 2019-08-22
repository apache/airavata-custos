package org.apache.samples.common;

import org.apache.custos.commons.exceptions.CustosSecurityException;
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
import org.json.JSONObject;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;

public class SamplesCommon {

    public static JSONObject getAccessToken(String username, String password, String clientId, String grant_type) throws CustosSecurityException {
        new KeyCloakSecurityManager().initializeSecurityInfra();
        if(password == null) {
            //keeping a default password for testing
            password = "123456";
        }
        String tokenEndPoint = "https://localhost:8443/auth/realms/default/protocol/openid-connect/token";

        CloseableHttpClient httpClient = HttpClients.createSystem();
        HttpPost httpPost = new HttpPost(tokenEndPoint);
        ;
        List<NameValuePair> formParams = new ArrayList<>();
        formParams.add(new BasicNameValuePair("grant_type", grant_type));
        formParams.add(new BasicNameValuePair("client_id", clientId));
        formParams.add(new BasicNameValuePair("username", username));
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
