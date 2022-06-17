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

package org.apache.custos.agent.profile.validator;

import org.apache.custos.agent.profile.service.AgentRequest;
import org.apache.custos.core.services.commons.Validator;
import org.apache.custos.core.services.commons.exceptions.MissingParameterException;
import org.springframework.stereotype.Component;

/**
 * This class validates the  requests
 */
@Component
public class AgentInputValidator implements Validator {

    /**
     * Input parameter validater
     *
     * @param methodName
     * @param obj
     * @return
     */
    public <ReqT> ReqT validate(String methodName, ReqT obj) {

        switch (methodName) {
            case "createAgent":
            case "updateAgent":
            case "deleteAgent":
            case "getAgent":
                validateAgentRequest(obj);
                break;
            default:

        }
        return obj;
    }

    private boolean validateAgentRequest(Object obj) {
        if (obj instanceof AgentRequest) {
            AgentRequest request = (AgentRequest) obj;
            if (request.getTenantId() == 0) {
                throw new MissingParameterException("TenantId should not be null", null);
            }
            if (request.getAgent() == null) {
                throw new MissingParameterException("Agent should not be null", null);
            }

            if (request.getAgent().getId() == null || request.getAgent().getId().equals("")) {
                throw new MissingParameterException("AgentId should not be null", null);
            }

        }
        return true;
    }


}
