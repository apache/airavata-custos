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
import org.apache.custos.ssl.certificate.manager.helpers.QueryStringHelper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.net.HttpURLConnection;
import java.net.URL;


/**
 * Nginx client class to handle nginx related requests
 */
public class NginxClient {
    private static final Logger logger = LoggerFactory.getLogger(NginxClient.class);
    private final String caChallengeUrl;

    public NginxClient(NginxConfiguration config) {
        caChallengeUrl = config.getUrl() + "/ca/challenge";
    }


    /**
     * This method creates the necessary file in nginx which is needed
     * to validate the acme challenge
     *
     * @param fileName Name of the file
     * @param fileContent Content of the file
     * @return true if request is success else false
     */
    public boolean createAcmeChallenge(String fileName, String fileContent) {
        QueryStringHelper query = new QueryStringHelper("file", fileName);
        query.add("content", fileContent);
        return send("POST", query.toString());
    }

    /**
     * This method removes the challenge file created in nginx after
     * validation process is completed
     *
     * @param fileName Name of the file
     * @return true if request is success else false
     */
    public boolean deleteAcmeChallenge(String fileName) {
        QueryStringHelper query = new QueryStringHelper("file", fileName);
        return send("DELETE", query.toString());
    }

    private boolean send(String method, String query) {
        HttpURLConnection con = null;
        int status = 0;
        try {
            URL url = new URL(caChallengeUrl + "?" + query);
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
