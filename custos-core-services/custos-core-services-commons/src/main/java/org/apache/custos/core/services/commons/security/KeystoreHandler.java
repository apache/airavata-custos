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

package org.apache.custos.core.services.commons.security;


import org.apache.catalina.security.SecurityUtil;
import org.apache.commons.codec.binary.Base64;
import org.apache.custos.core.services.commons.util.Constants;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.*;
import java.security.KeyStore;
import java.security.KeyStoreException;
import java.security.NoSuchAlgorithmException;
import java.security.cert.Certificate;
import java.security.cert.CertificateException;
import java.security.cert.CertificateFactory;
import java.security.cert.X509Certificate;
import java.util.Collection;
import java.util.Date;

public class KeystoreHandler {

    private static final Logger LOGGER = LoggerFactory.getLogger(KeystoreHandler.class);

    private static KeyStore loadKeyStore(String trustStorePath, String trustStorePassword) throws IOException,
            KeyStoreException, CertificateException, NoSuchAlgorithmException {

        File trustStoreFile = new File(trustStorePath);
        InputStream is;
        if (trustStoreFile.exists()) {
            LOGGER.debug("Loading trust store file from path " + trustStorePath);
            is = new FileInputStream(trustStorePath);
        } else {
            LOGGER.debug("Trying to load trust store file form class path " + trustStorePath);
            is = SecurityUtil.class.getClassLoader().getResourceAsStream(trustStorePath);
            if (is != null) {
                LOGGER.debug("Trust store file was loaded form class path " + trustStorePath);
            }
        }

        if (is == null) {
            throw new RuntimeException("Could not find a trust store file in path " + trustStorePath);
        }

        KeyStore trustStore = KeyStore.getInstance(KeyStore.getDefaultType());

        char[] trustPassword = trustStorePassword.toCharArray();

        trustStore.load(is, trustPassword);

        return trustStore;
    }

    public static boolean UpdateKeyStore(String trustStorePath, String trustStorePassword, String alias, InputStream inputStream) throws CertificateException,
            NoSuchAlgorithmException, KeyStoreException, IOException {
        KeyStore store = loadKeyStore(trustStorePath, trustStorePassword);

        Base64 encoder = new Base64(64);

        String files = new String(inputStream.readAllBytes());

        byte[] decoded = encoder.decode(files.replaceAll("-----BEGIN CERTIFICATE-----", "").replaceAll("-----END CERTIFICATE-----", ""));
        CertificateFactory cf = CertificateFactory.getInstance("X.509");
        Collection<Certificate> certificates = (Collection<Certificate>) cf.generateCertificates(new ByteArrayInputStream(decoded));
        store.setCertificateEntry(alias, certificates.iterator().next());

        FileOutputStream out = new FileOutputStream(trustStorePath);
        store.store(out, trustStorePassword.toCharArray());
        out.close();

        return true;
    }

    public static boolean isValidCertificate(String trustStorePath, String trustStorePassword, String alias) {
        try {
            KeyStore keyStore = loadKeyStore(trustStorePath, trustStorePassword);
            Certificate certificate = keyStore.getCertificate(alias);
            if (certificate.getType().equalsIgnoreCase(Constants.X509CERTIFICATE)) {
                Date date = ((X509Certificate) certificate).getNotAfter();
                Date currentDate = new Date();
                if (currentDate.getTime() > date.getTime()) {
                    return false;
                } else {
                    return true;
                }
            }
        } catch (Exception ex) {
            LOGGER.error("Exception occuured while checking validity of certificate ", ex);
        }
        return false;
    }

}
