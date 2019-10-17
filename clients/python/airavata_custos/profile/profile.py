from airavata_custos import utils
from airavata_custos.security.keycloak_connectors import AuthzToken


class User(object):

    def __init__(self, user_name, tenant, emails, first_name=None, last_name=None, middle_name=None, name_prefix=None,
                 name_suffix=None,
                 phones=None, country=None, nationality=None, labeled_URI=None, time_zone=None):
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
        self.user_name = user_name
        self.tenant = tenant
        self.first_name = first_name
        self.last_name = last_name
        self.middle_name = middle_name
        self.namePrefix = name_prefix
        self.nameSuffix = name_suffix
        self.emails = emails
        self.phones = phones
        self.country = country
        self.nationality = nationality
        self.labeled_URI = labeled_URI
        self.time_zone = time_zone


class Profile(object):

    def __init__(self):
        host = ""
        port = ""
        self.profile_service_pool = utils.initialize_userprofile_client_pool(host, port)

    def create_user(self, authorization_token: AuthzToken, user_profile: User = None) -> User:
        """
        This method creates a new user in custos profile service
        :param authorization_token: object of class AuthzToken
        :param user_profile: object of class User
        :return: boolean true if user is created successfully otherwise false
        """
        if user_profile:
            return self.profile_service_pool.addUserProfile(authorization_token, user_profile, user_profile.tenant)
        else:
            return self.profile_service_pool.initializeUserProfile(authorization_token)

    def update_user(self, authorization_token, user_profile):
        """
        This method updates the user in custos profile service
        :param authorization_token: object of class AuthzToken
        :param user_profile: updated user info
        :param tenant: tenant identifier to which the user belongs
        :return: boolean true if user is updated successfully otherwise false
        """
        return self.updateUserProfile(authorization_token, user_profile)

    def delete_user(self, authorization_token, user_name, tenant):
        """
        This method deletes the user in the custos profile service
        :param authorization_token: object of class AuthzToken
        :param user_name: unique identifier of the user
        :param tenant: tenant identifier to which the user belongs
        :return: boolean true if user is deleted successfully otherwise false
        """
        return self.deleteUserProfile(authorization_token, user_name, tenant)

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

