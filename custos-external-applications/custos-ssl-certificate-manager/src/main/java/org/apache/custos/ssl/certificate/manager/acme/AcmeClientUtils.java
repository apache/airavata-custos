/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing,
 *  software distributed under the License is distributed on an
 *  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *  KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

package org.apache.custos.ssl.certificate.manager.acme;

import org.shredzone.acme4j.util.KeyPairUtils;

import java.io.File;
import java.io.IOException;
import java.io.Reader;
import java.io.StringReader;
import java.security.KeyPair;

public class AcmeClientUtils {

    static KeyPair userKeyPair(String userKey, int keySize) throws IOException {
        if (userKey != null && !userKey.isEmpty()) {
            try (Reader reader = new StringReader(userKey)) {
                return KeyPairUtils.readKeyPair(reader);
            }
        } else {
            File f = new File(userKey);
            KeyPair userKeyPair = KeyPairUtils.createKeyPair(keySize);
            return userKeyPair;
        }
    }

    static KeyPair domainKeyPair(String domainKey, int keySize) throws IOException {
        if (domainKey != null && !domainKey.isEmpty()) {
            try (Reader reader = new StringReader(domainKey)) {
                return KeyPairUtils.readKeyPair(reader);
            }
        } else {
            File f = new File(domainKey);
            KeyPair domainKeyPair = KeyPairUtils.createKeyPair(keySize);
            return domainKeyPair;
        }
    }
}
