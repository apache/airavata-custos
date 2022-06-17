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

package org.apache.custos.credential.store.validator;


import org.apache.custos.core.services.commons.Validator;
import org.apache.custos.credential.store.exceptions.MissingParameterException;
import org.apache.custos.credential.store.service.*;
import org.springframework.stereotype.Component;

/**
 * This class validates the  requests
 */
@Component
public class CredentialStoreInputValidator implements Validator {

    /**
     * Input parameter validater
     *
     * @param methodName
     * @param obj
     * @return
     */
    public <ReqT> ReqT validate(String methodName, ReqT obj) {

        switch (methodName) {
            case "putCredential":
                validatePutCredential(obj, methodName);
                break;
            case "deleteCredential":
                validateDeleteCredential(obj, methodName);
                break;
            case "getCredential":
                validateGetCredential(obj, methodName);
                break;
            case "getAllCredentials":
                validateGetAllCredentials(obj, methodName);
                break;
            case "getNewCustosCredential":
                validateGetNewCustosCredential(obj, methodName);
                break;
            case "getOwnerIdFromToken":
            case "getCredentialFromToken":
            case "getAllCredentialsFromJWTToken":
            case "getAllCredentialsFromToken":
            case "getCustosCredentialFromToken":
            case "validateAgentJWTToken":
                validateTokenRequest(obj, methodName);
                break;
            case "getCredentialFromClientId":
                validateGetCredentialFromClientId(obj, methodName);
                break;
            default:
        }
        return obj;
    }

    private boolean validatePutCredential(Object obj, String method) {
        if (obj instanceof CredentialMetadata) {
            CredentialMetadata metadata = (CredentialMetadata) obj;

            if (metadata.getOwnerId() == 0) {
                throw new MissingParameterException("OwnerId cannot be null at " + method, null);
            }

            if (metadata.getId() == null || metadata.getId().trim().equals("")) {
                throw new MissingParameterException("Id cannot be null at " + method, null);
            }

            if (metadata.getSecret() == null || metadata.getSecret().trim().equals("")) {
                throw new MissingParameterException("Secret cannot be null at " + method, null);
            }

            if (metadata.getType() == null) {
                throw new MissingParameterException("Type cannot be null at " + method, null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method  " + method);
        }
        return true;
    }

    private boolean validateDeleteCredential(Object obj, String method) {
        if (obj instanceof DeleteCredentialRequest) {
            DeleteCredentialRequest metadata = (DeleteCredentialRequest) obj;
            if (metadata.getOwnerId() == 0) {
                throw new MissingParameterException("OwnerId cannot be null at " + method, null);
            }


            if (metadata.getType() == null) {
                throw new MissingParameterException("Type cannot be null at " + method, null);
            }


        } else {
            throw new RuntimeException("Unexpected input type for method  " + method);
        }
        return true;
    }

    private boolean validateGetCredential(Object obj, String method) {
        if (obj instanceof GetCredentialRequest) {
            GetCredentialRequest metadata = (GetCredentialRequest) obj;
            if (metadata.getOwnerId() == 0) {
                throw new MissingParameterException("OwnerId cannot be null at " + method, null);
            }

            if (metadata.getType() == null) {
                throw new MissingParameterException("Type cannot be null at " + method, null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method  " + method);
        }
        return true;

    }

    private boolean validateGetNewCustosCredential(Object obj, String method) {
        if (obj instanceof GetNewCustosCredentialRequest) {
            GetNewCustosCredentialRequest metadata = (GetNewCustosCredentialRequest) obj;
            if (metadata.getOwnerId() == 0) {
                throw new MissingParameterException("OwnerId cannot be null at " + method, null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method  " + method);
        }
        return true;

    }

    private boolean validateTokenRequest(Object obj, String method) {
        if (obj instanceof TokenRequest) {
            TokenRequest metadata = (TokenRequest) obj;
            if (metadata.getToken() == null || metadata.getToken().trim().equals("")) {
                throw new MissingParameterException("Token cannot be null at " + method, null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method  " + method);
        }
        return true;

    }

    private boolean validateGetAllCredentials(Object obj, String method) {
        if (obj instanceof GetAllCredentialsRequest) {
            GetAllCredentialsRequest request = (GetAllCredentialsRequest) obj;
            if (request.getOwnerId() == 0) {
                throw new MissingParameterException("OwnerId cannot be null at " + method, null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method  " + method);
        }
        return true;

    }

    private boolean validateGetCredentialFromClientId(Object obj, String method) {
        if (obj instanceof GetCredentialRequest) {
            GetCredentialRequest request = (GetCredentialRequest) obj;
            if (request.getId() == null || request.getId().equals("")) {
                throw new MissingParameterException("Client Id cannot be null at " + method, null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method  " + method);
        }
        return true;
    }

}
