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

public final class AcmeConfiguration implements Configuration {

    private String url;
    private String userKey;
    private String domainKey;
    private String domainCertificate = "domain.crt";
    private int keySize = 2048;
    private List<String> domains;

    public AcmeConfiguration() {
    }

    public AcmeConfiguration(Map<String, String> env) {
        this.url = env.get(Constants.ACME_URL);
        this.domains = Arrays.asList(env.get(Constants.ACME_DOMAINS).split(" "));
        this.userKey = env.get(Constants.ACME_USER_KEY);
        this.domainKey = env.get(Constants.ACME_DOMAIN_KEY);
    }

    public String getUrl() {
        return url;
    }

    public void setUrl(String url) {
        this.url = url;
    }

    public String getUserKey() {
        return userKey;
    }

    public void setUserKey(String userKey) {
        this.userKey = userKey;
    }

    public String getDomainKey() {
        return domainKey;
    }

    public void setDomainKey(String domainKey) {
        this.domainKey = domainKey;
    }

    public String getDomainCertificate() {
        return domainCertificate;
    }

    public void setDomainCertificate(String domainCertificate) {
        this.domainCertificate = domainCertificate;
    }

    public int getKeySize() {
        return keySize;
    }

    public void setKeySize(int keySize) {
        this.keySize = keySize;
    }

    public List<String> getDomains() {
        return domains;
    }

    public void setDomains(List<String> domains) {
        this.domains = domains;
    }

    @Override
    public void validate() {

    }
}
