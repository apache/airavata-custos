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


class ProfileSettings(object):
    # Profile Service Configuration

    def __init__(self):
        self.PROFILE_SERVICE_HOST = '0.0.0.0'
        self.PROFILE_SERVICE_PORT = '8081'


class IAMSettings(object):

    def __init__(self):
        self.KEYCLOAK_AUTHORIZE_URL = 'https://localhost:8443/auth/realms/default/protocol/openid-connect/auth'
        self.KEYCLOAK_TOKEN_URL = 'https://localhost:8443/auth/realms/default/protocol/openid-connect/token'
        self.KEYCLOAK_USERINFO_URL = 'https://localhost:8443/auth/realms/default/protocol/openid-connect/userinfo'
        self.KEYCLOAK_LOGOUT_URL = 'https://localhost:8443/auth/realms/default/protocol/openid-connect/logout'
        self.VERIFY_SSL = False

# Seconds each connection in the pool is able to stay alive. If open connection
# has lived longer than this period, it will be closed.
# (https://github.com/Thriftpy/thrift_connector)
THRIFT_CLIENT_POOL_KEEPALIVE = 5