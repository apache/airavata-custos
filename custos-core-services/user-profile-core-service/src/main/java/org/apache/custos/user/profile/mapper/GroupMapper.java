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

import org.apache.custos.user.profile.persistance.model.Group;
import org.apache.custos.user.profile.persistance.model.GroupAttribute;
import org.apache.custos.user.profile.persistance.model.GroupRole;
import org.apache.custos.user.profile.persistance.model.GroupToGroupMembership;
import org.apache.custos.user.profile.utils.Constants;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.*;

/**
 * This class maps persistence data model to gRPC data model
 */
public class GroupMapper {

    private static final Logger LOGGER = LoggerFactory.getLogger(GroupMapper.class);


    public static Group createGroupEntity(org.apache.custos.user.profile.service.Group group, long tenantId) {

        Group groupEntity = new Group();
        String id = group.getId() + "@" + tenantId;

        String parentId = group.getParentId();

        if (parentId != null && !parentId.trim().equals("")) {
            parentId = group.getParentId() + "@" + tenantId;
        }

        groupEntity.setId(id);
        groupEntity.setExternalId(group.getId());
        groupEntity.setName(group.getName());
        groupEntity.setTenantId(tenantId);
        groupEntity.setParentId(parentId);



        if (group.getDescription() != null && !group.getDescription().trim().equals("")) {
            groupEntity.setDescription(group.getDescription());
        }

        if (!group.getAttributesList().isEmpty()) {

            Set<GroupAttribute> groupList = new HashSet<>();
            group.getAttributesList().forEach(atr -> {

                for (String value : atr.getValueList()) {
                    GroupAttribute groupAttribute = new GroupAttribute();
                    groupAttribute.setKeyValue(atr.getKey());
                    groupAttribute.setValue(value);
                    groupAttribute.setGroup(groupEntity);
                    groupList.add(groupAttribute);

                }

            });

            groupEntity.setGroupAttribute(groupList);

        }

        Set<GroupRole> groupRoles = new HashSet<>();

        if (!group.getClientRolesList().isEmpty()) {


            group.getClientRolesList().forEach(role -> {

                GroupRole userRole = new GroupRole();
                userRole.setValue(role);
                userRole.setType(Constants.ROLE_TYPE_CLIENT);
                userRole.setGroup(groupEntity);
                groupRoles.add(userRole);

            });


        }


        if (!group.getRealmRolesList().isEmpty()) {

            group.getRealmRolesList().forEach(role -> {

                GroupRole userRole = new GroupRole();
                userRole.setValue(role);
                userRole.setType(Constants.ROLE_TYPE_REALM);
                userRole.setGroup(groupEntity);
                groupRoles.add(userRole);

            });

        }

        groupEntity.setGroupRole(groupRoles);


        return groupEntity;
    }


    public static org.apache.custos.user.profile.service.Group createGroup(Group group, String ownerId) {

        org.apache.custos.user.profile.service.Group.Builder groupBuilder = org.apache.custos.user.profile.service.Group
                .newBuilder()
                .setId(group.getExternalId())
                .setName(group.getName())
                .setParentId(group.getParentId())
                .setCreatedTime(group.getCreatedAt().getTime())
                .setLastModifiedTime(group.getLastModifiedAt().getTime())
                .setOwnerId(ownerId);


        if (group.getDescription() != null) {
            groupBuilder.setDescription(group.getDescription());
        }

        List<String> clientRoles = new ArrayList<>();
        List<String> realmRoles = new ArrayList<>();

        if (group.getGroupRole() != null && !group.getGroupRole().isEmpty()) {

            group.getGroupRole().forEach(role -> {
                if (role.getType().equals(Constants.ROLE_TYPE_CLIENT)) {
                    clientRoles.add(role.getValue());
                } else {
                    realmRoles.add(role.getValue());
                }
            });
        }

        List<org.apache.custos.user.profile.service.GroupAttribute> attributeList = new ArrayList<>();
        Map<String, List<String>> atrMap = new HashMap<>();
        if (group.getGroupAttribute() != null && !group.getGroupAttribute().isEmpty()) {

            group.getGroupAttribute().forEach(atr -> {

                if (atrMap.get(atr.getKeyValue()) == null) {
                    atrMap.put(atr.getKeyValue(), new ArrayList<String>());
                }
                atrMap.get(atr.getKeyValue()).add(atr.getValue());

            });

            groupBuilder = groupBuilder.addAllClientRoles(clientRoles).addAllRealmRoles(realmRoles);
        }

        atrMap.keySet().forEach(key -> {
            org.apache.custos.user.profile.service.GroupAttribute attribute = org.apache.custos.user.profile.service
                    .GroupAttribute
                    .newBuilder()
                    .setKey(key)
                    .addAllValue(atrMap.get(key))
                    .build();
            attributeList.add(attribute);
        });

        return groupBuilder.addAllAttributes(attributeList).build();

    }


    public static Group setParentGroupMembership(Group parent, Group child) {

        GroupToGroupMembership groupToGroupMembership = new GroupToGroupMembership();
        groupToGroupMembership.setChild(child);
        groupToGroupMembership.setParent(parent);
        groupToGroupMembership.setTenantId(child.getTenantId());


        Set<GroupToGroupMembership> groupList = new HashSet<>();
        groupList.add(groupToGroupMembership);
        child.setParentGroups(groupList);
        return child;

    }

    public static GroupToGroupMembership groupToGroupMembership(Group child, Group parent) {
        GroupToGroupMembership groupToGroupMembership = new GroupToGroupMembership();
        groupToGroupMembership.setChild(child);
        groupToGroupMembership.setParent(parent);
        groupToGroupMembership.setTenantId(child.getTenantId());
        return groupToGroupMembership;

    }

}
