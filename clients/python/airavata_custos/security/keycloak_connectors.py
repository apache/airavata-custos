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
from oauthlib.oauth2 import LegacyApplicationClient
from requests_oauthlib import OAuth2Session
import requests
from airavata_custos import settings


class KeycloakBackend(object):
    def authenticate(self, client_credentials):
        """This method authenticates a client with keycloak

        Parameters:
        client_credentials (client_credentials): This object has client credentials which needs to be authenticated

        Returns:
        String:token
        String:userInfo
       """
        try:
            if client_credentials.username and client_credentials.password:
                token, userinfo = self._get_token_and_userinfo_password_flow(client_credentials)
            elif client_credentials.refresh_token:
                token = self._get_token_from_refresh_token(client_credentials)
            elif client_credentials.red:
                token, userinfo = self._get_token_and_userinfo_redirect_flow(client_credentials)

            return token, userinfo
        except Exception as e:
            return None

    def _get_token_and_userinfo_password_flow(self, client_credentials):

        oauth2_session = OAuth2Session(client=LegacyApplicationClient(client_id=client_credentials.client_id))
        token = oauth2_session.fetch_token(token_url=settings.KEYCLOAK_TOKEN_URL,
                                           username=client_credentials.username,
                                           password=client_credentials.password,
                                           client_id=client_credentials.client_id,
                                           client_secret=client_credentials.client_secret,
                                           verify=client_credentials.verify_ssl)
        userinfo = oauth2_session.get(settings.KEYCLOAK_USERINFO_URL).json()
        return token, userinfo

    def _get_token_and_userinfo_redirect_flow(self, client_credentials):
        oauth2_session = OAuth2Session(client_credentials.client_id,
                                       scope='openid',
                                       redirect_uri=client_credentials.redirect_uri,
                                       state=client_credentials.state)
        token = oauth2_session.fetch_token(settings.KEYCLOAK_TOKEN_URL,
                                           client_secret=client_credentials.client_secret,
                                           authorization_response=client_credentials.authorization_code_url,
                                           verify=client_credentials.verify_ssl)
        userinfo = oauth2_session.get(settings.KEYCLOAK_USERINFO_URL).json()
        return token, userinfo

    def _get_token_from_refresh_token(self, client_credentials):

        oauth2_session = OAuth2Session(client_credentials.client_id, scope='openid')
        auth = requests.auth.HTTPBasicAuth(client_credentials.client_id, client_credentials.client_secret)
        token = oauth2_session.refresh_token(token_url=settings.KEYCLOAK_TOKEN_URL,
                                             refresh_token=client_credentials.refresh_token,
                                             auth=auth,
                                             verify=client_credentials.verify_ssl)
        return token