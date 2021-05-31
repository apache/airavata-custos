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

import org.apache.custos.ssl.certificate.manager.nginx.NginxClient;
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

import java.io.FileWriter;
import java.io.IOException;
import java.net.URI;
import java.security.KeyPair;

public class AcmeClient {

    private static final Logger logger = LoggerFactory.getLogger(AcmeClient.class);

    AcmeConfiguration config;

    public AcmeClient(AcmeConfiguration config) {
        this.config = config;
    }

    public Order getCertificateOrder() throws AcmeException, IOException {
        Session session = new Session(URI.create(config.getUri()));
        Account account = new AccountBuilder()
                .agreeToTermsOfService()
                .useKeyPair(this.getUserKeyPair())
                .createLogin(session)
                .getAccount();

        logger.info("Registered a new user, URL: {}", account.getLocation());

        Order order = account.newOrder().domains(config.getDomains()).create();
        return order;
    }

    public Certificate getCertificateCredentials(Order order) throws IOException, AcmeException {
        KeyPair domainKeyPair = this.getDomainKeyPair();
        CSRBuilder csrb = new CSRBuilder();
        csrb.addDomains(this.config.getDomains());
        csrb.sign(domainKeyPair);

        order.execute(csrb.getEncoded());

        try {
            AcmeTasks.completeOrder(order);
        } catch (InterruptedException e) {
            logger.error("Couldn't complete order. Interrupted.");
            throw new AcmeException("Order failed.");
        }

        Certificate certificate = order.getCertificate();
        logger.info("Success! The certificate for domains {} has been generated!", config.getDomains());
        logger.info("Certificate URL: {}", certificate.getLocation());

        return certificate;
    }

    public void authorizeDomain(Order order, NginxClient nginxClient) throws AcmeException {
        for (Authorization auth : order.getAuthorizations()) {
            logger.info("Authorization for domain {}", auth.getIdentifier().getDomain());

            if (auth.getStatus() == Status.VALID) {
                return;
            }

            Http01Challenge challenge = httpChallenge(auth);
            if (challenge.getStatus() == Status.VALID) {
                return;
            }

            boolean success = nginxClient.createChallenge(challenge.getToken(), challenge.getAuthorization());
            if (!success) {
                logger.error("Couldn't create challenge in nginx server");
                throw new AcmeException("Challenge failed.");
            }

            challenge.trigger();
            try {
                AcmeTasks.validateChallenge(challenge);
            } catch (InterruptedException e) {
                logger.error("Couldn't validate challenge. Interrupted.");
                throw new AcmeException("Challenge failed.");
            }

            if (challenge.getStatus() != Status.VALID) {
                throw new AcmeException("Failed to pass the challenge for domain " + auth.getIdentifier().getDomain());
            }

            logger.info("Challenge has been completed.");
            // TODO - Remove validation resource
        }
    }

    private KeyPair getUserKeyPair() throws IOException {
        KeyPair accountKey = AcmeClientUtils.userKeyPair(config.getUserKey(), config.getKeySize());
        return accountKey;
    }

    private KeyPair getDomainKeyPair() throws IOException {
        KeyPair domainKey = AcmeClientUtils.domainKeyPair(config.getDomainKey(), config.getKeySize());
        return domainKey;
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
