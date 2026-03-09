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
import org.apache.custos.access.ci.service.model.amie.PacketEntity;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

/**
 * Handles the 'inform_transaction_complete' packet.
 * <p>
 * This is a terminal packet that declares the end of a transaction. This requires no reply.
 */
@Component
public class InformTransactionCompleteHandler implements PacketHandler {

    private static final Logger LOGGER = LoggerFactory.getLogger(InformTransactionCompleteHandler.class);

    @Override
    public String supportsType() {
        return "inform_transaction_complete";
    }

    @Override
    public void handle(JsonNode packetJson, PacketEntity packetEntity) {
        // This packet is purely informational and completes the transaction
        JsonNode body = packetJson.path("body");
        String statusCode = body.path("StatusCode").asText("Unknown");
        String message = body.path("Message").asText("");

        LOGGER.info("Received 'inform_transaction_complete' for packet amie_id [{}]. Status: [{}], Message: [{}]. Transaction is now closed.",
                packetEntity.getAmieId(), statusCode, message);
    }
}
