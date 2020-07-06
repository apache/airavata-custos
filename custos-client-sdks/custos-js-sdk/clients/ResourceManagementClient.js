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

const {} = require('../stubs/integration-services/resource-secret-management/ResourceSecretManagementService_pb');
const {ResourceSecretManagementServiceClient} = require('../stubs/integration-services/resource-secret-management/ResourceSecretManagementService_grpc_web_pb');
const {GetSecretRequest, SecretMetadata, ResourceOwnerType, ResourceType} = require('../stubs/core-services/resource-secret-service/ResourceSecretService_pb');

var resourceSecretService = new ResourceSecretManagementServiceClient("https://custos.scigap.org/grpcweb", null, null);

var request = new GetSecretRequest();


var resourceOwnerType = ResourceOwnerType.CUSTOS;
var resourceType = ResourceType.SERVER_CERTIFICATE;


var secretMetadata = new SecretMetadata();
secretMetadata.setResourceType(resourceType);
secretMetadata.setOwnerType(resourceOwnerType);
request.setMetadata(secretMetadata);


const header = {'Authorization': 'Bearer XXX'}

resourceSecretService.getSecret(request, header, (err, response) => {

    if (err) {
        console.log(err)
    } else {
        console.log(response)
    }

})