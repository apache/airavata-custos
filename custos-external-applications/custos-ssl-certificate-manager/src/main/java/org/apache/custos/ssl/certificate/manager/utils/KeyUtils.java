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
 * specific language governing permissions and limitations
 * under the License.
 */

package org.apache.custos.ssl.certificate.manager.utils;

import org.shredzone.acme4j.util.KeyPairUtils;

import java.io.IOException;
import java.io.Reader;
import java.io.StringReader;
import java.io.StringWriter;
import java.io.Writer;
import java.security.KeyPair;

public class KeyUtils {

    public static KeyPair getKeyPair(int keySize) {
        KeyPair keyPair = KeyPairUtils.createKeyPair(keySize);
        return keyPair;
    }

    public static KeyPair convertToKeyPair(String key) throws IOException {
        try (Reader reader = new StringReader(key)) {
            return KeyPairUtils.readKeyPair(reader);
        }
    }

    public static String convertToString(KeyPair key) throws IOException {
        try (Writer writer = new StringWriter()) {
            KeyPairUtils.writeKeyPair(key, writer);
            return writer.toString();
        }
    }
}
