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

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.LinkedHashMap;
import java.util.Map;

@Component
public class AuthzCache extends LinkedHashMap<AuthzCacheIndex, AuthzCacheEntry> {

    private final static Logger LOGGER = LoggerFactory.getLogger(AuthzCache.class);

    private int MAX_SIZE;

    public AuthzCache(@Value("${custos.identity.auth.cache.size:1024}") int initialCapacity) {
        super(initialCapacity);
        MAX_SIZE = initialCapacity;
    }

    @Override
    protected boolean removeEldestEntry(Map.Entry<AuthzCacheIndex, AuthzCacheEntry> eldest) {
        if (size() > MAX_SIZE) {
            LOGGER.info("Authz cache max size exceeded. Removing the old entries.");
        }
        return size() > MAX_SIZE;
    }
}
