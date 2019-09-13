import thrift_connector.connection_pool as connection_pool
from airavata_custos import settings
from custos.service.profile.iam.admin.services.cpi import IamAdminServices, constants

iamadmin_client_pool = connection_pool.ClientPool(
    IamAdminServices,
    settings.PROFILE_SERVICE_HOST,
    settings.PROFILE_SERVICE_PORT,
    connection_class=IAMAdminServiceThriftClient,
    keepalive=settings.THRIFT_CLIENT_POOL_KEEPALIVE
)

class IAMAdminServiceThriftClient(MultiplexThriftClientMixin,
                                  CustomThriftClient):
    service_name = constants.IAM_ADMIN_SERVICES_CPI_NAME
    secure = settings.PROFILE_SERVICE_SECURE