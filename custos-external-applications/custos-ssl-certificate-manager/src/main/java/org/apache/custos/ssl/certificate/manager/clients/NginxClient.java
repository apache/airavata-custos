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

import java.io.IOException;
import java.net.HttpURLConnection;
import java.net.URL;

public class NginxClient {

    private String baseUrl;

    public NginxClient(NginxConfiguration config) {
        this.baseUrl = config.getUrl() + config.getFolderPath();
    }

    public boolean createChallenge(String fileName, String content) {
        HttpURLConnection con = null;
        int status = 0;
        try {
            URL url = new URL(this.baseUrl + "?file=" + fileName + "&content=" + content);
            con = (HttpURLConnection) url.openConnection();
            status = con.getResponseCode();
        } catch (IOException e) {
            e.printStackTrace();
        } finally {
            con.disconnect();
        }

        return status == 200 ? true : false;
    }
}
