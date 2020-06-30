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
 * @fileoverview gRPC-Web generated client stub for org.apache.custos.group.management.service
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');


var UserProfileService_pb = require('../../core-services/user-profile/UserProfileService_pb')

var IamAdminService_pb = require('../../core-services/iam-admin-service/IamAdminService_pb')
const proto = {};
proto.org = {};
proto.org.apache = {};
proto.org.apache.custos = {};
proto.org.apache.custos.group = {};
proto.org.apache.custos.group.management = {};
proto.org.apache.custos.group.management.service = require('./GroupManagementService_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?Object} options
 * @constructor
 * @struct
 * @final
 */
proto.org.apache.custos.group.management.service.GroupManagementServiceClient =
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
proto.org.apache.custos.group.management.service.GroupManagementServicePromiseClient =
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
 *   !proto.org.apache.custos.iam.service.GroupsRequest,
 *   !proto.org.apache.custos.iam.service.GroupsResponse>}
 */
const methodDescriptor_GroupManagementService_createGroups = new grpc.web.MethodDescriptor(
  '/org.apache.custos.group.management.service.GroupManagementService/createGroups',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.GroupsRequest,
  IamAdminService_pb.GroupsResponse,
  /**
   * @param {!proto.org.apache.custos.iam.service.GroupsRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.GroupsResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.GroupsRequest,
 *   !proto.org.apache.custos.iam.service.GroupsResponse>}
 */
const methodInfo_GroupManagementService_createGroups = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.GroupsResponse,
  /**
   * @param {!proto.org.apache.custos.iam.service.GroupsRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.GroupsResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.GroupsRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.GroupsResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.GroupsResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.group.management.service.GroupManagementServiceClient.prototype.createGroups =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/createGroups',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_createGroups,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.GroupsRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.GroupsResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.group.management.service.GroupManagementServicePromiseClient.prototype.createGroups =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/createGroups',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_createGroups);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.GroupRequest,
 *   !proto.org.apache.custos.iam.service.GroupRepresentation>}
 */
const methodDescriptor_GroupManagementService_updateGroup = new grpc.web.MethodDescriptor(
  '/org.apache.custos.group.management.service.GroupManagementService/updateGroup',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.GroupRequest,
  IamAdminService_pb.GroupRepresentation,
  /**
   * @param {!proto.org.apache.custos.iam.service.GroupRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.GroupRepresentation.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.GroupRequest,
 *   !proto.org.apache.custos.iam.service.GroupRepresentation>}
 */
const methodInfo_GroupManagementService_updateGroup = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.GroupRepresentation,
  /**
   * @param {!proto.org.apache.custos.iam.service.GroupRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.GroupRepresentation.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.GroupRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.GroupRepresentation)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.GroupRepresentation>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.group.management.service.GroupManagementServiceClient.prototype.updateGroup =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/updateGroup',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_updateGroup,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.GroupRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.GroupRepresentation>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.group.management.service.GroupManagementServicePromiseClient.prototype.updateGroup =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/updateGroup',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_updateGroup);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.GroupRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_GroupManagementService_deleteGroup = new grpc.web.MethodDescriptor(
  '/org.apache.custos.group.management.service.GroupManagementService/deleteGroup',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.GroupRequest,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.GroupRequest} request
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
 *   !proto.org.apache.custos.iam.service.GroupRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_GroupManagementService_deleteGroup = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.GroupRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.GroupRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.group.management.service.GroupManagementServiceClient.prototype.deleteGroup =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/deleteGroup',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_deleteGroup,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.GroupRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.group.management.service.GroupManagementServicePromiseClient.prototype.deleteGroup =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/deleteGroup',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_deleteGroup);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.GroupRequest,
 *   !proto.org.apache.custos.iam.service.GroupRepresentation>}
 */
const methodDescriptor_GroupManagementService_findGroup = new grpc.web.MethodDescriptor(
  '/org.apache.custos.group.management.service.GroupManagementService/findGroup',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.GroupRequest,
  IamAdminService_pb.GroupRepresentation,
  /**
   * @param {!proto.org.apache.custos.iam.service.GroupRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.GroupRepresentation.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.GroupRequest,
 *   !proto.org.apache.custos.iam.service.GroupRepresentation>}
 */
const methodInfo_GroupManagementService_findGroup = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.GroupRepresentation,
  /**
   * @param {!proto.org.apache.custos.iam.service.GroupRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.GroupRepresentation.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.GroupRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.GroupRepresentation)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.GroupRepresentation>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.group.management.service.GroupManagementServiceClient.prototype.findGroup =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/findGroup',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_findGroup,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.GroupRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.GroupRepresentation>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.group.management.service.GroupManagementServicePromiseClient.prototype.findGroup =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/findGroup',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_findGroup);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.GroupRequest,
 *   !proto.org.apache.custos.iam.service.GroupsResponse>}
 */
const methodDescriptor_GroupManagementService_getAllGroups = new grpc.web.MethodDescriptor(
  '/org.apache.custos.group.management.service.GroupManagementService/getAllGroups',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.GroupRequest,
  IamAdminService_pb.GroupsResponse,
  /**
   * @param {!proto.org.apache.custos.iam.service.GroupRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.GroupsResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.GroupRequest,
 *   !proto.org.apache.custos.iam.service.GroupsResponse>}
 */
const methodInfo_GroupManagementService_getAllGroups = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.GroupsResponse,
  /**
   * @param {!proto.org.apache.custos.iam.service.GroupRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.GroupsResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.GroupRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.GroupsResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.GroupsResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.group.management.service.GroupManagementServiceClient.prototype.getAllGroups =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/getAllGroups',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_getAllGroups,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.GroupRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.GroupsResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.group.management.service.GroupManagementServicePromiseClient.prototype.getAllGroups =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/getAllGroups',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_getAllGroups);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.UserGroupMappingRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_GroupManagementService_addUserToGroup = new grpc.web.MethodDescriptor(
  '/org.apache.custos.group.management.service.GroupManagementService/addUserToGroup',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.UserGroupMappingRequest,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.UserGroupMappingRequest} request
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
 *   !proto.org.apache.custos.iam.service.UserGroupMappingRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_GroupManagementService_addUserToGroup = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.UserGroupMappingRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.UserGroupMappingRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.group.management.service.GroupManagementServiceClient.prototype.addUserToGroup =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/addUserToGroup',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_addUserToGroup,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.UserGroupMappingRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.group.management.service.GroupManagementServicePromiseClient.prototype.addUserToGroup =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/addUserToGroup',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_addUserToGroup);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.UserGroupMappingRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_GroupManagementService_removeUserFromGroup = new grpc.web.MethodDescriptor(
  '/org.apache.custos.group.management.service.GroupManagementService/removeUserFromGroup',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.UserGroupMappingRequest,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.UserGroupMappingRequest} request
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
 *   !proto.org.apache.custos.iam.service.UserGroupMappingRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_GroupManagementService_removeUserFromGroup = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.UserGroupMappingRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.UserGroupMappingRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.group.management.service.GroupManagementServiceClient.prototype.removeUserFromGroup =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/removeUserFromGroup',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_removeUserFromGroup,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.UserGroupMappingRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.group.management.service.GroupManagementServicePromiseClient.prototype.removeUserFromGroup =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/removeUserFromGroup',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_removeUserFromGroup);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.user.profile.service.GroupToGroupMembership,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_GroupManagementService_addChildGroupToParentGroup = new grpc.web.MethodDescriptor(
  '/org.apache.custos.group.management.service.GroupManagementService/addChildGroupToParentGroup',
  grpc.web.MethodType.UNARY,
  UserProfileService_pb.GroupToGroupMembership,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.user.profile.service.GroupToGroupMembership} request
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
 *   !proto.org.apache.custos.user.profile.service.GroupToGroupMembership,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_GroupManagementService_addChildGroupToParentGroup = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.user.profile.service.GroupToGroupMembership} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.user.profile.service.GroupToGroupMembership} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.group.management.service.GroupManagementServiceClient.prototype.addChildGroupToParentGroup =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/addChildGroupToParentGroup',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_addChildGroupToParentGroup,
      callback);
};


/**
 * @param {!proto.org.apache.custos.user.profile.service.GroupToGroupMembership} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.group.management.service.GroupManagementServicePromiseClient.prototype.addChildGroupToParentGroup =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/addChildGroupToParentGroup',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_addChildGroupToParentGroup);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.user.profile.service.GroupToGroupMembership,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_GroupManagementService_removeChildGroupFromParentGroup = new grpc.web.MethodDescriptor(
  '/org.apache.custos.group.management.service.GroupManagementService/removeChildGroupFromParentGroup',
  grpc.web.MethodType.UNARY,
  UserProfileService_pb.GroupToGroupMembership,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.user.profile.service.GroupToGroupMembership} request
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
 *   !proto.org.apache.custos.user.profile.service.GroupToGroupMembership,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_GroupManagementService_removeChildGroupFromParentGroup = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.user.profile.service.GroupToGroupMembership} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.user.profile.service.GroupToGroupMembership} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.group.management.service.GroupManagementServiceClient.prototype.removeChildGroupFromParentGroup =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/removeChildGroupFromParentGroup',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_removeChildGroupFromParentGroup,
      callback);
};


/**
 * @param {!proto.org.apache.custos.user.profile.service.GroupToGroupMembership} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.group.management.service.GroupManagementServicePromiseClient.prototype.removeChildGroupFromParentGroup =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/removeChildGroupFromParentGroup',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_removeChildGroupFromParentGroup);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.user.profile.service.UserProfileRequest,
 *   !proto.org.apache.custos.user.profile.service.GetAllGroupsResponse>}
 */
const methodDescriptor_GroupManagementService_getAllGroupsOfUser = new grpc.web.MethodDescriptor(
  '/org.apache.custos.group.management.service.GroupManagementService/getAllGroupsOfUser',
  grpc.web.MethodType.UNARY,
  UserProfileService_pb.UserProfileRequest,
  UserProfileService_pb.GetAllGroupsResponse,
  /**
   * @param {!proto.org.apache.custos.user.profile.service.UserProfileRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  UserProfileService_pb.GetAllGroupsResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.user.profile.service.UserProfileRequest,
 *   !proto.org.apache.custos.user.profile.service.GetAllGroupsResponse>}
 */
const methodInfo_GroupManagementService_getAllGroupsOfUser = new grpc.web.AbstractClientBase.MethodInfo(
  UserProfileService_pb.GetAllGroupsResponse,
  /**
   * @param {!proto.org.apache.custos.user.profile.service.UserProfileRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  UserProfileService_pb.GetAllGroupsResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.user.profile.service.UserProfileRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.user.profile.service.GetAllGroupsResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.user.profile.service.GetAllGroupsResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.group.management.service.GroupManagementServiceClient.prototype.getAllGroupsOfUser =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/getAllGroupsOfUser',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_getAllGroupsOfUser,
      callback);
};


/**
 * @param {!proto.org.apache.custos.user.profile.service.UserProfileRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.user.profile.service.GetAllGroupsResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.group.management.service.GroupManagementServicePromiseClient.prototype.getAllGroupsOfUser =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/getAllGroupsOfUser',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_getAllGroupsOfUser);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.user.profile.service.GroupRequest,
 *   !proto.org.apache.custos.user.profile.service.GetAllGroupsResponse>}
 */
const methodDescriptor_GroupManagementService_getAllParentGroupsOfGroup = new grpc.web.MethodDescriptor(
  '/org.apache.custos.group.management.service.GroupManagementService/getAllParentGroupsOfGroup',
  grpc.web.MethodType.UNARY,
  UserProfileService_pb.GroupRequest,
  UserProfileService_pb.GetAllGroupsResponse,
  /**
   * @param {!proto.org.apache.custos.user.profile.service.GroupRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  UserProfileService_pb.GetAllGroupsResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.user.profile.service.GroupRequest,
 *   !proto.org.apache.custos.user.profile.service.GetAllGroupsResponse>}
 */
const methodInfo_GroupManagementService_getAllParentGroupsOfGroup = new grpc.web.AbstractClientBase.MethodInfo(
  UserProfileService_pb.GetAllGroupsResponse,
  /**
   * @param {!proto.org.apache.custos.user.profile.service.GroupRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  UserProfileService_pb.GetAllGroupsResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.user.profile.service.GroupRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.user.profile.service.GetAllGroupsResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.user.profile.service.GetAllGroupsResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.group.management.service.GroupManagementServiceClient.prototype.getAllParentGroupsOfGroup =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/getAllParentGroupsOfGroup',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_getAllParentGroupsOfGroup,
      callback);
};


/**
 * @param {!proto.org.apache.custos.user.profile.service.GroupRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.user.profile.service.GetAllGroupsResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.group.management.service.GroupManagementServicePromiseClient.prototype.getAllParentGroupsOfGroup =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/getAllParentGroupsOfGroup',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_getAllParentGroupsOfGroup);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.user.profile.service.GroupRequest,
 *   !proto.org.apache.custos.user.profile.service.GetAllUserProfilesResponse>}
 */
const methodDescriptor_GroupManagementService_getAllChildUsers = new grpc.web.MethodDescriptor(
  '/org.apache.custos.group.management.service.GroupManagementService/getAllChildUsers',
  grpc.web.MethodType.UNARY,
  UserProfileService_pb.GroupRequest,
  UserProfileService_pb.GetAllUserProfilesResponse,
  /**
   * @param {!proto.org.apache.custos.user.profile.service.GroupRequest} request
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
 *   !proto.org.apache.custos.user.profile.service.GroupRequest,
 *   !proto.org.apache.custos.user.profile.service.GetAllUserProfilesResponse>}
 */
const methodInfo_GroupManagementService_getAllChildUsers = new grpc.web.AbstractClientBase.MethodInfo(
  UserProfileService_pb.GetAllUserProfilesResponse,
  /**
   * @param {!proto.org.apache.custos.user.profile.service.GroupRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  UserProfileService_pb.GetAllUserProfilesResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.user.profile.service.GroupRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.user.profile.service.GetAllUserProfilesResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.user.profile.service.GetAllUserProfilesResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.group.management.service.GroupManagementServiceClient.prototype.getAllChildUsers =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/getAllChildUsers',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_getAllChildUsers,
      callback);
};


/**
 * @param {!proto.org.apache.custos.user.profile.service.GroupRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.user.profile.service.GetAllUserProfilesResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.group.management.service.GroupManagementServicePromiseClient.prototype.getAllChildUsers =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/getAllChildUsers',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_getAllChildUsers);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.user.profile.service.GroupRequest,
 *   !proto.org.apache.custos.user.profile.service.GetAllGroupsResponse>}
 */
const methodDescriptor_GroupManagementService_getAllChildGroups = new grpc.web.MethodDescriptor(
  '/org.apache.custos.group.management.service.GroupManagementService/getAllChildGroups',
  grpc.web.MethodType.UNARY,
  UserProfileService_pb.GroupRequest,
  UserProfileService_pb.GetAllGroupsResponse,
  /**
   * @param {!proto.org.apache.custos.user.profile.service.GroupRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  UserProfileService_pb.GetAllGroupsResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.user.profile.service.GroupRequest,
 *   !proto.org.apache.custos.user.profile.service.GetAllGroupsResponse>}
 */
const methodInfo_GroupManagementService_getAllChildGroups = new grpc.web.AbstractClientBase.MethodInfo(
  UserProfileService_pb.GetAllGroupsResponse,
  /**
   * @param {!proto.org.apache.custos.user.profile.service.GroupRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  UserProfileService_pb.GetAllGroupsResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.user.profile.service.GroupRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.user.profile.service.GetAllGroupsResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.user.profile.service.GetAllGroupsResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.group.management.service.GroupManagementServiceClient.prototype.getAllChildGroups =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/getAllChildGroups',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_getAllChildGroups,
      callback);
};


/**
 * @param {!proto.org.apache.custos.user.profile.service.GroupRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.user.profile.service.GetAllGroupsResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.group.management.service.GroupManagementServicePromiseClient.prototype.getAllChildGroups =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/getAllChildGroups',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_getAllChildGroups);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.user.profile.service.GroupMembership,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_GroupManagementService_changeUserMembershipType = new grpc.web.MethodDescriptor(
  '/org.apache.custos.group.management.service.GroupManagementService/changeUserMembershipType',
  grpc.web.MethodType.UNARY,
  UserProfileService_pb.GroupMembership,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.user.profile.service.GroupMembership} request
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
 *   !proto.org.apache.custos.user.profile.service.GroupMembership,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_GroupManagementService_changeUserMembershipType = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.user.profile.service.GroupMembership} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.user.profile.service.GroupMembership} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.group.management.service.GroupManagementServiceClient.prototype.changeUserMembershipType =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/changeUserMembershipType',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_changeUserMembershipType,
      callback);
};


/**
 * @param {!proto.org.apache.custos.user.profile.service.GroupMembership} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.group.management.service.GroupManagementServicePromiseClient.prototype.changeUserMembershipType =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/changeUserMembershipType',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_changeUserMembershipType);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.user.profile.service.GroupMembership,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_GroupManagementService_hasAccess = new grpc.web.MethodDescriptor(
  '/org.apache.custos.group.management.service.GroupManagementService/hasAccess',
  grpc.web.MethodType.UNARY,
  UserProfileService_pb.GroupMembership,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.user.profile.service.GroupMembership} request
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
 *   !proto.org.apache.custos.user.profile.service.GroupMembership,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_GroupManagementService_hasAccess = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.user.profile.service.GroupMembership} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.user.profile.service.GroupMembership} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.group.management.service.GroupManagementServiceClient.prototype.hasAccess =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/hasAccess',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_hasAccess,
      callback);
};


/**
 * @param {!proto.org.apache.custos.user.profile.service.GroupMembership} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.group.management.service.GroupManagementServicePromiseClient.prototype.hasAccess =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.group.management.service.GroupManagementService/hasAccess',
      request,
      metadata || {},
      methodDescriptor_GroupManagementService_hasAccess);
};


module.exports = proto.org.apache.custos.group.management.service;

