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
class RequestUserModifyHandlerTest {

    @Mock
    private AmieClient amieClient;

    @Mock
    private PersonService personService;

    private RequestUserModifyHandler handler;
    private ObjectMapper objectMapper;

    @BeforeEach
    void setUp() {
        handler = new RequestUserModifyHandler(amieClient, personService);
        objectMapper = new ObjectMapper();
    }

    @Test
    void supportsType_shouldReturnCorrectType() {
        assertThat(handler.supportsType()).isEqualTo("request_user_modify");
    }

    @Test
    void handle_withReplaceAction_shouldProcessSuccessfully() throws Exception {
        JsonNode incomingPacket = JsonTestUtils.loadMockPacket("request_user_modify_replace", "incoming-request.json");
        JsonNode expectedReply = JsonTestUtils.loadMockPacket("request_user_modify_replace", "outgoing-inform.json");

        PacketEntity packetEntity = new PacketEntity();
        packetEntity.setAmieId(233497921L);
        packetEntity.setType("request_user_modify");

        handler.handle(incomingPacket, packetEntity);

        verify(personService).replaceFromModifyPacket(any(JsonNode.class));

        @SuppressWarnings("unchecked")
        ArgumentCaptor<Map<String, Object>> replyCaptor = ArgumentCaptor.forClass(Map.class);
        verify(amieClient).replyToPacket(eq(233497921L), replyCaptor.capture());

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
    void handle_withDeleteAction_shouldProcessSuccessfully() throws Exception {
        JsonNode incomingPacket = JsonTestUtils.loadMockPacket("request_user_modify_delete", "incoming-request.json");
        JsonNode expectedReply = JsonTestUtils.loadMockPacket("request_user_modify_delete", "outgoing-inform.json");

        PacketEntity packetEntity = new PacketEntity();
        packetEntity.setAmieId(233497922L);
        packetEntity.setType("request_user_modify");

        handler.handle(incomingPacket, packetEntity);

        verify(personService).deleteFromModifyPacket(any(JsonNode.class));

        @SuppressWarnings("unchecked")
        ArgumentCaptor<Map<String, Object>> replyCaptor = ArgumentCaptor.forClass(Map.class);
        verify(amieClient).replyToPacket(eq(233497922L), replyCaptor.capture());

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
    void handle_withUpperCaseAction_shouldProcessSuccessfully() throws Exception {
        JsonNode packetJson = createValidPacketJson("REPLACE");
        PacketEntity packetEntity = createPacketEntity();

        handler.handle(packetJson, packetEntity);

        verify(personService).replaceFromModifyPacket(any(JsonNode.class));
        //noinspection unchecked
        verify(amieClient).replyToPacket(eq(12345L), any(Map.class));
    }

    @Test
    void handle_withMissingActionType_shouldThrowException() {
        JsonNode packetJson = createPacketJsonWithMissingField("ActionType");
        PacketEntity packetEntity = createPacketEntity();

        assertThatThrownBy(() -> handler.handle(packetJson, packetEntity))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("'ActionType' must not be empty (replace|delete)");
    }

    @Test
    void handle_withEmptyActionType_shouldThrowException() {
        JsonNode packetJson = createPacketJsonWithEmptyField("ActionType");
        PacketEntity packetEntity = createPacketEntity();

        assertThatThrownBy(() -> handler.handle(packetJson, packetEntity))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("'ActionType' must not be empty (replace|delete)");
    }

    @Test
    void handle_withUnsupportedActionType_shouldThrowException() {
        JsonNode packetJson = createValidPacketJson("unsupported");
        PacketEntity packetEntity = createPacketEntity();

        assertThatThrownBy(() -> handler.handle(packetJson, packetEntity))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("Unsupported ActionType: unsupported");
    }

    private JsonNode createValidPacketJson(String actionType) {
        return objectMapper.createObjectNode()
                .set("body", objectMapper.createObjectNode()
                        .put("ActionType", actionType)
                        .put("UserGlobalID", "USER123456")
                        .put("UserFirstName", "John")
                        .put("UserLastName", "Doe")
                        .put("UserEmail", "john.doe@example.com"));
    }

    private JsonNode createPacketJsonWithMissingField(String missingField) {
        ObjectNode body = objectMapper.createObjectNode()
                .put("ActionType", "replace")
                .put("UserGlobalID", "USER123456")
                .put("UserFirstName", "John")
                .put("UserLastName", "Doe")
                .put("UserEmail", "john.doe@example.com");

        if (body.has(missingField)) {
            body.remove(missingField);
        }

        return objectMapper.createObjectNode().set("body", body);
    }

    private JsonNode createPacketJsonWithEmptyField(String fieldName) {
        ObjectNode body = objectMapper.createObjectNode()
                .put("ActionType", "replace")
                .put("UserGlobalID", "USER123456")
                .put("UserFirstName", "John")
                .put("UserLastName", "Doe")
                .put("UserEmail", "john.doe@example.com");

        body.put(fieldName, "");

        return objectMapper.createObjectNode().set("body", body);
    }

    private PacketEntity createPacketEntity() {
        PacketEntity entity = new PacketEntity();
        entity.setAmieId(12345L);
        entity.setType("request_user_modify");
        return entity;
    }
}
