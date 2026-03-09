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

from custos.clients.group_management_client import GroupManagementClient

from custos.transport.settings import CustosServerClientSettings
import custos.clients.utils.utilities as utl

# load root directoty
BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

# get settings file path (settings file path reside in configs folder under home directory)
settings_path = os.path.join(BASE_DIR, 'configs', "settings.ini")

# read settings
custos_settings = CustosServerClientSettings(configuration_file_location=settings_path)

# create custos user management client
group_management_client = GroupManagementClient(custos_settings)

# obtain base 64 encoded token for tenant
b64_encoded_custos_token = utl.get_token(custos_settings=custos_settings)


def create_group(name, description, owner_id):
    response = group_management_client.create_groups(token=b64_encoded_custos_token, name=name,
                                                     description=description,
                                                     owner_id=owner_id)
    print(response)
    return response


def add_user_to_group(username, group_id, membership_type):
    response = group_management_client.add_user_to_group(token=b64_encoded_custos_token,
                                                         username=username,
                                                         group_id=group_id,
                                                         membership_type=membership_type)
    print(response)


def add_child_group_to_parent_group(parent_group_id, child_group_id):
    response = group_management_client.add_child_group(token=b64_encoded_custos_token, parent_group_id=parent_group_id,
                                                       child_group_id=child_group_id)
    print(response)


def remove_child_group(parent_group_id, child_group_id):
    response = group_management_client.add_child_group(token=b64_encoded_custos_token, parent_group_id=parent_group_id,
                                                       child_group_id=child_group_id)
    print(response)


create_group("Group A", "Paren group", "TestUser4")
create_group("Group B", "Child group", "TestUser4")

add_user_to_group("Testuser5", "602336d5-e193-41ac-bde6-eb36a73f687e", "Member")

add_child_group_to_parent_group("8b0f8241-e995-496e-a4f5-bdbde4235215", "602336d5-e193-41ac-bde6-eb36a73f687e")
remove_child_group("8b0f8241-e995-496e-a4f5-bdbde4235215", "602336d5-e193-41ac-bde6-eb36a73f687e")
