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
import org.apache.custos.core.services.commons.exceptions.MissingParameterException;
import org.apache.custos.sharing.service.*;
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
            case "getAllDirectSharings":
                validateGetAllDirectSharings(obj, methodName);
                break;
            default:
                throw new RuntimeException("Method not implemented");
        }

    }

    private boolean validateEntityTypeRequest(Object obj, String methodName) {
        if (obj instanceof EntityTypeRequest) {
            EntityTypeRequest entityTypeRequest = (EntityTypeRequest) obj;
            if (entityTypeRequest.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id not found ", null);
            }
            if (entityTypeRequest.getEntityType() == null) {
                throw new MissingParameterException("Entity type not found", null);
            }
            if (entityTypeRequest.getEntityType().getId() == null || entityTypeRequest.
                    getEntityType().getId().trim().equals("")) {
                throw new MissingParameterException("Entity type id not found", null);
            }
            if (entityTypeRequest.getEntityType().getName() == null || entityTypeRequest.
                    getEntityType().getName().trim().equals("")) {
                throw new MissingParameterException("Entity type name not found", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method " + methodName);
        }
        return true;
    }

    private boolean validateEntityTypeIdentifyingRequest(Object obj, String methodName) {
        if (obj instanceof EntityTypeRequest) {
            EntityTypeRequest entityTypeRequest = (EntityTypeRequest) obj;
            if (entityTypeRequest.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id not found ", null);
            }
            if (entityTypeRequest.getEntityType() == null) {
                throw new MissingParameterException("Entity type not found", null);
            }
            if (entityTypeRequest.getEntityType().getId() == null || entityTypeRequest.
                    getEntityType().getId().equals("")) {
                throw new MissingParameterException("Entity type id not found", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method " + methodName);
        }
        return true;
    }


    private boolean validateTenantIdOnly(Object obj, String methodName) {
        if (obj instanceof SearchRequest) {
            SearchRequest entityTypeRequest = (SearchRequest) obj;
            if (entityTypeRequest.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id not found ", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method " + methodName);
        }
        return true;
    }

    private boolean validateCreatePermissionTypeRequest(Object obj, String methodName) {
        if (obj instanceof PermissionTypeRequest) {
            PermissionTypeRequest permissionTypeRequest = (PermissionTypeRequest) obj;
            if (permissionTypeRequest.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id not found ", null);
            }
            if (permissionTypeRequest.getPermissionType() == null) {
                throw new MissingParameterException("Permission type not found", null);
            }
            if (permissionTypeRequest.getPermissionType().getId() == null || permissionTypeRequest.
                    getPermissionType().getId().equals("")) {
                throw new MissingParameterException("Permission type id not found", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method " + methodName);
        }
        return true;
    }

    private boolean validateEntityRequest(Object obj, String methodName) {

        if (obj instanceof EntityRequest) {
            EntityRequest entityRequest = (EntityRequest) obj;
            if (entityRequest.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id not found ", null);
            }

            if (entityRequest.getEntity() == null) {
                throw new MissingParameterException("Entity not found ", null);
            }

            if (entityRequest.getEntity().getId() == null || entityRequest.getEntity().getId().equals("")) {
                throw new MissingParameterException("Entity Id  not found", null);
            }
            if (entityRequest.getEntity().getType() == null || entityRequest.getEntity().getType().equals("")) {
                throw new MissingParameterException("Entity type  not found", null);
            }

            if (entityRequest.getEntity().getName() == null || entityRequest.getEntity().getName().equals("")) {
                throw new MissingParameterException("Entity name  not found", null);
            }

            if (entityRequest.getEntity().getOwnerId() == null || entityRequest.getEntity().getOwnerId().equals("")) {
                throw new MissingParameterException("Owner Id  not found", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method " + methodName);
        }
        return true;

    }

    private boolean validateEntityIdentifierRequest(Object object, String methodName) {

        if (object instanceof EntityRequest) {
            EntityRequest entityRequest = (EntityRequest) object;
            if (entityRequest.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id not found ", null);
            }

            if (entityRequest.getEntity() == null) {
                throw new MissingParameterException("Entity not found ", null);
            }

            if (entityRequest.getEntity().getId() == null || entityRequest.getEntity().getId().equals("")) {
                throw new MissingParameterException("Entity Id  not found", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method " + methodName);
        }
        return true;

    }

    private boolean validateSearchEntityRequest(Object object, String methodName) {

        if (object instanceof SearchRequest) {
            SearchRequest entityRequest = (SearchRequest) object;
            if (entityRequest.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id not found ", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method " + methodName);
        }
        return true;
    }


    private boolean validateGetSharingRequest(Object object, String methodName) {
        if (object instanceof SharingRequest) {
            SharingRequest entityRequest = (SharingRequest) object;
            if (entityRequest.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id not found ", null);
            }

            if (entityRequest.getEntity() == null) {
                throw new MissingParameterException("Entity  is not found ", null);
            }

            if (entityRequest.getEntity().getId() == null || entityRequest.getEntity().getId().equals("")) {
                throw new MissingParameterException("Entity id is  not found ", null);
            }

            if (entityRequest.getPermissionType() == null) {
                throw new MissingParameterException("Permission Type is  not found ", null);
            }

            if (entityRequest.getPermissionType().getId() == null || entityRequest.getPermissionType().getId().equals("")) {
                throw new MissingParameterException("Permission Type Id is  not found ", null);
            }


        } else {
            throw new RuntimeException("Unexpected input type for method " + methodName);
        }
        return true;
    }

    private boolean validateSharingRequest(Object object, String methodName) {
        if (object instanceof SharingRequest) {
            SharingRequest entityRequest = (SharingRequest) object;
            if (entityRequest.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id not found ", null);
            }

            if (entityRequest.getEntity() == null) {
                throw new MissingParameterException("Entity  is not found ", null);
            }

            if (entityRequest.getEntity().getId() == null || entityRequest.getEntity().getId().equals("")) {
                throw new MissingParameterException("Entity id is  not found ", null);
            }

            if (entityRequest.getPermissionType() == null) {
                throw new MissingParameterException("Permission Type is  not found ", null);
            }

            if (entityRequest.getPermissionType().getId() == null || entityRequest.getPermissionType().getId().equals("")) {
                throw new MissingParameterException("Permission Type Id is  not found ", null);
            }

            if (entityRequest.getOwnerIdList() == null || entityRequest.getOwnerIdList().isEmpty()) {
                throw new MissingParameterException("Owner  Id list   not found ", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method " + methodName);
        }
        return true;
    }

    private boolean validateCheckAccessRequest(Object object, String methodName) {
        if (object instanceof SharingRequest) {
            SharingRequest entityRequest = (SharingRequest) object;
            if (entityRequest.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id not found ", null);
            }

            if (entityRequest.getEntity() == null) {
                throw new MissingParameterException("Entity  is not found ", null);
            }

            if (entityRequest.getEntity().getId() == null || entityRequest.getEntity().getId().equals("")) {
                throw new MissingParameterException("Entity id is  not found ", null);
            }

            if (entityRequest.getPermissionType() == null) {
                throw new MissingParameterException("Permission Type is  not found ", null);
            }

            if (entityRequest.getPermissionType().getId() == null || entityRequest.getPermissionType().getId().equals("")) {
                throw new MissingParameterException("Permission Type Id is  not found ", null);
            }

            if (entityRequest.getOwnerIdList() == null || entityRequest.getOwnerIdList().isEmpty()) {
                throw new MissingParameterException("Owner  Id list   not found ", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method " + methodName);
        }
        return true;
    }

    private boolean validateGetAllDirectSharings(Object object, String methodName) {
        if (object instanceof SharingRequest) {
            SharingRequest entityRequest = (SharingRequest) object;
            if (entityRequest.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id not found ", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method " + methodName);
        }
        return true;
    }


}
