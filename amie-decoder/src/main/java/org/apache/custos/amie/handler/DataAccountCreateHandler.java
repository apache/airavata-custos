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

/**
 * Handles the 'data_account_create' packet.
 * <p>
 * This packet is the third step in the account creation transaction.
 */
@Component
public class DataAccountCreateHandler implements PacketHandler {

    private static final Logger log = LoggerFactory.getLogger(DataAccountCreateHandler.class);

    private final AmieClient amieClient;

    public DataAccountCreateHandler(AmieClient amieClient) {
        this.amieClient = amieClient;
    }

    @Override
    public String supportsType() {
        return "data_account_create";
    }

    @Override
    public void handle(JsonNode packetJson, PacketEntity packetEntity) {
        log.info("Starting 'data_account_create' handler for packet amie_id [{}].", packetEntity.getAmieId());

        JsonNode body = packetJson.path("body");
        String projectId = body.path("ProjectID").asText();
        String personId = body.path("PersonID").asText();
        JsonNode dnList = body.path("DnList");

        Assert.hasText(projectId, "'ProjectID' must not be empty.");
        Assert.hasText(personId, "'PersonID' must not be empty.");
        log.info("Packet validated. ProjectID: [{}], PersonID: [{}].", projectId, personId);


        // TODO - perform the business logic
        //  - find the user's record by 'personId' (localID) and update the distinguished names (dnList)
        if (dnList.isArray() && !dnList.isEmpty()) {
            log.info("Received DnList for user [{}]. In a real implementation, this would be saved to the user's profile.", personId);
            // TODO userService.updateUserDnList(personId, dnList);
        }

        // Send the 'inform_transaction_complete' reply to close the transaction.
        sendSuccessReply(packetEntity.getAmieId());
    }

    private void sendSuccessReply(long packetRecId) {
        Map<String, Object> reply = new HashMap<>();
        Map<String, Object> itcBody = new HashMap<>();

        itcBody.put("StatusCode", "Success");
        itcBody.put("DetailCode", 1);
        itcBody.put("Message", "Transaction completed successfully by handler.");

        reply.put("type", "inform_transaction_complete");
        reply.put("body", itcBody);

        amieClient.replyToPacket(packetRecId, reply);

        log.info("Successfully sent 'inform_transaction_complete' for data_account_create packet_rec_id [{}].", packetRecId);
    }
}
