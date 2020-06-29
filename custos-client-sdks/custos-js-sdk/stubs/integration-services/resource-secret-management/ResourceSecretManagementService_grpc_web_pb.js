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
 * @fileoverview gRPC-Web generated client stub for org.apache.custos.resource.secret.management.service
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');


// var google_api_annotations_pb = require('../../../google/api/annotations_pb.js')
//
// var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js')

var google_protobuf_struct_pb = require('google-protobuf/google/protobuf/struct_pb.js')

var ResourceSecretService_pb = require('./../../core-services/resource-secret-service/ResourceSecretService_pb.js')

var IdentityService_pb = require('./../../core-services/identity-service/IdentityService_pb.js')
const proto = {};
proto.org = {};
proto.org.apache = {};
proto.org.apache.custos = {};
proto.org.apache.custos.resource = {};
proto.org.apache.custos.resource.secret = {};
proto.org.apache.custos.resource.secret.management = {};
proto.org.apache.custos.resource.secret.management.service = require('./ResourceSecretManagementService_pb');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?Object} options
 * @constructor
 * @struct
 * @final
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServiceClient =
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
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServicePromiseClient =
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
 *   !proto.org.apache.custos.resource.secret.service.GetSecretRequest,
 *   !proto.org.apache.custos.resource.secret.service.SecretMetadata>}
 */
const methodDescriptor_ResourceSecretManagementService_getSecret = new grpc.web.MethodDescriptor(
  '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getSecret',
  grpc.web.MethodType.UNARY,
  ResourceSecretService_pb.GetSecretRequest,
  ResourceSecretService_pb.SecretMetadata,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.GetSecretRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.SecretMetadata.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.resource.secret.service.GetSecretRequest,
 *   !proto.org.apache.custos.resource.secret.service.SecretMetadata>}
 */
const methodInfo_ResourceSecretManagementService_getSecret = new grpc.web.AbstractClientBase.MethodInfo(
  ResourceSecretService_pb.SecretMetadata,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.GetSecretRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.SecretMetadata.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.resource.secret.service.GetSecretRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.resource.secret.service.SecretMetadata)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.resource.secret.service.SecretMetadata>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServiceClient.prototype.getSecret =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getSecret',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_getSecret,
      callback);
};


/**
 * @param {!proto.org.apache.custos.resource.secret.service.GetSecretRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.resource.secret.service.SecretMetadata>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServicePromiseClient.prototype.getSecret =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getSecret',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_getSecret);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.identity.service.GetJWKSRequest,
 *   !proto.google.protobuf.Struct>}
 */
const methodDescriptor_ResourceSecretManagementService_getJWKS = new grpc.web.MethodDescriptor(
  '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getJWKS',
  grpc.web.MethodType.UNARY,
  IdentityService_pb.GetJWKSRequest,
  google_protobuf_struct_pb.Struct,
  /**
   * @param {!proto.org.apache.custos.identity.service.GetJWKSRequest} request
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
 *   !proto.org.apache.custos.identity.service.GetJWKSRequest,
 *   !proto.google.protobuf.Struct>}
 */
const methodInfo_ResourceSecretManagementService_getJWKS = new grpc.web.AbstractClientBase.MethodInfo(
  google_protobuf_struct_pb.Struct,
  /**
   * @param {!proto.org.apache.custos.identity.service.GetJWKSRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  google_protobuf_struct_pb.Struct.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.identity.service.GetJWKSRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.google.protobuf.Struct)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.google.protobuf.Struct>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServiceClient.prototype.getJWKS =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getJWKS',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_getJWKS,
      callback);
};


/**
 * @param {!proto.org.apache.custos.identity.service.GetJWKSRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.google.protobuf.Struct>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServicePromiseClient.prototype.getJWKS =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getJWKS',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_getJWKS);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest,
 *   !proto.org.apache.custos.resource.secret.service.SecretMetadata>}
 */
const methodDescriptor_ResourceSecretManagementService_getResourceCredentialSummary = new grpc.web.MethodDescriptor(
  '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getResourceCredentialSummary',
  grpc.web.MethodType.UNARY,
  ResourceSecretService_pb.GetResourceCredentialByTokenRequest,
  ResourceSecretService_pb.SecretMetadata,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.SecretMetadata.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest,
 *   !proto.org.apache.custos.resource.secret.service.SecretMetadata>}
 */
const methodInfo_ResourceSecretManagementService_getResourceCredentialSummary = new grpc.web.AbstractClientBase.MethodInfo(
  ResourceSecretService_pb.SecretMetadata,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.SecretMetadata.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.resource.secret.service.SecretMetadata)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.resource.secret.service.SecretMetadata>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServiceClient.prototype.getResourceCredentialSummary =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getResourceCredentialSummary',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_getResourceCredentialSummary,
      callback);
};


/**
 * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.resource.secret.service.SecretMetadata>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServicePromiseClient.prototype.getResourceCredentialSummary =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getResourceCredentialSummary',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_getResourceCredentialSummary);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.resource.secret.service.GetResourceCredentialSummariesRequest,
 *   !proto.org.apache.custos.resource.secret.service.ResourceCredentialSummaries>}
 */
const methodDescriptor_ResourceSecretManagementService_getAllResourceCredentialSummaries = new grpc.web.MethodDescriptor(
  '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getAllResourceCredentialSummaries',
  grpc.web.MethodType.UNARY,
  ResourceSecretService_pb.GetResourceCredentialSummariesRequest,
  ResourceSecretService_pb.ResourceCredentialSummaries,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialSummariesRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.ResourceCredentialSummaries.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.resource.secret.service.GetResourceCredentialSummariesRequest,
 *   !proto.org.apache.custos.resource.secret.service.ResourceCredentialSummaries>}
 */
const methodInfo_ResourceSecretManagementService_getAllResourceCredentialSummaries = new grpc.web.AbstractClientBase.MethodInfo(
  ResourceSecretService_pb.ResourceCredentialSummaries,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialSummariesRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.ResourceCredentialSummaries.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialSummariesRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.resource.secret.service.ResourceCredentialSummaries)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.resource.secret.service.ResourceCredentialSummaries>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServiceClient.prototype.getAllResourceCredentialSummaries =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getAllResourceCredentialSummaries',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_getAllResourceCredentialSummaries,
      callback);
};


/**
 * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialSummariesRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.resource.secret.service.ResourceCredentialSummaries>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServicePromiseClient.prototype.getAllResourceCredentialSummaries =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getAllResourceCredentialSummaries',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_getAllResourceCredentialSummaries);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.resource.secret.service.SSHCredential,
 *   !proto.org.apache.custos.resource.secret.service.AddResourceCredentialResponse>}
 */
const methodDescriptor_ResourceSecretManagementService_addSSHCredential = new grpc.web.MethodDescriptor(
  '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/addSSHCredential',
  grpc.web.MethodType.UNARY,
  ResourceSecretService_pb.SSHCredential,
  ResourceSecretService_pb.AddResourceCredentialResponse,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.SSHCredential} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.AddResourceCredentialResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.resource.secret.service.SSHCredential,
 *   !proto.org.apache.custos.resource.secret.service.AddResourceCredentialResponse>}
 */
const methodInfo_ResourceSecretManagementService_addSSHCredential = new grpc.web.AbstractClientBase.MethodInfo(
  ResourceSecretService_pb.AddResourceCredentialResponse,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.SSHCredential} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.AddResourceCredentialResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.resource.secret.service.SSHCredential} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.resource.secret.service.AddResourceCredentialResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.resource.secret.service.AddResourceCredentialResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServiceClient.prototype.addSSHCredential =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/addSSHCredential',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_addSSHCredential,
      callback);
};


/**
 * @param {!proto.org.apache.custos.resource.secret.service.SSHCredential} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.resource.secret.service.AddResourceCredentialResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServicePromiseClient.prototype.addSSHCredential =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/addSSHCredential',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_addSSHCredential);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.resource.secret.service.PasswordCredential,
 *   !proto.org.apache.custos.resource.secret.service.AddResourceCredentialResponse>}
 */
const methodDescriptor_ResourceSecretManagementService_addPasswordCredential = new grpc.web.MethodDescriptor(
  '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/addPasswordCredential',
  grpc.web.MethodType.UNARY,
  ResourceSecretService_pb.PasswordCredential,
  ResourceSecretService_pb.AddResourceCredentialResponse,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.PasswordCredential} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.AddResourceCredentialResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.resource.secret.service.PasswordCredential,
 *   !proto.org.apache.custos.resource.secret.service.AddResourceCredentialResponse>}
 */
const methodInfo_ResourceSecretManagementService_addPasswordCredential = new grpc.web.AbstractClientBase.MethodInfo(
  ResourceSecretService_pb.AddResourceCredentialResponse,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.PasswordCredential} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.AddResourceCredentialResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.resource.secret.service.PasswordCredential} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.resource.secret.service.AddResourceCredentialResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.resource.secret.service.AddResourceCredentialResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServiceClient.prototype.addPasswordCredential =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/addPasswordCredential',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_addPasswordCredential,
      callback);
};


/**
 * @param {!proto.org.apache.custos.resource.secret.service.PasswordCredential} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.resource.secret.service.AddResourceCredentialResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServicePromiseClient.prototype.addPasswordCredential =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/addPasswordCredential',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_addPasswordCredential);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.resource.secret.service.CertificateCredential,
 *   !proto.org.apache.custos.resource.secret.service.AddResourceCredentialResponse>}
 */
const methodDescriptor_ResourceSecretManagementService_addCertificateCredential = new grpc.web.MethodDescriptor(
  '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/addCertificateCredential',
  grpc.web.MethodType.UNARY,
  ResourceSecretService_pb.CertificateCredential,
  ResourceSecretService_pb.AddResourceCredentialResponse,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.CertificateCredential} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.AddResourceCredentialResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.resource.secret.service.CertificateCredential,
 *   !proto.org.apache.custos.resource.secret.service.AddResourceCredentialResponse>}
 */
const methodInfo_ResourceSecretManagementService_addCertificateCredential = new grpc.web.AbstractClientBase.MethodInfo(
  ResourceSecretService_pb.AddResourceCredentialResponse,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.CertificateCredential} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.AddResourceCredentialResponse.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.resource.secret.service.CertificateCredential} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.resource.secret.service.AddResourceCredentialResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.resource.secret.service.AddResourceCredentialResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServiceClient.prototype.addCertificateCredential =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/addCertificateCredential',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_addCertificateCredential,
      callback);
};


/**
 * @param {!proto.org.apache.custos.resource.secret.service.CertificateCredential} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.resource.secret.service.AddResourceCredentialResponse>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServicePromiseClient.prototype.addCertificateCredential =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/addCertificateCredential',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_addCertificateCredential);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest,
 *   !proto.org.apache.custos.resource.secret.service.SSHCredential>}
 */
const methodDescriptor_ResourceSecretManagementService_getSSHCredential = new grpc.web.MethodDescriptor(
  '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getSSHCredential',
  grpc.web.MethodType.UNARY,
  ResourceSecretService_pb.GetResourceCredentialByTokenRequest,
  ResourceSecretService_pb.SSHCredential,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.SSHCredential.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest,
 *   !proto.org.apache.custos.resource.secret.service.SSHCredential>}
 */
const methodInfo_ResourceSecretManagementService_getSSHCredential = new grpc.web.AbstractClientBase.MethodInfo(
  ResourceSecretService_pb.SSHCredential,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.SSHCredential.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.resource.secret.service.SSHCredential)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.resource.secret.service.SSHCredential>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServiceClient.prototype.getSSHCredential =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getSSHCredential',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_getSSHCredential,
      callback);
};


/**
 * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.resource.secret.service.SSHCredential>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServicePromiseClient.prototype.getSSHCredential =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getSSHCredential',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_getSSHCredential);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest,
 *   !proto.org.apache.custos.resource.secret.service.PasswordCredential>}
 */
const methodDescriptor_ResourceSecretManagementService_getPasswordCredential = new grpc.web.MethodDescriptor(
  '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getPasswordCredential',
  grpc.web.MethodType.UNARY,
  ResourceSecretService_pb.GetResourceCredentialByTokenRequest,
  ResourceSecretService_pb.PasswordCredential,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.PasswordCredential.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest,
 *   !proto.org.apache.custos.resource.secret.service.PasswordCredential>}
 */
const methodInfo_ResourceSecretManagementService_getPasswordCredential = new grpc.web.AbstractClientBase.MethodInfo(
  ResourceSecretService_pb.PasswordCredential,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.PasswordCredential.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.resource.secret.service.PasswordCredential)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.resource.secret.service.PasswordCredential>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServiceClient.prototype.getPasswordCredential =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getPasswordCredential',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_getPasswordCredential,
      callback);
};


/**
 * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.resource.secret.service.PasswordCredential>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServicePromiseClient.prototype.getPasswordCredential =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getPasswordCredential',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_getPasswordCredential);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest,
 *   !proto.org.apache.custos.resource.secret.service.CertificateCredential>}
 */
const methodDescriptor_ResourceSecretManagementService_getCertificateCredential = new grpc.web.MethodDescriptor(
  '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getCertificateCredential',
  grpc.web.MethodType.UNARY,
  ResourceSecretService_pb.GetResourceCredentialByTokenRequest,
  ResourceSecretService_pb.CertificateCredential,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.CertificateCredential.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest,
 *   !proto.org.apache.custos.resource.secret.service.CertificateCredential>}
 */
const methodInfo_ResourceSecretManagementService_getCertificateCredential = new grpc.web.AbstractClientBase.MethodInfo(
  ResourceSecretService_pb.CertificateCredential,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.CertificateCredential.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.resource.secret.service.CertificateCredential)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.resource.secret.service.CertificateCredential>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServiceClient.prototype.getCertificateCredential =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getCertificateCredential',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_getCertificateCredential,
      callback);
};


/**
 * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.resource.secret.service.CertificateCredential>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServicePromiseClient.prototype.getCertificateCredential =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/getCertificateCredential',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_getCertificateCredential);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest,
 *   !proto.org.apache.custos.resource.secret.service.ResourceCredentialOperationStatus>}
 */
const methodDescriptor_ResourceSecretManagementService_deleteSSHCredential = new grpc.web.MethodDescriptor(
  '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/deleteSSHCredential',
  grpc.web.MethodType.UNARY,
  ResourceSecretService_pb.GetResourceCredentialByTokenRequest,
  ResourceSecretService_pb.ResourceCredentialOperationStatus,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.ResourceCredentialOperationStatus.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest,
 *   !proto.org.apache.custos.resource.secret.service.ResourceCredentialOperationStatus>}
 */
const methodInfo_ResourceSecretManagementService_deleteSSHCredential = new grpc.web.AbstractClientBase.MethodInfo(
  ResourceSecretService_pb.ResourceCredentialOperationStatus,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.ResourceCredentialOperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.resource.secret.service.ResourceCredentialOperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.resource.secret.service.ResourceCredentialOperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServiceClient.prototype.deleteSSHCredential =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/deleteSSHCredential',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_deleteSSHCredential,
      callback);
};


/**
 * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.resource.secret.service.ResourceCredentialOperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServicePromiseClient.prototype.deleteSSHCredential =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/deleteSSHCredential',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_deleteSSHCredential);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest,
 *   !proto.org.apache.custos.resource.secret.service.ResourceCredentialOperationStatus>}
 */
const methodDescriptor_ResourceSecretManagementService_deletePWDCredential = new grpc.web.MethodDescriptor(
  '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/deletePWDCredential',
  grpc.web.MethodType.UNARY,
  ResourceSecretService_pb.GetResourceCredentialByTokenRequest,
  ResourceSecretService_pb.ResourceCredentialOperationStatus,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.ResourceCredentialOperationStatus.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest,
 *   !proto.org.apache.custos.resource.secret.service.ResourceCredentialOperationStatus>}
 */
const methodInfo_ResourceSecretManagementService_deletePWDCredential = new grpc.web.AbstractClientBase.MethodInfo(
  ResourceSecretService_pb.ResourceCredentialOperationStatus,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.ResourceCredentialOperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.resource.secret.service.ResourceCredentialOperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.resource.secret.service.ResourceCredentialOperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServiceClient.prototype.deletePWDCredential =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/deletePWDCredential',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_deletePWDCredential,
      callback);
};


/**
 * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.resource.secret.service.ResourceCredentialOperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServicePromiseClient.prototype.deletePWDCredential =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/deletePWDCredential',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_deletePWDCredential);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest,
 *   !proto.org.apache.custos.resource.secret.service.ResourceCredentialOperationStatus>}
 */
const methodDescriptor_ResourceSecretManagementService_deleteCertificateCredential = new grpc.web.MethodDescriptor(
  '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/deleteCertificateCredential',
  grpc.web.MethodType.UNARY,
  ResourceSecretService_pb.GetResourceCredentialByTokenRequest,
  ResourceSecretService_pb.ResourceCredentialOperationStatus,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.ResourceCredentialOperationStatus.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest,
 *   !proto.org.apache.custos.resource.secret.service.ResourceCredentialOperationStatus>}
 */
const methodInfo_ResourceSecretManagementService_deleteCertificateCredential = new grpc.web.AbstractClientBase.MethodInfo(
  ResourceSecretService_pb.ResourceCredentialOperationStatus,
  /**
   * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  ResourceSecretService_pb.ResourceCredentialOperationStatus.deserializeBinary
);


/**
 * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.org.apache.custos.resource.secret.service.ResourceCredentialOperationStatus)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.org.apache.custos.resource.secret.service.ResourceCredentialOperationStatus>|undefined}
 *     The XHR Node Readable Stream
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServiceClient.prototype.deleteCertificateCredential =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/deleteCertificateCredential',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_deleteCertificateCredential,
      callback);
};


/**
 * @param {!proto.org.apache.custos.resource.secret.service.GetResourceCredentialByTokenRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.org.apache.custos.resource.secret.service.ResourceCredentialOperationStatus>}
 *     A native promise that resolves to the response
 */
proto.org.apache.custos.resource.secret.management.service.ResourceSecretManagementServicePromiseClient.prototype.deleteCertificateCredential =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/org.apache.custos.resource.secret.management.service.ResourceSecretManagementService/deleteCertificateCredential',
      request,
      metadata || {},
      methodDescriptor_ResourceSecretManagementService_deleteCertificateCredential);
};


module.exports = proto.org.apache.custos.resource.secret.management.service;

