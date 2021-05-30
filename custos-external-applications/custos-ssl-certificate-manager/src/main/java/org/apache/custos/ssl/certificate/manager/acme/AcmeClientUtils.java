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
import java.io.FileReader;
import java.io.FileWriter;
import java.io.IOException;
import java.security.KeyPair;

public class AcmeClientUtils {

    static KeyPair userKeyPair(String userKey, int keySize) throws IOException {
        File f = new File(userKey);
        if (f.exists()) {
            try (FileReader fr = new FileReader(f)) {
                return KeyPairUtils.readKeyPair(fr);
            }

        } else {
            KeyPair userKeyPair = KeyPairUtils.createKeyPair(keySize);
            try (FileWriter fw = new FileWriter(f)) {
                KeyPairUtils.writeKeyPair(userKeyPair, fw);
            }
            return userKeyPair;
        }
    }

    static KeyPair domainKeyPair(String domainKey, int keySize) throws IOException {
        File f = new File(domainKey);
        if (f.exists()) {
            try (FileReader fr = new FileReader(f)) {
                return KeyPairUtils.readKeyPair(fr);
            }
        } else {
            KeyPair domainKeyPair = KeyPairUtils.createKeyPair(keySize);
            try (FileWriter fw = new FileWriter(f)) {
                KeyPairUtils.writeKeyPair(domainKeyPair, fw);
            }
            return domainKeyPair;
        }
    }
}
