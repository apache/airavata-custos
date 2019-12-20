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

import org.apache.custos.tenant.profile.persistance.model.AttributeUpdateMetadata;
import org.apache.custos.tenant.profile.persistance.model.Tenant;
import org.apache.custos.tenant.profile.service.TenantAttributeUpdateMetadata;

import java.util.HashSet;
import java.util.Set;

/**
 * This class maps attributes between grpc TenantAttributeUpdateMetadata to DB AttributeUpdateMetadata table
 */
public class AttributeUpdateMetadataMapper {


    /**
     * This creates Attribute update entity List from comparing oldTenant and newTenant
     *
     * @param oldTenant
     * @param newTenant
     * @param updatedBy
     * @return
     */
    public static Set<AttributeUpdateMetadata> createAttributeUpdateMetadataEntity(Tenant oldTenant, Tenant newTenant, String updatedBy) {

        Set<AttributeUpdateMetadata> metadataSet = new HashSet<>();
        if (!oldTenant.getName().equals(newTenant.getName())) {
            AttributeUpdateMetadata attributeUpdateMetadata = new AttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("name");
            attributeUpdateMetadata.setUpdatedFieldValue(newTenant.getName());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }

        if (!oldTenant.getAdminEmail().equals(newTenant.getAdminEmail())) {
            AttributeUpdateMetadata attributeUpdateMetadata = new AttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("adminEmail");
            attributeUpdateMetadata.setUpdatedFieldValue(newTenant.getAdminEmail());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }

        if (!oldTenant.getAdminFirstName().equals(newTenant.getAdminFirstName())) {
            AttributeUpdateMetadata attributeUpdateMetadata = new AttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("adminFirstName");
            attributeUpdateMetadata.setUpdatedFieldValue(newTenant.getAdminFirstName());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }

        if (!oldTenant.getAdminLastName().equals(newTenant.getAdminLastName())) {
            AttributeUpdateMetadata attributeUpdateMetadata = new AttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("adminLastName");
            attributeUpdateMetadata.setUpdatedFieldValue(newTenant.getAdminLastName());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }

        if (!oldTenant.getDomain().equals(newTenant.getDomain())) {
            AttributeUpdateMetadata attributeUpdateMetadata = new AttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("domain");
            attributeUpdateMetadata.setUpdatedFieldValue(newTenant.getDomain());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }

        if (!oldTenant.getLogoURI().equals(newTenant.getLogoURI())) {
            AttributeUpdateMetadata attributeUpdateMetadata = new AttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("logoURI");
            attributeUpdateMetadata.setUpdatedFieldValue(newTenant.getLogoURI());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }

        if (!oldTenant.getRequesterEmail().equals(newTenant.getRequesterEmail())) {
            AttributeUpdateMetadata attributeUpdateMetadata = new AttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("requesterEmail");
            attributeUpdateMetadata.setUpdatedFieldValue(newTenant.getRequesterEmail());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }

        if (!oldTenant.getRequesterUsername().equals(newTenant.getRequesterUsername())) {
            AttributeUpdateMetadata attributeUpdateMetadata = new AttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("requesterUsername");
            attributeUpdateMetadata.setUpdatedFieldValue(newTenant.getRequesterUsername());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }

        if (!oldTenant.getScope().equals(newTenant.getScope())) {
            AttributeUpdateMetadata attributeUpdateMetadata = new AttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("scope");
            attributeUpdateMetadata.setUpdatedFieldValue(newTenant.getScope());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }

        if (!oldTenant.getRedirectURIS().equals(newTenant.getRedirectURIS())) {
            AttributeUpdateMetadata attributeUpdateMetadata = new AttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("redirectURIS");
            attributeUpdateMetadata.setUpdatedFieldValue(newTenant.getRedirectURIS().toString());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }

        if (!oldTenant.getContacts().equals(newTenant.getContacts())) {
            AttributeUpdateMetadata attributeUpdateMetadata = new AttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("contacts");
            attributeUpdateMetadata.setUpdatedFieldValue(newTenant.getContacts().toString());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }
        return metadataSet;
    }


    /**
     * create attribute update metadata from db entity
     * @param metadata
     * @return
     */
    public static TenantAttributeUpdateMetadata createAttributeUpdateMetadataFromEntity (AttributeUpdateMetadata metadata) {

        return TenantAttributeUpdateMetadata.newBuilder()
                 .setUpdatedAt(metadata.getUpdatedAt().toString())
                .setUpdatedBy(metadata.getUpdatedBy())
                .setUpdatedAttributeValue(metadata.getUpdatedFieldValue())
                .setUpdatedAttribute(metadata.getUpdatedFieldKey())
                .build();
    }



}
