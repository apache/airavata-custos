/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing,
 *  software distributed under the License is distributed on an
 *  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *  KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

package org.apache.custos.ssl.certificate.manager.clients;

import org.apache.custos.ssl.certificate.manager.configurations.NginxConfiguration;
import org.apache.custos.ssl.certificate.manager.helpers.QueryString;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.net.HttpURLConnection;
import java.net.URL;

public class NginxClient {
    private static final Logger logger = LoggerFactory.getLogger(NginxClient.class);
    private final String acmeChallengeUrl;

    public NginxClient(NginxConfiguration config) {
        acmeChallengeUrl = config.getUrl() + config.getFolderPath();
    }

    public boolean createAcmeChallenge(String fileName, String fileContent) {
        QueryString query = new QueryString("file", fileName);
        query.add("content", fileContent);
        return send("POST", query.toString());
    }

    public boolean deleteAcmeChallenge(String fileName) {
        QueryString query = new QueryString("file", fileName);
        return send("DELETE", query.toString());
    }

    private boolean send(String method, String query) {
        HttpURLConnection con = null;
        int status = 0;
        try {
            URL url = new URL(acmeChallengeUrl + "?" + query);
            con = (HttpURLConnection) url.openConnection();
            con.setRequestMethod(method);
            status = con.getResponseCode();
        } catch (IOException e) {
            logger.info("Error in nginx client: {}", e.getMessage());
        } finally {
            if (con != null) {
                con.disconnect();
            }
        }

        return status == HttpURLConnection.HTTP_OK;
    }
}
