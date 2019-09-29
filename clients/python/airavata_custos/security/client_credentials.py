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


class ClientCredentials(object):
    """
    This is the base class for passing parameters required to authenticate with keycloak
    """
    def __init__(self, client_id, client_secret):
        """
        This is the constructor for ClientCredentials class
        :param client_id: client identifier received after registering the tenant
        :param client_secret: client password received after registering the tenant
        """
        self.client_id = client_id
        self.client_secret = client_secret


class UserCredentials(ClientCredentials):
    """
    This class inherits from ClientCredentials class. Used for passing parameters required to authenticate user
    with keycloak
    """
    def __init__(self, client_id, client_secret, username, password):
        """
        This is the constructor for UserCredentials class
        :param client_id: client identifier received after registering the tenant
        :param client_secret: client password received after registering the tenant
        :param username: username of the user which needs to be authenticated
        :param password: password of the user which needs to be authenticated
        """
        super().__init__(client_id, client_secret)
        self.username = username
        self.password = password


class IdpCredentials(ClientCredentials):
    """
    This class inherits from ClientCredentials class. Used for passing parameters required to authenticate service
    account with keycloak
    """
    def __init__(self, client_id, client_secret, authorization_code_url, state, redirect_uri):
        """
        This is the constructor for AccountCredentials class
        :param client_id: client identifier received after registering the tenant
        :param client_secret: client password received after registering the tenant
        :param authorization_code_url: The URL that the user will be redirected back from the keycloak to the client
        :param state: An state string for CSRF protection.
        :param redirect_uri: URI for the callback entry point of the client
        """
        super().__init__(client_id, client_secret)
        self.authorization_code_url = authorization_code_url
        self.state = state
        self.redirect_uri = redirect_uri
