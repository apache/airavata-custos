/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the specific language
 * governing permissions and limitations under the License.
 *
 */
package org.apache.custos.signer.service.audit;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.custos.signer.service.model.CertificateIssuanceLog;
import org.apache.custos.signer.service.model.RevocationEvent;
import org.apache.custos.signer.service.repo.CertificateIssuanceLogRepository;
import org.apache.custos.signer.service.repo.RevocationEventRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.slf4j.MDC;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.HashMap;
import java.util.Map;

/**
 * Service for comprehensive audit logging of certificate operations.
 * Logs to both database and structured JSON logs for compliance and debugging.
 */
@Service
public class AuditLogger {

    private static final Logger logger = LoggerFactory.getLogger(AuditLogger.class);
    private static final Logger auditLogger = LoggerFactory.getLogger("audit");

    @Autowired
    private CertificateIssuanceLogRepository issuanceLogRepository;

    @Autowired
    private RevocationEventRepository revocationEventRepository;

    @Autowired
    private ObjectMapper objectMapper;

    /**
     * Log certificate issuance event
     */
    public void logCertificateIssuance(String tenantId, String clientId, long serialNumber,
                                       String keyId, String principal, String publicKeyFingerprint,
                                       String caFingerprint, LocalDateTime validAfter, LocalDateTime validBefore,
                                       String sourceIp, String userAccessTokenHash, Map<String, Object> metadata) {
        try {
            // Create database log entry
            CertificateIssuanceLog logEntry = new CertificateIssuanceLog();
            logEntry.setTenantId(tenantId);
            logEntry.setClientId(clientId);
            logEntry.setSerialNumber(serialNumber);
            logEntry.setKeyId(keyId);
            logEntry.setPrincipal(principal);
            logEntry.setPublicKeyFingerprint(publicKeyFingerprint);
            logEntry.setCaFingerprint(caFingerprint);
            logEntry.setValidAfter(validAfter);
            logEntry.setValidBefore(validBefore);
            logEntry.setSourceIp(sourceIp);
            logEntry.setUserAccessTokenHash(userAccessTokenHash);

            if (metadata != null && !metadata.isEmpty()) {
                logEntry.setRequestMetadata(objectMapper.writeValueAsString(metadata));
            }

            issuanceLogRepository.save(logEntry);

            // Create structured JSON log
            Map<String, Object> auditEvent = new HashMap<>();
            auditEvent.put("event_type", "certificate_issuance");
            auditEvent.put("timestamp", LocalDateTime.now().atOffset(ZoneOffset.UTC).toString());
            auditEvent.put("tenant_id", tenantId);
            auditEvent.put("client_id", clientId);
            auditEvent.put("serial_number", serialNumber);
            auditEvent.put("key_id", keyId);
            auditEvent.put("principal", principal);
            auditEvent.put("public_key_fingerprint", publicKeyFingerprint);
            auditEvent.put("ca_fingerprint", caFingerprint);
            auditEvent.put("valid_after", validAfter.atOffset(ZoneOffset.UTC).toString());
            auditEvent.put("valid_before", validBefore.atOffset(ZoneOffset.UTC).toString());
            auditEvent.put("source_ip", sourceIp);
            auditEvent.put("user_token_hash", userAccessTokenHash);

            if (metadata != null) {
                auditEvent.put("metadata", metadata);
            }

            // Add correlation ID if available
            String correlationId = MDC.get("traceId");
            if (correlationId != null) {
                auditEvent.put("correlation_id", correlationId);
            }

            auditLogger.info("Certificate issued: {}", objectMapper.writeValueAsString(auditEvent));

        } catch (Exception e) {
            logger.error("Failed to log certificate issuance event", e);
        }
    }

    /**
     * Log certificate revocation event
     */
    public void logCertificateRevocation(String tenantId, String clientId, Long serialNumber,
                                         String keyId, String caFingerprint, String reason, String revokedBy) {
        try {
            // Create database log entry
            RevocationEvent revocationEvent;

            if (serialNumber != null) {
                revocationEvent = RevocationEvent.forSerialNumber(tenantId, clientId, serialNumber, reason, revokedBy);
            } else if (keyId != null) {
                revocationEvent = RevocationEvent.forKeyId(tenantId, clientId, keyId, reason, revokedBy);
            } else if (caFingerprint != null) {
                revocationEvent = RevocationEvent.forCaFingerprint(tenantId, clientId, caFingerprint, reason, revokedBy);
            } else {
                throw new IllegalArgumentException("At least one revocation identifier must be provided");
            }

            revocationEventRepository.save(revocationEvent);

            // Create structured JSON log
            Map<String, Object> auditEvent = new HashMap<>();
            auditEvent.put("event_type", "certificate_revocation");
            auditEvent.put("timestamp", LocalDateTime.now().atOffset(ZoneOffset.UTC).toString());
            auditEvent.put("tenant_id", tenantId);
            auditEvent.put("client_id", clientId);
            auditEvent.put("reason", reason);
            auditEvent.put("revoked_by", revokedBy);

            if (serialNumber != null) {
                auditEvent.put("serial_number", serialNumber);
            }
            if (keyId != null) {
                auditEvent.put("key_id", keyId);
            }
            if (caFingerprint != null) {
                auditEvent.put("ca_fingerprint", caFingerprint);
            }

            // Add correlation ID if available
            String correlationId = MDC.get("traceId");
            if (correlationId != null) {
                auditEvent.put("correlation_id", correlationId);
            }

            auditLogger.info("Certificate revoked: {}", objectMapper.writeValueAsString(auditEvent));

        } catch (Exception e) {
            logger.error("Failed to log certificate revocation event", e);
        }
    }

    /**
     * Log authentication failure
     */
    public void logAuthenticationFailure(String clientId, String reason, String sourceIp) {
        try {
            Map<String, Object> auditEvent = new HashMap<>();
            auditEvent.put("event_type", "authentication_failure");
            auditEvent.put("timestamp", LocalDateTime.now().atOffset(ZoneOffset.UTC).toString());
            auditEvent.put("client_id", clientId);
            auditEvent.put("reason", reason);
            auditEvent.put("source_ip", sourceIp);

            // Add correlation ID if available
            String correlationId = MDC.get("traceId");
            if (correlationId != null) {
                auditEvent.put("correlation_id", correlationId);
            }

            auditLogger.warn("Authentication failed: {}", objectMapper.writeValueAsString(auditEvent));

        } catch (Exception e) {
            logger.error("Failed to log authentication failure", e);
        }
    }

    /**
     * Log policy violation
     */
    public void logPolicyViolation(String tenantId, String clientId, String principal,
                                   String violationType, String reason, String sourceIp) {
        try {
            Map<String, Object> auditEvent = new HashMap<>();
            auditEvent.put("event_type", "policy_violation");
            auditEvent.put("timestamp", LocalDateTime.now().atOffset(ZoneOffset.UTC).toString());
            auditEvent.put("tenant_id", tenantId);
            auditEvent.put("client_id", clientId);
            auditEvent.put("principal", principal);
            auditEvent.put("violation_type", violationType);
            auditEvent.put("reason", reason);
            auditEvent.put("source_ip", sourceIp);

            // Add correlation ID if available
            String correlationId = MDC.get("traceId");
            if (correlationId != null) {
                auditEvent.put("correlation_id", correlationId);
            }

            auditLogger.warn("Policy violation: {}", objectMapper.writeValueAsString(auditEvent));

        } catch (Exception e) {
            logger.error("Failed to log policy violation", e);
        }
    }

    /**
     * Log CA rotation event
     */
    public void logCaRotation(String tenantId, String clientId, String oldCaFingerprint,
                              String newCaFingerprint, String rotatedBy) {
        try {
            Map<String, Object> auditEvent = new HashMap<>();
            auditEvent.put("event_type", "ca_rotation");
            auditEvent.put("timestamp", LocalDateTime.now().atOffset(ZoneOffset.UTC).toString());
            auditEvent.put("tenant_id", tenantId);
            auditEvent.put("client_id", clientId);
            auditEvent.put("old_ca_fingerprint", oldCaFingerprint);
            auditEvent.put("new_ca_fingerprint", newCaFingerprint);
            auditEvent.put("rotated_by", rotatedBy);

            // Add correlation ID if available
            String correlationId = MDC.get("traceId");
            if (correlationId != null) {
                auditEvent.put("correlation_id", correlationId);
            }

            auditLogger.info("CA rotated: {}", objectMapper.writeValueAsString(auditEvent));

        } catch (Exception e) {
            logger.error("Failed to log CA rotation event", e);
        }
    }

    /**
     * Log KRL generation event
     */
    public void logKrlGeneration(String tenantId, String clientId, int revokedCount,
                                 String krlPath, long generationTimeMs) {
        try {
            Map<String, Object> auditEvent = new HashMap<>();
            auditEvent.put("event_type", "krl_generation");
            auditEvent.put("timestamp", LocalDateTime.now().atOffset(ZoneOffset.UTC).toString());
            auditEvent.put("tenant_id", tenantId);
            auditEvent.put("client_id", clientId);
            auditEvent.put("revoked_count", revokedCount);
            auditEvent.put("krl_path", krlPath);
            auditEvent.put("generation_time_ms", generationTimeMs);

            // Add correlation ID if available
            String correlationId = MDC.get("traceId");
            if (correlationId != null) {
                auditEvent.put("correlation_id", correlationId);
            }

            auditLogger.info("KRL generated: {}", objectMapper.writeValueAsString(auditEvent));

        } catch (Exception e) {
            logger.error("Failed to log KRL generation event", e);
        }
    }
}
