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

import java.util.HashMap;
import java.util.Map;
import java.util.UUID;

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

    public RequestAccountCreateHandler(AmieClient amieClient) {
        this.amieClient = amieClient;
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
        String userFirstName = body.path("UserFirstName").asText();
        String userLastName = body.path("UserLastName").asText();
        String userEmail = body.path("UserEmail").asText();
        String userOrgCode = body.path("UserOrgCode").asText();

        Assert.hasText(projectId, "'ProjectID' (the local project ID) must not be empty.");
        Assert.hasText(userFirstName, "'UserFirstName' must not be empty.");
        Assert.hasText(userLastName, "'UserLastName' must not be empty.");
        LOGGER.info("Packet validated successfully for user [{} {}] on project [{}].", userFirstName, userLastName, projectId);

        // TODO invoke actual cluster's user provisioning service. For the time being generating a local user ID and a username
        String localUserPersonId = UUID.randomUUID().toString();
        String localUsername = (userFirstName.trim().charAt(0) + userLastName.trim().replace(" ", "-")).toLowerCase();

        LOGGER.info("Created local user account with PersonID [{}] and username [{}]", localUserPersonId, localUsername);

        // Build and send the 'notify_account_create' reply
        Map<String, Object> replyBody = new HashMap<>();
        Map<String, Object> bodyContent = new HashMap<>();
        bodyContent.put("UserOrgCode", userOrgCode);
        bodyContent.put("ProjectID", projectId);
        bodyContent.put("UserPersonID", localUserPersonId);

        // User login name (Unix username) on the actual resource
        bodyContent.put("UserRemoteSiteLogin", localUsername);

        bodyContent.put("GrantNumber", grantNumber);

        replyBody.put("type", "notify_account_create");
        replyBody.put("body", bodyContent);

        amieClient.replyToPacket(packetEntity.getAmieId(), replyBody);

        LOGGER.info("Successfully completed 'request_account_create' handler and sent reply for packet amie_id [{}].", packetEntity.getAmieId());
    }
}
