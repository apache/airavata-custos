# (c) 2020, NetApp, Inc
# BSD-3 Clause (see COPYING or https://opensource.org/licenses/BSD-3-Clause)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

from ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_ldap import NetAppESeriesLdap
from units.modules.utils import ModuleTestCase, set_module_args, AnsibleFailJson, AnsibleExitJson
from units.compat import mock


class LdapTest(ModuleTestCase):
    REQUIRED_PARAMS = {
        "api_username": "admin",
        "api_password": "password",
        "api_url": "http://localhost",
        "ssid": "1"}
    REQ_FUNC = "ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_ldap.NetAppESeriesLdap.request"
    BASE_REQ_FUNC = "ansible_collections.netapp_eseries.santricity.plugins.module_utils.santricity.request"

    GET_DOMAINS = {"version": "3",
                   "ldapDomains": [{"id": "test1",
                                    "bindLookupUser": {"password": "***", "user": "CN=cn,OU=accounts,DC=test1,DC=example,DC=com"},
                                    "groupAttributes": ["memberOf"],
                                    "ldapUrl": "ldap://test.example.com:389",
                                    "names": ["test.example.com"],
                                    "roleMapCollection": [{"groupRegex": ".*", "ignoreCase": False, "name": "storage.monitor"}],
                                    "searchBase": "OU=accounts,DC=test,DC=example,DC=com",
                                    "userAttribute": "sAMAccountName"},
                                   {"id": "test2",
                                    "bindLookupUser": {"password": "***", "user": "CN=cn,OU=accounts,DC=test2,DC=example,DC=com"},
                                    "groupAttributes": ["memberOf"],
                                    "ldapUrl": "ldap://test2.example.com:389",
                                    "names": ["test2.example.com"],
                                    "roleMapCollection": [{"groupRegex": ".*", "ignoreCase": False, "name": "storage.admin"},
                                                          {"groupRegex": ".*", "ignoreCase": False, "name": "support.admin"},
                                                          {"groupRegex": ".*", "ignoreCase": False, "name": "security.admin"},
                                                          {"groupRegex": ".*", "ignoreCase": False, "name": "storage.monitor"}],
                                    "searchBase": "OU=accounts,DC=test2,DC=example,DC=com",
                                    "userAttribute": "sAMAccountName"}]}

    def _set_args(self, args=None):
        module_args = self.REQUIRED_PARAMS.copy()
        if args is not None:
            module_args.update(args)
        set_module_args(module_args)

    def test_valid_options_pass(self):
        """Verify valid options."""
        options_list = [{"state": "disabled"},
                        {"state": "absent", "identifier": "test_domain"},
                        {"state": "present", "identifier": "test_domain", "server_url": "ldap://test.example.com:389",
                         "search_base": "ou=accounts,DC=test,DC=example,DC=com"},
                        {"state": "present", "identifier": "test_domain", "server_url": "ldap://test.example.com:389",
                         "search_base": "ou=accounts,DC=test,DC=example,DC=com", "bind_user": "admin", "bind_password": "adminpass"},
                        {"state": "present", "identifier": "test_domain", "server_url": "ldap://test.example.com:389",
                         "search_base": "ou=accounts,DC=test,DC=example,DC=com", "bind_user": "admin", "bind_password": "adminpass",
                         "names": ["name1", "name2"], "group_attributes": ["group_attr1", "group_attr1"], "user_attribute": "user_attr"}]

        for options in options_list:
            with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
                self._set_args(options)
                ldap = NetAppESeriesLdap()
        for options in options_list:
            with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": False})]):
                self._set_args(options)
                ldap = NetAppESeriesLdap()

    def test_get_domain_pass(self):
        """Verify get_domain returns expected data structure."""
        options = {"state": "present", "identifier": "test_domain", "server_url": "ldap://test.example.com:389",
                   "search_base": "ou=accounts,DC=test,DC=example,DC=com", "bind_user": "admin", "bind_password": "adminpass",
                   "names": ["name1", "name2"], "group_attributes": ["group_attr1", "group_attr1"], "user_attribute": "user_attr"}
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
            with mock.patch(self.REQ_FUNC, return_value=(200, self.GET_DOMAINS)):
                self._set_args(options)
                ldap = NetAppESeriesLdap()
                self.assertEquals(ldap.get_domains(), self.GET_DOMAINS["ldapDomains"])

    def test_get_domain_fail(self):
        """Verify get_domain throws expected exceptions."""
        options = {"state": "present", "identifier": "test_domain", "server_url": "ldap://test.example.com:389",
                   "search_base": "ou=accounts,DC=test,DC=example,DC=com", "bind_user": "admin", "bind_password": "adminpass",
                   "names": ["name1", "name2"], "group_attributes": ["group_attr1", "group_attr1"], "user_attribute": "user_attr"}
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                with self.assertRaisesRegexp(AnsibleFailJson, "Failed to retrieve current LDAP configuration."):
                    self._set_args(options)
                    ldap = NetAppESeriesLdap()
                    ldap.get_domains()

    def test_build_request_body_pass(self):
        """Verify build_request_body builds expected data structure."""
        options_list = [{"state": "present", "identifier": "test_domain", "server_url": "ldap://test.example.com:389",
                         "search_base": "ou=accounts,DC=test,DC=example,DC=com"},
                        {"state": "present", "identifier": "test_domain", "server_url": "ldap://test.example.com:389",
                         "search_base": "ou=accounts,DC=test,DC=example,DC=com", "bind_user": "admin", "bind_password": "adminpass"},
                        {"state": "present", "identifier": "test_domain", "server_url": "ldap://test.example.com:389",
                         "search_base": "ou=accounts,DC=test,DC=example,DC=com", "bind_user": "admin", "bind_password": "adminpass",
                         "names": ["name1", "name2"], "group_attributes": ["group_attr1", "group_attr1"], "user_attribute": "user_attr"}]
        expectation_list = [{'id': 'test_domain', 'groupAttributes': ['memberOf'], 'ldapUrl': 'ldap://test.example.com:389', 'names': ['test.example.com'],
                             'roleMapCollection': [], 'searchBase': 'ou=accounts,DC=test,DC=example,DC=com', 'userAttribute': 'sAMAccountName'},
                            {'id': 'test_domain', 'groupAttributes': ['memberOf'], 'ldapUrl': 'ldap://test.example.com:389', 'names': ['test.example.com'],
                             'roleMapCollection': [], 'searchBase': 'ou=accounts,DC=test,DC=example,DC=com', 'userAttribute': 'sAMAccountName',
                             'bindLookupUser': {'password': 'adminpass', 'user': 'admin'}},
                            {'id': 'test_domain', 'groupAttributes': ['group_attr1', 'group_attr1'], 'ldapUrl': 'ldap://test.example.com:389',
                             'names': ['name1', 'name2'], 'roleMapCollection': [], 'searchBase': 'ou=accounts,DC=test,DC=example,DC=com',
                             'userAttribute': 'user_attr', 'bindLookupUser': {'password': 'adminpass', 'user': 'admin'}}]
        for index in range(len(options_list)):
            with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
                self._set_args(options_list[index])
                ldap = NetAppESeriesLdap()
                ldap.build_request_body()
                self.assertEquals(ldap.body, expectation_list[index])

    def test_are_changes_required_pass(self):
        """Verify build_request_body builds expected data structure."""
        options_list = [{"state": "present", "identifier": "test_domain", "server_url": "ldap://test.example.com:389",
                         "search_base": "ou=accounts,DC=test,DC=example,DC=com"},
                        {"state": "present", "identifier": "test_domain", "server_url": "ldap://test.example.com:389",
                         "search_base": "ou=accounts,DC=test,DC=example,DC=com", "bind_user": "admin", "bind_password": "adminpass"},
                        {"state": "present", "identifier": "test_domain", "server_url": "ldap://test.example.com:389",
                         "search_base": "ou=accounts,DC=test,DC=example,DC=com", "bind_user": "admin", "bind_password": "adminpass",
                         "names": ["name1", "name2"], "group_attributes": ["group_attr1", "group_attr1"], "user_attribute": "user_attr"}]

        for index in range(len(options_list)):
            with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
                self._set_args(options_list[index])
                ldap = NetAppESeriesLdap()
                ldap.get_domains = lambda: self.GET_DOMAINS["ldapDomains"]
                self.assertTrue(ldap.are_changes_required())

        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
            self._set_args({"state": "disabled"})
            ldap = NetAppESeriesLdap()
            ldap.get_domains = lambda: self.GET_DOMAINS["ldapDomains"]
            self.assertTrue(ldap.are_changes_required())
            self.assertEquals(ldap.existing_domain_ids, ["test1", "test2"])

        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
            self._set_args({"state": "absent", "identifier": "test_domain"})
            ldap = NetAppESeriesLdap()
            ldap.get_domains = lambda: self.GET_DOMAINS["ldapDomains"]
            self.assertFalse(ldap.are_changes_required())

        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
            self._set_args({"state": "present", "identifier": "test2", "server_url": "ldap://test2.example.com:389",
                            "search_base": "ou=accounts,DC=test2,DC=example,DC=com",
                            "bind_user": "CN=cn,OU=accounts,DC=test2,DC=example,DC=com", "bind_password": "adminpass",
                            "role_mappings": {".*": ["storage.admin", "support.admin", "security.admin", "storage.monitor"]},
                            "names": ["test2.example.com"], "group_attributes": ["memberOf"], "user_attribute": "sAMAccountName"})
            ldap = NetAppESeriesLdap()
            ldap.build_request_body()
            ldap.get_domains = lambda: self.GET_DOMAINS["ldapDomains"]
            ldap.add_domain = lambda temporary, skip_test: {"id": "ANSIBLE_TMP_DOMAIN"}

            with mock.patch(self.REQ_FUNC, return_value=(200, [{"id": "test2", "result": {"authenticationTestResult": "ok"}},
                                                               {"id": "ANSIBLE_TMP_DOMAIN", "result": {"authenticationTestResult": "ok"}}])):
                self.assertFalse(ldap.are_changes_required())

        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
            self._set_args({"state": "present", "identifier": "test2", "server_url": "ldap://test2.example.com:389",
                            "search_base": "ou=accounts,DC=test,DC=example,DC=com",
                            "bind_user": "CN=cn,OU=accounts,DC=test2,DC=example,DC=com", "bind_password": "adminpass",
                            "role_mappings": {".*": ["storage.admin", "support.admin", "security.admin", "storage.monitor"]},
                            "names": ["test2.example.com"], "group_attributes": ["memberOf"], "user_attribute": "sAMAccountName"})
            ldap = NetAppESeriesLdap()
            ldap.build_request_body()
            ldap.get_domains = lambda: self.GET_DOMAINS["ldapDomains"]
            ldap.add_domain = lambda temporary, skip_test: {"id": "ANSIBLE_TMP_DOMAIN"}

            with mock.patch(self.REQ_FUNC, return_value=(200, [{"id": "test2", "result": {"authenticationTestResult": "fail"}},
                                                               {"id": "ANSIBLE_TMP_DOMAIN", "result": {"authenticationTestResult": "ok"}}])):
                self.assertTrue(ldap.are_changes_required())

    def test_are_changes_required_fail(self):
        """Verify are_changes_required throws expected exception."""
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
            self._set_args({"state": "present", "identifier": "test2", "server_url": "ldap://test2.example.com:389",
                            "search_base": "ou=accounts,DC=test2,DC=example,DC=com",
                            "bind_user": "CN=cn,OU=accounts,DC=test2,DC=example,DC=com", "bind_password": "adminpass",
                            "role_mappings": {".*": ["storage.admin", "support.admin", "security.admin", "storage.monitor"]},
                            "names": ["test2.example.com"], "group_attributes": ["memberOf"], "user_attribute": "sAMAccountName"})
            ldap = NetAppESeriesLdap()
            ldap.build_request_body()
            ldap.get_domains = lambda: self.GET_DOMAINS["ldapDomains"]
            ldap.add_domain = lambda temporary, skip_test: {"id": "ANSIBLE_TMP_DOMAIN"}
            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to authenticate bind credentials!"):
                with mock.patch(self.REQ_FUNC, return_value=(200, [{"id": "test2", "result": {"authenticationTestResult": "fail"}},
                                                                   {"id": "ANSIBLE_TMP_DOMAIN", "result": {"authenticationTestResult": "fail"}}])):
                    ldap.are_changes_required()

        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
            self._set_args({"state": "present", "identifier": "test2", "server_url": "ldap://test2.example.com:389",
                            "search_base": "ou=accounts,DC=test2,DC=example,DC=com",
                            "bind_user": "CN=cn,OU=accounts,DC=test2,DC=example,DC=com", "bind_password": "adminpass",
                            "role_mappings": {".*": ["storage.admin", "support.admin", "security.admin", "storage.monitor"]},
                            "names": ["test2.example.com"], "group_attributes": ["memberOf"], "user_attribute": "sAMAccountName"})
            ldap = NetAppESeriesLdap()
            ldap.build_request_body()
            ldap.get_domains = lambda: self.GET_DOMAINS["ldapDomains"]
            ldap.add_domain = lambda temporary, skip_test: {"id": "ANSIBLE_TMP_DOMAIN"}
            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to authenticate bind credentials!"):
                with mock.patch(self.REQ_FUNC, return_value=(200, [{"id": "test2", "result": {"authenticationTestResult": "ok"}},
                                                                   {"id": "ANSIBLE_TMP_DOMAIN", "result": {"authenticationTestResult": "fail"}}])):
                    ldap.are_changes_required()

    def test_add_domain_pass(self):
        """Verify add_domain returns expected data."""
        self._set_args({"state": "present", "identifier": "test2", "server_url": "ldap://test2.example.com:389",
                        "search_base": "ou=accounts,DC=test,DC=example,DC=com",
                        "bind_user": "CN=cn,OU=accounts,DC=test2,DC=example,DC=com", "bind_password": "adminpass",
                        "role_mappings": {".*": ["storage.admin", "support.admin", "security.admin", "storage.monitor"]},
                        "names": ["test2.example.com"], "group_attributes": ["memberOf"], "user_attribute": "sAMAccountName"})

        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
            ldap = NetAppESeriesLdap()
            ldap.build_request_body()
            with mock.patch(self.REQ_FUNC, return_value=(200, {"ldapDomains": [{"id": "test2"}]})):
                self.assertEquals(ldap.add_domain(), {"id": "test2"})

    def test_add_domain_fail(self):
        """Verify add_domain returns expected data."""
        self._set_args({"state": "present", "identifier": "test2", "server_url": "ldap://test2.example.com:389",
                        "search_base": "ou=accounts,DC=test,DC=example,DC=com",
                        "bind_user": "CN=cn,OU=accounts,DC=test2,DC=example,DC=com", "bind_password": "adminpass",
                        "role_mappings": {".*": ["storage.admin", "support.admin", "security.admin", "storage.monitor"]},
                        "names": ["test2.example.com"], "group_attributes": ["memberOf"], "user_attribute": "sAMAccountName"})

        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
            ldap = NetAppESeriesLdap()
            ldap.build_request_body()
            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to create LDAP domain."):
                with mock.patch(self.REQ_FUNC, return_value=Exception()):
                    ldap.add_domain()

    def test_update_domain_pass(self):
        """Verify update_domain returns expected data."""
        self._set_args({"state": "present", "identifier": "test2", "server_url": "ldap://test2.example.com:389",
                        "search_base": "ou=accounts,DC=test,DC=example,DC=com",
                        "bind_user": "CN=cn,OU=accounts,DC=test2,DC=example,DC=com", "bind_password": "adminpass",
                        "role_mappings": {".*": ["storage.admin", "support.admin", "security.admin", "storage.monitor"]},
                        "names": ["test2.example.com"], "group_attributes": ["memberOf"], "user_attribute": "sAMAccountName"})

        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
            ldap = NetAppESeriesLdap()
            ldap.build_request_body()
            ldap.domain = {"id": "test2"}
            with mock.patch(self.REQ_FUNC, return_value=(200, None)):
                ldap.update_domain()

    def test_update_domain_fail(self):
        """Verify update_domain returns expected data."""
        self._set_args({"state": "present", "identifier": "test2", "server_url": "ldap://test2.example.com:389",
                        "search_base": "ou=accounts,DC=test,DC=example,DC=com",
                        "bind_user": "CN=cn,OU=accounts,DC=test2,DC=example,DC=com", "bind_password": "adminpass",
                        "role_mappings": {".*": ["storage.admin", "support.admin", "security.admin", "storage.monitor"]},
                        "names": ["test2.example.com"], "group_attributes": ["memberOf"], "user_attribute": "sAMAccountName"})

        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
            ldap = NetAppESeriesLdap()
            ldap.build_request_body()
            ldap.domain = {"id": "test2"}
            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to update LDAP domain."):
                with mock.patch(self.REQ_FUNC, return_value=Exception()):
                    ldap.update_domain()

    def test_delete_domain_pass(self):
        """Verify delete_domain returns expected data."""
        self._set_args({"state": "present", "identifier": "test2", "server_url": "ldap://test2.example.com:389",
                        "search_base": "ou=accounts,DC=test,DC=example,DC=com",
                        "bind_user": "CN=cn,OU=accounts,DC=test2,DC=example,DC=com", "bind_password": "adminpass",
                        "role_mappings": {".*": ["storage.admin", "support.admin", "security.admin", "storage.monitor"]},
                        "names": ["test2.example.com"], "group_attributes": ["memberOf"], "user_attribute": "sAMAccountName"})

        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
            ldap = NetAppESeriesLdap()
            with mock.patch(self.REQ_FUNC, return_value=(200, None)):
                ldap.delete_domain("test2")

    def test_delete_domain_fail(self):
        """Verify delete_domain returns expected data."""
        self._set_args({"state": "present", "identifier": "test2", "server_url": "ldap://test2.example.com:389",
                        "search_base": "ou=accounts,DC=test,DC=example,DC=com",
                        "bind_user": "CN=cn,OU=accounts,DC=test2,DC=example,DC=com", "bind_password": "adminpass",
                        "role_mappings": {".*": ["storage.admin", "support.admin", "security.admin", "storage.monitor"]},
                        "names": ["test2.example.com"], "group_attributes": ["memberOf"], "user_attribute": "sAMAccountName"})

        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
            ldap = NetAppESeriesLdap()
            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to delete LDAP domain."):
                with mock.patch(self.REQ_FUNC, return_value=Exception()):
                    ldap.delete_domain("test2")

    def test_disable_domains_pass(self):
        """Verify disable_domains completes successfully."""
        self._set_args({"state": "present", "identifier": "test2", "server_url": "ldap://test2.example.com:389",
                        "search_base": "ou=accounts,DC=test,DC=example,DC=com",
                        "bind_user": "CN=cn,OU=accounts,DC=test2,DC=example,DC=com", "bind_password": "adminpass",
                        "role_mappings": {".*": ["storage.admin", "support.admin", "security.admin", "storage.monitor"]},
                        "names": ["test2.example.com"], "group_attributes": ["memberOf"], "user_attribute": "sAMAccountName"})

        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
            ldap = NetAppESeriesLdap()
            ldap.delete_domain = lambda x: None
            ldap.existing_domain_ids = ["id1", "id2", "id3"]
            ldap.disable_domains()

    def test_apply_pass(self):
        """Verify apply exits as expected."""
        self._set_args({"state": "present", "identifier": "test2", "server_url": "ldap://test2.example.com:389",
                        "search_base": "ou=accounts,DC=test,DC=example,DC=com",
                        "bind_user": "CN=cn,OU=accounts,DC=test2,DC=example,DC=com", "bind_password": "adminpass",
                        "role_mappings": {".*": ["storage.admin", "support.admin", "security.admin", "storage.monitor"]},
                        "names": ["test2.example.com"], "group_attributes": ["memberOf"], "user_attribute": "sAMAccountName"})

        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
            ldap = NetAppESeriesLdap()
            ldap.build_request_body = lambda: None
            ldap.are_changes_required = lambda: False
            with self.assertRaisesRegexp(AnsibleExitJson, "No changes have been made to the LDAP configuration."):
                ldap.apply()

        self._set_args({"state": "present", "identifier": "test2", "server_url": "ldap://test2.example.com:389",
                        "search_base": "ou=accounts,DC=test,DC=example,DC=com",
                        "bind_user": "CN=cn,OU=accounts,DC=test2,DC=example,DC=com", "bind_password": "adminpass",
                        "role_mappings": {".*": ["storage.admin", "support.admin", "security.admin", "storage.monitor"]},
                        "names": ["test2.example.com"], "group_attributes": ["memberOf"], "user_attribute": "sAMAccountName"})

        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
            ldap = NetAppESeriesLdap()
            ldap.build_request_body = lambda: None
            ldap.are_changes_required = lambda: True
            ldap.add_domain = lambda: None
            ldap.domain = {}
            with self.assertRaisesRegexp(AnsibleExitJson, "LDAP domain has been added."):
                ldap.apply()

        self._set_args({"state": "present", "identifier": "test2", "server_url": "ldap://test2.example.com:389",
                        "search_base": "ou=accounts,DC=test,DC=example,DC=com",
                        "bind_user": "CN=cn,OU=accounts,DC=test2,DC=example,DC=com", "bind_password": "adminpass",
                        "role_mappings": {".*": ["storage.admin", "support.admin", "security.admin", "storage.monitor"]},
                        "names": ["test2.example.com"], "group_attributes": ["memberOf"], "user_attribute": "sAMAccountName"})

        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
            ldap = NetAppESeriesLdap()
            ldap.build_request_body = lambda: None
            ldap.are_changes_required = lambda: True
            ldap.update_domain = lambda: None
            ldap.domain = {"id": "test"}
            with self.assertRaisesRegexp(AnsibleExitJson, "LDAP domain has been updated."):
                ldap.apply()

        self._set_args({"state": "absent", "identifier": "test2"})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
            ldap = NetAppESeriesLdap()
            ldap.build_request_body = lambda: None
            ldap.are_changes_required = lambda: True
            ldap.delete_domain = lambda x: None
            ldap.domain = {"id": "test"}
            with self.assertRaisesRegexp(AnsibleExitJson, "LDAP domain has been removed."):
                ldap.apply()

        self._set_args({"state": "disabled"})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.10.0000.0001"}), (200, {"runningAsProxy": True})]):
            ldap = NetAppESeriesLdap()
            ldap.build_request_body = lambda: None
            ldap.are_changes_required = lambda: True
            ldap.disable_domain = lambda: None
            ldap.domain = {"id": "test"}
            with self.assertRaisesRegexp(AnsibleExitJson, "All LDAP domains have been removed."):
                ldap.apply()
