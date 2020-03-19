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
import org.apache.custos.scim.resource.manager.ResourceManager;
import org.apache.custos.scim.utils.Constants;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.wso2.charon3.core.protocol.SCIMResponse;
import org.wso2.charon3.core.protocol.endpoints.UserResourceManager;

import java.net.URI;
import java.util.Map;


@RestController
@RequestMapping(value = {"/v2/Users"})
@Api(value = "User Resource Management")
public class UserResource {

    private final static Logger LOGGER = LoggerFactory.getLogger(UserResource.class);


    @Autowired
    private ResourceManager resourceManager;


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
                                  @RequestParam(Constants.ATTRIBUTES) String attribute,
                                  @ApiParam(value = Constants.EXCLUDED_ATTRIBUTES_DESC, required = false)
                                  @RequestParam(Constants.EXCLUDE_ATTRIBUTES) String excludedAttributes) {

        return null;
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
                                     String resourceString) {

        LOGGER.info("Request Received "+ resourceString);

        // create charon-SCIM user endpoint and hand-over the request.
        UserResourceManager userResourceManager = new UserResourceManager();

        SCIMResponse response = userResourceManager.create(resourceString, resourceManager,
                attribute, excludedAttributes);


        LOGGER.info("Status" + response.getResponseStatus());
        LOGGER.info("Message" + response.getResponseMessage());
        Map<String, String> headerMap = response.getHeaderParamMap();

        for (String key : headerMap.keySet()) {
            LOGGER.info("Key: " + key);
            LOGGER.info("Value: " + headerMap.get(key));
        }

        ResponseEntity responseEntity = ResponseEntity.created(URI.create("http://lcoalhost:8080")).build();
        return responseEntity;
    }

    @ApiOperation(
            value = "Delete the user with the given id",
            notes = "Returns HTTP 204 if the user is successfully deleted.")

    @ApiResponses(value = {
            @ApiResponse(code = 204, message = "User is deleted"),
            @ApiResponse(code = 404, message = "Valid user is not found")})

    @DeleteMapping(path = {"/{id}"}, produces = {"application/json", "application/scim+json"})
    public ResponseEntity deleteUser(@ApiParam(value = Constants.ID_DESC, required = true)
                                     @PathVariable(Constants.ID) String id) {
        return null;

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
                                  @RequestParam(Constants.EXCLUDE_ATTRIBUTES) String excludedAttributes,
                                  @ApiParam(value = Constants.FILTER_DESC, required = false)
                                  @RequestParam(Constants.FILTER) String filter,
                                  @ApiParam(value = Constants.START_INDEX_DESC, required = false)
                                  @RequestParam(Constants.START_INDEX) int startIndex,
                                  @ApiParam(value = Constants.COUNT_DESC, required = false)
                                  @RequestParam(Constants.COUNT) int count,
                                  @ApiParam(value = Constants.SORT_BY_DESC, required = false)
                                  @RequestParam(Constants.SORT_BY) String sortBy,
                                  @ApiParam(value = Constants.SORT_ORDER_DESC, required = false)
                                  @RequestParam(Constants.SORT_ORDER) String sortOrder,
                                  @ApiParam(value = Constants.DOMAIN_DESC, required = false)
                                  @RequestParam(Constants.DOMAIN) String domainName) {
        return null;

    }


    @ApiOperation(
            value = "Return users according to the filter, sort and pagination parameters",
            notes = "Returns HTTP 404 if the users are not found.")

    @ApiResponses(value = {
            @ApiResponse(code = 200, message = "Valid users are found"),
            @ApiResponse(code = 404, message = "Valid users are not found")})

    @PostMapping(value = {"/.search"}, produces = {"application/json", "application/scim+json"}, consumes = {"application/scim+json"})
    public ResponseEntity getUsersByPost(String resourceString) {

        return null;
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
                                     @RequestParam(Constants.ATTRIBUTES) String attribute,
                                     @ApiParam(value = Constants.EXCLUDED_ATTRIBUTES_DESC, required = false)
                                     @RequestParam(Constants.EXCLUDE_ATTRIBUTES) String excludedAttributes,
                                     String resourceString) {

        return null;
    }

}
