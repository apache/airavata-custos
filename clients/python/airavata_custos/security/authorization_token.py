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
from airavata_custos import settings
from oauthlib.oauth2 import BackendApplicationClient
from requests_oauthlib import OAuth2Session
from airavata_custos.security.model.ttypes import AuthzToken


def create_authorization_token(client_credentials, tenant_id, username=None):
    client = BackendApplicationClient(client_id=client_credentials.client_id)
    oauth = OAuth2Session(client=client)
    token = oauth.fetch_token(
        token_url=settings.token_url,
        client_id=client_credentials.client_id,
        client_secret=client_credentials.client_secret,
        verify=client_credentials.verify_ssl)

    access_token = token.get('access_token')
    return AuthzToken(
        accessToken=access_token,
        # This is a service account, so leaving out userName for now
        claimsMap={'gatewayID': tenant_id, 'userName': username})
