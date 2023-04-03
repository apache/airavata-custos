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
import org.json.JSONArray;
import org.json.JSONObject;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.wso2.charon3.core.protocol.SCIMResponse;
import org.wso2.charon3.core.protocol.endpoints.GroupResourceManager;
import org.wso2.charon3.core.protocol.endpoints.UserResourceManager;

import java.util.Map;
import java.util.Optional;

@RestController
@RequestMapping(value = {"/v2/Groups"})
@Api(value = "Group Resource Management")
public class GroupResource extends AbstractResource {
    private final static Logger LOGGER = LoggerFactory.getLogger(GroupResource.class);

    @Autowired
    private ResourceManager resourceManager;

    @Autowired
    private AuthHandler authHandler;

    @ApiOperation(
            value = "Return the group with the given id",
            notes = "Returns HTTP 200 if the group is found.")

    @ApiResponses(value = {
            @ApiResponse(code = 200, message = "Valid group is found"),
            @ApiResponse(code = 404, message = "Valid group is not found")})

    @GetMapping(value = {"/{id}"}, produces = {"application/json", "application/scim+json"})
    public ResponseEntity getGroup(@ApiParam(value = Constants.ID_DESC, required = true)
                                   @PathVariable(Constants.ID) String id,
                                   @ApiParam(value = Constants.ATTRIBUTES_DESC, required = false)
                                   @RequestParam(value = Constants.ATTRIBUTES, required = false) String attribute,
                                   @ApiParam(value = Constants.EXCLUDED_ATTRIBUTES_DESC, required = false)
                                   @RequestParam(value = Constants.EXCLUDE_ATTRIBUTES, required = false) String excludedAttributes,
                                   @RequestHeader(value = Constants.AUTHORIZATION) String authorizationHeader) {

        Optional<AuthClaim> claim = authHandler.validateAndConfigure(authorizationHeader, false);

        JSONObject newObj = new JSONObject();

        if (claim.isPresent()) {
            newObj.put(Constants.CLIENT_ID, claim.get().getIamAuthId());
            newObj.put(Constants.CLIENT_SEC, claim.get().getIamAuthSecret());
            newObj.put(Constants.ID, id);
            newObj.put(Constants.TENANT_ID, String.valueOf(claim.get().getTenantId()));
            newObj.put(Constants.ACCESS_TOKEN, authHandler.getToken(authorizationHeader));
        }

        JSONObject custosExtention = new JSONObject();
        custosExtention.put(Constants.CUSTOS_EXTENSION, newObj);

        // create charon-SCIM user endpoint and hand-over the request.
        GroupResourceManager groupResourceManager = new GroupResourceManager();

        SCIMResponse response = groupResourceManager.get(custosExtention.toString(), resourceManager, attribute, excludedAttributes);

        return buildResponse(response);

    }

    @ApiOperation(
            value = "Return the group which was created",
            notes = "Returns HTTP 201 if the group is successfully created.")
    @ApiResponses(value = {
            @ApiResponse(code = 201, message = "Valid group is created"),
            @ApiResponse(code = 404, message = "Group is not found")})

    @PostMapping(produces = {"application/json", "application/scim+json"}, consumes = {"application/scim+json"})
    public ResponseEntity createGroup(@ApiParam(value = Constants.ATTRIBUTES_DESC, required = false)
                                      @RequestParam(value = Constants.ATTRIBUTES, required = false) String attribute,
                                      @ApiParam(value = Constants.EXCLUDED_ATTRIBUTES_DESC, required = false)
                                      @RequestParam(value = Constants.EXCLUDE_ATTRIBUTES, required = false) String excludedAttributes,
                                      @RequestBody Map<String, Object> payload,
                                      @RequestHeader(value = Constants.AUTHORIZATION) String authorizationHeader) {
        Optional<AuthClaim> claim = authHandler.validateAndConfigure(authorizationHeader, false);

        JSONObject object = new JSONObject(payload);

//        Object custosExtension = object.get(Constants.MEMBERS);
//
//        JSONArray cust = null;
//        if (custosExtension == null) {
        JSONArray cust = new JSONArray();
//        } else if (custosExtension instanceof JSONArray) {
//            cust = (JSONArray) custosExtension;
//        }

        JSONObject jsonObject = new JSONObject();

        if (claim.isPresent()) {
            jsonObject.put(Constants.CLIENT_ID, claim.get().getIamAuthId());
            jsonObject.put(Constants.CLIENT_SEC, claim.get().getIamAuthSecret());
            jsonObject.put(Constants.ACCESS_TOKEN, authHandler.getToken(authorizationHeader));
            jsonObject.put(Constants.TENANT_ID, String.valueOf(claim.get().getTenantId()));
        }


        // create charon-SCIM user endpoint and hand-over the request.
        GroupResourceManager groupResourceManager = new GroupResourceManager();

        JSONObject member = new JSONObject();

        member.put(Constants.VALUE, jsonObject.toString());
        member.put("display", Constants.CUSTOS_EXTENSION);

        cust.put(member);

        object.put(Constants.MEMBERS, cust);


        SCIMResponse response = groupResourceManager.create(object.toString(), resourceManager,
                attribute, excludedAttributes);


        return buildResponse(response);
    }

    @ApiOperation(
            value = "Delete the group with the given id",
            notes = "Returns HTTP 204 if the group is successfully deleted.")

    @ApiResponses(value = {
            @ApiResponse(code = 204, message = "Group is deleted"),
            @ApiResponse(code = 404, message = "Valid group is not found")})

    @DeleteMapping(value = {"/{id}"}, produces = {"application/json", "application/scim+json"})
    public ResponseEntity deleteGroup(@ApiParam(value = Constants.ID_DESC, required = true)
                                      @PathVariable(Constants.ID) String id,
                                      @RequestHeader(value = Constants.AUTHORIZATION) String authorizationHeader) {

        Optional<AuthClaim> claim = authHandler.validateAndConfigure(authorizationHeader, false);


        JSONObject newObj = new JSONObject();

        if (claim.isPresent()) {
            newObj.put(Constants.CLIENT_ID, claim.get().getIamAuthId());
            newObj.put(Constants.CLIENT_SEC, claim.get().getIamAuthSecret());
            newObj.put(Constants.ID, id);
            newObj.put(Constants.TENANT_ID, String.valueOf(claim.get().getTenantId()));
            newObj.put(Constants.ACCESS_TOKEN, authHandler.getToken(authorizationHeader));
        }


        JSONObject custosExtention = new JSONObject();
        custosExtention.put(Constants.CUSTOS_EXTENSION, newObj);

        // create charon-SCIM user endpoint and hand-over the request.
        GroupResourceManager groupResourceManager = new GroupResourceManager();
        LOGGER.info("Id Before  " + custosExtention.toString());
        SCIMResponse response = groupResourceManager.delete(custosExtention.toString(), resourceManager);

        return buildResponse(response);
    }

    @ApiOperation(
            value = "Return the updated group",
            notes = "Returns HTTP 404 if the group is not found.")

    @ApiResponses(value = {
            @ApiResponse(code = 200, message = "Group is updated"),
            @ApiResponse(code = 404, message = "Valid group is not found")})

    @PutMapping(value = {"/{id}"}, produces = {"application/json", "application/scim+json"}, consumes = {"application/scim+json"})
    public ResponseEntity updateGroup(@ApiParam(value = Constants.ID_DESC, required = true)
                                      @PathVariable(Constants.ID) String id,
                                      @ApiParam(value = Constants.ATTRIBUTES_DESC, required = false)
                                      @RequestParam(value = Constants.ATTRIBUTES, required = false) String attribute,
                                      @ApiParam(value = Constants.EXCLUDED_ATTRIBUTES_DESC, required = false)
                                      @RequestParam(value = Constants.EXCLUDE_ATTRIBUTES, required = false) String excludedAttributes,
                                      @RequestBody Map<String, Object> payload,
                                      @RequestHeader(value = Constants.AUTHORIZATION) String authorizationHeader) {

        Optional<AuthClaim> claim = authHandler.validateAndConfigure(authorizationHeader, false);

        JSONObject object = new JSONObject(payload);


//        Object custosExtension = object.get(Constants.MEMBERS);
//
//        JSONArray cust = null;
//        if (custosExtension == null) {
        JSONArray cust = new JSONArray();
//        } else if (custosExtension instanceof JSONArray) {
//            cust = (JSONArray) custosExtension;
//        }

        JSONObject jsonObject = new JSONObject();
        JSONObject newObj = new JSONObject();
        if (claim.isPresent()) {
            jsonObject.put(Constants.CLIENT_ID, claim.get().getIamAuthId());
            jsonObject.put(Constants.CLIENT_SEC, claim.get().getIamAuthSecret());
            jsonObject.put(Constants.ACCESS_TOKEN, authHandler.getToken(authorizationHeader));
            jsonObject.put(Constants.TENANT_ID, String.valueOf(claim.get().getTenantId()));


            JSONObject member = new JSONObject();

            member.put(Constants.VALUE, jsonObject.toString());
            member.put("display", Constants.CUSTOS_EXTENSION);

            cust.put(member);

            object.put(Constants.MEMBERS, cust);


            newObj.put(Constants.CLIENT_ID, claim.get().getIamAuthId());
            newObj.put(Constants.CLIENT_SEC, claim.get().getIamAuthSecret());
            newObj.put(Constants.ID, id);
            newObj.put(Constants.TENANT_ID, String.valueOf(claim.get().getTenantId()));
            newObj.put(Constants.ACCESS_TOKEN, authHandler.getToken(authorizationHeader));

        }
        JSONObject custosExt = new JSONObject();

        custosExt.put(Constants.CUSTOS_EXTENSION, newObj);


        // create charon-SCIM user endpoint and hand-over the request.
        GroupResourceManager groupResourceManager = new GroupResourceManager();


        SCIMResponse response = groupResourceManager.updateWithPUT(custosExt.toString(), object.toString(), resourceManager,
                attribute, excludedAttributes);


        return buildResponse(response);

    }

    @ApiOperation(
            value = "Return groups according to the filter, sort and pagination parameters",
            notes = "Returns HTTP 404 if the groups are not found.")

    @ApiResponses(value = {
            @ApiResponse(code = 200, message = "Valid groups are found"),
            @ApiResponse(code = 404, message = "Valid groups are not found")})

    @PostMapping(value = ("/.search"), produces = {"application/json", "application/scim+json"}, consumes = {"application/scim+json"})
    public ResponseEntity getGroupsByPost(@RequestBody  String resourceString, @RequestHeader(value = Constants.AUTHORIZATION) String authorizationHeader) {

        Optional<AuthClaim> claim = authHandler.validateAndConfigure(authorizationHeader, false);


        JSONObject newObj = new JSONObject();
        JSONObject custosExtention = new JSONObject(resourceString);
        if (claim.isPresent()) {
            newObj.put(Constants.CLIENT_ID, claim.get().getIamAuthId());
            newObj.put(Constants.CLIENT_SEC, claim.get().getIamAuthSecret());
            newObj.put(Constants.TENANT_ID, String.valueOf(claim.get().getTenantId()));
            newObj.put(Constants.ACCESS_TOKEN, authHandler.getToken(authorizationHeader));
        }

        custosExtention.put(Constants.DOMAIN, newObj.toString());

        GroupResourceManager groupResourceManager = new GroupResourceManager();

        SCIMResponse response = groupResourceManager.listWithPOST(custosExtention.toString(), resourceManager);

        return buildResponse(response);
    }

    @ApiOperation(
            value = "Return groups according to the filter, sort and pagination parameters",
            notes = "Returns HTTP 404 if the groups are not found.")

    @ApiResponses(value = {
            @ApiResponse(code = 200, message = "Valid groups are found"),
            @ApiResponse(code = 404, message = "Valid groups are not found")})

    @GetMapping(produces = {"application/json", "application/scim+json"})
    public ResponseEntity getGroup(@ApiParam(value = Constants.ATTRIBUTES_DESC, required = false)
                                   @RequestParam(value = Constants.ATTRIBUTES, required = false) String attribute,
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

        GroupResourceManager groupResourceManager = new GroupResourceManager();

        SCIMResponse response = groupResourceManager.listWithGET(resourceManager,
                filter, startIndex, count, sortBy, sortOrder, domainName, attribute, excludedAttributes);

        return buildResponse(response);

    }

}
