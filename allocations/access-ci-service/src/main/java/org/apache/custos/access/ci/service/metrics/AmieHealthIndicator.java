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
 *  specific language governing permissions and limitations
 *  under the License.
 */
package org.apache.custos.access.ci.service.metrics;

import org.apache.custos.access.ci.service.config.AmieProperties;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.boot.actuate.health.AbstractHealthIndicator;
import org.springframework.boot.actuate.health.Health;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestClientException;
import org.springframework.web.client.RestTemplate;

/**
 * Actuator health indicator for the upstream AMIE API.
 *
 * <p>Performs an HTTP GET to the configured AMIE base URL and reports
 * {@link Health#up()} when the endpoint responds, or {@link Health#down()} when
 * a network or HTTP error occurs.
 */
@Component
public class AmieHealthIndicator extends AbstractHealthIndicator {

    private static final Logger LOGGER = LoggerFactory.getLogger(AmieHealthIndicator.class);

    private final AmieProperties amieProperties;
    private final RestTemplate restTemplate;

    public AmieHealthIndicator(AmieProperties amieProperties,
                               @Qualifier("amieRestTemplate") RestTemplate restTemplate) {
        super("AMIE API health check failed");
        this.amieProperties = amieProperties;
        this.restTemplate = restTemplate;
    }

    @Override
    protected void doHealthCheck(Health.Builder builder) {
        String baseUrl = amieProperties.getBaseUrl();
        String siteCode = amieProperties.getSiteCode();
        try {
            ResponseEntity<String> response = restTemplate.getForEntity(baseUrl, String.class);
            if (response.getStatusCode().is2xxSuccessful() || response.getStatusCode().is3xxRedirection()) {
                builder.up()
                        .withDetail("url", baseUrl)
                        .withDetail("siteCode", siteCode)
                        .withDetail("httpStatus", response.getStatusCode().value());
            } else {
                builder.down()
                        .withDetail("url", baseUrl)
                        .withDetail("siteCode", siteCode)
                        .withDetail("httpStatus", response.getStatusCode().value());
            }
        } catch (RestClientException ex) {
            LOGGER.warn("AMIE API health check failed for URL [{}]: {}", baseUrl, ex.getMessage());
            builder.down(ex)
                    .withDetail("url", baseUrl)
                    .withDetail("siteCode", siteCode);
        }
    }
}
