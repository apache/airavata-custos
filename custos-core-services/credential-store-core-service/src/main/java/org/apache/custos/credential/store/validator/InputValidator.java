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


import org.apache.custos.credential.store.exceptions.MissingParameterException;
import org.apache.custos.credential.store.service.*;

/**
 * This class validates the  requests
 */
public class InputValidator {

    /**
     * Input parameter validater
     *
     * @param methodName
     * @param obj
     * @return
     */
    public static void validate(String methodName, Object obj) {

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
                validateGetOwnerIdFromToken(obj, methodName);
                break;
            default:
        }
    }

    private static boolean validatePutCredential(Object obj, String method) {
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

    private static boolean validateDeleteCredential(Object obj, String method) {
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

    private static boolean validateGetCredential(Object obj, String method) {
        if (obj instanceof GetCredentialRequest) {
            GetCredentialRequest metadata = (GetCredentialRequest) obj;
            if (metadata.getOwnerId() == 0) {
                throw new MissingParameterException("OwnerId cannot be null at " + method, null);
            }

            if (metadata.getId() == null || metadata.getId().trim().equals("")) {
                throw new MissingParameterException("Id cannot be null at " + method, null);
            }


            if (metadata.getType() == null) {
                throw new MissingParameterException("Type cannot be null at " + method, null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method  " + method);
        }
        return true;

    }

    private static boolean validateGetNewCustosCredential (Object obj, String method) {
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

    private static boolean validateGetOwnerIdFromToken (Object obj, String method) {
        if (obj instanceof GetOwnerIdFromTokenRequest) {
            GetOwnerIdFromTokenRequest metadata = (GetOwnerIdFromTokenRequest) obj;
            if (metadata.getToken() == null || metadata.getToken().trim().equals("") ){
                throw new MissingParameterException("OwnerId cannot be null at " + method, null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method  " + method);
        }
        return true;

    }

    private static boolean validateGetAllCredentials(Object obj, String method) {
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

}
