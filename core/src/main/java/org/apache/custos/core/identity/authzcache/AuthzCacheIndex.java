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

package org.apache.custos.core.identity.authzcache;

/**
 * Cache index of the default authorization cache.
 */
public class AuthzCacheIndex {

    private String subject;
    private String oauthAccessToken;
    private String tenantId;

    public AuthzCacheIndex(String userName, String tenantId, String accessToken) {
        this.subject = userName;
        this.oauthAccessToken = accessToken;
        this.tenantId = tenantId;
    }

    public String getSubject() {
        return subject;
    }

    public void setSubject(String subject) {
        this.subject = subject;
    }

    public String getOauthAccessToken() {
        return oauthAccessToken;
    }

    public void setOauthAccessToken(String oauthAccessToken) {
        this.oauthAccessToken = oauthAccessToken;
    }

    public String getTenantId() {
        return tenantId;
    }

    public void setTenantId(String tenantId) {
        this.tenantId = tenantId;
    }

    /*Equals and hash code methods are overridden since this is being used as an index of a map and that containsKey method
     * should return true if the values of two index objects are equal.*/
    @Override
    public boolean equals(Object other) {
        if (other == null || other.getClass() != getClass()) {
            return false;
        }
        return ((this.getSubject().equals(((AuthzCacheIndex) other).getSubject()))
                && (this.getTenantId().equals(((AuthzCacheIndex) other).getTenantId()))
                && (this.getOauthAccessToken().equals(((AuthzCacheIndex) other).getOauthAccessToken())));
    }

    @Override
    public int hashCode() {
        return this.getSubject().hashCode() + this.getOauthAccessToken().hashCode() + this.getTenantId().hashCode();
    }
}
