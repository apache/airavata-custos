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

package org.apache.custos.messaging.events.model;


import com.google.protobuf.util.JsonFormat;
import org.apache.custos.messaging.service.Message;
import org.apache.kafka.common.serialization.Deserializer;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Map;

public class MessageDeserializer implements Deserializer<Message> {
    private static final Logger LOGGER = LoggerFactory.getLogger(MessageDeserializer.class);

    @Override
    public void configure(Map<String, ?> map, boolean b) {

    }

    @Override
    public Message deserialize(String topic, byte[] bytes) {
        try {
            String deserialized = new String(bytes);
            final var jsonParser = JsonFormat.parser();
            Message.Builder messageBuilder = Message.newBuilder();
            jsonParser.merge(deserialized, messageBuilder);
            return messageBuilder.build();
        } catch (Exception ex) {
            String msg = "Error occurred while processing message " + ex.getMessage();
            LOGGER.error(msg, ex);
        }
        return null;
    }

    @Override
    public void close() {

    }
}
