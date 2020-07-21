import pytest

from clients.user_management_client import UserManagementClient
from clients.identity_management_client import IdentityManagementClient
from clients.super_tenant_management_client import SuperTenantManagementClient

from transport.settings import CustosServerClientSettings
import clients.utils.utilities as utl

# load APIServerClient with default configuration
client = UserManagementClient()
id_client = IdentityManagementClient()

custos_settings = CustosServerClientSettings()
token = utl.get_token(custos_settings)

admin_client = SuperTenantManagementClient()


@pytest.fixture(scope="module")
def admin_access_token():
    admin_token = id_client.token(token, None, None, "admin", "1234", None, "password")
    return admin_token["access_token"]


def test_register_user(username):
    response = client.register_user(token, username, "Test123", "Smith", "12345", "jhon@iu.edu", True)
    assert response.is_registered == True


def test_get_user(username):
    response = client.get_user(token, username)
    assert response.first_name == "Test123"
    assert response.last_name == "Smith"
    assert response.email == "jhon@iu.edu"
    assert response.state == "PENDING_CONFIRMATION"


def test_find_user(username):
    response = client.find_users(token, 0, 2, username, None, None, None)


def test_enable_user(username):
    client.enable_user(token, username)
    response = client.is_user_enabled(token, username)
    assert response.status == True


def test_is_username_available(username):
    response = client.is_username_available(token, username)
    assert response.status == False, "User exists, should be True"


def test_add_user_attribute(username, admin_access_token):
    attributes = [
        {
            "key": "phone",
            "values": ["8123915386"]
        }
    ]
    users = [username]
    response = client.add_user_attributes(admin_access_token, attributes, users)
    assert response.status == True
    response_user = client.get_user(token, username)
    assert response_user.attributes[0].key == "phone"
    assert response_user.attributes[0].values[0] == "8123915386"


def test_delete_user_attribute(username, admin_access_token):
    attributes = [
        {
            "key": "phone",
            "values": ["8123915386"]
        }
    ]
    users = [username]
    response = client.delete_user_attributes(admin_access_token, attributes, users)
    assert response.status == True


@pytest. mark. skip(reason="no way of currently testing this")
def test_add_user_roles(username, admin_access_token):
    roles = ["testing"]
    users = [username]
    response = client.add_roles_to_users(admin_access_token, users, roles, False)
    assert response.status == True


@pytest. mark. skip(reason="no way of currently testing this")
def test_delete_user_roles(username, admin_access_token):
    roles = ["testing"]
    users = [username]
    response = client.delete_user_roles(admin_access_token, users, roles, None)
    assert response.status == True


def test_delete_user(username):
    admin_token = id_client.token(token, None, None, "admin", "1234", None, "password")
    admin_access_token = admin_token["access_token"]
    response = client.delete_user(admin_access_token, username)
    assert response.status == True