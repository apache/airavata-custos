#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
import time
from oauthlib.oauth2 import LegacyApplicationClient
import requests
import configparser
from airavata_custos.settings import IAMSettings
from oauthlib.oauth2 import BackendApplicationClient
from requests_oauthlib import OAuth2Session
from custos.commons.model.security.ttypes import AuthzToken
from urllib.parse import quote


class KeycloakBackend(object):

    def __init__(self, configuration_file_location):
        """
        constructor for KeycloakBackend class
        :param configuration_file_location: takes the location of the ini file containing server configuration
        """
        self.keycloak_settings = IAMSettings()
        self._load_settings(configuration_file_location)

    def authenticate_user(self, user_credentials):
        """
        Method to authenticate a gateway user with keycloak
        :param user_credentials: object of UserCredentials class
        :return: openid token, openid user information
        """
        try:
            token, user_info = self._get_token_and_user_info_password_flow(user_credentials)
            return token, user_info
        except Exception as e:
            return None

    def authenticate_account(self, account_credentials):
        """

        :param account_credentials: object of AccountCredentials class
        :return: openid token, openid user information
        """
        try:
            token, account_info = self._get_token_and_user_info_redirect_flow(account_credentials)
            return token, account_info
        except Exception as e:
            return None

    def authenticate_using_refresh_token(self, client_credentials, refresh_token):
        """

        :param client_credentials: object of ClientCredentials class
        :param refresh_token: openid connect refresh token
        :return: openid token
        """
        try:
            token = self._get_token_from_refresh_token(client_credentials, refresh_token)
            return token
        except Exception as e:
            return None

    def idp_login(self, client_id, redirect_uri, idp_alias):
        """

        :param client_id: client identifier received after registering the tenant
        :param redirect_uri: redirect url
        :param idp_alias: idp to which redirection has to be made
        :return:authorization_url, redirect_uri, state
        """
        redirect_uri += '?idp_alias=' + quote(idp_alias)
        base_authorize_url = self.keycloak_settings.KEYCLOAK_AUTHORIZE_URL
        oauth2_session = OAuth2Session(client_id, scope='openid', redirect_uri=redirect_uri)
        authorization_url, state = oauth2_session.authorization_url(base_authorize_url)
        authorization_url += '&kc_idp_hint=' + quote(idp_alias)
        return authorization_url, redirect_uri, state

    def get_authorization_token(self, client_credentials, tenant_id, username=None):
        """
        This method created a authorization token for the user or a service account
        In case of a service account username will be null
        :param client_credentials: object of class client_credentials
        :param tenant_id: gateway id of the client
        :param username: username of the user for which authorization token is being created
        :return: AuthzToken
        """
        client = BackendApplicationClient(client_id=client_credentials.client_id)
        oauth = OAuth2Session(client=client)
        token = oauth.fetch_token(
            token_url=self.keycloak_settings.KEYCLOAK_TOKEN_URL,
            client_id=client_credentials.client_id,
            client_secret=client_credentials.client_secret,
            verify=client_credentials.verify_ssl)

        access_token = token.get('access_token')
        return AuthzToken(
            accessToken=access_token,
            claimsMap={'gatewayID': tenant_id, 'userName': username})

    def _get_token_and_user_info_password_flow(self, client_credentials):

        oauth2_session = OAuth2Session(client=LegacyApplicationClient(client_id=client_credentials.client_id))
        token = oauth2_session.fetch_token(token_url=self.keycloak_settings.KEYCLOAK_TOKEN_URL,
                                           username=client_credentials.username,
                                           password=client_credentials.password,
                                           client_id=client_credentials.client_id,
                                           client_secret=client_credentials.client_secret,
                                           verify=self.keycloak_settings.VERIFY_SSL)
        user_info = oauth2_session.get(self.keycloak_settings.KEYCLOAK_USERINFO_URL).json()
        return token, user_info

    def _get_token_and_user_info_redirect_flow(self, client_credentials):
        oauth2_session = OAuth2Session(client_credentials.client_id,
                                       scope='openid',
                                       redirect_uri=client_credentials.redirect_uri,
                                       state=client_credentials.state)
        token = oauth2_session.fetch_token(self.keycloak_settings.KEYCLOAK_TOKEN_URL,
                                           client_secret=client_credentials.client_secret,
                                           authorization_response=client_credentials.authorization_code_url,
                                           verify=self.keycloak_settings.VERIFY_SSL)
        user_info = oauth2_session.get(self.keycloak_settings.KEYCLOAK_USERINFO_URL).json()
        return token, user_info

    def _get_token_from_refresh_token(self, client_credentials, refresh_token):

        oauth2_session = OAuth2Session(client_credentials.client_id, scope='openid')
        auth = requests.auth.HTTPBasicAuth(client_credentials.client_id, client_credentials.client_secret)
        token = oauth2_session.refresh_token(token_url=self.keycloak_settings.KEYCLOAK_TOKEN_URL,
                                             refresh_token=refresh_token,
                                             auth=auth,
                                             verify=self.keycloak_settings.VERIFY_SSL)
        return token

    def _load_settings(self, configuration_file_location):
        config = configparser.ConfigParser()
        config.read(configuration_file_location)
        settings = config['IAMSeverSettings']
        self.keycloak_settings.KEYCLOAK_AUTHORIZE_URL = settings['KEYCLOAK_AUTHORIZE_URL']
        self.keycloak_settings.KEYCLOAK_LOGOUT_URL = settings['KEYCLOAK_LOGOUT_URL']
        self.keycloak_settings.KEYCLOAK_TOKEN_URL = settings['KEYCLOAK_TOKEN_URL']
        self.keycloak_settings.KEYCLOAK_USERINFO_URL = settings['KEYCLOAK_USERINFO_URL']
        self.keycloak_settings.VERIFY_SSL = settings.getboolean('VERIFY_SSL')

