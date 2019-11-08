package org.apache.custos.profile.iam.admin.services.core;

import org.apache.custos.commons.exceptions.ApplicationSettingsException;
import org.apache.custos.commons.utils.ServerSettings;
import org.apache.custos.profile.iam.admin.services.cpi.exception.IamAdminServicesException;
import org.apache.http.Consts;
import org.apache.http.HttpHeaders;
import org.apache.http.NameValuePair;
import org.apache.http.client.entity.UrlEncodedFormEntity;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;
import org.json.JSONArray;
import org.json.JSONObject;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.util.Base64;
import java.util.List;

public class CILogonManagementImpl {

    private static String CLIENT_NAME = "client_name";
    private static String REDIRECT_URIS = "redirect_uris";
    private static String COMMENTS = "comments";



    public static JSONObject create(String clientId, String clientSecret, String client_name, List<String> redirect_uris) throws ApplicationSettingsException, IamAdminServicesException {
        CloseableHttpClient httpClient = HttpClients.createSystem();

        HttpPost httpPost = new HttpPost(ServerSettings.getCILogonURL());
        String encoded = Base64.getEncoder().encodeToString((clientId+":"+clientSecret).getBytes(StandardCharsets.UTF_8));
        httpPost.setHeader(HttpHeaders.AUTHORIZATION, "Bearer " + encoded);
        JSONObject data = new JSONObject();
        data.put(CLIENT_NAME, client_name);
        data.put(REDIRECT_URIS, new JSONArray(redirect_uris));
        data.put(COMMENTS, "Created by custos");
        httpPost.setEntity(new StringEntity(data.toString(), ContentType.APPLICATION_JSON));
        try {
            CloseableHttpResponse response = httpClient.execute(httpPost);
            if(response.getStatusLine().getStatusCode() < 200 ||  response.getStatusLine().getStatusCode() > 299) throw new IamAdminServicesException("Could not create a cilogon client");
            try {
                JSONObject tokenInfo = new JSONObject(EntityUtils.toString(response.getEntity()));
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

    //TODO
    public static JSONObject get(String clientId, String clientSecret, String uri) throws ApplicationSettingsException{
        CloseableHttpClient httpClient = HttpClients.createSystem();
        HttpGet httpGet = new HttpGet(ServerSettings.getCILogonURL());
        String encoded = Base64.getEncoder().encodeToString((clientId+":"+clientSecret).getBytes(StandardCharsets.UTF_8));
        httpGet.setHeader(HttpHeaders.AUTHORIZATION, "Bearer " + encoded);
        return null;
    }
}

