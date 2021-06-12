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

package org.apache.custos.ssl.certificate.manager;

import org.apache.custos.ssl.certificate.manager.clients.AcmeClient;
import org.apache.custos.ssl.certificate.manager.clients.CustosClient;
import org.apache.custos.ssl.certificate.manager.clients.NginxClient;
import org.apache.custos.ssl.certificate.manager.configurations.AcmeConfiguration;
import org.apache.custos.ssl.certificate.manager.configurations.CustosConfiguration;
import org.apache.custos.ssl.certificate.manager.configurations.NginxConfiguration;
import org.apache.custos.ssl.certificate.manager.utils.CertUtils;
import org.bouncycastle.jce.provider.BouncyCastleProvider;
import org.quartz.Job;
import org.quartz.JobExecutionContext;
import org.shredzone.acme4j.Certificate;
import org.shredzone.acme4j.Order;
import org.shredzone.acme4j.Status;
import org.shredzone.acme4j.challenge.Http01Challenge;
import org.shredzone.acme4j.exception.AcmeException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.security.KeyPair;
import java.security.Security;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Map;

public class CertUpdater implements Job {

    private static final Logger logger = LoggerFactory.getLogger(CertUpdater.class);
    private final String CERT_UPDATER_USER_KEY = "cert_updater_user_key";
    private final String CERT_UPDATER_DOMAIN_KEY = "cert_updater_domain_key";

    @Override
    public void execute(JobExecutionContext jobExecutionContext) {
        Map<String, String> env = new HashMap<>();
        for (Map.Entry<String, Object> entry : jobExecutionContext.getJobDetail().getJobDataMap().entrySet()) {
            env.put(entry.getKey(), (String) entry.getValue());
        }
        NginxConfiguration nginxConfiguration = new NginxConfiguration(env);
        AcmeConfiguration acmeConfiguration = new AcmeConfiguration(env);
        CustosConfiguration custosConfiguration = new CustosConfiguration(env);
        try {
            CustosClient custosClient = new CustosClient(custosConfiguration);
            AcmeClient acmeClient = new AcmeClient(acmeConfiguration);
            NginxClient nginxClient = new NginxClient(nginxConfiguration);

            KeyPair userKeyPair = this.getKeyPair(CERT_UPDATER_USER_KEY, acmeConfiguration.getUserKey(), custosClient);
            Order order = acmeClient.getCertificateOrder(userKeyPair);
            ArrayList<Http01Challenge> challenges = acmeClient.getChallenges(order);

            for (Http01Challenge challenge : challenges) {
                String fileName = challenge.getToken();
                String fileContent = challenge.getAuthorization();
                boolean createChallengeSuccess = nginxClient.createAcmeChallenge(fileName, fileContent);
                if (!createChallengeSuccess) {
                    logger.error("Couldn't create challenge in nginx server");
                    throw new AcmeException("Challenge failed.");
                }

                challenge.trigger();
                boolean success = acmeClient.validateChallenge(challenge);
                if (!success) {
                    logger.error("Validating challenge timeout before completing.");
                    throw new AcmeException("Failed to pass the challenge for domain.");
                } else {
                    logger.error("Validating challenge completes.");
                }

                if (challenge.getStatus() != Status.VALID) {
                    throw new AcmeException("Failed to pass the challenge for domain.");
                }

                boolean deleteChallengeSuccess = nginxClient.deleteAcmeChallenge(fileName);
                if (!deleteChallengeSuccess) {
                    logger.error("Couldn't delete challenge file from nginx server.");
                }
            }

            KeyPair domainKeyPair = this.getKeyPair(CERT_UPDATER_DOMAIN_KEY, acmeConfiguration.getDomainKey(),
                    custosClient);
            Certificate certificate = acmeClient.getCertificateCredentials(order, domainKeyPair);
            String token = custosClient.addCertificate(CertUtils.convertToString(domainKeyPair),
                    CertUtils.certificateToString(certificate));
            if (token == null || token.isEmpty()) {
                logger.error("Error has occurred while adding certificate to Custos ");
            } else {
                logger.info("Certificate successfully saved in custos: {}", token);
            }
            custosClient.close();
        } catch (AcmeException e) {
            logger.error("Acme Exception : {} ", e.getMessage());
        } catch (IOException e) {
            logger.error("IO Exception: {}", e.getMessage());
        } catch (InterruptedException e) {
            logger.error("Couldn't validate challenge. Interrupted.");
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    private KeyPair getKeyPair(String key, String value, CustosClient custosClient) throws IOException {
        String keyPairValue = value;
        if (keyPairValue == null || keyPairValue.isEmpty()) {
            try {
                keyPairValue = custosClient.getKVCredentials(key);
                logger.info("{} is available in Custos.", key);
                return CertUtils.convertToKeyPair(keyPairValue);
            } catch (Exception e) {
                logger.error("Key {} isn't available in Custos.", key);
            }
        }

        logger.info("Creating new {} key pair", key);
        KeyPair keyPair = CertUtils.getKeyPair(2048);
        custosClient.addKVCredential(key, CertUtils.convertToString(keyPair));
        return keyPair;
    }

    public static void setupSecurityProvider() {
        Security.addProvider(new BouncyCastleProvider());
    }
}
