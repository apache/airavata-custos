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
import configparser


class Configuration(object):

    def __init__(self, configuration_file_location):
        config = configparser.ConfigParser()
        config.read(configuration_file_location)

        profile_settings = config['ProfileServerSettings']
        self.PROFILE_SERVICE_HOST = profile_settings['PROFILE_SERVICE_HOST']
        self.PROFILE_SERVICE_PORT = profile_settings['PROFILE_SERVICE_PORT']

        keycloak_settings = config['IAMServerSettings']
        self.KEYCLOAK_AUTHORIZE_URL = keycloak_settings['KEYCLOAK_AUTHORIZE_URL']
        self.KEYCLOAK_LOGOUT_URL = keycloak_settings['KEYCLOAK_LOGOUT_URL']
        self.KEYCLOAK_TOKEN_URL = keycloak_settings['KEYCLOAK_TOKEN_URL']
        self.KEYCLOAK_USERINFO_URL = keycloak_settings['KEYCLOAK_USERINFO_URL']
        self.VERIFY_SSL = keycloak_settings.getboolean('VERIFY_SSL')

