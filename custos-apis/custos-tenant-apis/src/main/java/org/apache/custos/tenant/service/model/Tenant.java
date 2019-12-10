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
package org.apache.custos.tenant.service.model;

import com.fasterxml.jackson.annotation.JsonProperty;
import org.springframework.validation.annotation.Validated;

import javax.validation.Valid;
import javax.validation.constraints.NotNull;
import java.util.ArrayList;
import java.util.List;
import java.util.Objects;

/**
 * Tenant
 */
@Validated

public class Tenant   {
    @JsonProperty("tenantName")
    private String tenantName = null;

    @JsonProperty("requesterEmail")
    private String requesterEmail = null;

    @JsonProperty("tenantURI")
    private String tenantURI = null;

    @JsonProperty("redirectURIs")
    @Valid
    private List<String> redirectURIs = new ArrayList<String>();

    @JsonProperty("tenantId")
    private String tenantId = null;

    @JsonProperty("logoURI")
    private String logoURI = null;

    @JsonProperty("scope")
    private String scope = null;

    @JsonProperty("domain")
    private String domain = null;

    @JsonProperty("tenantAdminFirstName")
    private String tenantAdminFirstName = null;

    @JsonProperty("tenantAdminLastName")
    private String tenantAdminLastName = null;

    @JsonProperty("contacts")
    @Valid
    private List<String> contacts = null;

    public Tenant tenantName(String tenantName) {
        this.tenantName = tenantName;
        return this;
    }

    /**
     * Get tenantName
     * @return tenantName
     **/
    @NotNull
    public String getTenantName() {
        return tenantName;
    }

    public void setTenantName(String tenantName) {
        this.tenantName = tenantName;
    }

    public Tenant requesterEmail(String requesterEmail) {
        this.requesterEmail = requesterEmail;
        return this;
    }

    /**
     * Get requesterEmail
     * @return requesterEmail
     **/
    @NotNull
    public String getRequesterEmail() {
        return requesterEmail;
    }

    public void setRequesterEmail(String requesterEmail) {
        this.requesterEmail = requesterEmail;
    }

    public Tenant tenantURI(String tenantURI) {
        this.tenantURI = tenantURI;
        return this;
    }

    /**
     * URL string of a web page providing information about the client.The value of this field MUST point to a valid web page.
     * @return tenantURI
     **/

    @NotNull
    public String getTenantURI() {
        return tenantURI;
    }

    public void setTenantURI(String tenantURI) {
        this.tenantURI = tenantURI;
    }

    public Tenant redirectURIs(List<String> redirectURIs) {
        this.redirectURIs = redirectURIs;
        return this;
    }

    public Tenant addRedirectURIsItem(String redirectURIsItem) {
        this.redirectURIs.add(redirectURIsItem);
        return this;
    }

    /**
     * Get redirectURIs
     * @return redirectURIs
     **/
    @NotNull
    public List<String> getRedirectURIs() {
        return redirectURIs;
    }

    public void setRedirectURIs(List<String> redirectURIs) {
        this.redirectURIs = redirectURIs;
    }

    public Tenant tenantId(String tenantId) {
        this.tenantId = tenantId;
        return this;
    }

    /**
     * Get tenantId
     * @return tenantId
     **/

    public String getTenantId() {
        return tenantId;
    }

    public void setTenantId(String tenantId) {
        this.tenantId = tenantId;
    }

    public Tenant logoURI(String logoURI) {
        this.logoURI = logoURI;
        return this;
    }

    /**
     * URL string that references a logo for the client.
     * @return logoURI
     **/

    public String getLogoURI() {
        return logoURI;
    }

    public void setLogoURI(String logoURI) {
        this.logoURI = logoURI;
    }

    public Tenant scope(String scope) {
        this.scope = scope;
        return this;
    }

    /**
     * Get scope
     * @return scope
     **/

    public String getScope() {
        return scope;
    }

    public void setScope(String scope) {
        this.scope = scope;
    }

    public Tenant domain(String domain) {
        this.domain = domain;
        return this;
    }

    /**
     * Get domain
     * @return domain
     **/

    public String getDomain() {
        return domain;
    }

    public void setDomain(String domain) {
        this.domain = domain;
    }

    public Tenant tenantAdminFirstName(String tenantAdminFirstName) {
        this.tenantAdminFirstName = tenantAdminFirstName;
        return this;
    }

    /**
     * Get tenantAdminFirstName
     * @return tenantAdminFirstName
     **/

    public String getTenantAdminFirstName() {
        return tenantAdminFirstName;
    }

    public void setTenantAdminFirstName(String tenantAdminFirstName) {
        this.tenantAdminFirstName = tenantAdminFirstName;
    }

    public Tenant tenantAdminLastName(String tenantAdminLastName) {
        this.tenantAdminLastName = tenantAdminLastName;
        return this;
    }

    /**
     * Get tenantAdminLastName
     * @return tenantAdminLastName
     **/
    public String getTenantAdminLastName() {
        return tenantAdminLastName;
    }

    public void setTenantAdminLastName(String tenantAdminLastName) {
        this.tenantAdminLastName = tenantAdminLastName;
    }

    public Tenant contacts(List<String> contacts) {
        this.contacts = contacts;
        return this;
    }

    public Tenant addContactsItem(String contactsItem) {
        if (this.contacts == null) {
            this.contacts = new ArrayList<String>();
        }
        this.contacts.add(contactsItem);
        return this;
    }

    /**
     * Get contacts
     * @return contacts
     **/
    public List<String> getContacts() {
        return contacts;
    }

    public void setContacts(List<String> contacts) {
        this.contacts = contacts;
    }


    @Override
    public boolean equals(java.lang.Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        Tenant tenant = (Tenant) o;
        return Objects.equals(this.tenantName, tenant.tenantName) &&
                Objects.equals(this.requesterEmail, tenant.requesterEmail) &&
                Objects.equals(this.tenantURI, tenant.tenantURI) &&
                Objects.equals(this.redirectURIs, tenant.redirectURIs) &&
                Objects.equals(this.tenantId, tenant.tenantId) &&
                Objects.equals(this.logoURI, tenant.logoURI) &&
                Objects.equals(this.scope, tenant.scope) &&
                Objects.equals(this.domain, tenant.domain) &&
                Objects.equals(this.tenantAdminFirstName, tenant.tenantAdminFirstName) &&
                Objects.equals(this.tenantAdminLastName, tenant.tenantAdminLastName) &&
                Objects.equals(this.contacts, tenant.contacts);
    }

    @Override
    public int hashCode() {
        return Objects.hash(tenantName, requesterEmail, tenantURI, redirectURIs, tenantId, logoURI, scope, domain, tenantAdminFirstName, tenantAdminLastName, contacts);
    }

    @Override
    public String toString() {
        StringBuilder sb = new StringBuilder();
        sb.append("class Tenant {\n");

        sb.append("    tenantName: ").append(toIndentedString(tenantName)).append("\n");
        sb.append("    requesterEmail: ").append(toIndentedString(requesterEmail)).append("\n");
        sb.append("    tenantURI: ").append(toIndentedString(tenantURI)).append("\n");
        sb.append("    redirectURIs: ").append(toIndentedString(redirectURIs)).append("\n");
        sb.append("    tenantId: ").append(toIndentedString(tenantId)).append("\n");
        sb.append("    logoURI: ").append(toIndentedString(logoURI)).append("\n");
        sb.append("    scope: ").append(toIndentedString(scope)).append("\n");
        sb.append("    domain: ").append(toIndentedString(domain)).append("\n");
        sb.append("    tenantAdminFirstName: ").append(toIndentedString(tenantAdminFirstName)).append("\n");
        sb.append("    tenantAdminLastName: ").append(toIndentedString(tenantAdminLastName)).append("\n");
        sb.append("    contacts: ").append(toIndentedString(contacts)).append("\n");
        sb.append("}");
        return sb.toString();
    }

    /**
     * Convert the given object to string with each line indented by 4 spaces
     * (except the first line).
     */
    private String toIndentedString(java.lang.Object o) {
        if (o == null) {
            return "null";
        }
        return o.toString().replace("\n", "\n    ");
    }
}
