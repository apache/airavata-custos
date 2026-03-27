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

import io.micrometer.core.instrument.Counter;
import io.micrometer.core.instrument.Timer;
import io.micrometer.core.instrument.simple.SimpleMeterRegistry;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

@Tag("unit")
class AmieMetricsTest {

    private SimpleMeterRegistry registry;
    private AmieMetrics amieMetrics;

    @BeforeEach
    void setUp() {
        registry = new SimpleMeterRegistry();
        amieMetrics = new AmieMetrics(registry);
    }

    @Test
    void recordPacketReceived_shouldIncrementCounterWithTypeTag() {
        amieMetrics.recordPacketReceived("request_project_create");
        amieMetrics.recordPacketReceived("request_project_create");
        amieMetrics.recordPacketReceived("request_user_modify");

        Counter createCounter = registry.find("amie_packets_received_total")
                .tag("type", "request_project_create")
                .counter();
        Counter modifyCounter = registry.find("amie_packets_received_total")
                .tag("type", "request_user_modify")
                .counter();

        assertThat(createCounter).isNotNull();
        assertThat(createCounter.count()).isEqualTo(2.0);
        assertThat(modifyCounter).isNotNull();
        assertThat(modifyCounter.count()).isEqualTo(1.0);
    }

    @Test
    void recordPacketProcessed_shouldIncrementCounterWithTypeAndOutcomeTags() {
        amieMetrics.recordPacketProcessed("request_project_create", "success");
        amieMetrics.recordPacketProcessed("request_project_create", "failure");
        amieMetrics.recordPacketProcessed("request_project_create", "success");

        Counter successCounter = registry.find("amie_packets_processed_total")
                .tag("type", "request_project_create")
                .tag("outcome", "success")
                .counter();
        Counter failureCounter = registry.find("amie_packets_processed_total")
                .tag("type", "request_project_create")
                .tag("outcome", "failure")
                .counter();

        assertThat(successCounter).isNotNull();
        assertThat(successCounter.count()).isEqualTo(2.0);
        assertThat(failureCounter).isNotNull();
        assertThat(failureCounter.count()).isEqualTo(1.0);
    }

    @Test
    void recordRetry_shouldIncrementRetryCounter() {
        amieMetrics.recordRetry();
        amieMetrics.recordRetry();
        amieMetrics.recordRetry();

        Counter retryCounter = registry.find("amie_events_retry_total").counter();

        assertThat(retryCounter).isNotNull();
        assertThat(retryCounter.count()).isEqualTo(3.0);
    }

    @Test
    void startAndStopProcessingTimer_shouldRecordDurationWithHandlerTag() throws InterruptedException {
        Timer.Sample sample = amieMetrics.startProcessingTimer();
        Thread.sleep(5);
        amieMetrics.stopProcessingTimer(sample, "RequestProjectCreateHandler");

        Timer timer = registry.find("amie_packet_processing_duration_seconds")
                .tag("handler", "RequestProjectCreateHandler")
                .timer();

        assertThat(timer).isNotNull();
        assertThat(timer.count()).isEqualTo(1);
        assertThat(timer.totalTime(java.util.concurrent.TimeUnit.MILLISECONDS)).isGreaterThan(0);
    }

    @Test
    void recordPollerFetch_shouldIncrementByCount() {
        amieMetrics.recordPollerFetch(5);
        amieMetrics.recordPollerFetch(3);

        Counter fetchCounter = registry.find("amie_poller_packets_fetched").counter();

        assertThat(fetchCounter).isNotNull();
        assertThat(fetchCounter.count()).isEqualTo(8.0);
    }

    @Test
    void recordPollerFetch_withZeroCount_shouldNotChangeCounter() {
        amieMetrics.recordPollerFetch(0);

        Counter fetchCounter = registry.find("amie_poller_packets_fetched").counter();

        assertThat(fetchCounter).isNotNull();
        assertThat(fetchCounter.count()).isEqualTo(0.0);
    }
}
