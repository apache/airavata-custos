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

package org.apache.custos.core.mapper.tenant;

import org.apache.custos.core.model.tenant.Contact;
import org.apache.custos.core.model.tenant.RedirectURI;
import org.apache.custos.core.model.tenant.Tenant;
import org.apache.custos.core.tenant.profile.api.TenantStatus;

import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;


/**
 * The TenantMapper class provides methods for mapping between Tenant objects and Tenant entity objects.
 */
public class TenantMapper {

    /**
     * Creates a Tenant entity object from a Tenant object.
     *
     * @param tenant the Tenant object to convert
     * @return the created Tenant entity object
     */
    public static Tenant createTenantEntityFromTenant(org.apache.custos.core.tenant.profile.api.Tenant tenant) {
        Tenant tenantEntity = new Tenant();
        tenantEntity.setId(tenant.getTenantId());
        tenantEntity.setName(tenant.getClientName());
        tenantEntity.setStatus(tenant.getTenantStatus().name());
        tenantEntity.setAdminFirstName(tenant.getAdminFirstName());
        tenantEntity.setAdminLastName(tenant.getAdminLastName());
        tenantEntity.setAdminEmail(tenant.getAdminEmail());
        tenantEntity.setRequesterEmail(tenant.getRequesterEmail());
        tenantEntity.setLogoURI(tenant.getLogoUri());
        tenantEntity.setScope(tenant.getScope());
        tenantEntity.setDomain(tenant.getDomain());
        tenantEntity.setAdminUsername(tenant.getAdminUsername());
        tenantEntity.setComment(tenant.getComment());
        tenantEntity.setUri(tenant.getClientUri());
        tenantEntity.setParentId(tenant.getParentTenantId());
        tenantEntity.setApplicationType(tenant.getApplicationType());
        tenantEntity.setJwksUri(tenant.getJwksUri());
        tenantEntity.setExample_extension_parameter(tenant.getExampleExtensionParameter());
        tenantEntity.setTosUri(tenant.getTosUri());
        tenantEntity.setPolicyUri(tenant.getPolicyUri());
        tenantEntity.setSoftwareId(tenant.getSoftwareId());
        tenantEntity.setSoftwareVersion(tenant.getSoftwareVersion());
        tenantEntity.setRefreshTokenLifetime(tenant.getRefeshTokenLifetime());

        Set<Contact> contactSet = new HashSet<>();
        for (int i = 0; i < tenant.getContactsCount(); i++) {
            String contact = tenant.getContacts(i);
            Contact contactEntity = new Contact();
            contactEntity.setTenant(tenantEntity);
            contactEntity.setContactInfo(contact);
            contactSet.add(contactEntity);
        }

        tenantEntity.setContacts(contactSet);
        Set<RedirectURI> redirectURIS = new HashSet<>();
        for (int i = 0; i < tenant.getRedirectUrisCount(); i++) {
            String uri = tenant.getRedirectUris(i);
            RedirectURI redirectURIEntity = new RedirectURI();
            redirectURIEntity.setTenant(tenantEntity);
            redirectURIEntity.setRedirectURI(uri);
            redirectURIS.add(redirectURIEntity);
        }

        tenantEntity.setRedirectURIS(redirectURIS);

        return tenantEntity;
    }

    /**
     * Converts a Tenant entity object to a Tenant object.
     *
     * @param tenantEntity the Tenant entity object to convert
     * @return the created Tenant object
     */
    public static org.apache.custos.core.tenant.profile.api.Tenant createTenantFromTenantEntity(Tenant tenantEntity) {
        Set<Contact> contacts = tenantEntity.getContacts();
        List<String> contactList = new ArrayList<>();
        if (contacts != null && !contacts.isEmpty()) {
            for (Contact contact : contacts) {
                contactList.add(contact.getContactInfo());
            }
        }

        Set<RedirectURI> redirectURIS = tenantEntity.getRedirectURIS();
        List<String> uriList = new ArrayList<>();
        if (redirectURIS != null && !redirectURIS.isEmpty()) {
            for (RedirectURI redirectURI : redirectURIS) {
                uriList.add(redirectURI.getRedirectURI());
            }
        }

        return org.apache.custos.core.tenant.profile.api.Tenant.newBuilder()
                .setAdminEmail(tenantEntity.getAdminEmail())
                .setAdminFirstName(tenantEntity.getAdminFirstName())
                .setAdminLastName(tenantEntity.getAdminLastName())
                .setDomain(tenantEntity.getDomain())
                .setClientUri(tenantEntity.getUri() == null ? tenantEntity.getLogoURI() : tenantEntity.getUri())
                .setRequesterEmail(tenantEntity.getRequesterEmail())
                .setScope(tenantEntity.getScope())
                .addAllContacts(contactList)
                .addAllRedirectUris(uriList)
                .setClientName(tenantEntity.getName())
                .setTenantId(tenantEntity.getId())
                .setTenantStatus(TenantStatus.valueOf(tenantEntity.getStatus()))
                .setAdminUsername(tenantEntity.getAdminUsername())
                .setComment(tenantEntity.getComment())
                .setLogoUri(tenantEntity.getLogoURI())
                .setApplicationType(tenantEntity.getApplicationType())
                .setJwksUri(tenantEntity.getJwksUri())
                .setExampleExtensionParameter(tenantEntity.getExample_extension_parameter())
                .setTosUri(tenantEntity.getTosUri())
                .setPolicyUri(tenantEntity.getPolicyUri())
                .setSoftwareId(tenantEntity.getSoftwareId())
                .setSoftwareVersion(tenantEntity.getSoftwareVersion())
                .setRefeshTokenLifetime(tenantEntity.getRefreshTokenLifetime())
                .setParentTenantId(tenantEntity.getParentId())
                .build();
    }

    /**
     * Retrieves the information of a tenant object as a formatted string.
     *
     * @param tenant the tenant object to retrieve the information from
     * @return the tenant information as a string
     */
    public static String getTenantInfoAsString(org.apache.custos.core.tenant.profile.api.Tenant tenant) {
        StringBuffer buffer = new StringBuffer();
        buffer.append("tenantName : ").append(tenant.getClientName());
        buffer.append("\n");
        buffer.append("tenantId : ").append(tenant.getTenantId());
        buffer.append("\n");
        buffer.append("tenantAdminEmail : ").append(tenant.getAdminEmail());
        buffer.append("\n");
        buffer.append("tenantAdminFirstName : ").append(tenant.getAdminFirstName());
        buffer.append("\n");
        buffer.append("tenantAdminLastName : ").append(tenant.getAdminLastName());
        buffer.append("domain : ").append(tenant.getDomain());
        buffer.append("\n");
        buffer.append("logoURI : ").append(tenant.getClientUri());
        buffer.append("\n");
        buffer.append("\n");
        buffer.append("requesterEmail : ").append(tenant.getRequesterEmail());
        buffer.append("\n");
        buffer.append("tenantScope : ").append(tenant.getScope());
        buffer.append("\n");
        buffer.append("contacts  : ").append(tenant.getContactsList());
        buffer.append("\n");
        buffer.append("redirectURIs : ").append(tenant.getRedirectUrisList());
        buffer.append("\n");

        return buffer.toString();
    }
}
