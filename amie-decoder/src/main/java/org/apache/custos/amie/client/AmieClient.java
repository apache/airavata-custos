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
package org.apache.custos.amie.client;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.custos.amie.config.AmieProperties;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpMethod;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestClientException;
import org.springframework.web.client.RestTemplate;
import org.springframework.web.util.UriComponentsBuilder;

import java.net.URI;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

/**
 * A client for Access CI AMIE REST API
 */
@Component
public class AmieClient {

    private static final Logger LOGGER = LoggerFactory.getLogger(AmieClient.class);
    private static final String PACKETS_PATH = "/packets/{site}/";
    private static final String HEADER_XA_SITE = "XA-SITE";
    private static final String HEADER_XA_API_KEY = "XA-API-KEY";

    private final RestTemplate restTemplate;
    private final AmieProperties properties;
    private final ObjectMapper objectMapper = new ObjectMapper();

    public AmieClient(@Qualifier("amieRestTemplate") RestTemplate restTemplate, AmieProperties properties) {
        this.restTemplate = restTemplate;
        this.properties = properties;
    }

    /**
     * Fetches all in-progress packets from the AMIE API.
     *
     * @return A list of packets. Returns an empty list if the call fails or no packets are found.
     */
    public List<JsonNode> fetchInProgressPackets() {
        URI uri = UriComponentsBuilder.fromHttpUrl(properties.getBaseUrl())
                .path(PACKETS_PATH)
                .build(properties.getSiteCode());

        HttpHeaders headers = new HttpHeaders();
        headers.set(HEADER_XA_SITE, properties.getSiteCode());
        headers.set(HEADER_XA_API_KEY, properties.getApiKey());
        headers.setAccept(Collections.singletonList(MediaType.APPLICATION_JSON));

        HttpEntity<String> entity = new HttpEntity<>(headers);

        try {
            LOGGER.debug("Fetching in-progress packets from {}", uri);
            ResponseEntity<String> response = restTemplate.exchange(uri, HttpMethod.GET, entity, String.class);

            if (!response.getStatusCode().is2xxSuccessful() || response.getBody() == null) {
                LOGGER.warn("Received a non-successful response from AMIE API: {} {}", response.getStatusCode(), response.getBody());
                return Collections.emptyList();
            }

            return parsePacketsFromResponse(response.getBody());

        } catch (RestClientException e) {
            LOGGER.error("Failed to fetch packets from AMIE API at {}", uri, e);
            return Collections.emptyList();
        }
    }

    private List<JsonNode> parsePacketsFromResponse(String responseBody) {
        List<JsonNode> packets = new ArrayList<>();
        try {
            JsonNode root = objectMapper.readTree(responseBody);

            // According to AMIE documentation, the response might be a direct array or an object
            // containing a "result" key which holds the data.
            if (root.isArray()) {
                root.forEach(packets::add);
            } else if (root.has("result")) {
                JsonNode resultNode = root.get("result");
                if (resultNode.isArray()) {
                    resultNode.forEach(packets::add);
                } else {
                    packets.add(resultNode);
                }
            } else {
                // Handle cases when the response is a single packet object not in an array
                packets.add(root);
            }
            LOGGER.info("Successfully fetched and parsed {} packets from AMIE API.", packets.size());
            return packets;

        } catch (Exception e) {
            LOGGER.error("Failed to parse JSON response from AMIE API. Response body: {}", responseBody, e);
            return Collections.emptyList();
        }
    }
}
