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
import grpc

from custos.transport.settings import CustosServerClientSettings

from custos.server.integration.AgentManagementService_pb2_grpc import AgentManagementServiceStub
from custos.server.core.IamAdminService_pb2 import AgentClientMetadata, RegisterUserRequest, \
    UserRepresentation, UserAttribute, AddUserAttributesRequest, DeleteUserAttributeRequest, AddUserRolesRequest, \
    DeleteUserRolesRequest, AddProtocolMapperRequest, ClaimJSONTypes, MapperTypes
from custos.server.integration.AgentManagementService_pb2 import AgentSearchRequest
from custos.clients.utils.certificate_fetching_rest_client import CertificateFetchingRestClient

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)


class AgentManagementClient(object):

    def __init__(self, custos_server_setting):
        self.custos_settings = custos_server_setting
        self.target = self.custos_settings.CUSTOS_SERVER_HOST + ":" + str(self.custos_settings.CUSTOS_SERVER_PORT)
        certManager = CertificateFetchingRestClient(custos_server_setting)
        certManager.load_certificate()
        with open(self.custos_settings.CUSTOS_CERT_PATH, 'rb') as f:
            trusted_certs = f.read()
        self.channel_credentials = grpc.ssl_channel_credentials(root_certificates=trusted_certs)
        self.channel = grpc.secure_channel(target=self.target, credentials=self.channel_credentials)
        self.agent_stub = AgentManagementServiceStub(self.channel)

    def enable_agents(self, token):
        """
        Enable agent registration for realm
        :param token:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            return self.agent_stub.enableAgents(metadata=metadata)
        except Exception:
            logger.exception("Error occurred while enabling agents")
            raise

    def configure_agent_client(self, token, access_token_life_time):
        """
        Configure agent client
        :param token:
        :param access_token_life_time:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            request = AgentClientMetadata(access_token_life_time=access_token_life_time)

            return self.agent_stub.enableAgents(request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred while configuring agent client")
            raise

    def register_and_enable_agent(self, token, agent):
        """
        Register and enable agent
        :param agent:
        :param token:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            attributeList = []
            for atr in agent['attributes']:
                attribute = UserAttribute(key=atr['key'], values=atr['values'])
                attributeList.append(attribute)
            id = agent['id']
            realm_roles = agent['realm_roles']
            user = UserRepresentation(id=id, realm_roles=realm_roles, attributes=attributeList)
            request = RegisterUserRequest(user=user)

            return self.agent_stub.registerAndEnableAgent(request=request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred while enabling agents")
            raise

    def get_agent(self, token, id):
        """
        Get agent having id
        :param token:
        :param id:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            request = AgentSearchRequest(id=id)
            return self.agent_stub.getAgent(request=request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred while fetching agent")
            raise

    def delete_agent(self, token, id):
        """
        Delete agent having id
        :param token:
        :param id:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            request = AgentSearchRequest(id=id)
            return self.agent_stub.deleteAgent(request=request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred while deleting agent")
            raise

    def disable_agent(self, token, id):
        """
        Disable agent having id
        :param token:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            request = AgentSearchRequest(id=id)
            return self.agent_stub.disableAgent(request=request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred while disabling agent")
            raise

    def enable_agent(self, token, id):
        """
        Enable agent having id
        :param token:
        :param id:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            request = AgentSearchRequest(id=id)
            return self.agent_stub.enableAgent(request=request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred while enabling agent")
            raise

    def add_agent_attributes(self, token, agents, attributes):
        """
        Add attributes to agents
        :param token:
        :param agents:
        :param attributes:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)
            attributeList = []
            for atr in attributes:
                attribute = UserAttribute(key=atr['key'], values=atr['values'])
                attributeList.append(attribute)

            request = AddUserAttributesRequest(attributes=attributeList, agents=agents)
            return self.agent_stub.addAgentAttributes(request=request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred while adding agent attributes")
            raise

    def delete_agent_attributes(self, token, agents, attributes):
        """
        Delete agent attributes of agents
        :param token:
        :param agents:
        :param attributes:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)
            attributeList = []
            for atr in attributes:
                attribute = UserAttribute(key=atr['key'], values=atr['values'])
                attributeList.append(attribute)

            request = DeleteUserAttributeRequest(attributes=attributeList, agents=agents)
            return self.agent_stub.deleteAgentAttributes(request=request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred while deleting agent attributes")
            raise

    def add_roles_to_agents(self, token, agents, roles):
        """
        Add roles to agents
        :param token:
        :param agents:
        :param roles:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)
            request = AddUserRolesRequest(agents=agents, roles=roles)
            return self.agent_stub.addRolesToAgent(request=request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred whiling adding roles to agents")
            raise

    def delete_roles_from_agent(self, token, id, roles):
        """
        Delete roles from agent
        :param token:
        :param id:
       :param roles:
       :return:
       """

        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)
            request = DeleteUserRolesRequest(id=id, roles=roles)
            return self.agent_stub.deleteRolesFromAgent(request=request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred while enabling agents")
            raise

    def add_protocol_mapper(self, token, name, attribute_name, claim_name, claim_type, mapper_type,
                            add_to_id_token, add_to_access_token, add_to_user_info, multi_valued,
                            aggregate_attribute_values):
        """
        Add protocol mapper to agent client
        :param token:
        :param name:
        :param attribute_name:
        :param claim_name:
        :param claim_type:
        :param mapper_type:
        :param add_to_id_token:
        :param add_to_access_token:
        :param add_to_user_info:
        :param multi_valued:
        :param aggregate_attribute_values:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            wrapped_json_type = ClaimJSONTypes.Value(claim_type)
            wrapped_mapper_type = MapperTypes.Value(mapper_type)

            request = AddProtocolMapperRequest(name=name,
                                               attribute_name=attribute_name,
                                               claim_name=claim_name,
                                               claim_type=wrapped_json_type,
                                               mapper_type=wrapped_mapper_type,
                                               add_to_id_token=add_to_id_token,
                                               add_to_access_token=add_to_access_token,
                                               add_to_user_info=add_to_user_info,
                                               multi_valued=multi_valued,
                                               aggregate_attribute_values=aggregate_attribute_values
                                               )

            return self.agent_stub.addProtocolMapper(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in add_protocol_mapper, probably due to invalid parameters")
            raise
