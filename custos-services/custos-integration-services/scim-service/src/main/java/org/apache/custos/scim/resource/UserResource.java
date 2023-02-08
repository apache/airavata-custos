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

package org.apache.custos.scim.resource;

import io.swagger.annotations.*;
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.scim.resource.manager.ResourceManager;
import org.apache.custos.scim.utils.AuthHandler;
import org.apache.custos.scim.utils.Constants;
import org.json.JSONObject;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.wso2.charon3.core.protocol.SCIMResponse;
import org.wso2.charon3.core.protocol.endpoints.UserResourceManager;

import java.util.Map;
import java.util.Optional;


@RestController
@RequestMapping(value = {"/v2/Users"})
@Api(value = "User Resource Management")
public class UserResource extends AbstractResource {

    private final static Logger LOGGER = LoggerFactory.getLogger(UserResource.class);


    @Autowired
    private ResourceManager resourceManager;

    @Autowired
    private AuthHandler authHandler;


    @ApiOperation(
            value = "Return the user with the given id",
            notes = "Returns HTTP 200 if the user is found.")

    @ApiResponses(value = {
            @ApiResponse(code = 200, message = "Valid user is found"),
            @ApiResponse(code = 404, message = "Valid user is not found")})

    @GetMapping(value = {"/{id}"}, produces = {"application/json", "application/scim+json"})
    public ResponseEntity getUser(@ApiParam(value = Constants.ID_DESC, required = true)
                                  @PathVariable(Constants.ID) String id,
                                  @ApiParam(value = Constants.ATTRIBUTES_DESC, required = false)
                                  @RequestParam(value = Constants.ATTRIBUTES, required = false) String attribute,
                                  @ApiParam(value = Constants.EXCLUDED_ATTRIBUTES_DESC, required = false)
                                  @RequestParam(value = Constants.EXCLUDE_ATTRIBUTES, required = false) String excludedAttributes,
                                  @RequestHeader(value = Constants.AUTHORIZATION) String authorizationHeader) {

        Optional<AuthClaim> claim = authHandler.validateAndConfigure(authorizationHeader, false);

        JSONObject newObj = new JSONObject();
        JSONObject custosExtention = new JSONObject();
        if (claim.isPresent()) {
            newObj.put(Constants.CLIENT_ID, claim.get().getIamAuthId());
            newObj.put(Constants.CLIENT_SEC, claim.get().getIamAuthSecret());
            newObj.put(Constants.ID, id);
            newObj.put(Constants.TENANT_ID, String.valueOf(claim.get().getTenantId()));
            newObj.put(Constants.ACCESS_TOKEN, authHandler.getToken(authorizationHeader));


        }
        custosExtention.put(Constants.CUSTOS_EXTENSION, newObj);

        // create charon-SCIM user endpoint and hand-over the request.
        UserResourceManager userResourceManager = new UserResourceManager();
        LOGGER.debug("Id Before  " + custosExtention.toString());
        SCIMResponse response = userResourceManager.get(custosExtention.toString(), resourceManager, attribute, excludedAttributes);

        return buildResponse(response);
    }


    @ApiOperation(
            value = "Return the user which was created",
            notes = "Returns HTTP 201 if the user is successfully created.")

    @ApiResponses(value = {
            @ApiResponse(code = 201, message = "Valid user is created"),
            @ApiResponse(code = 404, message = "User is not found")})

    @PostMapping(produces = {"application/json", "application/scim+json"}, consumes = {"application/scim+json"})
    public ResponseEntity createUser(@ApiParam(value = Constants.ATTRIBUTES_DESC, required = false)
                                     @RequestParam(value = Constants.ATTRIBUTES, required = false) String attribute,
                                     @ApiParam(value = Constants.EXCLUDED_ATTRIBUTES_DESC, required = false)
                                     @RequestParam(value = Constants.EXCLUDE_ATTRIBUTES, required = false) String excludedAttributes,
                                     @RequestBody Map<String, Object> payload,
                                     @RequestHeader(value = Constants.AUTHORIZATION) String authorizationHeader) {

        Optional<AuthClaim> claim = authHandler.validateAndConfigure(authorizationHeader, false);

        JSONObject object = new JSONObject(payload);

        JSONObject custosExtension = new JSONObject();

        if (claim.isPresent()) {
            custosExtension.put(Constants.CLIENT_ID, claim.get().getIamAuthId());
            custosExtension.put(Constants.CLIENT_SEC, claim.get().getIamAuthSecret());
            custosExtension.put(Constants.TENANT_ID, claim.get().getTenantId());
            custosExtension.put(Constants.ACCESS_TOKEN, authHandler.getToken(authorizationHeader));

            object.put(Constants.CUSTOS_EXTENSION, custosExtension);
        }
        // create charon-SCIM user endpoint and hand-over the request.
        UserResourceManager userResourceManager = new UserResourceManager();

        LOGGER.debug(object.toString());

        SCIMResponse response = userResourceManager.create(object.toString(), resourceManager,
                attribute, excludedAttributes);


        return buildResponse(response);
    }

    @ApiOperation(
            value = "Delete the user with the given id",
            notes = "Returns HTTP 204 if the user is successfully deleted.")

    @ApiResponses(value = {
            @ApiResponse(code = 204, message = "User is deleted"),
            @ApiResponse(code = 404, message = "Valid user is not found")})

    @DeleteMapping(path = {"/{id}"}, produces = {"application/json", "application/scim+json"})
    public ResponseEntity deleteUser(@ApiParam(value = Constants.ID_DESC, required = true)
                                     @PathVariable(Constants.ID) String id,
                                     @RequestHeader(value = Constants.AUTHORIZATION) String authorizationHeader) {

        Optional<AuthClaim> claim = authHandler.validateAndConfigure(authorizationHeader, false);


        JSONObject newObj = new JSONObject();
        JSONObject custosExtention = new JSONObject();
        if (claim.isPresent()) {
            newObj.put(Constants.CLIENT_ID, claim.get().getIamAuthId());
            newObj.put(Constants.CLIENT_SEC, claim.get().getIamAuthSecret());
            newObj.put(Constants.ID, id);
            newObj.put(Constants.TENANT_ID, String.valueOf(claim.get().getTenantId()));
            newObj.put(Constants.ACCESS_TOKEN, authHandler.getToken(authorizationHeader));
        }

        custosExtention.put(Constants.CUSTOS_EXTENSION, newObj);

        // create charon-SCIM user endpoint and hand-over the request.
        UserResourceManager userResourceManager = new UserResourceManager();
        LOGGER.debug("Id Before  " + custosExtention.toString());
        SCIMResponse response = userResourceManager.delete(custosExtention.toString(), resourceManager);

        return buildResponse(response);

    }


    @ApiOperation(
            value = "Return users according to the filter, sort and pagination parameters",
            notes = "Returns HTTP 404 if the users are not found.")

    @ApiResponses(value = {
            @ApiResponse(code = 200, message = "Valid users are found"),
            @ApiResponse(code = 404, message = "Valid users are not found")})

    @GetMapping(produces = {"application/json", "application/scim+json"})
    public ResponseEntity getUser(@ApiParam(value = Constants.ATTRIBUTES_DESC, required = false)
                                  @RequestParam(Constants.ATTRIBUTES) String attribute,
                                  @ApiParam(value = Constants.EXCLUDED_ATTRIBUTES_DESC, required = false)
                                  @RequestParam(value = Constants.EXCLUDE_ATTRIBUTES, required = false) String excludedAttributes,
                                  @ApiParam(value = Constants.FILTER_DESC, required = false)
                                  @RequestParam(value = Constants.FILTER, required = false) String filter,
                                  @ApiParam(value = Constants.START_INDEX_DESC, required = false)
                                  @RequestParam(value = Constants.START_INDEX, required = false) int startIndex,
                                  @ApiParam(value = Constants.COUNT_DESC, required = false)
                                  @RequestParam(value = Constants.COUNT, required = false) int count,
                                  @ApiParam(value = Constants.SORT_BY_DESC, required = false)
                                  @RequestParam(value = Constants.SORT_BY, required = false) String sortBy,
                                  @ApiParam(value = Constants.SORT_ORDER_DESC, required = false)
                                  @RequestParam(value = Constants.SORT_ORDER, required = false) String sortOrder,
                                  @ApiParam(value = Constants.DOMAIN_DESC, required = false)
                                  @RequestParam(value = Constants.DOMAIN, required = false) String domainName,
                                  @RequestHeader(value = Constants.AUTHORIZATION) String authorizationHeader) {
        authHandler.validateAndConfigure(authorizationHeader, false);

        UserResourceManager userResourceManager = new UserResourceManager();

        SCIMResponse response = userResourceManager.listWithGET(resourceManager,
                filter, startIndex, count, sortBy, sortOrder, domainName, attribute, excludedAttributes);

        return buildResponse(response);

    }


    @ApiOperation(
            value = "Return users according to the filter, sort and pagination parameters",
            notes = "Returns HTTP 404 if the users are not found.")

    @ApiResponses(value = {
            @ApiResponse(code = 200, message = "Valid users are found"),
            @ApiResponse(code = 404, message = "Valid users are not found")})

    @PostMapping(value = {"/.search"}, produces = {"application/json", "application/scim+json"}, consumes = {"application/scim+json"})
    public ResponseEntity getUsersByPost(@RequestBody  String resourceString, @RequestHeader(value = Constants.AUTHORIZATION) String authorizationHeader) {

        Optional<AuthClaim> claim = authHandler.validateAndConfigure(authorizationHeader, false);


        JSONObject newObj = new JSONObject();
        JSONObject custosExtention = new JSONObject(resourceString);
        if (claim.isPresent()) {
            newObj.put(Constants.CLIENT_ID, claim.get().getIamAuthId());
            newObj.put(Constants.CLIENT_SEC, claim.get().getIamAuthSecret());
            newObj.put(Constants.TENANT_ID, String.valueOf(claim.get().getTenantId()));
            newObj.put(Constants.ACCESS_TOKEN, authHandler.getToken(authorizationHeader));
        }

        custosExtention.put(Constants.CUSTOS_EXTENSION, newObj);

        UserResourceManager userResourceManager = new UserResourceManager();

        SCIMResponse response = userResourceManager.listWithPOST(custosExtention.toString(), resourceManager);

        return buildResponse(response);

    }


    @ApiOperation(
            value = "Return the updated user",
            notes = "Returns HTTP 404 if the user is not found.")

    @ApiResponses(value = {
            @ApiResponse(code = 200, message = "User is updated"),
            @ApiResponse(code = 404, message = "Valid user is not found")})

    @PutMapping(path = "/{id}", produces = {"application/json", "application/scim+json"}, consumes = {"application/scim+json"})
    public ResponseEntity updateUser(@ApiParam(value = Constants.ID_DESC, required = true)
                                     @PathVariable(Constants.ID) String id,
                                     @ApiParam(value = Constants.ATTRIBUTES_DESC, required = false)
                                     @RequestParam(value = Constants.ATTRIBUTES, required = false) String attribute,
                                     @ApiParam(value = Constants.EXCLUDED_ATTRIBUTES_DESC, required = false)
                                     @RequestParam(value = Constants.EXCLUDE_ATTRIBUTES, required = false) String excludedAttributes,
                                     @RequestBody Map<String, Object> payload,
                                     @RequestHeader(value = Constants.AUTHORIZATION) String authorizationHeader) {

        Optional<AuthClaim> claim = authHandler.validateAndConfigure(authorizationHeader, false);

        JSONObject object = new JSONObject(payload);

        JSONObject custosExtension = new JSONObject();
        JSONObject newObj = new JSONObject();

        if (claim.isPresent()) {
            custosExtension.put(Constants.CLIENT_ID, claim.get().getIamAuthId());
            custosExtension.put(Constants.CLIENT_SEC, claim.get().getIamAuthSecret());
            custosExtension.put(Constants.ACCESS_TOKEN, authHandler.getToken(authorizationHeader));
            custosExtension.put(Constants.TENANT_ID, claim.get().getTenantId());
            object.put(Constants.CUSTOS_EXTENSION, custosExtension);


            newObj.put(Constants.CLIENT_ID, claim.get().getIamAuthId());
            newObj.put(Constants.CLIENT_SEC, claim.get().getIamAuthSecret());
            newObj.put(Constants.ID, id);
            newObj.put(Constants.TENANT_ID, String.valueOf(claim.get().getTenantId()));
            newObj.put(Constants.ACCESS_TOKEN, authHandler.getToken(authorizationHeader));

        }
        JSONObject custosExt = new JSONObject();
        custosExt.put(Constants.CUSTOS_EXTENSION, newObj);


        // create charon-SCIM user endpoint and hand-over the request.
        UserResourceManager userResourceManager = new UserResourceManager();


        SCIMResponse response = userResourceManager.updateWithPUT(custosExt.toString(), object.toString(), resourceManager,
                attribute, excludedAttributes);


        return buildResponse(response);
    }


}
