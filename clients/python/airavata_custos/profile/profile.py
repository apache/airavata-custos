from airavata_custos import utils
from custos.commons.model.security.ttypes import AuthzToken
from custos.profile.model.User.ttypes import UserProfile
from custos.profile.model.User.ttypes import Status


class User(object):

    def __init__(self, user_name, tenant, status=Status.ACTIVE, emails=None, first_name=None, last_name=None, middle_name=None, name_prefix=None,
                 name_suffix=None,
                 phones=None, country=None, nationality=None, labeled_URI=None, time_zone=None):

        UserProfile(user_name, tenant, status, emails, first_name, last_name, middle_name, name_prefix,
                           name_suffix ,phones, country, nationality, labeled_URI,time_zone )


class Profile(object):

    def __init__(self):
        host = ""
        port = ""
        self.profile_service_pool = utils.initialize_userprofile_client_pool(host, port)

    def initialize_user(self, user_name, tenant, status, emails, first_name, last_name, middle_name, name_prefix,
                    name_suffix, phones, country, nationality, time_zone, nsf_demographics, home_organization,
                    original_affiliation, labeled_URI, comments, user_model_version, ordcidId)
        """
        constructor to create a user object
        :param user_name: unique identifier of the user PRIMARY
        :param tenant: unique identifier of the tenant PRIMARY
        :param emails: user emails
        :param first_name:  First name as asserted by the user
        :param last_name: Last name as asserted by the user
        :param middle_name: middle name as asserted by the user
        :param name_prefix: prefix to the users name as asserted by the user
        :param name_suffix: suffix to the users name as asserted by the user
        :param phones: Telephone numbers
        :param country: Country of Residence
        :param nationality: Countries of citizenship
        :param labeled_URI: Google Scholar, Web of Science, ACS, e.t.c
        :param time_zone: User’s preferred timezone
        """
        return UserProfile(user_name, tenant, status, emails, first_name, last_name, middle_name, name_prefix,
                           name_suffix ,phones, country, nationality, labeled_URI,time_zone )

    def create_user(self, authorization_token: AuthzToken) -> User:
        """
        This method creates a new user in custos profile service
        :param authorization_token: object of class AuthzToken
        :param user_profile: object of class User
        :return: boolean true if user is created successfully otherwise false
        """
        return self.profile_service_pool.initializeUserProfile(authorization_token)

    def update_user(self, authorization_token, user_profile):
        """
        This method updates the user in custos profile service
        :param authorization_token: object of class AuthzToken
        :param user_profile: updated user info
        :param tenant: tenant identifier to which the user belongs
        :return: boolean true if user is updated successfully otherwise false
        """
        return self.profile_service_pool.updateUserProfile(authorization_token, user_profile)

    def delete_user(self, authorization_token, user_name, tenant):
        """
        This method deletes the user in the custos profile service
        :param authorization_token: object of class AuthzToken
        :param user_name: unique identifier of the user
        :param tenant: tenant identifier to which the user belongs
        :return: boolean true if user is deleted successfully otherwise false
        """
        return self.profile_service_pool.deleteUserProfile(authorization_token, user_name, tenant)

    def get_user(self, authorization_token, user_name, tenant):
        """
        To retrieve user info
        :param authorization_token:  object of class AuthzToken
        :param user_name: unique identifier of the user
        :param tenant: tenant identifier to which the user belongs
        :return: object of class User
        """
        return self.profile_service_pool.getUserProfileById(authorization_token, user_name, tenant)

    def get_all_users(self, authorization_token, tenant, offset=0, limit=-1):
        """
        To retrieve all users in the tenant
        :param authorization_token: object of class AuthzToken
        :param tenant: tenant identifier
        :param offset: to limit the number of users
        :param limit: to limit the number of users
        :return: list of users
        """
        return self.profile_service_pool.getAllUserProfilesInGateway(authorization_token, tenant, offset, limit)

