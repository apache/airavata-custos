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

package org.apache.custos.core.identity.authzcache;


import org.apache.custos.core.identity.exceptions.AuthSecurityException;

/**
 * The AuthzCacheManager interface represents a cache manager for authorization decisions.
 */
public interface AuthzCacheManager {

    /**
     * Retrieves the cached status of the authorization for the given AuthzCacheIndex
     *
     * @param authzCacheIndex The AuthzCacheIndex representing the authorization request.
     * @return The AuthzCachedStatus enum indicating the status of the authorization cache.
     * @throws AuthSecurityException If there is a security exception while retrieving the cached status.
     */
    AuthzCachedStatus getAuthzCachedStatus(AuthzCacheIndex authzCacheIndex) throws AuthSecurityException;

    /**
     * Adds an entry to the authorization cache.
     *
     * @param authzCacheIndex The AuthzCacheIndex representing the authorization request.
     * @param authzCacheEntry The AuthzCacheEntry containing the authorization decision and metadata.
     * @throws AuthSecurityException If there is a security exception while adding the entry to the cache.
     */
    void addToAuthzCache(AuthzCacheIndex authzCacheIndex, AuthzCacheEntry authzCacheEntry) throws AuthSecurityException;

    /**
     * Checks if the authorization decision is cached for the given AuthzCacheIndex.
     *
     * @param authzCacheIndex The AuthzCacheIndex representing the authorization request.
     * @return true if the authorization decision is cached, false otherwise.
     * @throws AuthSecurityException If there is a security exception while checking the authorization cache.
     */
    boolean isAuthzDecisionCached(AuthzCacheIndex authzCacheIndex) throws AuthSecurityException;

    /**
     * Retrieves the authorization cache entry for the given AuthzCacheIndex.
     *
     * @param authzCacheIndex The AuthzCacheIndex representing the authorization request.
     * @return The AuthzCacheEntry containing the authorization decision and metadata.
     * @throws AuthSecurityException If there is a security exception while retrieving the cache entry.
     */
    AuthzCacheEntry getAuthzCacheEntry(AuthzCacheIndex authzCacheIndex) throws AuthSecurityException;

    /**
     * Removes an entry from the authorization cache.
     *
     * @param authzCacheIndex The AuthzCacheIndex representing the authorization request.
     * @throws AuthSecurityException If there is a security exception while removing the entry from the cache.
     */
    void removeAuthzCacheEntry(AuthzCacheIndex authzCacheIndex) throws AuthSecurityException;

    /**
     * Clears the cache used for authorization decisions.
     *
     * @throws AuthSecurityException if there is a security exception while clearing the cache.
     */
    void clearCache() throws AuthSecurityException;

}
