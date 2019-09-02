package org.apache.custos.commons.utils;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.custos.commons.exceptions.ServiceConnectionException;
import org.apache.http.HttpResponse;
import org.apache.http.HttpStatus;
import org.apache.http.client.HttpClient;
import org.apache.http.client.methods.*;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.HttpClientBuilder;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;
/*
* This class is a helper class.
* It acts as an HTTP client and returns the http response back to the caller
*/

@Component
public class ServiceRequestClient {

    private final static Logger logger = LoggerFactory.getLogger(ServiceRequestClient.class);

    public HttpResponse httpGet(String baseUrl, String requestUrl) throws ServiceConnectionException{

        try {
            HttpClient client = HttpClientBuilder.create().build();
            HttpGet request = new HttpGet(baseUrl + requestUrl);
            HttpResponse response = client.execute(request);
            return response;
        }catch (Exception ex){
            String error = String.format("Could not fulfill request from Url: %s", baseUrl+requestUrl);
            logger.error(error, ex);
            throw new ServiceConnectionException(error, ex);
        }
    }

    public HttpResponse httpPost(String baseUrl, String requestUrl, Object jsonObject) throws ServiceConnectionException{
        try{
            HttpClient client = HttpClientBuilder.create().build();
            HttpPost post = new HttpPost(baseUrl+requestUrl);
            ObjectMapper obj = new ObjectMapper();
            String postBodyString = obj.writeValueAsString(jsonObject);
            StringEntity requestEntity = new StringEntity(postBodyString, ContentType.APPLICATION_JSON);
            post.setEntity(requestEntity);
            HttpResponse response = client.execute(post);
            return response;
    }catch (Exception ex){
            String error = String.format("Could not fulfill request from Url: %s", baseUrl+requestUrl);
            logger.error(error, ex);
            throw new ServiceConnectionException(error, ex);
        }
    }

    public HttpResponse httpDelete(String baseUrl, String requestUrl){
        try{
            HttpClient client = HttpClientBuilder.create().build();
            HttpDelete delete = new HttpDelete(baseUrl+requestUrl);
            HttpResponse response = client.execute(delete);
            return response;
        }catch (Exception ex){
            String error = String.format("Could not fulfill request from Url: %s", baseUrl+requestUrl);
            logger.error(error, ex);
            throw new ServiceConnectionException(error, ex);
        }
    }

    public HttpResponse httpHead(String baseUrl, String requestUrl){
        try{
            HttpClient client = HttpClientBuilder.create().build();
            HttpHead head = new HttpHead(baseUrl+requestUrl);
            HttpResponse response = client.execute(head);
            return response;
        }catch (Exception ex){
            String error = String.format("Could not fulfill request from Url: %s", baseUrl+requestUrl);
            logger.error(error, ex);
            throw new ServiceConnectionException(error, ex);
        }
    }

    public HttpResponse httpPut(String baseUrl, String requestUrl, Object jsonObject){
        try{
            HttpClient client = HttpClientBuilder.create().build();
            HttpPut put = new HttpPut(baseUrl+requestUrl);
            ObjectMapper obj = new ObjectMapper();
            String putBodyString = obj.writeValueAsString(jsonObject);
            StringEntity requestEntity = new StringEntity(putBodyString, ContentType.APPLICATION_JSON);
            put.setEntity(requestEntity);
            HttpResponse response = client.execute(put);
            return response;
        }catch (Exception ex){
            String error = String.format("Could not fulfill request from Url: %s", baseUrl+requestUrl);
            logger.error(error, ex);
            throw new ServiceConnectionException(error, ex);
        }
    }
}
