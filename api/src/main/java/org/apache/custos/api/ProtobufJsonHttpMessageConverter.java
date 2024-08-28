/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */

package org.apache.custos.api;

import com.google.protobuf.Message;
import com.google.protobuf.util.JsonFormat;
import org.springframework.http.HttpInputMessage;
import org.springframework.http.HttpOutputMessage;
import org.springframework.http.MediaType;
import org.springframework.http.converter.AbstractHttpMessageConverter;
import org.springframework.http.converter.HttpMessageConversionException;
import org.springframework.http.converter.HttpMessageNotReadableException;
import org.springframework.http.converter.HttpMessageNotWritableException;

import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;

public class ProtobufJsonHttpMessageConverter extends AbstractHttpMessageConverter<Message> {

    public ProtobufJsonHttpMessageConverter() {
        super(new MediaType("application", "json", StandardCharsets.UTF_8));
    }

    @Override
    protected boolean supports(Class<?> clazz) {
        return Message.class.isAssignableFrom(clazz);
    }

    @Override
    protected Message readInternal(Class<? extends Message> clazz, HttpInputMessage inputMessage) throws HttpMessageNotReadableException {
        try (InputStreamReader reader = new InputStreamReader(inputMessage.getBody(), StandardCharsets.UTF_8)) {
            Message.Builder builder = (Message.Builder) clazz.getMethod("newBuilder").invoke(null);
            JsonFormat.parser().ignoringUnknownFields().merge(reader, builder);
            return builder.build();
        } catch (Exception e) {
            throw new HttpMessageNotReadableException("Error reading Protobuf message", e, inputMessage);
        }
    }

    @Override
    protected void writeInternal(Message message, HttpOutputMessage outputMessage) throws HttpMessageNotWritableException {
        try {
            outputMessage.getBody().write(JsonFormat.printer().preservingProtoFieldNames().print(message).getBytes(StandardCharsets.UTF_8));
        } catch (Exception e) {
            throw new HttpMessageConversionException("Error writing Protobuf message", e);
        }
    }

}
