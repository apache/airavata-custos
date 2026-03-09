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

from custos.clients.sharing_management_client import SharingManagementClient

from custos.transport.settings import CustosServerClientSettings
import custos.clients.utils.utilities as utl

# load root directoty
BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

# get settings file path (settings file path reside in configs folder under home directory)
settings_path = os.path.join(BASE_DIR, 'configs', "settings.ini")

# read settings
custos_settings = CustosServerClientSettings(configuration_file_location=settings_path)

# create custos user management client
sharing_management_client = SharingManagementClient(custos_settings)

# obtain base 64 encoded token for tenant
b64_encoded_custos_token = utl.get_token(custos_settings=custos_settings)


def create_permission_type(id, name, description):
    response = sharing_management_client.create_permission_type(token=b64_encoded_custos_token,
                                                                client_id=custos_settings.CUSTOS_CLIENT_ID,
                                                                id=id,
                                                                name=name,
                                                                description=description)
    print(response)


def create_entity_type(id, name, description):
    response = sharing_management_client.create_entity_type(token=b64_encoded_custos_token,
                                                            client_id=custos_settings.CUSTOS_CLIENT_ID,
                                                            id=id,
                                                            name=name,
                                                            description=description)
    print(response)


def create_entity(id, name, description, owner_id, type):
    response = sharing_management_client.create_entity(token=b64_encoded_custos_token,
                                                       client_id=custos_settings.CUSTOS_CLIENT_ID,
                                                       id=id,
                                                       name=name,
                                                       description=description,
                                                       owner_id=owner_id,
                                                       type=type,
                                                       parent_id='')
    print(response)


def share_entity_with_user(entity_id, permission_type, user_id):
    response = sharing_management_client.share_entity_with_users(token=b64_encoded_custos_token,
                                                                 client_id=custos_settings.CUSTOS_CLIENT_ID,
                                                                 entity_id=entity_id,
                                                                 permission_type=permission_type,
                                                                 user_id=user_id)
    print(response)


def share_entity_with_group(entity_id, permission_type, group_id):
    response = sharing_management_client.share_entity_with_groups(token=b64_encoded_custos_token,
                                                                  client_id=custos_settings.CUSTOS_CLIENT_ID,
                                                                  entity_id=entity_id,
                                                                  permission_type=permission_type,
                                                                  group_id=group_id)
    print(response)


def check_user_has_access(entity_id, permission_type, user_id):
    response = sharing_management_client.user_has_access(token=b64_encoded_custos_token,
                                                         client_id=custos_settings.CUSTOS_CLIENT_ID,
                                                         entity_id=entity_id,
                                                         permission_type=permission_type,
                                                         user_id=user_id)
    print(response)


# create_permission_type('RW', 'Read and Write', 'Read write permissions')
# create_entity_type("CRED_TOKEN", "Credential Token", "This is credential token")

create_entity('655e8845-9afa-4251-bb0e-c1ebasasx42bec2fcASD', 'Password Token', 'This is password token', 'TestUser5',
              'CRED_TOKEN')
# share_entity_with_user('655e8845-9afa-4251-bb0e-c1eb42bec2fc', 'RW', 'TestUser4')
# share_entity_with_group('655e8845-9afa-4251-bb0e-c1eb42bec2fc', 'RW', '602336d5-e193-41ac-bde6-eb36a73f687e')
#
# check_user_has_access('655e8845-9afa-4251-bb0e-c1eb42bec2fc', 'RW', 'TestUser4')
