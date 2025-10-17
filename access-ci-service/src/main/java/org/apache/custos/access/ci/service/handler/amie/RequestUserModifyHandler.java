/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
package org.apache.custos.access.ci.service.handler.amie;

import com.fasterxml.jackson.databind.JsonNode;
import org.apache.custos.access.ci.service.client.amie.AmieClient;
import org.apache.custos.access.ci.service.model.amie.PacketEntity;
import org.apache.custos.access.ci.service.service.PersonService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;
import org.springframework.util.Assert;

import java.util.HashMap;
import java.util.Map;

/**
 * Handles the 'request_user_modify' AMIE packet (ActionType replace | delete).
 */
@Component
public class RequestUserModifyHandler implements PacketHandler {

    private static final Logger LOGGER = LoggerFactory.getLogger(RequestUserModifyHandler.class);

    private final AmieClient amieClient;
    private final PersonService personService;

    public RequestUserModifyHandler(AmieClient amieClient, PersonService personService) {
        this.amieClient = amieClient;
        this.personService = personService;
    }

    @Override
    public String supportsType() {
        return "request_user_modify";
    }

    @Override
    public void handle(JsonNode packetJson, PacketEntity packetEntity) throws Exception {
        LOGGER.info("Starting 'request_user_modify' handler for packet amie_id [{}].", packetEntity.getAmieId());

        JsonNode body = packetJson.path("body");
        String actionType = body.path("ActionType").asText(null);
        Assert.hasText(actionType, "'ActionType' must not be empty (replace|delete).");

        switch (actionType.toLowerCase()) {
            case "replace":
                personService.replaceFromModifyPacket(body);
                break;
            case "delete":
                personService.deleteFromModifyPacket(body);
                break;
            default:
                throw new IllegalArgumentException("Unsupported ActionType: " + actionType);
        }

        sendSuccessReply(packetEntity.getAmieId());
    }

    private void sendSuccessReply(long packetRecId) {
        Map<String, Object> reply = new HashMap<>();
        Map<String, Object> itcBody = new HashMap<>();

        itcBody.put("StatusCode", "Success");
        itcBody.put("DetailCode", 1);
        itcBody.put("Message", "Transaction completed successfully");

        reply.put("type", "inform_transaction_complete");
        reply.put("body", itcBody);

        amieClient.replyToPacket(packetRecId, reply);

        LOGGER.info("Successfully sent 'inform_transaction_complete' for 'request_user_modify' packet_rec_id [{}].", packetRecId);
    }
}


