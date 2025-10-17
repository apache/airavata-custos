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
package org.apache.custos.access.ci.service.handler.amie;

import com.fasterxml.jackson.databind.JsonNode;
import org.apache.custos.access.ci.service.client.amie.AmieClient;
import org.apache.custos.access.ci.service.model.amie.PacketEntity;
import org.apache.custos.access.ci.service.service.ProjectMembershipService;
import org.apache.custos.access.ci.service.service.ProjectService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;
import org.springframework.util.Assert;

import java.util.HashMap;
import java.util.Map;

/**
 * Handles the 'request_project_inactivate' AMIE packet
 */
@Component
public class RequestProjectInactivateHandler implements PacketHandler {

    private static final Logger LOGGER = LoggerFactory.getLogger(RequestProjectInactivateHandler.class);

    private final AmieClient amieClient;
    private final ProjectService projectService;
    private final ProjectMembershipService membershipService;

    public RequestProjectInactivateHandler(AmieClient amieClient, ProjectService projectService, ProjectMembershipService membershipService) {
        this.amieClient = amieClient;
        this.projectService = projectService;
        this.membershipService = membershipService;
    }

    @Override
    public String supportsType() {
        return "request_project_inactivate";
    }

    @Override
    public void handle(JsonNode packetJson, PacketEntity packetEntity) {
        LOGGER.info("Starting 'request_project_inactivate' handler for packet amie_id [{}].", packetEntity.getAmieId());

        JsonNode body = packetJson.path("body");
        String projectId = body.path("ProjectID").asText();

        Assert.hasText(projectId, "'ProjectID' must not be empty.");
        LOGGER.info("Packet validated. ProjectID to inactivate: [{}].", projectId);

        projectService.inactivateProject(projectId);
        membershipService.inactivateAllMembershipsForProject(projectId);
        LOGGER.info("Inactivated project [{}] and all associated memberships.", projectId);

        sendSuccessReply(packetEntity.getAmieId(), body);

        LOGGER.info("Successfully completed 'request_project_inactivate' handler and sent reply for packet amie_id [{}].", packetEntity.getAmieId());
    }

    private void sendSuccessReply(long packetRecId, JsonNode originalBody) {
        Map<String, Object> reply = new HashMap<>();
        Map<String, Object> bodyContent = new HashMap<>();

        bodyContent.put("ProjectID", originalBody.path("ProjectID").asText());
        bodyContent.put("PersonID", originalBody.path("PersonID").asText(null));
        bodyContent.put("ResourceList", originalBody.path("ResourceList"));

        reply.put("type", "notify_project_inactivate");
        reply.put("body", bodyContent);

        amieClient.replyToPacket(packetRecId, reply);
    }
}
