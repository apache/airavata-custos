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

package org.apache.custos.logging.validator;

import org.apache.custos.core.services.commons.Validator;
import org.apache.custos.core.services.commons.exceptions.MissingParameterException;
import org.apache.custos.logging.service.LogEvent;
import org.apache.custos.logging.service.LogEventRequest;
import org.apache.custos.logging.service.LoggingConfigurationRequest;
import org.springframework.stereotype.Component;


/**
 * This class validates the  requests
 */
@Component
public class CustosLoggingInputValidator implements Validator {

    /**
     * Input parameter validater
     *
     * @param methodName
     * @param obj
     * @return
     */
    public <ReqT> ReqT validate(String methodName, ReqT obj) {

        switch (methodName) {
            case "addLogEvent":
                validateAddLogEvent(methodName, obj);
                break;

            case "getLogEvents":
                validateGetLogEvents(methodName, obj);
                break;
            case "isLogEnabled":
                validateIsLogEnabled(methodName, obj);
                break;
            case "enable":
                validateEnable(methodName, obj);
                break;
        }
        return obj;
    }


    private boolean validateAddLogEvent(String methodName, Object obj) {
        if (obj instanceof LogEvent) {
            LogEvent entityTypeRequest = (LogEvent) obj;
            if (entityTypeRequest.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id not found ", null);
            }
            if (entityTypeRequest.getClientId() == null || entityTypeRequest.getClientId().equals("")) {
                throw new MissingParameterException("Client Id not found ", null);
            }
            if (entityTypeRequest.getServiceName() == null || entityTypeRequest.getServiceName().equals("")) {
                throw new MissingParameterException("Service name  not found ", null);
            }

            if (entityTypeRequest.getEventType() == null || entityTypeRequest.getEventType().equals("")) {
                throw new MissingParameterException("Event type  not found ", null);
            }

            if (entityTypeRequest.getExternalIp() == null || entityTypeRequest.getExternalIp().equals("")) {
                throw new MissingParameterException("External Ip  not found ", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method " + methodName);
        }
        return true;
    }

    private boolean validateGetLogEvents(String methodName, Object obj) {
        if (obj instanceof LogEventRequest) {
            LogEventRequest entityTypeRequest = (LogEventRequest) obj;
            if (entityTypeRequest.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id not found ", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method " + methodName);
        }
        return true;
    }

    private boolean validateIsLogEnabled(String methodName, Object obj) {
        if (obj instanceof LoggingConfigurationRequest) {
            LoggingConfigurationRequest entityTypeRequest = (LoggingConfigurationRequest) obj;
            if (entityTypeRequest.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id not found ", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method " + methodName);
        }
        return true;
    }

    private boolean validateEnable(String methodName, Object obj) {
        if (obj instanceof LoggingConfigurationRequest) {
            LoggingConfigurationRequest entityTypeRequest = (LoggingConfigurationRequest) obj;
            if (entityTypeRequest.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id not found ", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method " + methodName);
        }
        return true;
    }


}
