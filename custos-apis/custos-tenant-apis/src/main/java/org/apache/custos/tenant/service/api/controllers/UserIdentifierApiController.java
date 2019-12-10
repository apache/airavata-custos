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
package org.apache.custos.tenant.service.api.controllers;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.custos.tenant.service.model.Tenant;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestHeader;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RequestPart;
import org.springframework.web.multipart.MultipartFile;

import javax.validation.constraints.*;
import javax.validation.Valid;
import javax.servlet.http.HttpServletRequest;
import java.io.IOException;
import java.util.List;

@Controller
public class UserIdentifierApiController implements UserIdentifierApi {

    private static final Logger log = LoggerFactory.getLogger(UserIdentifierApiController.class);

    private final ObjectMapper objectMapper;

    private final HttpServletRequest request;

    @org.springframework.beans.factory.annotation.Autowired
    public UserIdentifierApiController(ObjectMapper objectMapper, HttpServletRequest request) {
        this.objectMapper = objectMapper;
        this.request = request;
    }

    public ResponseEntity<List<Tenant>> getTenantsForUser(@PathVariable("userIdentifier") String userIdentifier) {
        String accept = request.getHeader("Accept");
        if (accept != null && accept.contains("application/json")) {
            try {
                return new ResponseEntity<List<Tenant>>(objectMapper.readValue("[ {  \"tenantAdminLastName\" : \"tenantAdminLastName\",  \"requesterEmail\" : \"xyz@iu.edu\",  \"tenantURI\" : \"galaxy.com\",  \"tenantName\" : \"Galaxy\",  \"scope\" : \"scope\",  \"domain\" : \"domain\",  \"tenantId\" : \"galaxy\",  \"tenantAdminFirstName\" : \"tenantAdminFirstName\",  \"logoURI\" : \"logoURI\",  \"redirectURIs\" : [ \"galaxy/oauth\", \"galaxy/oauth\" ],  \"contacts\" : [ \"contacts\", \"contacts\" ]}, {  \"tenantAdminLastName\" : \"tenantAdminLastName\",  \"requesterEmail\" : \"xyz@iu.edu\",  \"tenantURI\" : \"galaxy.com\",  \"tenantName\" : \"Galaxy\",  \"scope\" : \"scope\",  \"domain\" : \"domain\",  \"tenantId\" : \"galaxy\",  \"tenantAdminFirstName\" : \"tenantAdminFirstName\",  \"logoURI\" : \"logoURI\",  \"redirectURIs\" : [ \"galaxy/oauth\", \"galaxy/oauth\" ],  \"contacts\" : [ \"contacts\", \"contacts\" ]} ]", List.class), HttpStatus.NOT_IMPLEMENTED);
            } catch (IOException e) {
                log.error("Couldn't serialize response for content type application/json", e);
                return new ResponseEntity<List<Tenant>>(HttpStatus.INTERNAL_SERVER_ERROR);
            }
        }

        return new ResponseEntity<List<Tenant>>(HttpStatus.NOT_IMPLEMENTED);
    }

}
