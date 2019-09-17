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
import thrift_connector.connection_pool as connection_pool
from thrift.protocol import TBinaryProtocol
from thrift.protocol.TMultiplexedProtocol import TMultiplexedProtocol
from thrift.transport import TSocket, TSSLSocket, TTransport

from custos.profile.iam.admin.services.cpi import IamAdminServices, constants
from airavata_custos import settings

log = logging.getLogger(__name__)


class MultiplexThriftClientMixin:
    service_name = None

    @classmethod
    def get_protoco_factory(cls):
        def factory(transport):
            protocol = TBinaryProtocol.TBinaryProtocol(transport)
            multiplex_prot = TMultiplexedProtocol(protocol, cls.service_name)
            return multiplex_prot
        return factory


class CustomThriftClient(connection_pool.ThriftClient):
    secure = False
    validate = False

    @classmethod
    def get_socket_factory(cls):
        if not cls.secure:
            return super().get_socket_factory()
        else:
            def factory(host, port):
                return TSSLSocket.TSSLSocket(host, port, validate=cls.validate)
            return factory

    def ping(self):
        try:
            self.client.getAPIVersion()
        except Exception as e:
            log.debug("getAPIVersion failed: {}".format(str(e)))
            raise


class IAMAdminServiceThriftClient(MultiplexThriftClientMixin,
                                  CustomThriftClient):
    service_name = constants.IAM_ADMIN_SERVICES_CPI_NAME
    secure = True


def initialize_iamadmin_client_pool(host, port):
    iamadmin_client_pool = connection_pool.ClientPool(
        IamAdminServices,
        host,
        port,
        connection_class=IAMAdminServiceThriftClient,
        keepalive=settings.THRIFT_CLIENT_POOL_KEEPALIVE
    )
    return iamadmin_client_pool

