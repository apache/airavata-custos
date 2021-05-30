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

package org.apache.custos.ssl.certificate.manager.acme;

import java.util.List;

public final class AcmeConfiguration {

    private String sessionUri;
    private String userKey = "user.key";
    private String domainKey = "domain.key";
    private String domainCsr = "domain.csr";
    private String domainChain = "domain-chain.crt";
    private int keySize = 2048;
    private List<String> domains;

    public String getSessionUri() {
        return sessionUri;
    }

    public String getUserKey() {
        return userKey;
    }

    public String getDomainKey() {
        return domainKey;
    }

    public String getDomainCsr() {
        return domainCsr;
    }

    public String getDomainChain() {
        return domainChain;
    }

    public List<String> getDomains() {
        return domains;
    }

    public int getKeySize() {
        return keySize;
    }

    public void setSessionUri(String sessionUri) {
        this.sessionUri = sessionUri;
    }

    public void setUserKey(String userKey) {
        this.userKey = userKey;
    }

    public void setDomainKey(String domainKey) {
        this.domainKey = domainKey;
    }

    public void setDomainCsr(String domainCsr) {
        this.domainCsr = domainCsr;
    }

    public void setDomainChain(String domainChain) {
        this.domainChain = domainChain;
    }

    public void setDomains(List<String> domains) {
        this.domains = domains;
    }
}
