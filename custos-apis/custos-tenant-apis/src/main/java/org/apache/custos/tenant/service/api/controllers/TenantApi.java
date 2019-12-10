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

import org.apache.custos.tenant.service.model.Tenant;
import org.apache.custos.tenant.service.model.TenantApprovalStatus;
import org.apache.custos.tenant.service.model.TenantCredentials;
import org.springframework.http.ResponseEntity;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import javax.validation.Valid;
import javax.validation.constraints.*;
import java.util.List;

@RestController

public interface TenantApi {


    @RequestMapping(value = "/tenant/{tenantId}/status",
        produces = { "application/json" }, 
        method = RequestMethod.GET)
    ResponseEntity<TenantApprovalStatus> checkApporvalStatus(@PathVariable("tenantId") Long tenantId);


    @RequestMapping( value = "/tenant",
        produces = { "application/json" }, 
        consumes = { "application/json" },
        method = RequestMethod.POST)
    ResponseEntity<Tenant> createTenant(@Valid @RequestBody Tenant body);



    @RequestMapping(value = "/tenant/{tenantId}",
        produces = { "application/json" }, 
        method = RequestMethod.DELETE)
    ResponseEntity<Void> deleteTenant(@PathVariable("tenantId") Long tenantId);


    @RequestMapping(value = "/tenant/{tenantId}/credentials",
        produces = { "application/json" }, 
        method = RequestMethod.GET)
    ResponseEntity<TenantCredentials> getCredentials(@PathVariable("tenantId") Long tenantId);


    @RequestMapping(value = "/tenant/{tenantId}",
        produces = { "application/json" }, 
        method = RequestMethod.GET)
    ResponseEntity<Tenant> getTenantById(@PathVariable("tenantId") Long tenantId);


    @RequestMapping(value = "/tenant",
        produces = { "application/json" }, 
        consumes = { "application/json" },
        method = RequestMethod.PATCH)
    ResponseEntity<Void> updateTenant(@Valid @RequestBody Tenant body);

}
