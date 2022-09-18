# (c) 2020, NetApp, Inc
# BSD-3 Clause (see COPYING or https://opensource.org/licenses/BSD-3-Clause)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

from ansible.module_utils import six
from units.modules.utils import AnsibleExitJson, AnsibleFailJson, ModuleTestCase, set_module_args
from ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_proxy_systems import NetAppESeriesProxySystems
from units.compat import mock


class StoragePoolTest(ModuleTestCase):
    REQUIRED_PARAMS = {"api_username": "username",
                       "api_password": "password",
                       "api_url": "http://localhost/devmgr/v2",
                       "validate_certs": "no"}

    REQUEST_FUNC = "ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_proxy_systems.NetAppESeriesProxySystems.request"
    _REQUEST_FUNC = "ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_proxy_systems.NetAppESeriesProxySystems._request"
    TIME_FUNC = "ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_proxy_systems.sleep"

    def _set_args(self, args=None):
        module_args = self.REQUIRED_PARAMS.copy()
        if args is not None:
            module_args.update(args)
        set_module_args(module_args)

    def test_valid_options_pass(self):
        """Verify valid options."""
        options_list = [{"password": "password", "systems": [{"ssid": "10", "serial": "021633035190"},
                                                             {"addresses": ["192.168.1.100"]},
                                                             {"serial": "021628016299"}]},
                        {"password": "password", "systems": ["021178889999", "022348016297", "021625436296"]},
                        {"password": "password", "systems": []}, {}]

        for options in options_list:
            self._set_args(options)
            systems = NetAppESeriesProxySystems()

        self._set_args(options_list[0])
        systems = NetAppESeriesProxySystems()
        self.assertEquals(systems.systems, [
            {"ssid": "10", "serial": "021633035190", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
             "meta_tags": [], "controller_addresses": [], "embedded_available": None, "accept_certificate": False, "current_info": {}, "changes": {},
             "updated_required": False, "failed": False, "discovered": False},
            {"ssid": "192.168.1.100", "serial": "", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
             "meta_tags": [], "controller_addresses": ["192.168.1.100"], "embedded_available": None, "accept_certificate": False, "current_info": {},
             "changes": {}, "updated_required": False, "failed": False, "discovered": False},
            {"ssid": "021628016299", "serial": "021628016299", "password": "password", "password_valid": None, "password_set": None,
             "stored_password_valid": None, "meta_tags": [], "controller_addresses": [], "embedded_available": None, "accept_certificate": False,
             "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": False}])

    def test_invalid_options_fail(self):
        """Verify invalid systems option throws expected exception."""
        self._set_args({"password": "password", "systems": [[]]})
        with self.assertRaisesRegexp(AnsibleFailJson, "Invalid system! All systems must either be a simple serial number or a dictionary."):
            systems = NetAppESeriesProxySystems()

    def test_discover_array_pass(self):
        """Verify the discover_array method."""
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24",
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        response = {"discoverProcessRunning": False, "storageSystems": [{"serialNumber": "1", "ipAddresses": ["192.168.1.5", "192.168.1.6"],
                                                                         "supportedManagementPorts": ["https", "symbol"]},
                                                                        {"serialNumber": "2", "ipAddresses": ["192.168.1.15", "192.168.1.16"],
                                                                         "supportedManagementPorts": ["symbol"]},
                                                                        {"serialNumber": "3", "ipAddresses": ["192.168.1.25", "192.168.1.26"],
                                                                         "supportedManagementPorts": ["https", "symbol"]},
                                                                        {"serialNumber": "4", "ipAddresses": ["192.168.1.35", "192.168.1.36"],
                                                                         "supportedManagementPorts": ["symbol"]}]}
        systems = NetAppESeriesProxySystems()
        with mock.patch(self.TIME_FUNC, return_value=None):
            with mock.patch(self.REQUEST_FUNC, side_effect=[(200, {"requestId": "1"}), (200, {"discoverProcessRunning": True}), (200, response)]):
                systems.discover_array()
                self.assertEquals(systems.systems, [
                    {"ssid": "1", "serial": "1", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
                     "meta_tags": [], "controller_addresses": ["192.168.1.5", "192.168.1.6"], "embedded_available": True, "accept_certificate": True,
                     "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True},
                    {"ssid": "192.168.1.36", "serial": "", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
                     "meta_tags": [], "controller_addresses": ["192.168.1.35", "192.168.1.36"], "embedded_available": False, "accept_certificate": False,
                     "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True},
                    {"ssid": "2", "serial": "2", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
                     "meta_tags": [], "controller_addresses": ["192.168.1.15", "192.168.1.16"], "embedded_available": False, "accept_certificate": False,
                     "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True}])

        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24", "add_discovered_systems": True})
        response = {"discoverProcessRunning": False, "storageSystems": [{"serialNumber": "1", "ipAddresses": ["192.168.1.5", "192.168.1.6"],
                                                                         "supportedManagementPorts": ["https", "symbol"]},
                                                                        {"serialNumber": "2", "ipAddresses": ["192.168.1.15", "192.168.1.16"],
                                                                         "supportedManagementPorts": ["symbol"]},
                                                                        {"serialNumber": "3", "ipAddresses": ["192.168.1.25", "192.168.1.26"],
                                                                         "supportedManagementPorts": ["https", "symbol"]},
                                                                        {"serialNumber": "4", "ipAddresses": ["192.168.1.35", "192.168.1.36"],
                                                                         "supportedManagementPorts": ["symbol"]}]}
        systems = NetAppESeriesProxySystems()
        with mock.patch(self.TIME_FUNC, return_value=None):
            with mock.patch(self.REQUEST_FUNC, side_effect=[(200, {"requestId": "1"}), (200, {"discoverProcessRunning": True}), (200, response)]):
                systems.discover_array()
                self.assertEquals(systems.systems, [
                    {"ssid": "1", "serial": "1", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
                     "meta_tags": [], "controller_addresses": ["192.168.1.5", "192.168.1.6"], "embedded_available": True, "accept_certificate": True,
                     "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True},
                    {"ssid": "2", "serial": "2", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
                     "meta_tags": [], "controller_addresses": ["192.168.1.15", "192.168.1.16"], "embedded_available": False, "accept_certificate": False,
                     "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True},
                    {"ssid": "3", "serial": "3", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
                     "meta_tags": [], "controller_addresses": ["192.168.1.25", "192.168.1.26"], "embedded_available": True, "accept_certificate": True,
                     "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True},
                    {"ssid": "4", "serial": "4", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
                     "meta_tags": [], "controller_addresses": ["192.168.1.35", "192.168.1.36"], "embedded_available": False, "accept_certificate": False,
                     "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True}])

    def test_discover_array_fail(self):
        """Verify discover_array method throws expected exceptions."""
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24", "add_discovered_systems": True})
        systems = NetAppESeriesProxySystems()
        with self.assertRaisesRegex(AnsibleFailJson, "Failed to initiate array discovery."):
            with mock.patch(self.TIME_FUNC, return_value=None):
                with mock.patch(self.REQUEST_FUNC, return_value=Exception()):
                    systems.discover_array()

        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24", "add_discovered_systems": True})
        systems = NetAppESeriesProxySystems()
        with self.assertRaisesRegex(AnsibleFailJson, "Failed to get the discovery results."):
            with mock.patch(self.TIME_FUNC, return_value=None):
                with mock.patch(self.REQUEST_FUNC, side_effect=[(200, {"requestId": "1"}), Exception()]):
                    systems.discover_array()

        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24", "add_discovered_systems": True})
        systems = NetAppESeriesProxySystems()
        with self.assertRaisesRegex(AnsibleFailJson, "Timeout waiting for array discovery process."):
            with mock.patch(self.TIME_FUNC, return_value=None):
                with mock.patch(self.REQUEST_FUNC, side_effect=[(200, {"requestId": "1"})] + [(200, {"discoverProcessRunning": True})] * 1000):
                    systems.discover_array()

    def test_update_storage_systems_info_pass(self):
        """Verify update_storage_systems_info method performs correctly."""
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24",
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        systems.systems = [
            {"ssid": "1", "serial": "1", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
             "meta_tags": [], "controller_addresses": ["192.168.1.5", "192.168.1.6"], "embedded_available": True, "accept_certificate": True,
             "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True},
            {"ssid": "192.168.1.36", "serial": "", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
             "meta_tags": [], "controller_addresses": ["192.168.1.35", "192.168.1.36"], "embedded_available": False, "accept_certificate": False,
             "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True},
            {"ssid": "2", "serial": "2", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
             "meta_tags": [], "controller_addresses": ["192.168.1.15", "192.168.1.16"], "embedded_available": False, "accept_certificate": False,
             "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True}]

        with mock.patch(self.REQUEST_FUNC, return_value=(200, [{"id": "1", "passwordStatus": "valid", "metaTags": []},
                                                               {"id": "5", "passwordStatus": "valid", "metaTags": []}])):
            systems.update_storage_systems_info()
            self.assertEquals(systems.systems_to_remove, ["5"])
            self.assertEquals(systems.systems_to_add, [
                {"ssid": "192.168.1.36", "serial": "", "password": "password", "password_valid": None, "password_set": None,
                 "stored_password_valid": None, "meta_tags": [], "controller_addresses": ["192.168.1.35", "192.168.1.36"], "embedded_available": False,
                 "accept_certificate": False, "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True},
                {"ssid": "2", "serial": "2", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
                 "meta_tags": [], "controller_addresses": ["192.168.1.15", "192.168.1.16"], "embedded_available": False, "accept_certificate": False,
                 "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True}])

    def test_update_storage_systems_info_fail(self):
        """Verify update_storage_systems_info throws expected exceptions."""
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24",
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        systems.systems = [
            {"ssid": "1", "serial": "1", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
             "meta_tags": [], "controller_addresses": ["192.168.1.5", "192.168.1.6"], "embedded_available": True, "accept_certificate": True,
             "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True},
            {"ssid": "192.168.1.36", "serial": "", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
             "meta_tags": [], "controller_addresses": ["192.168.1.35", "192.168.1.36"], "embedded_available": False, "accept_certificate": False,
             "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True},
            {"ssid": "2", "serial": "2", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
             "meta_tags": [], "controller_addresses": ["192.168.1.15", "192.168.1.16"], "embedded_available": False, "accept_certificate": False,
             "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True}]

        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to retrieve storage systems."):
            with mock.patch(self.REQUEST_FUNC, return_value=Exception()):
                systems.update_storage_systems_info()

    def test_set_password_pass(self):
        """Verify set_password completes as expected."""
        system = {"ssid": "1", "serial": "1", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
                  "meta_tags": [], "controller_addresses": ["192.168.1.5", "192.168.1.6"], "embedded_available": True, "accept_certificate": True,
                  "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True}
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24",
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        with mock.patch(self.TIME_FUNC, return_value=None):
            with mock.patch(self._REQUEST_FUNC, return_value=(200, None)):
                systems.set_password(system)
                self.assertFalse(system["password_set"])

        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24",
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        with mock.patch(self.TIME_FUNC, return_value=None):
            with mock.patch(self._REQUEST_FUNC, return_value=(401, None)):
                systems.set_password(system)
                self.assertTrue(system["password_set"])

    def test_set_password_fail(self):
        """Verify set_password throws expected exceptions."""
        system = {"ssid": "1", "serial": "1", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
                  "meta_tags": [], "controller_addresses": ["192.168.1.5", "192.168.1.6"], "embedded_available": True, "accept_certificate": True,
                  "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True}
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24",
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        with mock.patch(self.TIME_FUNC, return_value=None):
            with mock.patch(self._REQUEST_FUNC, return_value=Exception()):
                systems.set_password(system)
                self.assertTrue(system["failed"])

        system = {"ssid": "1", "serial": "1", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
                  "meta_tags": [], "controller_addresses": ["192.168.1.5", "192.168.1.6"], "embedded_available": True, "accept_certificate": True,
                  "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True}
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24",
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        with mock.patch(self.TIME_FUNC, return_value=None):
            with mock.patch(self._REQUEST_FUNC, side_effect=[(200, None), Exception(), Exception(), Exception()]):
                systems.set_password(system)
                self.assertTrue(system["failed"])

    def test_update_system_changes_pass(self):
        """Verify system changes."""
        system = {"ssid": "1", "serial": "1", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
                  "meta_tags": [], "controller_addresses": ["192.168.1.5", "192.168.1.6"], "embedded_available": True, "accept_certificate": True,
                  "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True}
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24",
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        systems.update_system_changes(system)
        self.assertEquals(system["changes"], {})

        system = {"ssid": "1", "serial": "1", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
                  "meta_tags": [], "controller_addresses": ["192.168.1.5", "192.168.1.6"], "embedded_available": True, "accept_certificate": True,
                  "current_info": {"managementPaths": ["192.168.1.25", "192.168.1.6"], "metaTags": [],
                                   "controllers": [{"certificateStatus": "trusted"}, {"certificateStatus": "trusted"}]},
                  "changes": {}, "updated_required": False, "failed": False, "discovered": True}
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24",
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        systems.update_system_changes(system)
        self.assertEquals(system["changes"], {"controllerAddresses": ["192.168.1.5", "192.168.1.6"]})

        system = {"ssid": "1", "serial": "1", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
                  "meta_tags": [], "controller_addresses": ["192.168.1.5", "192.168.1.6"], "embedded_available": True, "accept_certificate": True,
                  "current_info": {"managementPaths": ["192.168.1.5", "192.168.1.6"], "metaTags": [], "ip1": "192.168.1.5", "ip2": "192.168.1.6",
                                   "controllers": [{"certificateStatus": "trusted"}, {"certificateStatus": "unknown"}]},
                  "changes": {}, "updated_required": False, "failed": False, "discovered": True}
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24",
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        systems.update_system_changes(system)
        self.assertEquals(system["changes"], {"acceptCertificate": True})

        system = {"ssid": "1", "serial": "1", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
                  "meta_tags": [], "controller_addresses": ["192.168.1.5", "192.168.1.6"], "embedded_available": True, "accept_certificate": True,
                  "current_info": {"managementPaths": ["192.168.1.5", "192.168.1.6"], "metaTags": [{"key": "key", "value": "1"}], "ip1": "192.168.1.5",
                                   "ip2": "192.168.1.6",
                                   "controllers": [{"certificateStatus": "trusted"}, {"certificateStatus": "trusted"}]},
                  "changes": {}, "updated_required": False, "failed": False, "discovered": True}
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24",
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        systems.update_system_changes(system)
        self.assertEquals(system["changes"], {"removeAllTags": True})

        system = {"ssid": "1", "serial": "1", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
                  "meta_tags": [{"key": "key", "value": "1"}], "controller_addresses": ["192.168.1.5", "192.168.1.6"], "embedded_available": True,
                  "accept_certificate": True,
                  "current_info": {"managementPaths": ["192.168.1.5", "192.168.1.6"], "metaTags": [], "ip1": "192.168.1.5", "ip2": "192.168.1.6",
                                   "controllers": [{"certificateStatus": "trusted"}, {"certificateStatus": "trusted"}]},
                  "changes": {}, "updated_required": False, "failed": False, "discovered": True}
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24",
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        systems.update_system_changes(system)
        self.assertEquals(system["changes"], {"metaTags": [{"key": "key", "value": "1"}]})

    def test_add_system_pass(self):
        """Validate add_system method."""
        system = {"ssid": "1", "serial": "1", "password": "password", "meta_tags": [{"key": "key", "value": "1"}],
                  "controller_addresses": ["192.168.1.5", "192.168.1.6"], "accept_certificate": True}
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24",
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        systems.set_password = lambda x: None
        with mock.patch(self.TIME_FUNC, return_value=None):
            with mock.patch(self.REQUEST_FUNC, side_effect=[(200, None), (200, None)]):
                systems.add_system(system)

        system = {"ssid": "1", "serial": "1", "password": "password", "meta_tags": [],
                  "controller_addresses": ["192.168.1.5", "192.168.1.6"], "accept_certificate": False}
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24",
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        systems.set_password = lambda x: None
        with mock.patch(self.TIME_FUNC, return_value=None):
            with mock.patch(self.REQUEST_FUNC, side_effect=[(200, None), (200, None)]):
                systems.add_system(system)

        # Test warning situations, tests should still succeed
        system = {"ssid": "1", "serial": "1", "password": "password", "meta_tags": [{"key": "key", "value": "1"}],
                  "controller_addresses": ["192.168.1.5", "192.168.1.6"], "accept_certificate": True}
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24",
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        systems.set_password = lambda x: None
        with mock.patch(self.TIME_FUNC, return_value=None):
            with mock.patch(self.REQUEST_FUNC, side_effect=[Exception(), Exception()]):
                systems.add_system(system)

        system = {"ssid": "1", "serial": "1", "password": "password", "meta_tags": [{"key": "key", "value": "1"}],
                  "controller_addresses": ["192.168.1.5", "192.168.1.6"], "accept_certificate": True}
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24",
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        systems.set_password = lambda x: None
        with mock.patch(self.TIME_FUNC, return_value=None):
            with mock.patch(self.REQUEST_FUNC, side_effect=[(200, None), Exception()]):
                systems.add_system(system)

    def test_update_system_pass(self):
        """Validate update_system method."""
        system = {"ssid": "1", "changes": {}}
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24",
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        systems.set_password = lambda x: None
        with mock.patch(self.TIME_FUNC, return_value=None):
            with mock.patch(self.REQUEST_FUNC, return_value=(200, None)):
                systems.update_system(system)

        system = {"ssid": "1", "changes": {}}
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24",
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        systems.set_password = lambda x: None
        with mock.patch(self.TIME_FUNC, return_value=None):
            with mock.patch(self.REQUEST_FUNC, return_value=Exception()):
                systems.update_system(system)

    def test_remove_system_pass(self):
        """Validate remove_system method."""
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24",
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        systems.set_password = lambda x: None
        with mock.patch(self.TIME_FUNC, return_value=None):
            with mock.patch(self.REQUEST_FUNC, return_value=(200, None)):
                systems.remove_system("1")

        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24",
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        systems.set_password = lambda x: None
        with mock.patch(self.TIME_FUNC, return_value=None):
            with mock.patch(self.REQUEST_FUNC, return_value=Exception()):
                systems.remove_system("1")

    def test_apply_pass(self):
        """Validate apply method."""
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24", "add_discovered_systems": False,
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        systems.is_embedded = lambda: False
        systems.discover_array = lambda: None
        systems.update_storage_systems_info = lambda: None
        systems.update_system_changes = lambda x: None
        systems.remove_system = lambda x: None
        systems.add_system = lambda x: None
        systems.update_system = lambda x: None
        systems.systems = [{"ssid": "1", "serial": "1", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
                            "meta_tags": [], "controller_addresses": ["192.168.1.5", "192.168.1.6"], "embedded_available": True, "accept_certificate": True,
                            "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True},
                           {"ssid": "192.168.1.36", "serial": "", "password": "password", "password_valid": None, "password_set": None,
                            "stored_password_valid": None,
                            "meta_tags": [], "controller_addresses": ["192.168.1.35", "192.168.1.36"], "embedded_available": False, "accept_certificate": False,
                            "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True},
                           {"ssid": "2", "serial": "2", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
                            "meta_tags": [], "controller_addresses": ["192.168.1.15", "192.168.1.16"], "embedded_available": False, "accept_certificate": False,
                            "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True}]
        systems.systems_to_remove = ["5"]
        systems.systems_to_add = [{"ssid": "192.168.1.36", "serial": "", "password": "password", "password_valid": None, "password_set": None,
                                   "stored_password_valid": None, "meta_tags": [], "controller_addresses": ["192.168.1.35", "192.168.1.36"],
                                   "embedded_available": False,
                                   "accept_certificate": False, "current_info": {}, "changes": {}, "updated_required": False, "failed": False,
                                   "discovered": True},
                                  {"ssid": "2", "serial": "2", "password": "password", "password_valid": None, "password_set": None,
                                   "stored_password_valid": None,
                                   "meta_tags": [], "controller_addresses": ["192.168.1.15", "192.168.1.16"], "embedded_available": False,
                                   "accept_certificate": False,
                                   "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True}]
        systems.systems_to_update = [{"ssid": "192.168.1.36", "serial": "", "password": "password", "password_valid": None, "password_set": None,
                                      "stored_password_valid": None, "meta_tags": [], "controller_addresses": ["192.168.1.35", "192.168.1.36"],
                                      "embedded_available": False,
                                      "accept_certificate": False, "current_info": {}, "changes": {}, "updated_required": False, "failed": False,
                                      "discovered": True},
                                     {"ssid": "2", "serial": "2", "password": "password", "password_valid": None, "password_set": None,
                                      "stored_password_valid": None,
                                      "meta_tags": [], "controller_addresses": ["192.168.1.15", "192.168.1.16"], "embedded_available": False,
                                      "accept_certificate": False,
                                      "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True}]
        with self.assertRaisesRegexp(AnsibleExitJson, "systems added.*?systems updated.*?system removed"):
            systems.apply()

        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24", "add_discovered_systems": False,
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        systems.is_embedded = lambda: False
        systems.discover_array = lambda: None
        systems.update_storage_systems_info = lambda: None
        systems.update_system_changes = lambda x: None
        systems.remove_system = lambda x: None
        systems.add_system = lambda x: None
        systems.update_system = lambda x: None
        systems.systems = [{"ssid": "1", "serial": "1", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
                            "meta_tags": [], "controller_addresses": ["192.168.1.5", "192.168.1.6"], "embedded_available": True, "accept_certificate": True,
                            "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True},
                           {"ssid": "192.168.1.36", "serial": "", "password": "password", "password_valid": None, "password_set": None,
                            "stored_password_valid": None,
                            "meta_tags": [], "controller_addresses": ["192.168.1.35", "192.168.1.36"], "embedded_available": False, "accept_certificate": False,
                            "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True},
                           {"ssid": "2", "serial": "2", "password": "password", "password_valid": None, "password_set": None, "stored_password_valid": None,
                            "meta_tags": [], "controller_addresses": ["192.168.1.15", "192.168.1.16"], "embedded_available": False, "accept_certificate": False,
                            "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True}]
        systems.systems_to_remove = ["5"]
        systems.systems_to_add = [{"ssid": "192.168.1.36", "serial": "", "password": "password", "password_valid": None, "password_set": None,
                                   "stored_password_valid": None, "meta_tags": [], "controller_addresses": ["192.168.1.35", "192.168.1.36"],
                                   "embedded_available": False,
                                   "accept_certificate": False, "current_info": {}, "changes": {}, "updated_required": False, "failed": False,
                                   "discovered": True},
                                  {"ssid": "2", "serial": "2", "password": "password", "password_valid": None, "password_set": None,
                                   "stored_password_valid": None,
                                   "meta_tags": [], "controller_addresses": ["192.168.1.15", "192.168.1.16"], "embedded_available": False,
                                   "accept_certificate": False,
                                   "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True}]
        systems.systems_to_update = [{"ssid": "192.168.1.36", "serial": "", "password": "password", "password_valid": None, "password_set": None,
                                      "stored_password_valid": None, "meta_tags": [], "controller_addresses": ["192.168.1.35", "192.168.1.36"],
                                      "embedded_available": False,
                                      "accept_certificate": False, "current_info": {}, "changes": {}, "updated_required": False, "failed": False,
                                      "discovered": True},
                                     {"ssid": "2", "serial": "2", "password": "password", "password_valid": None, "password_set": None,
                                      "stored_password_valid": None,
                                      "meta_tags": [], "controller_addresses": ["192.168.1.15", "192.168.1.16"], "embedded_available": False,
                                      "accept_certificate": False,
                                      "current_info": {}, "changes": {}, "updated_required": False, "failed": False, "discovered": True}]
        systems.undiscovered_systems = ["5", "6"]
        with self.assertRaises(AnsibleFailJson):
            systems.apply()

        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24", "add_discovered_systems": False,
                        "systems": []})
        systems = NetAppESeriesProxySystems()
        systems.is_embedded = lambda: False
        systems.discover_array = lambda: None
        systems.update_storage_systems_info = lambda: None
        systems.update_system_changes = lambda x: None
        systems.remove_system = lambda x: None
        systems.add_system = lambda x: None
        systems.systems = []
        systems.systems_to_remove = []
        systems.systems_to_add = []
        systems.systems_to_update = []
        with self.assertRaisesRegexp(AnsibleExitJson, "No changes were made."):
            systems.apply()

        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24", "add_discovered_systems": False,
                        "systems": []})
        systems = NetAppESeriesProxySystems()
        systems.is_embedded = lambda: False
        systems.discover_array = lambda: None
        systems.update_storage_systems_info = lambda: None
        systems.update_system_changes = lambda x: None
        systems.remove_system = lambda x: None
        systems.add_system = lambda x: None
        systems.systems = []
        systems.systems_to_remove = []
        systems.systems_to_add = []
        systems.undiscovered_systems = ["5", "6"]
        with self.assertRaises(AnsibleFailJson):
            systems.apply()

    def test_apply_fail(self):
        """Validate apply method throws expected exceptions."""
        self._set_args({"password": "password", "subnet_mask": "192.168.1.0/24", "add_discovered_systems": False,
                        "systems": [{"ssid": "1", "serial": "1"}, {"addresses": ["192.168.1.36"]}, {"serial": "2"}, {"serial": "5"}]})
        systems = NetAppESeriesProxySystems()
        systems.is_embedded = lambda: True
        with self.assertRaisesRegexp(AnsibleFailJson, "Cannot add/remove storage systems to SANtricity Web Services Embedded instance."):
            systems.apply()
