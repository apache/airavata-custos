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
 * @fileoverview gRPC-Web generated client stub for org.apache.custos.agent.management.service
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');

var IamAdminService_pb = require('../../core-services/iam-admin-service/IamAdminService_pb')
const proto = {};
proto.org = {};
proto.org.apache = {};
proto.org.apache.custos = {};
proto.org.apache.custos.agent = {};
proto.org.apache.custos.agent.management = {};
proto.org.apache.custos.agent.management.service = require('./AgentManagementService_pb');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?Object} options
 * @constructor
 * @struct
 * @final
 */
proto.org.apache.custos.agent.management.service.AgentManagementServiceClient =
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
proto.org.apache.custos.agent.management.service.AgentManagementServicePromiseClient =
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
 *   !proto.org.apache.custos.iam.service.AgentClientMetadata,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_AgentManagementService_enableAgents = new grpc.web.MethodDescriptor(
  '/org.apache.custos.agent.management.service.AgentManagementService/enableAgents',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.AgentClientMetadata,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.AgentClientMetadata} request
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
 *   !proto.org.apache.custos.iam.service.AgentClientMetadata,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_AgentManagementService_enableAgents = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.AgentClientMetadata} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.AgentClientMetadata} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.agent.management.service.AgentManagementServiceClient.prototype.enableAgents =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/enableAgents',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_enableAgents,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.AgentClientMetadata} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.agent.management.service.AgentManagementServicePromiseClient.prototype.enableAgents =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/enableAgents',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_enableAgents);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.AgentClientMetadata,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_AgentManagementService_configureAgentClient = new grpc.web.MethodDescriptor(
  '/org.apache.custos.agent.management.service.AgentManagementService/configureAgentClient',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.AgentClientMetadata,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.AgentClientMetadata} request
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
 *   !proto.org.apache.custos.iam.service.AgentClientMetadata,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_AgentManagementService_configureAgentClient = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.iam.service.AgentClientMetadata} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.AgentClientMetadata} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.agent.management.service.AgentManagementServiceClient.prototype.configureAgentClient =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/configureAgentClient',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_configureAgentClient,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.AgentClientMetadata} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.agent.management.service.AgentManagementServicePromiseClient.prototype.configureAgentClient =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/configureAgentClient',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_configureAgentClient);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.RegisterUserRequest,
 *   !proto.org.apache.custos.agent.management.service.AgentRegistrationResponse>}
 */
const methodDescriptor_AgentManagementService_registerAndEnableAgent = new grpc.web.MethodDescriptor(
  '/org.apache.custos.agent.management.service.AgentManagementService/registerAndEnableAgent',
  grpc.web.MethodType.UNARY,
  IamAdminService_pb.RegisterUserRequest,
  proto.org.apache.custos.agent.management.service.AgentRegistrationResponse,
  /**
   * @param {!proto.org.apache.custos.iam.service.RegisterUserRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.org.apache.custos.agent.management.service.AgentRegistrationResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.iam.service.RegisterUserRequest,
 *   !proto.org.apache.custos.agent.management.service.AgentRegistrationResponse>}
 */
const methodInfo_AgentManagementService_registerAndEnableAgent = new grpc.web.AbstractClientBase.MethodInfo(
  proto.org.apache.custos.agent.management.service.AgentRegistrationResponse,
  /**
   * @param {!proto.org.apache.custos.iam.service.RegisterUserRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.org.apache.custos.agent.management.service.AgentRegistrationResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.iam.service.RegisterUserRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.agent.management.service.AgentRegistrationResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.agent.management.service.AgentRegistrationResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.agent.management.service.AgentManagementServiceClient.prototype.registerAndEnableAgent =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/registerAndEnableAgent',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_registerAndEnableAgent,
      callback);
};


/**
 * @param {!proto.org.apache.custos.iam.service.RegisterUserRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.agent.management.service.AgentRegistrationResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.agent.management.service.AgentManagementServicePromiseClient.prototype.registerAndEnableAgent =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/registerAndEnableAgent',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_registerAndEnableAgent);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.agent.management.service.AgentSearchRequest,
 *   !proto.org.apache.custos.iam.service.Agent>}
 */
const methodDescriptor_AgentManagementService_getAgent = new grpc.web.MethodDescriptor(
  '/org.apache.custos.agent.management.service.AgentManagementService/getAgent',
  grpc.web.MethodType.UNARY,
  proto.org.apache.custos.agent.management.service.AgentSearchRequest,
  IamAdminService_pb.Agent,
  /**
   * @param {!proto.org.apache.custos.agent.management.service.AgentSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.Agent.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.agent.management.service.AgentSearchRequest,
 *   !proto.org.apache.custos.iam.service.Agent>}
 */
const methodInfo_AgentManagementService_getAgent = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.Agent,
  /**
   * @param {!proto.org.apache.custos.agent.management.service.AgentSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.Agent.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.agent.management.service.AgentSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.Agent)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.Agent>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.agent.management.service.AgentManagementServiceClient.prototype.getAgent =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/getAgent',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_getAgent,
      callback);
};


/**
 * @param {!proto.org.apache.custos.agent.management.service.AgentSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.Agent>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.agent.management.service.AgentManagementServicePromiseClient.prototype.getAgent =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/getAgent',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_getAgent);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.agent.management.service.AgentSearchRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_AgentManagementService_deleteAgent = new grpc.web.MethodDescriptor(
  '/org.apache.custos.agent.management.service.AgentManagementService/deleteAgent',
  grpc.web.MethodType.UNARY,
  proto.org.apache.custos.agent.management.service.AgentSearchRequest,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.agent.management.service.AgentSearchRequest} request
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
 *   !proto.org.apache.custos.agent.management.service.AgentSearchRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_AgentManagementService_deleteAgent = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.agent.management.service.AgentSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.agent.management.service.AgentSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.agent.management.service.AgentManagementServiceClient.prototype.deleteAgent =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/deleteAgent',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_deleteAgent,
      callback);
};


/**
 * @param {!proto.org.apache.custos.agent.management.service.AgentSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.agent.management.service.AgentManagementServicePromiseClient.prototype.deleteAgent =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/deleteAgent',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_deleteAgent);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.agent.management.service.AgentSearchRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_AgentManagementService_disableAgent = new grpc.web.MethodDescriptor(
  '/org.apache.custos.agent.management.service.AgentManagementService/disableAgent',
  grpc.web.MethodType.UNARY,
  proto.org.apache.custos.agent.management.service.AgentSearchRequest,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.agent.management.service.AgentSearchRequest} request
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
 *   !proto.org.apache.custos.agent.management.service.AgentSearchRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_AgentManagementService_disableAgent = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.agent.management.service.AgentSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.agent.management.service.AgentSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.agent.management.service.AgentManagementServiceClient.prototype.disableAgent =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/disableAgent',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_disableAgent,
      callback);
};


/**
 * @param {!proto.org.apache.custos.agent.management.service.AgentSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.agent.management.service.AgentManagementServicePromiseClient.prototype.disableAgent =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/disableAgent',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_disableAgent);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.agent.management.service.AgentSearchRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_AgentManagementService_enableAgent = new grpc.web.MethodDescriptor(
  '/org.apache.custos.agent.management.service.AgentManagementService/enableAgent',
  grpc.web.MethodType.UNARY,
  proto.org.apache.custos.agent.management.service.AgentSearchRequest,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.agent.management.service.AgentSearchRequest} request
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
 *   !proto.org.apache.custos.agent.management.service.AgentSearchRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_AgentManagementService_enableAgent = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.agent.management.service.AgentSearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.agent.management.service.AgentSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.agent.management.service.AgentManagementServiceClient.prototype.enableAgent =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/enableAgent',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_enableAgent,
      callback);
};


/**
 * @param {!proto.org.apache.custos.agent.management.service.AgentSearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.agent.management.service.AgentManagementServicePromiseClient.prototype.enableAgent =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/enableAgent',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_enableAgent);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.AddUserAttributesRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_AgentManagementService_addAgentAttributes = new grpc.web.MethodDescriptor(
  '/org.apache.custos.agent.management.service.AgentManagementService/addAgentAttributes',
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
const methodInfo_AgentManagementService_addAgentAttributes = new grpc.web.AbstractClientBase.MethodInfo(
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
proto.org.apache.custos.agent.management.service.AgentManagementServiceClient.prototype.addAgentAttributes =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/addAgentAttributes',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_addAgentAttributes,
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
proto.org.apache.custos.agent.management.service.AgentManagementServicePromiseClient.prototype.addAgentAttributes =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/addAgentAttributes',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_addAgentAttributes);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.DeleteUserAttributeRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_AgentManagementService_deleteAgentAttributes = new grpc.web.MethodDescriptor(
  '/org.apache.custos.agent.management.service.AgentManagementService/deleteAgentAttributes',
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
const methodInfo_AgentManagementService_deleteAgentAttributes = new grpc.web.AbstractClientBase.MethodInfo(
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
proto.org.apache.custos.agent.management.service.AgentManagementServiceClient.prototype.deleteAgentAttributes =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/deleteAgentAttributes',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_deleteAgentAttributes,
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
proto.org.apache.custos.agent.management.service.AgentManagementServicePromiseClient.prototype.deleteAgentAttributes =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/deleteAgentAttributes',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_deleteAgentAttributes);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.AddUserRolesRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_AgentManagementService_addRolesToAgent = new grpc.web.MethodDescriptor(
  '/org.apache.custos.agent.management.service.AgentManagementService/addRolesToAgent',
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
const methodInfo_AgentManagementService_addRolesToAgent = new grpc.web.AbstractClientBase.MethodInfo(
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
proto.org.apache.custos.agent.management.service.AgentManagementServiceClient.prototype.addRolesToAgent =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/addRolesToAgent',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_addRolesToAgent,
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
proto.org.apache.custos.agent.management.service.AgentManagementServicePromiseClient.prototype.addRolesToAgent =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/addRolesToAgent',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_addRolesToAgent);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.DeleteUserRolesRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_AgentManagementService_deleteRolesFromAgent = new grpc.web.MethodDescriptor(
  '/org.apache.custos.agent.management.service.AgentManagementService/deleteRolesFromAgent',
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
const methodInfo_AgentManagementService_deleteRolesFromAgent = new grpc.web.AbstractClientBase.MethodInfo(
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
proto.org.apache.custos.agent.management.service.AgentManagementServiceClient.prototype.deleteRolesFromAgent =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/deleteRolesFromAgent',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_deleteRolesFromAgent,
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
proto.org.apache.custos.agent.management.service.AgentManagementServicePromiseClient.prototype.deleteRolesFromAgent =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/deleteRolesFromAgent',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_deleteRolesFromAgent);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.iam.service.AddProtocolMapperRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_AgentManagementService_addProtocolMapper = new grpc.web.MethodDescriptor(
  '/org.apache.custos.agent.management.service.AgentManagementService/addProtocolMapper',
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
const methodInfo_AgentManagementService_addProtocolMapper = new grpc.web.AbstractClientBase.MethodInfo(
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
proto.org.apache.custos.agent.management.service.AgentManagementServiceClient.prototype.addProtocolMapper =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/addProtocolMapper',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_addProtocolMapper,
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
proto.org.apache.custos.agent.management.service.AgentManagementServicePromiseClient.prototype.addProtocolMapper =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/addProtocolMapper',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_addProtocolMapper);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.agent.management.service.SynchronizeAgentDBRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodDescriptor_AgentManagementService_synchronizeAgentDBs = new grpc.web.MethodDescriptor(
  '/org.apache.custos.agent.management.service.AgentManagementService/synchronizeAgentDBs',
  grpc.web.MethodType.UNARY,
  proto.org.apache.custos.agent.management.service.SynchronizeAgentDBRequest,
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.agent.management.service.SynchronizeAgentDBRequest} request
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
 *   !proto.org.apache.custos.agent.management.service.SynchronizeAgentDBRequest,
 *   !proto.org.apache.custos.iam.service.OperationStatus>}
 */
const methodInfo_AgentManagementService_synchronizeAgentDBs = new grpc.web.AbstractClientBase.MethodInfo(
  IamAdminService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.agent.management.service.SynchronizeAgentDBRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IamAdminService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.agent.management.service.SynchronizeAgentDBRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.iam.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.iam.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.agent.management.service.AgentManagementServiceClient.prototype.synchronizeAgentDBs =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/synchronizeAgentDBs',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_synchronizeAgentDBs,
      callback);
};


/**
 * @param {!proto.org.apache.custos.agent.management.service.SynchronizeAgentDBRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.iam.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.agent.management.service.AgentManagementServicePromiseClient.prototype.synchronizeAgentDBs =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.agent.management.service.AgentManagementService/synchronizeAgentDBs',
      request,
      metadata || {},
      methodDescriptor_AgentManagementService_synchronizeAgentDBs);
};


module.exports = proto.org.apache.custos.agent.management.service;

