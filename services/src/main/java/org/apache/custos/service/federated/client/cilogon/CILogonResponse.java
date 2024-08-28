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

package org.apache.custos.service.federated.client.cilogon;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
public class CILogonResponse {

    @JsonProperty(value = "client_id")
    private String clientId;

    @JsonProperty(value = "client_secret")
    private String clientSecret;

    @JsonProperty(value = "client_secret_expired_at")
    private long clientSecretExpiredAt;

    @JsonProperty(value = "registration_client_uri")
    private String registrationClientURI;

    @JsonProperty(value = "client_name")
    private String clientName;

    @JsonProperty(value = "redirect_uris")
    private String[] redirectURIs = new String[0];

    @JsonProperty(value = "scope")
    private String[] scope = new String[0];

    @JsonProperty(value = "grant_types")
    private String[] grantTypes;

    @JsonProperty(value = "comment")
    private String comment;

    @JsonProperty(value = "client_id_issued_at")
    private long clientIdIssuedAt;

    @JsonProperty(value = "contacts")
    private String[] contacts = new String[0];

    @JsonProperty(value = "client_uri")
    private String clientURI;


    public String getClientId() {
        return clientId;
    }

    public void setClientId(String clientId) {
        this.clientId = clientId;
    }

    public String getClientSecret() {
        return clientSecret;
    }

    public void setClientSecret(String clientSecret) {
        this.clientSecret = clientSecret;
    }

    public long getClientSecretExpiredAt() {
        return clientSecretExpiredAt;
    }

    public void setClientSecretExpiredAt(long clientSecretExpiredAt) {
        this.clientSecretExpiredAt = clientSecretExpiredAt;
    }

    public String getRegistrationClientURI() {
        return registrationClientURI;
    }

    public void setRegistrationClientURI(String registrationClientURI) {
        this.registrationClientURI = registrationClientURI;
    }

    public String getClientName() {
        return clientName;
    }

    public void setClientName(String clientName) {
        this.clientName = clientName;
    }

    public String[] getRedirectURIs() {
        return redirectURIs;
    }

    public void setRedirectURIs(String[] redirectURIs) {
        this.redirectURIs = redirectURIs;
    }

    public String[] getScope() {
        return scope;
    }

    public void setScope(String[] scope) {
        this.scope = scope;
    }

    public String[] getGrantTypes() {
        return grantTypes;
    }

    public void setGrantTypes(String[] grantTypes) {
        this.grantTypes = grantTypes;
    }

    public String getComment() {
        return comment;
    }

    public void setComment(String comment) {
        this.comment = comment;
    }

    public long getClientIdIssuedAt() {
        return clientIdIssuedAt;
    }

    public void setClientIdIssuedAt(long clientIdIssuedAt) {
        this.clientIdIssuedAt = clientIdIssuedAt;
    }

    public String[] getContacts() {
        return contacts;
    }

    public void setContacts(String[] contacts) {
        this.contacts = contacts;
    }

    public String getClientURI() {
        return clientURI;
    }

    public void setClientURI(String clientURI) {
        this.clientURI = clientURI;
    }
}
