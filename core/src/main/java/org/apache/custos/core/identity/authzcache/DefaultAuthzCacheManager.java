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

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Component
public class DefaultAuthzCacheManager implements AuthzCacheManager {

    @Autowired
    private AuthzCache cache;

    @Override
    public AuthzCachedStatus getAuthzCachedStatus(AuthzCacheIndex authzCacheIndex) {
        if (isAuthzDecisionCached(authzCacheIndex)) {
            AuthzCacheEntry cacheEntry = getAuthzCacheEntry(authzCacheIndex);
            long expiryTime = cacheEntry.getExpiryTime();
            long currentTime = System.currentTimeMillis();
            long timePassed = (currentTime - cacheEntry.getEntryTimestamp()) / 1000;

            if (expiryTime > timePassed) {
                // access token is still valid. Hence, return the cached decision
                if (cacheEntry.getDecision()) {
                    return AuthzCachedStatus.AUTHORIZED;
                } else {
                    return AuthzCachedStatus.NOT_AUTHORIZED;
                }
            } else {
                // access token has been expired. Hence, remove the entry and return.
                removeAuthzCacheEntry(authzCacheIndex);
                return AuthzCachedStatus.NOT_CACHED;
            }

        } else {
            return AuthzCachedStatus.NOT_CACHED;
        }
    }

    @Override
    public void addToAuthzCache(AuthzCacheIndex authzCacheIndex, AuthzCacheEntry authzCacheEntry) {
        cache.put(authzCacheIndex, authzCacheEntry);
    }

    @Override
    public boolean isAuthzDecisionCached(AuthzCacheIndex authzCacheIndex) {
        return cache.containsKey(authzCacheIndex);
    }

    @Override
    public AuthzCacheEntry getAuthzCacheEntry(AuthzCacheIndex authzCacheIndex) {
        return cache.get(authzCacheIndex);
    }

    @Override
    public void removeAuthzCacheEntry(AuthzCacheIndex authzCacheIndex) {
        cache.remove(authzCacheIndex);
    }

    @Override
    public void clearCache() {
        cache.clear();
    }
}