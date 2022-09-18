from __future__ import (absolute_import, division, print_function)

__metaclass__ = type

import os
import testinfra.utils.ansible_runner

testinfra_hosts = [os.environ['COMPOSE_PROJECT_NAME'] + '_ansible_1']


def test_retrieved_secret(host):
    """
    Verify that the as_file parameter makes the lookup plugin return the path to a temporary file
    containing the secret.
    """
    lookup_output_file = host.file('/lookup_output.txt')
    assert lookup_output_file.exists

    secret_file = host.file(lookup_output_file.content_string)
    assert secret_file.exists
    assert secret_file.mode == 0o600
    assert secret_file.content_string == "test_secret_in_file_password"
