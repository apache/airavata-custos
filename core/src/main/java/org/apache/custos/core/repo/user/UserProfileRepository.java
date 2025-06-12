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

package org.apache.custos.core.repo.user;

import org.apache.custos.core.model.user.UserProfile;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

import java.util.List;

public interface UserProfileRepository extends JpaRepository<UserProfile, String> {

    List<UserProfile> findByTenantId(long tenantId);

    @Query(value = "SELECT * FROM user_profile t WHERE t.tenant_id LIKE ?1 ORDER BY t.id limit ?2 OFFSET ?3", nativeQuery = true)
    List<UserProfile> findByTenantIdWithPagination(long tenantId, long limit, long offset);

    @Query(value = "SELECT * FROM user_profile u " +
            "WHERE u.tenant_id = :tenantId " +
            "AND LOWER(u.first_name) LIKE LOWER(CONCAT('%', :firstName, '%')) " +
            "ORDER BY u.id " +
            "LIMIT :limit OFFSET :offset", nativeQuery = true)
    List<UserProfile> searchUserProfilesByFirstName(@Param("tenantId") long tenantId,
                                        @Param("firstName") String firstName,
                                        @Param("limit") int limit,
                                        @Param("offset") int offset);

    @Query(value = "SELECT * FROM user_profile u " +
            "WHERE u.tenant_id = :tenantId " +
            "AND LOWER(u.last_name) LIKE LOWER(CONCAT('%', :lastName, '%')) " +
            "ORDER BY u.id " +
            "LIMIT :limit OFFSET :offset", nativeQuery = true)
    List<UserProfile> searchUserProfilesByLastName(@Param("tenantId") long tenantId,
                                       @Param("lastName") String lastName,
                                       @Param("limit") int limit,
                                       @Param("offset") int offset);

    @Query(value = "SELECT * FROM user_profile u " +
            "WHERE u.tenant_id = :tenantId " +
            "AND LOWER(u.username) LIKE LOWER(CONCAT('%', :username, '%')) " +
            "ORDER BY u.id " +
            "LIMIT :limit OFFSET :offset", nativeQuery = true)
    List<UserProfile> searchUserProfilesByUsername(@Param("tenantId") long tenantId,
                                       @Param("username") String username,
                                       @Param("limit") int limit,
                                       @Param("offset") int offset);

    // 🔹 Get Total Count for First Name Search
    @Query(value = "SELECT COUNT(*) FROM user_profile u " +
            "WHERE u.tenant_id = :tenantId " +
            "AND LOWER(u.first_name) LIKE LOWER(CONCAT('%', :firstName, '%'))", nativeQuery = true)
    long countByFirstName(@Param("tenantId") long tenantId,
                          @Param("firstName") String firstName);

    // 🔹 Get Total Count for Last Name Search
    @Query(value = "SELECT COUNT(*) FROM user_profile u " +
            "WHERE u.tenant_id = :tenantId " +
            "AND LOWER(u.last_name) LIKE LOWER(CONCAT('%', :lastName, '%'))", nativeQuery = true)
    long countByLastName(@Param("tenantId") long tenantId,
                         @Param("lastName") String lastName);

    // 🔹 Get Total Count for Username Search
    @Query(value = "SELECT COUNT(*) FROM user_profile u " +
            "WHERE u.tenant_id = :tenantId " +
            "AND LOWER(u.username) LIKE LOWER(CONCAT('%', :username, '%'))", nativeQuery = true)
    long countByUsername(@Param("tenantId") long tenantId,
                         @Param("username") String username);

}
