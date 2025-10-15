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
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package org.apache.custos.amie.handler;

import com.fasterxml.jackson.databind.JsonNode;
import org.apache.custos.amie.client.AmieClient;
import org.apache.custos.amie.model.PacketEntity;
import org.apache.custos.amie.service.ProjectMembershipService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;
import org.springframework.util.Assert;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Handles the 'request_account_reactivate' AMIE packet.
 * <p>
 * This transaction asks the local site to reactivate a user's association with a specific project.
 * Upon successful local processing, it sends a 'notify_account_reactivate' reply.
 */
@Component
public class RequestAccountReactivateHandler implements PacketHandler {

    private static final Logger LOGGER = LoggerFactory.getLogger(RequestAccountReactivateHandler.class);

    private final AmieClient amieClient;
    private final ProjectMembershipService membershipService;

    public RequestAccountReactivateHandler(AmieClient amieClient, ProjectMembershipService membershipService) {
        this.amieClient = amieClient;
        this.membershipService = membershipService;
    }

    @Override
    public String supportsType() {
        return "request_account_reactivate";
    }

    @Override
    public void handle(JsonNode packetJson, PacketEntity packetEntity) {
        LOGGER.info("Starting 'request_account_reactivate' handler for packet amie_id [{}].", packetEntity.getAmieId());

        JsonNode body = packetJson.path("body");
        String projectId = body.path("ProjectID").asText();
        String personId = body.path("PersonID").asText();

        Assert.hasText(projectId, "'ProjectID' must not be empty.");
        Assert.hasText(personId, "'PersonID' must not be empty.");
        LOGGER.info("Packet validated. Reactivating account for user [{}] on project [{}].", personId, projectId);

        membershipService.reactivateMembershipsByPersonAndProject(projectId, personId);

        sendSuccessReply(packetEntity.getAmieId(), body);

        LOGGER.info("Successfully completed 'request_account_reactivate' handler and sent reply for packet amie_id [{}].", packetEntity.getAmieId());
    }

    private void sendSuccessReply(long packetRecId, JsonNode body) {
        Map<String, Object> reply = new HashMap<>();
        Map<String, Object> bodyContent = new HashMap<>();

        bodyContent.put("ProjectID", body.path("ProjectID").asText());
        bodyContent.put("PersonID", body.path("PersonID").asText());

        List<String> resourceList = new ArrayList<>();
        JsonNode rlNode = body.path("ResourceList");
        if (rlNode.isArray()) {
            rlNode.forEach(node -> resourceList.add(node.asText()));
        }
        bodyContent.put("ResourceList", resourceList);

        reply.put("type", "notify_account_reactivate");
        reply.put("body", bodyContent);

        amieClient.replyToPacket(packetRecId, reply);
    }
}
