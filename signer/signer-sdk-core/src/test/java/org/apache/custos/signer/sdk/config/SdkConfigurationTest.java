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
 */
package org.apache.custos.signer.sdk.config;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;

class SdkConfigurationTest {

    @Test
    void defaultKeyTypeIsEd25519() {
        SdkConfiguration config = new SdkConfiguration.Builder()
                .tenantId("t1")
                .signerServiceAddress("localhost:9095")
                .addClient("a1", "c1", "s1")
                .build();

        SdkConfiguration.ClientConfig client = config.getClientConfig("a1").orElseThrow();
        assertEquals("ed25519", client.getKeyType());
    }

    @Test
    void keyTypeCanBeOverriddenPerClient() {
        SdkConfiguration config = new SdkConfiguration.Builder()
                .tenantId("t1")
                .signerServiceAddress("localhost:9095")
                .addClient("a1", "c1", "s1", "rsa")
                .addClient("a2", "c2", "s2", "ecdsa")
                .build();

        assertEquals("rsa", config.getClientConfig("a1").orElseThrow().getKeyType());
        assertEquals("ecdsa", config.getClientConfig("a2").orElseThrow().getKeyType());
    }
}
