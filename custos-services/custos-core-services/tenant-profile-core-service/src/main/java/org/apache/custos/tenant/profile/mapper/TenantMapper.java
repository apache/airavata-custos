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

package org.apache.custos.tenant.profile.mapper;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.custos.tenant.profile.persistance.model.Contact;
import org.apache.custos.tenant.profile.persistance.model.RedirectURI;
import org.apache.custos.tenant.profile.persistance.model.Tenant;
import org.apache.custos.tenant.profile.service.TenantStatus;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.*;


/**
 * This class maps attributes between grpc Tenant to DB Tenant table
 */
public class TenantMapper {

    private static final Logger LOGGER = LoggerFactory.getLogger(TenantMapper.class);

    /**
     * Maps gRPC Tenant Model to DB Layer Tenant Entity
     *
     * @param {@link org.apache.custos.tenant.profile.service.Tenant} tenant
     * @return Tenant
     */
    public static Tenant createTenantEntityFromTenant(org.apache.custos.tenant.profile.service.Tenant tenant) {
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
//        tenantEntity.setJwks(tenant.getJwksCount() > 0 ? getJWKSAsString(tenant.getJwksMap()) : null);
        tenantEntity.setJwksUri(tenant.getJwksUri());
        tenantEntity.setExample_extension_parameter(tenant.getExampleExtensionParameter());
        tenantEntity.setTosUri(tenant.getTosUri());
        tenantEntity.setPolicyUri(tenant.getPolicyUri());
        tenantEntity.setSoftwareId(tenant.getSoftwareId());
        tenantEntity.setSoftwareVersion(tenant.getSoftwareVersion());
        tenantEntity.setRefreshTokenLifetime(tenant.getRefeshTokenLifetime());


        Set<Contact> contactSet = new HashSet<Contact>();

        for (int i = 0; i < tenant.getContactsCount(); i++) {

            String contact = tenant.getContacts(i);
            Contact contactEntity = new Contact();
            contactEntity.setTenant(tenantEntity);
            contactEntity.setContactInfo(contact);
            contactSet.add(contactEntity);
        }

        tenantEntity.setContacts(contactSet);

        Set<RedirectURI> redirectURIS = new HashSet<RedirectURI>();
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
     * Transform TenantEntity to Tenant
     *
     * @param tenantEntity
     * @return tenant
     */
    public static org.apache.custos.tenant.profile.service.Tenant createTenantFromTenantEntity(Tenant tenantEntity) {

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


        return org.apache.custos.tenant.profile.service.Tenant.newBuilder()
                .setAdminEmail(tenantEntity.getAdminEmail())
                .setAdminFirstName(tenantEntity.getAdminFirstName())
                .setAdminLastName(tenantEntity.getAdminLastName())
                .setDomain(tenantEntity.getDomain())
                .setClientUri(tenantEntity.getUri()==null?tenantEntity.getLogoURI():tenantEntity.getUri())
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


    public static String getTenantInfoAsString(org.apache.custos.tenant.profile.service.Tenant tenant) {
        StringBuffer buffer = new StringBuffer();
        buffer.append("tenantName : " + tenant.getClientName());
        buffer.append("\n");
        buffer.append("tenantId : " + tenant.getTenantId());
        buffer.append("\n");
        buffer.append("tenantAdminEmail : " + tenant.getAdminEmail());
        buffer.append("\n");
        buffer.append("tenantAdminFirstName : " + tenant.getAdminFirstName());
        buffer.append("\n");
        buffer.append("tenantAdminLastName : " + tenant.getAdminLastName());
        buffer.append("domain : " + tenant.getDomain());
        buffer.append("\n");
        buffer.append("logoURI : " + tenant.getClientUri());
        buffer.append("\n");
        buffer.append("\n");
        buffer.append("requesterEmail : " + tenant.getRequesterEmail());
        buffer.append("\n");
        buffer.append("tenantScope : " + tenant.getScope());
        buffer.append("\n");
        buffer.append("contacts  : " + tenant.getContactsList().toString());
        buffer.append("\n");
        buffer.append("redirectURIs : " + tenant.getRedirectUrisList().toString());
        buffer.append("\n");

        return buffer.toString();

    }


    private static String getJWKSAsString(Map<String, String> jwksMap) {
        ObjectMapper objectMapper = new ObjectMapper();

        try {
            return objectMapper.writeValueAsString(jwksMap);
        } catch (JsonProcessingException e) {
            LOGGER.error("Error occurred while printing json ", e);
            return null;
        }
    }


}



