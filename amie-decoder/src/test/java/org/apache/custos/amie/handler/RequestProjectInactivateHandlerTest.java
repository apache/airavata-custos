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
import org.apache.custos.amie.service.ProjectMembershipService;
import org.apache.custos.amie.service.ProjectService;
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
class RequestProjectInactivateHandlerTest {

    @Mock
    private AmieClient amieClient;

    @Mock
    private ProjectService projectService;

    @Mock
    private ProjectMembershipService membershipService;

    private RequestProjectInactivateHandler handler;
    private ObjectMapper objectMapper;

    @BeforeEach
    void setUp() {
        handler = new RequestProjectInactivateHandler(amieClient, projectService, membershipService);
        objectMapper = new ObjectMapper();
    }

    @Test
    void supportsType_shouldReturnCorrectType() {
        assertThat(handler.supportsType()).isEqualTo("request_project_inactivate");
    }

    @Test
    void handle_withValidPacket_shouldProcessSuccessfully() {
        JsonNode incomingPacket = JsonTestUtils.loadMockPacket("request_project_inactivate", "incoming-request.json");
        JsonNode expectedReply = JsonTestUtils.loadMockPacket("request_project_inactivate", "outgoing-notify.json");

        PacketEntity packetEntity = new PacketEntity();
        packetEntity.setAmieId(233497911L);
        packetEntity.setType("request_project_inactivate");

        handler.handle(incomingPacket, packetEntity);

        verify(projectService).inactivateProject("test-project-123");
        verify(membershipService).inactivateAllMembershipsForProject("test-project-123");

        @SuppressWarnings("unchecked")
        ArgumentCaptor<Map<String, Object>> replyCaptor = ArgumentCaptor.forClass(Map.class);
        verify(amieClient).replyToPacket(eq(233497911L), replyCaptor.capture());

        Map<String, Object> sentReply = replyCaptor.getValue();
        assertThat(sentReply).containsKey("type");
        assertThat(sentReply.get("type")).isEqualTo("notify_project_inactivate");
        assertThat(sentReply).containsKey("body");

        @SuppressWarnings("unchecked")
        Map<String, Object> sentBody = (Map<String, Object>) sentReply.get("body");
        JsonNode expectedBody = expectedReply.path("body");

        assertThat(sentBody).containsKey("ProjectID");
        assertThat(sentBody).containsKey("PersonID");
        assertThat(sentBody).containsKey("ResourceList");

        assertThat(sentBody.get("ProjectID")).isEqualTo(expectedBody.path("ProjectID").asText());
        assertThat(sentBody.get("PersonID")).isEqualTo(expectedBody.path("PersonID").asText());
    }

    @Test
    void handle_withMissingProjectId_shouldThrowException() {
        JsonNode packetJson = createPacketJsonWithMissingProjectIdField();
        PacketEntity packetEntity = createPacketEntity();

        assertThatThrownBy(() -> handler.handle(packetJson, packetEntity))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("'ProjectID' must not be empty");
    }

    @Test
    void handle_withEmptyProjectId_shouldThrowException() {
        JsonNode packetJson = createPacketJsonWithEmptyField("ProjectID");
        PacketEntity packetEntity = createPacketEntity();

        assertThatThrownBy(() -> handler.handle(packetJson, packetEntity))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("'ProjectID' must not be empty");
    }

    @Test
    void handle_withOptionalFields_shouldProcessSuccessfully() {
        JsonNode packetJson = createPacketJsonWithOptionalFields();
        PacketEntity packetEntity = createPacketEntity();

        handler.handle(packetJson, packetEntity);

        verify(projectService).inactivateProject("PRJ-TEST123");
        verify(membershipService).inactivateAllMembershipsForProject("PRJ-TEST123");
        //noinspection unchecked
        verify(amieClient).replyToPacket(eq(12345L), any(Map.class));
    }

    private JsonNode createPacketJsonWithMissingProjectIdField() {
        ObjectNode body = objectMapper.createObjectNode().put("ProjectID", "PRJ-TEST123");

        if (body.has("ProjectID")) {
            body.remove("ProjectID");
        }

        return objectMapper.createObjectNode().set("body", body);
    }

    private JsonNode createPacketJsonWithEmptyField(String fieldName) {
        return objectMapper.createObjectNode().set("body", objectMapper.createObjectNode().put(fieldName, ""));
    }

    private JsonNode createPacketJsonWithOptionalFields() {
        return objectMapper.createObjectNode().set("body", objectMapper.createObjectNode()
                .put("ProjectID", "PRJ-TEST123")
                .put("PersonID", "person-123")
                .put("ResourceList", "[]"));
    }

    private PacketEntity createPacketEntity() {
        PacketEntity entity = new PacketEntity();
        entity.setAmieId(12345L);
        entity.setType("request_project_inactivate");
        return entity;
    }
}
