/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing,
 *  software distributed under the License is distributed on an
 *  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *  KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

package org.apache.custos.sharing.persistance.repository;

import org.apache.custos.sharing.persistance.model.Sharing;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

public interface SharingRepository extends JpaRepository<Sharing, String> {

    @Query(value = "select * from sharing s where s.tenant_id = ?1 and s.entity_id = ?2 " +
            "and s.sharing_type IN  ?3", nativeQuery = true)
    public List<Sharing> findSharingForEntityOfTenant(long tenantId, String entityId, List<String> sharingTypes);

    @Modifying
    @Transactional
    @Query(value = "delete  from sharing s where s.tenant_id = ?1 and s.entity_id = ?2 " +
            "and s.sharing_type = ?3", nativeQuery = true)
    public void removeGivenCascadingPermissionsForEntity(long tenantId, String entityId, String sharingType);


    @Query(value = "select * from sharing s where s.tenant_id = ?1 and s.entity_id = ?2 " +
            "and s.permission_type_id = ?3 and s.sharing_type IN ?4", nativeQuery = true)
    public List<Sharing> findAllByEntityAndSharingTypeAndPermissionType
            (long tenantId, String entityId, String permissionTypeId, List<String> sharingTypes);


    @Query(value = "select * from sharing s where s.tenant_id = ?1 and s.entity_id = ?2 " +
            "and s.permission_type_id = ?3 and s.associating_id_type = ?4", nativeQuery = true)
    public List<Sharing> findAllByEntityAndPermissionTypeAndOwnerType
            (long tenantId, String entityId, String permissionTypeId, String associatingIdType);

    @Query(value = "select * from sharing s where s.tenant_id = ?1 and s.entity_id = ?2 " +
            "and s.permission_type_id = ?3 and s.associating_id_type = ?4 and s.sharing_type IN ?5", nativeQuery = true)
    public List<Sharing> findAllByEntityAndPermissionTypeAndOwnerTypeAndSharingType
            (long tenantId, String entityId, String permissionTypeId, String associatingIdType, List<String> sharingList);



    @Transactional
    public List<Sharing> deleteAllByInheritedParentIdAndPermissionTypeIdAndTenantIdAndSharingTypeAndAssociatingId(
            String inheritedParentId, String permissionTypeId, long tenantId, String sharingType, String associatedId);


    @Transactional
    public void deleteAllByEntityIdAndPermissionTypeIdAndAssociatingIdAndTenantIdAndInheritedParentId(
            String entityId, String permissionTypeId, String associatingId, long tenantId, String inheritedParentId);

    @Query(value = "select * from sharing s where s.tenant_id = ?1 and s.entity_id = ?2 " +
            "and s.permission_type_id IN ?3 and s.associating_id IN  ?4", nativeQuery = true)
    public List<Sharing>  findAllSharingOfEntityForGroupsUnderPermissions(long tenantId, String entityId,
                                                                          List<String> permissionTypes,
                                                                          List<String> associatedIds);


    @Query(value = "select * from sharing s where s.tenant_id = ?1 and s.associating_id IN  ?2 " +
            "and s.entity_id  IN ?3", nativeQuery = true)
    public List<Sharing>  findAllSharingEntitiesForUsers(long tenantId,
                                                                          List<String> associatedIds,
                                                                          List<String> entityIds);




}
