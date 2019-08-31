/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */

/*
 * Component Programming Interface definition for Apache Custos Tenant profile Service.
 *
*/

include "../../authentication-service/authentication_service_cpi_errors.thrift"
include "../../authentication-service/security_model.thrift"
include "../data-models/user-tenant-group-models/tenant_profile_model.thrift"
include "../data-models/user-tenant-group-models/workspace_model.thrift"
include "profile_tenant_cpi_errors.thrift"

namespace java org.apache.custos.profile.tenant.cpi
namespace php Custos.Profile.Tenant.CPI
namespace py custos.profile.tenant.cpi

const string TENANT_PROFILE_CPI_VERSION = "0.17"
const string TENANT_PROFILE_CPI_NAME = "TenantProfileService"

service TenantProfileService {

    string getAPIVersion ()
                       throws (1: profile_tenant_cpi_errors.TenantProfileServiceException tpe)

    workspace_model.Gateway addGateway (1: required security_model.AuthzToken authzToken,
                       2: required workspace_model.Gateway gateway)
                    throws (1: profile_tenant_cpi_errors.TenantProfileServiceException tpe)

    workspace_model.Gateway updateGateway (1: required security_model.AuthzToken authzToken,
                        2: required workspace_model.Gateway updatedGateway)
                     throws (1: profile_tenant_cpi_errors.TenantProfileServiceException tpe)

    workspace_model.Gateway getGateway (1: required security_model.AuthzToken authzToken,
                                        2: required string custosInternalGatewayId)
                                     throws (1: profile_tenant_cpi_errors.TenantProfileServiceException tpe)

    workspace_model.Gateway getGatewayUsingGatewayId (1: required security_model.AuthzToken authzToken,
                                            2: required string gatewayId)
                                         throws (1: profile_tenant_cpi_errors.TenantProfileServiceException tpe)

    workspace_model.Gateway deleteGateway (1: required security_model.AuthzToken authzToken,
                        2: required string custosInternalGatewayId,
                        3: required string gatewayId)
                     throws (1: profile_tenant_cpi_errors.TenantProfileServiceException tpe)

    list<workspace_model.Gateway> getAllGateways (1: required security_model.AuthzToken authzToken)
                                               throws (1: profile_tenant_cpi_errors.TenantProfileServiceException tpe)

    bool isGatewayExist (1: required security_model.AuthzToken authzToken,
                         2: required string gatewayId)
                      throws (1: profile_tenant_cpi_errors.TenantProfileServiceException tpe)

    list<workspace_model.Gateway> getAllGatewaysForUser (1: required security_model.AuthzToken authzToken,
                                                         2: required string requesterUsername)
                                               throws (1: profile_tenant_cpi_errors.TenantProfileServiceException tpe)

    bool checkDuplicateGateway(1: required security_model.AuthzToken authzToken,
                               2: required workspace_model.Gateway gateway)
                       throws (1: profile_tenant_cpi_errors.TenantProfileServiceException tpe)
}