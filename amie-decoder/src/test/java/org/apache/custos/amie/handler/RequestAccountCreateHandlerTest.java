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
import org.apache.custos.amie.client.AmieClient;
import org.apache.custos.amie.model.ClusterAccountEntity;
import org.apache.custos.amie.model.PacketEntity;
import org.apache.custos.amie.model.PersonEntity;
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
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
@Tag("unit")
class RequestAccountCreateHandlerTest {

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

    private RequestAccountCreateHandler handler;
    private ObjectMapper objectMapper;

    @BeforeEach
    void setUp() {
        handler = new RequestAccountCreateHandler(amieClient, personService, userAccountService, projectService, membershipService);
        objectMapper = new ObjectMapper();
    }

    @Test
    void supportsType_shouldReturnCorrectType() {
        assertThat(handler.supportsType()).isEqualTo("request_account_create");
    }

    @Test
    void handle_withValidPacket_shouldProcessSuccessfully() throws Exception {
        JsonNode incomingPacket = JsonTestUtils.loadMockPacket("request_account_create", "incoming-request.json");
        JsonNode expectedReply = JsonTestUtils.loadMockPacket("request_account_create", "outgoing-notify.json");

        PacketEntity packetEntity = new PacketEntity();
        packetEntity.setAmieId(233497917L);
        packetEntity.setType("request_account_create");

        PersonEntity personEntity = createPersonEntityFromRealData();

        when(personService.findOrCreatePersonFromPacket(any(JsonNode.class))).thenReturn(personEntity);
        when(userAccountService.provisionClusterAccount(personEntity)).thenReturn(createClusterAccount());

        handler.handle(incomingPacket, packetEntity);

        verify(personService).findOrCreatePersonFromPacket(any(JsonNode.class));
        verify(userAccountService).provisionClusterAccount(personEntity);
        verify(projectService).createOrFindProject("test-project-456", "TEST123");
        verify(membershipService).createMembership("test-project-456", "account-123", "USER");

        @SuppressWarnings("unchecked")
        ArgumentCaptor<Map<String, Object>> replyCaptor = ArgumentCaptor.forClass(Map.class);
        verify(amieClient).replyToPacket(eq(233497917L), replyCaptor.capture());

        Map<String, Object> sentReply = replyCaptor.getValue();
        assertThat(sentReply).containsKey("type");
        assertThat(sentReply.get("type")).isEqualTo("notify_account_create");
        assertThat(sentReply).containsKey("body");

        @SuppressWarnings("unchecked")
        Map<String, Object> sentBody = (Map<String, Object>) sentReply.get("body");
        JsonNode expectedBody = expectedReply.path("body");

        assertThat(sentBody).containsKey("ProjectID");
        assertThat(sentBody).containsKey("GrantNumber");
        assertThat(sentBody).containsKey("UserPersonID");
        assertThat(sentBody).containsKey("UserRemoteSiteLogin");
        assertThat(sentBody).containsKey("UserOrgCode");
        assertThat(sentBody).containsKey("ResourceList");

        assertThat(sentBody.get("ProjectID")).isEqualTo(expectedBody.path("ProjectID").asText());
        assertThat(sentBody.get("GrantNumber")).isEqualTo(expectedBody.path("GrantNumber").asText());
        assertThat(sentBody.get("UserPersonID")).isEqualTo(expectedBody.path("UserPersonID").asText());
    }

    @Test
    void handle_withMissingProjectId_shouldThrowException() {
        JsonNode packetJson = createPacketJsonWithMissingField("ProjectID");
        PacketEntity packetEntity = createPacketEntity();

        assertThatThrownBy(() -> handler.handle(packetJson, packetEntity))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("'ProjectID' (the local project ID) must not be empty");
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
    void handle_withMissingUserGlobalId_shouldThrowException() {
        JsonNode packetJson = createPacketJsonWithMissingField("UserGlobalID");
        PacketEntity packetEntity = createPacketEntity();

        assertThatThrownBy(() -> handler.handle(packetJson, packetEntity))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("'UserGlobalID' must not be empty");
    }

    private JsonNode createPacketJsonWithMissingField(String missingField) {
        JsonNode body = objectMapper.createObjectNode()
                .put("ProjectID", "PRJ-TEST123")
                .put("GrantNumber", "TEST123")
                .put("UserGlobalID", "USER123456")
                .put("UserFirstName", "John")
                .put("UserLastName", "Doe")
                .put("UserEmail", "john.doe@example.com")
                .put("UserOrganization", "Test University")
                .put("UserOrgCode", "TU")
                .put("NsfStatusCode", "ACTIVE")
                .put("ResourceList", "[]")
                .set("UserDnList", objectMapper.createArrayNode());

        if (body.has(missingField)) {
            ((com.fasterxml.jackson.databind.node.ObjectNode) body).remove(missingField);
        }

        return objectMapper.createObjectNode().set("body", body);
    }

    private PacketEntity createPacketEntity() {
        PacketEntity entity = new PacketEntity();
        entity.setAmieId(12345L);
        entity.setType("request_account_create");
        return entity;
    }

    private PersonEntity createPersonEntityFromRealData() {
        PersonEntity entity = new PersonEntity();
        entity.setId("person-123");
        entity.setAccessGlobalId("12345");
        entity.setFirstName("Test");
        entity.setLastName("User");
        entity.setEmail("testuser@example.com");
        return entity;
    }

    private ClusterAccountEntity createClusterAccount() {
        ClusterAccountEntity entity = new ClusterAccountEntity();
        entity.setId("account-123");
        entity.setUsername("testuser");
        return entity;
    }
}
