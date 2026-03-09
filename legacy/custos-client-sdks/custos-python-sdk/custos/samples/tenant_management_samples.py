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
from custos.clients.tenant_management_client import TenantManagementClient
from custos.clients.super_tenant_management_client import SuperTenantManagementClient
from custos.clients.identity_management_client import IdentityManagementClient
from custos.transport.settings import CustosServerClientSettings
import custos.clients.utils.utilities as utl

logger = logging.getLogger(__name__)

logger.setLevel(logging.DEBUG)
# create console handler with a higher log level
handler = logging.StreamHandler()
handler.setLevel(logging.DEBUG)

custos_settings = CustosServerClientSettings()
# load APIServerClient with default configuration
client = TenantManagementClient(custos_settings)
admin_client = SuperTenantManagementClient(custos_settings)
id_client = IdentityManagementClient(custos_settings)

token = utl.get_token(custos_settings)


def create_tenant():
    contacts = ["2345634324"]
    redirect_uris = ["http://localhost:8080,http://localhost:8080/user/external_ids"]
    response = client.create_admin_tenant("SAMPLE",
                                          "XXX@iu.edu", "First Name", "LastName", "email", "admin",
                                          "1234",
                                          contacts, redirect_uris, "https://domain.org/",
                                          "openid profile email org.cilogon.userinfo", "domain.org",
                                          "https://domain.org/static/favicon.png", "Galaxy Portal")
    print(response)


def get_tenant():
    client_id = "custos-8p4baddxvbiusmjorjch-10000401"
    response = client.get_tenant(client_token=token, client_id=client_id)
    print(response)


def update_tenant():
    client_id = "custos-6nwoqodstpe5mvcq09lh-10000101"
    contacts = ["8123915386"]
    redirect_uris = ["https://custos.scigap.org/callback ", "http://127.0.0.1:8000/auth/callback/",
                     "http://127.0.0.1:8000/"]
    response = client.update_tenant(token, client_id, "Custos Portal",
                                    "irjanith@gmail.com", "Isuru", "Ranawaka", "irjanith@gmail.com", "isjarana",
                                    "Custos1234",
                                    contacts, redirect_uris, "https://custos.scigap.org/",
                                    "openid profile email org.cilogon.userinfo", "domain.org",
                                    "https://custos.scigap.org/", "Custos Portal")
    print(response)


def add_tenant_roles():
    roles = [{"name": "testing", "composite": False, "description": "testing realm"}]
    response = client.add_tenant_roles(token, roles, False)
    print(response)


def add_protocol_mapper():
    response = client.add_protocol_mapper(token, "phone_atr", "phone", "phone", "STRING", "USER_ATTRIBUTE", True, True,
                                          True, False, False)
    print(response)


def get_child_tenants():
    response = client.get_child_tenants(token, 0, 5, "ACTIVE")
    print(response)


def get_all_tenants():
    response = admin_client.get_all_tenants(token, 0, 5, "ACTIVE")
    print(response)


def delete_tenant():
    response = client.delete_tenant(token, "custos-pv3fqfs9z1hps0xily2t-10000000")
    print(response)
