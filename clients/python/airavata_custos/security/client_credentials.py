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

    Attributes:
        client_id:  Client ID, which you get from tenant registration with Keycloak
        client_secret: Client Secret, which you get from tenant registration with Keycloak
        username:  Username of the tenant user that needs to be authenticated
        password:  Password of the tenant user that needs to be authenticated
        authorization_code_url: URL of the authorization server’s authorization endpoint
        state:
        redirect_uri: Redirect URI you registered as callback
        refresh_token:
        verify_ssl: Flag to indicate ssl verification is required

    """
    client_id = None
    client_secret = None
    verify_ssl = None
    authorization_code_url = None
    state = None
    redirect_uri = None
    username = None
    password = None
    refresh_token = None

    def __init__(self, client_id, client_secret, verify_ssl=False, authorization_code_url=None, state=None, redirect_uri=None, username=None, password=None, refresh_token=None):
        self.client_id = client_id
        self.client_secret = client_secret
        self.verify_ssl = verify_ssl
        self.authorization_code_url = authorization_code_url
        self.state = state
        self.redirect_uri = redirect_uri
        self.username = username
        self.password = password
        self.refresh_token = refresh_token


