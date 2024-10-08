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

// source: src/main/proto/IamAdminService.proto
/**
 * @fileoverview
 * @enhanceable
 * @suppress {messageConventions} JS Compiler reports an error if a variable or
 *     field starts with 'MSG_' and isn't a translatable message.
 * @public
 */
// GENERATED CODE -- DO NOT EDIT!

var jspb = require('google-protobuf');
var goog = jspb;
var global = Function('return this')();

var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js');
goog.object.extend(proto, google_protobuf_empty_pb);
goog.exportSymbol('proto.org.apache.custos.iam.service.AddProtocolMapperRequest', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.AddRolesRequest', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.AddUserAttributesRequest', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.AddUserResponse', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.AddUserRolesRequest', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.Agent', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.AgentClientMetadata', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.AllRoles', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.CheckingResponse', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.ClaimJSONTypes', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.DeleteTenantRequest', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.DeleteUserAttributeRequest', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.DeleteUserRolesRequest', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.EventPersistenceRequest', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.FederateIDPResponse', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.FederatedIDPs', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.FindUsersRequest', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.FindUsersResponse', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.GetAllResources', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.GetAllResourcesResponse', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.GetOperationsMetadataRequest', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.GetOperationsMetadataResponse', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.GroupRepresentation', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.GroupRequest', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.GroupsRequest', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.GroupsResponse', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.IsUsernameAvailableRequest', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.MapperTypes', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.OperationMetadata', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.OperationStatus', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.RegisterUserRequest', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.RegisterUserResponse', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.RegisterUsersRequest', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.RegisterUsersResponse', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.ResetUserPassword', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.ResourceTypes', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.RoleRepresentation', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.SetUpTenantRequest', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.SetUpTenantResponse', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.UpdateUserProfileRequest', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.UserAttribute', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.UserGroupMappingRequest', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.UserRepresentation', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.UserSearchMetadata', null, global);
goog.exportSymbol('proto.org.apache.custos.iam.service.UserSearchRequest', null, global);
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.org.apache.custos.iam.service.SetUpTenantRequest.repeatedFields_, null);
};
goog.inherits(proto.org.apache.custos.iam.service.SetUpTenantRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.SetUpTenantRequest.displayName = 'proto.org.apache.custos.iam.service.SetUpTenantRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.displayName = 'proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.FederateIDPResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.FederateIDPResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.FederateIDPResponse.displayName = 'proto.org.apache.custos.iam.service.FederateIDPResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.SetUpTenantResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.SetUpTenantResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.SetUpTenantResponse.displayName = 'proto.org.apache.custos.iam.service.SetUpTenantResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.IsUsernameAvailableRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.IsUsernameAvailableRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.IsUsernameAvailableRequest.displayName = 'proto.org.apache.custos.iam.service.IsUsernameAvailableRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.CheckingResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.CheckingResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.CheckingResponse.displayName = 'proto.org.apache.custos.iam.service.CheckingResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.UserRepresentation = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.org.apache.custos.iam.service.UserRepresentation.repeatedFields_, null);
};
goog.inherits(proto.org.apache.custos.iam.service.UserRepresentation, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.UserRepresentation.displayName = 'proto.org.apache.custos.iam.service.UserRepresentation';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.GroupRepresentation = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.org.apache.custos.iam.service.GroupRepresentation.repeatedFields_, null);
};
goog.inherits(proto.org.apache.custos.iam.service.GroupRepresentation, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.GroupRepresentation.displayName = 'proto.org.apache.custos.iam.service.GroupRepresentation';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.RegisterUserRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.RegisterUserRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.RegisterUserRequest.displayName = 'proto.org.apache.custos.iam.service.RegisterUserRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.RegisterUsersRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.org.apache.custos.iam.service.RegisterUsersRequest.repeatedFields_, null);
};
goog.inherits(proto.org.apache.custos.iam.service.RegisterUsersRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.RegisterUsersRequest.displayName = 'proto.org.apache.custos.iam.service.RegisterUsersRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.RegisterUserResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.RegisterUserResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.RegisterUserResponse.displayName = 'proto.org.apache.custos.iam.service.RegisterUserResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.RegisterUsersResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.org.apache.custos.iam.service.RegisterUsersResponse.repeatedFields_, null);
};
goog.inherits(proto.org.apache.custos.iam.service.RegisterUsersResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.RegisterUsersResponse.displayName = 'proto.org.apache.custos.iam.service.RegisterUsersResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.UserSearchMetadata = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.UserSearchMetadata, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.UserSearchMetadata.displayName = 'proto.org.apache.custos.iam.service.UserSearchMetadata';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.FindUsersRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.FindUsersRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.FindUsersRequest.displayName = 'proto.org.apache.custos.iam.service.FindUsersRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.UserSearchRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.UserSearchRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.UserSearchRequest.displayName = 'proto.org.apache.custos.iam.service.UserSearchRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.FindUsersResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.org.apache.custos.iam.service.FindUsersResponse.repeatedFields_, null);
};
goog.inherits(proto.org.apache.custos.iam.service.FindUsersResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.FindUsersResponse.displayName = 'proto.org.apache.custos.iam.service.FindUsersResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.ResetUserPassword = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.ResetUserPassword, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.ResetUserPassword.displayName = 'proto.org.apache.custos.iam.service.ResetUserPassword';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.org.apache.custos.iam.service.DeleteUserRolesRequest.repeatedFields_, null);
};
goog.inherits(proto.org.apache.custos.iam.service.DeleteUserRolesRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.DeleteUserRolesRequest.displayName = 'proto.org.apache.custos.iam.service.DeleteUserRolesRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.org.apache.custos.iam.service.AddUserRolesRequest.repeatedFields_, null);
};
goog.inherits(proto.org.apache.custos.iam.service.AddUserRolesRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.AddUserRolesRequest.displayName = 'proto.org.apache.custos.iam.service.AddUserRolesRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.UpdateUserProfileRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.UpdateUserProfileRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.UpdateUserProfileRequest.displayName = 'proto.org.apache.custos.iam.service.UpdateUserProfileRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.AddUserResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.AddUserResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.AddUserResponse.displayName = 'proto.org.apache.custos.iam.service.AddUserResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.GetOperationsMetadataRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.GetOperationsMetadataRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.GetOperationsMetadataRequest.displayName = 'proto.org.apache.custos.iam.service.GetOperationsMetadataRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.OperationMetadata = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.OperationMetadata, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.OperationMetadata.displayName = 'proto.org.apache.custos.iam.service.OperationMetadata';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.GetOperationsMetadataResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.org.apache.custos.iam.service.GetOperationsMetadataResponse.repeatedFields_, null);
};
goog.inherits(proto.org.apache.custos.iam.service.GetOperationsMetadataResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.GetOperationsMetadataResponse.displayName = 'proto.org.apache.custos.iam.service.GetOperationsMetadataResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.DeleteTenantRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.DeleteTenantRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.DeleteTenantRequest.displayName = 'proto.org.apache.custos.iam.service.DeleteTenantRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.AddRolesRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.org.apache.custos.iam.service.AddRolesRequest.repeatedFields_, null);
};
goog.inherits(proto.org.apache.custos.iam.service.AddRolesRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.AddRolesRequest.displayName = 'proto.org.apache.custos.iam.service.AddRolesRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.RoleRepresentation = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.RoleRepresentation, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.RoleRepresentation.displayName = 'proto.org.apache.custos.iam.service.RoleRepresentation';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.AllRoles = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.org.apache.custos.iam.service.AllRoles.repeatedFields_, null);
};
goog.inherits(proto.org.apache.custos.iam.service.AllRoles, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.AllRoles.displayName = 'proto.org.apache.custos.iam.service.AllRoles';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.AddProtocolMapperRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.AddProtocolMapperRequest.displayName = 'proto.org.apache.custos.iam.service.AddProtocolMapperRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.OperationStatus = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.OperationStatus, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.OperationStatus.displayName = 'proto.org.apache.custos.iam.service.OperationStatus';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.org.apache.custos.iam.service.AddUserAttributesRequest.repeatedFields_, null);
};
goog.inherits(proto.org.apache.custos.iam.service.AddUserAttributesRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.AddUserAttributesRequest.displayName = 'proto.org.apache.custos.iam.service.AddUserAttributesRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.repeatedFields_, null);
};
goog.inherits(proto.org.apache.custos.iam.service.DeleteUserAttributeRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.displayName = 'proto.org.apache.custos.iam.service.DeleteUserAttributeRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.UserAttribute = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.org.apache.custos.iam.service.UserAttribute.repeatedFields_, null);
};
goog.inherits(proto.org.apache.custos.iam.service.UserAttribute, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.UserAttribute.displayName = 'proto.org.apache.custos.iam.service.UserAttribute';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.EventPersistenceRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.EventPersistenceRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.EventPersistenceRequest.displayName = 'proto.org.apache.custos.iam.service.EventPersistenceRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.GroupsRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.org.apache.custos.iam.service.GroupsRequest.repeatedFields_, null);
};
goog.inherits(proto.org.apache.custos.iam.service.GroupsRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.GroupsRequest.displayName = 'proto.org.apache.custos.iam.service.GroupsRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.GroupRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.GroupRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.GroupRequest.displayName = 'proto.org.apache.custos.iam.service.GroupRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.GroupsResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.org.apache.custos.iam.service.GroupsResponse.repeatedFields_, null);
};
goog.inherits(proto.org.apache.custos.iam.service.GroupsResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.GroupsResponse.displayName = 'proto.org.apache.custos.iam.service.GroupsResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.UserGroupMappingRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.UserGroupMappingRequest.displayName = 'proto.org.apache.custos.iam.service.UserGroupMappingRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.AgentClientMetadata = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.org.apache.custos.iam.service.AgentClientMetadata.repeatedFields_, null);
};
goog.inherits(proto.org.apache.custos.iam.service.AgentClientMetadata, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.AgentClientMetadata.displayName = 'proto.org.apache.custos.iam.service.AgentClientMetadata';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.Agent = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.org.apache.custos.iam.service.Agent.repeatedFields_, null);
};
goog.inherits(proto.org.apache.custos.iam.service.Agent, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.Agent.displayName = 'proto.org.apache.custos.iam.service.Agent';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.GetAllResources = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.org.apache.custos.iam.service.GetAllResources, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.GetAllResources.displayName = 'proto.org.apache.custos.iam.service.GetAllResources';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.org.apache.custos.iam.service.GetAllResourcesResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.org.apache.custos.iam.service.GetAllResourcesResponse.repeatedFields_, null);
};
goog.inherits(proto.org.apache.custos.iam.service.GetAllResourcesResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.org.apache.custos.iam.service.GetAllResourcesResponse.displayName = 'proto.org.apache.custos.iam.service.GetAllResourcesResponse';
}

/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.repeatedFields_ = [10];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.SetUpTenantRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.SetUpTenantRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    tenantid: jspb.Message.getFieldWithDefault(msg, 1, 0),
    tenantname: jspb.Message.getFieldWithDefault(msg, 2, ""),
    adminusername: jspb.Message.getFieldWithDefault(msg, 3, ""),
    adminfirstname: jspb.Message.getFieldWithDefault(msg, 4, ""),
    adminlastname: jspb.Message.getFieldWithDefault(msg, 5, ""),
    adminemail: jspb.Message.getFieldWithDefault(msg, 6, ""),
    adminpassword: jspb.Message.getFieldWithDefault(msg, 7, ""),
    tenanturl: jspb.Message.getFieldWithDefault(msg, 8, ""),
    requesteremail: jspb.Message.getFieldWithDefault(msg, 9, ""),
    redirecturisList: (f = jspb.Message.getRepeatedField(msg, 10)) == null ? undefined : f,
    custosclientid: jspb.Message.getFieldWithDefault(msg, 11, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.SetUpTenantRequest}
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.SetUpTenantRequest;
  return proto.org.apache.custos.iam.service.SetUpTenantRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.SetUpTenantRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.SetUpTenantRequest}
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantid(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setTenantname(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setAdminusername(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setAdminfirstname(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setAdminlastname(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setAdminemail(value);
      break;
    case 7:
      var value = /** @type {string} */ (reader.readString());
      msg.setAdminpassword(value);
      break;
    case 8:
      var value = /** @type {string} */ (reader.readString());
      msg.setTenanturl(value);
      break;
    case 9:
      var value = /** @type {string} */ (reader.readString());
      msg.setRequesteremail(value);
      break;
    case 10:
      var value = /** @type {string} */ (reader.readString());
      msg.addRedirecturis(value);
      break;
    case 11:
      var value = /** @type {string} */ (reader.readString());
      msg.setCustosclientid(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.SetUpTenantRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.SetUpTenantRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTenantid();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
  f = message.getTenantname();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getAdminusername();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getAdminfirstname();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getAdminlastname();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getAdminemail();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getAdminpassword();
  if (f.length > 0) {
    writer.writeString(
      7,
      f
    );
  }
  f = message.getTenanturl();
  if (f.length > 0) {
    writer.writeString(
      8,
      f
    );
  }
  f = message.getRequesteremail();
  if (f.length > 0) {
    writer.writeString(
      9,
      f
    );
  }
  f = message.getRedirecturisList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      10,
      f
    );
  }
  f = message.getCustosclientid();
  if (f.length > 0) {
    writer.writeString(
      11,
      f
    );
  }
};


/**
 * optional int64 tenantId = 1;
 * @return {number}
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.getTenantid = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.SetUpTenantRequest} returns this
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.setTenantid = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional string tenantName = 2;
 * @return {string}
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.getTenantname = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.SetUpTenantRequest} returns this
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.setTenantname = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string adminUsername = 3;
 * @return {string}
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.getAdminusername = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.SetUpTenantRequest} returns this
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.setAdminusername = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string adminFirstname = 4;
 * @return {string}
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.getAdminfirstname = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.SetUpTenantRequest} returns this
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.setAdminfirstname = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string adminLastname = 5;
 * @return {string}
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.getAdminlastname = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.SetUpTenantRequest} returns this
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.setAdminlastname = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional string adminEmail = 6;
 * @return {string}
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.getAdminemail = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.SetUpTenantRequest} returns this
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.setAdminemail = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * optional string adminPassword = 7;
 * @return {string}
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.getAdminpassword = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 7, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.SetUpTenantRequest} returns this
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.setAdminpassword = function(value) {
  return jspb.Message.setProto3StringField(this, 7, value);
};


/**
 * optional string tenantURL = 8;
 * @return {string}
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.getTenanturl = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 8, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.SetUpTenantRequest} returns this
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.setTenanturl = function(value) {
  return jspb.Message.setProto3StringField(this, 8, value);
};


/**
 * optional string requesterEmail = 9;
 * @return {string}
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.getRequesteremail = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 9, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.SetUpTenantRequest} returns this
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.setRequesteremail = function(value) {
  return jspb.Message.setProto3StringField(this, 9, value);
};


/**
 * repeated string redirectURIs = 10;
 * @return {!Array<string>}
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.getRedirecturisList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 10));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.org.apache.custos.iam.service.SetUpTenantRequest} returns this
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.setRedirecturisList = function(value) {
  return jspb.Message.setField(this, 10, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.SetUpTenantRequest} returns this
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.addRedirecturis = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 10, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.SetUpTenantRequest} returns this
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.clearRedirecturisList = function() {
  return this.setRedirecturisList([]);
};


/**
 * optional string custosClientId = 11;
 * @return {string}
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.getCustosclientid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 11, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.SetUpTenantRequest} returns this
 */
proto.org.apache.custos.iam.service.SetUpTenantRequest.prototype.setCustosclientid = function(value) {
  return jspb.Message.setProto3StringField(this, 11, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    tenantid: jspb.Message.getFieldWithDefault(msg, 1, 0),
    type: jspb.Message.getFieldWithDefault(msg, 2, 0),
    clientid: jspb.Message.getFieldWithDefault(msg, 3, ""),
    clientsec: jspb.Message.getFieldWithDefault(msg, 4, ""),
    configmapMap: (f = msg.getConfigmapMap()) ? f.toObject(includeInstance, undefined) : [],
    requesteremail: jspb.Message.getFieldWithDefault(msg, 6, ""),
    idpid: jspb.Message.getFieldWithDefault(msg, 7, ""),
    scope: jspb.Message.getFieldWithDefault(msg, 8, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest}
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest;
  return proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest}
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantid(value);
      break;
    case 2:
      var value = /** @type {!proto.org.apache.custos.iam.service.FederatedIDPs} */ (reader.readEnum());
      msg.setType(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientid(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientsec(value);
      break;
    case 5:
      var value = msg.getConfigmapMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readString, null, "", "");
         });
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setRequesteremail(value);
      break;
    case 7:
      var value = /** @type {string} */ (reader.readString());
      msg.setIdpid(value);
      break;
    case 8:
      var value = /** @type {string} */ (reader.readString());
      msg.setScope(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTenantid();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
  f = message.getType();
  if (f !== 0.0) {
    writer.writeEnum(
      2,
      f
    );
  }
  f = message.getClientid();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getClientsec();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getConfigmapMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(5, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeString);
  }
  f = message.getRequesteremail();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getIdpid();
  if (f.length > 0) {
    writer.writeString(
      7,
      f
    );
  }
  f = message.getScope();
  if (f.length > 0) {
    writer.writeString(
      8,
      f
    );
  }
};


/**
 * optional int64 tenantId = 1;
 * @return {number}
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.prototype.getTenantid = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest} returns this
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.prototype.setTenantid = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional FederatedIDPs type = 2;
 * @return {!proto.org.apache.custos.iam.service.FederatedIDPs}
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.prototype.getType = function() {
  return /** @type {!proto.org.apache.custos.iam.service.FederatedIDPs} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {!proto.org.apache.custos.iam.service.FederatedIDPs} value
 * @return {!proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest} returns this
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.prototype.setType = function(value) {
  return jspb.Message.setProto3EnumField(this, 2, value);
};


/**
 * optional string clientID = 3;
 * @return {string}
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.prototype.getClientid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest} returns this
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.prototype.setClientid = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string clientSec = 4;
 * @return {string}
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.prototype.getClientsec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest} returns this
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.prototype.setClientsec = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * map<string, string> configMap = 5;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.prototype.getConfigmapMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 5, opt_noLazyCreate,
      null));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest} returns this
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.prototype.clearConfigmapMap = function() {
  this.getConfigmapMap().clear();
  return this;};


/**
 * optional string requesterEmail = 6;
 * @return {string}
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.prototype.getRequesteremail = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest} returns this
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.prototype.setRequesteremail = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * optional string idpId = 7;
 * @return {string}
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.prototype.getIdpid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 7, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest} returns this
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.prototype.setIdpid = function(value) {
  return jspb.Message.setProto3StringField(this, 7, value);
};


/**
 * optional string scope = 8;
 * @return {string}
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.prototype.getScope = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 8, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest} returns this
 */
proto.org.apache.custos.iam.service.ConfigureFederateIDPRequest.prototype.setScope = function(value) {
  return jspb.Message.setProto3StringField(this, 8, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.FederateIDPResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.FederateIDPResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.FederateIDPResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.FederateIDPResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    status: jspb.Message.getBooleanFieldWithDefault(msg, 1, false)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.FederateIDPResponse}
 */
proto.org.apache.custos.iam.service.FederateIDPResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.FederateIDPResponse;
  return proto.org.apache.custos.iam.service.FederateIDPResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.FederateIDPResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.FederateIDPResponse}
 */
proto.org.apache.custos.iam.service.FederateIDPResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setStatus(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.FederateIDPResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.FederateIDPResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.FederateIDPResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.FederateIDPResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getStatus();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
};


/**
 * optional bool status = 1;
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.FederateIDPResponse.prototype.getStatus = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.org.apache.custos.iam.service.FederateIDPResponse} returns this
 */
proto.org.apache.custos.iam.service.FederateIDPResponse.prototype.setStatus = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.SetUpTenantResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.SetUpTenantResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.SetUpTenantResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.SetUpTenantResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    clientid: jspb.Message.getFieldWithDefault(msg, 1, ""),
    clientsecret: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.SetUpTenantResponse}
 */
proto.org.apache.custos.iam.service.SetUpTenantResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.SetUpTenantResponse;
  return proto.org.apache.custos.iam.service.SetUpTenantResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.SetUpTenantResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.SetUpTenantResponse}
 */
proto.org.apache.custos.iam.service.SetUpTenantResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientid(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientsecret(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.SetUpTenantResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.SetUpTenantResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.SetUpTenantResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.SetUpTenantResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getClientid();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getClientsecret();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string clientId = 1;
 * @return {string}
 */
proto.org.apache.custos.iam.service.SetUpTenantResponse.prototype.getClientid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.SetUpTenantResponse} returns this
 */
proto.org.apache.custos.iam.service.SetUpTenantResponse.prototype.setClientid = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string clientSecret = 2;
 * @return {string}
 */
proto.org.apache.custos.iam.service.SetUpTenantResponse.prototype.getClientsecret = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.SetUpTenantResponse} returns this
 */
proto.org.apache.custos.iam.service.SetUpTenantResponse.prototype.setClientsecret = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.IsUsernameAvailableRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.IsUsernameAvailableRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.IsUsernameAvailableRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.IsUsernameAvailableRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    tenantid: jspb.Message.getFieldWithDefault(msg, 1, 0),
    accesstoken: jspb.Message.getFieldWithDefault(msg, 2, ""),
    username: jspb.Message.getFieldWithDefault(msg, 3, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.IsUsernameAvailableRequest}
 */
proto.org.apache.custos.iam.service.IsUsernameAvailableRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.IsUsernameAvailableRequest;
  return proto.org.apache.custos.iam.service.IsUsernameAvailableRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.IsUsernameAvailableRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.IsUsernameAvailableRequest}
 */
proto.org.apache.custos.iam.service.IsUsernameAvailableRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantid(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setAccesstoken(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setUsername(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.IsUsernameAvailableRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.IsUsernameAvailableRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.IsUsernameAvailableRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.IsUsernameAvailableRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTenantid();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
  f = message.getAccesstoken();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getUsername();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
};


/**
 * optional int64 tenantId = 1;
 * @return {number}
 */
proto.org.apache.custos.iam.service.IsUsernameAvailableRequest.prototype.getTenantid = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.IsUsernameAvailableRequest} returns this
 */
proto.org.apache.custos.iam.service.IsUsernameAvailableRequest.prototype.setTenantid = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional string accessToken = 2;
 * @return {string}
 */
proto.org.apache.custos.iam.service.IsUsernameAvailableRequest.prototype.getAccesstoken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.IsUsernameAvailableRequest} returns this
 */
proto.org.apache.custos.iam.service.IsUsernameAvailableRequest.prototype.setAccesstoken = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string userName = 3;
 * @return {string}
 */
proto.org.apache.custos.iam.service.IsUsernameAvailableRequest.prototype.getUsername = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.IsUsernameAvailableRequest} returns this
 */
proto.org.apache.custos.iam.service.IsUsernameAvailableRequest.prototype.setUsername = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.CheckingResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.CheckingResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.CheckingResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.CheckingResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    isExist: jspb.Message.getBooleanFieldWithDefault(msg, 1, false)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.CheckingResponse}
 */
proto.org.apache.custos.iam.service.CheckingResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.CheckingResponse;
  return proto.org.apache.custos.iam.service.CheckingResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.CheckingResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.CheckingResponse}
 */
proto.org.apache.custos.iam.service.CheckingResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setIsExist(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.CheckingResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.CheckingResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.CheckingResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.CheckingResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getIsExist();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
};


/**
 * optional bool is_exist = 1;
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.CheckingResponse.prototype.getIsExist = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.org.apache.custos.iam.service.CheckingResponse} returns this
 */
proto.org.apache.custos.iam.service.CheckingResponse.prototype.setIsExist = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.org.apache.custos.iam.service.UserRepresentation.repeatedFields_ = [9,10,11];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.UserRepresentation.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.UserRepresentation} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.UserRepresentation.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, ""),
    username: jspb.Message.getFieldWithDefault(msg, 3, ""),
    firstName: jspb.Message.getFieldWithDefault(msg, 4, ""),
    lastName: jspb.Message.getFieldWithDefault(msg, 5, ""),
    password: jspb.Message.getFieldWithDefault(msg, 6, ""),
    email: jspb.Message.getFieldWithDefault(msg, 7, ""),
    temporaryPassword: jspb.Message.getBooleanFieldWithDefault(msg, 8, false),
    realmRolesList: (f = jspb.Message.getRepeatedField(msg, 9)) == null ? undefined : f,
    clientRolesList: (f = jspb.Message.getRepeatedField(msg, 10)) == null ? undefined : f,
    attributesList: jspb.Message.toObjectList(msg.getAttributesList(),
    proto.org.apache.custos.iam.service.UserAttribute.toObject, includeInstance),
    state: jspb.Message.getFieldWithDefault(msg, 12, ""),
    creationTime: jspb.Message.getFloatingPointFieldWithDefault(msg, 13, 0.0),
    lastLoginAt: jspb.Message.getFloatingPointFieldWithDefault(msg, 14, 0.0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation}
 */
proto.org.apache.custos.iam.service.UserRepresentation.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.UserRepresentation;
  return proto.org.apache.custos.iam.service.UserRepresentation.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.UserRepresentation} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation}
 */
proto.org.apache.custos.iam.service.UserRepresentation.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setUsername(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setFirstName(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setLastName(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setPassword(value);
      break;
    case 7:
      var value = /** @type {string} */ (reader.readString());
      msg.setEmail(value);
      break;
    case 8:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setTemporaryPassword(value);
      break;
    case 9:
      var value = /** @type {string} */ (reader.readString());
      msg.addRealmRoles(value);
      break;
    case 10:
      var value = /** @type {string} */ (reader.readString());
      msg.addClientRoles(value);
      break;
    case 11:
      var value = new proto.org.apache.custos.iam.service.UserAttribute;
      reader.readMessage(value,proto.org.apache.custos.iam.service.UserAttribute.deserializeBinaryFromReader);
      msg.addAttributes(value);
      break;
    case 12:
      var value = /** @type {string} */ (reader.readString());
      msg.setState(value);
      break;
    case 13:
      var value = /** @type {number} */ (reader.readDouble());
      msg.setCreationTime(value);
      break;
    case 14:
      var value = /** @type {number} */ (reader.readDouble());
      msg.setLastLoginAt(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.UserRepresentation.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.UserRepresentation} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.UserRepresentation.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getUsername();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getFirstName();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getLastName();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getPassword();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getEmail();
  if (f.length > 0) {
    writer.writeString(
      7,
      f
    );
  }
  f = message.getTemporaryPassword();
  if (f) {
    writer.writeBool(
      8,
      f
    );
  }
  f = message.getRealmRolesList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      9,
      f
    );
  }
  f = message.getClientRolesList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      10,
      f
    );
  }
  f = message.getAttributesList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      11,
      f,
      proto.org.apache.custos.iam.service.UserAttribute.serializeBinaryToWriter
    );
  }
  f = message.getState();
  if (f.length > 0) {
    writer.writeString(
      12,
      f
    );
  }
  f = message.getCreationTime();
  if (f !== 0.0) {
    writer.writeDouble(
      13,
      f
    );
  }
  f = message.getLastLoginAt();
  if (f !== 0.0) {
    writer.writeDouble(
      14,
      f
    );
  }
};


/**
 * optional string id = 1;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation} returns this
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.setId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string username = 3;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.getUsername = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation} returns this
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.setUsername = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string first_name = 4;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.getFirstName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation} returns this
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.setFirstName = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string last_name = 5;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.getLastName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation} returns this
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.setLastName = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional string password = 6;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.getPassword = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation} returns this
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.setPassword = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * optional string email = 7;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.getEmail = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 7, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation} returns this
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.setEmail = function(value) {
  return jspb.Message.setProto3StringField(this, 7, value);
};


/**
 * optional bool temporary_password = 8;
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.getTemporaryPassword = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 8, false));
};


/**
 * @param {boolean} value
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation} returns this
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.setTemporaryPassword = function(value) {
  return jspb.Message.setProto3BooleanField(this, 8, value);
};


/**
 * repeated string realm_roles = 9;
 * @return {!Array<string>}
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.getRealmRolesList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 9));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation} returns this
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.setRealmRolesList = function(value) {
  return jspb.Message.setField(this, 9, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation} returns this
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.addRealmRoles = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 9, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation} returns this
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.clearRealmRolesList = function() {
  return this.setRealmRolesList([]);
};


/**
 * repeated string client_roles = 10;
 * @return {!Array<string>}
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.getClientRolesList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 10));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation} returns this
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.setClientRolesList = function(value) {
  return jspb.Message.setField(this, 10, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation} returns this
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.addClientRoles = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 10, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation} returns this
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.clearClientRolesList = function() {
  return this.setClientRolesList([]);
};


/**
 * repeated UserAttribute attributes = 11;
 * @return {!Array<!proto.org.apache.custos.iam.service.UserAttribute>}
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.getAttributesList = function() {
  return /** @type{!Array<!proto.org.apache.custos.iam.service.UserAttribute>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.org.apache.custos.iam.service.UserAttribute, 11));
};


/**
 * @param {!Array<!proto.org.apache.custos.iam.service.UserAttribute>} value
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation} returns this
*/
proto.org.apache.custos.iam.service.UserRepresentation.prototype.setAttributesList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 11, value);
};


/**
 * @param {!proto.org.apache.custos.iam.service.UserAttribute=} opt_value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.UserAttribute}
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.addAttributes = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 11, opt_value, proto.org.apache.custos.iam.service.UserAttribute, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation} returns this
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.clearAttributesList = function() {
  return this.setAttributesList([]);
};


/**
 * optional string state = 12;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.getState = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 12, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation} returns this
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.setState = function(value) {
  return jspb.Message.setProto3StringField(this, 12, value);
};


/**
 * optional double creation_time = 13;
 * @return {number}
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.getCreationTime = function() {
  return /** @type {number} */ (jspb.Message.getFloatingPointFieldWithDefault(this, 13, 0.0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation} returns this
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.setCreationTime = function(value) {
  return jspb.Message.setProto3FloatField(this, 13, value);
};


/**
 * optional double last_login_at = 14;
 * @return {number}
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.getLastLoginAt = function() {
  return /** @type {number} */ (jspb.Message.getFloatingPointFieldWithDefault(this, 14, 0.0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation} returns this
 */
proto.org.apache.custos.iam.service.UserRepresentation.prototype.setLastLoginAt = function(value) {
  return jspb.Message.setProto3FloatField(this, 14, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.org.apache.custos.iam.service.GroupRepresentation.repeatedFields_ = [3,4,5,6,7];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.GroupRepresentation.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.GroupRepresentation} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.GroupRepresentation.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 1, ""),
    id: jspb.Message.getFieldWithDefault(msg, 2, ""),
    realmRolesList: (f = jspb.Message.getRepeatedField(msg, 3)) == null ? undefined : f,
    clientRolesList: (f = jspb.Message.getRepeatedField(msg, 4)) == null ? undefined : f,
    attributesList: jspb.Message.toObjectList(msg.getAttributesList(),
    proto.org.apache.custos.iam.service.UserAttribute.toObject, includeInstance),
    usersList: jspb.Message.toObjectList(msg.getUsersList(),
    proto.org.apache.custos.iam.service.UserRepresentation.toObject, includeInstance),
    subGroupsList: jspb.Message.toObjectList(msg.getSubGroupsList(),
    proto.org.apache.custos.iam.service.GroupRepresentation.toObject, includeInstance)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.GroupRepresentation}
 */
proto.org.apache.custos.iam.service.GroupRepresentation.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.GroupRepresentation;
  return proto.org.apache.custos.iam.service.GroupRepresentation.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.GroupRepresentation} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.GroupRepresentation}
 */
proto.org.apache.custos.iam.service.GroupRepresentation.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.addRealmRoles(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.addClientRoles(value);
      break;
    case 5:
      var value = new proto.org.apache.custos.iam.service.UserAttribute;
      reader.readMessage(value,proto.org.apache.custos.iam.service.UserAttribute.deserializeBinaryFromReader);
      msg.addAttributes(value);
      break;
    case 6:
      var value = new proto.org.apache.custos.iam.service.UserRepresentation;
      reader.readMessage(value,proto.org.apache.custos.iam.service.UserRepresentation.deserializeBinaryFromReader);
      msg.addUsers(value);
      break;
    case 7:
      var value = new proto.org.apache.custos.iam.service.GroupRepresentation;
      reader.readMessage(value,proto.org.apache.custos.iam.service.GroupRepresentation.deserializeBinaryFromReader);
      msg.addSubGroups(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.GroupRepresentation.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.GroupRepresentation} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.GroupRepresentation.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getRealmRolesList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      3,
      f
    );
  }
  f = message.getClientRolesList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      4,
      f
    );
  }
  f = message.getAttributesList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      5,
      f,
      proto.org.apache.custos.iam.service.UserAttribute.serializeBinaryToWriter
    );
  }
  f = message.getUsersList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      6,
      f,
      proto.org.apache.custos.iam.service.UserRepresentation.serializeBinaryToWriter
    );
  }
  f = message.getSubGroupsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      7,
      f,
      proto.org.apache.custos.iam.service.GroupRepresentation.serializeBinaryToWriter
    );
  }
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.GroupRepresentation} returns this
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string id = 2;
 * @return {string}
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.GroupRepresentation} returns this
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.setId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * repeated string realm_roles = 3;
 * @return {!Array<string>}
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.getRealmRolesList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 3));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.org.apache.custos.iam.service.GroupRepresentation} returns this
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.setRealmRolesList = function(value) {
  return jspb.Message.setField(this, 3, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.GroupRepresentation} returns this
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.addRealmRoles = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 3, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.GroupRepresentation} returns this
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.clearRealmRolesList = function() {
  return this.setRealmRolesList([]);
};


/**
 * repeated string client_roles = 4;
 * @return {!Array<string>}
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.getClientRolesList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 4));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.org.apache.custos.iam.service.GroupRepresentation} returns this
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.setClientRolesList = function(value) {
  return jspb.Message.setField(this, 4, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.GroupRepresentation} returns this
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.addClientRoles = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 4, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.GroupRepresentation} returns this
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.clearClientRolesList = function() {
  return this.setClientRolesList([]);
};


/**
 * repeated UserAttribute attributes = 5;
 * @return {!Array<!proto.org.apache.custos.iam.service.UserAttribute>}
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.getAttributesList = function() {
  return /** @type{!Array<!proto.org.apache.custos.iam.service.UserAttribute>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.org.apache.custos.iam.service.UserAttribute, 5));
};


/**
 * @param {!Array<!proto.org.apache.custos.iam.service.UserAttribute>} value
 * @return {!proto.org.apache.custos.iam.service.GroupRepresentation} returns this
*/
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.setAttributesList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 5, value);
};


/**
 * @param {!proto.org.apache.custos.iam.service.UserAttribute=} opt_value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.UserAttribute}
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.addAttributes = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 5, opt_value, proto.org.apache.custos.iam.service.UserAttribute, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.GroupRepresentation} returns this
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.clearAttributesList = function() {
  return this.setAttributesList([]);
};


/**
 * repeated UserRepresentation users = 6;
 * @return {!Array<!proto.org.apache.custos.iam.service.UserRepresentation>}
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.getUsersList = function() {
  return /** @type{!Array<!proto.org.apache.custos.iam.service.UserRepresentation>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.org.apache.custos.iam.service.UserRepresentation, 6));
};


/**
 * @param {!Array<!proto.org.apache.custos.iam.service.UserRepresentation>} value
 * @return {!proto.org.apache.custos.iam.service.GroupRepresentation} returns this
*/
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.setUsersList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 6, value);
};


/**
 * @param {!proto.org.apache.custos.iam.service.UserRepresentation=} opt_value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation}
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.addUsers = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 6, opt_value, proto.org.apache.custos.iam.service.UserRepresentation, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.GroupRepresentation} returns this
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.clearUsersList = function() {
  return this.setUsersList([]);
};


/**
 * repeated GroupRepresentation sub_groups = 7;
 * @return {!Array<!proto.org.apache.custos.iam.service.GroupRepresentation>}
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.getSubGroupsList = function() {
  return /** @type{!Array<!proto.org.apache.custos.iam.service.GroupRepresentation>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.org.apache.custos.iam.service.GroupRepresentation, 7));
};


/**
 * @param {!Array<!proto.org.apache.custos.iam.service.GroupRepresentation>} value
 * @return {!proto.org.apache.custos.iam.service.GroupRepresentation} returns this
*/
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.setSubGroupsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 7, value);
};


/**
 * @param {!proto.org.apache.custos.iam.service.GroupRepresentation=} opt_value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.GroupRepresentation}
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.addSubGroups = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 7, opt_value, proto.org.apache.custos.iam.service.GroupRepresentation, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.GroupRepresentation} returns this
 */
proto.org.apache.custos.iam.service.GroupRepresentation.prototype.clearSubGroupsList = function() {
  return this.setSubGroupsList([]);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.RegisterUserRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.RegisterUserRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.RegisterUserRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.RegisterUserRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    tenantid: jspb.Message.getFieldWithDefault(msg, 1, 0),
    accesstoken: jspb.Message.getFieldWithDefault(msg, 2, ""),
    clientid: jspb.Message.getFieldWithDefault(msg, 3, ""),
    clientsec: jspb.Message.getFieldWithDefault(msg, 4, ""),
    user: (f = msg.getUser()) && proto.org.apache.custos.iam.service.UserRepresentation.toObject(includeInstance, f),
    performedby: jspb.Message.getFieldWithDefault(msg, 6, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.RegisterUserRequest}
 */
proto.org.apache.custos.iam.service.RegisterUserRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.RegisterUserRequest;
  return proto.org.apache.custos.iam.service.RegisterUserRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.RegisterUserRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.RegisterUserRequest}
 */
proto.org.apache.custos.iam.service.RegisterUserRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantid(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setAccesstoken(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientid(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientsec(value);
      break;
    case 5:
      var value = new proto.org.apache.custos.iam.service.UserRepresentation;
      reader.readMessage(value,proto.org.apache.custos.iam.service.UserRepresentation.deserializeBinaryFromReader);
      msg.setUser(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setPerformedby(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.RegisterUserRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.RegisterUserRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.RegisterUserRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.RegisterUserRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTenantid();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
  f = message.getAccesstoken();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getClientid();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getClientsec();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getUser();
  if (f != null) {
    writer.writeMessage(
      5,
      f,
      proto.org.apache.custos.iam.service.UserRepresentation.serializeBinaryToWriter
    );
  }
  f = message.getPerformedby();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
};


/**
 * optional int64 tenantId = 1;
 * @return {number}
 */
proto.org.apache.custos.iam.service.RegisterUserRequest.prototype.getTenantid = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.RegisterUserRequest} returns this
 */
proto.org.apache.custos.iam.service.RegisterUserRequest.prototype.setTenantid = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional string accessToken = 2;
 * @return {string}
 */
proto.org.apache.custos.iam.service.RegisterUserRequest.prototype.getAccesstoken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.RegisterUserRequest} returns this
 */
proto.org.apache.custos.iam.service.RegisterUserRequest.prototype.setAccesstoken = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string clientId = 3;
 * @return {string}
 */
proto.org.apache.custos.iam.service.RegisterUserRequest.prototype.getClientid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.RegisterUserRequest} returns this
 */
proto.org.apache.custos.iam.service.RegisterUserRequest.prototype.setClientid = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string clientSec = 4;
 * @return {string}
 */
proto.org.apache.custos.iam.service.RegisterUserRequest.prototype.getClientsec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.RegisterUserRequest} returns this
 */
proto.org.apache.custos.iam.service.RegisterUserRequest.prototype.setClientsec = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional UserRepresentation user = 5;
 * @return {?proto.org.apache.custos.iam.service.UserRepresentation}
 */
proto.org.apache.custos.iam.service.RegisterUserRequest.prototype.getUser = function() {
  return /** @type{?proto.org.apache.custos.iam.service.UserRepresentation} */ (
    jspb.Message.getWrapperField(this, proto.org.apache.custos.iam.service.UserRepresentation, 5));
};


/**
 * @param {?proto.org.apache.custos.iam.service.UserRepresentation|undefined} value
 * @return {!proto.org.apache.custos.iam.service.RegisterUserRequest} returns this
*/
proto.org.apache.custos.iam.service.RegisterUserRequest.prototype.setUser = function(value) {
  return jspb.Message.setWrapperField(this, 5, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.org.apache.custos.iam.service.RegisterUserRequest} returns this
 */
proto.org.apache.custos.iam.service.RegisterUserRequest.prototype.clearUser = function() {
  return this.setUser(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.RegisterUserRequest.prototype.hasUser = function() {
  return jspb.Message.getField(this, 5) != null;
};


/**
 * optional string performedBy = 6;
 * @return {string}
 */
proto.org.apache.custos.iam.service.RegisterUserRequest.prototype.getPerformedby = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.RegisterUserRequest} returns this
 */
proto.org.apache.custos.iam.service.RegisterUserRequest.prototype.setPerformedby = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.org.apache.custos.iam.service.RegisterUsersRequest.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.RegisterUsersRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.RegisterUsersRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.RegisterUsersRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.RegisterUsersRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    usersList: jspb.Message.toObjectList(msg.getUsersList(),
    proto.org.apache.custos.iam.service.UserRepresentation.toObject, includeInstance),
    tenantid: jspb.Message.getFieldWithDefault(msg, 2, 0),
    accesstoken: jspb.Message.getFieldWithDefault(msg, 3, ""),
    clientid: jspb.Message.getFieldWithDefault(msg, 4, ""),
    performedby: jspb.Message.getFieldWithDefault(msg, 5, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.RegisterUsersRequest}
 */
proto.org.apache.custos.iam.service.RegisterUsersRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.RegisterUsersRequest;
  return proto.org.apache.custos.iam.service.RegisterUsersRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.RegisterUsersRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.RegisterUsersRequest}
 */
proto.org.apache.custos.iam.service.RegisterUsersRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.org.apache.custos.iam.service.UserRepresentation;
      reader.readMessage(value,proto.org.apache.custos.iam.service.UserRepresentation.deserializeBinaryFromReader);
      msg.addUsers(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantid(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setAccesstoken(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientid(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setPerformedby(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.RegisterUsersRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.RegisterUsersRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.RegisterUsersRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.RegisterUsersRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getUsersList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.org.apache.custos.iam.service.UserRepresentation.serializeBinaryToWriter
    );
  }
  f = message.getTenantid();
  if (f !== 0) {
    writer.writeInt64(
      2,
      f
    );
  }
  f = message.getAccesstoken();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getClientid();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getPerformedby();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
};


/**
 * repeated UserRepresentation users = 1;
 * @return {!Array<!proto.org.apache.custos.iam.service.UserRepresentation>}
 */
proto.org.apache.custos.iam.service.RegisterUsersRequest.prototype.getUsersList = function() {
  return /** @type{!Array<!proto.org.apache.custos.iam.service.UserRepresentation>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.org.apache.custos.iam.service.UserRepresentation, 1));
};


/**
 * @param {!Array<!proto.org.apache.custos.iam.service.UserRepresentation>} value
 * @return {!proto.org.apache.custos.iam.service.RegisterUsersRequest} returns this
*/
proto.org.apache.custos.iam.service.RegisterUsersRequest.prototype.setUsersList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.org.apache.custos.iam.service.UserRepresentation=} opt_value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation}
 */
proto.org.apache.custos.iam.service.RegisterUsersRequest.prototype.addUsers = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.org.apache.custos.iam.service.UserRepresentation, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.RegisterUsersRequest} returns this
 */
proto.org.apache.custos.iam.service.RegisterUsersRequest.prototype.clearUsersList = function() {
  return this.setUsersList([]);
};


/**
 * optional int64 tenantId = 2;
 * @return {number}
 */
proto.org.apache.custos.iam.service.RegisterUsersRequest.prototype.getTenantid = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.RegisterUsersRequest} returns this
 */
proto.org.apache.custos.iam.service.RegisterUsersRequest.prototype.setTenantid = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional string accessToken = 3;
 * @return {string}
 */
proto.org.apache.custos.iam.service.RegisterUsersRequest.prototype.getAccesstoken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.RegisterUsersRequest} returns this
 */
proto.org.apache.custos.iam.service.RegisterUsersRequest.prototype.setAccesstoken = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string clientId = 4;
 * @return {string}
 */
proto.org.apache.custos.iam.service.RegisterUsersRequest.prototype.getClientid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.RegisterUsersRequest} returns this
 */
proto.org.apache.custos.iam.service.RegisterUsersRequest.prototype.setClientid = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string performedBy = 5;
 * @return {string}
 */
proto.org.apache.custos.iam.service.RegisterUsersRequest.prototype.getPerformedby = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.RegisterUsersRequest} returns this
 */
proto.org.apache.custos.iam.service.RegisterUsersRequest.prototype.setPerformedby = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.RegisterUserResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.RegisterUserResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.RegisterUserResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.RegisterUserResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    isRegistered: jspb.Message.getBooleanFieldWithDefault(msg, 1, false)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.RegisterUserResponse}
 */
proto.org.apache.custos.iam.service.RegisterUserResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.RegisterUserResponse;
  return proto.org.apache.custos.iam.service.RegisterUserResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.RegisterUserResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.RegisterUserResponse}
 */
proto.org.apache.custos.iam.service.RegisterUserResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setIsRegistered(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.RegisterUserResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.RegisterUserResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.RegisterUserResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.RegisterUserResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getIsRegistered();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
};


/**
 * optional bool is_registered = 1;
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.RegisterUserResponse.prototype.getIsRegistered = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.org.apache.custos.iam.service.RegisterUserResponse} returns this
 */
proto.org.apache.custos.iam.service.RegisterUserResponse.prototype.setIsRegistered = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.org.apache.custos.iam.service.RegisterUsersResponse.repeatedFields_ = [2];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.RegisterUsersResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.RegisterUsersResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.RegisterUsersResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.RegisterUsersResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    alluseresregistered: jspb.Message.getBooleanFieldWithDefault(msg, 1, false),
    failedusersList: jspb.Message.toObjectList(msg.getFailedusersList(),
    proto.org.apache.custos.iam.service.UserRepresentation.toObject, includeInstance)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.RegisterUsersResponse}
 */
proto.org.apache.custos.iam.service.RegisterUsersResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.RegisterUsersResponse;
  return proto.org.apache.custos.iam.service.RegisterUsersResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.RegisterUsersResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.RegisterUsersResponse}
 */
proto.org.apache.custos.iam.service.RegisterUsersResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setAlluseresregistered(value);
      break;
    case 2:
      var value = new proto.org.apache.custos.iam.service.UserRepresentation;
      reader.readMessage(value,proto.org.apache.custos.iam.service.UserRepresentation.deserializeBinaryFromReader);
      msg.addFailedusers(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.RegisterUsersResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.RegisterUsersResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.RegisterUsersResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.RegisterUsersResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getAlluseresregistered();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
  f = message.getFailedusersList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      2,
      f,
      proto.org.apache.custos.iam.service.UserRepresentation.serializeBinaryToWriter
    );
  }
};


/**
 * optional bool allUseresRegistered = 1;
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.RegisterUsersResponse.prototype.getAlluseresregistered = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.org.apache.custos.iam.service.RegisterUsersResponse} returns this
 */
proto.org.apache.custos.iam.service.RegisterUsersResponse.prototype.setAlluseresregistered = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};


/**
 * repeated UserRepresentation failedUsers = 2;
 * @return {!Array<!proto.org.apache.custos.iam.service.UserRepresentation>}
 */
proto.org.apache.custos.iam.service.RegisterUsersResponse.prototype.getFailedusersList = function() {
  return /** @type{!Array<!proto.org.apache.custos.iam.service.UserRepresentation>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.org.apache.custos.iam.service.UserRepresentation, 2));
};


/**
 * @param {!Array<!proto.org.apache.custos.iam.service.UserRepresentation>} value
 * @return {!proto.org.apache.custos.iam.service.RegisterUsersResponse} returns this
*/
proto.org.apache.custos.iam.service.RegisterUsersResponse.prototype.setFailedusersList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 2, value);
};


/**
 * @param {!proto.org.apache.custos.iam.service.UserRepresentation=} opt_value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation}
 */
proto.org.apache.custos.iam.service.RegisterUsersResponse.prototype.addFailedusers = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 2, opt_value, proto.org.apache.custos.iam.service.UserRepresentation, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.RegisterUsersResponse} returns this
 */
proto.org.apache.custos.iam.service.RegisterUsersResponse.prototype.clearFailedusersList = function() {
  return this.setFailedusersList([]);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.UserSearchMetadata.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.UserSearchMetadata.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.UserSearchMetadata} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.UserSearchMetadata.toObject = function(includeInstance, msg) {
  var f, obj = {
    username: jspb.Message.getFieldWithDefault(msg, 1, ""),
    firstName: jspb.Message.getFieldWithDefault(msg, 2, ""),
    lastName: jspb.Message.getFieldWithDefault(msg, 3, ""),
    email: jspb.Message.getFieldWithDefault(msg, 4, ""),
    id: jspb.Message.getFieldWithDefault(msg, 5, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.UserSearchMetadata}
 */
proto.org.apache.custos.iam.service.UserSearchMetadata.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.UserSearchMetadata;
  return proto.org.apache.custos.iam.service.UserSearchMetadata.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.UserSearchMetadata} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.UserSearchMetadata}
 */
proto.org.apache.custos.iam.service.UserSearchMetadata.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setUsername(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setFirstName(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setLastName(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setEmail(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.UserSearchMetadata.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.UserSearchMetadata.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.UserSearchMetadata} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.UserSearchMetadata.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getUsername();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getFirstName();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getLastName();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getEmail();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
};


/**
 * optional string username = 1;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserSearchMetadata.prototype.getUsername = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserSearchMetadata} returns this
 */
proto.org.apache.custos.iam.service.UserSearchMetadata.prototype.setUsername = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string first_name = 2;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserSearchMetadata.prototype.getFirstName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserSearchMetadata} returns this
 */
proto.org.apache.custos.iam.service.UserSearchMetadata.prototype.setFirstName = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string last_name = 3;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserSearchMetadata.prototype.getLastName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserSearchMetadata} returns this
 */
proto.org.apache.custos.iam.service.UserSearchMetadata.prototype.setLastName = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string email = 4;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserSearchMetadata.prototype.getEmail = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserSearchMetadata} returns this
 */
proto.org.apache.custos.iam.service.UserSearchMetadata.prototype.setEmail = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string id = 5;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserSearchMetadata.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserSearchMetadata} returns this
 */
proto.org.apache.custos.iam.service.UserSearchMetadata.prototype.setId = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.FindUsersRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.FindUsersRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.FindUsersRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.FindUsersRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    user: (f = msg.getUser()) && proto.org.apache.custos.iam.service.UserSearchMetadata.toObject(includeInstance, f),
    offset: jspb.Message.getFieldWithDefault(msg, 4, 0),
    limit: jspb.Message.getFieldWithDefault(msg, 5, 0),
    tenantid: jspb.Message.getFieldWithDefault(msg, 1, 0),
    accesstoken: jspb.Message.getFieldWithDefault(msg, 2, ""),
    clientId: jspb.Message.getFieldWithDefault(msg, 6, ""),
    clientSec: jspb.Message.getFieldWithDefault(msg, 7, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.FindUsersRequest}
 */
proto.org.apache.custos.iam.service.FindUsersRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.FindUsersRequest;
  return proto.org.apache.custos.iam.service.FindUsersRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.FindUsersRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.FindUsersRequest}
 */
proto.org.apache.custos.iam.service.FindUsersRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 3:
      var value = new proto.org.apache.custos.iam.service.UserSearchMetadata;
      reader.readMessage(value,proto.org.apache.custos.iam.service.UserSearchMetadata.deserializeBinaryFromReader);
      msg.setUser(value);
      break;
    case 4:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setOffset(value);
      break;
    case 5:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setLimit(value);
      break;
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantid(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setAccesstoken(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientId(value);
      break;
    case 7:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientSec(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.FindUsersRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.FindUsersRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.FindUsersRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.FindUsersRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getUser();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.org.apache.custos.iam.service.UserSearchMetadata.serializeBinaryToWriter
    );
  }
  f = message.getOffset();
  if (f !== 0) {
    writer.writeInt32(
      4,
      f
    );
  }
  f = message.getLimit();
  if (f !== 0) {
    writer.writeInt32(
      5,
      f
    );
  }
  f = message.getTenantid();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
  f = message.getAccesstoken();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getClientId();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getClientSec();
  if (f.length > 0) {
    writer.writeString(
      7,
      f
    );
  }
};


/**
 * optional UserSearchMetadata user = 3;
 * @return {?proto.org.apache.custos.iam.service.UserSearchMetadata}
 */
proto.org.apache.custos.iam.service.FindUsersRequest.prototype.getUser = function() {
  return /** @type{?proto.org.apache.custos.iam.service.UserSearchMetadata} */ (
    jspb.Message.getWrapperField(this, proto.org.apache.custos.iam.service.UserSearchMetadata, 3));
};


/**
 * @param {?proto.org.apache.custos.iam.service.UserSearchMetadata|undefined} value
 * @return {!proto.org.apache.custos.iam.service.FindUsersRequest} returns this
*/
proto.org.apache.custos.iam.service.FindUsersRequest.prototype.setUser = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.org.apache.custos.iam.service.FindUsersRequest} returns this
 */
proto.org.apache.custos.iam.service.FindUsersRequest.prototype.clearUser = function() {
  return this.setUser(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.FindUsersRequest.prototype.hasUser = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional int32 offset = 4;
 * @return {number}
 */
proto.org.apache.custos.iam.service.FindUsersRequest.prototype.getOffset = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.FindUsersRequest} returns this
 */
proto.org.apache.custos.iam.service.FindUsersRequest.prototype.setOffset = function(value) {
  return jspb.Message.setProto3IntField(this, 4, value);
};


/**
 * optional int32 limit = 5;
 * @return {number}
 */
proto.org.apache.custos.iam.service.FindUsersRequest.prototype.getLimit = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.FindUsersRequest} returns this
 */
proto.org.apache.custos.iam.service.FindUsersRequest.prototype.setLimit = function(value) {
  return jspb.Message.setProto3IntField(this, 5, value);
};


/**
 * optional int64 tenantId = 1;
 * @return {number}
 */
proto.org.apache.custos.iam.service.FindUsersRequest.prototype.getTenantid = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.FindUsersRequest} returns this
 */
proto.org.apache.custos.iam.service.FindUsersRequest.prototype.setTenantid = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional string accessToken = 2;
 * @return {string}
 */
proto.org.apache.custos.iam.service.FindUsersRequest.prototype.getAccesstoken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.FindUsersRequest} returns this
 */
proto.org.apache.custos.iam.service.FindUsersRequest.prototype.setAccesstoken = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string client_id = 6;
 * @return {string}
 */
proto.org.apache.custos.iam.service.FindUsersRequest.prototype.getClientId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.FindUsersRequest} returns this
 */
proto.org.apache.custos.iam.service.FindUsersRequest.prototype.setClientId = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * optional string client_sec = 7;
 * @return {string}
 */
proto.org.apache.custos.iam.service.FindUsersRequest.prototype.getClientSec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 7, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.FindUsersRequest} returns this
 */
proto.org.apache.custos.iam.service.FindUsersRequest.prototype.setClientSec = function(value) {
  return jspb.Message.setProto3StringField(this, 7, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.UserSearchRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.UserSearchRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.UserSearchRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    user: (f = msg.getUser()) && proto.org.apache.custos.iam.service.UserSearchMetadata.toObject(includeInstance, f),
    tenantid: jspb.Message.getFieldWithDefault(msg, 2, 0),
    accesstoken: jspb.Message.getFieldWithDefault(msg, 3, ""),
    clientId: jspb.Message.getFieldWithDefault(msg, 4, ""),
    clientSec: jspb.Message.getFieldWithDefault(msg, 5, ""),
    performedby: jspb.Message.getFieldWithDefault(msg, 6, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.UserSearchRequest}
 */
proto.org.apache.custos.iam.service.UserSearchRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.UserSearchRequest;
  return proto.org.apache.custos.iam.service.UserSearchRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.UserSearchRequest}
 */
proto.org.apache.custos.iam.service.UserSearchRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.org.apache.custos.iam.service.UserSearchMetadata;
      reader.readMessage(value,proto.org.apache.custos.iam.service.UserSearchMetadata.deserializeBinaryFromReader);
      msg.setUser(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantid(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setAccesstoken(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientId(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientSec(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setPerformedby(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.UserSearchRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.UserSearchRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.UserSearchRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getUser();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.org.apache.custos.iam.service.UserSearchMetadata.serializeBinaryToWriter
    );
  }
  f = message.getTenantid();
  if (f !== 0) {
    writer.writeInt64(
      2,
      f
    );
  }
  f = message.getAccesstoken();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getClientId();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getClientSec();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getPerformedby();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
};


/**
 * optional UserSearchMetadata user = 1;
 * @return {?proto.org.apache.custos.iam.service.UserSearchMetadata}
 */
proto.org.apache.custos.iam.service.UserSearchRequest.prototype.getUser = function() {
  return /** @type{?proto.org.apache.custos.iam.service.UserSearchMetadata} */ (
    jspb.Message.getWrapperField(this, proto.org.apache.custos.iam.service.UserSearchMetadata, 1));
};


/**
 * @param {?proto.org.apache.custos.iam.service.UserSearchMetadata|undefined} value
 * @return {!proto.org.apache.custos.iam.service.UserSearchRequest} returns this
*/
proto.org.apache.custos.iam.service.UserSearchRequest.prototype.setUser = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.org.apache.custos.iam.service.UserSearchRequest} returns this
 */
proto.org.apache.custos.iam.service.UserSearchRequest.prototype.clearUser = function() {
  return this.setUser(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.UserSearchRequest.prototype.hasUser = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional int64 tenantId = 2;
 * @return {number}
 */
proto.org.apache.custos.iam.service.UserSearchRequest.prototype.getTenantid = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.UserSearchRequest} returns this
 */
proto.org.apache.custos.iam.service.UserSearchRequest.prototype.setTenantid = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional string accessToken = 3;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserSearchRequest.prototype.getAccesstoken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserSearchRequest} returns this
 */
proto.org.apache.custos.iam.service.UserSearchRequest.prototype.setAccesstoken = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string client_id = 4;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserSearchRequest.prototype.getClientId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserSearchRequest} returns this
 */
proto.org.apache.custos.iam.service.UserSearchRequest.prototype.setClientId = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string client_sec = 5;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserSearchRequest.prototype.getClientSec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserSearchRequest} returns this
 */
proto.org.apache.custos.iam.service.UserSearchRequest.prototype.setClientSec = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional string performedBy = 6;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserSearchRequest.prototype.getPerformedby = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserSearchRequest} returns this
 */
proto.org.apache.custos.iam.service.UserSearchRequest.prototype.setPerformedby = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.org.apache.custos.iam.service.FindUsersResponse.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.FindUsersResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.FindUsersResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.FindUsersResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.FindUsersResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    usersList: jspb.Message.toObjectList(msg.getUsersList(),
    proto.org.apache.custos.iam.service.UserRepresentation.toObject, includeInstance)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.FindUsersResponse}
 */
proto.org.apache.custos.iam.service.FindUsersResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.FindUsersResponse;
  return proto.org.apache.custos.iam.service.FindUsersResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.FindUsersResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.FindUsersResponse}
 */
proto.org.apache.custos.iam.service.FindUsersResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.org.apache.custos.iam.service.UserRepresentation;
      reader.readMessage(value,proto.org.apache.custos.iam.service.UserRepresentation.deserializeBinaryFromReader);
      msg.addUsers(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.FindUsersResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.FindUsersResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.FindUsersResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.FindUsersResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getUsersList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.org.apache.custos.iam.service.UserRepresentation.serializeBinaryToWriter
    );
  }
};


/**
 * repeated UserRepresentation users = 1;
 * @return {!Array<!proto.org.apache.custos.iam.service.UserRepresentation>}
 */
proto.org.apache.custos.iam.service.FindUsersResponse.prototype.getUsersList = function() {
  return /** @type{!Array<!proto.org.apache.custos.iam.service.UserRepresentation>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.org.apache.custos.iam.service.UserRepresentation, 1));
};


/**
 * @param {!Array<!proto.org.apache.custos.iam.service.UserRepresentation>} value
 * @return {!proto.org.apache.custos.iam.service.FindUsersResponse} returns this
*/
proto.org.apache.custos.iam.service.FindUsersResponse.prototype.setUsersList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.org.apache.custos.iam.service.UserRepresentation=} opt_value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation}
 */
proto.org.apache.custos.iam.service.FindUsersResponse.prototype.addUsers = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.org.apache.custos.iam.service.UserRepresentation, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.FindUsersResponse} returns this
 */
proto.org.apache.custos.iam.service.FindUsersResponse.prototype.clearUsersList = function() {
  return this.setUsersList([]);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.ResetUserPassword.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.ResetUserPassword.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.ResetUserPassword} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.ResetUserPassword.toObject = function(includeInstance, msg) {
  var f, obj = {
    username: jspb.Message.getFieldWithDefault(msg, 1, ""),
    password: jspb.Message.getFieldWithDefault(msg, 2, ""),
    tenantid: jspb.Message.getFieldWithDefault(msg, 3, 0),
    accesstoken: jspb.Message.getFieldWithDefault(msg, 4, ""),
    clientid: jspb.Message.getFieldWithDefault(msg, 5, ""),
    clientsec: jspb.Message.getFieldWithDefault(msg, 6, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.ResetUserPassword}
 */
proto.org.apache.custos.iam.service.ResetUserPassword.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.ResetUserPassword;
  return proto.org.apache.custos.iam.service.ResetUserPassword.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.ResetUserPassword} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.ResetUserPassword}
 */
proto.org.apache.custos.iam.service.ResetUserPassword.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setUsername(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setPassword(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantid(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setAccesstoken(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientid(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientsec(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.ResetUserPassword.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.ResetUserPassword.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.ResetUserPassword} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.ResetUserPassword.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getUsername();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getPassword();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getTenantid();
  if (f !== 0) {
    writer.writeInt64(
      3,
      f
    );
  }
  f = message.getAccesstoken();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getClientid();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getClientsec();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
};


/**
 * optional string username = 1;
 * @return {string}
 */
proto.org.apache.custos.iam.service.ResetUserPassword.prototype.getUsername = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.ResetUserPassword} returns this
 */
proto.org.apache.custos.iam.service.ResetUserPassword.prototype.setUsername = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string password = 2;
 * @return {string}
 */
proto.org.apache.custos.iam.service.ResetUserPassword.prototype.getPassword = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.ResetUserPassword} returns this
 */
proto.org.apache.custos.iam.service.ResetUserPassword.prototype.setPassword = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional int64 tenantId = 3;
 * @return {number}
 */
proto.org.apache.custos.iam.service.ResetUserPassword.prototype.getTenantid = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.ResetUserPassword} returns this
 */
proto.org.apache.custos.iam.service.ResetUserPassword.prototype.setTenantid = function(value) {
  return jspb.Message.setProto3IntField(this, 3, value);
};


/**
 * optional string accessToken = 4;
 * @return {string}
 */
proto.org.apache.custos.iam.service.ResetUserPassword.prototype.getAccesstoken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.ResetUserPassword} returns this
 */
proto.org.apache.custos.iam.service.ResetUserPassword.prototype.setAccesstoken = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string clientId = 5;
 * @return {string}
 */
proto.org.apache.custos.iam.service.ResetUserPassword.prototype.getClientid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.ResetUserPassword} returns this
 */
proto.org.apache.custos.iam.service.ResetUserPassword.prototype.setClientid = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional string clientSec = 6;
 * @return {string}
 */
proto.org.apache.custos.iam.service.ResetUserPassword.prototype.getClientsec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.ResetUserPassword} returns this
 */
proto.org.apache.custos.iam.service.ResetUserPassword.prototype.setClientsec = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.repeatedFields_ = [3,4];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.DeleteUserRolesRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    tenantId: jspb.Message.getFieldWithDefault(msg, 1, 0),
    username: jspb.Message.getFieldWithDefault(msg, 2, ""),
    clientRolesList: (f = jspb.Message.getRepeatedField(msg, 3)) == null ? undefined : f,
    rolesList: (f = jspb.Message.getRepeatedField(msg, 4)) == null ? undefined : f,
    accessToken: jspb.Message.getFieldWithDefault(msg, 5, ""),
    clientId: jspb.Message.getFieldWithDefault(msg, 6, ""),
    performedBy: jspb.Message.getFieldWithDefault(msg, 7, ""),
    id: jspb.Message.getFieldWithDefault(msg, 8, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest}
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.DeleteUserRolesRequest;
  return proto.org.apache.custos.iam.service.DeleteUserRolesRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest}
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setUsername(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.addClientRoles(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.addRoles(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setAccessToken(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientId(value);
      break;
    case 7:
      var value = /** @type {string} */ (reader.readString());
      msg.setPerformedBy(value);
      break;
    case 8:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.DeleteUserRolesRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTenantId();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
  f = message.getUsername();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getClientRolesList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      3,
      f
    );
  }
  f = message.getRolesList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      4,
      f
    );
  }
  f = message.getAccessToken();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getClientId();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getPerformedBy();
  if (f.length > 0) {
    writer.writeString(
      7,
      f
    );
  }
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      8,
      f
    );
  }
};


/**
 * optional int64 tenant_id = 1;
 * @return {number}
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.getTenantId = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.setTenantId = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional string username = 2;
 * @return {string}
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.getUsername = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.setUsername = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * repeated string client_roles = 3;
 * @return {!Array<string>}
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.getClientRolesList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 3));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.setClientRolesList = function(value) {
  return jspb.Message.setField(this, 3, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.addClientRoles = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 3, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.clearClientRolesList = function() {
  return this.setClientRolesList([]);
};


/**
 * repeated string roles = 4;
 * @return {!Array<string>}
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.getRolesList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 4));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.setRolesList = function(value) {
  return jspb.Message.setField(this, 4, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.addRoles = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 4, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.clearRolesList = function() {
  return this.setRolesList([]);
};


/**
 * optional string access_token = 5;
 * @return {string}
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.getAccessToken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.setAccessToken = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional string client_id = 6;
 * @return {string}
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.getClientId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.setClientId = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * optional string performed_by = 7;
 * @return {string}
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.getPerformedBy = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 7, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.setPerformedBy = function(value) {
  return jspb.Message.setProto3StringField(this, 7, value);
};


/**
 * optional string id = 8;
 * @return {string}
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 8, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserRolesRequest.prototype.setId = function(value) {
  return jspb.Message.setProto3StringField(this, 8, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.repeatedFields_ = [2,3,8];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.AddUserRolesRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.AddUserRolesRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    tenantId: jspb.Message.getFieldWithDefault(msg, 1, 0),
    usernamesList: (f = jspb.Message.getRepeatedField(msg, 2)) == null ? undefined : f,
    rolesList: (f = jspb.Message.getRepeatedField(msg, 3)) == null ? undefined : f,
    accessToken: jspb.Message.getFieldWithDefault(msg, 4, ""),
    clientId: jspb.Message.getFieldWithDefault(msg, 5, ""),
    clientLevel: jspb.Message.getBooleanFieldWithDefault(msg, 6, false),
    performedBy: jspb.Message.getFieldWithDefault(msg, 7, ""),
    agentsList: (f = jspb.Message.getRepeatedField(msg, 8)) == null ? undefined : f
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.AddUserRolesRequest}
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.AddUserRolesRequest;
  return proto.org.apache.custos.iam.service.AddUserRolesRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.AddUserRolesRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.AddUserRolesRequest}
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.addUsernames(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.addRoles(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setAccessToken(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientId(value);
      break;
    case 6:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setClientLevel(value);
      break;
    case 7:
      var value = /** @type {string} */ (reader.readString());
      msg.setPerformedBy(value);
      break;
    case 8:
      var value = /** @type {string} */ (reader.readString());
      msg.addAgents(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.AddUserRolesRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.AddUserRolesRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTenantId();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
  f = message.getUsernamesList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      2,
      f
    );
  }
  f = message.getRolesList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      3,
      f
    );
  }
  f = message.getAccessToken();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getClientId();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getClientLevel();
  if (f) {
    writer.writeBool(
      6,
      f
    );
  }
  f = message.getPerformedBy();
  if (f.length > 0) {
    writer.writeString(
      7,
      f
    );
  }
  f = message.getAgentsList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      8,
      f
    );
  }
};


/**
 * optional int64 tenant_id = 1;
 * @return {number}
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.getTenantId = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.AddUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.setTenantId = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * repeated string usernames = 2;
 * @return {!Array<string>}
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.getUsernamesList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 2));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.org.apache.custos.iam.service.AddUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.setUsernamesList = function(value) {
  return jspb.Message.setField(this, 2, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.AddUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.addUsernames = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 2, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.AddUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.clearUsernamesList = function() {
  return this.setUsernamesList([]);
};


/**
 * repeated string roles = 3;
 * @return {!Array<string>}
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.getRolesList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 3));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.org.apache.custos.iam.service.AddUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.setRolesList = function(value) {
  return jspb.Message.setField(this, 3, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.AddUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.addRoles = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 3, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.AddUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.clearRolesList = function() {
  return this.setRolesList([]);
};


/**
 * optional string access_token = 4;
 * @return {string}
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.getAccessToken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.AddUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.setAccessToken = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string client_id = 5;
 * @return {string}
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.getClientId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.AddUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.setClientId = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional bool client_level = 6;
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.getClientLevel = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 6, false));
};


/**
 * @param {boolean} value
 * @return {!proto.org.apache.custos.iam.service.AddUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.setClientLevel = function(value) {
  return jspb.Message.setProto3BooleanField(this, 6, value);
};


/**
 * optional string performed_by = 7;
 * @return {string}
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.getPerformedBy = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 7, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.AddUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.setPerformedBy = function(value) {
  return jspb.Message.setProto3StringField(this, 7, value);
};


/**
 * repeated string agents = 8;
 * @return {!Array<string>}
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.getAgentsList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 8));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.org.apache.custos.iam.service.AddUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.setAgentsList = function(value) {
  return jspb.Message.setField(this, 8, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.AddUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.addAgents = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 8, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.AddUserRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserRolesRequest.prototype.clearAgentsList = function() {
  return this.setAgentsList([]);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.UpdateUserProfileRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.UpdateUserProfileRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.UpdateUserProfileRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.UpdateUserProfileRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    accesstoken: jspb.Message.getFieldWithDefault(msg, 1, ""),
    tenantid: jspb.Message.getFieldWithDefault(msg, 2, 0),
    user: (f = msg.getUser()) && proto.org.apache.custos.iam.service.UserRepresentation.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.UpdateUserProfileRequest}
 */
proto.org.apache.custos.iam.service.UpdateUserProfileRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.UpdateUserProfileRequest;
  return proto.org.apache.custos.iam.service.UpdateUserProfileRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.UpdateUserProfileRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.UpdateUserProfileRequest}
 */
proto.org.apache.custos.iam.service.UpdateUserProfileRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setAccesstoken(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantid(value);
      break;
    case 3:
      var value = new proto.org.apache.custos.iam.service.UserRepresentation;
      reader.readMessage(value,proto.org.apache.custos.iam.service.UserRepresentation.deserializeBinaryFromReader);
      msg.setUser(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.UpdateUserProfileRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.UpdateUserProfileRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.UpdateUserProfileRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.UpdateUserProfileRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getAccesstoken();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getTenantid();
  if (f !== 0) {
    writer.writeInt64(
      2,
      f
    );
  }
  f = message.getUser();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.org.apache.custos.iam.service.UserRepresentation.serializeBinaryToWriter
    );
  }
};


/**
 * optional string accessToken = 1;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UpdateUserProfileRequest.prototype.getAccesstoken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UpdateUserProfileRequest} returns this
 */
proto.org.apache.custos.iam.service.UpdateUserProfileRequest.prototype.setAccesstoken = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional int64 tenantId = 2;
 * @return {number}
 */
proto.org.apache.custos.iam.service.UpdateUserProfileRequest.prototype.getTenantid = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.UpdateUserProfileRequest} returns this
 */
proto.org.apache.custos.iam.service.UpdateUserProfileRequest.prototype.setTenantid = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional UserRepresentation user = 3;
 * @return {?proto.org.apache.custos.iam.service.UserRepresentation}
 */
proto.org.apache.custos.iam.service.UpdateUserProfileRequest.prototype.getUser = function() {
  return /** @type{?proto.org.apache.custos.iam.service.UserRepresentation} */ (
    jspb.Message.getWrapperField(this, proto.org.apache.custos.iam.service.UserRepresentation, 3));
};


/**
 * @param {?proto.org.apache.custos.iam.service.UserRepresentation|undefined} value
 * @return {!proto.org.apache.custos.iam.service.UpdateUserProfileRequest} returns this
*/
proto.org.apache.custos.iam.service.UpdateUserProfileRequest.prototype.setUser = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.org.apache.custos.iam.service.UpdateUserProfileRequest} returns this
 */
proto.org.apache.custos.iam.service.UpdateUserProfileRequest.prototype.clearUser = function() {
  return this.setUser(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.UpdateUserProfileRequest.prototype.hasUser = function() {
  return jspb.Message.getField(this, 3) != null;
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.AddUserResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.AddUserResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.AddUserResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.AddUserResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    code: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.AddUserResponse}
 */
proto.org.apache.custos.iam.service.AddUserResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.AddUserResponse;
  return proto.org.apache.custos.iam.service.AddUserResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.AddUserResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.AddUserResponse}
 */
proto.org.apache.custos.iam.service.AddUserResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setCode(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.AddUserResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.AddUserResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.AddUserResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.AddUserResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCode();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string code = 1;
 * @return {string}
 */
proto.org.apache.custos.iam.service.AddUserResponse.prototype.getCode = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.AddUserResponse} returns this
 */
proto.org.apache.custos.iam.service.AddUserResponse.prototype.setCode = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.GetOperationsMetadataRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.GetOperationsMetadataRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.GetOperationsMetadataRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.GetOperationsMetadataRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    traceid: jspb.Message.getFieldWithDefault(msg, 1, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.GetOperationsMetadataRequest}
 */
proto.org.apache.custos.iam.service.GetOperationsMetadataRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.GetOperationsMetadataRequest;
  return proto.org.apache.custos.iam.service.GetOperationsMetadataRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.GetOperationsMetadataRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.GetOperationsMetadataRequest}
 */
proto.org.apache.custos.iam.service.GetOperationsMetadataRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTraceid(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.GetOperationsMetadataRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.GetOperationsMetadataRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.GetOperationsMetadataRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.GetOperationsMetadataRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTraceid();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
};


/**
 * optional int64 traceId = 1;
 * @return {number}
 */
proto.org.apache.custos.iam.service.GetOperationsMetadataRequest.prototype.getTraceid = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.GetOperationsMetadataRequest} returns this
 */
proto.org.apache.custos.iam.service.GetOperationsMetadataRequest.prototype.setTraceid = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.OperationMetadata.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.OperationMetadata.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.OperationMetadata} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.OperationMetadata.toObject = function(includeInstance, msg) {
  var f, obj = {
    event: jspb.Message.getFieldWithDefault(msg, 1, ""),
    status: jspb.Message.getFieldWithDefault(msg, 2, ""),
    timestamp: jspb.Message.getFieldWithDefault(msg, 3, ""),
    performedby: jspb.Message.getFieldWithDefault(msg, 4, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.OperationMetadata}
 */
proto.org.apache.custos.iam.service.OperationMetadata.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.OperationMetadata;
  return proto.org.apache.custos.iam.service.OperationMetadata.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.OperationMetadata} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.OperationMetadata}
 */
proto.org.apache.custos.iam.service.OperationMetadata.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setEvent(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setStatus(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setTimestamp(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setPerformedby(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.OperationMetadata.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.OperationMetadata.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.OperationMetadata} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.OperationMetadata.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getEvent();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getStatus();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getTimestamp();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getPerformedby();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
};


/**
 * optional string event = 1;
 * @return {string}
 */
proto.org.apache.custos.iam.service.OperationMetadata.prototype.getEvent = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.OperationMetadata} returns this
 */
proto.org.apache.custos.iam.service.OperationMetadata.prototype.setEvent = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string status = 2;
 * @return {string}
 */
proto.org.apache.custos.iam.service.OperationMetadata.prototype.getStatus = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.OperationMetadata} returns this
 */
proto.org.apache.custos.iam.service.OperationMetadata.prototype.setStatus = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string timeStamp = 3;
 * @return {string}
 */
proto.org.apache.custos.iam.service.OperationMetadata.prototype.getTimestamp = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.OperationMetadata} returns this
 */
proto.org.apache.custos.iam.service.OperationMetadata.prototype.setTimestamp = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string performedBy = 4;
 * @return {string}
 */
proto.org.apache.custos.iam.service.OperationMetadata.prototype.getPerformedby = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.OperationMetadata} returns this
 */
proto.org.apache.custos.iam.service.OperationMetadata.prototype.setPerformedby = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.org.apache.custos.iam.service.GetOperationsMetadataResponse.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.GetOperationsMetadataResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.GetOperationsMetadataResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.GetOperationsMetadataResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.GetOperationsMetadataResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    metadataList: jspb.Message.toObjectList(msg.getMetadataList(),
    proto.org.apache.custos.iam.service.OperationMetadata.toObject, includeInstance)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.GetOperationsMetadataResponse}
 */
proto.org.apache.custos.iam.service.GetOperationsMetadataResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.GetOperationsMetadataResponse;
  return proto.org.apache.custos.iam.service.GetOperationsMetadataResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.GetOperationsMetadataResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.GetOperationsMetadataResponse}
 */
proto.org.apache.custos.iam.service.GetOperationsMetadataResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.org.apache.custos.iam.service.OperationMetadata;
      reader.readMessage(value,proto.org.apache.custos.iam.service.OperationMetadata.deserializeBinaryFromReader);
      msg.addMetadata(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.GetOperationsMetadataResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.GetOperationsMetadataResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.GetOperationsMetadataResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.GetOperationsMetadataResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getMetadataList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.org.apache.custos.iam.service.OperationMetadata.serializeBinaryToWriter
    );
  }
};


/**
 * repeated OperationMetadata metadata = 1;
 * @return {!Array<!proto.org.apache.custos.iam.service.OperationMetadata>}
 */
proto.org.apache.custos.iam.service.GetOperationsMetadataResponse.prototype.getMetadataList = function() {
  return /** @type{!Array<!proto.org.apache.custos.iam.service.OperationMetadata>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.org.apache.custos.iam.service.OperationMetadata, 1));
};


/**
 * @param {!Array<!proto.org.apache.custos.iam.service.OperationMetadata>} value
 * @return {!proto.org.apache.custos.iam.service.GetOperationsMetadataResponse} returns this
*/
proto.org.apache.custos.iam.service.GetOperationsMetadataResponse.prototype.setMetadataList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.org.apache.custos.iam.service.OperationMetadata=} opt_value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.OperationMetadata}
 */
proto.org.apache.custos.iam.service.GetOperationsMetadataResponse.prototype.addMetadata = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.org.apache.custos.iam.service.OperationMetadata, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.GetOperationsMetadataResponse} returns this
 */
proto.org.apache.custos.iam.service.GetOperationsMetadataResponse.prototype.clearMetadataList = function() {
  return this.setMetadataList([]);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.DeleteTenantRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.DeleteTenantRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.DeleteTenantRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.DeleteTenantRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    tenantid: jspb.Message.getFieldWithDefault(msg, 1, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.DeleteTenantRequest}
 */
proto.org.apache.custos.iam.service.DeleteTenantRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.DeleteTenantRequest;
  return proto.org.apache.custos.iam.service.DeleteTenantRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.DeleteTenantRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.DeleteTenantRequest}
 */
proto.org.apache.custos.iam.service.DeleteTenantRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantid(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.DeleteTenantRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.DeleteTenantRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.DeleteTenantRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.DeleteTenantRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTenantid();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
};


/**
 * optional int64 tenantId = 1;
 * @return {number}
 */
proto.org.apache.custos.iam.service.DeleteTenantRequest.prototype.getTenantid = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.DeleteTenantRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteTenantRequest.prototype.setTenantid = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.org.apache.custos.iam.service.AddRolesRequest.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.AddRolesRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.AddRolesRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.AddRolesRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.AddRolesRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    rolesList: jspb.Message.toObjectList(msg.getRolesList(),
    proto.org.apache.custos.iam.service.RoleRepresentation.toObject, includeInstance),
    clientLevel: jspb.Message.getBooleanFieldWithDefault(msg, 2, false),
    tenantId: jspb.Message.getFieldWithDefault(msg, 3, 0),
    clientId: jspb.Message.getFieldWithDefault(msg, 4, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.AddRolesRequest}
 */
proto.org.apache.custos.iam.service.AddRolesRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.AddRolesRequest;
  return proto.org.apache.custos.iam.service.AddRolesRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.AddRolesRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.AddRolesRequest}
 */
proto.org.apache.custos.iam.service.AddRolesRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.org.apache.custos.iam.service.RoleRepresentation;
      reader.readMessage(value,proto.org.apache.custos.iam.service.RoleRepresentation.deserializeBinaryFromReader);
      msg.addRoles(value);
      break;
    case 2:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setClientLevel(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantId(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.AddRolesRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.AddRolesRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.AddRolesRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.AddRolesRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRolesList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.org.apache.custos.iam.service.RoleRepresentation.serializeBinaryToWriter
    );
  }
  f = message.getClientLevel();
  if (f) {
    writer.writeBool(
      2,
      f
    );
  }
  f = message.getTenantId();
  if (f !== 0) {
    writer.writeInt64(
      3,
      f
    );
  }
  f = message.getClientId();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
};


/**
 * repeated RoleRepresentation roles = 1;
 * @return {!Array<!proto.org.apache.custos.iam.service.RoleRepresentation>}
 */
proto.org.apache.custos.iam.service.AddRolesRequest.prototype.getRolesList = function() {
  return /** @type{!Array<!proto.org.apache.custos.iam.service.RoleRepresentation>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.org.apache.custos.iam.service.RoleRepresentation, 1));
};


/**
 * @param {!Array<!proto.org.apache.custos.iam.service.RoleRepresentation>} value
 * @return {!proto.org.apache.custos.iam.service.AddRolesRequest} returns this
*/
proto.org.apache.custos.iam.service.AddRolesRequest.prototype.setRolesList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.org.apache.custos.iam.service.RoleRepresentation=} opt_value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.RoleRepresentation}
 */
proto.org.apache.custos.iam.service.AddRolesRequest.prototype.addRoles = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.org.apache.custos.iam.service.RoleRepresentation, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.AddRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddRolesRequest.prototype.clearRolesList = function() {
  return this.setRolesList([]);
};


/**
 * optional bool client_level = 2;
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.AddRolesRequest.prototype.getClientLevel = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.org.apache.custos.iam.service.AddRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddRolesRequest.prototype.setClientLevel = function(value) {
  return jspb.Message.setProto3BooleanField(this, 2, value);
};


/**
 * optional int64 tenant_id = 3;
 * @return {number}
 */
proto.org.apache.custos.iam.service.AddRolesRequest.prototype.getTenantId = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.AddRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddRolesRequest.prototype.setTenantId = function(value) {
  return jspb.Message.setProto3IntField(this, 3, value);
};


/**
 * optional string client_id = 4;
 * @return {string}
 */
proto.org.apache.custos.iam.service.AddRolesRequest.prototype.getClientId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.AddRolesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddRolesRequest.prototype.setClientId = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.RoleRepresentation.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.RoleRepresentation.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.RoleRepresentation} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.RoleRepresentation.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 1, ""),
    description: jspb.Message.getFieldWithDefault(msg, 2, ""),
    composite: jspb.Message.getBooleanFieldWithDefault(msg, 3, false)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.RoleRepresentation}
 */
proto.org.apache.custos.iam.service.RoleRepresentation.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.RoleRepresentation;
  return proto.org.apache.custos.iam.service.RoleRepresentation.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.RoleRepresentation} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.RoleRepresentation}
 */
proto.org.apache.custos.iam.service.RoleRepresentation.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setDescription(value);
      break;
    case 3:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setComposite(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.RoleRepresentation.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.RoleRepresentation.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.RoleRepresentation} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.RoleRepresentation.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getDescription();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getComposite();
  if (f) {
    writer.writeBool(
      3,
      f
    );
  }
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.org.apache.custos.iam.service.RoleRepresentation.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.RoleRepresentation} returns this
 */
proto.org.apache.custos.iam.service.RoleRepresentation.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string description = 2;
 * @return {string}
 */
proto.org.apache.custos.iam.service.RoleRepresentation.prototype.getDescription = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.RoleRepresentation} returns this
 */
proto.org.apache.custos.iam.service.RoleRepresentation.prototype.setDescription = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional bool composite = 3;
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.RoleRepresentation.prototype.getComposite = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 3, false));
};


/**
 * @param {boolean} value
 * @return {!proto.org.apache.custos.iam.service.RoleRepresentation} returns this
 */
proto.org.apache.custos.iam.service.RoleRepresentation.prototype.setComposite = function(value) {
  return jspb.Message.setProto3BooleanField(this, 3, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.org.apache.custos.iam.service.AllRoles.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.AllRoles.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.AllRoles.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.AllRoles} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.AllRoles.toObject = function(includeInstance, msg) {
  var f, obj = {
    rolesList: jspb.Message.toObjectList(msg.getRolesList(),
    proto.org.apache.custos.iam.service.RoleRepresentation.toObject, includeInstance),
    scope: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.AllRoles}
 */
proto.org.apache.custos.iam.service.AllRoles.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.AllRoles;
  return proto.org.apache.custos.iam.service.AllRoles.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.AllRoles} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.AllRoles}
 */
proto.org.apache.custos.iam.service.AllRoles.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.org.apache.custos.iam.service.RoleRepresentation;
      reader.readMessage(value,proto.org.apache.custos.iam.service.RoleRepresentation.deserializeBinaryFromReader);
      msg.addRoles(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setScope(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.AllRoles.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.AllRoles.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.AllRoles} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.AllRoles.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRolesList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.org.apache.custos.iam.service.RoleRepresentation.serializeBinaryToWriter
    );
  }
  f = message.getScope();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * repeated RoleRepresentation roles = 1;
 * @return {!Array<!proto.org.apache.custos.iam.service.RoleRepresentation>}
 */
proto.org.apache.custos.iam.service.AllRoles.prototype.getRolesList = function() {
  return /** @type{!Array<!proto.org.apache.custos.iam.service.RoleRepresentation>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.org.apache.custos.iam.service.RoleRepresentation, 1));
};


/**
 * @param {!Array<!proto.org.apache.custos.iam.service.RoleRepresentation>} value
 * @return {!proto.org.apache.custos.iam.service.AllRoles} returns this
*/
proto.org.apache.custos.iam.service.AllRoles.prototype.setRolesList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.org.apache.custos.iam.service.RoleRepresentation=} opt_value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.RoleRepresentation}
 */
proto.org.apache.custos.iam.service.AllRoles.prototype.addRoles = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.org.apache.custos.iam.service.RoleRepresentation, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.AllRoles} returns this
 */
proto.org.apache.custos.iam.service.AllRoles.prototype.clearRolesList = function() {
  return this.setRolesList([]);
};


/**
 * optional string scope = 2;
 * @return {string}
 */
proto.org.apache.custos.iam.service.AllRoles.prototype.getScope = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.AllRoles} returns this
 */
proto.org.apache.custos.iam.service.AllRoles.prototype.setScope = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.AddProtocolMapperRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 1, ""),
    attributeName: jspb.Message.getFieldWithDefault(msg, 2, ""),
    claimName: jspb.Message.getFieldWithDefault(msg, 3, ""),
    claimType: jspb.Message.getFieldWithDefault(msg, 4, 0),
    tenantId: jspb.Message.getFieldWithDefault(msg, 6, 0),
    clientId: jspb.Message.getFieldWithDefault(msg, 7, ""),
    mapperType: jspb.Message.getFieldWithDefault(msg, 8, 0),
    addToIdToken: jspb.Message.getBooleanFieldWithDefault(msg, 9, false),
    addToAccessToken: jspb.Message.getBooleanFieldWithDefault(msg, 10, false),
    addToUserInfo: jspb.Message.getBooleanFieldWithDefault(msg, 11, false),
    multiValued: jspb.Message.getBooleanFieldWithDefault(msg, 12, false),
    aggregateAttributeValues: jspb.Message.getBooleanFieldWithDefault(msg, 13, false)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest}
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.AddProtocolMapperRequest;
  return proto.org.apache.custos.iam.service.AddProtocolMapperRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest}
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setAttributeName(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setClaimName(value);
      break;
    case 4:
      var value = /** @type {!proto.org.apache.custos.iam.service.ClaimJSONTypes} */ (reader.readEnum());
      msg.setClaimType(value);
      break;
    case 6:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantId(value);
      break;
    case 7:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientId(value);
      break;
    case 8:
      var value = /** @type {!proto.org.apache.custos.iam.service.MapperTypes} */ (reader.readEnum());
      msg.setMapperType(value);
      break;
    case 9:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setAddToIdToken(value);
      break;
    case 10:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setAddToAccessToken(value);
      break;
    case 11:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setAddToUserInfo(value);
      break;
    case 12:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setMultiValued(value);
      break;
    case 13:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setAggregateAttributeValues(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.AddProtocolMapperRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getAttributeName();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getClaimName();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getClaimType();
  if (f !== 0.0) {
    writer.writeEnum(
      4,
      f
    );
  }
  f = message.getTenantId();
  if (f !== 0) {
    writer.writeInt64(
      6,
      f
    );
  }
  f = message.getClientId();
  if (f.length > 0) {
    writer.writeString(
      7,
      f
    );
  }
  f = message.getMapperType();
  if (f !== 0.0) {
    writer.writeEnum(
      8,
      f
    );
  }
  f = message.getAddToIdToken();
  if (f) {
    writer.writeBool(
      9,
      f
    );
  }
  f = message.getAddToAccessToken();
  if (f) {
    writer.writeBool(
      10,
      f
    );
  }
  f = message.getAddToUserInfo();
  if (f) {
    writer.writeBool(
      11,
      f
    );
  }
  f = message.getMultiValued();
  if (f) {
    writer.writeBool(
      12,
      f
    );
  }
  f = message.getAggregateAttributeValues();
  if (f) {
    writer.writeBool(
      13,
      f
    );
  }
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest} returns this
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string attribute_name = 2;
 * @return {string}
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.getAttributeName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest} returns this
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.setAttributeName = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string claim_name = 3;
 * @return {string}
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.getClaimName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest} returns this
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.setClaimName = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional ClaimJSONTypes claim_type = 4;
 * @return {!proto.org.apache.custos.iam.service.ClaimJSONTypes}
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.getClaimType = function() {
  return /** @type {!proto.org.apache.custos.iam.service.ClaimJSONTypes} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {!proto.org.apache.custos.iam.service.ClaimJSONTypes} value
 * @return {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest} returns this
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.setClaimType = function(value) {
  return jspb.Message.setProto3EnumField(this, 4, value);
};


/**
 * optional int64 tenant_id = 6;
 * @return {number}
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.getTenantId = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 6, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest} returns this
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.setTenantId = function(value) {
  return jspb.Message.setProto3IntField(this, 6, value);
};


/**
 * optional string client_id = 7;
 * @return {string}
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.getClientId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 7, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest} returns this
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.setClientId = function(value) {
  return jspb.Message.setProto3StringField(this, 7, value);
};


/**
 * optional MapperTypes mapper_type = 8;
 * @return {!proto.org.apache.custos.iam.service.MapperTypes}
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.getMapperType = function() {
  return /** @type {!proto.org.apache.custos.iam.service.MapperTypes} */ (jspb.Message.getFieldWithDefault(this, 8, 0));
};


/**
 * @param {!proto.org.apache.custos.iam.service.MapperTypes} value
 * @return {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest} returns this
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.setMapperType = function(value) {
  return jspb.Message.setProto3EnumField(this, 8, value);
};


/**
 * optional bool add_to_id_token = 9;
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.getAddToIdToken = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 9, false));
};


/**
 * @param {boolean} value
 * @return {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest} returns this
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.setAddToIdToken = function(value) {
  return jspb.Message.setProto3BooleanField(this, 9, value);
};


/**
 * optional bool add_to_access_token = 10;
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.getAddToAccessToken = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 10, false));
};


/**
 * @param {boolean} value
 * @return {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest} returns this
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.setAddToAccessToken = function(value) {
  return jspb.Message.setProto3BooleanField(this, 10, value);
};


/**
 * optional bool add_to_user_info = 11;
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.getAddToUserInfo = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 11, false));
};


/**
 * @param {boolean} value
 * @return {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest} returns this
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.setAddToUserInfo = function(value) {
  return jspb.Message.setProto3BooleanField(this, 11, value);
};


/**
 * optional bool multi_valued = 12;
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.getMultiValued = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 12, false));
};


/**
 * @param {boolean} value
 * @return {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest} returns this
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.setMultiValued = function(value) {
  return jspb.Message.setProto3BooleanField(this, 12, value);
};


/**
 * optional bool aggregate_attribute_values = 13;
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.getAggregateAttributeValues = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 13, false));
};


/**
 * @param {boolean} value
 * @return {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest} returns this
 */
proto.org.apache.custos.iam.service.AddProtocolMapperRequest.prototype.setAggregateAttributeValues = function(value) {
  return jspb.Message.setProto3BooleanField(this, 13, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.OperationStatus.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.OperationStatus.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.OperationStatus} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.OperationStatus.toObject = function(includeInstance, msg) {
  var f, obj = {
    status: jspb.Message.getBooleanFieldWithDefault(msg, 1, false)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.OperationStatus}
 */
proto.org.apache.custos.iam.service.OperationStatus.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.OperationStatus;
  return proto.org.apache.custos.iam.service.OperationStatus.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.OperationStatus} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.OperationStatus}
 */
proto.org.apache.custos.iam.service.OperationStatus.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setStatus(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.OperationStatus.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.OperationStatus.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.OperationStatus} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.OperationStatus.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getStatus();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
};


/**
 * optional bool status = 1;
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.OperationStatus.prototype.getStatus = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.org.apache.custos.iam.service.OperationStatus} returns this
 */
proto.org.apache.custos.iam.service.OperationStatus.prototype.setStatus = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.repeatedFields_ = [1,2,7];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.AddUserAttributesRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.AddUserAttributesRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    attributesList: jspb.Message.toObjectList(msg.getAttributesList(),
    proto.org.apache.custos.iam.service.UserAttribute.toObject, includeInstance),
    usersList: (f = jspb.Message.getRepeatedField(msg, 2)) == null ? undefined : f,
    tenantId: jspb.Message.getFieldWithDefault(msg, 3, 0),
    clientId: jspb.Message.getFieldWithDefault(msg, 4, ""),
    accessToken: jspb.Message.getFieldWithDefault(msg, 5, ""),
    performedby: jspb.Message.getFieldWithDefault(msg, 6, ""),
    agentsList: (f = jspb.Message.getRepeatedField(msg, 7)) == null ? undefined : f
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.AddUserAttributesRequest}
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.AddUserAttributesRequest;
  return proto.org.apache.custos.iam.service.AddUserAttributesRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.AddUserAttributesRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.AddUserAttributesRequest}
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.org.apache.custos.iam.service.UserAttribute;
      reader.readMessage(value,proto.org.apache.custos.iam.service.UserAttribute.deserializeBinaryFromReader);
      msg.addAttributes(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.addUsers(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantId(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientId(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setAccessToken(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setPerformedby(value);
      break;
    case 7:
      var value = /** @type {string} */ (reader.readString());
      msg.addAgents(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.AddUserAttributesRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.AddUserAttributesRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getAttributesList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.org.apache.custos.iam.service.UserAttribute.serializeBinaryToWriter
    );
  }
  f = message.getUsersList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      2,
      f
    );
  }
  f = message.getTenantId();
  if (f !== 0) {
    writer.writeInt64(
      3,
      f
    );
  }
  f = message.getClientId();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getAccessToken();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getPerformedby();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getAgentsList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      7,
      f
    );
  }
};


/**
 * repeated UserAttribute attributes = 1;
 * @return {!Array<!proto.org.apache.custos.iam.service.UserAttribute>}
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.getAttributesList = function() {
  return /** @type{!Array<!proto.org.apache.custos.iam.service.UserAttribute>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.org.apache.custos.iam.service.UserAttribute, 1));
};


/**
 * @param {!Array<!proto.org.apache.custos.iam.service.UserAttribute>} value
 * @return {!proto.org.apache.custos.iam.service.AddUserAttributesRequest} returns this
*/
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.setAttributesList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.org.apache.custos.iam.service.UserAttribute=} opt_value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.UserAttribute}
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.addAttributes = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.org.apache.custos.iam.service.UserAttribute, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.AddUserAttributesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.clearAttributesList = function() {
  return this.setAttributesList([]);
};


/**
 * repeated string users = 2;
 * @return {!Array<string>}
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.getUsersList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 2));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.org.apache.custos.iam.service.AddUserAttributesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.setUsersList = function(value) {
  return jspb.Message.setField(this, 2, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.AddUserAttributesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.addUsers = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 2, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.AddUserAttributesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.clearUsersList = function() {
  return this.setUsersList([]);
};


/**
 * optional int64 tenant_id = 3;
 * @return {number}
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.getTenantId = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.AddUserAttributesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.setTenantId = function(value) {
  return jspb.Message.setProto3IntField(this, 3, value);
};


/**
 * optional string client_id = 4;
 * @return {string}
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.getClientId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.AddUserAttributesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.setClientId = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string access_token = 5;
 * @return {string}
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.getAccessToken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.AddUserAttributesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.setAccessToken = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional string performedBy = 6;
 * @return {string}
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.getPerformedby = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.AddUserAttributesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.setPerformedby = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * repeated string agents = 7;
 * @return {!Array<string>}
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.getAgentsList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 7));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.org.apache.custos.iam.service.AddUserAttributesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.setAgentsList = function(value) {
  return jspb.Message.setField(this, 7, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.AddUserAttributesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.addAgents = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 7, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.AddUserAttributesRequest} returns this
 */
proto.org.apache.custos.iam.service.AddUserAttributesRequest.prototype.clearAgentsList = function() {
  return this.setAgentsList([]);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.repeatedFields_ = [1,2,7];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    attributesList: jspb.Message.toObjectList(msg.getAttributesList(),
    proto.org.apache.custos.iam.service.UserAttribute.toObject, includeInstance),
    usersList: (f = jspb.Message.getRepeatedField(msg, 2)) == null ? undefined : f,
    tenantId: jspb.Message.getFieldWithDefault(msg, 3, 0),
    clientId: jspb.Message.getFieldWithDefault(msg, 4, ""),
    accessToken: jspb.Message.getFieldWithDefault(msg, 5, ""),
    performedby: jspb.Message.getFieldWithDefault(msg, 6, ""),
    agentsList: (f = jspb.Message.getRepeatedField(msg, 7)) == null ? undefined : f
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest}
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.DeleteUserAttributeRequest;
  return proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest}
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.org.apache.custos.iam.service.UserAttribute;
      reader.readMessage(value,proto.org.apache.custos.iam.service.UserAttribute.deserializeBinaryFromReader);
      msg.addAttributes(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.addUsers(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantId(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientId(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setAccessToken(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setPerformedby(value);
      break;
    case 7:
      var value = /** @type {string} */ (reader.readString());
      msg.addAgents(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getAttributesList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.org.apache.custos.iam.service.UserAttribute.serializeBinaryToWriter
    );
  }
  f = message.getUsersList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      2,
      f
    );
  }
  f = message.getTenantId();
  if (f !== 0) {
    writer.writeInt64(
      3,
      f
    );
  }
  f = message.getClientId();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getAccessToken();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getPerformedby();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getAgentsList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      7,
      f
    );
  }
};


/**
 * repeated UserAttribute attributes = 1;
 * @return {!Array<!proto.org.apache.custos.iam.service.UserAttribute>}
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.getAttributesList = function() {
  return /** @type{!Array<!proto.org.apache.custos.iam.service.UserAttribute>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.org.apache.custos.iam.service.UserAttribute, 1));
};


/**
 * @param {!Array<!proto.org.apache.custos.iam.service.UserAttribute>} value
 * @return {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest} returns this
*/
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.setAttributesList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.org.apache.custos.iam.service.UserAttribute=} opt_value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.UserAttribute}
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.addAttributes = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.org.apache.custos.iam.service.UserAttribute, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.clearAttributesList = function() {
  return this.setAttributesList([]);
};


/**
 * repeated string users = 2;
 * @return {!Array<string>}
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.getUsersList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 2));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.setUsersList = function(value) {
  return jspb.Message.setField(this, 2, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.addUsers = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 2, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.clearUsersList = function() {
  return this.setUsersList([]);
};


/**
 * optional int64 tenant_id = 3;
 * @return {number}
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.getTenantId = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.setTenantId = function(value) {
  return jspb.Message.setProto3IntField(this, 3, value);
};


/**
 * optional string client_id = 4;
 * @return {string}
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.getClientId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.setClientId = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string access_token = 5;
 * @return {string}
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.getAccessToken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.setAccessToken = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional string performedBy = 6;
 * @return {string}
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.getPerformedby = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.setPerformedby = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * repeated string agents = 7;
 * @return {!Array<string>}
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.getAgentsList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 7));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.setAgentsList = function(value) {
  return jspb.Message.setField(this, 7, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.addAgents = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 7, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest} returns this
 */
proto.org.apache.custos.iam.service.DeleteUserAttributeRequest.prototype.clearAgentsList = function() {
  return this.setAgentsList([]);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.org.apache.custos.iam.service.UserAttribute.repeatedFields_ = [2];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.UserAttribute.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.UserAttribute.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.UserAttribute} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.UserAttribute.toObject = function(includeInstance, msg) {
  var f, obj = {
    key: jspb.Message.getFieldWithDefault(msg, 1, ""),
    valuesList: (f = jspb.Message.getRepeatedField(msg, 2)) == null ? undefined : f
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.UserAttribute}
 */
proto.org.apache.custos.iam.service.UserAttribute.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.UserAttribute;
  return proto.org.apache.custos.iam.service.UserAttribute.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.UserAttribute} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.UserAttribute}
 */
proto.org.apache.custos.iam.service.UserAttribute.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setKey(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.addValues(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.UserAttribute.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.UserAttribute.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.UserAttribute} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.UserAttribute.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getKey();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getValuesList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      2,
      f
    );
  }
};


/**
 * optional string key = 1;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserAttribute.prototype.getKey = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserAttribute} returns this
 */
proto.org.apache.custos.iam.service.UserAttribute.prototype.setKey = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * repeated string values = 2;
 * @return {!Array<string>}
 */
proto.org.apache.custos.iam.service.UserAttribute.prototype.getValuesList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 2));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.org.apache.custos.iam.service.UserAttribute} returns this
 */
proto.org.apache.custos.iam.service.UserAttribute.prototype.setValuesList = function(value) {
  return jspb.Message.setField(this, 2, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.UserAttribute} returns this
 */
proto.org.apache.custos.iam.service.UserAttribute.prototype.addValues = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 2, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.UserAttribute} returns this
 */
proto.org.apache.custos.iam.service.UserAttribute.prototype.clearValuesList = function() {
  return this.setValuesList([]);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.EventPersistenceRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.EventPersistenceRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.EventPersistenceRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.EventPersistenceRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    tenantid: jspb.Message.getFieldWithDefault(msg, 1, 0),
    adminEvent: jspb.Message.getBooleanFieldWithDefault(msg, 2, false),
    event: jspb.Message.getFieldWithDefault(msg, 3, ""),
    enable: jspb.Message.getBooleanFieldWithDefault(msg, 4, false),
    persistenceTime: jspb.Message.getFieldWithDefault(msg, 5, 0),
    performedby: jspb.Message.getFieldWithDefault(msg, 6, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.EventPersistenceRequest}
 */
proto.org.apache.custos.iam.service.EventPersistenceRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.EventPersistenceRequest;
  return proto.org.apache.custos.iam.service.EventPersistenceRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.EventPersistenceRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.EventPersistenceRequest}
 */
proto.org.apache.custos.iam.service.EventPersistenceRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantid(value);
      break;
    case 2:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setAdminEvent(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setEvent(value);
      break;
    case 4:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setEnable(value);
      break;
    case 5:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setPersistenceTime(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setPerformedby(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.EventPersistenceRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.EventPersistenceRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.EventPersistenceRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.EventPersistenceRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTenantid();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
  f = message.getAdminEvent();
  if (f) {
    writer.writeBool(
      2,
      f
    );
  }
  f = message.getEvent();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getEnable();
  if (f) {
    writer.writeBool(
      4,
      f
    );
  }
  f = message.getPersistenceTime();
  if (f !== 0) {
    writer.writeInt64(
      5,
      f
    );
  }
  f = message.getPerformedby();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
};


/**
 * optional int64 tenantId = 1;
 * @return {number}
 */
proto.org.apache.custos.iam.service.EventPersistenceRequest.prototype.getTenantid = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.EventPersistenceRequest} returns this
 */
proto.org.apache.custos.iam.service.EventPersistenceRequest.prototype.setTenantid = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional bool admin_event = 2;
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.EventPersistenceRequest.prototype.getAdminEvent = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.org.apache.custos.iam.service.EventPersistenceRequest} returns this
 */
proto.org.apache.custos.iam.service.EventPersistenceRequest.prototype.setAdminEvent = function(value) {
  return jspb.Message.setProto3BooleanField(this, 2, value);
};


/**
 * optional string event = 3;
 * @return {string}
 */
proto.org.apache.custos.iam.service.EventPersistenceRequest.prototype.getEvent = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.EventPersistenceRequest} returns this
 */
proto.org.apache.custos.iam.service.EventPersistenceRequest.prototype.setEvent = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional bool enable = 4;
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.EventPersistenceRequest.prototype.getEnable = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 4, false));
};


/**
 * @param {boolean} value
 * @return {!proto.org.apache.custos.iam.service.EventPersistenceRequest} returns this
 */
proto.org.apache.custos.iam.service.EventPersistenceRequest.prototype.setEnable = function(value) {
  return jspb.Message.setProto3BooleanField(this, 4, value);
};


/**
 * optional int64 persistence_time = 5;
 * @return {number}
 */
proto.org.apache.custos.iam.service.EventPersistenceRequest.prototype.getPersistenceTime = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.EventPersistenceRequest} returns this
 */
proto.org.apache.custos.iam.service.EventPersistenceRequest.prototype.setPersistenceTime = function(value) {
  return jspb.Message.setProto3IntField(this, 5, value);
};


/**
 * optional string performedBy = 6;
 * @return {string}
 */
proto.org.apache.custos.iam.service.EventPersistenceRequest.prototype.getPerformedby = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.EventPersistenceRequest} returns this
 */
proto.org.apache.custos.iam.service.EventPersistenceRequest.prototype.setPerformedby = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.org.apache.custos.iam.service.GroupsRequest.repeatedFields_ = [6];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.GroupsRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.GroupsRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.GroupsRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.GroupsRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    tenantid: jspb.Message.getFieldWithDefault(msg, 1, 0),
    accesstoken: jspb.Message.getFieldWithDefault(msg, 2, ""),
    performedby: jspb.Message.getFieldWithDefault(msg, 3, ""),
    clientid: jspb.Message.getFieldWithDefault(msg, 4, ""),
    clientsec: jspb.Message.getFieldWithDefault(msg, 5, ""),
    groupsList: jspb.Message.toObjectList(msg.getGroupsList(),
    proto.org.apache.custos.iam.service.GroupRepresentation.toObject, includeInstance)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.GroupsRequest}
 */
proto.org.apache.custos.iam.service.GroupsRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.GroupsRequest;
  return proto.org.apache.custos.iam.service.GroupsRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.GroupsRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.GroupsRequest}
 */
proto.org.apache.custos.iam.service.GroupsRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantid(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setAccesstoken(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setPerformedby(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientid(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientsec(value);
      break;
    case 6:
      var value = new proto.org.apache.custos.iam.service.GroupRepresentation;
      reader.readMessage(value,proto.org.apache.custos.iam.service.GroupRepresentation.deserializeBinaryFromReader);
      msg.addGroups(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.GroupsRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.GroupsRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.GroupsRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.GroupsRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTenantid();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
  f = message.getAccesstoken();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getPerformedby();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getClientid();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getClientsec();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getGroupsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      6,
      f,
      proto.org.apache.custos.iam.service.GroupRepresentation.serializeBinaryToWriter
    );
  }
};


/**
 * optional int64 tenantId = 1;
 * @return {number}
 */
proto.org.apache.custos.iam.service.GroupsRequest.prototype.getTenantid = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.GroupsRequest} returns this
 */
proto.org.apache.custos.iam.service.GroupsRequest.prototype.setTenantid = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional string accessToken = 2;
 * @return {string}
 */
proto.org.apache.custos.iam.service.GroupsRequest.prototype.getAccesstoken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.GroupsRequest} returns this
 */
proto.org.apache.custos.iam.service.GroupsRequest.prototype.setAccesstoken = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string performedBy = 3;
 * @return {string}
 */
proto.org.apache.custos.iam.service.GroupsRequest.prototype.getPerformedby = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.GroupsRequest} returns this
 */
proto.org.apache.custos.iam.service.GroupsRequest.prototype.setPerformedby = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string clientId = 4;
 * @return {string}
 */
proto.org.apache.custos.iam.service.GroupsRequest.prototype.getClientid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.GroupsRequest} returns this
 */
proto.org.apache.custos.iam.service.GroupsRequest.prototype.setClientid = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string clientSec = 5;
 * @return {string}
 */
proto.org.apache.custos.iam.service.GroupsRequest.prototype.getClientsec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.GroupsRequest} returns this
 */
proto.org.apache.custos.iam.service.GroupsRequest.prototype.setClientsec = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * repeated GroupRepresentation groups = 6;
 * @return {!Array<!proto.org.apache.custos.iam.service.GroupRepresentation>}
 */
proto.org.apache.custos.iam.service.GroupsRequest.prototype.getGroupsList = function() {
  return /** @type{!Array<!proto.org.apache.custos.iam.service.GroupRepresentation>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.org.apache.custos.iam.service.GroupRepresentation, 6));
};


/**
 * @param {!Array<!proto.org.apache.custos.iam.service.GroupRepresentation>} value
 * @return {!proto.org.apache.custos.iam.service.GroupsRequest} returns this
*/
proto.org.apache.custos.iam.service.GroupsRequest.prototype.setGroupsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 6, value);
};


/**
 * @param {!proto.org.apache.custos.iam.service.GroupRepresentation=} opt_value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.GroupRepresentation}
 */
proto.org.apache.custos.iam.service.GroupsRequest.prototype.addGroups = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 6, opt_value, proto.org.apache.custos.iam.service.GroupRepresentation, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.GroupsRequest} returns this
 */
proto.org.apache.custos.iam.service.GroupsRequest.prototype.clearGroupsList = function() {
  return this.setGroupsList([]);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.GroupRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.GroupRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.GroupRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.GroupRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    tenantid: jspb.Message.getFieldWithDefault(msg, 1, 0),
    accesstoken: jspb.Message.getFieldWithDefault(msg, 2, ""),
    performedby: jspb.Message.getFieldWithDefault(msg, 3, ""),
    clientid: jspb.Message.getFieldWithDefault(msg, 4, ""),
    clientsec: jspb.Message.getFieldWithDefault(msg, 5, ""),
    id: jspb.Message.getFieldWithDefault(msg, 6, ""),
    group: (f = msg.getGroup()) && proto.org.apache.custos.iam.service.GroupRepresentation.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.GroupRequest}
 */
proto.org.apache.custos.iam.service.GroupRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.GroupRequest;
  return proto.org.apache.custos.iam.service.GroupRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.GroupRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.GroupRequest}
 */
proto.org.apache.custos.iam.service.GroupRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantid(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setAccesstoken(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setPerformedby(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientid(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientsec(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 7:
      var value = new proto.org.apache.custos.iam.service.GroupRepresentation;
      reader.readMessage(value,proto.org.apache.custos.iam.service.GroupRepresentation.deserializeBinaryFromReader);
      msg.setGroup(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.GroupRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.GroupRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.GroupRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.GroupRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTenantid();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
  f = message.getAccesstoken();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getPerformedby();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getClientid();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getClientsec();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getGroup();
  if (f != null) {
    writer.writeMessage(
      7,
      f,
      proto.org.apache.custos.iam.service.GroupRepresentation.serializeBinaryToWriter
    );
  }
};


/**
 * optional int64 tenantId = 1;
 * @return {number}
 */
proto.org.apache.custos.iam.service.GroupRequest.prototype.getTenantid = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.GroupRequest} returns this
 */
proto.org.apache.custos.iam.service.GroupRequest.prototype.setTenantid = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional string accessToken = 2;
 * @return {string}
 */
proto.org.apache.custos.iam.service.GroupRequest.prototype.getAccesstoken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.GroupRequest} returns this
 */
proto.org.apache.custos.iam.service.GroupRequest.prototype.setAccesstoken = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string performedBy = 3;
 * @return {string}
 */
proto.org.apache.custos.iam.service.GroupRequest.prototype.getPerformedby = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.GroupRequest} returns this
 */
proto.org.apache.custos.iam.service.GroupRequest.prototype.setPerformedby = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string clientId = 4;
 * @return {string}
 */
proto.org.apache.custos.iam.service.GroupRequest.prototype.getClientid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.GroupRequest} returns this
 */
proto.org.apache.custos.iam.service.GroupRequest.prototype.setClientid = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string clientSec = 5;
 * @return {string}
 */
proto.org.apache.custos.iam.service.GroupRequest.prototype.getClientsec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.GroupRequest} returns this
 */
proto.org.apache.custos.iam.service.GroupRequest.prototype.setClientsec = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional string id = 6;
 * @return {string}
 */
proto.org.apache.custos.iam.service.GroupRequest.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.GroupRequest} returns this
 */
proto.org.apache.custos.iam.service.GroupRequest.prototype.setId = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * optional GroupRepresentation group = 7;
 * @return {?proto.org.apache.custos.iam.service.GroupRepresentation}
 */
proto.org.apache.custos.iam.service.GroupRequest.prototype.getGroup = function() {
  return /** @type{?proto.org.apache.custos.iam.service.GroupRepresentation} */ (
    jspb.Message.getWrapperField(this, proto.org.apache.custos.iam.service.GroupRepresentation, 7));
};


/**
 * @param {?proto.org.apache.custos.iam.service.GroupRepresentation|undefined} value
 * @return {!proto.org.apache.custos.iam.service.GroupRequest} returns this
*/
proto.org.apache.custos.iam.service.GroupRequest.prototype.setGroup = function(value) {
  return jspb.Message.setWrapperField(this, 7, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.org.apache.custos.iam.service.GroupRequest} returns this
 */
proto.org.apache.custos.iam.service.GroupRequest.prototype.clearGroup = function() {
  return this.setGroup(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.GroupRequest.prototype.hasGroup = function() {
  return jspb.Message.getField(this, 7) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.org.apache.custos.iam.service.GroupsResponse.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.GroupsResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.GroupsResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.GroupsResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.GroupsResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    groupsList: jspb.Message.toObjectList(msg.getGroupsList(),
    proto.org.apache.custos.iam.service.GroupRepresentation.toObject, includeInstance)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.GroupsResponse}
 */
proto.org.apache.custos.iam.service.GroupsResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.GroupsResponse;
  return proto.org.apache.custos.iam.service.GroupsResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.GroupsResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.GroupsResponse}
 */
proto.org.apache.custos.iam.service.GroupsResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.org.apache.custos.iam.service.GroupRepresentation;
      reader.readMessage(value,proto.org.apache.custos.iam.service.GroupRepresentation.deserializeBinaryFromReader);
      msg.addGroups(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.GroupsResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.GroupsResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.GroupsResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.GroupsResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getGroupsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.org.apache.custos.iam.service.GroupRepresentation.serializeBinaryToWriter
    );
  }
};


/**
 * repeated GroupRepresentation groups = 1;
 * @return {!Array<!proto.org.apache.custos.iam.service.GroupRepresentation>}
 */
proto.org.apache.custos.iam.service.GroupsResponse.prototype.getGroupsList = function() {
  return /** @type{!Array<!proto.org.apache.custos.iam.service.GroupRepresentation>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.org.apache.custos.iam.service.GroupRepresentation, 1));
};


/**
 * @param {!Array<!proto.org.apache.custos.iam.service.GroupRepresentation>} value
 * @return {!proto.org.apache.custos.iam.service.GroupsResponse} returns this
*/
proto.org.apache.custos.iam.service.GroupsResponse.prototype.setGroupsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.org.apache.custos.iam.service.GroupRepresentation=} opt_value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.GroupRepresentation}
 */
proto.org.apache.custos.iam.service.GroupsResponse.prototype.addGroups = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.org.apache.custos.iam.service.GroupRepresentation, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.GroupsResponse} returns this
 */
proto.org.apache.custos.iam.service.GroupsResponse.prototype.clearGroupsList = function() {
  return this.setGroupsList([]);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.UserGroupMappingRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.UserGroupMappingRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    tenantid: jspb.Message.getFieldWithDefault(msg, 1, 0),
    accesstoken: jspb.Message.getFieldWithDefault(msg, 2, ""),
    performedby: jspb.Message.getFieldWithDefault(msg, 3, ""),
    clientid: jspb.Message.getFieldWithDefault(msg, 4, ""),
    clientsec: jspb.Message.getFieldWithDefault(msg, 5, ""),
    username: jspb.Message.getFieldWithDefault(msg, 6, ""),
    groupId: jspb.Message.getFieldWithDefault(msg, 7, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.UserGroupMappingRequest}
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.UserGroupMappingRequest;
  return proto.org.apache.custos.iam.service.UserGroupMappingRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.UserGroupMappingRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.UserGroupMappingRequest}
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantid(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setAccesstoken(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setPerformedby(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientid(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientsec(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setUsername(value);
      break;
    case 7:
      var value = /** @type {string} */ (reader.readString());
      msg.setGroupId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.UserGroupMappingRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.UserGroupMappingRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTenantid();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
  f = message.getAccesstoken();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getPerformedby();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getClientid();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getClientsec();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getUsername();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getGroupId();
  if (f.length > 0) {
    writer.writeString(
      7,
      f
    );
  }
};


/**
 * optional int64 tenantId = 1;
 * @return {number}
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest.prototype.getTenantid = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.UserGroupMappingRequest} returns this
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest.prototype.setTenantid = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional string accessToken = 2;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest.prototype.getAccesstoken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserGroupMappingRequest} returns this
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest.prototype.setAccesstoken = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string performedBy = 3;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest.prototype.getPerformedby = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserGroupMappingRequest} returns this
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest.prototype.setPerformedby = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string clientId = 4;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest.prototype.getClientid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserGroupMappingRequest} returns this
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest.prototype.setClientid = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string clientSec = 5;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest.prototype.getClientsec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserGroupMappingRequest} returns this
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest.prototype.setClientsec = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional string username = 6;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest.prototype.getUsername = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserGroupMappingRequest} returns this
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest.prototype.setUsername = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * optional string group_id = 7;
 * @return {string}
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest.prototype.getGroupId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 7, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.UserGroupMappingRequest} returns this
 */
proto.org.apache.custos.iam.service.UserGroupMappingRequest.prototype.setGroupId = function(value) {
  return jspb.Message.setProto3StringField(this, 7, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.repeatedFields_ = [3];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.AgentClientMetadata.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.AgentClientMetadata} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.toObject = function(includeInstance, msg) {
  var f, obj = {
    tenantid: jspb.Message.getFieldWithDefault(msg, 1, 0),
    tenanturl: jspb.Message.getFieldWithDefault(msg, 2, ""),
    redirecturisList: (f = jspb.Message.getRepeatedField(msg, 3)) == null ? undefined : f,
    clientname: jspb.Message.getFieldWithDefault(msg, 4, ""),
    accessTokenLifeTime: jspb.Message.getFieldWithDefault(msg, 5, 0),
    performedby: jspb.Message.getFieldWithDefault(msg, 6, ""),
    accessToken: jspb.Message.getFieldWithDefault(msg, 7, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.AgentClientMetadata}
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.AgentClientMetadata;
  return proto.org.apache.custos.iam.service.AgentClientMetadata.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.AgentClientMetadata} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.AgentClientMetadata}
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantid(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setTenanturl(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.addRedirecturis(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientname(value);
      break;
    case 5:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setAccessTokenLifeTime(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setPerformedby(value);
      break;
    case 7:
      var value = /** @type {string} */ (reader.readString());
      msg.setAccessToken(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.AgentClientMetadata.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.AgentClientMetadata} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTenantid();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
  f = message.getTenanturl();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getRedirecturisList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      3,
      f
    );
  }
  f = message.getClientname();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getAccessTokenLifeTime();
  if (f !== 0) {
    writer.writeInt64(
      5,
      f
    );
  }
  f = message.getPerformedby();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getAccessToken();
  if (f.length > 0) {
    writer.writeString(
      7,
      f
    );
  }
};


/**
 * optional int64 tenantId = 1;
 * @return {number}
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.prototype.getTenantid = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.AgentClientMetadata} returns this
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.prototype.setTenantid = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional string tenantURL = 2;
 * @return {string}
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.prototype.getTenanturl = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.AgentClientMetadata} returns this
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.prototype.setTenanturl = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * repeated string redirectURIs = 3;
 * @return {!Array<string>}
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.prototype.getRedirecturisList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 3));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.org.apache.custos.iam.service.AgentClientMetadata} returns this
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.prototype.setRedirecturisList = function(value) {
  return jspb.Message.setField(this, 3, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.AgentClientMetadata} returns this
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.prototype.addRedirecturis = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 3, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.AgentClientMetadata} returns this
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.prototype.clearRedirecturisList = function() {
  return this.setRedirecturisList([]);
};


/**
 * optional string clientName = 4;
 * @return {string}
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.prototype.getClientname = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.AgentClientMetadata} returns this
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.prototype.setClientname = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional int64 access_token_life_time = 5;
 * @return {number}
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.prototype.getAccessTokenLifeTime = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.AgentClientMetadata} returns this
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.prototype.setAccessTokenLifeTime = function(value) {
  return jspb.Message.setProto3IntField(this, 5, value);
};


/**
 * optional string performedBy = 6;
 * @return {string}
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.prototype.getPerformedby = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.AgentClientMetadata} returns this
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.prototype.setPerformedby = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * optional string access_token = 7;
 * @return {string}
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.prototype.getAccessToken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 7, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.AgentClientMetadata} returns this
 */
proto.org.apache.custos.iam.service.AgentClientMetadata.prototype.setAccessToken = function(value) {
  return jspb.Message.setProto3StringField(this, 7, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.org.apache.custos.iam.service.Agent.repeatedFields_ = [2,3];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.Agent.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.Agent.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.Agent} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.Agent.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, ""),
    realmRolesList: (f = jspb.Message.getRepeatedField(msg, 2)) == null ? undefined : f,
    attributesList: jspb.Message.toObjectList(msg.getAttributesList(),
    proto.org.apache.custos.iam.service.UserAttribute.toObject, includeInstance),
    isenabled: jspb.Message.getBooleanFieldWithDefault(msg, 4, false),
    creationTime: jspb.Message.getFloatingPointFieldWithDefault(msg, 5, 0.0),
    lastModifiedAt: jspb.Message.getFloatingPointFieldWithDefault(msg, 6, 0.0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.Agent}
 */
proto.org.apache.custos.iam.service.Agent.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.Agent;
  return proto.org.apache.custos.iam.service.Agent.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.Agent} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.Agent}
 */
proto.org.apache.custos.iam.service.Agent.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.addRealmRoles(value);
      break;
    case 3:
      var value = new proto.org.apache.custos.iam.service.UserAttribute;
      reader.readMessage(value,proto.org.apache.custos.iam.service.UserAttribute.deserializeBinaryFromReader);
      msg.addAttributes(value);
      break;
    case 4:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setIsenabled(value);
      break;
    case 5:
      var value = /** @type {number} */ (reader.readDouble());
      msg.setCreationTime(value);
      break;
    case 6:
      var value = /** @type {number} */ (reader.readDouble());
      msg.setLastModifiedAt(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.Agent.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.Agent.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.Agent} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.Agent.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getRealmRolesList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      2,
      f
    );
  }
  f = message.getAttributesList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      3,
      f,
      proto.org.apache.custos.iam.service.UserAttribute.serializeBinaryToWriter
    );
  }
  f = message.getIsenabled();
  if (f) {
    writer.writeBool(
      4,
      f
    );
  }
  f = message.getCreationTime();
  if (f !== 0.0) {
    writer.writeDouble(
      5,
      f
    );
  }
  f = message.getLastModifiedAt();
  if (f !== 0.0) {
    writer.writeDouble(
      6,
      f
    );
  }
};


/**
 * optional string id = 1;
 * @return {string}
 */
proto.org.apache.custos.iam.service.Agent.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.Agent} returns this
 */
proto.org.apache.custos.iam.service.Agent.prototype.setId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * repeated string realm_roles = 2;
 * @return {!Array<string>}
 */
proto.org.apache.custos.iam.service.Agent.prototype.getRealmRolesList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 2));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.org.apache.custos.iam.service.Agent} returns this
 */
proto.org.apache.custos.iam.service.Agent.prototype.setRealmRolesList = function(value) {
  return jspb.Message.setField(this, 2, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.Agent} returns this
 */
proto.org.apache.custos.iam.service.Agent.prototype.addRealmRoles = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 2, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.Agent} returns this
 */
proto.org.apache.custos.iam.service.Agent.prototype.clearRealmRolesList = function() {
  return this.setRealmRolesList([]);
};


/**
 * repeated UserAttribute attributes = 3;
 * @return {!Array<!proto.org.apache.custos.iam.service.UserAttribute>}
 */
proto.org.apache.custos.iam.service.Agent.prototype.getAttributesList = function() {
  return /** @type{!Array<!proto.org.apache.custos.iam.service.UserAttribute>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.org.apache.custos.iam.service.UserAttribute, 3));
};


/**
 * @param {!Array<!proto.org.apache.custos.iam.service.UserAttribute>} value
 * @return {!proto.org.apache.custos.iam.service.Agent} returns this
*/
proto.org.apache.custos.iam.service.Agent.prototype.setAttributesList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 3, value);
};


/**
 * @param {!proto.org.apache.custos.iam.service.UserAttribute=} opt_value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.UserAttribute}
 */
proto.org.apache.custos.iam.service.Agent.prototype.addAttributes = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 3, opt_value, proto.org.apache.custos.iam.service.UserAttribute, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.Agent} returns this
 */
proto.org.apache.custos.iam.service.Agent.prototype.clearAttributesList = function() {
  return this.setAttributesList([]);
};


/**
 * optional bool isEnabled = 4;
 * @return {boolean}
 */
proto.org.apache.custos.iam.service.Agent.prototype.getIsenabled = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 4, false));
};


/**
 * @param {boolean} value
 * @return {!proto.org.apache.custos.iam.service.Agent} returns this
 */
proto.org.apache.custos.iam.service.Agent.prototype.setIsenabled = function(value) {
  return jspb.Message.setProto3BooleanField(this, 4, value);
};


/**
 * optional double creation_time = 5;
 * @return {number}
 */
proto.org.apache.custos.iam.service.Agent.prototype.getCreationTime = function() {
  return /** @type {number} */ (jspb.Message.getFloatingPointFieldWithDefault(this, 5, 0.0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.Agent} returns this
 */
proto.org.apache.custos.iam.service.Agent.prototype.setCreationTime = function(value) {
  return jspb.Message.setProto3FloatField(this, 5, value);
};


/**
 * optional double last_modified_at = 6;
 * @return {number}
 */
proto.org.apache.custos.iam.service.Agent.prototype.getLastModifiedAt = function() {
  return /** @type {number} */ (jspb.Message.getFloatingPointFieldWithDefault(this, 6, 0.0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.Agent} returns this
 */
proto.org.apache.custos.iam.service.Agent.prototype.setLastModifiedAt = function(value) {
  return jspb.Message.setProto3FloatField(this, 6, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.GetAllResources.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.GetAllResources.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.GetAllResources} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.GetAllResources.toObject = function(includeInstance, msg) {
  var f, obj = {
    tenantid: jspb.Message.getFieldWithDefault(msg, 1, 0),
    clientid: jspb.Message.getFieldWithDefault(msg, 2, ""),
    resourceType: jspb.Message.getFieldWithDefault(msg, 3, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.GetAllResources}
 */
proto.org.apache.custos.iam.service.GetAllResources.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.GetAllResources;
  return proto.org.apache.custos.iam.service.GetAllResources.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.GetAllResources} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.GetAllResources}
 */
proto.org.apache.custos.iam.service.GetAllResources.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTenantid(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setClientid(value);
      break;
    case 3:
      var value = /** @type {!proto.org.apache.custos.iam.service.ResourceTypes} */ (reader.readEnum());
      msg.setResourceType(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.GetAllResources.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.GetAllResources.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.GetAllResources} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.GetAllResources.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTenantid();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
  f = message.getClientid();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getResourceType();
  if (f !== 0.0) {
    writer.writeEnum(
      3,
      f
    );
  }
};


/**
 * optional int64 tenantId = 1;
 * @return {number}
 */
proto.org.apache.custos.iam.service.GetAllResources.prototype.getTenantid = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.org.apache.custos.iam.service.GetAllResources} returns this
 */
proto.org.apache.custos.iam.service.GetAllResources.prototype.setTenantid = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional string clientId = 2;
 * @return {string}
 */
proto.org.apache.custos.iam.service.GetAllResources.prototype.getClientid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.org.apache.custos.iam.service.GetAllResources} returns this
 */
proto.org.apache.custos.iam.service.GetAllResources.prototype.setClientid = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional ResourceTypes resource_type = 3;
 * @return {!proto.org.apache.custos.iam.service.ResourceTypes}
 */
proto.org.apache.custos.iam.service.GetAllResources.prototype.getResourceType = function() {
  return /** @type {!proto.org.apache.custos.iam.service.ResourceTypes} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {!proto.org.apache.custos.iam.service.ResourceTypes} value
 * @return {!proto.org.apache.custos.iam.service.GetAllResources} returns this
 */
proto.org.apache.custos.iam.service.GetAllResources.prototype.setResourceType = function(value) {
  return jspb.Message.setProto3EnumField(this, 3, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.org.apache.custos.iam.service.GetAllResourcesResponse.repeatedFields_ = [1,2];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.org.apache.custos.iam.service.GetAllResourcesResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.org.apache.custos.iam.service.GetAllResourcesResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.org.apache.custos.iam.service.GetAllResourcesResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.GetAllResourcesResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    agentsList: jspb.Message.toObjectList(msg.getAgentsList(),
    proto.org.apache.custos.iam.service.Agent.toObject, includeInstance),
    usersList: jspb.Message.toObjectList(msg.getUsersList(),
    proto.org.apache.custos.iam.service.UserRepresentation.toObject, includeInstance)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.org.apache.custos.iam.service.GetAllResourcesResponse}
 */
proto.org.apache.custos.iam.service.GetAllResourcesResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.org.apache.custos.iam.service.GetAllResourcesResponse;
  return proto.org.apache.custos.iam.service.GetAllResourcesResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.org.apache.custos.iam.service.GetAllResourcesResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.org.apache.custos.iam.service.GetAllResourcesResponse}
 */
proto.org.apache.custos.iam.service.GetAllResourcesResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.org.apache.custos.iam.service.Agent;
      reader.readMessage(value,proto.org.apache.custos.iam.service.Agent.deserializeBinaryFromReader);
      msg.addAgents(value);
      break;
    case 2:
      var value = new proto.org.apache.custos.iam.service.UserRepresentation;
      reader.readMessage(value,proto.org.apache.custos.iam.service.UserRepresentation.deserializeBinaryFromReader);
      msg.addUsers(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.org.apache.custos.iam.service.GetAllResourcesResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.org.apache.custos.iam.service.GetAllResourcesResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.org.apache.custos.iam.service.GetAllResourcesResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.org.apache.custos.iam.service.GetAllResourcesResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getAgentsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.org.apache.custos.iam.service.Agent.serializeBinaryToWriter
    );
  }
  f = message.getUsersList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      2,
      f,
      proto.org.apache.custos.iam.service.UserRepresentation.serializeBinaryToWriter
    );
  }
};


/**
 * repeated Agent agents = 1;
 * @return {!Array<!proto.org.apache.custos.iam.service.Agent>}
 */
proto.org.apache.custos.iam.service.GetAllResourcesResponse.prototype.getAgentsList = function() {
  return /** @type{!Array<!proto.org.apache.custos.iam.service.Agent>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.org.apache.custos.iam.service.Agent, 1));
};


/**
 * @param {!Array<!proto.org.apache.custos.iam.service.Agent>} value
 * @return {!proto.org.apache.custos.iam.service.GetAllResourcesResponse} returns this
*/
proto.org.apache.custos.iam.service.GetAllResourcesResponse.prototype.setAgentsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.org.apache.custos.iam.service.Agent=} opt_value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.Agent}
 */
proto.org.apache.custos.iam.service.GetAllResourcesResponse.prototype.addAgents = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.org.apache.custos.iam.service.Agent, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.GetAllResourcesResponse} returns this
 */
proto.org.apache.custos.iam.service.GetAllResourcesResponse.prototype.clearAgentsList = function() {
  return this.setAgentsList([]);
};


/**
 * repeated UserRepresentation users = 2;
 * @return {!Array<!proto.org.apache.custos.iam.service.UserRepresentation>}
 */
proto.org.apache.custos.iam.service.GetAllResourcesResponse.prototype.getUsersList = function() {
  return /** @type{!Array<!proto.org.apache.custos.iam.service.UserRepresentation>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.org.apache.custos.iam.service.UserRepresentation, 2));
};


/**
 * @param {!Array<!proto.org.apache.custos.iam.service.UserRepresentation>} value
 * @return {!proto.org.apache.custos.iam.service.GetAllResourcesResponse} returns this
*/
proto.org.apache.custos.iam.service.GetAllResourcesResponse.prototype.setUsersList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 2, value);
};


/**
 * @param {!proto.org.apache.custos.iam.service.UserRepresentation=} opt_value
 * @param {number=} opt_index
 * @return {!proto.org.apache.custos.iam.service.UserRepresentation}
 */
proto.org.apache.custos.iam.service.GetAllResourcesResponse.prototype.addUsers = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 2, opt_value, proto.org.apache.custos.iam.service.UserRepresentation, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.org.apache.custos.iam.service.GetAllResourcesResponse} returns this
 */
proto.org.apache.custos.iam.service.GetAllResourcesResponse.prototype.clearUsersList = function() {
  return this.setUsersList([]);
};


/**
 * @enum {number}
 */
proto.org.apache.custos.iam.service.FederatedIDPs = {
  CILOGON: 0,
  FACEBOOK: 1,
  GOOGLE: 2,
  LINKEDIN: 3,
  TWITTER: 4,
  CUSTOM_OIDC: 5
};

/**
 * @enum {number}
 */
proto.org.apache.custos.iam.service.MapperTypes = {
  USER_ATTRIBUTE: 0,
  USER_REALM_ROLE: 1,
  USER_CLIENT_ROLE: 2
};

/**
 * @enum {number}
 */
proto.org.apache.custos.iam.service.ClaimJSONTypes = {
  STRING: 0,
  LONG: 1,
  INTEGER: 2,
  BOOLEAN: 3,
  JSON: 4
};

/**
 * @enum {number}
 */
proto.org.apache.custos.iam.service.ResourceTypes = {
  USER: 0,
  AGENT: 1
};

goog.object.extend(exports, proto.org.apache.custos.iam.service);
