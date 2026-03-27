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
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.boot.actuate.health.Health;
import org.springframework.boot.actuate.health.Status;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.client.ResourceAccessException;
import org.springframework.web.client.RestTemplate;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
@Tag("unit")
class AmieHealthIndicatorTest {

    private static final String BASE_URL = "https://a3mdev.xsede.org/amie-api-test";
    private static final String SITE_CODE = "NEXUS";

    @Mock
    private RestTemplate restTemplate;

    private AmieProperties amieProperties;
    private AmieHealthIndicator healthIndicator;

    @BeforeEach
    void setUp() {
        amieProperties = new AmieProperties();
        amieProperties.setBaseUrl(BASE_URL);
        amieProperties.setSiteCode(SITE_CODE);
        healthIndicator = new AmieHealthIndicator(amieProperties, restTemplate);
    }

    @Test
    void health_whenAmieApiReturns200_shouldBeUp() {
        when(restTemplate.getForEntity(BASE_URL, String.class))
                .thenReturn(ResponseEntity.ok("OK"));

        Health health = healthIndicator.health();

        assertThat(health.getStatus()).isEqualTo(Status.UP);
        assertThat(health.getDetails()).containsEntry("url", BASE_URL);
        assertThat(health.getDetails()).containsEntry("siteCode", SITE_CODE);
        assertThat(health.getDetails()).containsEntry("httpStatus", 200);
    }

    @Test
    void health_whenAmieApiReturns302_shouldBeUp() {
        when(restTemplate.getForEntity(BASE_URL, String.class))
                .thenReturn(ResponseEntity.status(HttpStatus.FOUND).build());

        Health health = healthIndicator.health();

        assertThat(health.getStatus()).isEqualTo(Status.UP);
        assertThat(health.getDetails()).containsEntry("httpStatus", 302);
    }

    @Test
    void health_whenAmieApiReturns500_shouldBeDown() {
        when(restTemplate.getForEntity(BASE_URL, String.class))
                .thenReturn(ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body("error"));

        Health health = healthIndicator.health();

        assertThat(health.getStatus()).isEqualTo(Status.DOWN);
        assertThat(health.getDetails()).containsEntry("url", BASE_URL);
        assertThat(health.getDetails()).containsEntry("httpStatus", 500);
    }

    @Test
    void health_whenRestClientExceptionThrown_shouldBeDown() {
        when(restTemplate.getForEntity(BASE_URL, String.class))
                .thenThrow(new ResourceAccessException("Connection refused"));

        Health health = healthIndicator.health();

        assertThat(health.getStatus()).isEqualTo(Status.DOWN);
        assertThat(health.getDetails()).containsEntry("url", BASE_URL);
        assertThat(health.getDetails()).containsEntry("siteCode", SITE_CODE);
    }

    @Test
    void health_whenNetworkTimeout_shouldIncludeUrlInDownDetails() {
        when(restTemplate.getForEntity(BASE_URL, String.class))
                .thenThrow(new ResourceAccessException("Read timed out"));

        Health health = healthIndicator.health();

        assertThat(health.getStatus()).isEqualTo(Status.DOWN);
        assertThat(health.getDetails()).containsKey("url");
        assertThat(health.getDetails().get("url")).isEqualTo(BASE_URL);
    }
}
