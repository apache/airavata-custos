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

package org.apache.custos.core.mapper.user;

import org.apache.custos.core.model.user.AttributeUpdateMetadata;
import org.apache.custos.core.model.user.UserProfile;
import org.apache.custos.core.user.profile.api.UserProfileAttributeUpdateMetadata;

import java.util.HashSet;
import java.util.Objects;
import java.util.Set;

/**
 * This class maps attributes between protobuf TenantAttributeUpdateMetadata to DB AttributeUpdateMetadata table
 */
public class AttributeUpdateMetadataMapper {


    /**
     * Creates a set of AttributeUpdateMetadata objects based on the changes made to a UserProfile object.
     *
     * @param oldProf   The old UserProfile object.
     * @param newProf   The new UserProfile object.
     * @param updatedBy The user who updated the UserProfile.
     * @return A set of AttributeUpdateMetadata objects representing the updated fields in the UserProfile.
     */
    public static Set<AttributeUpdateMetadata> createAttributeUpdateMetadataEntity(UserProfile oldProf, UserProfile newProf, String updatedBy) {

        Set<AttributeUpdateMetadata> metadataSet = new HashSet<>();
        if (!(Objects.equals(oldProf.getTenantId(), newProf.getTenantId()))) {
            AttributeUpdateMetadata attributeUpdateMetadata = new AttributeUpdateMetadata();
            attributeUpdateMetadata.setUserProfile(newProf);
            attributeUpdateMetadata.setUpdatedFieldKey("tenantId");
            attributeUpdateMetadata.setUpdatedFieldValue(String.valueOf(newProf.getTenantId()));
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            metadataSet.add(attributeUpdateMetadata);
        }

        if (!oldProf.getUsername().equals(newProf.getUsername())) {
            AttributeUpdateMetadata attributeUpdateMetadata = new AttributeUpdateMetadata();
            attributeUpdateMetadata.setUserProfile(newProf);
            attributeUpdateMetadata.setUpdatedFieldKey("username");
            attributeUpdateMetadata.setUpdatedFieldValue(newProf.getUsername());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            metadataSet.add(attributeUpdateMetadata);
        }

        if (!oldProf.getEmailAddress().equals(newProf.getEmailAddress())) {
            AttributeUpdateMetadata attributeUpdateMetadata = new AttributeUpdateMetadata();
            attributeUpdateMetadata.setUserProfile(newProf);
            attributeUpdateMetadata.setUpdatedFieldKey("emailAddress");
            attributeUpdateMetadata.setUpdatedFieldValue(newProf.getEmailAddress());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            metadataSet.add(attributeUpdateMetadata);
        }

        if (!oldProf.getFirstName().equals(newProf.getFirstName())) {
            AttributeUpdateMetadata attributeUpdateMetadata = new AttributeUpdateMetadata();
            attributeUpdateMetadata.setUserProfile(newProf);
            attributeUpdateMetadata.setUpdatedFieldKey("firstName");
            attributeUpdateMetadata.setUpdatedFieldValue(newProf.getFirstName());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            metadataSet.add(attributeUpdateMetadata);
        }

        if (!oldProf.getLastName().equals(newProf.getLastName())) {
            AttributeUpdateMetadata attributeUpdateMetadata = new AttributeUpdateMetadata();
            attributeUpdateMetadata.setUserProfile(newProf);
            attributeUpdateMetadata.setUpdatedFieldKey("lastName");
            attributeUpdateMetadata.setUpdatedFieldValue(newProf.getLastName());
            attributeUpdateMetadata.setUpdatedBy(updatedBy);
            metadataSet.add(attributeUpdateMetadata);
        }

        return metadataSet;
    }


    /**
     * Creates a UserProfileAttributeUpdateMetadata object from an AttributeUpdateMetadata object.
     *
     * @param metadata The AttributeUpdateMetadata object to convert.
     * @return A UserProfileAttributeUpdateMetadata object with values populated from the metadata parameter.
     */
    public static UserProfileAttributeUpdateMetadata createAttributeUpdateMetadataFromEntity(AttributeUpdateMetadata metadata) {
        return UserProfileAttributeUpdateMetadata.newBuilder()
                .setUpdatedAt(metadata.getUpdatedAt().toString())
                .setUpdatedBy(metadata.getUpdatedBy())
                .setUpdatedAttributeValue(metadata.getUpdatedFieldValue())
                .setUpdatedAttribute(metadata.getUpdatedFieldKey())
                .build();
    }
}
