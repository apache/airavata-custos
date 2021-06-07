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
import java.io.UnsupportedEncodingException;
import java.net.URI;
import java.security.KeyPair;
import java.util.ArrayList;

public class AcmeClient {

    private static final Logger logger = LoggerFactory.getLogger(AcmeClient.class);
    private final AcmeConfiguration config;

    public AcmeClient(AcmeConfiguration config) {
        this.config = config;
    }

    public Order getCertificateOrder(KeyPair userKey) throws AcmeException, IOException {
        Session session = new Session(URI.create(config.getUrl()));
        Account account = new AccountBuilder()
                .agreeToTermsOfService()
                .useKeyPair(userKey)
                .createLogin(session)
                .getAccount();
        logger.info("Registered user, URL: {}", account.getLocation());
        return account.newOrder().domains(config.getDomains()).create();
    }

    public Certificate getCertificateCredentials(Order order, KeyPair domainKeyPair) throws IOException, AcmeException {
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

    public ArrayList<Http01Challenge> getChallenges(Order order) throws AcmeException, UnsupportedEncodingException {
        ArrayList<Http01Challenge> challenges = new ArrayList<>();
        for (Authorization auth : order.getAuthorizations()) {
            logger.info("Creating challenge for {}", auth.getIdentifier().getDomain());
            if (auth.getStatus() == Status.VALID) {
                continue;
            }

            Http01Challenge challenge = httpChallenge(auth);
            if (challenge.getStatus() == Status.VALID) {
                continue;
            }

            challenges.add(challenge);
        }
        return challenges;
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
