#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

"""
Tests for the well_knowns module.
"""

from airavata_custos import well_knowns


def test_get_well_known_oidc_uri_for_realm():
    well_known_oidc_uri = well_knowns.get_well_known_oidc_uri_for_realm(
        "test-realm")
    assert (well_known_oidc_uri ==
            "https://iam.scigap.org/auth/realms/"
            "test-realm/.well-known/openid-configuration")


def test_get_dev_well_known_oidc_uri_for_realm():
    well_known_oidc_uri = well_knowns.get_dev_well_known_oidc_uri_for_realm(
        "test-realm")
    assert (well_known_oidc_uri ==
            "https://iamdev.scigap.org/auth/realms/"
            "test-realm/.well-known/openid-configuration")
