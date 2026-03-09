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
import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.custos.access.ci.service.model.amie.PacketEntity;
import org.apache.custos.access.ci.service.util.JsonTestUtils;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

@ExtendWith(MockitoExtension.class)
@Tag("unit")
class InformTransactionCompleteHandlerTest {

    private InformTransactionCompleteHandler handler;
    private ObjectMapper objectMapper;

    @BeforeEach
    void setUp() {
        handler = new InformTransactionCompleteHandler();
        objectMapper = new ObjectMapper();
    }

    @Test
    void supportsType_shouldReturnCorrectType() {
        assertThat(handler.supportsType()).isEqualTo("inform_transaction_complete");
    }

    @Test
    void handle_withValidPacket_shouldProcessSuccessfully() {
        JsonNode incomingPacket = JsonTestUtils.loadMockPacket("inform_transaction_complete", "incoming-inform.json");

        PacketEntity packetEntity = new PacketEntity();
        packetEntity.setAmieId(233497913L);
        packetEntity.setType("inform_transaction_complete");

        handler.handle(incomingPacket, packetEntity);
    }

    @Test
    void handle_withMinimalPacket_shouldProcessSuccessfully() {
        JsonNode packetJson = createMinimalPacketJson();
        PacketEntity packetEntity = createPacketEntity();

        handler.handle(packetJson, packetEntity);
    }

    @Test
    void handle_withNullPacketJson_shouldThrowException() {
        PacketEntity packetEntity = createPacketEntity();

        //noinspection DataFlowIssue
        assertThatThrownBy(() -> handler.handle(null, packetEntity)).isInstanceOf(NullPointerException.class);
    }

    @Test
    void handle_withNullPacketEntity_shouldThrowException() {
        JsonNode packetJson = createValidPacketJson();

        //noinspection DataFlowIssue
        assertThatThrownBy(() -> handler.handle(packetJson, null)).isInstanceOf(NullPointerException.class);
    }

    private JsonNode createValidPacketJson() {
        return objectMapper.createObjectNode().set("body", objectMapper.createObjectNode()
                .put("StatusCode", "Success")
                .put("DetailCode", 1)
                .put("Message", "Transaction completed successfully"));
    }

    private JsonNode createMinimalPacketJson() {
        return objectMapper.createObjectNode().set("body", objectMapper.createObjectNode());
    }

    private PacketEntity createPacketEntity() {
        PacketEntity entity = new PacketEntity();
        entity.setAmieId(12345L);
        entity.setType("inform_transaction_complete");
        return entity;
    }
}
