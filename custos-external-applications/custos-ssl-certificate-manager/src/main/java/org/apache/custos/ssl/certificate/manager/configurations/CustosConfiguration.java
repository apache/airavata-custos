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

package org.apache.custos.ssl.certificate.manager.configurations;

import java.util.Map;

public class CustosConfiguration {

    private String url;
    private int port;
    private String clientId;
    private String clientSecret;
    private String ownerId;

    public CustosConfiguration(Map<String, String> env) {
        this.url = env.get(Constants.CUSTOS_URL);
        this.port = Integer.parseInt(env.get(Constants.CUSTOS_PORT));
        this.clientId = env.get(Constants.CUSTOS_CLIENT_ID);
        this.clientSecret = env.get(Constants.CUSTOS_CLIENT_SECRET);
        this.ownerId = env.get(Constants.CUSTOS_OWNER_ID);
    }

    public String getUrl() {
        return url;
    }

    public int getPort() {
        return port;
    }

    public String getClientId() {
        return clientId;
    }

    public String getClientSecret() {
        return clientSecret;
    }

    public String getOwnerId() {
        return ownerId;
    }

}
