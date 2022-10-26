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

from custos.server.integration.IdentityManagementService_pb2_grpc import IdentityManagementServiceStub
from custos.server.core.IdentityService_pb2 import AuthenticationRequest, AuthToken, Claim, \
    GetUserManagementSATokenRequest, \
    GetTokenRequest, GetOIDCConfiguration, EndSessionRequest
from custos.server.integration.IdentityManagementService_pb2 import AuthorizationRequest, GetCredentialsRequest, \
    EndSessionRequest as Er, GetAgentTokenRequest
from custos.clients.utils.certificate_fetching_rest_client import CertificateFetchingRestClient

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)


class IdentityManagementClient(object):

    def __init__(self, custos_server_setting):
        self.custos_settings = custos_server_setting
        self.target = self.custos_settings.CUSTOS_SERVER_HOST + ":" + str(self.custos_settings.CUSTOS_SERVER_PORT)
        certManager = CertificateFetchingRestClient(custos_server_setting)
        certManager.load_certificate()
        with open(self.custos_settings.CUSTOS_CERT_PATH, 'rb') as f:
            trusted_certs = f.read()
        self.channel_credentials = grpc.ssl_channel_credentials(root_certificates=trusted_certs)
        self.channel = grpc.secure_channel(target=self.target, credentials=self.channel_credentials)
        self.identity_stub = IdentityManagementServiceStub(self.channel)

    def authenticate(self, token, username, password):
        """
        Used for local authentication
        :param token client credentials
        :param username Users username
        :param password Users password
        :return: Access token
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            request = AuthenticationRequest(username=username, password=password)

            return self.identity_stub.authenticate(request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred in authenticate, probably due to invalid parameters")
            raise

    def is_authenticated(self, token, user_access_token, username):
        """
        Check access token is valid
        :param token: client credential token
        :param user_access_token access token of user
        :param username
        :return: status (TRUE, FALSE)
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            claim = Claim(key="username", value=username)

            claims = []
            claims.append(claim)

            request = AuthToken(accessToken=user_access_token, claims=claims)

            return self.identity_stub.isAuthenticated(request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred in is_authenticated, probably due to invalid parameters")
            raise

    def get_service_account_access_token(self, token):
        """
        Get service account access token
        :param token: client credentials
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            request = GetUserManagementSATokenRequest();

            return self.identity_stub.getUserManagementServiceAccountAccessToken(request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred in get_service_account_access_token, probably due to invalid parameters")
            raise

    def authorize(self, client_id, redirect_uri, response_type, scope, state):
        """
        return authorize url of keycloak
        :param redirect_uri: redirect URI of client
        :param response_type: response type (code)
        :param scope: (openid email profile)
        :param state: (random number)
        :return:
        """
        try:
            request = AuthorizationRequest(client_id=client_id, redirect_uri=redirect_uri, response_type=response_type,
                                           scope=scope,
                                           state=state)

            return self.identity_stub.authorize(request)
        except Exception:
            logger.exception("Error occurred in authorize, probably due to invalid parameters")
            raise

    def token(self, token, redirect_uri=None, code=None, username=None, password=None, refresh_token=None,
              grant_type=None):
        """
        provide user access token
        :param token: client credentials
        :param redirect_uri: redirect uri
        :param code: code returned from token
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            request = GetTokenRequest(redirect_uri=redirect_uri, code=code,
                                      username=username, password=password, refresh_token=refresh_token,
                                      grant_type=grant_type)

            return self.identity_stub.token(request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred in token, probably due to invalid parameters")
            raise

    def get_credentials(self, token, client_id):
        """
        provides IAM and CILogon credentials
        :param token
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            request = GetCredentialsRequest(client_id=client_id)

            return self.identity_stub.getCredentials(request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred in get_credentials, probably due to invalid parameters")
            raise

    def get_oidc_configuration(self, token, client_id):
        """
        send the OIDC config
        :param token client credentials
        :param client_id
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            request = GetOIDCConfiguration(client_id=client_id)

            return self.identity_stub.getOIDCConfiguration(request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred in get_OIDC_Configuration, probably due to invalid parameters")
            raise

    def end_user_session(self, token, refresh_token):
        """
        End user session
        :param token:
        :param refresh_token:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            body = EndSessionRequest(refresh_token=refresh_token)
            request = Er(body=body)

            return self.identity_stub.endUserSession(request=request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred while ending user session")
            raise

    def get_agent_token(self, token, client_id, grant_type, refresh_token=None):
        """
        Get agent token
        :param token: base64Encoded(agentId:agentSec)
        :param client_id: parent Client Id
        :param grant_type: client_credentials, refresh_token
        :param refresh_token:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            request = GetAgentTokenRequest(client_id=client_id, grant_type=grant_type, refresh_token=refresh_token)

            return self.identity_stub.getAgentToken(request=request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred while fetching agent token")
            raise

    def end_agent_session(self, token, refresh_token):
        """
        End user session
        :param token: base64Encoded(agentId:agentSec)
        :param refresh_token:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            body = EndSessionRequest(refresh_token=refresh_token)
            request = Er(body=body)

            return self.identity_stub.endAgentSession(request=request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred while ending agent session")
            raise
