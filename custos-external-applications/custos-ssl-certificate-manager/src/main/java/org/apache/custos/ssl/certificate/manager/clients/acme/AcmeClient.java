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

package org.apache.custos.ssl.certificate.manager.clients.acme;

import org.apache.custos.ssl.certificate.manager.clients.CustosClient;
import org.apache.custos.ssl.certificate.manager.clients.NginxClient;
import org.apache.custos.ssl.certificate.manager.configurations.AcmeConfiguration;
import org.shredzone.acme4j.Account;
import org.shredzone.acme4j.AccountBuilder;
import org.shredzone.acme4j.Authorization;
import org.shredzone.acme4j.Certificate;
import org.shredzone.acme4j.Order;
import org.shredzone.acme4j.Session;
import org.shredzone.acme4j.Status;
import org.shredzone.acme4j.challenge.Http01Challenge;
import org.shredzone.acme4j.exception.AcmeException;
import org.shredzone.acme4j.util.CSRBuilder;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.net.URI;
import java.security.KeyPair;

public class AcmeClient {

    private static final Logger logger = LoggerFactory.getLogger(AcmeClient.class);
    private AcmeConfiguration config;
    private CustosClient custosClient;
    private NginxClient nginxClient;

    public AcmeClient(AcmeConfiguration config, CustosClient custosClient, NginxClient nginxClient) {
        this.config = config;
        this.custosClient = custosClient;
        this.nginxClient = nginxClient;
    }

    public Order getCertificateOrder() throws AcmeException, IOException {
        KeyPair userKey = this.getKeyPair("user_key", config.getUserKey());
        Session session = new Session(URI.create(config.getUrl()));
        Account account = new AccountBuilder()
                .agreeToTermsOfService()
                .useKeyPair(userKey)
                .createLogin(session)
                .getAccount();

        logger.info("Registered a new user, URL: {}", account.getLocation());

        Order order = account.newOrder().domains(config.getDomains()).create();
        return order;
    }

    public Certificate getCertificateCredentials(Order order) throws IOException, AcmeException {
        KeyPair domainKeyPair = this.getKeyPair("domain_key", config.getUserKey());
        CSRBuilder csrBuilder = new CSRBuilder();
        csrBuilder.addDomains(this.config.getDomains());
        csrBuilder.sign(domainKeyPair);

        order.execute(csrBuilder.getEncoded());

        try {
            AcmeClientTasks.completeOrder(order);
        } catch (InterruptedException e) {
            logger.error("Couldn't complete order. Interrupted.");
            throw new AcmeException("Order failed.");
        }

        Certificate certificate = order.getCertificate();
        logger.info("Success! The certificate for domains {} has been generated!", config.getDomains());
        logger.info("Certificate URL: {}", certificate.getLocation());

        return certificate;
    }

    public void authorizeDomain(Order order) throws AcmeException {
        for (Authorization auth : order.getAuthorizations()) {
            logger.info("Authorization for domain {}", auth.getIdentifier().getDomain());

            if (auth.getStatus() == Status.VALID) {
                return;
            }

            Http01Challenge challenge = httpChallenge(auth);
            if (challenge.getStatus() == Status.VALID) {
                return;
            }

            boolean success = this.nginxClient.createChallenge(challenge.getToken(), challenge.getAuthorization());
            if (!success) {
                logger.error("Couldn't create challenge in nginx server");
                throw new AcmeException("Challenge failed.");
            }

            challenge.trigger();
            try {
                AcmeClientTasks.validateChallenge(challenge);
            } catch (InterruptedException e) {
                logger.error("Couldn't validate challenge. Interrupted.");
                throw new AcmeException("Challenge failed.");
            }

            if (challenge.getStatus() != Status.VALID) {
                throw new AcmeException("Failed to pass the challenge for domain " + auth.getIdentifier().getDomain());
            }

            nginxClient.deleteResource(challenge.getToken());
            logger.info("Challenge has been completed.");
        }
    }

    private KeyPair getKeyPair(String key, String value) throws IOException {
        String keyPairValue = value;
        if (keyPairValue == null || keyPairValue.isEmpty()) {
            try {
                keyPairValue = this.custosClient.getKVCredentials(key);
                return AcmeClientUtils.convertToKeyPair(keyPairValue);
            } catch (Exception e) {
                logger.error("Key {} not in custos.", key);
            }
        }

        KeyPair keyPair = AcmeClientUtils.getKeyPair(config.getKeySize());
        this.custosClient.addKVCredential(key, AcmeClientUtils.convertToString(keyPair));
        return keyPair;
    }

    private Http01Challenge httpChallenge(Authorization auth) throws AcmeException {
        Http01Challenge challenge = auth.findChallenge(Http01Challenge.class);
        if (challenge == null) {
            throw new AcmeException("Found no " + Http01Challenge.TYPE + " challenge");
        }

        logger.info("Domain : {}", auth.getIdentifier().getDomain());
        logger.info("File : {}", challenge.getToken());
        logger.info("Content: {}", challenge.getAuthorization());
        return challenge;
    }
}
