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
import org.apache.custos.core.model.tenant.TenantAttributeUpdateMetadata;

import java.util.HashSet;
import java.util.Set;

/**
 * This class is responsible for creating and mapping attribute update metadata for a Tenant object.
 */
public class AttributeUpdateMetadataMapper {

    /**
     * Creates a set of TenantAttributeUpdateMetadata entities based on the differences between the old and new Tenant objects.
     *
     * @param oldTenant The old Tenant object.
     * @param newTenant The new Tenant object.
     * @param updatedBy The user who updated the Tenant object.
     * @return A set of TenantAttributeUpdateMetadata entities representing the attribute updates.
     */
    public static Set<TenantAttributeUpdateMetadata> createAttributeUpdateMetadataEntity(Tenant oldTenant, Tenant newTenant, String updatedBy) {

        Set<TenantAttributeUpdateMetadata> metadataSet = new HashSet<>();
        if (!oldTenant.getName().equals(newTenant.getName())) {
            TenantAttributeUpdateMetadata attributeUpdateMetadata = new TenantAttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("name");
            attributeUpdateMetadata.setUpdatedFieldValue(newTenant.getName());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }

        if (!oldTenant.getAdminEmail().equals(newTenant.getAdminEmail())) {
            TenantAttributeUpdateMetadata attributeUpdateMetadata = new TenantAttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("adminEmail");
            attributeUpdateMetadata.setUpdatedFieldValue(newTenant.getAdminEmail());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }

        if (!oldTenant.getAdminFirstName().equals(newTenant.getAdminFirstName())) {
            TenantAttributeUpdateMetadata attributeUpdateMetadata = new TenantAttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("adminFirstName");
            attributeUpdateMetadata.setUpdatedFieldValue(newTenant.getAdminFirstName());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }

        if (!oldTenant.getAdminLastName().equals(newTenant.getAdminLastName())) {
            TenantAttributeUpdateMetadata attributeUpdateMetadata = new TenantAttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("adminLastName");
            attributeUpdateMetadata.setUpdatedFieldValue(newTenant.getAdminLastName());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }

        if (!oldTenant.getDomain().equals(newTenant.getDomain())) {
            TenantAttributeUpdateMetadata attributeUpdateMetadata = new TenantAttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("domain");
            attributeUpdateMetadata.setUpdatedFieldValue(newTenant.getDomain());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }

        if (!oldTenant.getLogoURI().equals(newTenant.getLogoURI())) {
            TenantAttributeUpdateMetadata attributeUpdateMetadata = new TenantAttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("logoURI");
            attributeUpdateMetadata.setUpdatedFieldValue(newTenant.getLogoURI());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }

        if (!oldTenant.getRequesterEmail().equals(newTenant.getRequesterEmail())) {
            TenantAttributeUpdateMetadata attributeUpdateMetadata = new TenantAttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("requesterEmail");
            attributeUpdateMetadata.setUpdatedFieldValue(newTenant.getRequesterEmail());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }


        if (!oldTenant.getScope().equals(newTenant.getScope())) {
            TenantAttributeUpdateMetadata attributeUpdateMetadata = new TenantAttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("scope");
            attributeUpdateMetadata.setUpdatedFieldValue(newTenant.getScope());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }

        Set diff = difference(oldTenant.getRedirectURIS(), newTenant.getRedirectURIS());

        if (!diff.isEmpty()) {
            TenantAttributeUpdateMetadata attributeUpdateMetadata = new TenantAttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("redirectURIS");
            attributeUpdateMetadata.setUpdatedFieldValue(setToString(diff));
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }

        Set conDiff = difference(oldTenant.getContacts(), newTenant.getContacts());
        if (!conDiff.isEmpty()) {
            TenantAttributeUpdateMetadata attributeUpdateMetadata = new TenantAttributeUpdateMetadata();
            attributeUpdateMetadata.setTenant(newTenant);
            attributeUpdateMetadata.setUpdatedFieldKey("contacts");
            attributeUpdateMetadata.setUpdatedFieldValue(setToString(conDiff));
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            attributeUpdateMetadata.setTenant(newTenant);
            metadataSet.add(attributeUpdateMetadata);
        }
        return metadataSet;
    }

    /**
     * Creates a TenantAttributeUpdateMetadata object based on the given TenantAttributeUpdateMetadata entity.
     *
     * @param metadata The TenantAttributeUpdateMetadata entity to create the TenantAttributeUpdateMetadata object from.
     * @return A TenantAttributeUpdateMetadata object representing the attribute update metadata.
     */
    public static org.apache.custos.core.tenant.profile.api.TenantAttributeUpdateMetadata createAttributeUpdateMetadataFromEntity(TenantAttributeUpdateMetadata metadata) {

        return org.apache.custos.core.tenant.profile.api.TenantAttributeUpdateMetadata.newBuilder()
                .setUpdatedAt(metadata.getUpdatedAt().toString())
                .setUpdatedBy(metadata.getUpdatedBy())
                .setUpdatedAttributeValue(metadata.getUpdatedFieldValue())
                .setUpdatedAttribute(metadata.getUpdatedFieldKey())
                .build();
    }

    private static <T> Set<T> difference(final Set<T> setOne, final Set<T> setTwo) {
        Set<T> result = new HashSet<T>(setOne);
        result.removeIf(setTwo::contains);
        return result;
    }

    private static String setToString(Set<Object> result) {
        StringBuilder builder = new StringBuilder();
        if (!result.isEmpty()) {
            for (Object t : result) {
                if (t instanceof RedirectURI) {
                    builder.append(((RedirectURI) t).getRedirectURI());
                } else {
                    builder.append(((Contact) t).getContactInfo());
                }
                builder.append(",");
            }
        }
        return builder.toString();
    }
}
