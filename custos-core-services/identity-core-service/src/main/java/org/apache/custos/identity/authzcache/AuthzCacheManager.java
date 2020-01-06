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
package org.apache.custos.identity.authzcache;


import org.apache.custos.identity.exceptions.CustosSecurityException;

/**
 * This is the interface through which security manager accesses the underlying caching implementation
 * See the DefaultAuthzCacheManager.java for an example implementation of this interface.
 */
public interface AuthzCacheManager {
    /**
     * Returns the status of the cache w.r.t the given authorization request which is encapsulated in
     * the AuthzCacheIndex.
     *
     * @param authzCacheIndex
     * @return
     */
    public AuthzCachedStatus getAuthzCachedStatus(AuthzCacheIndex authzCacheIndex) throws CustosSecurityException;

    /**
     * Add to cache the authorization decision pertaining to a given authorization request.
     *
     * @param authzCacheIndex
     * @param authzCacheEntry
     * @throws CustosSecurityException
     */
    public void addToAuthzCache(AuthzCacheIndex authzCacheIndex, AuthzCacheEntry authzCacheEntry) throws CustosSecurityException;

    /**
     * Check if a valid decision is cached for a given authorization request.
     *
     * @param authzCacheIndex
     * @return
     */
    public boolean isAuthzDecisionCached(AuthzCacheIndex authzCacheIndex) throws CustosSecurityException;

    /**
     * Returns the AuthzCacheEntry for a given authorization request.
     *
     * @param authzCacheIndex
     * @return
     * @throws CustosSecurityException
     */
    public AuthzCacheEntry getAuthzCacheEntry(AuthzCacheIndex authzCacheIndex) throws CustosSecurityException;

    /**
     * Removes the authorization cache entry for a given authorization request.
     *
     * @param authzCacheIndex
     * @throws CustosSecurityException
     */
    public void removeAuthzCacheEntry(AuthzCacheIndex authzCacheIndex) throws CustosSecurityException;

    /**
     * Clear the authorization cache.
     *
     * @return
     */
    public void clearCache() throws CustosSecurityException;

}
