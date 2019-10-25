from airavata_custos import utils
from custos.commons.model.security.ttypes import AuthzToken
from custos.profile.model.User.ttypes import UserProfile
from airavata_custos.settings import ProfileSettings
import configparser


class Profile(object):

    def __init__(self, configuration_file_location: str):
        self.profile_settings = ProfileSettings()
        self._load_settings(configuration_file_location)
        self.profile_service_pool = utils.initialize_userprofile_client_pool(self.profile_settings.PROFILE_SERVICE_HOST,
                                                                             self.profile_settings.PROFILE_SERVICE_PORT)

    def create_user(self, authorization_token: AuthzToken) -> UserProfile:
        """
        This method creates a new user in custos profile service
        :param authorization_token: object of class AuthzToken
        :return: boolean true if user is created successfully otherwise false
        """
        return self.profile_service_pool.initializeUserProfile(authorization_token)

    def update_user(self, authorization_token: AuthzToken, user_profile: UserProfile) -> bool:
        """
        This method updates the user in custos profile service
        :param authorization_token: object of class AuthzToken
        :param user_profile: updated user info
        :param tenant: tenant identifier to which the user belongs
        :return: boolean true if user is updated successfully otherwise false
        """
        return self.profile_service_pool.updateUserProfile(authorization_token, user_profile)

    def delete_user(self, authorization_token: AuthzToken, user_name: str, tenant: str) -> bool:
        """
        This method deletes the user in the custos profile service
        :param authorization_token: object of class AuthzToken
        :param user_name: unique identifier of the user
        :param tenant: tenant identifier to which the user belongs
        :return: boolean true if user is deleted successfully otherwise false
        """
        return self.profile_service_pool.deleteUserProfile(authorization_token, user_name, tenant)

    def get_user(self, authorization_token: AuthzToken, user_name: str, tenant: str) -> UserProfile:
        """
        To retrieve user info
        :param authorization_token:  object of class AuthzToken
        :param user_name: unique identifier of the user
        :param tenant: tenant identifier to which the user belongs
        :return: object of class User
        """
        return self.profile_service_pool.getUserProfileById(authorization_token, user_name, tenant)

    def get_all_users(self, authorization_token: AuthzToken, tenant: str, offset: int = 0, limit: int = -1) -> list:
        """
        To retrieve all users in the tenant
        :param authorization_token: object of class AuthzToken
        :param tenant: tenant identifier
        :param offset: to limit the number of users
        :param limit: to limit the number of users
        :return: list of users
        """
        return self.profile_service_pool.getAllUserProfilesInGateway(authorization_token, tenant, offset, limit)

    def _load_settings(self, configuration_file_location):
        config = configparser.ConfigParser()
        config.read(configuration_file_location)
        settings = config['ProfileServerSettings']
        self.profile_settings.PROFILE_SERVICE_HOST = settings['PROFILE_SERVICE_HOST']
        self.profile_settings.PROFILE_SERVICE_PORT = settings['PROFILE_SERVICE_PORT']