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

from custos.clients.resource_secret_management_client import ResourceSecretManagementClient

from custos.transport.settings import CustosServerClientSettings
import custos.clients.utils.utilities as utl

# load root directoty
BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

# get settings file path (settings file path reside in configs folder under home directory)
settings_path = os.path.join(BASE_DIR, 'configs', "settings.ini")

# read settings
custos_settings = CustosServerClientSettings(configuration_file_location=settings_path)

# create custos user management client
resource_management_client = ResourceSecretManagementClient(custos_settings)

# obtain base 64 encoded token for tenant
b64_encoded_custos_token = utl.get_token(custos_settings=custos_settings)


def generateSSH_key(owner_id, description):
    response = resource_management_client.add_ssh_credential(token=b64_encoded_custos_token,
                                                             client_id=custos_settings.CUSTOS_CLIENT_ID,
                                                             owner_id=owner_id,
                                                             description=description)
    print(response)
    return response


def getSSHKey(ssh_credential_token):
    response = resource_management_client.get_ssh_credential(token=b64_encoded_custos_token,
                                                             client_id=custos_settings.CUSTOS_CLIENT_ID,
                                                             ssh_credential_token=ssh_credential_token)
    print(response)
    return response


def addPasswordCredential(owner_id, description, password):
    response = resource_management_client.add_password_credential(token=b64_encoded_custos_token,
                                                                  client_id=custos_settings.CUSTOS_CLIENT_ID,
                                                                  owner_id=owner_id,
                                                                  description=description,
                                                                  password=password)
    print(response)
    return response


def getPasswordCredential(password_credential_token):
    response = resource_management_client.get_password_credential(token=b64_encoded_custos_token,
                                                                  client_id=custos_settings.CUSTOS_CLIENT_ID,
                                                                  password_credential_token=password_credential_token)
    print(response)
    return response


generateSSH_key("TestUser5", "My Gateway SSH Key")
addPasswordCredential("TestUser5", "My admin password", "asaxsxasxasx")
getSSHKey("655e8845-9afa-4251-bb0e-c1eb42bec2fc")
getPasswordCredential("a575726a-3e2a-41d1-ad08-06a97dbae903")
