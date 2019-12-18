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

import org.apache.custos.tenant.profile.persistance.model.*;

import java.util.HashSet;
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
        if (tenant.getTenantId() != null) {
            Long castedId = Long.valueOf(tenant.getTenantId());
            tenantEntity.setId(castedId);
        }
        tenantEntity.setName(tenant.getTenantName());
        tenantEntity.setStatus(tenant.getTenantStatus());
        tenantEntity.setRequesterUsername(tenant.getRequesterUsername());
        tenantEntity.setAdminFirstName(tenant.getAdminFirstName());
        tenantEntity.setAdminLastName(tenant.getAdminLastName());
        tenantEntity.setAdminEmail(tenant.getAdminEmail());
        tenantEntity.setRequesterUsername(tenant.getRequesterUsername());
        tenantEntity.setRequesterEmail(tenant.getRequesterEmail());
        tenantEntity.setLogoURI(tenant.getLogoURI());
        tenantEntity.setScope(tenant.getScope());
        tenantEntity.setDomain(tenant.getDomain());

        Set<Contact> contactSet = new HashSet<Contact>();
        for (int i = 0; i < tenant.getContactsCount(); i++) {

            String contact = tenant.getContacts(i);
            Contact contactEntity = new Contact();
            contactEntity.setContactInfo(contact);
            contactSet.add(contactEntity);
        }

        tenantEntity.setContacts(contactSet);

        Set<RedirectURI> redirectURIS = new HashSet<RedirectURI>();
        for (int i = 0; i < tenant.getRedirectURIsCount(); i++) {

            String uri = tenant.getRedirectURIs(i);
            RedirectURI redirectURIEntity = new RedirectURI();
            redirectURIEntity.setRedirectURI(uri);
            redirectURIS.add(redirectURIEntity);
        }

        tenantEntity.setRedirectURIS(redirectURIS);

        return tenantEntity;

    }






}
