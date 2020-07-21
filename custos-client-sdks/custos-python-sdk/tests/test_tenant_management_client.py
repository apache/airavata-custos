import pytest

import logging
from clients.tenant_management_client import TenantManagementClient
from clients.super_tenant_management_client import SuperTenantManagementClient
from clients.identity_management_client import IdentityManagementClient
from custos.integration import TenantManagementService_pb2
from transport.settings import CustosServerClientSettings
import clients.utils.utilities as utl

logger = logging.getLogger(__name__)

logger.setLevel(logging.DEBUG)
# create console handler with a higher log level
handler = logging.StreamHandler()
handler.setLevel(logging.DEBUG)

# load APIServerClient with default configuration
client = TenantManagementClient()
admin_client = SuperTenantManagementClient()
id_client = IdentityManagementClient()
custos_settings = CustosServerClientSettings()
token = utl.get_token(custos_settings)


class TestTenantManagementClient:
    client_name = "SAMPLE"
    requester_email = "XXX@iu.edu"
    contacts = ["2345634324"]
    redirect_uris = ["http://localhost:8080,http://localhost:8080/user/external_ids"]
    admin_first_name = "Admin First Name"
    admin_last_name = "LastName"
    admin_email = "email"
    client_uri = "https://domain.org/"
    scope =  "openid profile email org.cilogon.userinfo"
    domain = "domain.org"
    logo_uri = "https://domain.org/static/favicon.png"
    comment = "Galaxy Portal"

    @pytest.fixture
    def tenant(self):
        return client.create_admin_tenant(self.client_name,
                                          self.requester_email,
                                          self.admin_first_name,
                                          self.admin_last_name,
                                          self.admin_email,
                                          "admin",
                                          "1234",
                                          self.contacts,
                                          self.redirect_uris,
                                          self.client_uri,
                                          self.scope,
                                          self.domain,
                                          self.logo_uri,
                                          self.comment)

    def test_create_tenant(self, tenant):
        assert type(tenant) is TenantManagementService_pb2.CreateTenantResponse

    def test_get_tenant(self, tenant):
        client_id = tenant.client_id
        response = client.get_tenant(client_token=token, client_id=client_id)
        print(response)
        assert type(response) is TenantManagementService_pb2.GetTenantResponse
        assert response.client_id == tenant.client_id
        assert response.client_name == self.client_name
        assert response.requester_email == self.requester_email
        assert response.admin_first_name == self.admin_first_name
        assert response.admin_last_name == self.admin_last_name
        assert response.admin_email == self.admin_email
        assert all([a == b for a, b in zip(response.contacts, self.contacts)])
        assert all([a == b for a, b in zip(response.redirect_uris, self.redirect_uris)])
        assert response.scope == self.scope
        assert response.domain == self.domain
        assert response.logo_uri == self.logo_uri
        assert response.comment == self.comment

    @pytest.mark.skip(reason="no way of currently testing this")
    def test_delete_tenant(self, tenant):
        print("Client_id: ", tenant.client_id)
        response = client.delete_tenant(token, "custos-al6tf9r1aggetx6nrt1p-10000519")
        print(response)
