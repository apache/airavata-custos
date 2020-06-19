import pytest
import logging

from clients.identity_management_client import IdentityManagementClient
from transport.settings import CustosServerClientSettings
import clients.utils.utilities as utl

logger = logging.getLogger(__name__)

logger.setLevel(logging.DEBUG)
# create console handler with a higher log level
handler = logging.StreamHandler()
handler.setLevel(logging.DEBUG)

# load IdentityManagementClient with default configuration
client = IdentityManagementClient()

custos_settings = CustosServerClientSettings()
main_token = utl.get_token(custos_settings)


@pytest.mark.skip(reason="no way of currently testing this")
def test_authenticate():
    with pytest.raises(Exception) as e:
        assert client.authenticate(main_token, "isjarana", "Custos1234")
    print(e.__class__.__name__)


@pytest.mark.skip(reason="no way of currently testing this")
def test_authorize():
    response = client.authorize("custos-xgect9otrwawa8uwztym-10000006", "http://custos.lk", "code",
                                "openid email profile", "asdadasdewde")
    print(response)