# content of conftest.py
import random
import string

import pytest
from clients.identity_management_client import IdentityManagementClient


def get_random_alpha_numeric_string():

    ascii_characters = string.ascii_letters + string.digits
    return ''.join((random.choice(ascii_characters) for _ in range(8)))


@pytest.fixture(scope="module")
def username():
    return get_random_alpha_numeric_string()


