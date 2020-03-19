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
import org.apache.custos.scim.utils.Constants;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping(value = {"/v2/Groups"})
@Api(value = "Group Resource Management")
public class GroupResource {


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
                                   @RequestParam(Constants.ATTRIBUTES) String attribute,
                                   @ApiParam(value = Constants.EXCLUDED_ATTRIBUTES_DESC, required = false)
                                   @RequestParam(Constants.EXCLUDE_ATTRIBUTES) String excludedAttributes) {

        return null;

    }

    @ApiOperation(
            value = "Return the group which was created",
            notes = "Returns HTTP 201 if the group is successfully created.")
    @ApiResponses(value = {
            @ApiResponse(code = 201, message = "Valid group is created"),
            @ApiResponse(code = 404, message = "Group is not found")})

    @PostMapping(produces = {"application/json", "application/scim+json"}, consumes = {"application/scim+json"})
    public ResponseEntity createGroup(@ApiParam(value = Constants.ATTRIBUTES_DESC, required = false)
                                      @RequestParam(Constants.ATTRIBUTES) String attribute,
                                      @ApiParam(value = Constants.EXCLUDED_ATTRIBUTES_DESC, required = false)
                                      @RequestParam(Constants.EXCLUDE_ATTRIBUTES) String excludedAttributes,
                                      String resourceString) {
        return null;
    }

    @ApiOperation(
            value = "Delete the group with the given id",
            notes = "Returns HTTP 204 if the group is successfully deleted.")

    @ApiResponses(value = {
            @ApiResponse(code = 204, message = "Group is deleted"),
            @ApiResponse(code = 404, message = "Valid group is not found")})

    @DeleteMapping(value = {"/{id}"}, produces = {"application/json", "application/scim+json"})
    public ResponseEntity deleteGroup(@ApiParam(value = Constants.ID_DESC, required = true)
                                      @PathVariable(Constants.ID) String id) {

        return null;
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
                                      @RequestParam(Constants.ATTRIBUTES) String attribute,
                                      @ApiParam(value = Constants.EXCLUDED_ATTRIBUTES_DESC, required = false)
                                      @RequestParam(Constants.EXCLUDE_ATTRIBUTES) String excludedAttributes,
                                      String resourceString) {

        return null;
    }

    @ApiOperation(
            value = "Return groups according to the filter, sort and pagination parameters",
            notes = "Returns HTTP 404 if the groups are not found.")

    @ApiResponses(value = {
            @ApiResponse(code = 200, message = "Valid groups are found"),
            @ApiResponse(code = 404, message = "Valid groups are not found")})

    @PostMapping(value = ("/.search"), produces = {"application/json", "application/scim+json"}, consumes = {"application/scim+json"})
    public ResponseEntity getGroupsByPost(String resourceString) {

        return null;
    }

    @ApiOperation(
            value = "Return groups according to the filter, sort and pagination parameters",
            notes = "Returns HTTP 404 if the groups are not found.")

    @ApiResponses(value = {
            @ApiResponse(code = 200, message = "Valid groups are found"),
            @ApiResponse(code = 404, message = "Valid groups are not found")})

    @GetMapping(produces = {"application/json", "application/scim+json"})
    public ResponseEntity getGroup(@ApiParam(value = Constants.ATTRIBUTES_DESC, required = false)
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
                                   @RequestParam(value = Constants.DOMAIN) String domainName) {

        return null;

    }

}
