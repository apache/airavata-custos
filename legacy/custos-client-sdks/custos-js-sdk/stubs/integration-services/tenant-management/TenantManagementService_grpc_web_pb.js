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
 * @fileoverview gRPC-Web generated client stub for org.apache.custos.tenant.management.service
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');


var google_api_annotations_pb = require('../../../google/api/annotations_pb.js')

var TenantProfileService_pb = require('../../../TenantProfileService_pb.js')

var google_rpc_error_details_pb = require('../../../google/rpc/error_details_pb.js')

var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js')

var IamAdminService_pb = require('../../../IamAdminService_pb.js')
const proto = {};
proto.org = {};
proto.org.apache = {};
proto.org.apache.custos = {};
proto.org.apache.custos.tenant = {};
proto.org.apache.custos.tenant.management = {};
proto.org.apache.custos.tenant.management.service = require('./TenantManagementService_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?Object} options
 * @constructor
 * @struct
 * @final
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServiceClient =
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
proto.org.apache.custos.tenant.management.service.TenantManagementServicePromiseClient =
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
 *   !proto.org.apache.custos.tenant.profile.service.Tenant,
 *   !proto.org.apache.custos.tenant.management.service.CreateTenantResponse>}
 */
const methodDescriptor_TenantManagementService_createTenant = new grpc.web.MethodDescriptor(
  '/org.apache.custos.tenant.management.service.TenantManagementService/createTenant',
  grpc.web.MethodType.UNARY,
  TenantProfileService_pb.Tenant,
  proto.org.apache.custos.tenant.management.service.CreateTenantResponse,
  /**
   * @param {!proto.org.apache.custos.tenant.profile.service.Tenant} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.org.apache.custos.tenant.management.service.CreateTenantResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.tenant.profile.service.Tenant,
 *   !proto.org.apache.custos.tenant.management.service.CreateTenantResponse>}
 */
const methodInfo_TenantManagementService_createTenant = new grpc.web.AbstractClientBase.MethodInfo(
  proto.org.apache.custos.tenant.management.service.CreateTenantResponse,
  /**
   * @param {!proto.org.apache.custos.tenant.profile.service.Tenant} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.org.apache.custos.tenant.management.service.CreateTenantResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.tenant.profile.service.Tenant} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.tenant.management.service.CreateTenantResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.tenant.management.service.CreateTenantResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServiceClient.prototype.createTenant =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/createTenant',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_createTenant,
      callback);
};


/**
 * @param {!proto.org.apache.custos.tenant.profile.service.Tenant} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.tenant.management.service.CreateTenantResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServicePromiseClient.prototype.createTenant =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/createTenant',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_createTenant);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.tenant.management.service.GetTenantRequest,
 *   !proto.org.apache.custos.tenant.management.service.GetTenantResponse>}
 */
const methodDescriptor_TenantManagementService_getTenant = new grpc.web.MethodDescriptor(
  '/org.apache.custos.tenant.management.service.TenantManagementService/getTenant',
  grpc.web.MethodType.UNARY,
  proto.org.apache.custos.tenant.management.service.GetTenantRequest,
  proto.org.apache.custos.tenant.management.service.GetTenantResponse,
  /**
   * @param {!proto.org.apache.custos.tenant.management.service.GetTenantRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.org.apache.custos.tenant.management.service.GetTenantResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.tenant.management.service.GetTenantRequest,
 *   !proto.org.apache.custos.tenant.management.service.GetTenantResponse>}
 */
const methodInfo_TenantManagementService_getTenant = new grpc.web.AbstractClientBase.MethodInfo(
  proto.org.apache.custos.tenant.management.service.GetTenantResponse,
  /**
   * @param {!proto.org.apache.custos.tenant.management.service.GetTenantRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.org.apache.custos.tenant.management.service.GetTenantResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.tenant.management.service.GetTenantRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.tenant.management.service.GetTenantResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.tenant.management.service.GetTenantResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServiceClient.prototype.getTenant =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/getTenant',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_getTenant,
      callback);
};


/**
 * @param {!proto.org.apache.custos.tenant.management.service.GetTenantRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.tenant.management.service.GetTenantResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServicePromiseClient.prototype.getTenant =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/getTenant',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_getTenant);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.tenant.management.service.UpdateTenantRequest,
 *   !proto.org.apache.custos.tenant.management.service.GetTenantResponse>}
 */
const methodDescriptor_TenantManagementService_updateTenant = new grpc.web.MethodDescriptor(
  '/org.apache.custos.tenant.management.service.TenantManagementService/updateTenant',
  grpc.web.MethodType.UNARY,
  proto.org.apache.custos.tenant.management.service.UpdateTenantRequest,
  proto.org.apache.custos.tenant.management.service.GetTenantResponse,
  /**
   * @param {!proto.org.apache.custos.tenant.management.service.UpdateTenantRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.org.apache.custos.tenant.management.service.GetTenantResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.tenant.management.service.UpdateTenantRequest,
 *   !proto.org.apache.custos.tenant.management.service.GetTenantResponse>}
 */
const methodInfo_TenantManagementService_updateTenant = new grpc.web.AbstractClientBase.MethodInfo(
  proto.org.apache.custos.tenant.management.service.GetTenantResponse,
  /**
   * @param {!proto.org.apache.custos.tenant.management.service.UpdateTenantRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.org.apache.custos.tenant.management.service.GetTenantResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.tenant.management.service.UpdateTenantRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.tenant.management.service.GetTenantResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.tenant.management.service.GetTenantResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServiceClient.prototype.updateTenant =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/updateTenant',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_updateTenant,
      callback);
};


/**
 * @param {!proto.org.apache.custos.tenant.management.service.UpdateTenantRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.tenant.management.service.GetTenantResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServicePromiseClient.prototype.updateTenant =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/updateTenant',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_updateTenant);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.tenant.management.service.DeleteTenantRequest,
 *   !proto.google.protobuf.Empty>}
 */
const methodDescriptor_TenantManagementService_deleteTenant = new grpc.web.MethodDescriptor(
  '/org.apache.custos.tenant.management.service.TenantManagementService/deleteTenant',
  grpc.web.MethodType.UNARY,
  proto.org.apache.custos.tenant.management.service.DeleteTenantRequest,
  google_protobuf_empty_pb.Empty,
  /**
   * @param {!proto.org.apache.custos.tenant.management.service.DeleteTenantRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  google_protobuf_empty_pb.Empty.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.tenant.management.service.DeleteTenantRequest,
 *   !proto.google.protobuf.Empty>}
 */
const methodInfo_TenantManagementService_deleteTenant = new grpc.web.AbstractClientBase.MethodInfo(
  google_protobuf_empty_pb.Empty,
  /**
   * @param {!proto.org.apache.custos.tenant.management.service.DeleteTenantRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  google_protobuf_empty_pb.Empty.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.tenant.management.service.DeleteTenantRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.google.protobuf.Empty)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.google.protobuf.Empty>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServiceClient.prototype.deleteTenant =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/deleteTenant',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_deleteTenant,
      callback);
};


/**
 * @param {!proto.org.apache.custos.tenant.management.service.DeleteTenantRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.google.protobuf.Empty>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServicePromiseClient.prototype.deleteTenant =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/deleteTenant',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_deleteTenant);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.AddRolesRequest,
 *   !proto.org.apache.custos.iam.service.AllRoles>}
 */
const methodDescriptor_TenantManagementService_addTenantRoles = new grpc.web.MethodDescriptor(
  '/org.apache.custos.tenant.management.service.TenantManagementService/addTenantRoles',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.AddRolesRequest,
  IamAdminService_pb.AllRoles,
  /**
   * @param {!proto.org.apache.custos.iam.service.AddRolesRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.AllRoles.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.AddRolesRequest,
 *   !proto.org.apache.custos.iam.service.AllRoles>}
 */
const methodInfo_TenantManagementService_addTenantRoles = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.AllRoles,
  /**
   * @param {!proto.org.apache.custos.iam.service.AddRolesRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.AllRoles.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.AddRolesRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.AllRoles)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.AllRoles>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServiceClient.prototype.addTenantRoles =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/addTenantRoles',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_addTenantRoles,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.AddRolesRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.AllRoles>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServicePromiseClient.prototype.addTenantRoles =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/addTenantRoles',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_addTenantRoles);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.AddProtocolMapperRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_TenantManagementService_addProtocolMapper = new grpc.web.MethodDescriptor(
  '/org.apache.custos.tenant.management.service.TenantManagementService/addProtocolMapper',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.AddProtocolMapperRequest,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest} request
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
 *   !proto.org.apache.custos.iam.service.AddProtocolMapperRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_TenantManagementService_addProtocolMapper = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServiceClient.prototype.addProtocolMapper =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/addProtocolMapper',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_addProtocolMapper,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.AddProtocolMapperRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServicePromiseClient.prototype.addProtocolMapper =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/addProtocolMapper',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_addProtocolMapper);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.EventPersistenceRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_TenantManagementService_configureEventPersistence = new grpc.web.MethodDescriptor(
  '/org.apache.custos.tenant.management.service.TenantManagementService/configureEventPersistence',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.EventPersistenceRequest,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.EventPersistenceRequest} request
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
 *   !proto.org.apache.custos.iam.service.EventPersistenceRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_TenantManagementService_configureEventPersistence = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.EventPersistenceRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.EventPersistenceRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServiceClient.prototype.configureEventPersistence =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/configureEventPersistence',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_configureEventPersistence,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.EventPersistenceRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServicePromiseClient.prototype.configureEventPersistence =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/configureEventPersistence',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_configureEventPersistence);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.tenant.profile.service.UpdateStatusRequest,
 *   !proto.org.apache.custos.tenant.profile.service.UpdateStatusResponse>}
 */
const methodDescriptor_TenantManagementService_updateTenantStatus = new grpc.web.MethodDescriptor(
  '/org.apache.custos.tenant.management.service.TenantManagementService/updateTenantStatus',
  grpc.web.MethodType.UNARY,
  TenantProfileService_pb.UpdateStatusRequest,
  TenantProfileService_pb.UpdateStatusResponse,
  /**
   * @param {!proto.org.apache.custos.tenant.profile.service.UpdateStatusRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  TenantProfileService_pb.UpdateStatusResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.tenant.profile.service.UpdateStatusRequest,
 *   !proto.org.apache.custos.tenant.profile.service.UpdateStatusResponse>}
 */
const methodInfo_TenantManagementService_updateTenantStatus = new grpc.web.AbstractClientBase.MethodInfo(
  TenantProfileService_pb.UpdateStatusResponse,
  /**
   * @param {!proto.org.apache.custos.tenant.profile.service.UpdateStatusRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  TenantProfileService_pb.UpdateStatusResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.tenant.profile.service.UpdateStatusRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.tenant.profile.service.UpdateStatusResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.tenant.profile.service.UpdateStatusResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServiceClient.prototype.updateTenantStatus =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/updateTenantStatus',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_updateTenantStatus,
      callback);
};


/**
 * @param {!proto.org.apache.custos.tenant.profile.service.UpdateStatusRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.tenant.profile.service.UpdateStatusResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServicePromiseClient.prototype.updateTenantStatus =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/updateTenantStatus',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_updateTenantStatus);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.tenant.profile.service.GetTenantsRequest,
 *   !proto.org.apache.custos.tenant.profile.service.GetAllTenantsResponse>}
 */
const methodDescriptor_TenantManagementService_getAllTenants = new grpc.web.MethodDescriptor(
  '/org.apache.custos.tenant.management.service.TenantManagementService/getAllTenants',
  grpc.web.MethodType.UNARY,
  TenantProfileService_pb.GetTenantsRequest,
  TenantProfileService_pb.GetAllTenantsResponse,
  /**
   * @param {!proto.org.apache.custos.tenant.profile.service.GetTenantsRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  TenantProfileService_pb.GetAllTenantsResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.tenant.profile.service.GetTenantsRequest,
 *   !proto.org.apache.custos.tenant.profile.service.GetAllTenantsResponse>}
 */
const methodInfo_TenantManagementService_getAllTenants = new grpc.web.AbstractClientBase.MethodInfo(
  TenantProfileService_pb.GetAllTenantsResponse,
  /**
   * @param {!proto.org.apache.custos.tenant.profile.service.GetTenantsRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  TenantProfileService_pb.GetAllTenantsResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.tenant.profile.service.GetTenantsRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.tenant.profile.service.GetAllTenantsResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.tenant.profile.service.GetAllTenantsResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServiceClient.prototype.getAllTenants =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/getAllTenants',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_getAllTenants,
      callback);
};


/**
 * @param {!proto.org.apache.custos.tenant.profile.service.GetTenantsRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.tenant.profile.service.GetAllTenantsResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServicePromiseClient.prototype.getAllTenants =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/getAllTenants',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_getAllTenants);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.tenant.profile.service.GetTenantsRequest,
 *   !proto.org.apache.custos.tenant.profile.service.GetAllTenantsResponse>}
 */
const methodDescriptor_TenantManagementService_getChildTenants = new grpc.web.MethodDescriptor(
  '/org.apache.custos.tenant.management.service.TenantManagementService/getChildTenants',
  grpc.web.MethodType.UNARY,
  TenantProfileService_pb.GetTenantsRequest,
  TenantProfileService_pb.GetAllTenantsResponse,
  /**
   * @param {!proto.org.apache.custos.tenant.profile.service.GetTenantsRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  TenantProfileService_pb.GetAllTenantsResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.tenant.profile.service.GetTenantsRequest,
 *   !proto.org.apache.custos.tenant.profile.service.GetAllTenantsResponse>}
 */
const methodInfo_TenantManagementService_getChildTenants = new grpc.web.AbstractClientBase.MethodInfo(
  TenantProfileService_pb.GetAllTenantsResponse,
  /**
   * @param {!proto.org.apache.custos.tenant.profile.service.GetTenantsRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  TenantProfileService_pb.GetAllTenantsResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.tenant.profile.service.GetTenantsRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.tenant.profile.service.GetAllTenantsResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.tenant.profile.service.GetAllTenantsResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServiceClient.prototype.getChildTenants =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/getChildTenants',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_getChildTenants,
      callback);
};


/**
 * @param {!proto.org.apache.custos.tenant.profile.service.GetTenantsRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.tenant.profile.service.GetAllTenantsResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServicePromiseClient.prototype.getChildTenants =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/getChildTenants',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_getChildTenants);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.tenant.profile.service.GetAllTenantsForUserRequest,
 *   !proto.org.apache.custos.tenant.profile.service.GetAllTenantsForUserResponse>}
 */
const methodDescriptor_TenantManagementService_getAllTenantsForUser = new grpc.web.MethodDescriptor(
  '/org.apache.custos.tenant.management.service.TenantManagementService/getAllTenantsForUser',
  grpc.web.MethodType.UNARY,
  TenantProfileService_pb.GetAllTenantsForUserRequest,
  TenantProfileService_pb.GetAllTenantsForUserResponse,
  /**
   * @param {!proto.org.apache.custos.tenant.profile.service.GetAllTenantsForUserRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  TenantProfileService_pb.GetAllTenantsForUserResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.tenant.profile.service.GetAllTenantsForUserRequest,
 *   !proto.org.apache.custos.tenant.profile.service.GetAllTenantsForUserResponse>}
 */
const methodInfo_TenantManagementService_getAllTenantsForUser = new grpc.web.AbstractClientBase.MethodInfo(
  TenantProfileService_pb.GetAllTenantsForUserResponse,
  /**
   * @param {!proto.org.apache.custos.tenant.profile.service.GetAllTenantsForUserRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  TenantProfileService_pb.GetAllTenantsForUserResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.tenant.profile.service.GetAllTenantsForUserRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.tenant.profile.service.GetAllTenantsForUserResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.tenant.profile.service.GetAllTenantsForUserResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServiceClient.prototype.getAllTenantsForUser =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/getAllTenantsForUser',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_getAllTenantsForUser,
      callback);
};


/**
 * @param {!proto.org.apache.custos.tenant.profile.service.GetAllTenantsForUserRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.tenant.profile.service.GetAllTenantsForUserResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServicePromiseClient.prototype.getAllTenantsForUser =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/getAllTenantsForUser',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_getAllTenantsForUser);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.tenant.profile.service.GetAuditTrailRequest,
 *   !proto.org.apache.custos.tenant.profile.service.GetStatusUpdateAuditTrailResponse>}
 */
const methodDescriptor_TenantManagementService_getTenantStatusUpdateAuditTrail = new grpc.web.MethodDescriptor(
  '/org.apache.custos.tenant.management.service.TenantManagementService/getTenantStatusUpdateAuditTrail',
  grpc.web.MethodType.UNARY,
  TenantProfileService_pb.GetAuditTrailRequest,
  TenantProfileService_pb.GetStatusUpdateAuditTrailResponse,
  /**
   * @param {!proto.org.apache.custos.tenant.profile.service.GetAuditTrailRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  TenantProfileService_pb.GetStatusUpdateAuditTrailResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.tenant.profile.service.GetAuditTrailRequest,
 *   !proto.org.apache.custos.tenant.profile.service.GetStatusUpdateAuditTrailResponse>}
 */
const methodInfo_TenantManagementService_getTenantStatusUpdateAuditTrail = new grpc.web.AbstractClientBase.MethodInfo(
  TenantProfileService_pb.GetStatusUpdateAuditTrailResponse,
  /**
   * @param {!proto.org.apache.custos.tenant.profile.service.GetAuditTrailRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  TenantProfileService_pb.GetStatusUpdateAuditTrailResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.tenant.profile.service.GetAuditTrailRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.tenant.profile.service.GetStatusUpdateAuditTrailResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.tenant.profile.service.GetStatusUpdateAuditTrailResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServiceClient.prototype.getTenantStatusUpdateAuditTrail =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/getTenantStatusUpdateAuditTrail',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_getTenantStatusUpdateAuditTrail,
      callback);
};


/**
 * @param {!proto.org.apache.custos.tenant.profile.service.GetAuditTrailRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.tenant.profile.service.GetStatusUpdateAuditTrailResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServicePromiseClient.prototype.getTenantStatusUpdateAuditTrail =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/getTenantStatusUpdateAuditTrail',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_getTenantStatusUpdateAuditTrail);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.tenant.profile.service.GetAuditTrailRequest,
 *   !proto.org.apache.custos.tenant.profile.service.GetAttributeUpdateAuditTrailResponse>}
 */
const methodDescriptor_TenantManagementService_getTenantAttributeUpdateAuditTrail = new grpc.web.MethodDescriptor(
  '/org.apache.custos.tenant.management.service.TenantManagementService/getTenantAttributeUpdateAuditTrail',
  grpc.web.MethodType.UNARY,
  TenantProfileService_pb.GetAuditTrailRequest,
  TenantProfileService_pb.GetAttributeUpdateAuditTrailResponse,
  /**
   * @param {!proto.org.apache.custos.tenant.profile.service.GetAuditTrailRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  TenantProfileService_pb.GetAttributeUpdateAuditTrailResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.tenant.profile.service.GetAuditTrailRequest,
 *   !proto.org.apache.custos.tenant.profile.service.GetAttributeUpdateAuditTrailResponse>}
 */
const methodInfo_TenantManagementService_getTenantAttributeUpdateAuditTrail = new grpc.web.AbstractClientBase.MethodInfo(
  TenantProfileService_pb.GetAttributeUpdateAuditTrailResponse,
  /**
   * @param {!proto.org.apache.custos.tenant.profile.service.GetAuditTrailRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  TenantProfileService_pb.GetAttributeUpdateAuditTrailResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.tenant.profile.service.GetAuditTrailRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.tenant.profile.service.GetAttributeUpdateAuditTrailResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.tenant.profile.service.GetAttributeUpdateAuditTrailResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServiceClient.prototype.getTenantAttributeUpdateAuditTrail =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/getTenantAttributeUpdateAuditTrail',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_getTenantAttributeUpdateAuditTrail,
      callback);
};


/**
 * @param {!proto.org.apache.custos.tenant.profile.service.GetAuditTrailRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.tenant.profile.service.GetAttributeUpdateAuditTrailResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.tenant.management.service.TenantManagementServicePromiseClient.prototype.getTenantAttributeUpdateAuditTrail =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.tenant.management.service.TenantManagementService/getTenantAttributeUpdateAuditTrail',
      request,
      metadata || {},
      methodDescriptor_TenantManagementService_getTenantAttributeUpdateAuditTrail);
};


module.exports = proto.org.apache.custos.tenant.management.service;

