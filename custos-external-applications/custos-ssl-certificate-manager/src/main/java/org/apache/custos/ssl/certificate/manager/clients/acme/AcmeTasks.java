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

import org.shredzone.acme4j.Order;
import org.shredzone.acme4j.Status;
import org.shredzone.acme4j.challenge.Challenge;
import org.shredzone.acme4j.exception.AcmeException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.TimerTask;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

public class AcmeTasks {

    private static final Logger logger = LoggerFactory.getLogger(AcmeTasks.class);

    final static int PERIOD = 3;
    final static int RETRY_COUNT = 10;

    public static void validateChallenge(final Challenge challenge) throws InterruptedException {
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
        executor.awaitTermination(PERIOD * RETRY_COUNT, TimeUnit.SECONDS);
    }

    public static void completeOrder(final Order order) throws InterruptedException {
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
        executor.awaitTermination(PERIOD * RETRY_COUNT, TimeUnit.SECONDS);
    }
}
