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
import org.apache.custos.access.ci.service.model.ClusterAccountEntity;
import org.apache.custos.access.ci.service.model.PersonEntity;
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
 * Handles the 'request_account_create' (RAC) AMIE packet.
 * <p>
 * This transaction asks the local site to create an account for a user on a project.
 * This includes creating a local user account with the specified project (a Unix account) if one does not exist.
 * Upon successful processing sends a 'notify_account_create' (NAC) reply back to AMIE.
 */
@Component
public class RequestAccountCreateHandler implements PacketHandler {

    private static final Logger LOGGER = LoggerFactory.getLogger(RequestAccountCreateHandler.class);

    private final AmieClient amieClient;
    private final PersonService personService;
    private final UserAccountService userAccountService;
    private final ProjectService projectService;
    private final ProjectMembershipService membershipService;

    public RequestAccountCreateHandler(AmieClient amieClient, PersonService personService,
                                       UserAccountService userAccountService, ProjectService projectService,
                                       ProjectMembershipService membershipService) {
        this.amieClient = amieClient;
        this.personService = personService;
        this.userAccountService = userAccountService;
        this.projectService = projectService;
        this.membershipService = membershipService;
    }

    @Override
    public String supportsType() {
        return "request_account_create";
    }

    @Override
    public void handle(JsonNode packetJson, PacketEntity packetEntity) throws Exception {
        LOGGER.info("Starting 'request_account_create' handler for packet amie_id [{}].", packetEntity.getAmieId());

        JsonNode body = packetJson.path("body");
        String projectId = body.path("ProjectID").asText();
        String grantNumber = body.path("GrantNumber").asText();
        String userGlobalId = body.path("UserGlobalID").asText();

        Assert.hasText(projectId, "'ProjectID' (the local project ID) must not be empty.");
        Assert.hasText(grantNumber, "'GrantNumber' must not be empty.");
        Assert.hasText(userGlobalId, "'UserGlobalID' must not be empty.");
        LOGGER.info("Packet validated for UserGlobalID [{}] on project [{}].", userGlobalId, projectId);

        PersonEntity person = personService.findOrCreatePersonFromPacket(body);
        LOGGER.info("Ensured person record exists with local ID [{}].", person.getId());

        ClusterAccountEntity clusterAccount = userAccountService.provisionClusterAccount(person);
        LOGGER.info("Provisioned new cluster account [{}] with username [{}].", clusterAccount.getId(), clusterAccount.getUsername());

        projectService.createOrFindProject(projectId, grantNumber);
        LOGGER.info("Ensured project [{}] exists.", projectId);

        membershipService.createMembership(projectId, clusterAccount.getId(), "USER");
        LOGGER.info("Created 'USER' membership for cluster account [{}] on project [{}].", clusterAccount.getId(), projectId);

        sendSuccessReply(packetEntity.getAmieId(), body, person.getId(), clusterAccount.getUsername());
        LOGGER.info("Successfully completed 'request_account_create' handler and sent reply for packet amie_id [{}].", packetEntity.getAmieId());
    }

    private void sendSuccessReply(long packetRecId, JsonNode originalBody, String localPersonId, String localUsername) {
        Map<String, Object> reply = new HashMap<>();
        Map<String, Object> bodyContent = new HashMap<>();

        bodyContent.put("ProjectID", originalBody.path("ProjectID").asText());
        bodyContent.put("GrantNumber", originalBody.path("GrantNumber").asText());
        bodyContent.put("UserPersonID", localPersonId);
        bodyContent.put("UserRemoteSiteLogin", localUsername);

        bodyContent.put("UserOrgCode", originalBody.path("UserOrgCode").asText(null));
        bodyContent.put("ResourceList", originalBody.path("ResourceList"));

        reply.put("type", "notify_account_create");
        reply.put("body", bodyContent);

        amieClient.replyToPacket(packetRecId, reply);
    }
}
