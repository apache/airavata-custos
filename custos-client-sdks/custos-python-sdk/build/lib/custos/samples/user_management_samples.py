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

token = utl.get_token(custos_settings)

admin_client = SuperTenantManagementClient(custos_settings)


def register_user():
    response = client.register_user(token, "TestingUser", "Jhon", "Smith", "12345", "jhon@iu.edu", True)
    print(response)


def register_and_enable_users():
    response = id_client.authenticate(token, "isjarana", "Custos1234")

    users = [
        {
            "username": "test123",
            "first_name": "user1",
            "last_name": "last",
            "password": "1234",
            "email": "irjanith1@gmail.com",
            "temporary_password": True,
            "realm_roles": [

            ],
            "client_roles": [

            ],
            "attributes": [

            ]
        }
    ]

    response = client.register_and_enable_users(response.accessToken, users)
    print(response)


def add_user_attributes():
    response = id_client.authenticate(token, "isjarana", "Custos1234")
    attributes = [
        {
            "key": "phone",
            "values": ["8123915386"]
        }
    ]
    users = ["janith"]
    response = client.add_user_attributes(response.accessToken, attributes, users)
    print(response)


def delete_user_attributes():
    response = id_client.authenticate(token, "isjarana", "Custos1234")
    attributes = [
        {
            "key": "phone",
            "values": ["8123915386"]
        }
    ]
    users = ["janith"]
    response = client.delete_user_attributes(response.accessToken, attributes, users)
    print(response)


def add_roles_to_user():
    response = id_client.authenticate(token, "issa", "1234")
    roles = ["testing"]
    users = ["janith"]
    response = client.add_roles_to_users(response.accessToken, users, roles, False)
    print(response)


def find_users():
    response = client.find_users(token, 0, 3, username="isjarana")
    print(response)
