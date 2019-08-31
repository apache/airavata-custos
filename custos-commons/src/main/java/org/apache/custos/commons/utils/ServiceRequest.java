package org.apache.custos.commons.utils;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.springframework.stereotype.Component;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;
import java.util.Map;

@Component
public class ServiceRequest {

    public String httpGet(String baseUrl, String requestUrl) throws Exception{
            StringBuilder result = new StringBuilder();
            URL url = new URL(baseUrl + requestUrl);
            HttpURLConnection conn = (HttpURLConnection) url.openConnection();
            conn.setRequestMethod("GET");
            BufferedReader rd = new BufferedReader(new InputStreamReader(conn.getInputStream()));
            String line;
            while ((line = rd.readLine()) != null) {
                result.append(line);
            }
            rd.close();
            return result.toString();
    }

    public void httpPost(String baseUrl, String requestUrl, Object jsonObject){
//        ObjectMapper obj = new ObjectMapper();
//        String postBodyString = obj.writeValueAsString(jsonObject);
//        StringEntity requestEntity = new StringEntity(postBodyString, ContentType.APPLICATION_JSON);
    }
}
