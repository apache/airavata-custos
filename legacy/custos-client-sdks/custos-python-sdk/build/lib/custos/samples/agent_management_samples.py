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
from custos.clients.agent_management_client import AgentManagementClient

from custos.transport.settings import CustosServerClientSettings
import custos.clients.utils.utilities as utl

logger = logging.getLogger(__name__)

logger.setLevel(logging.DEBUG)
# create console handler with a higher log level
handler = logging.StreamHandler()
handler.setLevel(logging.DEBUG)

custos_settings = CustosServerClientSettings()
# load APIServerClient with default configuration
client = AgentManagementClient(custos_settings)
id_client = IdentityManagementClient(custos_settings)

token = utl.get_token(custos_settings)


def register_and_enable():
    agent = {
        "id": "agent-asdasda-ebnmvf",
        "realm_roles": [],
        "attributes": [{
            "key": "agent_cluster_id",
            "values": ["123123131"]
        }]
    }
    id_res = id_client.token(token, username="isjarana", password="Custos1234", grant_type="password")
    response = client.register_and_enable_agent(id_res['access_token'], agent)
    print(response)
