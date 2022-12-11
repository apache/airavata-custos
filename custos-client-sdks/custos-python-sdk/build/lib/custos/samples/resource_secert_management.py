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
from custos.clients.user_management_client import UserManagementClient
from custos.clients.identity_management_client import IdentityManagementClient
from custos.clients.super_tenant_management_client import SuperTenantManagementClient
from custos.clients.resource_secret_management_client import ResourceSecretManagementClient
from google.protobuf.json_format import MessageToDict

from custos.transport.settings import CustosServerClientSettings
import custos.clients.utils.utilities as utl

logger = logging.getLogger(__name__)

logger.setLevel(logging.DEBUG)
# create console handler with a higher log level
handler = logging.StreamHandler()
handler.setLevel(logging.DEBUG)

custos_settings = CustosServerClientSettings()
# load APIServerClient with default configuration
client = UserManagementClient(custos_settings)
id_client = IdentityManagementClient(custos_settings)
resource_secret_client = ResourceSecretManagementClient(custos_settings)

token = utl.get_token(custos_settings)

admin_client = SuperTenantManagementClient(custos_settings)


def user_login():
    response = id_client.token(token=
                               token,
                               username="USERNAME",
                               password="PASSWORD",
                               grant_type="password")
    dict_obj = MessageToDict(response)
    return dict_obj['access_token']


def setKVCredential():
    token = user_login()
    resp = resource_secret_client.set_KV_credential(token=token, user_token=token,
                                                    client_id='CHANGE_ME', key='Your key',
                                                    value='Your Value')
    print(resp)


def getKVCredential():
    token = user_login()
    resp = resource_secret_client.get_KV_credential(token=token, user_token=token,
                                                    client_id='CHANGE_ME', key='Your key')
    print(resp)


def updateKVCredential():
    token = user_login()
    resp = resource_secret_client.update_KV_credential(token=token, user_token=token,
                                                       client_id='CHANGE_ME', key='Your key')
    print(resp)


def deleteKVCredential():
    token = user_login()
    resp = resource_secret_client.delete_KV_credential(token=token, user_token=token,
                                                       client_id='CHANGE_ME', key='Your key')
    print(resp)


