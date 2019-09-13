import logging
from airavata_custos.utils import iamadmin_client_pool

logger = logging.getLogger(__name__)


def is_username_available(authz_token, username):
    """
    This method validates if the username is available or not
    :param authz_token: Object of AuthzToken class containing access token, username, gatewayId of the active user
    :param username: The username whose availability needs to be verified
    :return: boolean
    """
    return iamadmin_client_pool.isUsernameAvailable(authz_token, username)


def register_user(authz_token, username, email_address, first_name, last_name, password):
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
    return iamadmin_client_pool.registerUser(
        authz_token,
        username,
        email_address,
        first_name,
        last_name,
        password)


def is_user_enabled(authz_token, username):
    """
    Checks the user is enabled/disabled in keycloak. Only the enabled user can login
    :param authz_token: Object of AuthzToken class containing access token, username, gatewayId of the active user
    :param username: The username of the user
    :return: boolean
    """
    return iamadmin_client_pool.isUserEnabled(authz_token, username)


def enable_user(authz_token, username):
    """
    The method to enable a disabled user
    :param authz_token: Object of AuthzToken class containing access token, username, gatewayId of the active user
    :param username: The username of the user
    :return: Object of UserProfile class, containing user details
    """
    return iamadmin_client_pool.enableUser(authz_token, username)


def delete_user(authz_token, username):
    """
    This method deleted the user from keycloak. Returns true if delete is successful
    :param authz_token: Object of AuthzToken class containing access token, username, gatewayId of the active user
    :param username: The username of the user
    :return: boolean
    """
    return iamadmin_client_pool.deleteUser(authz_token, username)


def is_user_exist(authz_token, username):
    """
    This method checks if the user exists in keycloak. Returns true if the user exists otherwise returns false
    :param authz_token: Object of AuthzToken class containing access token, username, gatewayId of the active user
    :param username: The username of the user
    :return: boolean
    """
    return iamadmin_client_pool.isUserExist(authz_token, username)


def get_user(authz_token, username):
    """

    :param authz_token:
    :param username:
    :return:
    """
    return iamadmin_client_pool.getUser(authz_token, username)


def get_users(authz_token, offset, limit, search=None):
    """

    :param authz_token:
    :param offset:
    :param limit:
    :param search:
    :return:
    """
    return iamadmin_client_pool.getUsers(authz_token, offset, limit, search)


def reset_user_password(authz_token, username, new_password):
    """

    :param authz_token:
    :param username:
    :param new_password:
    :return:
    """
    return iamadmin_client_pool.resetUserPassword(
        authz_token, username, new_password)

def set_up_tenant(authz_token, gateway, tenantAdminPasswordCredentials):
    pass
