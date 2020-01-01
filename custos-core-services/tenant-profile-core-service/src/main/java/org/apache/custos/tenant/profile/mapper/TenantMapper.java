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

import org.apache.custos.tenant.profile.persistance.model.Contact;
import org.apache.custos.tenant.profile.persistance.model.RedirectURI;
import org.apache.custos.tenant.profile.persistance.model.Tenant;

import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;


/**
 * This class maps attributes between grpc Tenant to DB Tenant table
 */
public class TenantMapper {


    /**
     * Maps gRPC Tenant Model to DB Layer Tenant Entity
     *
     * @param {@link org.apache.custos.tenant.profile.service.Tenant} tenant
     * @return Tenant
     */
    public static Tenant createTenantEntityFromTenant(org.apache.custos.tenant.profile.service.Tenant tenant) {
        Tenant tenantEntity = new Tenant();
        if (tenant.getTenantId() != 0){
            tenantEntity.setId(tenant.getTenantId());
        }
        tenantEntity.setName(tenant.getTenantName());
        tenantEntity.setStatus(tenant.getTenantStatus());
        tenantEntity.setAdminFirstName(tenant.getAdminFirstName());
        tenantEntity.setAdminLastName(tenant.getAdminLastName());
        tenantEntity.setAdminEmail(tenant.getAdminEmail());
        tenantEntity.setRequesterEmail(tenant.getRequesterEmail());
        tenantEntity.setLogoURI(tenant.getTenantURI());
        tenantEntity.setScope(tenant.getScope());
        tenantEntity.setDomain(tenant.getDomain());

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
        for (int i = 0; i < tenant.getRedirectURIsCount(); i++) {

            String uri = tenant.getRedirectURIs(i);
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
                .setTenantURI(tenantEntity.getLogoURI())
                .setRequesterEmail(tenantEntity.getRequesterEmail())
                .setScope(tenantEntity.getScope())
                .addAllContacts(contactList)
                .addAllRedirectURIs(uriList)
                .setTenantName(tenantEntity.getName())
                .setTenantId(tenantEntity.getId())
                .setTenantStatus(tenantEntity.getStatus())
                .build();


    }


    public static String getTenantInfoAsString (org.apache.custos.tenant.profile.service.Tenant tenant) {
        StringBuffer buffer = new StringBuffer();
        buffer.append("tenantName : "+tenant.getTenantName());
        buffer.append("\n");
        buffer.append("tenantId : "+tenant.getTenantId());
        buffer.append("\n");
        buffer.append("tenantAdminEmail : "+tenant.getAdminEmail());
        buffer.append("\n");
        buffer.append("tenantAdminFirstName : "+tenant.getAdminFirstName());
        buffer.append("\n");
        buffer.append("tenantAdminLastName : "+tenant.getAdminLastName());
        buffer.append("domain : "+tenant.getDomain());
        buffer.append("\n");
        buffer.append("logoURI : "+tenant.getTenantURI());
        buffer.append("\n");
        buffer.append("\n");
        buffer.append("requesterEmail : "+tenant.getRequesterEmail());
        buffer.append("\n");
        buffer.append("tenantScope : "+tenant.getScope());
        buffer.append("\n");
        buffer.append("contacts  : "+tenant.getContactsList().toString());
        buffer.append("\n");
        buffer.append("redirectURIs : "+ tenant.getRedirectURIsList().toString());
        buffer.append("\n");

        return buffer.toString();

    }


//    private <T> Set<T>  difference(Set<T> newSet, Set<T> oldSet) {
//        Set<T> diffSet = new HashSet<>();
//
//
//
//    }


}



