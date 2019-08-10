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
namespace java org.apache.custos.profile.model.workspace
namespace php Custos.Profile.Model.Workspace
namespace cpp apache.custos.profile.model.workspace
namespace py custos.profile.model.workspace

enum GatewayApprovalStatus {
    REQUESTED,
    APPROVED,
    ACTIVE,
    DEACTIVATED,
    CANCELLED,
    DENIED,
    CREATED,
    DEPLOYED
}

struct Gateway {
    1: optional string custosInternalGatewayId,
    2: required string gatewayId,
    3: required GatewayApprovalStatus gatewayApprovalStatus,
    4: optional string gatewayName,
    5: optional string domain,
    6: optional string emailAddress
    7: optional string gatewayAcronym,
    8: optional string gatewayURL,
    9: optional string gatewayPublicAbstract,
    10: optional string reviewProposalDescription,
    11: optional string gatewayAdminFirstName,
    12: optional string gatewayAdminLastName,
    13: optional string gatewayAdminEmail,
    14: optional string identityServerUserName,
    15: optional string identityServerPasswordToken,
    16: optional string declinedReason,
    17: optional string oauthClientId,
    18: optional string oauthClientSecret,
    19: optional i64 requestCreationTime,
    20: optional string requesterUsername
}