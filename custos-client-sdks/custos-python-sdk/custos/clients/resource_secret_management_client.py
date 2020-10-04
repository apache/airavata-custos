#  Licensed to the Apache Software Foundation (ASF) under one or more
#  contributor license agreements.  See the NOTICE file distributed with
#  this work for additional information regarding copyright ownership.
#  The ASF licenses this file to You under the Apache License, Version 2.0
#  (the "License"); you may not use this file except in compliance with
#  the License.  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#


import logging
import grpc

from custos.transport.settings import CustosServerClientSettings

from custos.server.integration.ResourceSecretManagementService_pb2_grpc import ResourceSecretManagementServiceStub
from custos.server.core.IdentityService_pb2 import GetJWKSRequest
from custos.server.core.ResourceSecretService_pb2 import GetSecretRequest, SecretMetadata, ResourceOwnerType, \
    ResourceSource, \
    ResourceType, SSHCredential, PasswordCredential, GetResourceCredentialByTokenRequest
from google.protobuf.json_format import MessageToJson
from custos.clients.utils.certificate_fetching_rest_client import CertificateFetchingRestClient

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)


class ResourceSecretManagementClient(object):

    def __init__(self, custos_server_setting):
        self.custos_settings = custos_server_setting
        self.target = self.custos_settings.CUSTOS_SERVER_HOST + ":" + str(self.custos_settings.CUSTOS_SERVER_PORT)
        certManager = CertificateFetchingRestClient(custos_server_setting)
        certManager.load_certificate()
        with open(self.custos_settings.CUSTOS_CERT_PATH, 'rb') as f:
            trusted_certs = f.read()
        self.channel_credentials = grpc.ssl_channel_credentials(root_certificates=trusted_certs)
        self.channel = grpc.secure_channel(target=self.target, credentials=self.channel_credentials)
        self.resource_sec_client = ResourceSecretManagementServiceStub(self.channel)

    def get_secret(self, token, owner_type, resource_type, source, name):
        """
        Get secret from secret service
        :param token:
        :param owner_type:
        :param resource_type:
        :param source:
        :param name:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            owner_type = ResourceOwnerType.Value(owner_type)
            resource_type = ResourceType.Value(resource_type)
            source = ResourceSource.Value(source)
            met = SecretMetadata(owner_type=owner_type,
                                 resource_type=resource_type, source=source, name=name)
            request = GetSecretRequest(metadata=met)
            msg = self.resource_sec_client.getSecret(request=request, metadata=metadata)
            return MessageToJson(msg)
        except Exception:
            logger.exception("Error occurred while fetching secrets")
            raise

    def get_JWKS(self, token):
        """
        Get JWKS resources
        :param token:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)
            request = GetJWKSRequest()
            msg = self.resource_sec_client.getJWKS(request=request, metadata=metadata)
            return MessageToJson(msg)
        except Exception:
            logger.exception("Error occurred while fetching JWKS request")
            raise

    def add_ssh_credential(self, token, client_id, owner_id, description):

        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)
            secret_metadata = SecretMetadata(client_id=client_id, owner_id=owner_id, description=description)
            ssh_cred = SSHCredential(metadata=secret_metadata)

            return self.resource_sec_client.addSSHCredential(request=ssh_cred, metadata=metadata)

        except Exception:
            logger.exception("Error occurred while creating ssh key")
            raise

    def add_password_credential(self, token, client_id, owner_id, description, password):

        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)
            secret_metadata = SecretMetadata(client_id=client_id, owner_id=owner_id, description=description)
            password_cred = PasswordCredential(metadata=secret_metadata, password=password)

            return self.resource_sec_client.addPasswordCredential(request=password_cred, metadata=metadata)

        except Exception:
            logger.exception("Error occurred while creating password key")
            raise

    def get_ssh_credential(self, token, client_id, ssh_credential_token):

        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)
            request = GetResourceCredentialByTokenRequest(client_id=client_id, token=ssh_credential_token)

            msg = self.resource_sec_client.getSSHCredential(request=request, metadata=metadata)
            return MessageToJson(msg)
        except Exception:
            logger.exception("Error occurred while creating ssh key")
            raise

    def get_password_credential(self, token, client_id, password_credential_token):

        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)
            request = GetResourceCredentialByTokenRequest(client_id=client_id, token=password_credential_token)

            msg = self.resource_sec_client.getPasswordCredential(request=request, metadata=metadata)
            return MessageToJson(msg)
        except Exception:
            logger.exception("Error occurred while creating password key")
            raise
