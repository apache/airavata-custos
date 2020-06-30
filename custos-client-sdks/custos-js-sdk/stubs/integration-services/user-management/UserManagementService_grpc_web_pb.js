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

/**
 * @fileoverview gRPC-Web generated client stub for org.apache.custos.user.management.service
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');


var UserProfileService_pb = require('../../core-services/user-profile/UserProfileService_pb.js')

var IamAdminService_pb = require('../../core-services/iam-admin-service/IamAdminService_pb.js')
const proto = {};
proto.org = {};
proto.org.apache = {};
proto.org.apache.custos = {};
proto.org.apache.custos.user = {};
proto.org.apache.custos.user.management = {};
proto.org.apache.custos.user.management.service = require('./UserManagementService_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?Object} options
 * @constructor
 * @struct
 * @final
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options['format'] = 'text';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?Object} options
 * @constructor
 * @struct
 * @final
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options['format'] = 'text';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.RegisterUserRequest,
 *   !proto.org.apache.custos.iam.service.RegisterUserResponse>}
 */
const methodDescriptor_UserManagementService_registerUser = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/registerUser',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.RegisterUserRequest,
  IamAdminService_pb.RegisterUserResponse,
  /**
   * @param {!proto.org.apache.custos.iam.service.RegisterUserRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.RegisterUserResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.RegisterUserRequest,
 *   !proto.org.apache.custos.iam.service.RegisterUserResponse>}
 */
const methodInfo_UserManagementService_registerUser = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.RegisterUserResponse,
  /**
   * @param {!proto.org.apache.custos.iam.service.RegisterUserRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.RegisterUserResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.RegisterUserRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.RegisterUserResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.RegisterUserResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.registerUser =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/registerUser',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_registerUser,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.RegisterUserRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.RegisterUserResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.registerUser =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/registerUser',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_registerUser);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.RegisterUsersRequest,
 *   !proto.org.apache.custos.iam.service.RegisterUsersResponse>}
 */
const methodDescriptor_UserManagementService_registerAndEnableUsers = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/registerAndEnableUsers',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.RegisterUsersRequest,
  IamAdminService_pb.RegisterUsersResponse,
  /**
   * @param {!proto.org.apache.custos.iam.service.RegisterUsersRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.RegisterUsersResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.RegisterUsersRequest,
 *   !proto.org.apache.custos.iam.service.RegisterUsersResponse>}
 */
const methodInfo_UserManagementService_registerAndEnableUsers = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.RegisterUsersResponse,
  /**
   * @param {!proto.org.apache.custos.iam.service.RegisterUsersRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.RegisterUsersResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.RegisterUsersRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.RegisterUsersResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.RegisterUsersResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.registerAndEnableUsers =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/registerAndEnableUsers',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_registerAndEnableUsers,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.RegisterUsersRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.RegisterUsersResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.registerAndEnableUsers =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/registerAndEnableUsers',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_registerAndEnableUsers);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.AddUserAttributesRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_UserManagementService_addUserAttributes = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/addUserAttributes',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.AddUserAttributesRequest,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.AddUserAttributesRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.AddUserAttributesRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_UserManagementService_addUserAttributes = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.AddUserAttributesRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.AddUserAttributesRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.addUserAttributes =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/addUserAttributes',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_addUserAttributes,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.AddUserAttributesRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.addUserAttributes =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/addUserAttributes',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_addUserAttributes);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.DeleteUserAttributeRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_UserManagementService_deleteUserAttributes = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/deleteUserAttributes',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.DeleteUserAttributeRequest,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.DeleteUserAttributeRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_UserManagementService_deleteUserAttributes = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.deleteUserAttributes =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/deleteUserAttributes',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_deleteUserAttributes,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.DeleteUserAttributeRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.deleteUserAttributes =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/deleteUserAttributes',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_deleteUserAttributes);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.UserSearchRequest,
 *   !proto.org.apache.custos.iam.service.UserRepresentation>}
 */
const methodDescriptor_UserManagementService_enableUser = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/enableUser',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.UserSearchRequest,
  IamAdminService_pb.UserRepresentation,
  /**
   * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.UserRepresentation.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.UserSearchRequest,
 *   !proto.org.apache.custos.iam.service.UserRepresentation>}
 */
const methodInfo_UserManagementService_enableUser = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.UserRepresentation,
  /**
   * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.UserRepresentation.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.UserRepresentation)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.UserRepresentation>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.enableUser =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/enableUser',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_enableUser,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.UserRepresentation>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.enableUser =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/enableUser',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_enableUser);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.UserSearchRequest,
 *   !proto.org.apache.custos.iam.service.UserRepresentation>}
 */
const methodDescriptor_UserManagementService_disableUser = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/disableUser',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.UserSearchRequest,
  IamAdminService_pb.UserRepresentation,
  /**
   * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.UserRepresentation.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.UserSearchRequest,
 *   !proto.org.apache.custos.iam.service.UserRepresentation>}
 */
const methodInfo_UserManagementService_disableUser = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.UserRepresentation,
  /**
   * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.UserRepresentation.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.UserRepresentation)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.UserRepresentation>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.disableUser =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/disableUser',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_disableUser,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.UserRepresentation>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.disableUser =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/disableUser',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_disableUser);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.UserSearchRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_UserManagementService_grantAdminPrivileges = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/grantAdminPrivileges',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.UserSearchRequest,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.UserSearchRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_UserManagementService_grantAdminPrivileges = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.grantAdminPrivileges =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/grantAdminPrivileges',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_grantAdminPrivileges,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.grantAdminPrivileges =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/grantAdminPrivileges',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_grantAdminPrivileges);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.UserSearchRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_UserManagementService_removeAdminPrivileges = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/removeAdminPrivileges',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.UserSearchRequest,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.UserSearchRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_UserManagementService_removeAdminPrivileges = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.removeAdminPrivileges =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/removeAdminPrivileges',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_removeAdminPrivileges,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.removeAdminPrivileges =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/removeAdminPrivileges',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_removeAdminPrivileges);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.AddUserRolesRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_UserManagementService_addRolesToUsers = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/addRolesToUsers',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.AddUserRolesRequest,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.AddUserRolesRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.AddUserRolesRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_UserManagementService_addRolesToUsers = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.AddUserRolesRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.AddUserRolesRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.addRolesToUsers =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/addRolesToUsers',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_addRolesToUsers,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.AddUserRolesRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.addRolesToUsers =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/addRolesToUsers',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_addRolesToUsers);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.UserSearchRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_UserManagementService_isUserEnabled = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/isUserEnabled',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.UserSearchRequest,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.UserSearchRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_UserManagementService_isUserEnabled = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.isUserEnabled =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/isUserEnabled',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_isUserEnabled,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.isUserEnabled =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/isUserEnabled',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_isUserEnabled);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.UserSearchRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_UserManagementService_isUsernameAvailable = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/isUsernameAvailable',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.UserSearchRequest,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.UserSearchRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_UserManagementService_isUsernameAvailable = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.isUsernameAvailable =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/isUsernameAvailable',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_isUsernameAvailable,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.isUsernameAvailable =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/isUsernameAvailable',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_isUsernameAvailable);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.UserSearchRequest,
 *   !proto.org.apache.custos.iam.service.UserRepresentation>}
 */
const methodDescriptor_UserManagementService_getUser = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/getUser',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.UserSearchRequest,
  IamAdminService_pb.UserRepresentation,
  /**
   * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.UserRepresentation.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.UserSearchRequest,
 *   !proto.org.apache.custos.iam.service.UserRepresentation>}
 */
const methodInfo_UserManagementService_getUser = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.UserRepresentation,
  /**
   * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.UserRepresentation.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.UserRepresentation)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.UserRepresentation>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.getUser =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/getUser',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_getUser,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.UserRepresentation>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.getUser =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/getUser',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_getUser);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.FindUsersRequest,
 *   !proto.org.apache.custos.iam.service.FindUsersResponse>}
 */
const methodDescriptor_UserManagementService_findUsers = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/findUsers',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.FindUsersRequest,
  IamAdminService_pb.FindUsersResponse,
  /**
   * @param {!proto.org.apache.custos.iam.service.FindUsersRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.FindUsersResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.FindUsersRequest,
 *   !proto.org.apache.custos.iam.service.FindUsersResponse>}
 */
const methodInfo_UserManagementService_findUsers = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.FindUsersResponse,
  /**
   * @param {!proto.org.apache.custos.iam.service.FindUsersRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.FindUsersResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.FindUsersRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.FindUsersResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.FindUsersResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.findUsers =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/findUsers',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_findUsers,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.FindUsersRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.FindUsersResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.findUsers =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/findUsers',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_findUsers);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.ResetUserPassword,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_UserManagementService_resetPassword = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/resetPassword',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.ResetUserPassword,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.ResetUserPassword} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.ResetUserPassword,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_UserManagementService_resetPassword = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.ResetUserPassword} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.ResetUserPassword} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.resetPassword =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/resetPassword',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_resetPassword,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.ResetUserPassword} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.resetPassword =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/resetPassword',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_resetPassword);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.UserSearchRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_UserManagementService_deleteUser = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/deleteUser',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.UserSearchRequest,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.UserSearchRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_UserManagementService_deleteUser = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.deleteUser =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/deleteUser',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_deleteUser,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.UserSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.deleteUser =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/deleteUser',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_deleteUser);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.DeleteUserRolesRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_UserManagementService_deleteUserRoles = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/deleteUserRoles',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.DeleteUserRolesRequest,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.DeleteUserRolesRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_UserManagementService_deleteUserRoles = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.deleteUserRoles =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/deleteUserRoles',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_deleteUserRoles,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.DeleteUserRolesRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.deleteUserRoles =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/deleteUserRoles',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_deleteUserRoles);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.user.management.service.UserProfileRequest,
 *   !proto.org.apache.custos.user.profile.service.UserProfile>}
 */
const methodDescriptor_UserManagementService_updateUserProfile = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/updateUserProfile',
  grpc.web.MethodType.UNARY,
  proto.org.apache.custos.user.management.service.UserProfileRequest,
  UserProfileService_pb.UserProfile,
  /**
   * @param {!proto.org.apache.custos.user.management.service.UserProfileRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  UserProfileService_pb.UserProfile.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.user.management.service.UserProfileRequest,
 *   !proto.org.apache.custos.user.profile.service.UserProfile>}
 */
const methodInfo_UserManagementService_updateUserProfile = new grpc.web.AbstractClientBase.MethodInfo(
  UserProfileService_pb.UserProfile,
  /**
   * @param {!proto.org.apache.custos.user.management.service.UserProfileRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  UserProfileService_pb.UserProfile.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.user.management.service.UserProfileRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.user.profile.service.UserProfile)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.user.profile.service.UserProfile>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.updateUserProfile =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/updateUserProfile',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_updateUserProfile,
      callback);
};


/**
 * @param {!proto.org.apache.custos.user.management.service.UserProfileRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.user.profile.service.UserProfile>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.updateUserProfile =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/updateUserProfile',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_updateUserProfile);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.user.management.service.UserProfileRequest,
 *   !proto.org.apache.custos.user.profile.service.UserProfile>}
 */
const methodDescriptor_UserManagementService_getUserProfile = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/getUserProfile',
  grpc.web.MethodType.UNARY,
  proto.org.apache.custos.user.management.service.UserProfileRequest,
  UserProfileService_pb.UserProfile,
  /**
   * @param {!proto.org.apache.custos.user.management.service.UserProfileRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  UserProfileService_pb.UserProfile.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.user.management.service.UserProfileRequest,
 *   !proto.org.apache.custos.user.profile.service.UserProfile>}
 */
const methodInfo_UserManagementService_getUserProfile = new grpc.web.AbstractClientBase.MethodInfo(
  UserProfileService_pb.UserProfile,
  /**
   * @param {!proto.org.apache.custos.user.management.service.UserProfileRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  UserProfileService_pb.UserProfile.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.user.management.service.UserProfileRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.user.profile.service.UserProfile)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.user.profile.service.UserProfile>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.getUserProfile =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/getUserProfile',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_getUserProfile,
      callback);
};


/**
 * @param {!proto.org.apache.custos.user.management.service.UserProfileRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.user.profile.service.UserProfile>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.getUserProfile =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/getUserProfile',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_getUserProfile);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.user.management.service.UserProfileRequest,
 *   !proto.org.apache.custos.user.profile.service.UserProfile>}
 */
const methodDescriptor_UserManagementService_deleteUserProfile = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/deleteUserProfile',
  grpc.web.MethodType.UNARY,
  proto.org.apache.custos.user.management.service.UserProfileRequest,
  UserProfileService_pb.UserProfile,
  /**
   * @param {!proto.org.apache.custos.user.management.service.UserProfileRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  UserProfileService_pb.UserProfile.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.user.management.service.UserProfileRequest,
 *   !proto.org.apache.custos.user.profile.service.UserProfile>}
 */
const methodInfo_UserManagementService_deleteUserProfile = new grpc.web.AbstractClientBase.MethodInfo(
  UserProfileService_pb.UserProfile,
  /**
   * @param {!proto.org.apache.custos.user.management.service.UserProfileRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  UserProfileService_pb.UserProfile.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.user.management.service.UserProfileRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.user.profile.service.UserProfile)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.user.profile.service.UserProfile>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.deleteUserProfile =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/deleteUserProfile',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_deleteUserProfile,
      callback);
};


/**
 * @param {!proto.org.apache.custos.user.management.service.UserProfileRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.user.profile.service.UserProfile>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.deleteUserProfile =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/deleteUserProfile',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_deleteUserProfile);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.user.management.service.UserProfileRequest,
 *   !proto.org.apache.custos.user.profile.service.GetAllUserProfilesResponse>}
 */
const methodDescriptor_UserManagementService_getAllUserProfilesInTenant = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/getAllUserProfilesInTenant',
  grpc.web.MethodType.UNARY,
  proto.org.apache.custos.user.management.service.UserProfileRequest,
  UserProfileService_pb.GetAllUserProfilesResponse,
  /**
   * @param {!proto.org.apache.custos.user.management.service.UserProfileRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  UserProfileService_pb.GetAllUserProfilesResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.user.management.service.UserProfileRequest,
 *   !proto.org.apache.custos.user.profile.service.GetAllUserProfilesResponse>}
 */
const methodInfo_UserManagementService_getAllUserProfilesInTenant = new grpc.web.AbstractClientBase.MethodInfo(
  UserProfileService_pb.GetAllUserProfilesResponse,
  /**
   * @param {!proto.org.apache.custos.user.management.service.UserProfileRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  UserProfileService_pb.GetAllUserProfilesResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.user.management.service.UserProfileRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.user.profile.service.GetAllUserProfilesResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.user.profile.service.GetAllUserProfilesResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.getAllUserProfilesInTenant =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/getAllUserProfilesInTenant',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_getAllUserProfilesInTenant,
      callback);
};


/**
 * @param {!proto.org.apache.custos.user.management.service.UserProfileRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.user.profile.service.GetAllUserProfilesResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.getAllUserProfilesInTenant =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/getAllUserProfilesInTenant',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_getAllUserProfilesInTenant);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.user.management.service.LinkUserProfileRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_UserManagementService_linkUserProfile = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/linkUserProfile',
  grpc.web.MethodType.UNARY,
  proto.org.apache.custos.user.management.service.LinkUserProfileRequest,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.user.management.service.LinkUserProfileRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.user.management.service.LinkUserProfileRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_UserManagementService_linkUserProfile = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.user.management.service.LinkUserProfileRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.user.management.service.LinkUserProfileRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.linkUserProfile =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/linkUserProfile',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_linkUserProfile,
      callback);
};


/**
 * @param {!proto.org.apache.custos.user.management.service.LinkUserProfileRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.linkUserProfile =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/linkUserProfile',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_linkUserProfile);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.user.profile.service.GetUpdateAuditTrailRequest,
 *   !proto.org.apache.custos.user.profile.service.GetUpdateAuditTrailResponse>}
 */
const methodDescriptor_UserManagementService_getUserProfileAuditTrails = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/getUserProfileAuditTrails',
  grpc.web.MethodType.UNARY,
  UserProfileService_pb.GetUpdateAuditTrailRequest,
  UserProfileService_pb.GetUpdateAuditTrailResponse,
  /**
   * @param {!proto.org.apache.custos.user.profile.service.GetUpdateAuditTrailRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  UserProfileService_pb.GetUpdateAuditTrailResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.user.profile.service.GetUpdateAuditTrailRequest,
 *   !proto.org.apache.custos.user.profile.service.GetUpdateAuditTrailResponse>}
 */
const methodInfo_UserManagementService_getUserProfileAuditTrails = new grpc.web.AbstractClientBase.MethodInfo(
  UserProfileService_pb.GetUpdateAuditTrailResponse,
  /**
   * @param {!proto.org.apache.custos.user.profile.service.GetUpdateAuditTrailRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  UserProfileService_pb.GetUpdateAuditTrailResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.user.profile.service.GetUpdateAuditTrailRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.user.profile.service.GetUpdateAuditTrailResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.user.profile.service.GetUpdateAuditTrailResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.getUserProfileAuditTrails =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/getUserProfileAuditTrails',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_getUserProfileAuditTrails,
      callback);
};


/**
 * @param {!proto.org.apache.custos.user.profile.service.GetUpdateAuditTrailRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.user.profile.service.GetUpdateAuditTrailResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.getUserProfileAuditTrails =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/getUserProfileAuditTrails',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_getUserProfileAuditTrails);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.user.management.service.SynchronizeUserDBRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_UserManagementService_synchronizeUserDBs = new grpc.web.MethodDescriptor(
  '/org.apache.custos.user.management.service.UserManagementService/synchronizeUserDBs',
  grpc.web.MethodType.UNARY,
  proto.org.apache.custos.user.management.service.SynchronizeUserDBRequest,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.user.management.service.SynchronizeUserDBRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.user.management.service.SynchronizeUserDBRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_UserManagementService_synchronizeUserDBs = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.user.management.service.SynchronizeUserDBRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.user.management.service.SynchronizeUserDBRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.user.management.service.UserManagementServiceClient.prototype.synchronizeUserDBs =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/synchronizeUserDBs',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_synchronizeUserDBs,
      callback);
};


/**
 * @param {!proto.org.apache.custos.user.management.service.SynchronizeUserDBRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.user.management.service.UserManagementServicePromiseClient.prototype.synchronizeUserDBs =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.user.management.service.UserManagementService/synchronizeUserDBs',
      request,
      metadata || {},
      methodDescriptor_UserManagementService_synchronizeUserDBs);
};


module.exports = proto.org.apache.custos.user.management.service;

