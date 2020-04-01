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

package org.apache.custos.agent.profile.mapper;

import org.apache.custos.agent.profile.persistance.model.Agent;
import org.apache.custos.agent.profile.persistance.model.AgentAttribute;
import org.apache.custos.agent.profile.persistance.model.AgentRole;
import org.apache.custos.agent.profile.service.AgentStatus;
import org.apache.custos.agent.profile.utils.Constants;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.*;

/**
 * Mapping class of service models to persistence model
 */
public class AgentMapper {

    private static final Logger LOGGER = LoggerFactory.getLogger(AgentMapper.class);


    public static Agent createAgent(org.apache.custos.agent.profile.service.Agent agent, long tenantId) {

        Agent persistenceModel = new Agent();

        persistenceModel.setId(agent.getId());
        persistenceModel.setTenantId(tenantId);
        persistenceModel.setStatus(agent.getStatus().name());


        Set<AgentAttribute> attributeSet = new HashSet<>();
        if (agent.getAttributesList() != null && !agent.getAttributesList().isEmpty()) {

            agent.getAttributesList().forEach(atr -> {
                if (atr.getValueList() != null && !atr.getValueList().isEmpty()) {
                    for (String value : atr.getValueList()) {
                        AgentAttribute userAttribute = new AgentAttribute();
                        userAttribute.setKeyValue(atr.getKey());
                        userAttribute.setValue(value);
                        userAttribute.setAgent(persistenceModel);
                        attributeSet.add(userAttribute);
                    }
                }
            });
        }
        persistenceModel.setAgentAttribute(attributeSet);
        Set<AgentRole> userRoleSet = new HashSet<>();
        if (agent.getRolesList() != null && !agent.getRolesList().isEmpty()) {

            agent.getRolesList().forEach(role -> {
                AgentRole userRole = new AgentRole();
                userRole.setValue(role);
                userRole.setType(Constants.ROLE_TYPE_REALM);
                userRole.setAgent(persistenceModel);
                userRoleSet.add(userRole);
            });
        }

        persistenceModel.setAgentRole(userRoleSet);

        return persistenceModel;
    }


    public static org.apache.custos.agent.profile.service.Agent createAgent(Agent agent) {

        org.apache.custos.agent.profile.service.Agent serviceAgent =
                org.apache.custos.agent.profile.service.Agent
                        .newBuilder()
                        .setId(agent.getId())
                        .setCreatedAt(agent.getCreatedAt().getTime())
                        .setLastModifiedAt(agent.getLast_modified_at().getTime())
                        .setStatus(AgentStatus.valueOf(agent.getStatus()))
                        .build();

        List<org.apache.custos.agent.profile.service.AgentAttribute> attributeList = new ArrayList<>();
        if (agent.getAgentAttribute() != null && !agent.getAgentAttribute().isEmpty()) {

            Map<String, List<String>> atrMap = new HashMap<>();

            agent.getAgentAttribute().forEach(atr -> {

                if (atrMap.get(atr.getKeyValue()) == null) {
                    atrMap.put(atr.getKeyValue(), new ArrayList<String>());
                }
                atrMap.get(atr.getKeyValue()).add(atr.getValue());

            });


            atrMap.keySet().forEach(key -> {
                org.apache.custos.agent.profile.service.AgentAttribute attribute = org.apache.custos.agent.profile.service
                        .AgentAttribute
                        .newBuilder()
                        .setKey(key)
                        .addAllValue(atrMap.get(key))
                        .build();
                attributeList.add(attribute);
            });
        }

        List<String> roleList = new ArrayList<>();
        if (agent.getAgentRole() != null && !agent.getAgentRole().isEmpty()) {

            agent.getAgentRole().forEach(role -> {
                if (role.getType().equals(Constants.ROLE_TYPE_CLIENT)) {
                    roleList.add(role.getValue());
                }
            });
        }

        serviceAgent = serviceAgent.toBuilder()
                .addAllRoles(roleList)
                .addAllAttributes(attributeList)
                .build();


        return serviceAgent;

    }


}
