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
from clients.user_management_client import UserManagementClient
from clients.identity_management_client import IdentityManagementClient

logger = logging.getLogger(__name__)

logger.setLevel(logging.DEBUG)
# create console handler with a higher log level
handler = logging.StreamHandler()
handler.setLevel(logging.DEBUG)

# load APIServerClient with default configuration
client = UserManagementClient()
id_client = IdentityManagementClient()


def register_user():
    token = "Y3VzdG9zLXhnZWN0OW90cndhd2E4dXd6dHltLTEwMDAwMDA2Ok9wUWljMWlBNXVOcldJUDNRRGFwa2x6WXZPUDNCeXA1V3ZjZGMyVDU=";
    response = client.register_user(token, "TestingUser", "Jhon", "Smith", "12345", "jhon@iu.edu", True)
    print(response)


def register_and_enable_users():
    token = "Y3VzdG9zLXhnZWN0OW90cndhd2E4dXd6dHltLTEwMDAwMDA2Ok9wUWljMWlBNXVOcldJUDNRRGFwa2x6WXZPUDNCeXA1V3ZjZGMyVDU="
    response = id_client.authenticate(token, "issa", "1234")

    users = [
        {
            "username": "Janith",
            "first_name": "Isuru",
            "last_name": "Ranawaka",
            "password": "1234",
            "email": "irjanith@gmail.com",
            "temporary_password": True,
            "realm_roles": [
                "testing"
            ],
            "client_roles": [
                "owner"
            ],
            "attributes": [
                {
                    "key": "phone",
                    "values": ["8123915386"]
                },
                {
                    "key": "email",
                    "values": ["isjarana@iu.edu"]
                }
            ]
        }
    ]

    response = client.register_and_enable_users(response.accessToken, users)
    print(response)


def add_user_attributes():
    token = "Y3VzdG9zLXhnZWN0OW90cndhd2E4dXd6dHltLTEwMDAwMDA2Ok9wUWljMWlBNXVOcldJUDNRRGFwa2x6WXZPUDNCeXA1V3ZjZGMyVDU="
    response = id_client.authenticate(token, "issa", "1234")
    attributes = [
        {
            "key": "home_phone",
            "values": ["12345678"]
        }
    ]
    users = ["janith"]
    response = client.add_user_attributes(response.accessToken, attributes, users)
    print(response)


def delete_user_attributes():
    token = "Y3VzdG9zLXhnZWN0OW90cndhd2E4dXd6dHltLTEwMDAwMDA2Ok9wUWljMWlBNXVOcldJUDNRRGFwa2x6WXZPUDNCeXA1V3ZjZGMyVDU="
    response = id_client.authenticate(token, "issa", "1234")
    attributes = [
        {
            "key": "home_phone",
            "values": ["12345678"]
        }
    ]
    users = ["janith"]
    response = client.delete_user_attributes(response.accessToken, attributes, users)
    print(response)


def add_roles_to_user():
    token = "Y3VzdG9zLXhnZWN0OW90cndhd2E4dXd6dHltLTEwMDAwMDA2Ok9wUWljMWlBNXVOcldJUDNRRGFwa2x6WXZPUDNCeXA1V3ZjZGMyVDU="
    response = id_client.authenticate(token, "issa", "1234")
    roles = ["testing"]
    users = ["janith"]
    response = client.add_roles_to_users(response.accessToken, users, roles, False)
    print(response)


def find_users():
    token = "Y3VzdG9zLXhnZWN0OW90cndhd2E4dXd6dHltLTEwMDAwMDA2Ok9wUWljMWlBNXVOcldJUDNRRGFwa2x6WXZPUDNCeXA1V3ZjZGMyVDU="
    response = client.find_users(token, 0, 3, username="janith")
    print(response)


find_users()
