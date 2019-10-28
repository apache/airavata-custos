#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

from airavata_custos import utils
from custos.commons.model.security.ttypes import AuthzToken
from custos.profile.model.User.ttypes import UserProfile
from airavata_custos.configuration import Configuration


class Profile(object):

    def __init__(self, configuration: Configuration):
        """

        :param configuration: object of class Configuration which has hosts/ports of custos services
        """
        self.profile_service_pool = utils.initialize_userprofile_client_pool(configuration.PROFILE_SERVICE_HOST,
                                                                             configuration.PROFILE_SERVICE_PORT)

    def create_user(self, authorization_token: AuthzToken) -> UserProfile:
        """
        This method creates a new user in custos profile service
        :param authorization_token: object of class AuthzToken
        :return: boolean true if user is created successfully otherwise false
        """
        return self.profile_service_pool.initializeUserProfile(authorization_token)

    def update_user(self, authorization_token: AuthzToken, user_profile: UserProfile) -> bool:
        """
        This method updates the user in custos profile service
        :param authorization_token: object of class AuthzToken
        :param user_profile: updated user info
        :param tenant: tenant identifier to which the user belongs
        :return: boolean true if user is updated successfully otherwise false
        """
        return self.profile_service_pool.updateUserProfile(authorization_token, user_profile)

    def delete_user(self, authorization_token: AuthzToken, user_name: str, tenant: str) -> bool:
        """
        This method deletes the user in the custos profile service
        :param authorization_token: object of class AuthzToken
        :param user_name: unique identifier of the user
        :param tenant: tenant identifier to which the user belongs
        :return: boolean true if user is deleted successfully otherwise false
        """
        return self.profile_service_pool.deleteUserProfile(authorization_token, user_name, tenant)

    def get_user(self, authorization_token: AuthzToken, user_name: str, tenant: str) -> UserProfile:
        """
        To retrieve user info
        :param authorization_token:  object of class AuthzToken
        :param user_name: unique identifier of the user
        :param tenant: tenant identifier to which the user belongs
        :return: object of class User
        """
        return self.profile_service_pool.getUserProfileById(authorization_token, user_name, tenant)

    def get_all_users(self, authorization_token: AuthzToken, tenant: str, offset: int = 0, limit: int = -1) -> list:
        """
        To retrieve all users in the tenant
        :param authorization_token: object of class AuthzToken
        :param tenant: tenant identifier
        :param offset: to limit the number of users
        :param limit: to limit the number of users
        :return: list of users
        """
        return self.profile_service_pool.getAllUserProfilesInGateway(authorization_token, tenant, offset, limit)
