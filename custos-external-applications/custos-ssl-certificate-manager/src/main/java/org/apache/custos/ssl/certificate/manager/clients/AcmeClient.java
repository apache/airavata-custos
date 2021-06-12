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

package org.apache.custos.ssl.certificate.manager.clients;

import org.apache.custos.ssl.certificate.manager.configurations.AcmeConfiguration;
import org.shredzone.acme4j.Account;
import org.shredzone.acme4j.AccountBuilder;
import org.shredzone.acme4j.Authorization;
import org.shredzone.acme4j.Certificate;
import org.shredzone.acme4j.Order;
import org.shredzone.acme4j.Session;
import org.shredzone.acme4j.Status;
import org.shredzone.acme4j.challenge.Challenge;
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
import java.util.TimerTask;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

/**
 * Acme client class to perform CA related operations
 */
public class AcmeClient {

    final static int PERIOD = 3;
    final static int RETRY_COUNT = 10;

    private static final Logger logger = LoggerFactory.getLogger(AcmeClient.class);

    private final AcmeConfiguration config;

    public AcmeClient(AcmeConfiguration config) {
        this.config = config;
    }

    /**
     * Order the certificate
     *
     * @param userKey user key file
     * @return the certificate order
     * @throws AcmeException if an acme error occurs
     * @throws IOException   if an I/O error occurs
     */
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
            boolean success = this.completeOrder(order);
            if (!success) {
                logger.error("Process timeout before completing order.");
                throw new AcmeException("Order failed.");
            } else {
                logger.info("Order completes.");
            }
        } catch (InterruptedException e) {
            logger.error("Couldn't complete order. Interrupted.");
            throw new AcmeException("Order failed.");
        }

        Certificate certificate = order.getCertificate();
        logger.info("Success! The certificate for domains {} has been generated!", config.getDomains());
        logger.info("Certificate URL: {}", certificate.getLocation());
        return certificate;
    }

    /**
     * Get challenges for the order
     *
     * @param order Certificate order
     * @return challenges for the order
     * @throws AcmeException                if an acme error occurs
     * @throws UnsupportedEncodingException if an encode error occurs
     */
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

    /**
     * Validates the challenge
     *
     * @param challenge challenge to be validated
     * @return the status of scheduler
     * @throws InterruptedException if an interrupted exception occurs
     */
    public boolean validateChallenge(Challenge challenge) throws InterruptedException {
        return this.processChallenge(challenge);
    }

    private boolean processChallenge(Challenge challenge) throws InterruptedException {
        final ScheduledExecutorService executor = Executors.newSingleThreadScheduledExecutor();
        TimerTask task = new TimerTask() {
            short count = 0;

            @Override
            public void run() {
                try {
                    if (challenge.getStatus() == Status.VALID || count++ > RETRY_COUNT) {
                        executor.shutdown();
                        return;
                    }

                    if (challenge.getStatus() == Status.INVALID) {
                        logger.error("Challenge has failed, reason: {}", challenge.getError());
                        executor.shutdown();
                        return;
                    }

                    challenge.update();
                } catch (AcmeException e) {
                    logger.error("Challenge has failed, reason: {}", e.getMessage());
                }
            }
        };

        executor.scheduleAtFixedRate(task, 0, PERIOD, TimeUnit.SECONDS);
        return executor.awaitTermination(PERIOD * RETRY_COUNT, TimeUnit.SECONDS);
    }

    private boolean completeOrder(Order order) throws InterruptedException {
        final ScheduledExecutorService executor = Executors.newSingleThreadScheduledExecutor();
        TimerTask task = new TimerTask() {
            short count = 0;

            @Override
            public void run() {
                try {
                    if (order.getStatus() == Status.VALID || count++ > RETRY_COUNT) {
                        executor.shutdown();
                        return;
                    }

                    if (order.getStatus() == Status.INVALID) {
                        logger.error("Order has failed, reason: {}", order.getError());
                        executor.shutdown();
                        return;
                    }

                    order.update();
                } catch (AcmeException e) {
                    logger.error("Order has failed, reason: {}", e.getMessage());
                }
            }
        };

        executor.scheduleAtFixedRate(task, 0, PERIOD, TimeUnit.SECONDS);
        return executor.awaitTermination(PERIOD * RETRY_COUNT, TimeUnit.SECONDS);
    }

    private Http01Challenge httpChallenge(Authorization auth) throws AcmeException {
        Http01Challenge challenge = auth.findChallenge(Http01Challenge.class);
        if (challenge == null) {
            throw new AcmeException("Found no " + Http01Challenge.TYPE + " challenge");
        }
        logger.info("Created challenge for domain : {}", auth.getIdentifier().getDomain());
        return challenge;
    }
}
