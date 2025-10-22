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
package org.apache.custos.signer.sdk.keystore;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.security.KeyPair;
import java.time.Duration;
import java.time.Instant;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

/**
 * In-memory key store implementation (default for v1).
 * Stores keypairs in memory with automatic expiry cleanup.
 * All keys are lost on restart - no persistence.
 */
public class InMemoryKeyStore implements KeyStoreProvider {

    private static final Logger logger = LoggerFactory.getLogger(InMemoryKeyStore.class);

    private final Map<String, KeyEntry> keyStore = new ConcurrentHashMap<>();
    private final ScheduledExecutorService cleanupExecutor = Executors.newSingleThreadScheduledExecutor(r -> {
        Thread t = new Thread(r, "InMemoryKeyStore-cleanup");
        t.setDaemon(true);
        return t;
    });

    public InMemoryKeyStore() {
        // Schedule cleanup every 60 seconds
        cleanupExecutor.scheduleAtFixedRate(this::cleanupExpiredKeys, 60, 60, TimeUnit.SECONDS);
        logger.debug("InMemoryKeyStore initialized with automatic cleanup");
    }

    @Override
    public void store(String userId, String contextId, KeyPair keyPair, Duration ttl) {
        String key = buildKey(userId, contextId);
        Instant expiryTime = Instant.now().plus(ttl);

        KeyEntry entry = new KeyEntry(keyPair, expiryTime);
        keyStore.put(key, entry);

        logger.debug("Stored keypair for user: {}, context: {}, expires: {}",
                userId, contextId, expiryTime);
    }

    @Override
    public Optional<KeyPair> retrieve(String userId, String contextId) {
        String key = buildKey(userId, contextId);
        KeyEntry entry = keyStore.get(key);

        if (entry == null) {
            logger.debug("No keypair found for user: {}, context: {}", userId, contextId);
            return Optional.empty();
        }

        if (entry.isExpired()) {
            logger.debug("Keypair expired for user: {}, context: {}, expired: {}",
                    userId, contextId, entry.getExpiryTime());
            keyStore.remove(key);
            return Optional.empty();
        }

        logger.debug("Retrieved keypair for user: {}, context: {}", userId, contextId);
        return Optional.of(entry.getKeyPair());
    }

    @Override
    public void delete(String userId, String contextId) {
        String key = buildKey(userId, contextId);
        KeyEntry removed = keyStore.remove(key);

        if (removed != null) {
            logger.debug("Deleted keypair for user: {}, context: {}", userId, contextId);
        } else {
            logger.debug("No keypair to delete for user: {}, context: {}", userId, contextId);
        }
    }

    @Override
    public boolean exists(String userId, String contextId) {
        String key = buildKey(userId, contextId);
        KeyEntry entry = keyStore.get(key);
        return entry != null && !entry.isExpired();
    }

    @Override
    public long size() {
        cleanupExpiredKeys(); // Clean up before counting
        return keyStore.size();
    }

    @Override
    public void clear() {
        int cleared = keyStore.size();
        keyStore.clear();
        logger.debug("Cleared {} keypairs from in-memory store", cleared);
    }

    /**
     * Shutdown the cleanup executor
     */
    public void shutdown() {
        cleanupExecutor.shutdown();
        try {
            if (!cleanupExecutor.awaitTermination(5, TimeUnit.SECONDS)) {
                cleanupExecutor.shutdownNow();
            }
        } catch (InterruptedException e) {
            cleanupExecutor.shutdownNow();
            Thread.currentThread().interrupt();
        }
        logger.debug("InMemoryKeyStore shutdown completed");
    }

    private String buildKey(String userId, String contextId) {
        return userId + ":" + contextId;
    }

    private void cleanupExpiredKeys() {
        Instant now = Instant.now();
        int removed = 0;

        keyStore.entrySet().removeIf(entry -> {
            if (entry.getValue().isExpired()) {
                logger.debug("Cleaned up expired keypair: {}", entry.getKey());
                return true;
            }
            return false;
        });

        if (removed > 0) {
            logger.debug("Cleaned up {} expired keypairs", removed);
        }
    }

    /**
     * Container for keypair with expiry time
     */
    private static class KeyEntry {
        private final KeyPair keyPair;
        private final Instant expiryTime;

        public KeyEntry(KeyPair keyPair, Instant expiryTime) {
            this.keyPair = keyPair;
            this.expiryTime = expiryTime;
        }

        public KeyPair getKeyPair() {
            return keyPair;
        }

        public Instant getExpiryTime() {
            return expiryTime;
        }

        public boolean isExpired() {
            return Instant.now().isAfter(expiryTime);
        }
    }
}
