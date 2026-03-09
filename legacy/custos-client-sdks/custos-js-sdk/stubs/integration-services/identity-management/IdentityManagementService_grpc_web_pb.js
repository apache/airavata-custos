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
 * @fileoverview gRPC-Web generated client stub for org.apache.custos.identity.management.service
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');




var IdentityService_pb = require('../../core-services/iam-admin-service/IamAdminService_pb')

var google_protobuf_struct_pb = require('google-protobuf/google/protobuf/struct_pb.js')

var google_protobuf_any_pb = require('google-protobuf/google/protobuf/any_pb.js')

var CredentialStoreService_pb = require('../../core-services/credential-store-service/CredentialStoreService_pb.js')
const proto = {};
proto.org = {};
proto.org.apache = {};
proto.org.apache.custos = {};
proto.org.apache.custos.identity = {};
proto.org.apache.custos.identity.management = {};
proto.org.apache.custos.identity.management.service = require('./IdentityManagementService_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?Object} options
 * @constructor
 * @struct
 * @final
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServiceClient =
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
proto.org.apache.custos.identity.management.service.IdentityManagementServicePromiseClient =
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
 *   !proto.org.apache.custos.identity.service.AuthenticationRequest,
 *   !proto.org.apache.custos.identity.service.AuthToken>}
 */
const methodDescriptor_IdentityManagementService_authenticate = new grpc.web.MethodDescriptor(
  '/org.apache.custos.identity.management.service.IdentityManagementService/authenticate',
  grpc.web.MethodType.UNARY,
  IdentityService_pb.AuthenticationRequest,
  IdentityService_pb.AuthToken,
  /**
   * @param {!proto.org.apache.custos.identity.service.AuthenticationRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IdentityService_pb.AuthToken.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.identity.service.AuthenticationRequest,
 *   !proto.org.apache.custos.identity.service.AuthToken>}
 */
const methodInfo_IdentityManagementService_authenticate = new grpc.web.AbstractClientBase.MethodInfo(
  IdentityService_pb.AuthToken,
  /**
   * @param {!proto.org.apache.custos.identity.service.AuthenticationRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IdentityService_pb.AuthToken.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.identity.service.AuthenticationRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.identity.service.AuthToken)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.identity.service.AuthToken>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServiceClient.prototype.authenticate =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/authenticate',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_authenticate,
      callback);
};


/**
 * @param {!proto.org.apache.custos.identity.service.AuthenticationRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.identity.service.AuthToken>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServicePromiseClient.prototype.authenticate =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/authenticate',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_authenticate);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.identity.service.AuthToken,
 *   !proto.org.apache.custos.identity.service.IsAuthenticateResponse>}
 */
const methodDescriptor_IdentityManagementService_isAuthenticated = new grpc.web.MethodDescriptor(
  '/org.apache.custos.identity.management.service.IdentityManagementService/isAuthenticated',
  grpc.web.MethodType.UNARY,
  IdentityService_pb.AuthToken,
  IdentityService_pb.IsAuthenticateResponse,
  /**
   * @param {!proto.org.apache.custos.identity.service.AuthToken} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IdentityService_pb.IsAuthenticateResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.identity.service.AuthToken,
 *   !proto.org.apache.custos.identity.service.IsAuthenticateResponse>}
 */
const methodInfo_IdentityManagementService_isAuthenticated = new grpc.web.AbstractClientBase.MethodInfo(
  IdentityService_pb.IsAuthenticateResponse,
  /**
   * @param {!proto.org.apache.custos.identity.service.AuthToken} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IdentityService_pb.IsAuthenticateResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.identity.service.AuthToken} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.identity.service.IsAuthenticateResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.identity.service.IsAuthenticateResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServiceClient.prototype.isAuthenticated =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/isAuthenticated',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_isAuthenticated,
      callback);
};


/**
 * @param {!proto.org.apache.custos.identity.service.AuthToken} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.identity.service.IsAuthenticateResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServicePromiseClient.prototype.isAuthenticated =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/isAuthenticated',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_isAuthenticated);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.identity.service.AuthToken,
 *   !proto.org.apache.custos.identity.service.User>}
 */
const methodDescriptor_IdentityManagementService_getUser = new grpc.web.MethodDescriptor(
  '/org.apache.custos.identity.management.service.IdentityManagementService/getUser',
  grpc.web.MethodType.UNARY,
  IdentityService_pb.AuthToken,
  IdentityService_pb.User,
  /**
   * @param {!proto.org.apache.custos.identity.service.AuthToken} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IdentityService_pb.User.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.identity.service.AuthToken,
 *   !proto.org.apache.custos.identity.service.User>}
 */
const methodInfo_IdentityManagementService_getUser = new grpc.web.AbstractClientBase.MethodInfo(
  IdentityService_pb.User,
  /**
   * @param {!proto.org.apache.custos.identity.service.AuthToken} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IdentityService_pb.User.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.identity.service.AuthToken} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.identity.service.User)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.identity.service.User>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServiceClient.prototype.getUser =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/getUser',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_getUser,
      callback);
};


/**
 * @param {!proto.org.apache.custos.identity.service.AuthToken} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.identity.service.User>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServicePromiseClient.prototype.getUser =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/getUser',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_getUser);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.identity.service.GetUserManagementSATokenRequest,
 *   !proto.org.apache.custos.identity.service.AuthToken>}
 */
const methodDescriptor_IdentityManagementService_getUserManagementServiceAccountAccessToken = new grpc.web.MethodDescriptor(
  '/org.apache.custos.identity.management.service.IdentityManagementService/getUserManagementServiceAccountAccessToken',
  grpc.web.MethodType.UNARY,
  IdentityService_pb.GetUserManagementSATokenRequest,
  IdentityService_pb.AuthToken,
  /**
   * @param {!proto.org.apache.custos.identity.service.GetUserManagementSATokenRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IdentityService_pb.AuthToken.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.identity.service.GetUserManagementSATokenRequest,
 *   !proto.org.apache.custos.identity.service.AuthToken>}
 */
const methodInfo_IdentityManagementService_getUserManagementServiceAccountAccessToken = new grpc.web.AbstractClientBase.MethodInfo(
  IdentityService_pb.AuthToken,
  /**
   * @param {!proto.org.apache.custos.identity.service.GetUserManagementSATokenRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IdentityService_pb.AuthToken.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.identity.service.GetUserManagementSATokenRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.identity.service.AuthToken)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.identity.service.AuthToken>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServiceClient.prototype.getUserManagementServiceAccountAccessToken =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/getUserManagementServiceAccountAccessToken',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_getUserManagementServiceAccountAccessToken,
      callback);
};


/**
 * @param {!proto.org.apache.custos.identity.service.GetUserManagementSATokenRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.identity.service.AuthToken>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServicePromiseClient.prototype.getUserManagementServiceAccountAccessToken =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/getUserManagementServiceAccountAccessToken',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_getUserManagementServiceAccountAccessToken);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.identity.management.service.EndSessionRequest,
 *   !proto.org.apache.custos.identity.service.OperationStatus>}
 */
const methodDescriptor_IdentityManagementService_endUserSession = new grpc.web.MethodDescriptor(
  '/org.apache.custos.identity.management.service.IdentityManagementService/endUserSession',
  grpc.web.MethodType.UNARY,
  proto.org.apache.custos.identity.management.service.EndSessionRequest,
  IdentityService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.identity.management.service.EndSessionRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IdentityService_pb.OperationStatus.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.identity.management.service.EndSessionRequest,
 *   !proto.org.apache.custos.identity.service.OperationStatus>}
 */
const methodInfo_IdentityManagementService_endUserSession = new grpc.web.AbstractClientBase.MethodInfo(
  IdentityService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.identity.management.service.EndSessionRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IdentityService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.identity.management.service.EndSessionRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.identity.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.identity.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServiceClient.prototype.endUserSession =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/endUserSession',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_endUserSession,
      callback);
};


/**
 * @param {!proto.org.apache.custos.identity.management.service.EndSessionRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.identity.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServicePromiseClient.prototype.endUserSession =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/endUserSession',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_endUserSession);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.identity.management.service.AuthorizationRequest,
 *   !proto.org.apache.custos.identity.management.service.AuthorizationResponse>}
 */
const methodDescriptor_IdentityManagementService_authorize = new grpc.web.MethodDescriptor(
  '/org.apache.custos.identity.management.service.IdentityManagementService/authorize',
  grpc.web.MethodType.UNARY,
  proto.org.apache.custos.identity.management.service.AuthorizationRequest,
  proto.org.apache.custos.identity.management.service.AuthorizationResponse,
  /**
   * @param {!proto.org.apache.custos.identity.management.service.AuthorizationRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.org.apache.custos.identity.management.service.AuthorizationResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.identity.management.service.AuthorizationRequest,
 *   !proto.org.apache.custos.identity.management.service.AuthorizationResponse>}
 */
const methodInfo_IdentityManagementService_authorize = new grpc.web.AbstractClientBase.MethodInfo(
  proto.org.apache.custos.identity.management.service.AuthorizationResponse,
  /**
   * @param {!proto.org.apache.custos.identity.management.service.AuthorizationRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.org.apache.custos.identity.management.service.AuthorizationResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.identity.management.service.AuthorizationRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.identity.management.service.AuthorizationResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.identity.management.service.AuthorizationResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServiceClient.prototype.authorize =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/authorize',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_authorize,
      callback);
};


/**
 * @param {!proto.org.apache.custos.identity.management.service.AuthorizationRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.identity.management.service.AuthorizationResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServicePromiseClient.prototype.authorize =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/authorize',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_authorize);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.identity.service.GetTokenRequest,
 *   !proto.google.protobuf.Struct>}
 */
const methodDescriptor_IdentityManagementService_token = new grpc.web.MethodDescriptor(
  '/org.apache.custos.identity.management.service.IdentityManagementService/token',
  grpc.web.MethodType.UNARY,
  IdentityService_pb.GetTokenRequest,
  google_protobuf_struct_pb.Struct,
  /**
   * @param {!proto.org.apache.custos.identity.service.GetTokenRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  google_protobuf_struct_pb.Struct.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.identity.service.GetTokenRequest,
 *   !proto.google.protobuf.Struct>}
 */
const methodInfo_IdentityManagementService_token = new grpc.web.AbstractClientBase.MethodInfo(
  google_protobuf_struct_pb.Struct,
  /**
   * @param {!proto.org.apache.custos.identity.service.GetTokenRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  google_protobuf_struct_pb.Struct.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.identity.service.GetTokenRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.google.protobuf.Struct)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.google.protobuf.Struct>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServiceClient.prototype.token =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/token',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_token,
      callback);
};


/**
 * @param {!proto.org.apache.custos.identity.service.GetTokenRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.google.protobuf.Struct>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServicePromiseClient.prototype.token =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/token',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_token);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.identity.management.service.GetCredentialsRequest,
 *   !proto.org.apache.custos.credential.store.service.Credentials>}
 */
const methodDescriptor_IdentityManagementService_getCredentials = new grpc.web.MethodDescriptor(
  '/org.apache.custos.identity.management.service.IdentityManagementService/getCredentials',
  grpc.web.MethodType.UNARY,
  proto.org.apache.custos.identity.management.service.GetCredentialsRequest,
  CredentialStoreService_pb.Credentials,
  /**
   * @param {!proto.org.apache.custos.identity.management.service.GetCredentialsRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  CredentialStoreService_pb.Credentials.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.identity.management.service.GetCredentialsRequest,
 *   !proto.org.apache.custos.credential.store.service.Credentials>}
 */
const methodInfo_IdentityManagementService_getCredentials = new grpc.web.AbstractClientBase.MethodInfo(
  CredentialStoreService_pb.Credentials,
  /**
   * @param {!proto.org.apache.custos.identity.management.service.GetCredentialsRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  CredentialStoreService_pb.Credentials.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.identity.management.service.GetCredentialsRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.credential.store.service.Credentials)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.credential.store.service.Credentials>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServiceClient.prototype.getCredentials =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/getCredentials',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_getCredentials,
      callback);
};


/**
 * @param {!proto.org.apache.custos.identity.management.service.GetCredentialsRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.credential.store.service.Credentials>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServicePromiseClient.prototype.getCredentials =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/getCredentials',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_getCredentials);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.identity.service.GetOIDCConfiguration,
 *   !proto.google.protobuf.Struct>}
 */
const methodDescriptor_IdentityManagementService_getOIDCConfiguration = new grpc.web.MethodDescriptor(
  '/org.apache.custos.identity.management.service.IdentityManagementService/getOIDCConfiguration',
  grpc.web.MethodType.UNARY,
  IdentityService_pb.GetOIDCConfiguration,
  google_protobuf_struct_pb.Struct,
  /**
   * @param {!proto.org.apache.custos.identity.service.GetOIDCConfiguration} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  google_protobuf_struct_pb.Struct.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.identity.service.GetOIDCConfiguration,
 *   !proto.google.protobuf.Struct>}
 */
const methodInfo_IdentityManagementService_getOIDCConfiguration = new grpc.web.AbstractClientBase.MethodInfo(
  google_protobuf_struct_pb.Struct,
  /**
   * @param {!proto.org.apache.custos.identity.service.GetOIDCConfiguration} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  google_protobuf_struct_pb.Struct.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.identity.service.GetOIDCConfiguration} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.google.protobuf.Struct)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.google.protobuf.Struct>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServiceClient.prototype.getOIDCConfiguration =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/getOIDCConfiguration',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_getOIDCConfiguration,
      callback);
};


/**
 * @param {!proto.org.apache.custos.identity.service.GetOIDCConfiguration} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.google.protobuf.Struct>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServicePromiseClient.prototype.getOIDCConfiguration =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/getOIDCConfiguration',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_getOIDCConfiguration);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.identity.management.service.GetAgentTokenRequest,
 *   !proto.google.protobuf.Struct>}
 */
const methodDescriptor_IdentityManagementService_getAgentToken = new grpc.web.MethodDescriptor(
  '/org.apache.custos.identity.management.service.IdentityManagementService/getAgentToken',
  grpc.web.MethodType.UNARY,
  proto.org.apache.custos.identity.management.service.GetAgentTokenRequest,
  google_protobuf_struct_pb.Struct,
  /**
   * @param {!proto.org.apache.custos.identity.management.service.GetAgentTokenRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  google_protobuf_struct_pb.Struct.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.identity.management.service.GetAgentTokenRequest,
 *   !proto.google.protobuf.Struct>}
 */
const methodInfo_IdentityManagementService_getAgentToken = new grpc.web.AbstractClientBase.MethodInfo(
  google_protobuf_struct_pb.Struct,
  /**
   * @param {!proto.org.apache.custos.identity.management.service.GetAgentTokenRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  google_protobuf_struct_pb.Struct.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.identity.management.service.GetAgentTokenRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.google.protobuf.Struct)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.google.protobuf.Struct>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServiceClient.prototype.getAgentToken =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/getAgentToken',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_getAgentToken,
      callback);
};


/**
 * @param {!proto.org.apache.custos.identity.management.service.GetAgentTokenRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.google.protobuf.Struct>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServicePromiseClient.prototype.getAgentToken =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/getAgentToken',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_getAgentToken);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.identity.management.service.EndSessionRequest,
 *   !proto.org.apache.custos.identity.service.OperationStatus>}
 */
const methodDescriptor_IdentityManagementService_endAgentSession = new grpc.web.MethodDescriptor(
  '/org.apache.custos.identity.management.service.IdentityManagementService/endAgentSession',
  grpc.web.MethodType.UNARY,
  proto.org.apache.custos.identity.management.service.EndSessionRequest,
  IdentityService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.identity.management.service.EndSessionRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IdentityService_pb.OperationStatus.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.identity.management.service.EndSessionRequest,
 *   !proto.org.apache.custos.identity.service.OperationStatus>}
 */
const methodInfo_IdentityManagementService_endAgentSession = new grpc.web.AbstractClientBase.MethodInfo(
  IdentityService_pb.OperationStatus,
  /**
   * @param {!proto.org.apache.custos.identity.management.service.EndSessionRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  IdentityService_pb.OperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.identity.management.service.EndSessionRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.identity.service.OperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.identity.service.OperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServiceClient.prototype.endAgentSession =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/endAgentSession',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_endAgentSession,
      callback);
};


/**
 * @param {!proto.org.apache.custos.identity.management.service.EndSessionRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.identity.service.OperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.identity.management.service.IdentityManagementServicePromiseClient.prototype.endAgentSession =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.identity.management.service.IdentityManagementService/endAgentSession',
      request,
      metadata || {},
      methodDescriptor_IdentityManagementService_endAgentSession);
};


module.exports = proto.org.apache.custos.identity.management.service;

