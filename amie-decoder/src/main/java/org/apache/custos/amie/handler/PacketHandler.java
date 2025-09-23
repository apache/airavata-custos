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
import org.apache.custos.amie.model.PacketEntity;

public interface PacketHandler {

    /**
     * Executes the business logic for a given AMIE packet
     *
     * @param packetJson   The raw packet content
     * @param packetEntity The entity for the packet
     * @throws Exception if processing fails
     */
    void handle(JsonNode packetJson, PacketEntity packetEntity) throws Exception;

    /**
     * Define which packet type this handler is responsible for.
     * For example, "request_account_create". "*" indicates a default handler.
     *
     * @return The AMIE packet type this handler supports
     */
    String supportsType();

}
