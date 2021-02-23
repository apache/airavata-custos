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

package org.apache.custos.resource.secret.manager.adaptor.outbound;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.protobuf.GeneratedMessageV3;
import org.json.JSONObject;

import java.util.HashMap;
import java.util.Map;

public class CredentialMap extends ResourceCredential {

    private Map<String, String> credentialMap;

    private JSONObject json;

    private String credentialString;

    public CredentialMap(GeneratedMessageV3 message) {
        super(message);
        if (message instanceof org.apache.custos.resource.secret.service.CredentialMap) {
            this.credentialMap = ((org.apache.custos.resource.secret.service.CredentialMap) message).getCredentialMapMap();
            if (this.credentialMap != null &&  !this.credentialMap.isEmpty()) {
                this.json = new JSONObject(this.credentialMap);
            }
        }
    }

    public Map<String, String> getCredentialMap() {
        return credentialMap;
    }

    public void setCredentialMap(Map<String, String> credentialMap) {
        this.credentialMap = credentialMap;
    }

    public JSONObject getJson() {
        return json;
    }

    public void setJson(JSONObject json) {
        this.json = json;
    }

    public String getCredentialString() {
        return this.json != null ? this.json.toString() : null;
    }

    public static Map<String, String> getCredentialMapFromString(String jsonString) throws JsonProcessingException {
        Map<String, String> map = new HashMap<String, String>();
        ObjectMapper mapper = new ObjectMapper();
        return mapper.readValue(jsonString,
                new TypeReference<HashMap<String, String>>() {
                });
    }
}
