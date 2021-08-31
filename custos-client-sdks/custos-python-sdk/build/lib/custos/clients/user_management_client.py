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

from custos.server.integration.UserManagementService_pb2_grpc import UserManagementServiceStub
from custos.server.core.IamAdminService_pb2 import RegisterUserRequest, UserRepresentation, RegisterUsersRequest, \
    UserAttribute, \
    AddUserAttributesRequest, DeleteUserAttributeRequest, UserSearchMetadata, FindUsersRequest, AddUserRolesRequest, \
    UserSearchRequest, ResetUserPassword, DeleteUserRolesRequest
from custos.server.core.UserProfileService_pb2 import UserProfile
from custos.server.integration.UserManagementService_pb2 import LinkUserProfileRequest, UserProfileRequest

from custos.clients.utils.certificate_fetching_rest_client import CertificateFetchingRestClient

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)


class UserManagementClient(object):

    def __init__(self, custos_server_setting):
        self.custos_settings = custos_server_setting
        self.target = self.custos_settings.CUSTOS_SERVER_HOST + ":" + str(self.custos_settings.CUSTOS_SERVER_PORT)
        certManager = CertificateFetchingRestClient(custos_server_setting)
        certManager.load_certificate()
        with open(self.custos_settings.CUSTOS_CERT_PATH, 'rb') as f:
            trusted_certs = f.read()
        self.channel_credentials = grpc.ssl_channel_credentials(root_certificates=trusted_certs)
        self.channel = grpc.secure_channel(target=self.target, credentials=self.channel_credentials)
        self.user_stub = UserManagementServiceStub(self.channel)

    def register_user(self, token, username, first_name, last_name, password, email, is_temp_password):
        """
        Register user in given tenant
        :param token:  client credentials
        :param username:
        :param first_name:
        :param last_name:
        :param password:
        :param email:
        :param is_temp_password:
        :return: registration status
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            user = UserRepresentation(username=username, password=password,
                                      first_name=first_name, last_name=last_name,
                                      email=email, temporary_password=is_temp_password)

            request = RegisterUserRequest(user=user)

            return self.user_stub.registerUser(request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred in register_user, probably due to invalid parameters")
            raise

    def register_and_enable_users(self, token, users):
        """
        register and enable users
        :param token: admin access token
        :param users:
        :return:
        """

        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            usersList = []

            for user in users:
                attributeList = []
                for atr in user['attributes']:
                    attribute = UserAttribute(key=atr['key'], values=atr['values'])
                    attributeList.append(attribute)
                username = user['username']
                password = user['password']
                first_name = user['first_name']
                last_name = user['last_name']
                email = user['email']
                temporary_password = user['temporary_password']
                realm_roles = user['realm_roles']
                client_roles = user['client_roles']
                attributes = attributeList
                user = UserRepresentation(username=username, password=password,
                                          first_name=first_name, last_name=last_name,
                                          email=email, temporary_password=temporary_password,
                                          realm_roles=realm_roles, client_roles=client_roles,
                                          attributes=attributes)

                usersList.append(user)

            request = RegisterUsersRequest(users=usersList)
            return self.user_stub.registerAndEnableUsers(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in register_and_enable_users, probably due to invalid parameters")
            raise

    def add_user_attributes(self, token, attributes, users):
        """
        Add user attributes
        :param token:
        :param attributes:
        :param users:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            attributeList = []
            for atr in attributes:
                attribute = UserAttribute(key=atr['key'], values=atr['values'])
                attributeList.append(attribute)

            request = AddUserAttributesRequest(attributes=attributeList, users=users)
            return self.user_stub.addUserAttributes(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in add_user_attributes, probably due to invalid parameters")
            raise

    def delete_user_attributes(self, token, attributes, users):
        """
        Delete user attributes
        :param token: user token
        :param attributes:
        :param users:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            attributeList = []
            for atr in attributes:
                attribute = UserAttribute(key=atr['key'], values=atr['values'])
                attributeList.append(attribute)

            request = DeleteUserAttributeRequest(attributes=attributeList, users=users)
            return self.user_stub.deleteUserAttributes(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in delete_user_attributes, probably due to invalid parameters")
            raise

    def enable_user(self, token, username):
        """
        Enable user request
        :param token: client credential
        :param attributes:
        :param users:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            user = UserSearchMetadata(username=username)
            request = UserSearchRequest(user=user)
            return self.user_stub.enableUser(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in enable_user, probably due to invalid parameters")
            raise

    def add_roles_to_users(self, token, usernames, roles, is_client_level):
        """
        Add roles to users
        :param token: admin token
        :param usernames list of usersname
        :param : roles list of roles
        :param is_client_level to add client level else realm level
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            request = AddUserRolesRequest(usernames=usernames, roles=roles, client_level=is_client_level)
            return self.user_stub.addRolesToUsers(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in add_roles_to_users, probably due to invalid parameters")
            raise

    def is_user_enabled(self, token, username):
        """
        Check the weather user is enabled
        :param token: client credential
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            user = UserSearchMetadata(username=username)
            request = UserSearchRequest(user=user)
            return self.user_stub.isUserEnabled(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in is_user_enabled, probably due to invalid parameters")
            raise

    def is_username_available(self, token, username):
        """
        Check the weather username  is available
        :param token: client credential
        :param username
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            user = UserSearchMetadata(username=username)
            request = UserSearchRequest(user=user)
            return self.user_stub.isUsernameAvailable(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in is_username_available, probably due to invalid parameters")
            raise

    def get_user(self, token, username):
        """
        Get user
        :param token: client credential
        :param username
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            user = UserSearchMetadata(username=username)
            request = UserSearchRequest(user=user)
            return self.user_stub.getUser(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in get_user, probably due to invalid parameters")
            raise

    def find_users(self, token, offset, limit, username=None, firstname=None, lastname=None, email=None, ):
        """
        Find users
        :param token: client credential
        :param username
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            user = UserSearchMetadata(username=username, first_name=firstname, last_name=lastname, email=email)
            request = FindUsersRequest(user=user, offset=offset, limit=limit)
            return self.user_stub.findUsers(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in find_users, probably due to invalid parameters")
            raise

    def reset_password(self, token, username, password):
        """
        Reset user password
        :param token: client credential
        :param username
        :param password
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            request = ResetUserPassword(username=username, password=password)
            return self.user_stub.resetPassword(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in reset_password, probably due to invalid parameters")
            raise

    def delete_user(self, token, username):
        """
        Delete user from a given realm
        :param token: admin credentials
        :param username:
        :return:
        """

        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            user = UserSearchMetadata(username=username)
            request = UserSearchRequest(user=user)
            return self.user_stub.deleteUser(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in delete_user, probably due to invalid parameters")
            raise

    def delete_user_roles(self, token, username, client_roles, realm_roles):
        """
        Delete user roles
        :param token: admin access token
        :param username:
        :param client_roles:
        :param realm_roles:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            request = DeleteUserRolesRequest(username=username, client_roles=client_roles, roles=realm_roles)
            return self.user_stub.deleteUserRoles(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in delete_user_roles, probably due to invalid parameters")
            raise

    def update_user_profile(self, token, username, email, first_name, last_name):
        """
        Update user profile
        :param token: user token
        :param username:
        :param email:
        :param first_name:
        :param last_name:
        :return:
        """

        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            profile = UserProfile(username=username, email=email, first_name=first_name, last_name=last_name)
            request = UserProfileRequest(user_profile=profile)
            return self.user_stub.updateUserProfile(request=request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in update_user_profile, probably due to invalid parameters")
            raise

    def link_user_profile(self, token, current_username, previous_username, linking_attributes=None):
        """
        Link existing user profile with previous user profile
        :param previous_username:
        :param current_username:
        :param linking_attributes:
        :return:
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            attributeList = []
            for atr in linking_attributes:
                attribute = UserAttribute(key=atr['key'], values=atr['values'])
                attributeList.append(attribute)

            request = LinkUserProfileRequest(current_username=current_username, previous_username=previous_username,
                                             linking_attributes=attributeList)
            return self.user_stub.linkUserProfile(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in update_user_profile, probably due to invalid parameters")
            raise
