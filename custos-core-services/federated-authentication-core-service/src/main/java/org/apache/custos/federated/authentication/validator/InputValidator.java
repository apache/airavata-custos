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

package org.apache.custos.federated.authentication.validator;


import org.apache.custos.core.services.commons.Validator;
import org.apache.custos.federated.authentication.exceptions.MissingParameterException;
import org.apache.custos.federated.authentication.service.ClientMetadata;
import org.apache.custos.federated.authentication.service.DeleteClientRequest;
import org.apache.custos.federated.authentication.service.GetClientRequest;

/**
 * This class validates the  requests
 */
public class InputValidator implements Validator {

    /**
     * Input parameter validater
     * @param methodName
     * @param obj
     * @return
     */
    public  void validate(String methodName, Object obj) {

        switch (methodName) {
            case "addClient":
            case "updateClient":
              validateClientMetadata(obj, methodName);
              break;
            case "getClient":
                validateGetClientRequest(obj);
                break;
            case "deleteClient":
                validateDeleteClientRequest(obj);
                break;
            default:
        }
    }

    private  boolean validateClientMetadata(Object obj, String method) {
        if (obj instanceof ClientMetadata) {
          ClientMetadata metadata = (ClientMetadata)obj;
          if (metadata.getTenantId() == 0) {
              throw new MissingParameterException("Tenant Id should not be null", null);
          }

          if (metadata.getTenantName() == null || metadata.getTenantName().trim().equals("")) {
              throw new MissingParameterException("Tenant name should not be null", null);
          }

          if (metadata.getRedirectURIsList() == null || metadata.getRedirectURIsCount() == 0) {
              throw new MissingParameterException("RedirectURIs should not be null", null);
          }
          if (metadata.getComment() == null || metadata.getComment().trim().equals("")) {
              throw new MissingParameterException("Comment should not be null", null);
          }

          if (metadata.getPerformedBy() == null || metadata.getPerformedBy().trim().equals("")) {
              throw new MissingParameterException("Performed by should not be null", null);
          }

        } else {
            throw new RuntimeException("Unexpected input type for method  "+method);
        }
        return true;
    }

    private  boolean validateGetClientRequest(Object obj) {
        if (obj instanceof GetClientRequest) {
            GetClientRequest request = (GetClientRequest)obj;

            if (request.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id should not be null", null);
            }

            if (request.getClientId() == null || request.getClientId().trim().equals("")) {
                throw new MissingParameterException("Client Id should not be null", null);
            }


        } else {
            throw new RuntimeException("Unexpected input type for method getClient");
        }
        return true;
    }

    private  boolean validateDeleteClientRequest(Object obj) {
        if (obj instanceof DeleteClientRequest) {
            DeleteClientRequest request = (DeleteClientRequest) obj;

            if (request.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id should not be null", null);
            }

            if (request.getClientId() == null || request.getClientId().trim().equals("")) {
                throw new MissingParameterException("Client Id should not be null", null);
            }

            if (request.getPerformedBy() == null || request.getPerformedBy().trim().equals("")) {
                throw new MissingParameterException("Performed By should not be null", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method deleteClient");
        }
        return true;
    }


}
