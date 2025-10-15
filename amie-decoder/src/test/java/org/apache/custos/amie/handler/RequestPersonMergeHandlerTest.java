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
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import org.apache.custos.amie.client.AmieClient;
import org.apache.custos.amie.model.PacketEntity;
import org.apache.custos.amie.service.PersonService;
import org.apache.custos.amie.util.JsonTestUtils;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.verify;

@ExtendWith(MockitoExtension.class)
@Tag("unit")
class RequestPersonMergeHandlerTest {

    @Mock
    private AmieClient amieClient;

    @Mock
    private PersonService personService;

    private RequestPersonMergeHandler handler;
    private ObjectMapper objectMapper;

    @BeforeEach
    void setUp() {
        handler = new RequestPersonMergeHandler(amieClient, personService);
        objectMapper = new ObjectMapper();
    }

    @Test
    void supportsType_shouldReturnCorrectType() {
        assertThat(handler.supportsType()).isEqualTo("request_person_merge");
    }

    @Test
    void handle_withValidPacket_shouldProcessSuccessfully() {
        JsonNode incomingPacket = JsonTestUtils.loadMockPacket("request_person_merge", "incoming-request.json");
        JsonNode expectedReply = JsonTestUtils.loadMockPacket("request_person_merge", "outgoing-inform.json");

        PacketEntity packetEntity = new PacketEntity();
        packetEntity.setAmieId(233497920L);
        packetEntity.setType("request_person_merge");

        handler.handle(incomingPacket, packetEntity);

        verify(personService).mergePersons("test-person-primary-123", "test-person-secondary-456");

        @SuppressWarnings("unchecked")
        ArgumentCaptor<Map<String, Object>> replyCaptor = ArgumentCaptor.forClass(Map.class);
        verify(amieClient).replyToPacket(eq(233497920L), replyCaptor.capture());

        Map<String, Object> sentReply = replyCaptor.getValue();
        assertThat(sentReply).containsKey("type");
        assertThat(sentReply.get("type")).isEqualTo("inform_transaction_complete");
        assertThat(sentReply).containsKey("body");

        @SuppressWarnings("unchecked")
        Map<String, Object> sentBody = (Map<String, Object>) sentReply.get("body");
        JsonNode expectedBody = expectedReply.path("body");

        assertThat(sentBody).containsKey("StatusCode");
        assertThat(sentBody).containsKey("DetailCode");
        assertThat(sentBody).containsKey("Message");

        assertThat(sentBody.get("StatusCode")).isEqualTo(expectedBody.path("StatusCode").asText());
        assertThat(sentBody.get("DetailCode")).isEqualTo(expectedBody.path("DetailCode").asInt());
        assertThat(sentBody.get("Message")).isEqualTo(expectedBody.path("Message").asText());
    }

    @Test
    void handle_withMissingKeepPersonId_shouldThrowException() {
        JsonNode packetJson = createPacketJsonWithMissingField("KeepPersonID");
        PacketEntity packetEntity = createPacketEntity();

        assertThatThrownBy(() -> handler.handle(packetJson, packetEntity))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("'KeepPersonID' (the surviving user's local ID) must not be empty");
    }

    @Test
    void handle_withMissingDeletePersonId_shouldThrowException() {
        JsonNode packetJson = createPacketJsonWithMissingField("DeletePersonID");
        PacketEntity packetEntity = createPacketEntity();

        assertThatThrownBy(() -> handler.handle(packetJson, packetEntity))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("'DeletePersonID' (the retiring user's local ID) must not be empty");
    }

    @Test
    void handle_withEmptyKeepPersonId_shouldThrowException() {
        JsonNode packetJson = createPacketJsonWithEmptyField("KeepPersonID");
        PacketEntity packetEntity = createPacketEntity();

        assertThatThrownBy(() -> handler.handle(packetJson, packetEntity))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("'KeepPersonID' (the surviving user's local ID) must not be empty");
    }

    @Test
    void handle_withEmptyDeletePersonId_shouldThrowException() {
        JsonNode packetJson = createPacketJsonWithEmptyField("DeletePersonID");
        PacketEntity packetEntity = createPacketEntity();

        assertThatThrownBy(() -> handler.handle(packetJson, packetEntity))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("'DeletePersonID' (the retiring user's local ID) must not be empty");
    }

    @Test
    void handle_withOptionalGlobalIds_shouldProcessSuccessfully() {
        JsonNode packetJson = createPacketJsonWithGlobalIds();
        PacketEntity packetEntity = createPacketEntity();

        handler.handle(packetJson, packetEntity);

        verify(personService).mergePersons("person-surviving", "person-retiring");
        //noinspection unchecked
        verify(amieClient).replyToPacket(eq(12345L), any(Map.class));
    }

    private JsonNode createPacketJsonWithMissingField(String missingField) {
        ObjectNode body = objectMapper.createObjectNode()
                .put("KeepPersonID", "person-surviving")
                .put("DeletePersonID", "person-retiring");

        if (body.has(missingField)) {
            body.remove(missingField);
        }

        return objectMapper.createObjectNode().set("body", body);
    }

    private JsonNode createPacketJsonWithEmptyField(String fieldName) {
        ObjectNode body = objectMapper.createObjectNode()
                .put("KeepPersonID", "person-surviving")
                .put("DeletePersonID", "person-retiring");

        body.put(fieldName, "");

        return objectMapper.createObjectNode().set("body", body);
    }

    private JsonNode createPacketJsonWithGlobalIds() {
        return objectMapper.createObjectNode()
                .set("body", objectMapper.createObjectNode()
                        .put("KeepPersonID", "person-surviving")
                        .put("KeepGlobalID", "GLOBAL123")
                        .put("DeletePersonID", "person-retiring")
                        .put("DeleteGlobalID", "GLOBAL456"));
    }

    private PacketEntity createPacketEntity() {
        PacketEntity entity = new PacketEntity();
        entity.setAmieId(12345L);
        entity.setType("request_person_merge");
        return entity;
    }
}
