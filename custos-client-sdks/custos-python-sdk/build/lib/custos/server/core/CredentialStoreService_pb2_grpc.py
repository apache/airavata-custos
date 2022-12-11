#  Licensed to the Apache Software Foundation (ASF) under one
#  or more contributor license agreements. See the NOTICE file
#  distributed with this work for additional information
#  regarding copyright ownership. The ASF licenses this file
#  to you under the Apache License, Version 2.0 (the
#  "License"); you may not use this file except in compliance
#  with the License. You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing,
#   software distributed under the License is distributed on an
#   "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
#   KIND, either express or implied. See the License for the
#   specific language governing permissions and limitations
#   under the License.

# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc

import custos.server.core.CredentialStoreService_pb2 as CredentialStoreService__pb2


class CredentialStoreServiceStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.putCredential = channel.unary_unary(
                '/org.apache.custos.credential.store.service.CredentialStoreService/putCredential',
                request_serializer=CredentialStoreService__pb2.CredentialMetadata.SerializeToString,
                response_deserializer=CredentialStoreService__pb2.OperationStatus.FromString,
                )
        self.deleteCredential = channel.unary_unary(
                '/org.apache.custos.credential.store.service.CredentialStoreService/deleteCredential',
                request_serializer=CredentialStoreService__pb2.DeleteCredentialRequest.SerializeToString,
                response_deserializer=CredentialStoreService__pb2.OperationStatus.FromString,
                )
        self.getCredential = channel.unary_unary(
                '/org.apache.custos.credential.store.service.CredentialStoreService/getCredential',
                request_serializer=CredentialStoreService__pb2.GetCredentialRequest.SerializeToString,
                response_deserializer=CredentialStoreService__pb2.CredentialMetadata.FromString,
                )
        self.getAllCredentials = channel.unary_unary(
                '/org.apache.custos.credential.store.service.CredentialStoreService/getAllCredentials',
                request_serializer=CredentialStoreService__pb2.GetAllCredentialsRequest.SerializeToString,
                response_deserializer=CredentialStoreService__pb2.GetAllCredentialsResponse.FromString,
                )
        self.getOperationMetadata = channel.unary_unary(
                '/org.apache.custos.credential.store.service.CredentialStoreService/getOperationMetadata',
                request_serializer=CredentialStoreService__pb2.GetOperationsMetadataRequest.SerializeToString,
                response_deserializer=CredentialStoreService__pb2.GetOperationsMetadataResponse.FromString,
                )
        self.getNewCustosCredential = channel.unary_unary(
                '/org.apache.custos.credential.store.service.CredentialStoreService/getNewCustosCredential',
                request_serializer=CredentialStoreService__pb2.GetNewCustosCredentialRequest.SerializeToString,
                response_deserializer=CredentialStoreService__pb2.CredentialMetadata.FromString,
                )
        self.getOwnerIdFromToken = channel.unary_unary(
                '/org.apache.custos.credential.store.service.CredentialStoreService/getOwnerIdFromToken',
                request_serializer=CredentialStoreService__pb2.TokenRequest.SerializeToString,
                response_deserializer=CredentialStoreService__pb2.GetOwnerIdResponse.FromString,
                )
        self.getCustosCredentialFromToken = channel.unary_unary(
                '/org.apache.custos.credential.store.service.CredentialStoreService/getCustosCredentialFromToken',
                request_serializer=CredentialStoreService__pb2.TokenRequest.SerializeToString,
                response_deserializer=CredentialStoreService__pb2.CredentialMetadata.FromString,
                )
        self.getCustosCredentialFromClientId = channel.unary_unary(
                '/org.apache.custos.credential.store.service.CredentialStoreService/getCustosCredentialFromClientId',
                request_serializer=CredentialStoreService__pb2.GetCredentialRequest.SerializeToString,
                response_deserializer=CredentialStoreService__pb2.CredentialMetadata.FromString,
                )
        self.getAllCredentialsFromToken = channel.unary_unary(
                '/org.apache.custos.credential.store.service.CredentialStoreService/getAllCredentialsFromToken',
                request_serializer=CredentialStoreService__pb2.TokenRequest.SerializeToString,
                response_deserializer=CredentialStoreService__pb2.GetAllCredentialsResponse.FromString,
                )
        self.getMasterCredentials = channel.unary_unary(
                '/org.apache.custos.credential.store.service.CredentialStoreService/getMasterCredentials',
                request_serializer=CredentialStoreService__pb2.GetCredentialRequest.SerializeToString,
                response_deserializer=CredentialStoreService__pb2.GetAllCredentialsResponse.FromString,
                )
        self.getAllCredentialsFromJWTToken = channel.unary_unary(
                '/org.apache.custos.credential.store.service.CredentialStoreService/getAllCredentialsFromJWTToken',
                request_serializer=CredentialStoreService__pb2.TokenRequest.SerializeToString,
                response_deserializer=CredentialStoreService__pb2.GetAllCredentialsResponse.FromString,
                )
        self.getBasicCredentials = channel.unary_unary(
                '/org.apache.custos.credential.store.service.CredentialStoreService/getBasicCredentials',
                request_serializer=CredentialStoreService__pb2.TokenRequest.SerializeToString,
                response_deserializer=CredentialStoreService__pb2.Credentials.FromString,
                )
        self.createAgentCredential = channel.unary_unary(
                '/org.apache.custos.credential.store.service.CredentialStoreService/createAgentCredential',
                request_serializer=CredentialStoreService__pb2.CredentialMetadata.SerializeToString,
                response_deserializer=CredentialStoreService__pb2.CredentialMetadata.FromString,
                )
        self.getAgentCredential = channel.unary_unary(
                '/org.apache.custos.credential.store.service.CredentialStoreService/getAgentCredential',
                request_serializer=CredentialStoreService__pb2.GetCredentialRequest.SerializeToString,
                response_deserializer=CredentialStoreService__pb2.CredentialMetadata.FromString,
                )
        self.deleteAgentCredential = channel.unary_unary(
                '/org.apache.custos.credential.store.service.CredentialStoreService/deleteAgentCredential',
                request_serializer=CredentialStoreService__pb2.CredentialMetadata.SerializeToString,
                response_deserializer=CredentialStoreService__pb2.OperationStatus.FromString,
                )
        self.getCredentialByAgentBasicAuth = channel.unary_unary(
                '/org.apache.custos.credential.store.service.CredentialStoreService/getCredentialByAgentBasicAuth',
                request_serializer=CredentialStoreService__pb2.TokenRequest.SerializeToString,
                response_deserializer=CredentialStoreService__pb2.GetAllCredentialsResponse.FromString,
                )
        self.getCredentialByAgentJWTToken = channel.unary_unary(
                '/org.apache.custos.credential.store.service.CredentialStoreService/getCredentialByAgentJWTToken',
                request_serializer=CredentialStoreService__pb2.TokenRequest.SerializeToString,
                response_deserializer=CredentialStoreService__pb2.GetAllCredentialsResponse.FromString,
                )
        self.validateAgentJWTToken = channel.unary_unary(
                '/org.apache.custos.credential.store.service.CredentialStoreService/validateAgentJWTToken',
                request_serializer=CredentialStoreService__pb2.TokenRequest.SerializeToString,
                response_deserializer=CredentialStoreService__pb2.OperationStatus.FromString,
                )


class CredentialStoreServiceServicer(object):
    """Missing associated documentation comment in .proto file."""

    def putCredential(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def deleteCredential(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def getCredential(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def getAllCredentials(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def getOperationMetadata(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def getNewCustosCredential(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def getOwnerIdFromToken(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def getCustosCredentialFromToken(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def getCustosCredentialFromClientId(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def getAllCredentialsFromToken(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def getMasterCredentials(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def getAllCredentialsFromJWTToken(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def getBasicCredentials(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def createAgentCredential(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def getAgentCredential(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def deleteAgentCredential(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def getCredentialByAgentBasicAuth(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def getCredentialByAgentJWTToken(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def validateAgentJWTToken(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_CredentialStoreServiceServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'putCredential': grpc.unary_unary_rpc_method_handler(
                    servicer.putCredential,
                    request_deserializer=CredentialStoreService__pb2.CredentialMetadata.FromString,
                    response_serializer=CredentialStoreService__pb2.OperationStatus.SerializeToString,
            ),
            'deleteCredential': grpc.unary_unary_rpc_method_handler(
                    servicer.deleteCredential,
                    request_deserializer=CredentialStoreService__pb2.DeleteCredentialRequest.FromString,
                    response_serializer=CredentialStoreService__pb2.OperationStatus.SerializeToString,
            ),
            'getCredential': grpc.unary_unary_rpc_method_handler(
                    servicer.getCredential,
                    request_deserializer=CredentialStoreService__pb2.GetCredentialRequest.FromString,
                    response_serializer=CredentialStoreService__pb2.CredentialMetadata.SerializeToString,
            ),
            'getAllCredentials': grpc.unary_unary_rpc_method_handler(
                    servicer.getAllCredentials,
                    request_deserializer=CredentialStoreService__pb2.GetAllCredentialsRequest.FromString,
                    response_serializer=CredentialStoreService__pb2.GetAllCredentialsResponse.SerializeToString,
            ),
            'getOperationMetadata': grpc.unary_unary_rpc_method_handler(
                    servicer.getOperationMetadata,
                    request_deserializer=CredentialStoreService__pb2.GetOperationsMetadataRequest.FromString,
                    response_serializer=CredentialStoreService__pb2.GetOperationsMetadataResponse.SerializeToString,
            ),
            'getNewCustosCredential': grpc.unary_unary_rpc_method_handler(
                    servicer.getNewCustosCredential,
                    request_deserializer=CredentialStoreService__pb2.GetNewCustosCredentialRequest.FromString,
                    response_serializer=CredentialStoreService__pb2.CredentialMetadata.SerializeToString,
            ),
            'getOwnerIdFromToken': grpc.unary_unary_rpc_method_handler(
                    servicer.getOwnerIdFromToken,
                    request_deserializer=CredentialStoreService__pb2.TokenRequest.FromString,
                    response_serializer=CredentialStoreService__pb2.GetOwnerIdResponse.SerializeToString,
            ),
            'getCustosCredentialFromToken': grpc.unary_unary_rpc_method_handler(
                    servicer.getCustosCredentialFromToken,
                    request_deserializer=CredentialStoreService__pb2.TokenRequest.FromString,
                    response_serializer=CredentialStoreService__pb2.CredentialMetadata.SerializeToString,
            ),
            'getCustosCredentialFromClientId': grpc.unary_unary_rpc_method_handler(
                    servicer.getCustosCredentialFromClientId,
                    request_deserializer=CredentialStoreService__pb2.GetCredentialRequest.FromString,
                    response_serializer=CredentialStoreService__pb2.CredentialMetadata.SerializeToString,
            ),
            'getAllCredentialsFromToken': grpc.unary_unary_rpc_method_handler(
                    servicer.getAllCredentialsFromToken,
                    request_deserializer=CredentialStoreService__pb2.TokenRequest.FromString,
                    response_serializer=CredentialStoreService__pb2.GetAllCredentialsResponse.SerializeToString,
            ),
            'getMasterCredentials': grpc.unary_unary_rpc_method_handler(
                    servicer.getMasterCredentials,
                    request_deserializer=CredentialStoreService__pb2.GetCredentialRequest.FromString,
                    response_serializer=CredentialStoreService__pb2.GetAllCredentialsResponse.SerializeToString,
            ),
            'getAllCredentialsFromJWTToken': grpc.unary_unary_rpc_method_handler(
                    servicer.getAllCredentialsFromJWTToken,
                    request_deserializer=CredentialStoreService__pb2.TokenRequest.FromString,
                    response_serializer=CredentialStoreService__pb2.GetAllCredentialsResponse.SerializeToString,
            ),
            'getBasicCredentials': grpc.unary_unary_rpc_method_handler(
                    servicer.getBasicCredentials,
                    request_deserializer=CredentialStoreService__pb2.TokenRequest.FromString,
                    response_serializer=CredentialStoreService__pb2.Credentials.SerializeToString,
            ),
            'createAgentCredential': grpc.unary_unary_rpc_method_handler(
                    servicer.createAgentCredential,
                    request_deserializer=CredentialStoreService__pb2.CredentialMetadata.FromString,
                    response_serializer=CredentialStoreService__pb2.CredentialMetadata.SerializeToString,
            ),
            'getAgentCredential': grpc.unary_unary_rpc_method_handler(
                    servicer.getAgentCredential,
                    request_deserializer=CredentialStoreService__pb2.GetCredentialRequest.FromString,
                    response_serializer=CredentialStoreService__pb2.CredentialMetadata.SerializeToString,
            ),
            'deleteAgentCredential': grpc.unary_unary_rpc_method_handler(
                    servicer.deleteAgentCredential,
                    request_deserializer=CredentialStoreService__pb2.CredentialMetadata.FromString,
                    response_serializer=CredentialStoreService__pb2.OperationStatus.SerializeToString,
            ),
            'getCredentialByAgentBasicAuth': grpc.unary_unary_rpc_method_handler(
                    servicer.getCredentialByAgentBasicAuth,
                    request_deserializer=CredentialStoreService__pb2.TokenRequest.FromString,
                    response_serializer=CredentialStoreService__pb2.GetAllCredentialsResponse.SerializeToString,
            ),
            'getCredentialByAgentJWTToken': grpc.unary_unary_rpc_method_handler(
                    servicer.getCredentialByAgentJWTToken,
                    request_deserializer=CredentialStoreService__pb2.TokenRequest.FromString,
                    response_serializer=CredentialStoreService__pb2.GetAllCredentialsResponse.SerializeToString,
            ),
            'validateAgentJWTToken': grpc.unary_unary_rpc_method_handler(
                    servicer.validateAgentJWTToken,
                    request_deserializer=CredentialStoreService__pb2.TokenRequest.FromString,
                    response_serializer=CredentialStoreService__pb2.OperationStatus.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'org.apache.custos.credential.store.service.CredentialStoreService', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class CredentialStoreService(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def putCredential(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/org.apache.custos.credential.store.service.CredentialStoreService/putCredential',
            CredentialStoreService__pb2.CredentialMetadata.SerializeToString,
            CredentialStoreService__pb2.OperationStatus.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def deleteCredential(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/org.apache.custos.credential.store.service.CredentialStoreService/deleteCredential',
            CredentialStoreService__pb2.DeleteCredentialRequest.SerializeToString,
            CredentialStoreService__pb2.OperationStatus.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def getCredential(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/org.apache.custos.credential.store.service.CredentialStoreService/getCredential',
            CredentialStoreService__pb2.GetCredentialRequest.SerializeToString,
            CredentialStoreService__pb2.CredentialMetadata.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def getAllCredentials(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/org.apache.custos.credential.store.service.CredentialStoreService/getAllCredentials',
            CredentialStoreService__pb2.GetAllCredentialsRequest.SerializeToString,
            CredentialStoreService__pb2.GetAllCredentialsResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def getOperationMetadata(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/org.apache.custos.credential.store.service.CredentialStoreService/getOperationMetadata',
            CredentialStoreService__pb2.GetOperationsMetadataRequest.SerializeToString,
            CredentialStoreService__pb2.GetOperationsMetadataResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def getNewCustosCredential(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/org.apache.custos.credential.store.service.CredentialStoreService/getNewCustosCredential',
            CredentialStoreService__pb2.GetNewCustosCredentialRequest.SerializeToString,
            CredentialStoreService__pb2.CredentialMetadata.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def getOwnerIdFromToken(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/org.apache.custos.credential.store.service.CredentialStoreService/getOwnerIdFromToken',
            CredentialStoreService__pb2.TokenRequest.SerializeToString,
            CredentialStoreService__pb2.GetOwnerIdResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def getCustosCredentialFromToken(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/org.apache.custos.credential.store.service.CredentialStoreService/getCustosCredentialFromToken',
            CredentialStoreService__pb2.TokenRequest.SerializeToString,
            CredentialStoreService__pb2.CredentialMetadata.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def getCustosCredentialFromClientId(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/org.apache.custos.credential.store.service.CredentialStoreService/getCustosCredentialFromClientId',
            CredentialStoreService__pb2.GetCredentialRequest.SerializeToString,
            CredentialStoreService__pb2.CredentialMetadata.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def getAllCredentialsFromToken(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/org.apache.custos.credential.store.service.CredentialStoreService/getAllCredentialsFromToken',
            CredentialStoreService__pb2.TokenRequest.SerializeToString,
            CredentialStoreService__pb2.GetAllCredentialsResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def getMasterCredentials(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/org.apache.custos.credential.store.service.CredentialStoreService/getMasterCredentials',
            CredentialStoreService__pb2.GetCredentialRequest.SerializeToString,
            CredentialStoreService__pb2.GetAllCredentialsResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def getAllCredentialsFromJWTToken(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/org.apache.custos.credential.store.service.CredentialStoreService/getAllCredentialsFromJWTToken',
            CredentialStoreService__pb2.TokenRequest.SerializeToString,
            CredentialStoreService__pb2.GetAllCredentialsResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def getBasicCredentials(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/org.apache.custos.credential.store.service.CredentialStoreService/getBasicCredentials',
            CredentialStoreService__pb2.TokenRequest.SerializeToString,
            CredentialStoreService__pb2.Credentials.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def createAgentCredential(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/org.apache.custos.credential.store.service.CredentialStoreService/createAgentCredential',
            CredentialStoreService__pb2.CredentialMetadata.SerializeToString,
            CredentialStoreService__pb2.CredentialMetadata.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def getAgentCredential(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/org.apache.custos.credential.store.service.CredentialStoreService/getAgentCredential',
            CredentialStoreService__pb2.GetCredentialRequest.SerializeToString,
            CredentialStoreService__pb2.CredentialMetadata.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def deleteAgentCredential(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/org.apache.custos.credential.store.service.CredentialStoreService/deleteAgentCredential',
            CredentialStoreService__pb2.CredentialMetadata.SerializeToString,
            CredentialStoreService__pb2.OperationStatus.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def getCredentialByAgentBasicAuth(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/org.apache.custos.credential.store.service.CredentialStoreService/getCredentialByAgentBasicAuth',
            CredentialStoreService__pb2.TokenRequest.SerializeToString,
            CredentialStoreService__pb2.GetAllCredentialsResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def getCredentialByAgentJWTToken(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/org.apache.custos.credential.store.service.CredentialStoreService/getCredentialByAgentJWTToken',
            CredentialStoreService__pb2.TokenRequest.SerializeToString,
            CredentialStoreService__pb2.GetAllCredentialsResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def validateAgentJWTToken(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/org.apache.custos.credential.store.service.CredentialStoreService/validateAgentJWTToken',
            CredentialStoreService__pb2.TokenRequest.SerializeToString,
            CredentialStoreService__pb2.OperationStatus.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)
