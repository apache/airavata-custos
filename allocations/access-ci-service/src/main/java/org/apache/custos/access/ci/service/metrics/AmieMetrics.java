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
import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.Timer;
import org.springframework.stereotype.Component;

/**
 * Metrics for the AMIE packet processing pipeline.
 *
 * <p>All metric names follow the Prometheus naming convention:
 * {@code amie_<subsystem>_<measurement>_<unit>}.
 */
@Component
public class AmieMetrics {

    private static final String PACKETS_RECEIVED_TOTAL = "amie_packets_received_total";
    private static final String PACKETS_PROCESSED_TOTAL = "amie_packets_processed_total";
    private static final String EVENTS_RETRY_TOTAL = "amie_events_retry_total";
    private static final String PROCESSING_DURATION_SECONDS = "amie_packet_processing_duration_seconds";
    private static final String POLLER_PACKETS_FETCHED = "amie_poller_packets_fetched";

    private static final String TAG_TYPE = "type";
    private static final String TAG_OUTCOME = "outcome";
    private static final String TAG_HANDLER = "handler";

    private final MeterRegistry meterRegistry;

    public AmieMetrics(MeterRegistry meterRegistry) {
        this.meterRegistry = meterRegistry;
    }

    /**
     * Increments the counter tracking raw packets received from the AMIE API, tagged by packet type.
     *
     * @param packetType the AMIE packet type (e.g., "request_project_create")
     */
    public void recordPacketReceived(String packetType) {
        Counter.builder(PACKETS_RECEIVED_TOTAL)
                .tag(TAG_TYPE, packetType)
                .description("Total number of AMIE packets received from the API")
                .register(meterRegistry)
                .increment();
    }

    /**
     * Increments the counter tracking processed packets, tagged by packet type and outcome.
     *
     * @param packetType the AMIE packet type (e.g., "request_project_create")
     * @param outcome    the processing outcome (e.g., "success", "failure")
     */
    public void recordPacketProcessed(String packetType, String outcome) {
        Counter.builder(PACKETS_PROCESSED_TOTAL)
                .tag(TAG_TYPE, packetType)
                .tag(TAG_OUTCOME, outcome)
                .description("Total number of AMIE packets that completed processing")
                .register(meterRegistry)
                .increment();
    }

    /**
     * Increments the counter tracking event retry attempts.
     */
    public void recordRetry() {
        Counter.builder(EVENTS_RETRY_TOTAL)
                .description("Total number of AMIE processing event retry attempts")
                .register(meterRegistry)
                .increment();
    }

    /**
     * Starts a timer sample for measuring packet processing duration.
     *
     * @return a {@link Timer.Sample} that must be stopped via {@link #stopProcessingTimer}
     */
    public Timer.Sample startProcessingTimer() {
        return Timer.start(meterRegistry);
    }

    /**
     * Stops a previously started timer sample and records the duration against the
     * {@code amie_packet_processing_duration_seconds} timer, tagged by handler type.
     *
     * @param sample      the sample returned by {@link #startProcessingTimer()}
     * @param handlerType the name of the handler that processed the packet
     */
    public void stopProcessingTimer(Timer.Sample sample, String handlerType) {
        Timer timer = Timer.builder(PROCESSING_DURATION_SECONDS)
                .tag(TAG_HANDLER, handlerType)
                .description("Time taken to process an AMIE packet by handler type")
                .register(meterRegistry);
        sample.stop(timer);
    }

    /**
     * Increments the counter tracking the number of packets fetched during a poller run.
     *
     * @param count the number of packets fetched
     */
    public void recordPollerFetch(int count) {
        Counter.builder(POLLER_PACKETS_FETCHED)
                .description("Total number of AMIE packets fetched by the poller")
                .register(meterRegistry)
                .increment(count);
    }
}
