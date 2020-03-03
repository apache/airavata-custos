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


import org.apache.custos.user.profile.persistance.model.StatusUpdateMetadata;
import org.apache.custos.user.profile.persistance.model.UserProfile;
import org.apache.custos.user.profile.service.UserProfileStatusUpdateMetadata;
import org.apache.custos.user.profile.service.UserStatus;

import java.util.HashSet;
import java.util.Set;

/**
 * This class maps attributes of grpc UserProfileStatusUpdateMetadata to DB StatusUpdateMetadata table
 */
public class StatusUpdateMetadataMapper {


    /**
     * Creates Status update entity to save in DB
     *
     * @param {@link    org.apache.custos.user.profile.UserProfile} profile
     * @param updatedBy
     * @return
     */
    public static Set<StatusUpdateMetadata> createStatusUpdateMetadataEntity(UserProfile tenant, String updatedBy) {

        Set<StatusUpdateMetadata> metaDataSet = new HashSet<>();

        StatusUpdateMetadata metadata = new StatusUpdateMetadata();
        metadata.setUserProfile(tenant);
        metadata.setUpdatedBy(updatedBy);
        metadata.setUpdatedStatus(tenant.getStatus());

        metaDataSet.add(metadata);
        return metaDataSet;

    }

    /**
     * convert TenantStatusUpdateMetadataEntity to gRPC TenantStatusUpdateMetadata
     *
     * @param metadata
     * @return TenantStatusUpdateMetadata
     */
    public static UserProfileStatusUpdateMetadata createUserProfileStatusMetadataFrom(StatusUpdateMetadata metadata) {
        return UserProfileStatusUpdateMetadata.newBuilder()
                .setUpdatedAt(metadata.getUpdatedAt().toString())
                .setUpdatedBy(metadata.getUpdatedBy())
                .setUpdatedStatus(UserStatus.valueOf(metadata.getUpdatedStatus())).build();

    }

}
