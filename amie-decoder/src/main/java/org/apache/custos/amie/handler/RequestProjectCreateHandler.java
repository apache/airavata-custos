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

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Handles 'request_project_create' (RPC) by replying with 'notify_project_create' (NPC).
 */
@Component
public class RequestProjectCreateHandler implements PacketHandler {

    private static final Logger LOGGER = LoggerFactory.getLogger(RequestProjectCreateHandler.class);

    private final AmieClient amieClient;

    public RequestProjectCreateHandler(AmieClient amieClient) {
        this.amieClient = amieClient;
    }

    @Override
    public void handle(JsonNode packetJson, PacketEntity packetEntity) {

        // TODO - refactor the sanity checks into Packet Router (if all the packets follow the same style)
        if (packetJson == null) {
            LOGGER.error("packet json is null");
            throw new IllegalArgumentException("packetJson is null");
        }

        JsonNode body = packetJson.path("body");
        if (body.isMissingNode()) {
            LOGGER.error("body is null");
            throw new IllegalArgumentException("packetJson.body is missing");
        }

        String grantNumber = body.path("GrantNumber").asText("");
        String projectTitle = body.path("ProjectTitle").asText("");
        String startDate = body.path("StartDate").asText("");
        String endDate = body.path("EndDate").asText("");
        String recordId = body.path("RecordID").asText("");
        String piOrgCode = body.path("PiOrgCode").asText("");
        String pfosNumber = body.path("PfosNumber").asText("");
        String piGlobalID = body.path("PiGlobalID").asText("");

        // ResourceList will have only one resource (According to AMIE documentation)
        List<Object> resourceList = new ArrayList<>();
        JsonNode rl = body.path("ResourceList");
        if (rl.isArray()) {
            rl.forEach(node -> resourceList.add(node.asText()));
        } else {
            LOGGER.warn("No resource list found for amie_id [{}]", packetEntity.getAmieId());
        }

        // TODO - Derive a local project identifier
        String localProjectId = deriveLocalProjectId(grantNumber);

        // Build the NPC reply body
        Map<String, Object> replyBody = new HashMap<>();
        Map<String, Object> npc = new HashMap<>();
        npc.put("GrantNumber", grantNumber);
        npc.put("ProjectTitle", projectTitle);
        npc.put("StartDate", startDate);
        npc.put("EndDate", endDate);
        npc.put("RecordID", recordId);
        npc.put("ProjectID", localProjectId);
        npc.put("PiOrgCode", piOrgCode);
        npc.put("PfosNumber", pfosNumber);
        npc.put("PiGlobalID", piGlobalID);
        if (!resourceList.isEmpty()) {
            npc.put("ResourceList", resourceList);
        }

        replyBody.put("type", "notify_project_create");
        replyBody.put("body", npc);

        long packetRecId = packetEntity.getAmieId();
        amieClient.replyToPacket(packetRecId, replyBody);

        LOGGER.info("NPC sent for RPC packet_rec_id={}, LocalProjectID={}", packetRecId, localProjectId);
    }

    @Override
    public String supportsType() {
        return "request_project_create";
    }

    private String deriveLocalProjectId(String grantNumber) {
        // TODO - need to keep a DB mapping between LocalProjectID and GrantNumber
        String gn = (grantNumber == null || grantNumber.isBlank()) ? "UNKNOWN" : grantNumber.trim();
        return "PRJ-" + gn;
    }
}
