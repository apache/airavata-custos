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
from clients.super_tenant_management_client import SuperTenantManagementClient

logger = logging.getLogger(__name__)

logger.setLevel(logging.DEBUG)
# create console handler with a higher log level
handler = logging.StreamHandler()
handler.setLevel(logging.DEBUG)

# load APIServerClient with default configuration
client = TenantManagementClient()
admin_client = SuperTenantManagementClient()


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


def get_child_tenants():
    token = "Y3VzdG9zLWNzOGp5Y2M4Y3U2NmpuYzJ0c3UzLTEwMDAwMDAyOnNZaDVKSXVuUVEzYU5zRzUzdkMxWlpLckNUOE1KbVJLemJSbXdmbGE="
    response = client.get_child_tenants(token, 0, 5, "ACTIVE")

    print(response)


def get_all_tenants():
    token = "eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJpQmtlN0c1RHBYU0pKZjAyRWswRElHaXNyc1BHOEF5d0dXSURyRE80WVNFIn0.eyJqdGkiOiI1OWI4ZTFhZC01MjViLTRmOWMtYjQwMi0xMTdiNjg0Njk5ODYiLCJleHAiOjE1ODM1MjczOTcsIm5iZiI6MCwiaWF0IjoxNTgzNTI1NTk3LCJpc3MiOiJodHRwczovL2tleWNsb2FrLmN1c3Rvcy5zY2lnYXAub3JnOjMxMDAwL2F1dGgvcmVhbG1zLzEwMDAwMDA2IiwiYXVkIjpbInJlYWxtLW1hbmFnZW1lbnQiLCJhY2NvdW50Il0sInN1YiI6ImE5MTc2YjdiLTNlZDQtNDc2Ny1hMzI3LTQ4MmY0YWZjNWRhNSIsInR5cCI6IkJlYXJlciIsImF6cCI6ImN1c3Rvcy14Z2VjdDlvdHJ3YXdhOHV3enR5bS0xMDAwMDAwNiIsImF1dGhfdGltZSI6MTU4MzUyNTU3MSwic2Vzc2lvbl9zdGF0ZSI6ImQ4YTZjZGFlLTYyYzItNGU3Mi1hYTVlLTMzNWYzZmVkNmY3OCIsImFjciI6IjEiLCJhbGxvd2VkLW9yaWdpbnMiOlsiaHR0cHM6Ly9jdXRvcy5weXRob24iLCJodHRwOi8vY3VzdG9zLmxrIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJvZmZsaW5lX2FjY2VzcyIsImFkbWluIiwidW1hX2F1dGhvcml6YXRpb24iXX0sInJlc291cmNlX2FjY2VzcyI6eyJyZWFsbS1tYW5hZ2VtZW50Ijp7InJvbGVzIjpbInZpZXctcmVhbG0iLCJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsIm1hbmFnZS1pZGVudGl0eS1wcm92aWRlcnMiLCJpbXBlcnNvbmF0aW9uIiwicmVhbG0tYWRtaW4iLCJjcmVhdGUtY2xpZW50IiwibWFuYWdlLXVzZXJzIiwicXVlcnktcmVhbG1zIiwidmlldy1hdXRob3JpemF0aW9uIiwicXVlcnktY2xpZW50cyIsInF1ZXJ5LXVzZXJzIiwibWFuYWdlLWV2ZW50cyIsIm1hbmFnZS1yZWFsbSIsInZpZXctZXZlbnRzIiwidmlldy11c2VycyIsInZpZXctY2xpZW50cyIsIm1hbmFnZS1hdXRob3JpemF0aW9uIiwibWFuYWdlLWNsaWVudHMiLCJxdWVyeS1ncm91cHMiXX0sImFjY291bnQiOnsicm9sZXMiOlsibWFuYWdlLWFjY291bnQiLCJtYW5hZ2UtYWNjb3VudC1saW5rcyIsInZpZXctcHJvZmlsZSJdfX0sInNjb3BlIjoib3BlbmlkIGVtYWlsIHByb2ZpbGUiLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwibmFtZSI6Ikphbml0aCBSYW5hd2FrYSIsInByZWZlcnJlZF91c2VybmFtZSI6Imlzc2EiLCJnaXZlbl9uYW1lIjoiSmFuaXRoIiwiZmFtaWx5X25hbWUiOiJSYW5hd2FrYSIsImVtYWlsIjoiaXJqYW5pdGhAZ21haWwuY29tIn0.DldRtDOypXVjmXjttJYjHucaHjmzDu5DXNqK-xEgAyY5dpnhmsNpqIHmd5nxS1YAQ7wSt-exc8_ZPiZbtOw6GBdzajo81oa6yc2j6fdLZBdnEOfhmb0sH9u2IEFds9W5gpjqOjNZ04BmtwA8Kp3_CTO6k4j6jdlKJDbW0k30niTWf2T-Y-a4H9Xvs8AnrLbj8sjEj7hxNi9yenquObtvfPBR5sUOCZV8Wfbsl_HKp0pt-LjhO9CXR3dnNue8tD5SYUE1Z5rG2ERIiRbPT4ESBWfKJMIbglzcDMhZKz6XEGYzHycxVbTEShvisLqLFNeH_M1qroBT7zJ5IoXLoR9fvQ"
    response = admin_client.get_all_tenants(token, 0, 5, "ACTIVE")
    print(response)

get_all_tenants()
