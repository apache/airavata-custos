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

import logging
import configparser
from airavata_custos import utils
from airavata_custos.settings import ProfileSettings

logger = logging.getLogger(__name__)


class IAMAdminClient(object):

    def __init__(self, configuration_file_location):
        """
        constructor for IAMAdminClient class
        :param configuration_file_location: takes the location of the ini file containing server configuration
        """
        self.profile_settings = ProfileSettings()
        self._load_settings(configuration_file_location)
        self.iamadmin_client_pool = utils.initialize_iamadmin_client_pool(self.profile_settings.PROFILE_SERVICE_HOST,
                                                                          self.profile_settings.PROFILE_SERVICE_PORT)

    def is_username_available(self, authz_token, username):
        """
        This method validates if the username is available or not
        :param authz_token: Object of AuthzToken class containing access token, username, gatewayId of the active user
        :param username: The username whose availability needs to be verified
        :return: boolean
        """
        return self.iamadmin_client_pool.isUsernameAvailable(authz_token, username)

    def register_user(self, authz_token, username, email_address, first_name, last_name, password):
        """
        This method registers the user with the keycloak instance returns true if successful, false if the registration fails
        :param authz_token: Object of AuthzToken class containing access token, username, gatewayId of the active user
        :param username: The username of the user that needs to be registered
        :param email_address: The email address of the user that needs to be registered
        :param first_name: The first name of the user that needs to be registered
        :param last_name: The last name of the user that needs to be registered
        :param password: The password of the user that needs to be registered
        :return: boolean
        """
        return self.iamadmin_client_pool.registerUser(
            authz_token,
            username,
            email_address,
            first_name,
            last_name,
            password)

    def is_user_enabled(self, authz_token, username):
        """
        Checks the user is enabled/disabled in keycloak. Only the enabled user can login
        :param authz_token: Object of AuthzToken class containing access token, username, gatewayId of the active user
        :param username: The username of the user
        :return: boolean
        """
        return self.iamadmin_client_pool.isUserEnabled(authz_token, username)

    def enable_user(self, authz_token, username):
        """
        The method to enable a disabled user
        :param authz_token: Object of AuthzToken class containing access token, username, gatewayId of the active user
        :param username: The username of the user
        :return: Object of UserProfile class, containing user details
        """
        return self.iamadmin_client_pool.enableUser(authz_token, username)

    def delete_user(self, authz_token, username):
        """
        This method deleted the user from keycloak. Returns true if delete is successful
        :param authz_token: Object of AuthzToken class containing access token, username, gatewayId of the active user
        :param username: The username of the user
        :return: boolean
        """
        return self.iamadmin_client_pool.deleteUser(authz_token, username)

    def is_user_exist(self, authz_token, username):
        """
        This method checks if the user exists in keycloak. Returns true if the user exists otherwise returns false
        :param authz_token: Object of AuthzToken class containing access token, username, gatewayId of the active user
        :param username: The username of the user
        :return: boolean
        """
        try:
            return self.iamadmin_client_pool.isUserExist(authz_token, username)
        except Exception:
            return None

    def get_user(self, authz_token, username):
        """

        :param authz_token: Object of AuthzToken class containing access token, username, gatewayId of the active user
        :param username: username of the user
        :return: object of class UserProfile
        """
        try:
            return self.iamadmin_client_pool.getUser(authz_token, username)
        except Exception:
            return None

    def get_users(self, authz_token, offset=0, limit=-1, search=None):
        """

        :param authz_token: Object of AuthzToken class containing access token, username, gatewayId of the active user
        :param offset: start index
        :param limit: end index
        :param search: search criteria for filtering users
        :return: list of UserProfile class objects
        """
        try:
            return self.iamadmin_client_pool.getUsers(authz_token, offset, limit, search)
        except Exception:
            return None

    def reset_user_password(self, authz_token, username, new_password):
        """

        :param authz_token: Object of AuthzToken class containing access token, username, gatewayId of the active user
        :param username: username of the user
        :param new_password: new password for the user
        :return:
        """
        try:
            return self.iamadmin_client_pool.resetUserPassword(
                authz_token, username, new_password)
        except Exception:
            return None

    def _load_settings(self, configuration_file_location):
        config = configparser.ConfigParser()
        config.read(configuration_file_location)
        settings = config['ProfileServerSettings']
        self.profile_settings.PROFILE_SERVICE_HOST = settings['PROFILE_SERVICE_HOST']
        self.profile_settings.PROFILE_SERVICE_PORT = settings['PROFILE_SERVICE_PORT']
