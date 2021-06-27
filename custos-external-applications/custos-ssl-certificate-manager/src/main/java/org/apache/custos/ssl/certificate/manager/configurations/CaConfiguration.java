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

import java.util.Arrays;
import java.util.List;
import java.util.Map;

/**
 * Configuration class for Acme
 */
public final class CaConfiguration {

    private String url;
    private String userKeyPath;
    private String domainKeyPath;
    private List<String> domains;

    /**
     * @param env Map containing acme configuration through environment
     *            variables or application.properties
     */
    public CaConfiguration(Map<String, String> env) {
        this.url = env.get(Constants.CA_URL);
        this.domains = Arrays.asList(env.get(Constants.CA_DOMAINS).split(" "));
        this.userKeyPath = env.get(Constants.CA_USER_KEY_PATH);
        this.domainKeyPath = env.get(Constants.CA_DOMAIN_KEY_PATH);
    }

    public String getUrl() {
        return url;
    }

    public String getUserKeyPath() {
        return userKeyPath;
    }

    public String getDomainKeyPath() {
        return domainKeyPath;
    }

    public List<String> getDomains() {
        return domains;
    }

}
