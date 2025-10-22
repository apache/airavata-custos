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

import java.security.KeyPair;
import java.time.Duration;
import java.util.Optional;

/**
 * Interface for pluggable SSH key storage backends.
 * Allows different implementations for storing ephemeral or persistent SSH keypairs.
 */
public interface KeyStoreProvider {

    /**
     * Store a keypair with TTL
     *
     * @param userId    User identifier
     * @param contextId Context identifier (e.g., target system, job ID)
     * @param keyPair   The SSH keypair to store
     * @param ttl       Time-to-live for the keypair
     */
    void store(String userId, String contextId, KeyPair keyPair, Duration ttl);

    /**
     * Retrieve a keypair
     *
     * @param userId    User identifier
     * @param contextId Context identifier
     * @return Optional containing the keypair if found and not expired
     */
    Optional<KeyPair> retrieve(String userId, String contextId);

    /**
     * Delete a keypair
     *
     * @param userId    User identifier
     * @param contextId Context identifier
     */
    void delete(String userId, String contextId);

    /**
     * Check if a keypair exists and is not expired
     *
     * @param userId    User identifier
     * @param contextId Context identifier
     * @return true if keypair exists and is valid
     */
    default boolean exists(String userId, String contextId) {
        return retrieve(userId, contextId).isPresent();
    }

    /**
     * Get the number of stored keypairs
     *
     * @return number of keypairs in storage
     */
    default long size() {
        return 0; // Default implementation for backends that don't support counting
    }

    /**
     * Clear all stored keypairs
     */
    default void clear() {
        // Default implementation - subclasses should override if they support bulk operations
    }
}
