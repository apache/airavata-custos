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

package org.apache.custos.user.profile.mapper;


import org.apache.custos.user.profile.persistance.model.UserAttribute;
import org.apache.custos.user.profile.persistance.model.UserProfile;
import org.apache.custos.user.profile.persistance.model.UserRole;
import org.apache.custos.user.profile.service.UserStatus;
import org.apache.custos.user.profile.utils.Constants;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.*;


/**
 * This class maps attributes between grpc UserProfile to DB UserProfile table
 */
public class UserProfileMapper {

    private static final Logger LOGGER = LoggerFactory.getLogger(UserProfileMapper.class);

    /**
     * Maps gRPC UserProfile Model to DB Layer UserProfile Entity
     *
     * @param {@link org.apache.custos.user.profile.service.UserProfile} tenant
     * @return Tenant
     */
    public static UserProfile createUserProfileEntityFromUserProfile(org.apache.custos.user.profile.service.UserProfile userProfile) {

        UserProfile entity = new UserProfile();

        entity.setUsername(userProfile.getUsername());
        entity.setEmailAddress(userProfile.getEmail());
        entity.setFirstName(userProfile.getFirstName());
        entity.setLastName(userProfile.getLastName());
        entity.setStatus(userProfile.getStatus().name());

        Set<UserAttribute> attributeSet = new HashSet<>();
        if (userProfile.getAttributesList() != null && !userProfile.getAttributesList().isEmpty()) {


            userProfile.getAttributesList().forEach(atr -> {
                if (atr.getValueList() != null && !atr.getValueList().isEmpty()) {
                    for (String value : atr.getValueList()) {
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
        if (userProfile.getRealmRolesList() != null && !userProfile.getRealmRolesList().isEmpty()) {


            userProfile.getRealmRolesList().forEach(role -> {
                UserRole userRole = new UserRole();
                userRole.setValue(role);
                userRole.setType(Constants.ROLE_TYPE_REALM);
                userRole.setUserProfile(entity);
                userRoleSet.add(userRole);
            });


        }

        if (userProfile.getClientRolesList() != null && !userProfile.getClientRolesList().isEmpty()) {
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
     * Transform UserProfileEntity to Tenant
     *
     * @param profileEntity
     * @return tenant
     */
    public static org.apache.custos.user.profile.service.UserProfile createUserProfileFromUserProfileEntity(UserProfile profileEntity) {


        org.apache.custos.user.profile.service.UserProfile.Builder builder =
                org.apache.custos.user.profile.service.UserProfile.newBuilder();


        if (profileEntity.getUserRole() != null && !profileEntity.getUserRole().isEmpty()) {

            profileEntity.getUserRole().forEach(role -> {
                if (role.getType().equals(Constants.ROLE_TYPE_CLIENT)) {
                    builder.addClientRoles(role.getValue());
                } else {
                    builder.addRealmRoles(role.getValue());
                }
            });
        }

        List<org.apache.custos.user.profile.service.UserAttribute> attributeList = new ArrayList<>();
        if (profileEntity.getUserAttribute() != null && !profileEntity.getUserAttribute().isEmpty()) {

            Map<String, List<String>> atrMap = new HashMap<>();

            profileEntity.getUserAttribute().forEach(atr -> {

                if (atrMap.get(atr.getKey()) == null) {
                    atrMap.put(atr.getKey(), new ArrayList<String>());
                }
                atrMap.get(atr.getKey()).add(atr.getValue());

            });


            atrMap.keySet().forEach(key -> {
                org.apache.custos.user.profile.service.UserAttribute attribute = org.apache.custos.user.profile.service
                        .UserAttribute
                        .newBuilder()
                        .setKey(key)
                        .addAllValue(atrMap.get(key))
                        .build();
                attributeList.add(attribute);
            });
        }


        return builder
                .setUsername(profileEntity.getUsername())
                .setEmail(profileEntity.getEmailAddress())
                .setFirstName(profileEntity.getFirstName())
                .setLastName(profileEntity.getLastName())
                .setCreatedAt(profileEntity.getCreatedAt().toString())
                .setLastModifiedAt(profileEntity.getLastModifiedAt() != null ? profileEntity.getLastModifiedAt().toString() : "")
                .setStatus(UserStatus.valueOf(profileEntity.getStatus()))
                .addAllAttributes(attributeList)
                .build();


    }


    public static String getUserInfoInfoAsString(org.apache.custos.user.profile.service.UserProfile userProfile) {
        StringBuffer buffer = new StringBuffer();
        buffer.append("username : " + userProfile.getUsername());
        buffer.append("\n");
        buffer.append("emailAddress : " + userProfile.getEmail());
        buffer.append("\n");
        buffer.append("firstName : " + userProfile.getFirstName());
        buffer.append("\n");
        buffer.append("lastName : " + userProfile.getLastName());
        buffer.append("\n");

        return buffer.toString();

    }


}



