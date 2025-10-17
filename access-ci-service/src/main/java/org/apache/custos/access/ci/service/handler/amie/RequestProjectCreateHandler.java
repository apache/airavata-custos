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
import com.fasterxml.jackson.databind.node.ObjectNode;
import org.apache.custos.access.ci.service.client.amie.AmieClient;
import org.apache.custos.access.ci.service.model.ClusterAccountEntity;
import org.apache.custos.access.ci.service.model.PersonEntity;
import org.apache.custos.access.ci.service.model.ProjectEntity;
import org.apache.custos.access.ci.service.model.amie.PacketEntity;
import org.apache.custos.access.ci.service.service.PersonService;
import org.apache.custos.access.ci.service.service.ProjectMembershipService;
import org.apache.custos.access.ci.service.service.ProjectService;
import org.apache.custos.access.ci.service.service.UserAccountService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;
import org.springframework.util.Assert;

import java.util.HashMap;
import java.util.Map;

/**
 * Handles 'request_project_create' (RPC) by replying with 'notify_project_create' (NPC).
 */
@Component
public class RequestProjectCreateHandler implements PacketHandler {

    private static final Logger LOGGER = LoggerFactory.getLogger(RequestProjectCreateHandler.class);

    private final AmieClient amieClient;
    private final PersonService personService;
    private final UserAccountService userAccountService;
    private final ProjectService projectService;
    private final ProjectMembershipService membershipService;

    public RequestProjectCreateHandler(AmieClient amieClient, PersonService personService, UserAccountService userAccountService,
                                       ProjectService projectService, ProjectMembershipService membershipService) {
        this.amieClient = amieClient;
        this.personService = personService;
        this.userAccountService = userAccountService;
        this.projectService = projectService;
        this.membershipService = membershipService;
    }

    @Override
    public String supportsType() {
        return "request_project_create";
    }

    @Override
    public void handle(JsonNode packetJson, PacketEntity packetEntity) throws Exception {
        LOGGER.info("Starting 'request_project_create' handler for packet amie_id [{}].", packetEntity.getAmieId());

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

        String grantNumber = body.path("GrantNumber").asText();
        String piGlobalId = body.path("PiGlobalID").asText();
        String piFirstName = body.path("PiFirstName").asText();
        String piLastName = body.path("PiLastName").asText();

        Assert.hasText(grantNumber, "'GrantNumber' must not be empty.");
        Assert.hasText(piGlobalId, "'PiGlobalID' must not be empty.");
        Assert.hasText(piFirstName, "'PiFirstName' must not be empty.");
        Assert.hasText(piLastName, "'PiLastName' must not be empty.");
        LOGGER.info("Packet validated for GrantNumber [{}] and PI Global ID [{}].", grantNumber, piGlobalId);

        // Find or Create Person record for the PI
        ObjectNode piAsUserNode = createPiAsUserNode(body);
        PersonEntity piPerson = personService.findOrCreatePersonFromPacket(piAsUserNode);
        LOGGER.info("PI person record exists with local ID [{}].", piPerson.getId());

        ClusterAccountEntity piClusterAccount = userAccountService.provisionClusterAccount(piPerson);
        LOGGER.info("Provisioned cluster account for PI [{}] with username [{}].", piPerson.getId(), piClusterAccount.getUsername());

        String localProjectId = "PRJ-" + grantNumber;
        ProjectEntity project = projectService.createOrFindProject(localProjectId, grantNumber);
        LOGGER.info("Project [{}] exists.", project.getId());

        membershipService.createMembership(project.getId(), piClusterAccount.getId(), "PI");
        LOGGER.info("Created 'PI' membership for cluster account [{}] on project [{}].", piClusterAccount.getId(), project.getId());

        sendSuccessReply(packetEntity.getAmieId(), body, project.getId(), piPerson.getId(), piClusterAccount.getUsername());
        LOGGER.info("Successfully completed 'request_project_create' handler and sent reply for packet amie_id [{}].", packetEntity.getAmieId());
    }

    private void sendSuccessReply(long packetRecId, JsonNode originalBody, String localProjectId, String localPiPersonId, String localPiUsername) {
        Map<String, Object> reply = new HashMap<>();
        Map<String, Object> npcBody = new HashMap<>();

        npcBody.put("ProjectID", localProjectId);
        npcBody.put("GrantNumber", originalBody.path("GrantNumber").asText());
        npcBody.put("PiPersonID", localPiPersonId);
        npcBody.put("PiRemoteSiteLogin", localPiUsername);
        npcBody.put("PiGlobalID", originalBody.path("PiGlobalID").asText());
        npcBody.put("ProjectTitle", originalBody.path("ProjectTitle").asText(null));
        npcBody.put("ResourceList", originalBody.path("ResourceList"));

        reply.put("type", "notify_project_create");
        reply.put("body", npcBody);

        amieClient.replyToPacket(packetRecId, reply);
    }

    private ObjectNode createPiAsUserNode(JsonNode rpcBody) {
        ObjectNode userNode = rpcBody.deepCopy();
        userNode.put("UserGlobalID", rpcBody.path("PiGlobalID").asText());
        userNode.put("UserFirstName", rpcBody.path("PiFirstName").asText());
        userNode.put("UserLastName", rpcBody.path("PiLastName").asText());
        userNode.put("UserEmail", rpcBody.path("PiEmail").asText());
        userNode.put("UserOrganization", rpcBody.path("PiOrganization").asText());
        userNode.put("UserOrgCode", rpcBody.path("PiOrgCode").asText());
        userNode.put("NsfStatusCode", rpcBody.path("NsfStatusCode").asText());
        userNode.set("UserDnList", rpcBody.path("PiDnList"));
        return userNode;
    }
}
