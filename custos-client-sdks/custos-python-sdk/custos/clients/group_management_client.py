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

from custos.server.integration.GroupManagementService_pb2_grpc import GroupManagementServiceStub
from custos.server.core.IamAdminService_pb2 import GroupRequest, GroupsRequest, UserGroupMappingRequest, \
    UserAttribute, GroupRepresentation

from custos.server.core.UserProfileService_pb2 import GroupToGroupMembership
from custos.clients.utils.certificate_fetching_rest_client import CertificateFetchingRestClient

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)


class GroupManagementClient(object):

    def __init__(self, custos_server_setting):
        self.custos_settings = custos_server_setting
        self.target = self.custos_settings.CUSTOS_SERVER_HOST + ":" + str(self.custos_settings.CUSTOS_SERVER_PORT)
        certManager = CertificateFetchingRestClient(custos_server_setting)
        certManager.load_certificate()
        with open(self.custos_settings.CUSTOS_CERT_PATH, 'rb') as f:
            trusted_certs = f.read()
        self.channel_credentials = grpc.ssl_channel_credentials(root_certificates=trusted_certs)
        self.channel = grpc.secure_channel(target=self.target, credentials=self.channel_credentials)
        self.group_stub = GroupManagementServiceStub(self.channel)

    def create_groups(self, token, name, description, owner_id):
        """
        Create groups
        :param owner_id:
        :param description:
        :param name:
        :param token:
        :return:
        """

        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)
            group_list = []
            rep = GroupRepresentation(name=name, realm_roles=[], client_roles=[],
                                      sub_groups=[], attributes=[], description=description,
                                      ownerId=owner_id)
            group_list.append(rep)
            request = GroupsRequest(groups=group_list)

            return self.group_stub.createGroups(request=request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred while creating groups")
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
            request = GroupRequest(id=id)
            return self.group_stub.deleteGroup(request=request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred while deleting group")
            raise

    def find_group(self, token, group_name, group_id):
        """
        find group using group name
        :param token:
        :param group_name:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            gr = GroupRepresentation(id=group_id, name=group_name)
            request = GroupRequest(id=group_id, group=gr)
            return self.group_stub.findGroup(request=request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred finding group")
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
            request = GroupRequest()
            return self.group_stub.getAllGroups(request=request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred while pulling groups")
            raise

    def add_user_to_group(self, token, username, group_id, membership_type):
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
            request = UserGroupMappingRequest(username=username, group_id=group_id, membership_type=membership_type)
            return self.group_stub.addUserToGroup(request=request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred while adding user to group")
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
            request = UserGroupMappingRequest(username=username, group_id=group_id)
            return self.group_stub.removeUserFromGroup(request=request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred while removing user from group")
            raise

    def add_child_group(self, token, parent_group_id, child_group_id):
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)
            grm = GroupToGroupMembership(child_id=child_group_id, parent_id=parent_group_id)
            return self.group_stub.addChildGroupToParentGroup(request=grm, metadata=metadata)
        except Exception:
            logger.exception("Error occurred while adding child group")
            raise

    def remove_child_group(self, token, parent_group_id, child_group_id):
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)
            grm = GroupToGroupMembership(child_id=child_group_id, parent_id=parent_group_id)
            return self.group_stub.removeChildGroupFromParentGroup(request=grm, metadata=metadata)
        except Exception:
            logger.exception("Error occurred while removing child group from group")
            raise
