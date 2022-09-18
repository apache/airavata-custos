from __future__ import (absolute_import, division, print_function)
__metaclass__ = type

import os
import testinfra.utils.ansible_runner

testinfra_hosts = [os.environ['COMPOSE_PROJECT_NAME'] + '_ansible_1']


def test_retrieved_secret(host):
    secrets_file = host.file('/conjur_secrets.txt')

    assert secrets_file.exists

    result = host.check_output("cat /conjur_secrets.txt", shell=True)

    assert result == "var_with_spaces_secret_password"
