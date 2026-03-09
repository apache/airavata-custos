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
package org.apache.custos.access.ci.service.model.amie;

public enum PacketStatus {

    /**
     * Packet has been received from the AMIE API and persisted but has not yet been processed.
     */
    NEW,
    /**
     * The packet's raw JSON content has been successfully parsed, and the initial processing event has been created.
     */
    DECODED,
    /**
     * All required processing events have been completed successfully.
     */
    PROCESSED,
    /**
     * The processing of this packet failed and will not be automatically retried.
     */
    FAILED
}
