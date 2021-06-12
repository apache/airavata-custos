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
     * @param keyPair
     * @return string from key pair
     * @throws IOException
     */
    public static String convertToString(KeyPair keyPair) throws IOException {
        try (Writer writer = new StringWriter()) {
            KeyPairUtils.writeKeyPair(keyPair, writer);
            return writer.toString();
        }
    }

    /**
     * Convert X.509 certificate to string
     *
     * @param certificate certificate
     * @return certificate string
     * @throws CertificateEncodingException
     * @throws IOException
     */
    public static String certificateToString(Certificate certificate) throws CertificateEncodingException, IOException {
        Iterator iterator = certificate.getCertificateChain().iterator();
        try (StringWriter out = new StringWriter()) {
            while (iterator.hasNext()) {
                X509Certificate cert = (X509Certificate) iterator.next();
                AcmeUtils.writeToPem(cert.getEncoded(), AcmeUtils.PemLabel.CERTIFICATE, out);
            }
            String data = out.toString();
            return data;
        }
    }
}
