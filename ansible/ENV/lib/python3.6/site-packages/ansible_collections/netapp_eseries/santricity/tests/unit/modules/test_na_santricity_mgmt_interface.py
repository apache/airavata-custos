# (c) 2020, NetApp, Inc
# BSD-3 Clause (see COPYING or https://opensource.org/licenses/BSD-3-Clause)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

from ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_mgmt_interface import NetAppESeriesMgmtInterface
from units.modules.utils import AnsibleExitJson, AnsibleFailJson, ModuleTestCase, set_module_args
from units.compat import mock


class MgmtInterfaceTest(ModuleTestCase):
    REQUIRED_PARAMS = {
        'api_username': 'rw',
        'api_password': 'password',
        'api_url': 'http://localhost',
        'ssid': '1',
    }

    TEST_DATA = [
        {"controllerRef": "070000000000000000000001",
         "controllerSlot": 1,
         "interfaceName": "wan0",
         "interfaceRef": "2800070000000000000000000001000000000000",
         "channel": 1,
         "alias": "creG1g-AP-a",
         "ipv4Enabled": True,
         "ipv4Address": "10.1.1.10",
         "linkStatus": "up",
         "ipv4SubnetMask": "255.255.255.0",
         "ipv4AddressConfigMethod": "configStatic",
         "ipv4GatewayAddress": "10.1.1.1",
         "ipv6Enabled": False,
         "physicalLocation": {"slot": 0},
         "dnsProperties": {"acquisitionProperties": {"dnsAcquisitionType": "stat",
                                                     "dnsServers": [{"addressType": "ipv4",
                                                                     "ipv4Address": "10.1.0.250"},
                                                                    {"addressType": "ipv4",
                                                                     "ipv4Address": "10.10.0.20"}]},
                           "dhcpAcquiredDnsServers": []},
         "ntpProperties": {"acquisitionProperties": {"ntpAcquisitionType": "disabled",
                                                     "ntpServers": None},
                           "dhcpAcquiredNtpServers": []}},
        {"controllerRef": "070000000000000000000001",
         "controllerSlot": 1,
         "interfaceName": "wan1",
         "interfaceRef": "2800070000000000000000000001000000000000",
         "channel": 2,
         "alias": "creG1g-AP-a",
         "ipv4Enabled": True,
         "linkStatus": "down",
         "ipv4Address": "0.0.0.0",
         "ipv4SubnetMask": "0.0.0.0",
         "ipv4AddressConfigMethod": "configDhcp",
         "ipv4GatewayAddress": "10.1.1.1",
         "ipv6Enabled": False,
         "physicalLocation": {"slot": 1},
         "dnsProperties": {"acquisitionProperties": {"dnsAcquisitionType": "stat",
                                                     "dnsServers": [{"addressType": "ipv4",
                                                                     "ipv4Address": "10.1.0.250",
                                                                     "ipv6Address": None},
                                                                    {"addressType": "ipv4",
                                                                     "ipv4Address": "10.10.0.20",
                                                                     "ipv6Address": None}]},
                           "dhcpAcquiredDnsServers": []},
         "ntpProperties": {"acquisitionProperties": {"ntpAcquisitionType": "disabled",
                                                     "ntpServers": None},
                           "dhcpAcquiredNtpServers": []}},
        {"controllerRef": "070000000000000000000002",
         "controllerSlot": 2,
         "interfaceName": "wan0",
         "interfaceRef": "2800070000000000000000000001000000000000",
         "channel": 1,
         "alias": "creG1g-AP-b",
         "ipv4Enabled": True,
         "ipv4Address": "0.0.0.0",
         "linkStatus": "down",
         "ipv4SubnetMask": "0.0.0.0",
         "ipv4AddressConfigMethod": "configDhcp",
         "ipv4GatewayAddress": "10.1.1.1",
         "ipv6Enabled": False,
         "physicalLocation": {"slot": 0},
         "dnsProperties": {"acquisitionProperties": {"dnsAcquisitionType": "stat",
                                                     "dnsServers": [{"addressType": "ipv4",
                                                                     "ipv4Address": "10.1.0.250",
                                                                     "ipv6Address": None}]},
                           "dhcpAcquiredDnsServers": []},
         "ntpProperties": {"acquisitionProperties": {"ntpAcquisitionType": "stat",
                                                     "ntpServers": [{"addrType": "ipvx",
                                                                     "domainName": None,
                                                                     "ipvxAddress": {"addressType": "ipv4",
                                                                                     "ipv4Address": "10.13.1.5",
                                                                                     "ipv6Address": None}},
                                                                    {"addrType": "ipvx",
                                                                     "domainName": None,
                                                                     "ipvxAddress": {"addressType": "ipv4",
                                                                                     "ipv4Address": "10.15.1.8",
                                                                                     "ipv6Address": None}}]},
                           "dhcpAcquiredNtpServers": []}},
        {"controllerRef": "070000000000000000000002",
         "controllerSlot": 2,
         "interfaceName": "wan1",
         "interfaceRef": "2801070000000000000000000001000000000000",
         "channel": 2,
         "alias": "creG1g-AP-b",
         "ipv4Enabled": True,
         "ipv4Address": "0.0.0.0",
         "linkStatus": "down",
         "ipv4SubnetMask": "0.0.0.0",
         "ipv4AddressConfigMethod": "configDhcp",
         "ipv4GatewayAddress": "10.1.1.1",
         "ipv6Enabled": False,
         "physicalLocation": {"slot": 1},
         "dnsProperties": {"acquisitionProperties": {"dnsAcquisitionType": "stat",
                                                     "dnsServers": [{"addressType": "ipv4",
                                                                     "ipv4Address": "10.19.1.2",
                                                                     "ipv6Address": None}]},
                           "dhcpAcquiredDnsServers": []},
         "ntpProperties": {"acquisitionProperties": {"ntpAcquisitionType": "stat",
                                                     "ntpServers": [{"addrType": "ipvx",
                                                                     "domainName": None,
                                                                     "ipvxAddress": {"addressType": "ipv4",
                                                                                     "ipv4Address": "10.13.1.5",
                                                                                     "ipv6Address": None}},
                                                                    {"addrType": "ipvx",
                                                                     "domainName": None,
                                                                     "ipvxAddress": {"addressType": "ipv4",
                                                                                     "ipv4Address": "10.15.1.18",
                                                                                     "ipv6Address": None}}]},
                           "dhcpAcquiredNtpServers": []}}]

    REQ_FUNC = 'ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_mgmt_interface.NetAppESeriesMgmtInterface.request'
    TIME_FUNC = 'ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_mgmt_interface.sleep'

    def _set_args(self, args=None):
        module_args = self.REQUIRED_PARAMS.copy()
        if args is not None:
            module_args.update(args)
        set_module_args(module_args)

    def test_get_controllers_pass(self):
        """Verify dictionary return from get_controllers."""
        initial = {
            "state": "enabled",
            "controller": "A",
            "port": "1",
            "address": "192.168.1.1",
            "subnet_mask": "255.255.255.1",
            "config_method": "static"}
        controller_request = [
            {"physicalLocation": {"slot": 2},
             "controllerRef": "070000000000000000000002",
             "networkSettings": {"remoteAccessEnabled": True}},
            {"physicalLocation": {"slot": 1},
             "controllerRef": "070000000000000000000001",
             "networkSettings": {"remoteAccessEnabled": False}}]
        expected = {
            'A': {'controllerRef': '070000000000000000000001',
                  'controllerSlot': 1, 'ssh': False},
            'B': {'controllerRef': '070000000000000000000002',
                  'controllerSlot': 2, 'ssh': True}}

        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()

        with mock.patch(self.REQ_FUNC, return_value=(200, controller_request)):
            response = mgmt_interface.get_controllers()
            self.assertTrue(response == expected)

    def test_controller_property_fail(self):
        """Verify controllers endpoint request failure causes AnsibleFailJson exception."""
        initial = {
            "state": "enabled",
            "controller": "A",
            "port": "1",
            "address": "192.168.1.1",
            "subnet_mask": "255.255.255.1",
            "config_method": "static"}
        controller_request = [
            {"physicalLocation": {"slot": 2},
             "controllerRef": "070000000000000000000002",
             "networkSettings": {"remoteAccessEnabled": True}},
            {"physicalLocation": {"slot": 1},
             "controllerRef": "070000000000000000000001",
             "networkSettings": {"remoteAccessEnabled": False}}]
        expected = {
            'A': {'controllerRef': '070000000000000000000001',
                  'controllerSlot': 1, 'ssh': False},
            'B': {'controllerRef': '070000000000000000000002',
                  'controllerSlot': 2, 'ssh': True}}

        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to retrieve the controller settings."):
            with mock.patch(self.REQ_FUNC, return_value=Exception):
                response = mgmt_interface.get_controllers()

    def test_update_target_interface_info_pass(self):
        """Verify return value from interface property."""
        initial = {
            "state": "enabled",
            "controller": "A",
            "port": "1",
            "address": "192.168.1.1",
            "subnet_mask": "255.255.255.0",
            "config_method": "static"}
        get_controller = {"A": {"controllerSlot": 1, "controllerRef": "070000000000000000000001", "ssh": False},
                          "B": {"controllerSlot": 2, "controllerRef": "070000000000000000000002", "ssh": True}}
        expected = {"channel": 1, "link_status": "up", "enabled": True, "address": "10.1.1.10", "gateway": "10.1.1.1", "subnet_mask": "255.255.255.0",
                    "dns_config_method": "stat",
                    "dns_servers": [{"addressType": "ipv4", "ipv4Address": "10.1.0.250"}, {"addressType": "ipv4", "ipv4Address": "10.10.0.20"}],
                    "ntp_config_method": "disabled", "ntp_servers": None, "config_method": "configStatic", "controllerRef": "070000000000000000000001",
                    "controllerSlot": 1, "ipv6_enabled": False, "id": "2800070000000000000000000001000000000000", "ssh": False}

        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.get_controllers = lambda: get_controller

        with mock.patch(self.REQ_FUNC, return_value=(200, self.TEST_DATA)):
            mgmt_interface.update_target_interface_info()
            self.assertEquals(mgmt_interface.interface_info, expected)

    def test_interface_property_request_exception_fail(self):
        """Verify ethernet-interfaces endpoint request failure results in AnsibleFailJson exception."""
        initial = {
            "state": "enabled",
            "controller": "A",
            "port": "1",
            "address": "192.168.1.1",
            "subnet_mask": "255.255.255.0",
            "config_method": "static"}
        get_controller = {"A": {"controllerSlot": 1, "controllerRef": "070000000000000000000001", "ssh": False},
                          "B": {"controllerSlot": 2, "controllerRef": "070000000000000000000002", "ssh": True}}

        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.get_controllers = lambda: get_controller

        with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to retrieve defined management interfaces."):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                mgmt_interface.update_target_interface_info()

    def test_update_target_interface_info_fail(self):
        """Verify return value from update_target_interface_info method."""
        initial = {
            "state": "enabled",
            "controller": "A",
            "port": "3",
            "address": "192.168.1.1",
            "subnet_mask": "255.255.255.1",
            "config_method": "static"}
        get_controller = {"A": {"controllerSlot": 1, "controllerRef": "070000000000000000000001", "ssh": False},
                          "B": {"controllerSlot": 2, "controllerRef": "070000000000000000000002", "ssh": True}}

        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.get_controllers = lambda: get_controller

        with self.assertRaisesRegexp(AnsibleFailJson, "Invalid port number! Controller .*? ports:"):
            with mock.patch(self.REQ_FUNC, return_value=(200, self.TEST_DATA)):
                mgmt_interface.update_target_interface_info()

    def test_update_body_enable_interface_setting_pass(self):
        """Validate update_body_enable_interface_setting updates properly."""
        initial = {"state": "enabled", "controller": "A", "port": "1", "address": "192.168.1.1", "subnet_mask": "255.255.255.1", "config_method": "static"}
        interface_info = {"channel": 1, "link_status": "up", "enabled": True, "address": "10.1.1.10", "gateway": "10.1.1.1",
                          "subnet_mask": "255.255.255.0",
                          "dns_config_method": "stat",
                          "dns_servers": [{"addressType": "ipv4", "ipv4Address": "10.1.0.250"},
                                          {"addressType": "ipv4", "ipv4Address": "10.10.0.20"}],
                          "ntp_config_method": "disabled", "ntp_servers": None, "config_method": "configStatic",
                          "controllerRef": "070000000000000000000001",
                          "controllerSlot": 1, "ipv6_enabled": True, "id": "2800070000000000000000000001000000000000", "ssh": False}
        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.interface_info = interface_info
        change_required = mgmt_interface.update_body_enable_interface_setting()
        self.assertFalse(change_required)
        self.assertTrue("ipv4Enabled" in mgmt_interface.body and mgmt_interface.body["ipv4Enabled"])

        initial = {"state": "disabled", "controller": "A", "port": "1", "address": "192.168.1.1", "subnet_mask": "255.255.255.1", "config_method": "static"}
        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.interface_info = interface_info
        change_required = mgmt_interface.update_body_enable_interface_setting()
        self.assertTrue(change_required)
        self.assertTrue("ipv4Enabled" in mgmt_interface.body and not mgmt_interface.body["ipv4Enabled"])

    def test_update_body_enable_interface_setting_fail(self):
        """Validate update_body_enable_interface_setting throws expected exception"""
        initial = {"state": "disabled", "controller": "A", "port": "1", "address": "192.168.1.1", "subnet_mask": "255.255.255.1", "config_method": "static"}
        interface_info = {"channel": 1, "link_status": "up", "enabled": True, "address": "10.1.1.10", "gateway": "10.1.1.1",
                          "subnet_mask": "255.255.255.0",
                          "dns_config_method": "stat",
                          "dns_servers": [{"addressType": "ipv4", "ipv4Address": "10.1.0.250"},
                                          {"addressType": "ipv4", "ipv4Address": "10.10.0.20"}],
                          "ntp_config_method": "disabled", "ntp_servers": None, "config_method": "configStatic",
                          "controllerRef": "070000000000000000000001",
                          "controllerSlot": 1, "ipv6_enabled": False, "id": "2800070000000000000000000001000000000000", "ssh": False}

        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.interface_info = interface_info
        with self.assertRaisesRegexp(AnsibleFailJson, "Either IPv4 or IPv6 must be enabled."):
            mgmt_interface.update_body_enable_interface_setting()

    def test_update_body_interface_settings_fail(self):
        """Validate update_body_interface_settings throws expected exception"""
        initial = {"state": "enabled", "controller": "A", "port": "1", "address": "192.168.1.1", "subnet_mask": "255.255.255.1", "config_method": "static"}
        interface_info = {"channel": 1, "link_status": "up", "enabled": True, "address": "10.1.1.10", "gateway": "10.1.1.1",
                          "subnet_mask": "255.255.255.0",
                          "dns_config_method": "stat",
                          "dns_servers": [{"addressType": "ipv4", "ipv4Address": "10.1.0.250"},
                                          {"addressType": "ipv4", "ipv4Address": "10.10.0.20"}],
                          "ntp_config_method": "disabled", "ntp_servers": None, "config_method": "configStatic",
                          "controllerRef": "070000000000000000000001",
                          "controllerSlot": 1, "ipv6_enabled": False, "id": "2800070000000000000000000001000000000000", "ssh": False}

        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.interface_info = interface_info
        self.assertTrue(mgmt_interface.update_body_interface_settings())
        self.assertEquals(mgmt_interface.body, {"ipv4AddressConfigMethod": "configStatic", "ipv4Address": "192.168.1.1", "ipv4SubnetMask": "255.255.255.1"})

        initial = {"state": "enabled", "controller": "A", "port": "1", "address": "192.168.1.100", "subnet_mask": "255.255.255.1", "gateway": "192.168.1.1",
                   "config_method": "static"}
        interface_info = {"channel": 1, "link_status": "up", "enabled": True, "address": "10.1.1.10", "gateway": "10.1.1.1",
                          "subnet_mask": "255.255.255.0",
                          "dns_config_method": "stat",
                          "dns_servers": [{"addressType": "ipv4", "ipv4Address": "10.1.0.250"},
                                          {"addressType": "ipv4", "ipv4Address": "10.10.0.20"}],
                          "ntp_config_method": "disabled", "ntp_servers": None, "config_method": "configStatic",
                          "controllerRef": "070000000000000000000001",
                          "controllerSlot": 1, "ipv6_enabled": False, "id": "2800070000000000000000000001000000000000", "ssh": False}

        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.interface_info = interface_info
        self.assertTrue(mgmt_interface.update_body_interface_settings())
        self.assertEquals(mgmt_interface.body, {"ipv4AddressConfigMethod": "configStatic", "ipv4Address": "192.168.1.100", "ipv4SubnetMask": "255.255.255.1",
                                                "ipv4GatewayAddress": "192.168.1.1"})

        initial = {"state": "enabled", "controller": "A", "port": "1", "config_method": "dhcp"}
        interface_info = {"channel": 1, "link_status": "up", "enabled": True, "address": "10.1.1.10", "gateway": "10.1.1.1",
                          "subnet_mask": "255.255.255.0",
                          "dns_config_method": "stat",
                          "dns_servers": [{"addressType": "ipv4", "ipv4Address": "10.1.0.250"},
                                          {"addressType": "ipv4", "ipv4Address": "10.10.0.20"}],
                          "ntp_config_method": "disabled", "ntp_servers": None, "config_method": "configStatic",
                          "controllerRef": "070000000000000000000001",
                          "controllerSlot": 1, "ipv6_enabled": False, "id": "2800070000000000000000000001000000000000", "ssh": False}

        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.interface_info = interface_info
        self.assertTrue(mgmt_interface.update_body_interface_settings())
        self.assertEquals(mgmt_interface.body, {"ipv4AddressConfigMethod": "configDhcp"})

        initial = {"state": "enabled", "controller": "A", "port": "1", "config_method": "dhcp"}
        interface_info = {"channel": 1, "link_status": "up", "enabled": True, "address": "10.1.1.10", "gateway": "10.1.1.1",
                          "subnet_mask": "255.255.255.0",
                          "dns_config_method": "stat",
                          "dns_servers": [{"addressType": "ipv4", "ipv4Address": "10.1.0.250"},
                                          {"addressType": "ipv4", "ipv4Address": "10.10.0.20"}],
                          "ntp_config_method": "disabled", "ntp_servers": None, "config_method": "configDhcp",
                          "controllerRef": "070000000000000000000001",
                          "controllerSlot": 1, "ipv6_enabled": False, "id": "2800070000000000000000000001000000000000", "ssh": False}

        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.interface_info = interface_info
        self.assertFalse(mgmt_interface.update_body_interface_settings())
        self.assertEquals(mgmt_interface.body, {"ipv4AddressConfigMethod": "configDhcp"})

    def test_update_body_dns_server_settings_pass(self):
        """Validate update_body_dns_server_settings throws expected exception"""
        interface_info = {"channel": 1, "link_status": "up", "enabled": True, "address": "10.1.1.10", "gateway": "10.1.1.1",
                          "subnet_mask": "255.255.255.0",
                          "dns_config_method": "stat",
                          "dns_servers": [{"addressType": "ipv4", "ipv4Address": "10.1.0.250"},
                                          {"addressType": "ipv4", "ipv4Address": "10.10.0.20"}],
                          "ntp_config_method": "disabled", "ntp_servers": None, "config_method": "configStatic",
                          "controllerRef": "070000000000000000000001",
                          "controllerSlot": 1, "ipv6_enabled": False, "id": "2800070000000000000000000001000000000000", "ssh": False}

        initial = {"state": "enabled", "controller": "A", "port": "1", "dns_config_method": "dhcp"}
        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.interface_info = interface_info
        self.assertTrue(mgmt_interface.update_body_dns_server_settings())
        self.assertEquals(mgmt_interface.body, {"dnsAcquisitionDescriptor": {"dnsAcquisitionType": "dhcp"}})

        initial = {"state": "enabled", "controller": "A", "port": "1", "dns_config_method": "static", "dns_address": "192.168.1.100"}
        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.interface_info = interface_info
        self.assertTrue(mgmt_interface.update_body_dns_server_settings())
        self.assertEquals(mgmt_interface.body, {"dnsAcquisitionDescriptor": {"dnsAcquisitionType": "stat",
                                                                             "dnsServers": [{"addressType": "ipv4", "ipv4Address": "192.168.1.100"}]}})

        initial = {"state": "enabled", "controller": "A", "port": "1", "dns_config_method": "static", "dns_address": "192.168.1.100",
                   "dns_address_backup": "192.168.1.102"}
        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.interface_info = interface_info
        self.assertTrue(mgmt_interface.update_body_dns_server_settings())
        self.assertEquals(mgmt_interface.body, {"dnsAcquisitionDescriptor": {"dnsAcquisitionType": "stat",
                                                                             "dnsServers": [{"addressType": "ipv4", "ipv4Address": "192.168.1.100"},
                                                                                            {"addressType": "ipv4", "ipv4Address": "192.168.1.102"}]}})

    def test_update_body_ntp_server_settings_pass(self):
        """Validate update_body_ntp_server_settings throws expected exception"""
        interface_info = {"channel": 1, "link_status": "up", "enabled": True, "address": "10.1.1.10", "gateway": "10.1.1.1",
                          "subnet_mask": "255.255.255.0",
                          "dns_config_method": "stat",
                          "dns_servers": [{"addressType": "ipv4", "ipv4Address": "10.1.0.250"},
                                          {"addressType": "ipv4", "ipv4Address": "10.10.0.20"}],
                          "ntp_config_method": "dhcp", "ntp_servers": None, "config_method": "configStatic",
                          "controllerRef": "070000000000000000000001",
                          "controllerSlot": 1, "ipv6_enabled": False, "id": "2800070000000000000000000001000000000000", "ssh": False}

        initial = {"state": "enabled", "controller": "A", "port": "1", "ntp_config_method": "disabled"}
        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.interface_info = interface_info
        self.assertTrue(mgmt_interface.update_body_ntp_server_settings())
        self.assertEquals(mgmt_interface.body, {"ntpAcquisitionDescriptor": {"ntpAcquisitionType": "disabled"}})

        initial = {"state": "enabled", "controller": "A", "port": "1", "ntp_config_method": "dhcp"}
        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.interface_info = interface_info
        self.assertFalse(mgmt_interface.update_body_ntp_server_settings())
        self.assertEquals(mgmt_interface.body, {"ntpAcquisitionDescriptor": {"ntpAcquisitionType": "dhcp"}})

        initial = {"state": "enabled", "controller": "A", "port": "1", "ntp_config_method": "static", "ntp_address": "192.168.1.200"}
        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.interface_info = interface_info
        self.assertTrue(mgmt_interface.update_body_ntp_server_settings())
        self.assertEquals(mgmt_interface.body, {"ntpAcquisitionDescriptor": {
            "ntpAcquisitionType": "stat", "ntpServers": [{"addrType": "ipvx", "ipvxAddress": {"addressType": "ipv4", "ipv4Address": "192.168.1.200"}}]}})

        initial = {"state": "enabled", "controller": "A", "port": "1", "ntp_config_method": "static", "ntp_address": "192.168.1.200",
                   "ntp_address_backup": "192.168.1.202"}
        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.interface_info = interface_info
        self.assertTrue(mgmt_interface.update_body_ntp_server_settings())
        self.assertEquals(mgmt_interface.body, {"ntpAcquisitionDescriptor": {
            "ntpAcquisitionType": "stat", "ntpServers": [{"addrType": "ipvx", "ipvxAddress": {"addressType": "ipv4", "ipv4Address": "192.168.1.200"}},
                                                         {"addrType": "ipvx", "ipvxAddress": {"addressType": "ipv4", "ipv4Address": "192.168.1.202"}}]}})

    def test_update_body_ssh_setting_pass(self):
        """Validate update_body_ssh_setting throws expected exception"""
        interface_info = {"channel": 1, "link_status": "up", "enabled": True, "address": "10.1.1.10", "gateway": "10.1.1.1",
                          "subnet_mask": "255.255.255.0",
                          "dns_config_method": "stat",
                          "dns_servers": [{"addressType": "ipv4", "ipv4Address": "10.1.0.250"},
                                          {"addressType": "ipv4", "ipv4Address": "10.10.0.20"}],
                          "ntp_config_method": "disabled", "ntp_servers": None, "config_method": "configStatic",
                          "controllerRef": "070000000000000000000001",
                          "controllerSlot": 1, "ipv6_enabled": False, "id": "2800070000000000000000000001000000000000", "ssh": False}

        initial = {"state": "enabled", "controller": "A", "port": "1", "config_method": "dhcp", "ssh": True}
        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.interface_info = interface_info
        self.assertTrue(mgmt_interface.update_body_ssh_setting())
        self.assertEquals(mgmt_interface.body, {"enableRemoteAccess": True})

        initial = {"state": "enabled", "controller": "A", "port": "1", "config_method": "dhcp", "ssh": False}
        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.interface_info = interface_info
        self.assertFalse(mgmt_interface.update_body_ssh_setting())
        self.assertEquals(mgmt_interface.body, {"enableRemoteAccess": False})

    def test_update_url_pass(self):
        """Verify update_url returns expected url."""
        initial = {"state": "enabled", "controller": "A", "port": "1", "config_method": "dhcp", "ssh": False}
        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.url = "https://192.168.1.100:8443/devmgr/v2/"
        mgmt_interface.alt_interface_addresses = ["192.168.1.102"]
        mgmt_interface.update_url()
        self.assertTrue(mgmt_interface.url, "https://192.168.1.102:8443/devmgr/v2/")

    def test_update_pass(self):
        """Verify update successfully completes."""
        initial = {"state": "enabled", "controller": "A", "port": "1", "config_method": "dhcp", "ssh": False}
        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.update_request_body = lambda: False
        mgmt_interface.is_embedded = lambda: False
        mgmt_interface.use_alternate_address = False
        with self.assertRaisesRegexp(AnsibleExitJson, "No changes are required."):
            with mock.patch(self.REQ_FUNC, return_value=(200, None)):
                mgmt_interface.update()

        def update_request_body():
            update_request_body.value = not update_request_body.value
            return update_request_body.value
        update_request_body.value = False

        initial = {"state": "enabled", "controller": "A", "port": "1", "config_method": "dhcp", "ssh": False}
        self._set_args(initial)
        mgmt_interface = NetAppESeriesMgmtInterface()
        mgmt_interface.update_request_body = update_request_body
        mgmt_interface.is_embedded = lambda: True
        mgmt_interface.use_alternate_address = False
        with self.assertRaisesRegexp(AnsibleExitJson, "The interface settings have been updated."):
            with mock.patch(self.REQ_FUNC, return_value=(200, None)):
                mgmt_interface.update()
