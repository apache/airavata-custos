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

package org.apache.custos.cert.updater;

import org.apache.custos.clients.core.ClientUtils;
import org.apache.custos.core.services.commons.security.KeystoreHandler;

import java.io.*;
import java.security.KeyStoreException;
import java.security.NoSuchAlgorithmException;
import java.security.cert.CertificateException;
import java.util.Properties;

public class CertUpdater {


    public static void main(String[] args) throws IOException, CertificateException, NoSuchAlgorithmException, KeyStoreException {

        String path = System.getProperty("user.dir");
        String filePath = path + "/src/main/resources/cert-updater.properties";

        Properties properties = getProperties(filePath);

        String host = properties.getProperty("custos.service.host");
        String clientId = properties.getProperty("custos.client.id");
        String clientSec = properties.getProperty("custos.client.sec");
        String update = properties.getProperty("custos.cert.update");
        String services = properties.getProperty("custos.updating.services");

        if (Boolean.parseBoolean(update)) {
            String[] servicesTobeUpdated = services.split(",");

            for (String service : servicesTobeUpdated) {
                if (!KeystoreHandler.isValidCertificate(path + "../../custos-core-services/" + service +
                                "src/main/resources/keycloak-client-truststore.pkcs12",
                        "keycloak", "ca")) {
                    InputStream inputStream = ClientUtils.getServerCertificate(host, clientId, clientSec);

                    KeystoreHandler.UpdateKeyStore(path + "/src/main/resources/keycloak-client-truststore.pkcs12",
                            "keycloak", "ca", inputStream);
                    boolean valid = KeystoreHandler.isValidCertificate(path + "/src/main/resources/keycloak-client-truststore.pkcs12",
                            "keycloak", "ca");
                }
            }
        }
    }


    private static Properties getProperties(String filePath) throws IOException {
        CertUpdater certUpdater = new CertUpdater();
        Properties properties = new Properties();
        File file = new File(filePath);
        InputStream inputStream = null;
        try {
            inputStream = new FileInputStream(file);
            if (inputStream != null) {
                properties.load(inputStream);
            }
        } catch (IOException e) {
            e.printStackTrace();
        } finally {
            if (inputStream != null) {
                inputStream.close();
            }
        }
        return properties;
    }

    private static boolean writeToFile(String dstFilePath, InputStream stream) throws IOException {
        File targetFile = new File(dstFilePath);
        OutputStream outputStream = new FileOutputStream(targetFile);
        byte[] buffer = new byte[stream.available()];
        stream.read(buffer);
        outputStream.write(buffer);
        outputStream.close();
        return true;

    }

}
