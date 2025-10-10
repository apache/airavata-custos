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
package org.apache.custos.amie.model;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.EnumType;
import jakarta.persistence.Enumerated;
import jakarta.persistence.Id;
import jakarta.persistence.Lob;
import jakarta.persistence.Table;

import java.time.Instant;
import java.util.UUID;

@Entity
@Table(name = "packets")
public class PacketEntity {

    @Id
    private String id;

    @Column(name = "amie_id", nullable = false, unique = true)
    private Long amieId;

    @Column(name = "type", nullable = false, length = 64)
    private String type;

    @Enumerated(EnumType.STRING)
    @Column(name = "status", columnDefinition = "VARCHAR", nullable = false, length = 32)
    private PacketStatus status = PacketStatus.NEW;

    @Lob
    @Column(name = "raw_json", columnDefinition = "TEXT", nullable = false)
    private String rawJson;

    @Column(name = "received_at", nullable = false)
    private Instant receivedAt = Instant.now();

    @Column(name = "decoded_at")
    private Instant decodedAt;

    @Column(name = "processed_at")
    private Instant processedAt;

    @Column(name = "retries", nullable = false)
    private int retries = 0;

    @Lob
    @Column(name = "last_error", columnDefinition = "TEXT")
    private String lastError;

    public PacketEntity() {
        this.id = UUID.randomUUID().toString();
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public Long getAmieId() {
        return amieId;
    }

    public void setAmieId(Long amieId) {
        this.amieId = amieId;
    }

    public String getType() {
        return type;
    }

    public void setType(String type) {
        this.type = type;
    }

    public PacketStatus getStatus() {
        return status;
    }

    public void setStatus(PacketStatus status) {
        this.status = status;
    }

    public String getRawJson() {
        return rawJson;
    }

    public void setRawJson(String rawJson) {
        this.rawJson = rawJson;
    }

    public Instant getReceivedAt() {
        return receivedAt;
    }

    public void setReceivedAt(Instant receivedAt) {
        this.receivedAt = receivedAt;
    }

    public Instant getDecodedAt() {
        return decodedAt;
    }

    public void setDecodedAt(Instant decodedAt) {
        this.decodedAt = decodedAt;
    }

    public Instant getProcessedAt() {
        return processedAt;
    }

    public void setProcessedAt(Instant processedAt) {
        this.processedAt = processedAt;
    }

    public int getRetries() {
        return retries;
    }

    public void setRetries(int retries) {
        this.retries = retries;
    }

    public String getLastError() {
        return lastError;
    }

    public void setLastError(String lastError) {
        this.lastError = lastError;
    }
}
