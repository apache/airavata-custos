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

package org.apache.custos.service.iam;

/**
 * Includes operations associated with keycloak
 */
public enum IAMOperations {

    SET_UP_TENANT,
    UPDATE_TENANT,
    REGISTER_USER,
    ENABLE_USER,
    DISABLE_USER,
    DELETE_USER,
    RESET_PASSWORD,
    UPDATE_USER_PROFILE,
    ADD_ROLE_TO_USER,
    DELETE_ROLE_FROM_USER,
    CONFIGURE_IDP,
    REGISTER_ENABLE_USERS,
    ADD_ROLES_TO_TENANT,
    ADD_PROTOCOL_MAPPER,
    ADD_USER_ATTRIBUTE,
    ADD_ROLES_TO_USERS,
    DELETE_USER_ATTRIBUTES,
    CONFIGURE_PERSISTANCE,
    CREATE_GROUP,
    UPDATE_GROUP,
    DELETE_GROUP,
    ADD_USER_TO_GROUP,
    REMOVE_USER_FROM_GROUP,
    CONFIGURE_AGENT_CLIENT,
    REGISTER_AGENT,
    ENABLE_AGENT,
    DISABLE_AGENT,
    DELETE_AGENT,
    ADD_ROLE_TO_AGENT,
    DELETE_ROLE_FROM_AGENT,
    ADD_AGENT_ATTRIBUTES,
    DELETE_AGENT_ATTRIBUTES,
    GRANT_ADMIN_PRIVILEGE,
    REMOVE_ADMIN_PRIVILEGE

}
