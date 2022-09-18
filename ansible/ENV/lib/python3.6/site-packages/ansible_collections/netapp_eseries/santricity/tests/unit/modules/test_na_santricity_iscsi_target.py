# (c) 2020, NetApp, Inc
# BSD-3 Clause (see COPYING or https://opensource.org/licenses/BSD-3-Clause)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

from ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_iscsi_target import NetAppESeriesIscsiTarget
from units.modules.utils import AnsibleFailJson, AnsibleExitJson, ModuleTestCase, set_module_args
from units.compat import mock


class IscsiTargetTest(ModuleTestCase):
    REQUIRED_PARAMS = {"api_username": "admin", "api_password": "adminpassword", "api_url": "http://localhost", "ssid": "1", "name": "abc"}
    CHAP_SAMPLE = "a" * 14
    REQ_FUNC = "ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_iscsi_target.NetAppESeriesIscsiTarget.request"
    TARGET_REQUEST_RESPONSE = [{"targetRef": "90000000600A098000A4B28D00334A065DA9D747",
                                "nodeName": {"ioInterfaceType": "iscsi",
                                             "iscsiNodeName": "iqn.1992-08.com.netapp:2806.600a098000a4b28d000000005da9d744",
                                             "remoteNodeWWN": None, "nvmeNodeName": None},
                                "alias": {"ioInterfaceType": "iscsi",
                                          "iscsiAlias": "target_name"},
                                "configuredAuthMethods": {"authMethodData": [{"authMethod": "none",
                                                                              "chapSecret": None}]},
                                "portals": [{"groupTag": 2,
                                             "ipAddress": {"addressType": "ipv4",
                                                           "ipv4Address": "10.10.10.110",
                                                           "ipv6Address": None},
                                             "tcpListenPort": 3260},
                                            {"groupTag": 2,
                                             "ipAddress": {"addressType": "ipv6",
                                                           "ipv4Address": None,
                                                           "ipv6Address": "FE8000000000000002A098FFFEA4B9D7"},
                                             "tcpListenPort": 3260},
                                            {"groupTag": 2,
                                             "ipAddress": {"addressType": "ipv4",
                                                           "ipv4Address": "10.10.10.112",
                                                           "ipv6Address": None},
                                             "tcpListenPort": 3260},
                                            {"groupTag": 1, "ipAddress": {"addressType": "ipv4",
                                                                          "ipv4Address": "10.10.11.110",
                                                                          "ipv6Address": None},
                                             "tcpListenPort": 3260},
                                            {"groupTag": 1,
                                             "ipAddress": {"addressType": "ipv6",
                                                           "ipv4Address": None,
                                                           "ipv6Address": "FE8000000000000002A098FFFEA4B293"},
                                             "tcpListenPort": 3260},
                                            {"groupTag": 1,
                                             "ipAddress": {"addressType": "ipv4",
                                                           "ipv4Address": "10.10.11.112",
                                                           "ipv6Address": None},
                                             "tcpListenPort": 3260}]}]
    ISCSI_ENTRY_DATA_RESPONSE = [{"icmpPingResponseEnabled": False,
                                  "unnamedDiscoverySessionsEnabled": False,
                                  "isnsServerTcpListenPort": 0,
                                  "ipv4IsnsServerAddressConfigMethod": "configDhcp",
                                  "ipv4IsnsServerAddress": "0.0.0.0",
                                  "ipv6IsnsServerAddressConfigMethod": "configStatic",
                                  "ipv6IsnsServerAddress": "00000000000000000000000000000000",
                                  "isnsRegistrationState": "__UNDEFINED",
                                  "isnsServerRegistrationEnabled": False,
                                  "hostPortsConfiguredDHCP": False}]

    def _set_args(self, args=None):
        module_args = self.REQUIRED_PARAMS.copy()
        if args is not None:
            module_args.update(args)
        set_module_args(module_args)

    def test_validate_params(self):
        """Ensure we can pass valid parameters to the module"""
        for i in range(12, 57):
            secret = 'a' * i
            self._set_args(dict(chap=secret))
            tgt = NetAppESeriesIscsiTarget()

    def test_invalid_chap_secret(self):
        for secret in [11 * 'a', 58 * 'a']:
            with self.assertRaisesRegexp(AnsibleFailJson, r'.*?CHAP secret is not valid.*') as result:
                self._set_args(dict(chap=secret))
                tgt = NetAppESeriesIscsiTarget()

    def test_target_pass(self):
        """Ensure target property returns the expected data structure."""
        expected_response = {"alias": "target_name", "chap": False, "iqn": "iqn.1992-08.com.netapp:2806.600a098000a4b28d000000005da9d744",
                             "ping": False, "unnamed_discovery": False}

        self._set_args({"name": "target_name", "ping": True, "unnamed_discovery": True})
        iscsi_target = NetAppESeriesIscsiTarget()

        with mock.patch(self.REQ_FUNC, side_effect=[(200, self.TARGET_REQUEST_RESPONSE), (200, self.ISCSI_ENTRY_DATA_RESPONSE)]):
            self.assertEquals(iscsi_target.target, expected_response)

    def test_target_fail(self):
        """Ensure target property returns the expected data structure."""
        self._set_args({"name": "target_name", "ping": True, "unnamed_discovery": True})
        iscsi_target = NetAppESeriesIscsiTarget()

        with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to retrieve the iSCSI target information."):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                result = iscsi_target.target

        with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to retrieve the iSCSI target information."):
            with mock.patch(self.REQ_FUNC, side_effect=[(200, self.TARGET_REQUEST_RESPONSE), Exception()]):
                result = iscsi_target.target

        with self.assertRaisesRegexp(AnsibleFailJson, r"This storage-system does not appear to have iSCSI interfaces."):
            with mock.patch(self.REQ_FUNC, return_value=(200, [])):
                result = iscsi_target.target

    def test_apply_iscsi_settings_pass(self):
        """Ensure apply_iscsi_settings succeeds properly."""
        self._set_args({"name": "not_target_name"})
        iscsi_target = NetAppESeriesIscsiTarget()
        with mock.patch(self.REQ_FUNC, side_effect=[(200, self.TARGET_REQUEST_RESPONSE), (200, self.ISCSI_ENTRY_DATA_RESPONSE), (200, [])]):
            self.assertTrue(iscsi_target.apply_iscsi_settings())

        self._set_args({"name": "target_name"})
        iscsi_target = NetAppESeriesIscsiTarget()
        with mock.patch(self.REQ_FUNC, side_effect=[(200, self.TARGET_REQUEST_RESPONSE), (200, self.ISCSI_ENTRY_DATA_RESPONSE), (200, [])]):
            self.assertFalse(iscsi_target.apply_iscsi_settings())

    def test_apply_iscsi_settings_fail(self):
        """Ensure apply_iscsi_settings fails properly."""
        self._set_args({"name": "not_target_name"})
        iscsi_target = NetAppESeriesIscsiTarget()
        with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to update the iSCSI target settings."):
            with mock.patch(self.REQ_FUNC, side_effect=[(200, self.TARGET_REQUEST_RESPONSE), (200, self.ISCSI_ENTRY_DATA_RESPONSE), Exception()]):
                self.assertTrue(iscsi_target.apply_iscsi_settings())

    def test_apply_target_changes_pass(self):
        """Ensure apply_iscsi_settings succeeds properly."""
        self._set_args({"name": "target_name", "ping": True, "unnamed_discovery": True})
        iscsi_target = NetAppESeriesIscsiTarget()
        with mock.patch(self.REQ_FUNC, side_effect=[(200, self.TARGET_REQUEST_RESPONSE), (200, self.ISCSI_ENTRY_DATA_RESPONSE), (200, [])]):
            self.assertTrue(iscsi_target.apply_target_changes())

        self._set_args({"name": "target_name", "ping": False, "unnamed_discovery": True})
        iscsi_target = NetAppESeriesIscsiTarget()
        with mock.patch(self.REQ_FUNC, side_effect=[(200, self.TARGET_REQUEST_RESPONSE), (200, self.ISCSI_ENTRY_DATA_RESPONSE), (200, [])]):
            self.assertTrue(iscsi_target.apply_target_changes())

        self._set_args({"name": "target_name", "ping": True, "unnamed_discovery": False})
        iscsi_target = NetAppESeriesIscsiTarget()
        with mock.patch(self.REQ_FUNC, side_effect=[(200, self.TARGET_REQUEST_RESPONSE), (200, self.ISCSI_ENTRY_DATA_RESPONSE), (200, [])]):
            self.assertTrue(iscsi_target.apply_target_changes())

        self._set_args({"name": "target_name", "ping": False, "unnamed_discovery": False})
        iscsi_target = NetAppESeriesIscsiTarget()
        with mock.patch(self.REQ_FUNC, side_effect=[(200, self.TARGET_REQUEST_RESPONSE), (200, self.ISCSI_ENTRY_DATA_RESPONSE), (200, [])]):
            self.assertFalse(iscsi_target.apply_target_changes())

    def test_apply_target_changes_fail(self):
        """Ensure apply_iscsi_settings fails properly."""
        self._set_args({"name": "target_name", "ping": True, "unnamed_discovery": True})
        iscsi_target = NetAppESeriesIscsiTarget()

        with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to update the iSCSI target settings."):
            with mock.patch(self.REQ_FUNC, side_effect=[(200, self.TARGET_REQUEST_RESPONSE), (200, self.ISCSI_ENTRY_DATA_RESPONSE), Exception()]):
                iscsi_target.apply_target_changes()

    def test_update_pass(self):
        """Ensure update successfully exists."""
        self._set_args({"name": "target_name", "ping": True, "unnamed_discovery": True})
        iscsi_target = NetAppESeriesIscsiTarget()

        iscsi_target.apply_iscsi_settings = lambda: True
        iscsi_target.apply_target_changes = lambda: True
        with self.assertRaisesRegexp(AnsibleExitJson, r"\'changed\': True"):
            with mock.patch(self.REQ_FUNC, side_effect=[(200, self.TARGET_REQUEST_RESPONSE), (200, self.ISCSI_ENTRY_DATA_RESPONSE)]):
                iscsi_target.update()

        iscsi_target.apply_iscsi_settings = lambda: False
        iscsi_target.apply_target_changes = lambda: True
        with self.assertRaisesRegexp(AnsibleExitJson, r"\'changed\': True"):
            with mock.patch(self.REQ_FUNC, side_effect=[(200, self.TARGET_REQUEST_RESPONSE), (200, self.ISCSI_ENTRY_DATA_RESPONSE)]):
                iscsi_target.update()

        iscsi_target.apply_iscsi_settings = lambda: True
        iscsi_target.apply_target_changes = lambda: False
        with self.assertRaisesRegexp(AnsibleExitJson, r"\'changed\': True"):
            with mock.patch(self.REQ_FUNC, side_effect=[(200, self.TARGET_REQUEST_RESPONSE), (200, self.ISCSI_ENTRY_DATA_RESPONSE)]):
                iscsi_target.update()

        iscsi_target.apply_iscsi_settings = lambda: False
        iscsi_target.apply_target_changes = lambda: False
        with self.assertRaisesRegexp(AnsibleExitJson, r"\'changed\': False"):
            with mock.patch(self.REQ_FUNC, side_effect=[(200, self.TARGET_REQUEST_RESPONSE), (200, self.ISCSI_ENTRY_DATA_RESPONSE)]):
                iscsi_target.update()
