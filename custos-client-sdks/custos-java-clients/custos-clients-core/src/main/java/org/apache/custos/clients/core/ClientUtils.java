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

package org.apache.custos.clients.core;

import io.grpc.Metadata;
import org.apache.custos.integration.core.utils.Constants;
import org.json.JSONObject;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.*;
import java.net.HttpURLConnection;
import java.net.URL;
import java.nio.charset.Charset;
import java.util.Base64;


/**
 * Utility methods for clients
 */
public class ClientUtils {

    private static final Logger LOGGER = LoggerFactory.getLogger(ClientUtils.class);

    /**
     * Creates authorization header
     *
     * @param clientId
     * @param clientSecret
     * @return
     */
    public static Metadata getAuthorizationHeader(String clientId, String clientSecret) {

        String credential = clientId + ":" + clientSecret;

        byte[] token = Base64.getEncoder().encode(credential.getBytes());

        String formattedString = new String(token);
        String headerStr = "Bearer " + formattedString;
        Metadata header = new Metadata();
        Metadata.Key<String> key = Metadata.Key.of("authorization", Metadata.ASCII_STRING_MARSHALLER);
        header.put(key, headerStr);
        return header;
    }


    public static Metadata getAuthorizationHeader(String accessToken) {

        String headerStr = "Bearer " + accessToken;
        Metadata header = new Metadata();
        Metadata.Key<String> key = Metadata.Key.of("authorization", Metadata.ASCII_STRING_MARSHALLER);
        header.put(key, headerStr);
        return header;
    }


    public static Metadata getUserTokenHeader(String accessToken) {
        Metadata header = new Metadata();
        Metadata.Key<String> key = Metadata.Key.of(Constants.USER_TOKEN, Metadata.ASCII_STRING_MARSHALLER);
        header.put(key, accessToken);
        return header;
    }

    /**
     * Provides file object
     *
     * @param classObj
     * @param fileName
     * @return
     */
    public static File getFile(Class classObj, final String fileName) {
        return new File(classObj
                .getResource(fileName)
                .getFile());
    }


    public static InputStream getServerCertificate(String serverHost, String clientId,
                                                   String clientSec) throws IOException {
        HttpURLConnection conn = null;
        InputStream inputStream = null;
        try {
            String url = "https://" + serverHost +
                    "/apiserver/resource-secret-management/v1.0.0/secret?metadata.owner_type=CUSTOS&metadata.resource_type=SERVER_CERTIFICATE";

            URL endpoint = new URL(url);
            conn = (HttpURLConnection) endpoint.openConnection();
            conn.setRequestMethod("GET");
            String userCredentials = clientId + ":" + clientSec;
            String basicAuth = "Basic " + new String(Base64.getEncoder().encode(userCredentials.getBytes()));
            conn.setRequestProperty("Accept", "application/json");
            conn.setRequestProperty("Authorization", basicAuth);
            if (conn.getResponseCode() != 200) {
                throw new RuntimeException("Failed : HTTP error code : "
                        + conn.getResponseCode());
            }

            StringBuilder responseStrBuilder = new StringBuilder();
            inputStream = conn.getInputStream();
            BufferedReader br = new BufferedReader(new InputStreamReader(
                    (inputStream)));

            String output;

            while ((output = br.readLine()) != null) {
                responseStrBuilder.append(output);
            }

            JSONObject jsonObject = new JSONObject(responseStrBuilder.toString());

            Object key = jsonObject.get("value");
            return new ByteArrayInputStream(((String) key).getBytes(Charset.forName("UTF-8")));

        } catch (Exception ex) {
            String msg = "Error occurred while fetching server certificate";
            LOGGER.error(msg, ex);
            throw ex;
        } finally {
            if (conn != null) {
                conn.disconnect();
                inputStream.close();

            }
        }

    }

}
