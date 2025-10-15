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
import org.apache.custos.amie.model.ClusterAccountEntity;
import org.apache.custos.amie.model.PacketEntity;
import org.apache.custos.amie.model.PersonEntity;
import org.apache.custos.amie.model.ProjectEntity;
import org.apache.custos.amie.service.PersonService;
import org.apache.custos.amie.service.ProjectMembershipService;
import org.apache.custos.amie.service.ProjectService;
import org.apache.custos.amie.service.UserAccountService;
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
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
@Tag("unit")
class RequestProjectCreateHandlerTest {

    @Mock
    private AmieClient amieClient;

    @Mock
    private PersonService personService;

    @Mock
    private UserAccountService userAccountService;

    @Mock
    private ProjectService projectService;

    @Mock
    private ProjectMembershipService membershipService;

    private RequestProjectCreateHandler handler;
    private ObjectMapper objectMapper;

    @BeforeEach
    void setUp() {
        handler = new RequestProjectCreateHandler(amieClient, personService, userAccountService, projectService, membershipService);
        objectMapper = new ObjectMapper();
    }

    @Test
    void supportsType_shouldReturnCorrectType() {
        assertThat(handler.supportsType()).isEqualTo("request_project_create");
    }

    @Test
    void handle_withValidPacket_shouldProcessSuccessfully() throws Exception {
        JsonNode incomingPacket = JsonTestUtils.loadMockPacket("request_project_create", "incoming-request.json");
        JsonNode expectedReply = JsonTestUtils.loadMockPacket("request_project_create", "outgoing-notify.json");

        PacketEntity packetEntity = new PacketEntity();
        packetEntity.setAmieId(233497907L);
        packetEntity.setType("request_project_create");

        PersonEntity personEntity = createPersonEntityFromRealData();
        ProjectEntity projectEntity = createProjectEntityFromRealData();

        when(personService.findOrCreatePersonFromPacket(any(JsonNode.class))).thenReturn(personEntity);
        when(userAccountService.provisionClusterAccount(personEntity)).thenReturn(createClusterAccount());
        when(projectService.createOrFindProject(anyString(), anyString())).thenReturn(projectEntity);

        handler.handle(incomingPacket, packetEntity);
        verify(personService).findOrCreatePersonFromPacket(any(JsonNode.class));
        verify(userAccountService).provisionClusterAccount(personEntity);
        verify(projectService).createOrFindProject(anyString(), anyString());
        verify(membershipService).createMembership(anyString(), anyString(), eq("PI"));

        @SuppressWarnings("unchecked")
        ArgumentCaptor<Map<String, Object>> replyCaptor = ArgumentCaptor.forClass(Map.class);
        verify(amieClient).replyToPacket(eq(233497907L), replyCaptor.capture());

        Map<String, Object> sentReply = replyCaptor.getValue();
        assertThat(sentReply).containsKey("type");
        assertThat(sentReply.get("type")).isEqualTo("notify_project_create");
        assertThat(sentReply).containsKey("body");

        @SuppressWarnings("unchecked")
        Map<String, Object> sentBody = (Map<String, Object>) sentReply.get("body");
        JsonNode expectedBody = expectedReply.path("body");

        assertThat(sentBody).containsKey("ProjectID");
        assertThat(sentBody).containsKey("GrantNumber");
        assertThat(sentBody).containsKey("PiPersonID");
        assertThat(sentBody).containsKey("PiRemoteSiteLogin");
        assertThat(sentBody).containsKey("PiGlobalID");
        assertThat(sentBody).containsKey("ProjectTitle");
        assertThat(sentBody).containsKey("ResourceList");

        assertThat(sentBody.get("GrantNumber")).isEqualTo(expectedBody.path("GrantNumber").asText());
        assertThat(sentBody.get("PiGlobalID")).isEqualTo(expectedBody.path("PiGlobalID").asText());
        assertThat(sentBody.get("ProjectTitle")).isEqualTo(expectedBody.path("ProjectTitle").asText());
    }

    @Test
    void handle_withNullPacketJson_shouldThrowException() {
        PacketEntity packetEntity = createPacketEntity();

        assertThatThrownBy(() -> handler.handle(null, packetEntity))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessage("packetJson is null");
    }

    @Test
    void handle_withMissingBody_shouldThrowException() {
        JsonNode packetJson = objectMapper.createObjectNode();
        PacketEntity packetEntity = createPacketEntity();

        assertThatThrownBy(() -> handler.handle(packetJson, packetEntity))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessage("packetJson.body is missing");
    }

    @Test
    void handle_withMissingGrantNumber_shouldThrowException() {
        JsonNode packetJson = createPacketJsonWithMissingField("GrantNumber");
        PacketEntity packetEntity = createPacketEntity();

        assertThatThrownBy(() -> handler.handle(packetJson, packetEntity))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("'GrantNumber' must not be empty");
    }

    @Test
    void handle_withMissingPiGlobalId_shouldThrowException() {
        JsonNode packetJson = createPacketJsonWithMissingField("PiGlobalID");
        PacketEntity packetEntity = createPacketEntity();

        assertThatThrownBy(() -> handler.handle(packetJson, packetEntity))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("'PiGlobalID' must not be empty");
    }


    private JsonNode createPacketJsonWithMissingField(String missingField) {
        ObjectNode body = objectMapper.createObjectNode()
                .put("GrantNumber", "TEST123")
                .put("PiGlobalID", "PI123456")
                .put("PiFirstName", "John")
                .put("PiLastName", "Doe")
                .put("PiEmail", "john.doe@example.com")
                .put("PiOrganization", "Test Org")
                .put("PiOrgCode", "TEST")
                .put("NsfStatusCode", "ACTIVE")
                .put("ProjectTitle", "Test Project")
                .put("ResourceList", "[]")
                .put("PiDnList", "[]");

        if (body.has(missingField)) {
            body.remove(missingField);
        }

        return objectMapper.createObjectNode().set("body", body);
    }

    private PacketEntity createPacketEntity() {
        PacketEntity entity = new PacketEntity();
        entity.setAmieId(12345L);
        entity.setType("request_project_create");
        return entity;
    }


    private ClusterAccountEntity createClusterAccount() {
        ClusterAccountEntity entity = new ClusterAccountEntity();
        entity.setId("account-123");
        entity.setUsername("jdoe");
        return entity;
    }

    private PersonEntity createPersonEntityFromRealData() {
        PersonEntity entity = new PersonEntity();
        entity.setId("person-123");
        entity.setAccessGlobalId("54806");
        entity.setFirstName("Hui");
        entity.setLastName("Wan");
        entity.setEmail("hwan@uccs.edu");
        return entity;
    }

    private ProjectEntity createProjectEntityFromRealData() {
        ProjectEntity entity = new ProjectEntity();
        entity.setId("project-123");
        entity.setGrantNumber("NNT259276");
        entity.setActive(true);
        return entity;
    }
}
