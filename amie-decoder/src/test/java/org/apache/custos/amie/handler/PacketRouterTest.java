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
import org.apache.custos.amie.model.PacketEntity;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.List;

import static org.mockito.Mockito.any;
import static org.mockito.Mockito.lenient;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;

@ExtendWith(MockitoExtension.class)
@Tag("unit")
class PacketRouterTest {

    @Mock
    private PacketHandler mockHandler1;

    @Mock
    private PacketHandler mockHandler2;

    private PacketRouter router;
    private ObjectMapper objectMapper;

    @BeforeEach
    void setUp() {
        lenient().when(mockHandler1.supportsType()).thenReturn("request_project_create");
        lenient().when(mockHandler2.supportsType()).thenReturn("request_account_create");

        router = new PacketRouter(List.of(mockHandler1, mockHandler2));
        objectMapper = new ObjectMapper();
    }

    @Test
    void route_withMatchingHandler_shouldCallCorrectHandler() throws Exception {
        JsonNode packetJson = objectMapper.createObjectNode();
        PacketEntity packetEntity = new PacketEntity();
        packetEntity.setType("request_project_create");

        router.route(packetJson, packetEntity);

        verify(mockHandler1).handle(packetJson, packetEntity);
        verify(mockHandler2, never()).handle(any(), any());
    }

    @Test
    void route_withDifferentMatchingHandler_shouldCallCorrectHandler() throws Exception {
        JsonNode packetJson = objectMapper.createObjectNode();
        PacketEntity packetEntity = new PacketEntity();
        packetEntity.setType("request_account_create");

        router.route(packetJson, packetEntity);

        verify(mockHandler2).handle(packetJson, packetEntity);
        verify(mockHandler1, never()).handle(any(), any());
    }

    @Test
    void route_withNoMatchingHandler_shouldCallDefaultHandler() throws Exception {
        JsonNode packetJson = objectMapper.createObjectNode();
        PacketEntity packetEntity = new PacketEntity();
        packetEntity.setType("unknown_packet_type");

        router.route(packetJson, packetEntity);

        verify(mockHandler1, never()).handle(any(), any());
        verify(mockHandler2, never()).handle(any(), any());
    }

    @Test
    void route_withCaseInsensitiveMatching_shouldCallCorrectHandler() throws Exception {

        JsonNode packetJson = objectMapper.createObjectNode();
        PacketEntity packetEntity = new PacketEntity();
        packetEntity.setType("REQUEST_PROJECT_CREATE");

        router.route(packetJson, packetEntity);

        verify(mockHandler1).handle(packetJson, packetEntity);
        verify(mockHandler2, never()).handle(any(), any());
    }
}
