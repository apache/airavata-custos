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

package org.apache.custos.user.profile.persistance.repository;

import org.apache.custos.user.profile.persistance.model.UserGroupMembership;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;

public interface GroupMembershipRepository extends JpaRepository<UserGroupMembership, String> {


    public List<UserGroupMembership> findAllByGroupId(String id);

    public List<UserGroupMembership> findAllByUserProfileId(String id);

    public List<UserGroupMembership> findAllByGroupIdAndUserProfileId(String groupEntityId, String userProfileId);


}
