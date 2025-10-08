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
package org.apache.custos.amie.handler;

import com.fasterxml.jackson.databind.JsonNode;
import org.apache.custos.amie.client.AmieClient;
import org.apache.custos.amie.model.PacketEntity;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;
import org.springframework.util.Assert;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Handles the 'request_account_inactivate' AMIE packet.
 * <p>
 * This transaction asks the local site to inactivate a user's association with a specific project.
 * Upon successful local processing, it sends a 'notify_account_inactivate' reply.
 */
@Component
public class RequestAccountInactivateHandler implements PacketHandler {

    private static final Logger LOGGER = LoggerFactory.getLogger(RequestAccountInactivateHandler.class);

    private final AmieClient amieClient;

    public RequestAccountInactivateHandler(AmieClient amieClient) {
        this.amieClient = amieClient;
    }

    @Override
    public String supportsType() {
        return "request_account_inactivate";
    }

    @Override
    public void handle(JsonNode packetJson, PacketEntity packetEntity) {
        LOGGER.info("Starting 'request_account_inactivate' handler for packet amie_id [{}].", packetEntity.getAmieId());

        JsonNode body = packetJson.path("body");
        String projectId = body.path("ProjectID").asText();
        String personId = body.path("PersonID").asText();

        Assert.hasText(projectId, "'ProjectID' must not be empty.");
        Assert.hasText(personId, "'PersonID' must not be empty.");
        LOGGER.info("Packet validated. Inactivating account for user [{}] on project [{}].", personId, projectId);

        // TODO - perform the business logic
        //  - find the user's project record and mark it as inactive
        //  - remove the user from project's Slurm account
        LOGGER.info("Simulating business logic: Removing user [{}] from project [{}]'s allocation.", personId, projectId);


        // Send the 'notify_account_inactivate' reply.
        Map<String, Object> replyBody = new HashMap<>();
        Map<String, Object> bodyContent = new HashMap<>();

        bodyContent.put("ProjectID", projectId);
        bodyContent.put("PersonID", personId);

        List<String> resourceList = new ArrayList<>();
        JsonNode rlNode = body.path("ResourceList");
        if (rlNode.isArray()) {
            rlNode.forEach(node -> resourceList.add(node.asText()));
        }
        bodyContent.put("ResourceList", resourceList);

        replyBody.put("type", "notify_account_inactivate");
        replyBody.put("body", bodyContent);

        amieClient.replyToPacket(packetEntity.getAmieId(), replyBody);

        LOGGER.info("Successfully completed 'request_account_inactivate' handler and sent reply for packet amie_id [{}].", packetEntity.getAmieId());
    }
}
