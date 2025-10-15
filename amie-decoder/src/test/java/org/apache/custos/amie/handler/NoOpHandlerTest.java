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
import org.mockito.junit.jupiter.MockitoExtension;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

@ExtendWith(MockitoExtension.class)
@Tag("unit")
class NoOpHandlerTest {

    private NoOpHandler handler;
    private ObjectMapper objectMapper;

    @BeforeEach
    void setUp() {
        handler = new NoOpHandler();
        objectMapper = new ObjectMapper();
    }

    @Test
    void supportsType_shouldReturnWildcard() {
        assertThat(handler.supportsType()).isEqualTo("*");
    }

    @Test
    void handle_shouldNotThrowException() {
        JsonNode packetJson = objectMapper.createObjectNode();
        PacketEntity packetEntity = new PacketEntity();
        packetEntity.setAmieId(12345L);
        packetEntity.setType("unknown_type");

        handler.handle(packetJson, packetEntity);
    }

    @Test
    void handle_withNullPacketJson_shouldNotThrowException() {
        PacketEntity packetEntity = new PacketEntity();
        packetEntity.setAmieId(12345L);
        packetEntity.setType("unknown_type");

        handler.handle(null, packetEntity);
    }

    @Test
    void handle_withNullPacketEntity_shouldNotThrowException() {
        JsonNode packetJson = objectMapper.createObjectNode();

        //noinspection DataFlowIssue
        assertThatThrownBy(() -> handler.handle(packetJson, null)).isInstanceOf(NullPointerException.class);
    }
}
