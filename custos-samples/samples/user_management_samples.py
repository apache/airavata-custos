import os

from custos.clients.user_management_client import UserManagementClient

from custos.transport.settings import CustosServerClientSettings
import custos.clients.utils.utilities as utl

# load root directoty
BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

#get settings file path (settings file path reside in configs folder under home directory)
settings_path = os.path.join(BASE_DIR, 'configs', "settings.ini")

# read settings
custos_settings = CustosServerClientSettings(configuration_file_location=settings_path)

# create custos user management client
user_management_client = UserManagementClient(custos_settings)

# obtain base 64 encoded token for tenant
b64_encoded_custos_token = utl.get_token(custos_settings=custos_settings)


def register_user():
    response = user_management_client.register_user(token=b64_encoded_custos_token,
                                                    username="TestUser5",
                                                    first_name="Watson",
                                                    last_name="Christe",
                                                    password="1234",
                                                    email="wat@gmail.com",
                                                    is_temp_password=False)
    print(response)


def enable_user():
    response = user_management_client.enable_user(token=b64_encoded_custos_token,
                                                  username="TestUser5")
    print(response)


def get_user():
    response = user_management_client.get_user(token=b64_encoded_custos_token,
                                               username="TestUser5")
    print(response)


def update_user():
    response = user_management_client.update_user_profile(token=b64_encoded_custos_token,
                                                          username="TestUser5",
                                                          first_name="Jimmy",
                                                          last_name="Jhon",
                                                          email="wat@gmail.com")
    print(response)


def find_users():
    response = user_management_client.find_users(token=b64_encoded_custos_token, offset=0, limit=1,
                                                 username="TestUser5")
    print(response)


register_user()
enable_user()
get_user()
update_user()
find_users()
