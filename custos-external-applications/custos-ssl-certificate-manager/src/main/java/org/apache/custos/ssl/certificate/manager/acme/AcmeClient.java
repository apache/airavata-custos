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
import java.io.Writer;
import java.net.URI;
import java.security.KeyPair;

public class AcmeClient {

    private static final Logger LOG = LoggerFactory.getLogger(AcmeClient.class);

    AcmeConfiguration config;

    public AcmeClient(AcmeConfiguration config) {
        this.config = config;
    }

    public void authorizeDomain(Order order, NginxClient nginxClient) throws AcmeException, InterruptedException {
        for (Authorization auth : order.getAuthorizations()) {
            LOG.info("Authorization for domain {}", auth.getIdentifier().getDomain());

            if (auth.getStatus() == Status.VALID) {
                return;
            }

            Http01Challenge challenge = httpChallenge(auth);
            if (challenge.getStatus() == Status.VALID) {
                return;
            }

            boolean success = nginxClient.createChallenge(challenge.getToken(), challenge.getAuthorization());
            if (!success){
                LOG.error("Couldn't create challenge in nginx server");
                throw new AcmeException("Challenge failed.");
            }

            challenge.trigger();

            try {
                int attempts = 10;
                while (challenge.getStatus() != Status.VALID && attempts-- > 0) {
                    if (challenge.getStatus() == Status.INVALID) {
                        LOG.error("Challenge has failed, reason: {}", challenge.getError());
                        throw new AcmeException("Challenge failed.");
                    }

                    Thread.sleep(3000L);
                    challenge.update();
                }
            } catch (InterruptedException ex) {
                LOG.error("Interrupted", ex);
                Thread.currentThread().interrupt();
            }

            if (challenge.getStatus() != Status.VALID) {
                throw new AcmeException("Failed to pass the challenge for domain " + auth.getIdentifier().getDomain());
            }

            LOG.info("Challenge has been completed.");
            // TODO - Remove validation resource
        }
    }

    public Order getCertificateOrder() throws AcmeException, IOException {
        Session session = new Session(URI.create(config.getSessionUri()));
        KeyPair accountKey = AcmeClientUtils.userKeyPair(config.getUserKey(), config.getKeySize());
        Account account = new AccountBuilder()
                .agreeToTermsOfService()
                .useKeyPair(accountKey)
                .createLogin(session)
                .getAccount();

        LOG.info("Registered a new user, URL: {}", account.getLocation());

        Order order = account.newOrder().domains(config.getDomains()).create();
        return order;
    }

    public void getCertificates(Order order) throws IOException, AcmeException {
        KeyPair domainKeyPair = AcmeClientUtils.domainKeyPair(config.getDomainKey(), config.getKeySize());

        CSRBuilder csrb = new CSRBuilder();
        csrb.addDomains(this.config.getDomains());
        csrb.sign(domainKeyPair);

        try (Writer out = new FileWriter(config.getDomainCsr())) {
            csrb.write(out);
        }

        order.execute(csrb.getEncoded());

        try {
            int attempts = 10;
            while (order.getStatus() != Status.VALID && attempts-- > 0) {
                if (order.getStatus() == Status.INVALID) {
                    LOG.error("Order has failed, reason: {}", order.getError());
                    throw new AcmeException("Order failed");
                }

                Thread.sleep(3000L);
                order.update();
            }
        } catch (InterruptedException ex) {
            LOG.error("Interrupted", ex);
            Thread.currentThread().interrupt();
        }

        Certificate certificate = order.getCertificate();
        LOG.info("Success! The certificate for domains {} has been generated!", config.getDomains());
        LOG.info("Certificate URL: {}", certificate.getLocation());

        try (FileWriter fw = new FileWriter(config.getDomainChain())) {
            certificate.writeCertificate(fw);
        }
    }

    private Http01Challenge httpChallenge(Authorization auth) throws AcmeException {
        Http01Challenge challenge = auth.findChallenge(Http01Challenge.class);
        if (challenge == null) {
            throw new AcmeException("Found no " + Http01Challenge.TYPE + " challenge");
        }

        LOG.info("Domain : {}", auth.getIdentifier().getDomain());
        LOG.info("File : {}", challenge.getToken());
        LOG.info("Content: {}", challenge.getAuthorization());
        return challenge;
    }
}
