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

import org.apache.custos.core.model.tenant.Tenant;
import org.apache.custos.core.model.tenant.TenantStatusUpdateMetadata;
import org.apache.custos.core.tenant.profile.api.TenantStatus;

import java.util.HashSet;
import java.util.Set;

/**
 * The StatusUpdateMetadataMapper class provides utility methods to create and convert status update metadata.
 */
public class StatusUpdateMetadataMapper {

    /**
     * Creates a set of TenantStatusUpdateMetadata entity based on the given parameters.
     *
     * @param tenant The Tenant object representing the tenant.
     * @param updatedBy The username of the user who updated the status.
     * @return A set of TenantStatusUpdateMetadata representing the status update metadata.
     */
    public static Set<TenantStatusUpdateMetadata> createStatusUpdateMetadataEntity(Tenant tenant, String updatedBy) {
        Set<TenantStatusUpdateMetadata> metaDataSet = new HashSet<>();
        TenantStatusUpdateMetadata metadata = new TenantStatusUpdateMetadata();
        metadata.setTenant(tenant);
        metadata.setUpdatedBy(updatedBy);
        metadata.setUpdatedStatus(tenant.getStatus());

        metaDataSet.add(metadata);
        return metaDataSet;
    }

    /**
     * Creates a new instance of {@link org.apache.custos.core.tenant.profile.api.TenantStatusUpdateMetadata}
     * by converting the given {@link TenantStatusUpdateMetadata} object.
     *
     * @param metadata The {@link TenantStatusUpdateMetadata} object to convert.
     * @return A new instance of {@link org.apache.custos.core.tenant.profile.api.TenantStatusUpdateMetadata}.
     */
    public static org.apache.custos.core.tenant.profile.api.TenantStatusUpdateMetadata createTenantStatusMetadataFrom(TenantStatusUpdateMetadata metadata) {
        return org.apache.custos.core.tenant.profile.api.TenantStatusUpdateMetadata.newBuilder()
                .setUpdatedAt(metadata.getUpdatedAt().toString())
                .setUpdatedBy(metadata.getUpdatedBy())
                .setUpdatedStatus(TenantStatus.valueOf(metadata.getUpdatedStatus())).build();
    }
}
