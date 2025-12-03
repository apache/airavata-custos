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
package org.apache.custos.signer.service.policy;

// TODO: move to shared/common module for reuse with SDK
public enum KeyType {
    ED25519("ed25519"),
    RSA("rsa"),
    ECDSA("ecdsa");

    private final String id;

    KeyType(String id) {
        this.id = id;
    }

    public String id() {
        return id;
    }

    public static KeyType from(String value) {
        if (value == null) {
            return ED25519;
        }
        String normalized = value.toLowerCase();
        return switch (normalized) {
            case "rsa", "ssh-rsa" -> RSA;
            case "ecdsa", "ecdsa-sha2-nistp256" -> ECDSA;
            default -> ED25519;
        };
    }
}
