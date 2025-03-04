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
import org.apache.custos.core.model.user.UserRole;
import org.apache.custos.core.constants.Operation;
import java.util.HashSet;
import java.util.Set;

/**
 * This class maps attributes between protobuf TenantAttributeUpdateMetadata to DB AttributeUpdateMetadata table
 */
public class AttributeUpdateMetadataMapper {

    private static AttributeUpdateMetadata createAttribute(UserProfile prof, String fieldKey, String fieldValue, String updatedBy, Operation operation) {
        AttributeUpdateMetadata attributeUpdateMetadata = new AttributeUpdateMetadata();
        attributeUpdateMetadata.setUserProfile(prof);
        attributeUpdateMetadata.setUpdatedFieldKey(fieldKey);
        attributeUpdateMetadata.setUpdatedFieldValue(fieldValue);
        attributeUpdateMetadata.setUpdatedBy(updatedBy);
        attributeUpdateMetadata.setOperation(operation);
        return attributeUpdateMetadata;
    }

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

        if (!oldProf.getUsername().equals(newProf.getUsername())) {
            metadataSet.add(createAttribute(oldProf, "username", newProf.getUsername(), updatedBy, Operation.UPDATE));
        }

        if (!oldProf.getEmailAddress().equals(newProf.getEmailAddress())) {
            metadataSet.add(createAttribute(oldProf, "emailAddress", newProf.getEmailAddress(), updatedBy, Operation.UPDATE));
        }

        if (!oldProf.getFirstName().equals(newProf.getFirstName())) {
            metadataSet.add(createAttribute(oldProf, "firstName", newProf.getFirstName(), updatedBy, Operation.UPDATE));
        }

        if (!oldProf.getLastName().equals(newProf.getLastName())) {
            metadataSet.add(createAttribute(oldProf, "lastName", newProf.getLastName(), updatedBy, Operation.UPDATE));
        }

        if (!oldProf.getUserRole().equals(newProf.getUserRole())) {
            // find the differences between oldProf and newProf
            Set<UserRole> oldUserRoles = new HashSet<>(oldProf.getUserRole());
            Set<UserRole> newUserRoles = new HashSet<>(newProf.getUserRole());
            Set<UserRole> addedRoles = new HashSet<>();
            for (UserRole role : newUserRoles) {
                if (oldUserRoles.contains(role)) {
                    oldUserRoles.remove(role);
                } else {
                    addedRoles.add(role);
                }
            }

            for (UserRole role : addedRoles) {
                metadataSet.add(createAttribute(oldProf, role.getType() + "Roles", role.getValue(), updatedBy, Operation.CREATE));
            }

            for (UserRole role : oldUserRoles) {
                metadataSet.add(createAttribute(oldProf, role.getType() + "Roles", role.getValue(), updatedBy, Operation.DELETE));
            }
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
