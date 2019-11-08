package org.apache.custos.rest.controller;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.custos.commons.model.security.AuthzToken;
import org.apache.custos.profile.iam.admin.services.cpi.exception.IamAdminServicesException;
import org.apache.custos.profile.tenant.cpi.TenantProfileService;
import org.apache.custos.rest.resources.CILogonClientInfoResource;
import org.apache.http.HttpHeaders;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;
import org.json.JSONArray;
import org.json.JSONObject;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.ws.rs.PathParam;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Base64;

@RestController
@RequestMapping("/cilogon")
public class CILogonController {

    private static AuthzToken authzToken = new AuthzToken("empy_token");

    private String cILogonAdminClientId = "";
    private String cILogonAdminClientSecret = "";
    private String cILogonAdminUrl = "";

    @Autowired
    private TenantProfileService.Client tenantClient;

    @RequestMapping(value = "client/{tenant}", method = RequestMethod.POST)
    public CILogonClientInfoResource createCILogOnClient(@PathVariable("tenant") String tenantId) throws Exception {
        // Check realm is created is first, is callback urls exist
        boolean gatewayExist = tenantClient.isGatewayExist(authzToken, tenantId);
        if (!gatewayExist) {
            throw new Exception("Tennant " + tenantId + " does not exist");
        }

        CloseableHttpClient httpClient = HttpClients.createSystem();

        HttpPost httpPost = new HttpPost(cILogonAdminUrl);
        String encoded = Base64.getEncoder().encodeToString((cILogonAdminClientId+":"+cILogonAdminClientSecret).getBytes(StandardCharsets.UTF_8));
        httpPost.setHeader(HttpHeaders.AUTHORIZATION, "Bearer " + encoded);
        JSONObject data = new JSONObject();
        data.put("client_name", cILogonAdminClientId);
        data.put("redirect_uris", new JSONArray(new ArrayList<>()));
        data.put("comments", "Created by custos");
        httpPost.setEntity(new StringEntity(data.toString(), ContentType.APPLICATION_JSON));

        try {
            CloseableHttpResponse response = httpClient.execute(httpPost);
            if(response.getStatusLine().getStatusCode() < 200 ||  response.getStatusLine().getStatusCode() > 299) {
                throw new Exception("Could not create a cilogon client");
            }
            try {
                ObjectMapper objectMapper = new ObjectMapper();

                CILogonClientInfoResource ciLogonClientInfoResource = objectMapper.readValue(EntityUtils.toString(response.getEntity()), CILogonClientInfoResource.class);
                return ciLogonClientInfoResource;
            } finally {
                response.close();
            }
        } finally {
            try {
                httpClient.close();
            } catch (IOException e) {
                // Ignore
            }
        }
    }

}
