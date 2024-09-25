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
import json

from custos.clients.user_management_client import UserManagementClient
from custos.clients.group_management_client import GroupManagementClient
from custos.clients.resource_secret_management_client import ResourceSecretManagementClient
from custos.clients.sharing_management_client import SharingManagementClient

from custos.transport.settings import CustosServerClientSettings
import custos.clients.utils.utilities as utl

from google.protobuf.json_format import MessageToJson

# load root directoty
BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

# get settings file path (settings file path reside in configs folder under home directory)
settings_path = os.path.join(BASE_DIR, 'configs', "settings.ini")

# read settings
custos_settings = CustosServerClientSettings(configuration_file_location=settings_path)

# create custos user management client
user_management_client = UserManagementClient(custos_settings)

# create custos group management client
group_management_client = GroupManagementClient(custos_settings)

# create custos resource secret client
resource_secret_client = ResourceSecretManagementClient(custos_settings)

# create sharing management client
sharing_management_client = SharingManagementClient(custos_settings)

# obtain base 64 encoded token for tenant
b64_encoded_custos_token = utl.get_token(custos_settings=custos_settings)

owner_id = "admin"

created_groups = {}

resource_ids = []


def verifiy_admin_user():
    response = user_management_client.get_user(token=b64_encoded_custos_token, username=owner_id)
    user_management_client.update_user_profile(
        token=b64_encoded_custos_token,
        username=response.username,
        email=response.email,
        first_name=response.first_name,
        last_name=response.last_name)


def register_users(users):
    for user in users:
        print("Registering user: " + user['username'])
        user_management_client.register_user(token=b64_encoded_custos_token,
                                             username=user['username'],
                                             first_name=user['first_name'],
                                             last_name=user['last_name'],
                                             password=user['password'],
                                             email=user['email'],
                                             is_temp_password=False)

        user_management_client.enable_user(token=b64_encoded_custos_token, username=user['username'])


def create_groups(groups):
    for group in groups:
        print("Creating group: " + group['name'])
        grResponse = group_management_client.create_groups(token=b64_encoded_custos_token,
                                                           name=group['name'],
                                                           description=group['description'],
                                                           owner_id=group['owner_id'])
        resp = MessageToJson(grResponse)
        respData = json.loads(resp)
        created_groups[respData['groups'][0]['name']] = respData['groups'][0]['id']


def allocate_users_to_groups(user_group_mapping):
    for usr_map in user_group_mapping:
        group_id = created_groups[usr_map['group_name']]
        print("Assigning user " + usr_map['username'] + " to group " + usr_map['group_name'])
        group_management_client.add_user_to_group(token=b64_encoded_custos_token,
                                                  username=usr_map['username'],
                                                  group_id=group_id,
                                                  membership_type='Member'
                                                  )


def allocate_child_group_to_parent_group(gr_gr_mapping):
    for gr_map in gr_gr_mapping:
        child_id = created_groups[gr_map['child_name']]
        parent_id = created_groups[gr_map['parent_name']]
        print("Assigning child group " + gr_map['child_name'] + " to parent group " + gr_map['parent_name'])
        group_management_client.add_child_group(token=b64_encoded_custos_token,
                                                parent_group_id=parent_id,
                                                child_group_id=child_id)


def create_resource():
    response = resource_secret_client.add_ssh_credential(
        token=b64_encoded_custos_token,
        client_id=custos_settings.CUSTOS_CLIENT_ID,
        owner_id='admin',
        description='Testing SSH Key')
    resource_ids.append(response.token)


def create_permissions(permissions):
    for perm in permissions:
        print("Creating permission " + perm['id'])
        sharing_management_client.create_permission_type(token=b64_encoded_custos_token,
                                                         client_id=custos_settings.CUSTOS_CLIENT_ID,
                                                         id=perm['id'],
                                                         name=perm['name'],
                                                         description=perm['description'])


def create_entity_types(entity_types):
    for type in entity_types:
        print("Creating entity types " + type['id'])
        sharing_management_client.create_entity_type(token=b64_encoded_custos_token,
                                                     client_id=custos_settings.CUSTOS_CLIENT_ID,
                                                     id=type['id'],
                                                     name=type['name'],
                                                     description=type['description'])


def register_resources(resources):
    for resource in resources:
        print("Register resources " + resource['id'])
        sharing_management_client.create_entity(token=b64_encoded_custos_token,
                                                client_id=custos_settings.CUSTOS_CLIENT_ID,
                                                id=resource['id'],
                                                name=resource['name'],
                                                description=resource['description'],
                                                owner_id=resource['user_id'],
                                                type=resource['type'],
                                                parent_id='')


def share_resource_with_user(sharings):
    for shr in sharings:
        print("Sharing entity " + shr['entity_id'] + " with user " + shr['user_id'] + " with permission " + shr[
            'permission_type'])
        sharing_management_client.share_entity_with_users(token=b64_encoded_custos_token,
                                                          client_id=custos_settings.CUSTOS_CLIENT_ID,
                                                          entity_id=shr['entity_id'],
                                                          permission_type=shr['permission_type'],
                                                          user_id=shr['user_id']
                                                          )


def share_resource_with_group(gr_sharings):
    for shr in gr_sharings:
        group_id = created_groups[shr['group_name']]
        print("Sharing entity " + shr['entity_id'] + " with group " + shr['group_name'] + " with permission " + shr[
            'permission_type'])
        sharing_management_client.share_entity_with_groups(token=b64_encoded_custos_token,
                                                           client_id=custos_settings.CUSTOS_CLIENT_ID,
                                                           entity_id=shr['entity_id'],
                                                           permission_type=shr['permission_type'],
                                                           group_id=group_id)


def check_user_permissions(users):
    for user in users:
        access = sharing_management_client.user_has_access(token=b64_encoded_custos_token,
                                                           client_id=custos_settings.CUSTOS_CLIENT_ID,
                                                           entity_id=resource_ids[0],
                                                           permission_type="READ",
                                                           user_id=user['username'])
        usr = user['username']
        print("Access for user " + usr + " : " + str(access))


users = [
    {
        'username': 'UserA',
        'first_name': 'Aaron',
        'last_name': 'Bob',
        'password': '1234',
        'email': 'a@gmail.com'
    },
    {
        'username': 'UserB',
        'first_name': 'Baron',
        'last_name': 'Bob',
        'password': '1234',
        'email': 'b@gmail.com'
    },
    {
        'username': 'UserC',
        'first_name': 'Caron',
        'last_name': 'Bob',
        'password': '1234',
        'email': 'c@gmail.com'
    },
    {
        'username': 'UserD',
        'first_name': 'Daron',
        'last_name': 'Bob',
        'password': '1234',
        'email': 'd@gmail.com'
    },
    {
        'username': 'UserE',
        'first_name': 'Earon',
        'last_name': 'Bob',
        'password': '1234',
        'email': 'e@gmail.com'
    },
    {
        'username': 'UserF',
        'first_name': 'Faron',
        'last_name': 'Bob',
        'password': '1234',
        'email': 'f@gmail.com'
    }
]

groups = [
    {
        'name': 'groupA',
        'description': 'Group L',
        'owner_id': 'admin'
    },
    {
        'name': 'groupB',
        'description': 'Group B',
        'owner_id': 'admin'
    },
    {
        'name': 'groupC',
        'description': 'Group C',
        'owner_id': 'admin'
    }
]

user_group_mapping = [
    {
        'group_name': 'groupA',
        'username': 'UserA'
    },
    {
        'group_name': 'groupA',
        'username': 'UserB'
    },
    {
        'group_name': 'groupB',
        'username': 'UserC'
    },
    {
        'group_name': 'groupB',
        'username': 'UserD'
    },
    {
        'group_name': 'groupC',
        'username': 'UserE'
    },
    {
        'group_name': 'groupC',
        'username': 'UserF'
    }
]

child_gr_parent_gr_mapping = [
    {
        "child_name": "groupB",
        "parent_name": "groupA"
    }
]

permissions = [
    {
        'id': 'OWNER',
        'name': 'OWNER',
        'description': 'Owner permission'
    },
    {
        'id': 'READ',
        'name': 'READ',
        'description': 'Read permission'
    },
    {
        'id': 'WRITE',
        'name': 'WRITE',
        'description': 'WRITE permission'
    }
]
entity_types = [
    {
        'id': 'SECRET',
        'name': 'SECRET',
        'description': 'SECRET Keys'
    }
]

verifiy_admin_user()

# Register users
register_users(users)

# Create groups
create_groups(groups)

# Allocate users to groups
allocate_users_to_groups(user_group_mapping)

# Allocate child groups to parent group
allocate_child_group_to_parent_group(child_gr_parent_gr_mapping)

# Creare resource
create_resource()

# Create permissions
create_permissions(permissions)

# Create entity types
create_entity_types(entity_types)

resources = [
    {
        'id': resource_ids[0],
        'name': 'SECRET',
        'description': 'Register SSH Key Id',
        'user_id': 'admin',
        'type': 'SECRET'
    }
]

sharings = [
    {
        "entity_id": resource_ids[0],
        "permission_type": "READ",
        "type": "SECRET",
        "user_id": "UserF"
    }
]

gr_sharings = [{

    "entity_id": resource_ids[0],
    "permission_type": "READ",
    "type": "SSH_KEY",
    "group_name": 'groupA'
}]

# Register resources
register_resources(resources)

# Share resource with users
share_resource_with_user(sharings)

# Share resource with group
share_resource_with_group(gr_sharings)

# Check user permissions
check_user_permissions(users)
