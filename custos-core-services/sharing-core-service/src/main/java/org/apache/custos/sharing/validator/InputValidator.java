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

package org.apache.custos.sharing.validator;


import org.apache.custos.core.services.commons.Validator;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;


/**
 * This class validates the  request parameter
 */
public class InputValidator implements Validator {

    private static final Logger LOGGER = LoggerFactory.getLogger(InputValidator.class);

    /**
     * Input parameter validater
     *
     * @param methodName
     * @param obj
     * @return
     */
    public void validate(String methodName, Object obj) {

        switch (methodName) {
            case "createEntityType":
            case "updateEntityType":
                validateEntityTypeRequest(obj, methodName);
                break;
            case "deleteEntityType":
            case "getEntityType":
                validateEntityTypeIdentifyingRequest(obj, methodName);
                break;
            case "getEntityTypes":
                validateTenantIdOnly(obj, methodName);
                break;
            case "createPermissionType":
            case "updatePermissionType":
            case "deletePermissionType":
            case "getPermissionType":
                validateCreatePermissionTypeRequest(obj, methodName);
                break;
            case "getPermissionTypes":
                validateTenantIdOnly(obj, methodName);
                break;
            case "createEntity":
            case "updateEntity":
                validateEntityRequest(obj, methodName);
                break;
            case "isEntityExists":
            case "getEntity":
            case "deleteEntity":
                validateEntityIdentifierRequest(obj, methodName);
                break;
            case "searchEntities":
                validateSearchEntityRequest(obj, methodName);
                break;
            case "getListOfSharedUsers":
            case "getListOfDirectlySharedUsers":
            case "getListOfSharedGroups":
            case "getListOfDirectlySharedGroups":
                validateGetSharingRequest(obj, methodName);
                break;

            case "shareEntityWithUsers":
            case "shareEntityWithGroups":
            case "revokeEntitySharingFromUsers":
            case "revokeEntitySharingFromGroups":
                validateSharingRequest(obj, methodName);
                break;

            case "userHasAccess":
                validateCheckAccessRequest(obj, methodName);
                break;
            default:
                throw new RuntimeException("Method not implemented");
        }

    }

    private boolean validateEntityTypeRequest(Object obj, String methodName) {

        return true;
    }

    private boolean validateEntityTypeIdentifyingRequest(Object obj, String methodName) {

        return true;
    }


    private boolean validateTenantIdOnly(Object obj, String methodName) {

        return true;
    }

    private boolean validateCreatePermissionTypeRequest(Object obj, String methodName) {

        return true;
    }

    private boolean validateEntityRequest(Object obj, String methodName) {

        return true;

    }

    private boolean validateEntityIdentifierRequest(Object object, String methodName) {

        return true;

    }

    private boolean validateSearchEntityRequest(Object object, String methodName) {
        
        return true;
    }


    private boolean validateGetSharingRequest(Object object, String methodName) {
        return true;
    }

    private boolean validateSharingRequest(Object object, String methodName) {
        return true;
    }

    private boolean validateCheckAccessRequest(Object object, String methodName) {
        return true;
    }


}
