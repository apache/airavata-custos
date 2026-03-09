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


import org.apache.custos.core.model.user.StatusUpdateMetadata;
import org.apache.custos.core.model.user.UserProfile;
import org.apache.custos.core.user.profile.api.UserProfileStatusUpdateMetadata;
import org.apache.custos.core.user.profile.api.UserStatus;

import java.util.HashSet;
import java.util.Set;

/**
 * This class maps attributes of grpc UserProfileStatusUpdateMetadata to DB StatusUpdateMetadata table
 */
public class StatusUpdateMetadataMapper {

    /**
     * Creates a status update metadata entity for a user profile.
     *
     * @param tenant    The user profile to create the metadata for.
     * @param updatedBy The username of the user who updated the status.
     * @return A set of status update metadata containing the updated status, updated by, and user profile.
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
     * Creates a UserProfileStatusUpdateMetadata object from a StatusUpdateMetadata object.
     *
     * @param metadata The StatusUpdateMetadata object to convert.
     * @return A UserProfileStatusUpdateMetadata object containing the converted data.
     */
    public static UserProfileStatusUpdateMetadata createUserProfileStatusMetadataFrom(StatusUpdateMetadata metadata) {
        return UserProfileStatusUpdateMetadata.newBuilder()
                .setUpdatedAt(metadata.getUpdatedAt().toString())
                .setUpdatedBy(metadata.getUpdatedBy())
                .setUpdatedStatus(UserStatus.valueOf(metadata.getUpdatedStatus())).build();

    }

}
