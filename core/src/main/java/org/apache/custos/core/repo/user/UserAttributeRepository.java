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

import org.apache.custos.core.model.user.UserAttribute;
import org.apache.custos.core.model.user.UserProfile;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;

import java.util.List;
import java.util.Optional;

public interface UserAttributeRepository extends JpaRepository<UserAttribute, Long> {

    @Query("SELECT DISTINCT atr.userProfile from UserAttribute atr where atr.keyValue = ?1 and atr.value =?2")
    List<UserProfile> findFilteredUserProfiles(String key, String value);

    Optional<UserAttribute> findUserAttributeByKeyValueAndValueAndUserProfile(String keyValue, String value, UserProfile userProfile);
}
