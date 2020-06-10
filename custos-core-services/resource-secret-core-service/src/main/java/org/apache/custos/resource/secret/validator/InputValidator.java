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

package org.apache.custos.resource.secret.validator;


import org.apache.custos.core.services.commons.Validator;
import org.apache.custos.core.services.commons.exceptions.MissingParameterException;
import org.apache.custos.resource.secret.service.*;


/**
 * This class validates the  requests
 */
public class InputValidator implements Validator {

    /**
     * Input parameter validater
     *
     * @param methodName
     * @param obj
     * @return
     */
    public void validate(String methodName, Object obj) {

        switch (methodName) {
            case "getAllResourceCredentialSummaries":
                validateGetAllResourceCredentialSummaries(obj, methodName);
                break;
            case "addSSHCredential":
                validateAddSSHCredential(obj, methodName);
                break;
            case "addPasswordCredential":
                validateAddPasswordCredential(obj, methodName);
                break;
            case "addCertificateCredential":
                validateAddCertificateCredential(obj, methodName);
                break;
            case "getPasswordCredential":
            case "getCertificateCredential":
            case "getSSHCredential":
            case "deleteSSHCredential":
            case "deletePWDCredential":
            case "getResourceCredentialSummary":
                validateGetResourceCredentialByToken(obj, methodName);
                break;
            default:
        }
    }

    private boolean validateGetResourceCredentialByToken(Object obj, String methodName) {
        if (obj instanceof GetResourceCredentialByTokenRequest) {
            GetResourceCredentialByTokenRequest request = (GetResourceCredentialByTokenRequest) obj;

            if (request.getTenantId() == 0) {
                throw new MissingParameterException("TenantId should be set", null);
            }

            if (request.getToken() == null || request.getToken().equals("")) {
                throw new MissingParameterException("Token should not be null", null);
            }


        } else {
            throw new RuntimeException("Unexpected input type for method  " + methodName);
        }
        return true;


    }

    private boolean validateGetAllResourceCredentialSummaries(Object obj, String methodName) {
        if (obj instanceof GetResourceCredentialSummariesRequest) {
            GetResourceCredentialByTokenRequest request = (GetResourceCredentialByTokenRequest) obj;

            if (request.getTenantId() == 0) {
                throw new MissingParameterException("TenantId should be set", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method  " + methodName);
        }
        return true;
    }

    private boolean validateAddSSHCredential(Object obj, String methodName) {
        if (obj instanceof SSHCredential) {
            SSHCredential request = (SSHCredential) obj;

            validateSecretMetadata(request.getMetadata());

        } else {
            throw new RuntimeException("Unexpected input type for method  " + methodName);
        }
        return true;
    }

    private boolean validateAddPasswordCredential(Object obj, String methodName) {
        if (obj instanceof PasswordCredential) {
            PasswordCredential request = (PasswordCredential) obj;

            validateSecretMetadata(request.getMetadata());

            if (request.getPassword() == null || request.getPassword().trim().equals("")) {
                throw new MissingParameterException("Password should not be  null ", null);
            }


        } else {
            throw new RuntimeException("Unexpected input type for method  " + methodName);
        }
        return true;
    }

    private boolean validateAddCertificateCredential(Object obj, String methodName) {
        if (obj instanceof CertificateCredential) {
            CertificateCredential request = (CertificateCredential) obj;

            validateSecretMetadata(request.getMetadata());

            if (request.getX509Cert() == null || request.getX509Cert().trim().equals("")) {
                throw new MissingParameterException("Certificate should not be null", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method  " + methodName);
        }
        return true;
    }


    private boolean validateSecretMetadata(SecretMetadata metadata) {

        if (metadata.getOwnerId() == null || metadata.getOwnerId().trim().equals("")) {
            throw new MissingParameterException("OwnerId should not be null", null);
        }

        if (metadata.getTenantId() == 0) {
            throw new MissingParameterException("TenantId should be set", null);
        }

        return true;
    }


}
