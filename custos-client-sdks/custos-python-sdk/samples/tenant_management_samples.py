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
from clients.tenant_management_client import TenantManagementClient

logger = logging.getLogger(__name__)

logger.setLevel(logging.DEBUG)
# create console handler with a higher log level
handler = logging.StreamHandler()
handler.setLevel(logging.DEBUG)

# load APIServerClient with default configuration
client = TenantManagementClient()


def create_tenant():
    contacts = ["8123345687"]
    redirect_uris = ["https://cutos.python/callback"]
    response = client.create_admin_tenant("Custos  Tenant",
                                          "irjanith@gmail.com", "Isuru", "Ranawaka", "irjanith@gmail.com", "Issa",
                                          "1234",
                                          contacts, redirect_uris, "http://custos.lk",
                                          "openid profile email org.cilogon.userinfo", "custos.python",
                                          "http://custos.lk", "Creating for test python SDK")

    print(response)


def get_credentials():
    token = "Y3VzdG9zLXhnZWN0OW90cndhd2E4dXd6dHltLTEwMDAwMDA2Ok9wUWljMWlBNXVOcldJUDNRRGFwa2x6WXZPUDNCeXA1V3ZjZGMyVDU="
    response = client.get_credentials(token=token)
    print(response)


def get_tenant():
    token = "Y3VzdG9zLXhnZWN0OW90cndhd2E4dXd6dHltLTEwMDAwMDA2Ok9wUWljMWlBNXVOcldJUDNRRGFwa2x6WXZPUDNCeXA1V3ZjZGMyVDU="
    client_id = "custos-xgect9otrwawa8uwztym-10000006"
    response = client.get_tenant(token=token, client_id=client_id)
    print(response)


def update_tenant():
    token = "Y3VzdG9zLXhnZWN0OW90cndhd2E4dXd6dHltLTEwMDAwMDA2Ok9wUWljMWlBNXVOcldJUDNRRGFwa2x6WXZPUDNCeXA1V3ZjZGMyVDU="
    client_id = "custos-xgect9otrwawa8uwztym-10000006"
    contacts = ["8123345687"]
    redirect_uris = ["https://cutos.python/callback"]
    response = client.update_tenant(token, client_id, "Custos Python Tenant",
                                    "irjanith@gmail.com", "Janith", "Ranawaka", "irjanith@gmail.com", "Issa",
                                    "1234",
                                    contacts, redirect_uris, "http://custos.lk",
                                    "openid profile email org.cilogon.userinfo", "custos.python",
                                    "http://custos.lk", "Creating for test python SDK")

    print(response)


def add_tenant_roles():
    token = "Y3VzdG9zLXhnZWN0OW90cndhd2E4dXd6dHltLTEwMDAwMDA2Ok9wUWljMWlBNXVOcldJUDNRRGFwa2x6WXZPUDNCeXA1V3ZjZGMyVDU="
    roles = [{"name": "testing", "composite": False, "description": "testing realm"}]
    response = client.add_tenant_roles(token, roles, False)

    print(response)


def add_protocol_mapper():
    token = "Y3VzdG9zLXhnZWN0OW90cndhd2E4dXd6dHltLTEwMDAwMDA2Ok9wUWljMWlBNXVOcldJUDNRRGFwa2x6WXZPUDNCeXA1V3ZjZGMyVDU="
    response = client.add_protocol_mapper(token, "phone_atr", "phone", "phone", "STRING", "USER_ATTRIBUTE", True, True,
                                          True, False, False)

    print(response)


create_tenant()
