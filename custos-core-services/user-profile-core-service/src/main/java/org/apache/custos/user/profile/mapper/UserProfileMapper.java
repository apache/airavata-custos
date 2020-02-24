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


import org.apache.custos.user.profile.persistance.model.UserProfileEntity;


/**
 * This class maps attributes between grpc UserProfile to DB UserProfile table
 */
public class UserProfileMapper {


    /**
     * Maps gRPC UserProfile Model to DB Layer UserProfile Entity
     *
     * @param {@link org.apache.custos.user.profile.service.UserProfile} tenant
     * @return Tenant
     */
    public static UserProfileEntity createUserProfileEntityFromUserProfile(org.apache.custos.user.profile.service.UserProfile userProfile) {

        UserProfileEntity entity = new UserProfileEntity();

        entity.setId(userProfile.getUserId());
        entity.setUsername(userProfile.getUsername());
        entity.setEmailAddress(userProfile.getEmail());
        entity.setFirstName(userProfile.getFirstName());
        entity.setLastName(userProfile.getLastName());
        entity.setTenantId(userProfile.getTenantId());
        entity.setStatus(userProfile.getStatus().name());

        return entity;

    }


    /**
     * Transform UserProfileEntity to Tenant
     *
     * @param profileEntity
     * @return tenant
     */
    public static org.apache.custos.user.profile.service.UserProfile createUserProfileFromUserProfileEntity(UserProfileEntity profileEntity) {

        return org.apache.custos.user.profile.service.UserProfile.newBuilder()
                .setUsername(profileEntity.getUsername())
                .setEmail(profileEntity.getEmailAddress())
                .setFirstName(profileEntity.getFirstName())
                .setLastName(profileEntity.getLastName())
                .setTenantId(profileEntity.getTenantId())
                .setUserId(profileEntity.getId())
                .build();


    }


    public static String getUserInfoInfoAsString(org.apache.custos.user.profile.service.UserProfile userProfile) {
        StringBuffer buffer = new StringBuffer();
        buffer.append("username : " + userProfile.getUsername());
        buffer.append("\n");
        buffer.append("tenantId : " + userProfile.getTenantId());
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



