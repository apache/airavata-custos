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

import org.apache.custos.core.constants.Constants;
import org.apache.custos.core.model.user.UserAttribute;
import org.apache.custos.core.model.user.UserProfile;
import org.apache.custos.core.model.user.UserRole;
import org.apache.custos.core.user.profile.api.UserStatus;
import org.apache.custos.core.user.profile.api.UserTypes;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;

/**
 * This class maps attributes between protobuf UserProfile to UserProfile model
 */
public class UserProfileMapper {

    /**
     * Creates a UserProfile entity from a UserProfile protobuf
     *
     * @param userProfile the UserProfile object to be converted
     * @return the created UserProfile entity
     */
    public static UserProfile createUserProfileEntityFromUserProfile(org.apache.custos.core.user.profile.api.UserProfile userProfile) {

        UserProfile entity = new UserProfile();

        entity.setUsername(userProfile.getUsername());
        userProfile.getEmail();
        if (!userProfile.getEmail().trim().isEmpty()) {
            entity.setEmailAddress(userProfile.getEmail());
        }

        userProfile.getFirstName();
        if (!userProfile.getFirstName().trim().isEmpty()) {
            entity.setFirstName(userProfile.getFirstName());
        }

        userProfile.getLastName();
        if (!userProfile.getLastName().trim().isEmpty()) {
            entity.setLastName(userProfile.getLastName());
        }

        entity.setType(userProfile.getType().name());

        entity.setStatus(userProfile.getStatus().name());

        Set<UserAttribute> attributeSet = new HashSet<>();
        if (!userProfile.getAttributesList().isEmpty()) {


            userProfile.getAttributesList().forEach(atr -> {
                if (!atr.getValuesList().isEmpty()) {
                    for (String value : atr.getValuesList()) {
                        UserAttribute userAttribute = new UserAttribute();
                        userAttribute.setKey(atr.getKey());
                        userAttribute.setValue(value);
                        userAttribute.setUserProfile(entity);
                        attributeSet.add(userAttribute);
                    }
                }
            });
        }

        entity.setUserAttribute(attributeSet);
        Set<UserRole> userRoleSet = new HashSet<>();
        if (!userProfile.getRealmRolesList().isEmpty()) {
            userProfile.getRealmRolesList().forEach(role -> {
                UserRole userRole = new UserRole();
                userRole.setValue(role);
                userRole.setType(Constants.ROLE_TYPE_REALM);
                userRole.setUserProfile(entity);
                userRoleSet.add(userRole);
            });
        }

        if (!userProfile.getClientRolesList().isEmpty()) {
            userProfile.getClientRolesList().forEach(role -> {
                UserRole userRole = new UserRole();
                userRole.setValue(role);
                userRole.setType(Constants.ROLE_TYPE_CLIENT);
                userRole.setUserProfile(entity);
                userRoleSet.add(userRole);
            });
        }
        entity.setUserRole(userRoleSet);

        return entity;
    }


    /**
     * Creates a protobuf UserProfile object from a UserProfileEntity object.
     *
     * @param profileEntity  The UserProfileEntity object to be converted.
     * @param membershipType The membership type of the user.
     * @return The created UserProfile object.
     */
    public static org.apache.custos.core.user.profile.api.UserProfile createUserProfileFromUserProfileEntity(UserProfile profileEntity, String membershipType) {
        org.apache.custos.core.user.profile.api.UserProfile.Builder builder = org.apache.custos.core.user.profile.api.UserProfile.newBuilder();

        if (profileEntity.getUserRole() != null && !profileEntity.getUserRole().isEmpty()) {

            profileEntity.getUserRole().forEach(role -> {
                if (role.getType().equals(Constants.ROLE_TYPE_CLIENT)) {
                    builder.addClientRoles(role.getValue());
                } else {
                    builder.addRealmRoles(role.getValue());
                }
            });
        }

        List<org.apache.custos.core.user.profile.api.UserAttribute> attributeList = new ArrayList<>();
        if (profileEntity.getUserAttribute() != null && !profileEntity.getUserAttribute().isEmpty()) {

            Map<String, List<String>> atrMap = new HashMap<>();
            profileEntity.getUserAttribute().forEach(atr -> {
                atrMap.computeIfAbsent(atr.getKey(), k -> new ArrayList<>());
                atrMap.get(atr.getKey()).add(atr.getValue());
            });

            atrMap.keySet().forEach(key -> {
                org.apache.custos.core.user.profile.api.UserAttribute attribute = org.apache.custos.core.user.profile.api.UserAttribute
                        .newBuilder()
                        .setKey(key)
                        .addAllValues(atrMap.get(key))
                        .build();
                attributeList.add(attribute);
            });
        }

        builder.setUsername(profileEntity.getUsername())
                .setCreatedAt(profileEntity.getCreatedAt().getTime())
                .setLastModifiedAt(profileEntity.getLastModifiedAt() != null ? profileEntity.getLastModifiedAt().getTime() : 0)
                .setStatus(UserStatus.valueOf(profileEntity.getStatus()))
                .addAllAttributes(attributeList);


        if (profileEntity.getEmailAddress() != null) {
            builder.setEmail(profileEntity.getEmailAddress());
        }

        if (profileEntity.getFirstName() != null) {
            builder.setFirstName(profileEntity.getFirstName());
        }

        if (profileEntity.getLastName() != null) {
            builder.setLastName(profileEntity.getLastName());
        }

        if (membershipType != null) {
            builder.setMembershipType(membershipType);
        }

        if (profileEntity.getType() == null) {
            builder.setType(UserTypes.END_USER);
        } else {
            builder.setType(UserTypes.valueOf(profileEntity.getType()));
        }

        return builder.build();
    }

    public static String getUserInfoInfoAsString(org.apache.custos.core.user.profile.api.UserProfile userProfile) {
        return "username : " + userProfile.getUsername() +
                "\n" +
                "emailAddress : " + userProfile.getEmail() +
                "\n" +
                "firstName : " + userProfile.getFirstName() +
                "\n" +
                "lastName : " + userProfile.getLastName() +
                "\n";
    }
}
