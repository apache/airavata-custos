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
from custos.clients.identity_management_client import IdentityManagementClient
from custos.clients.group_management_client import GroupManagementClient

from custos.transport.settings import CustosServerClientSettings
import custos.clients.utils.utilities as utl

logger = logging.getLogger(__name__)

logger.setLevel(logging.DEBUG)
# create console handler with a higher log level
handler = logging.StreamHandler()
handler.setLevel(logging.DEBUG)
custos_settings = CustosServerClientSettings()
# load APIServerClient with default configuration
client = GroupManagementClient(custos_settings)
id_client = IdentityManagementClient(custos_settings)

token = utl.get_token(custos_settings)

print(token)


def create_group():
    groups = [
        {
            "name": "testll",
            "realm_roles": [],
            "client_roles": [],
            "attributes": [{
                "key": "phone",
                "values": ["8123915386"]
            }],
            "sub_groups": [{
                "name": "testlj",
                "realm_roles": [],
                "client_roles": [],
                "attributes": [{
                    "key": "email",
                    "values": ["irjanith@gmail.com"]
                }],
                "sub_groups": []
            }]
        }
    ]
    id_res = id_client.token(token, username="isjarana", password="Custos1234", grant_type="password")
    response = client.create_groups(id_res['access_token'], groups)
    print(response)
