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
from custos.server.integration.TenantManagementService_pb2_grpc import TenantManagementServiceStub
from custos.server.core.TenantProfileService_pb2 import Tenant, GetTenantsRequest, GetAllTenantsForUserRequest
from custos.server.core.IamAdminService_pb2 import AddRolesRequest, RoleRepresentation, AddProtocolMapperRequest, \
    ClaimJSONTypes, MapperTypes
from custos.server.integration.TenantManagementService_pb2 import GetTenantRequest, \
    UpdateTenantRequest, DeleteTenantRequest
from custos.transport.settings import CustosServerClientSettings
from custos.clients.utils.certificate_fetching_rest_client import CertificateFetchingRestClient

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)


class TenantManagementClient(object):

    def __init__(self, custos_server_setting):
        self.custos_settings = custos_server_setting
        self.target = self.custos_settings.CUSTOS_SERVER_HOST + ":" + str(self.custos_settings.CUSTOS_SERVER_PORT)
        certManager = CertificateFetchingRestClient(custos_server_setting)
        certManager.load_certificate()
        with open(self.custos_settings.CUSTOS_CERT_PATH, 'rb') as f:
            trusted_certs = f.read()
        self.channel_credentials = grpc.ssl_channel_credentials(root_certificates=trusted_certs)
        self.channel = grpc.secure_channel(target=self.target, credentials=self.channel_credentials)
        self.tenant_stub = TenantManagementServiceStub(self.channel)

    def create_admin_tenant(self, client_name, requester_email, admin_frist_name,
                            admin_last_name, admin_email, admin_username, admin_password,
                            contacts, redirect_uris, client_uri, scope, domain, logo_uri, comment):
        """
         Creates admin tenant client. Needs to be approved by
         Custos Admin
        :return: Custos Credentials
        """
        try:
            tenant = Tenant(client_name=client_name,
                            requester_email=requester_email,
                            admin_first_name=admin_frist_name,
                            admin_last_name=admin_last_name,
                            admin_email=admin_email,
                            admin_username=admin_username,
                            admin_password=admin_password,
                            contacts=contacts,
                            redirect_uris=redirect_uris,
                            client_uri=client_uri,
                            scope=scope,
                            domain=domain,
                            logo_uri=logo_uri,
                            comment=comment,
                            application_type="web")
            return self.tenant_stub.createTenant(tenant)
        except Exception:
            logger.exception("Error occurred in create_admin_tenant, probably due to invalid parameters")
            raise

    def create_tenant(self, client_token, client_name, requester_email, admin_frist_name,
                      admin_last_name, admin_email, admin_username, admin_password,
                      contacts, redirect_uris, client_uri, scope, domain, logo_uri, comment):
        """
        Creates child tenant under admin tenant. Automatically activates
        :return: Custos credentials
        """
        try:
            tenant = Tenant(client_name=client_name,
                            requester_email=requester_email,
                            admin_first_name=admin_frist_name,
                            admin_last_name=admin_last_name,
                            admin_email=admin_email,
                            admin_username=admin_username,
                            admin_password=admin_password,
                            contacts=contacts,
                            redirect_uris=redirect_uris,
                            client_uri=client_uri,
                            scope=scope,
                            domain=domain,
                            logo_uri=logo_uri,
                            comment=comment,
                            application_type="web")
            token = "Bearer " + client_token
            metadata = (('authorization', token),)
            return self.tenant_stub.createTenant(tenant, metadata=metadata)
        except Exception:
            logger.exception("Error occurred in create_tenant, probably due to invalid parameters")
            raise

    def get_tenant(self, client_token, client_id):
        """
        Fetch tenant
        :return: Tenant
        """
        try:
            request = GetTenantRequest(client_id=client_id)
            token = "Bearer " + client_token
            metadata = (('authorization', token),)
            return self.tenant_stub.getTenant(request, metadata=metadata)
        except Exception:
            logger.exception("Error occurred in get_tenant, probably due to invalid parameters")
            raise

    def update_tenant(self, client_token, client_id, client_name, requester_email, admin_frist_name,
                      admin_last_name, admin_email, admin_username, admin_password,
                      contacts, redirect_uris, client_uri, scope, domain, logo_uri, comment):
        """
        Update given tenant by client Id
        :return: updated tenant
        """
        try:
            tenant = Tenant(client_name=client_name,
                            requester_email=requester_email,
                            admin_first_name=admin_frist_name,
                            admin_last_name=admin_last_name,
                            admin_email=admin_email,
                            admin_username=admin_username,
                            admin_password=admin_password,
                            contacts=contacts,
                            redirect_uris=redirect_uris,
                            client_uri=client_uri,
                            scope=scope,
                            domain=domain,
                            logo_uri=logo_uri,
                            comment=comment,
                            application_type="web")
            token = "Bearer " + client_token
            metadata = (('authorization', token),)

            request = UpdateTenantRequest(client_id=client_id, body=tenant)

            return self.tenant_stub.updateTenant(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in update_tenant, probably due to invalid parameters")
            raise

    def delete_tenant(self, token, client_id):
        """
        Delete given tenant by client Id
        :return:  void
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            request = DeleteTenantRequest(client_id=client_id)

            return self.tenant_stub.deleteTenant(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in delete_tenant, probably due to invalid parameters")
            raise

    def add_tenant_roles(self, token, roles, is_client_level):
        """
        :param token
        :param: roles include realm or client level roles as array
        :param is_client_level boolean to indicate to add roles to client
        :return:  void
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            rolesRepArray = []

            for role in roles:
                rolesRep = RoleRepresentation(name=role['name'], description=role['description'],
                                              composite=role['composite'])
                rolesRepArray.append(rolesRep)

            request = AddRolesRequest(roles=rolesRepArray, client_level=is_client_level)

            return self.tenant_stub.addTenantRoles(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in add_tenant_roles, probably due to invalid parameters")
            raise

    def add_protocol_mapper(self, token, name, attribute_name, claim_name, claim_type, mapper_type,
                            add_to_id_token, add_to_access_token, add_to_user_info, multi_valued,
                            aggregate_attribute_values):
        """
        Protocol mapper enables to add user attributes, user realm roles or user client roles to be
        added to ID token, Access token.
        :param token
        :param: roles include realm or client level roles as array
        :param is_client_level boolean to indicate to add roles to client
        :return:  void
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

            return self.tenant_stub.addProtocolMapper(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in add_protocol_mapper, probably due to invalid parameters")
            raise

    def get_child_tenants(self, token, offset, limit, status):
        """
        Get child tenants of the calling tenant
        :param token
        :param: offset omit initial number of results equalt to offset
        :param limit results should contain  maximum number of entries
        :param status (ACTIVE, REQUESTED, DENIED, CANCELLED, DEACTIVATED)
        :return:  Tenants
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            request = GetTenantsRequest(offset=offset, limit=limit, status=status)

            return self.tenant_stub.getChildTenants(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in get_child_tenants, probably due to invalid parameters")
            raise

    def get_all_tenants(self, token, email):
        """
        Get all tenants requested by given user
        :param token
        :param email get all tenants requested by email
        :return:  Tenants
        """
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            request = GetAllTenantsForUserRequest(email=email)

            return self.tenant_stub.getAllTenantsForUser(request, metadata=metadata)

        except Exception:
            logger.exception("Error occurred in get_all_tenants, probably due to invalid parameters")
            raise
