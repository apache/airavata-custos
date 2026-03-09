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
from custos.transport.settings import CustosServerClientSettings
import custos.clients.utils.utilities as utl

logger = logging.getLogger(__name__)

logger.setLevel(logging.DEBUG)
# create console handler with a higher log level
handler = logging.StreamHandler()
handler.setLevel(logging.DEBUG)

custos_settings = CustosServerClientSettings()
# load IdentityManagementClient with default configuration
client = IdentityManagementClient(custos_settings)

main_token = utl.get_token(custos_settings)


def authenticate():
    response = client.authenticate(main_token, "isjarana", "Custos1234")
    print(response)


def is_authenticated():
    access_token = client.authenticate(main_token, "issa", "1234")
    response = client.is_authenticated(main_token, access_token.accessToken, "issa")
    print(response)


def get_user_management_access_token():
    response = client.get_service_account_access_token(token)
    print(response)


def authorize():
    response = client.authorize("custos-xgect9otrwawa8uwztym-10000006", "http://custos.lk", "code",
                                "openid email profile", "asdadasdewde")
    print(response)


def token():
    response = client.token(main_token, "http://custos.lk", "asdasdasdadasd")
    print(response)


def get_credentials():
    response = client.get_credentials(main_token, "custos-xgect9otrwawa8uwztym-10000006")
    print(response)


def get_OIDC_config():
    response = client.get_oidc_configuration(main_token, "custos-pv3fqfs9z1hps0xily2t-10000000")
    print(response)
