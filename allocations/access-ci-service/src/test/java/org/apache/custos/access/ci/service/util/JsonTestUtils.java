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
package org.apache.custos.access.ci.service.util;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;

/**
 * Utility class for loading and validating JSON content
 */
public final class JsonTestUtils {

    private static final ObjectMapper OBJECT_MAPPER = new ObjectMapper();

    private JsonTestUtils() {
    }

    public static JsonNode loadJsonFromFile(String resourcePath) {
        try {
            String content = Files.readString(Path.of("src/test/resources", resourcePath));
            return OBJECT_MAPPER.readTree(content);

        } catch (IOException e) {
            throw new RuntimeException("Failed to load JSON from: " + resourcePath, e);
        }
    }

    public static JsonNode loadMockPacket(String handlerType, String filename) {
        String path = "mock-data/" + handlerType + "/" + filename;
        return loadJsonFromFile(path);
    }
}