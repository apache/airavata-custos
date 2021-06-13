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

import org.shredzone.acme4j.Certificate;
import org.shredzone.acme4j.toolbox.AcmeUtils;
import org.shredzone.acme4j.util.KeyPairUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.io.Reader;
import java.io.StringReader;
import java.io.StringWriter;
import java.io.Writer;
import java.security.KeyPair;
import java.security.cert.CertificateEncodingException;
import java.security.cert.X509Certificate;
import java.util.Iterator;

/**
 * Utility class for KeyPair
 */
public class CertUtils {
    private static final Logger logger = LoggerFactory.getLogger(CertUtils.class);

    /**
     * Get new key pair
     *
     * @param keySize key size
     * @return new keypair
     */
    public static KeyPair getKeyPair(int keySize) {
        KeyPair keyPair = KeyPairUtils.createKeyPair(keySize);
        return keyPair;
    }


    /**
     * Convert string to a keypair
     *
     * @param keyPair string key pair
     * @return KeyPair from string key pair
     * @throws IOException
     */
    public static KeyPair convertToKeyPair(String keyPair) throws IOException {
        try (Reader reader = new StringReader(keyPair)) {
            return KeyPairUtils.readKeyPair(reader);
        }
    }

    /**
     * Convert keypair to string
     *
     * @param credential
     * @return string from key pair
     * @throws IOException
     * @throws CertificateEncodingException
     */
    public static <T> String toString(T credential) throws IOException, CertificateEncodingException {
        if (credential instanceof KeyPair) {
            try (Writer writer = new StringWriter()) {
                KeyPairUtils.writeKeyPair((KeyPair) credential, writer);
                return writer.toString();
            }
        } else if (credential instanceof Certificate) {
            Iterator iterator = ((Certificate) credential).getCertificateChain().iterator();
            try (StringWriter out = new StringWriter()) {
                while (iterator.hasNext()) {
                    X509Certificate cert = (X509Certificate) iterator.next();
                    AcmeUtils.writeToPem(cert.getEncoded(), AcmeUtils.PemLabel.CERTIFICATE, out);
                }
                String data = out.toString();
                return data;
            }
        } else {
            logger.warn("Invalid credential format");
            return null;
        }
    }
}
