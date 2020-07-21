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
from custos.server.integration.TenantManagementService_pb2_grpc import TenantManagementServiceStub;
from custos.server.core.TenantProfileService_pb2 import GetTenantsRequest, UpdateStatusRequest
from custos.transport.settings import CustosServerClientSettings
from custos.clients.utils.certificate_fetching_rest_client import CertificateFetchingRestClient

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)


class SuperTenantManagementClient(object):

    def __init__(self, custos_server_setting):
        self.custos_settings =custos_server_setting
        self.target = self.custos_settings.CUSTOS_SERVER_HOST + ":" + str(self.custos_settings.CUSTOS_SERVER_PORT)
        certManager = CertificateFetchingRestClient(custos_server_setting)
        certManager.load_certificate()
        with open(self.custos_settings.CUSTOS_CERT_PATH, 'rb') as f:
            trusted_certs = f.read()
        self.channel_credentials = grpc.ssl_channel_credentials(root_certificates=trusted_certs)
        self.channel = grpc.secure_channel(target=self.target, credentials=self.channel_credentials)
        self.tenant_stub = TenantManagementServiceStub(self.channel)

    def get_all_tenants(self, token, offset, limit, status, requester_email=None):
        """
        Get all tenants
        :param token admin user token
        :param offset  omits the initial number of entries
        :param limit  contains maximum number of entries
        :param status (ACTIVE, REQUESTED, DENIED, CANCELLED, DEACTIVATED)
        :return: Tenants
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            request = GetTenantsRequest(offset=offset, limit=limit, status=status, requester_email=None)

            return self.tenant_stub.getAllTenants(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in get_all_tenants, probably due to invalid parameters")
            raise

    def update_tenant_status(self, token, client_id, status):
        """
        Update tenant status.
        :param token admin user token
        :param client_id  client id of tenant to be updated
        :param status (ACTIVE, REQUESTED, DENIED, CANCELLED, DEACTIVATED)
        :return: Operation Status
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            request = UpdateStatusRequest(client_id=client_id, status=status)

            return self.tenant_stub.updateTenantStatus(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in update_tenant_status, probably due to invalid parameters")
            raise
