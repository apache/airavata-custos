# (c) 2020, NetApp, Inc
# BSD-3 Clause (see COPYING or https://opensource.org/licenses/BSD-3-Clause)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

from ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_discover import NetAppESeriesDiscover
from units.modules.utils import AnsibleExitJson, AnsibleFailJson, ModuleTestCase, set_module_args
from units.compat import mock


class AlertsTest(ModuleTestCase):
    REQUIRED_PARAMS = {"subnet_mask": "192.168.1.0/24"}
    BASE_REQ_FUNC = 'ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_discover.request'
    SLEEP_FUNC = 'ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_discover.sleep'

    def _set_args(self, args=None):
        module_args = self.REQUIRED_PARAMS.copy()
        if args is not None:
            module_args.update(args)
        set_module_args(module_args)

    def test_valid_options_pass(self):
        """Verify constructor accepts valid options."""
        options_list = [{"ports": [1, 8443]},
                        {"ports": [8080, 65535]},
                        {"ports": [8443], "proxy_url": "https://192.168.1.1:8443/devmgr/v2/", "proxy_username": "admin", "proxy_password": "adminpass"},
                        {"ports": [8443], "proxy_url": "https://192.168.1.1:8443/devmgr/v2/", "proxy_username": "admin", "proxy_password": "adminpass",
                         "prefer_embedded": True},
                        {"ports": [8443], "proxy_url": "https://192.168.1.1:8443/devmgr/v2/", "proxy_username": "admin", "proxy_password": "adminpass",
                         "prefer_embedded": False},
                        {"ports": [8443], "proxy_url": "https://192.168.1.1:8443/devmgr/v2/", "proxy_username": "admin", "proxy_password": "adminpass",
                         "proxy_validate_certs": True},
                        {"ports": [8443], "proxy_url": "https://192.168.1.1:8443/devmgr/v2/", "proxy_username": "admin", "proxy_password": "adminpass",
                         "proxy_validate_certs": False}]

        for options in options_list:
            self._set_args(options)
            discover = NetAppESeriesDiscover()

    def test_valid_options_fail(self):
        """Verify constructor throws expected exceptions."""
        options_list = [{"ports": [0, 8443]}, {"ports": [8080, 65536]}, {"ports": [8080, "port"]}, {"ports": [8080, -10]}, {"ports": [8080, 70000]}]

        for options in options_list:
            self._set_args(options)
            with self.assertRaisesRegexp(AnsibleFailJson, "Invalid port! Ports must be positive numbers between 0 and 65536."):
                discover = NetAppESeriesDiscover()

    def test_check_ip_address_pass(self):
        """Verify check_ip_address successfully completes."""
        self._set_args()
        with mock.patch(self.BASE_REQ_FUNC, return_value=(200, {"chassisSerialNumber": "012345678901", "storageArrayLabel": "array_label"})):
            discover = NetAppESeriesDiscover()
            discover.check_ip_address(discover.systems_found, "192.168.1.100")
        self.assertEqual(discover.systems_found, {"012345678901": {"api_urls": ["https://192.168.1.100:8443/devmgr/v2/storage-systems/1/"],
                                                                   "label": "array_label", "addresses": [], "proxy_required": False}})

        self._set_args({"ports": [8080, 8443]})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(404, None), (401, None), (200, {"sa": {"saData": {"chassisSerialNumber": "012345678901",
                                                                                                            "storageArrayLabel": "array_label"}}})]):
            discover = NetAppESeriesDiscover()
            discover.check_ip_address(discover.systems_found, "192.168.1.101")
        self.assertEqual(discover.systems_found, {"012345678901": {"api_urls": ["https://192.168.1.101:8443/devmgr/v2/storage-systems/1/"],
                                                                   "label": "array_label", "addresses": [], "proxy_required": False}})

    def test_no_proxy_discover_pass(self):
        """Verify no_proxy_discover completes successfully."""
        self._set_args()
        discover = NetAppESeriesDiscover()
        discover.check_ip_address = lambda: None
        discover.no_proxy_discover()

    def test_verify_proxy_service_pass(self):
        """Verify verify_proxy_service completes successfully."""
        self._set_args({"proxy_url": "https://192.168.1.200", "proxy_username": "admin", "proxy_password": "adminpass"})
        discover = NetAppESeriesDiscover()
        with mock.patch(self.BASE_REQ_FUNC, return_value=(200, {"runningAsProxy": True})):
            discover.verify_proxy_service()

    def test_verify_proxy_service_fail(self):
        """Verify verify_proxy_service throws expected exception."""
        self._set_args({"proxy_url": "https://192.168.1.200", "proxy_username": "admin", "proxy_password": "adminpass"})
        discover = NetAppESeriesDiscover()
        with self.assertRaisesRegexp(AnsibleFailJson, "Web Services is not running as a proxy!"):
            with mock.patch(self.BASE_REQ_FUNC, return_value=(200, {"runningAsProxy": False})):
                discover.verify_proxy_service()

        self._set_args({"proxy_url": "https://192.168.1.200", "proxy_username": "admin", "proxy_password": "adminpass"})
        discover = NetAppESeriesDiscover()
        with self.assertRaisesRegexp(AnsibleFailJson, "Proxy is not available! Check proxy_url."):
            with mock.patch(self.BASE_REQ_FUNC, return_value=Exception()):
                discover.verify_proxy_service()

    def test_test_systems_found_pass(self):
        """Verify test_systems_found adds to systems_found dictionary."""
        self._set_args({"proxy_url": "https://192.168.1.200", "proxy_username": "admin", "proxy_password": "adminpass", "prefer_embedded": True})
        discover = NetAppESeriesDiscover()
        with mock.patch(self.BASE_REQ_FUNC, return_value=(200, {"runningAsProxy": True})):
            discover.test_systems_found(discover.systems_found, "012345678901", "array_label", ["192.168.1.100", "192.168.1.102"])
        self.assertEqual(discover.systems_found, {"012345678901": {"api_urls": ["https://192.168.1.100:8443/devmgr/v2/",
                                                                                "https://192.168.1.102:8443/devmgr/v2/"],
                                                                   "label": "array_label",
                                                                   "addresses": ["192.168.1.100", "192.168.1.102"],
                                                                   "proxy_required": False}})

    def test_proxy_discover_pass(self):
        """Verify proxy_discover completes successfully."""
        self._set_args({"subnet_mask": "192.168.1.0/30", "proxy_url": "https://192.168.1.200", "proxy_username": "admin", "proxy_password": "adminpass"})
        discover = NetAppESeriesDiscover()
        discover.verify_proxy_service = lambda: None
        with mock.patch(self.SLEEP_FUNC, return_value=None):
            with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"requestId": "1"}), (200, {"discoverProcessRunning": True}),
                                                             (200, {"discoverProcessRunning": False,
                                                                    "storageSystems": [{"controllers": [{"ipAddresses": ["192.168.1.100", "192.168.1.102"]}],
                                                                                        "supportedManagementPorts": ["https"], "serialNumber": "012345678901",
                                                                                        "label": "array_label"}]})]):
                discover.proxy_discover()

        self._set_args({"subnet_mask": "192.168.1.0/30", "proxy_url": "https://192.168.1.200", "proxy_username": "admin", "proxy_password": "adminpass"})
        discover = NetAppESeriesDiscover()
        discover.verify_proxy_service = lambda: None
        with mock.patch(self.SLEEP_FUNC, return_value=None):
            with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"requestId": "1"}), (200, {"discoverProcessRunning": True}),
                                                             (200, {"discoverProcessRunning": False,
                                                                    "storageSystems": [{"controllers": [{"ipAddresses": ["192.168.1.100", "192.168.1.102"]}],
                                                                                        "supportedManagementPorts": [], "serialNumber": "012345678901",
                                                                                        "label": "array_label"}]})]):
                discover.proxy_discover()

    def test_proxy_discover_fail(self):
        """Verify proxy_discover throws expected exceptions."""
        self._set_args({"subnet_mask": "192.168.1.0/30", "proxy_url": "https://192.168.1.200", "proxy_username": "admin", "proxy_password": "adminpass"})
        discover = NetAppESeriesDiscover()
        discover.verify_proxy_service = lambda: None
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to initiate array discovery."):
            with mock.patch(self.SLEEP_FUNC, return_value=None):
                with mock.patch(self.BASE_REQ_FUNC, return_value=Exception()):
                    discover.proxy_discover()

        self._set_args({"subnet_mask": "192.168.1.0/30", "proxy_url": "https://192.168.1.200", "proxy_username": "admin", "proxy_password": "adminpass"})
        discover = NetAppESeriesDiscover()
        discover.verify_proxy_service = lambda: None
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to get the discovery results."):
            with mock.patch(self.SLEEP_FUNC, return_value=None):
                with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"requestId": "1"}), Exception()]):
                    discover.proxy_discover()

        self._set_args({"subnet_mask": "192.168.1.0/30", "proxy_url": "https://192.168.1.200", "proxy_username": "admin", "proxy_password": "adminpass"})
        discover = NetAppESeriesDiscover()
        discover.verify_proxy_service = lambda: None
        with self.assertRaisesRegexp(AnsibleFailJson, "Timeout waiting for array discovery process."):
            with mock.patch(self.SLEEP_FUNC, return_value=None):
                with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"requestId": "1"})] + [(200, {"discoverProcessRunning": True})] * 300):
                    discover.proxy_discover()

    def test_discover_pass(self):
        """Verify discover successfully completes."""
        self._set_args({"subnet_mask": "192.168.1.0/30", "proxy_url": "https://192.168.1.200", "proxy_username": "admin", "proxy_password": "adminpass"})
        discover = NetAppESeriesDiscover()
        discover.proxy_discover = lambda: None
        with self.assertRaisesRegexp(AnsibleExitJson, "Discover process complete."):
            discover.discover()

        self._set_args()
        discover = NetAppESeriesDiscover()
        discover.no_proxy_discover = lambda: None
        with self.assertRaisesRegexp(AnsibleExitJson, "Discover process complete."):
            discover.discover()
