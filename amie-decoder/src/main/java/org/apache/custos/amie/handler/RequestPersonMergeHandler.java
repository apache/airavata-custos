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
import org.apache.custos.amie.service.PersonService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.Assert;

import java.util.HashMap;
import java.util.Map;

/**
 * Handles the 'request_person_merge' AMIE packet.
 * <p>
 * This transaction is to merge two local user identities into one.
 * This identifies a "surviving" user and a "retiring" user. The handler must
 * re-associate all resources from the retiring user to the surviving user and then
 * disable the retiring user's account.
 * Upon successful processing, this handler sends an 'inform_transaction_complete' reply to close the transaction.
 */
@Component
public class RequestPersonMergeHandler implements PacketHandler {

    private static final Logger LOGGER = LoggerFactory.getLogger(RequestPersonMergeHandler.class);

    private final AmieClient amieClient;
    private final PersonService personService;

    public RequestPersonMergeHandler(AmieClient amieClient, PersonService personService) {
        this.amieClient = amieClient;
        this.personService = personService;
    }

    @Override
    public String supportsType() {
        return "request_person_merge";
    }

    @Override
    @Transactional
    public void handle(JsonNode packetJson, PacketEntity packetEntity) {
        LOGGER.info("Starting 'request_person_merge' handler for packet amie_id [{}].", packetEntity.getAmieId());

        JsonNode body = packetJson.path("body");
        String survivingPersonLocalId = body.path("KeepPersonID").asText();
        String survivingPersonGlobalId = body.path("KeepGlobalID").asText();
        String retiringPersonLocalId = body.path("DeletePersonID").asText();
        String retiringPersonGlobalId = body.path("DeleteGlobalID").asText();

        Assert.hasText(survivingPersonLocalId, "'KeepPersonID' (the surviving user's local ID) must not be empty.");
        Assert.hasText(retiringPersonLocalId, "'DeletePersonID' (the retiring user's local ID) must not be empty.");
        LOGGER.info("Packet validated. Merging user with local ID [{}], global ID [{}] into user with local ID [{}], global ID [{}].",
                retiringPersonLocalId, retiringPersonGlobalId, survivingPersonLocalId, survivingPersonGlobalId);

        personService.mergePersons(survivingPersonLocalId, retiringPersonLocalId);

        sendSuccessReply(packetEntity.getAmieId());
    }

    private void sendSuccessReply(long packetRecId) {
        Map<String, Object> reply = new HashMap<>();
        Map<String, Object> itcBody = new HashMap<>();

        itcBody.put("StatusCode", "Success");
        itcBody.put("DetailCode", 1);
        itcBody.put("Message", "Person merge transaction completed successfully.");

        reply.put("type", "inform_transaction_complete");
        reply.put("body", itcBody);

        amieClient.replyToPacket(packetRecId, reply);

        LOGGER.info("Successfully sent 'inform_transaction_complete' for person_merge packet_rec_id [{}].", packetRecId);
    }
}
