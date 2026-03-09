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

import os

from custos.clients.identity_management_client import IdentityManagementClient

from custos.transport.settings import CustosServerClientSettings
import custos.clients.utils.utilities as utl

# load APIServerClient with default configuration
BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

settings_path = os.path.join(BASE_DIR, 'configs', "settings.ini")

# logger.info(settings_path)
custos_settings = CustosServerClientSettings(configuration_file_location=settings_path)

id_client = IdentityManagementClient(custos_settings)

b64_encoded_custos_token = utl.get_token(custos_settings=custos_settings)


def login(username, password):
    resp = id_client.authenticate(token=b64_encoded_custos_token, username=username, password=password)
    print(resp)


def obtain_access_token_from_code(redirect_uri, code):
    resp = id_client.token(token=b64_encoded_custos_token, redirect_uri=redirect_uri, code=code,
                           grant_type="authorization_code")
    print(resp)


def get_oidc_configuration(client_id):
    response = id_client.get_oidc_configuration(token=b64_encoded_custos_token, client_id=client_id)
    print(response)


def logout(refresh_token):
    response = id_client.end_user_session(token=b64_encoded_custos_token, refresh_token=refresh_token)
    print(response)

