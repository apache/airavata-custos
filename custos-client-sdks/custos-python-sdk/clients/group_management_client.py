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

from transport.settings import CustosServerClientSettings

from custos.integration.GroupManagementService_pb2_grpc import GroupManagementServiceStub
from custos.core.IamAdminService_pb2 import GroupRequest, GroupsRequest, UserGroupMappingRequest

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)


class GroupManagementClient(object):

    def __init__(self, configuration_file_location=None):
        self.custos_settings = CustosServerClientSettings(configuration_file_location)
        self.target = self.custos_settings.CUSTOS_SERVER_HOST + ":" + str(self.custos_settings.CUSTOS_SERVER_PORT)
        with open(self.custos_settings.CUSTOS_CERT_PATH, 'rb') as f:
            trusted_certs = f.read()
        self.channel_credentials = grpc.ssl_channel_credentials(root_certificates=trusted_certs)
        self.channel = grpc.secure_channel(target=self.target, credentials=self.channel_credentials)
        self.group_stub = GroupManagementServiceStub(self.channel)

    def create_groups(self, token, groups):
        """
        create groups
        :param token:
        :param groups:
        :return:
        """

        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            return self.agent_stub.enableAgents(metadata=metadata)
        except Exception:
            logger.exception("Error occurred while enabling agents")
            raise

    def update_group(self, token, group):
        """
        update already created group
        :param token:
        :param group:
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

    def delete_group(self, token, id):
        """
        delete group using group id
        :param token:
        :param id:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            attributeList = []
            for atr in agent:
                attribute = UserAttribute(key=atr['key'], values=atr['values'])
                attributeList.append(attribute)
            id = agent['id']
            realm_roles = agent['realm_roles']
            attributes = attributeList
            user = UserRepresentation(id=id, realm_roles=realm_roles, attributes=attributes)
            request = RegisterUserRequest(user=user)

            return self.agent_stub.registerAndEnableAgent(request=request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred while enabling agents")
            raise

    def find_group(self, token, group_name):
        """
        find group using group name
        :param token:
        :param group_name:
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

    def get_all_groups(self, token):
        """
        Get all groups of tenant
        :param token:
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

    def add_user_to_group(self, token, username, group_id):
        """
        Add user to group
        :param token:
        :param username:
        :param group_id:
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

    def remove_user_from_group(self, token, username, group_id):
        """
        Remove user from group
        :param token:
        :param username:
        :param group_id:
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
