# (c) 2020, NetApp, Inc
# BSD-3 Clause (see COPYING or https://opensource.org/licenses/BSD-3-Clause)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

from ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_nvme_interface import NetAppESeriesNvmeInterface
from units.modules.utils import AnsibleExitJson, AnsibleFailJson, ModuleTestCase, set_module_args
from units.compat import mock


class NvmeInterfaceTest(ModuleTestCase):
    REQUIRED_PARAMS = {"api_username": "rw",
                       "api_password": "password",
                       "api_url": "http://localhost",
                       "ssid": "1",
                       "state": "enabled",
                       "controller": "A",
                       "channel": 1}

    REQ_FUNC = "ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_nvme_interface.NetAppESeriesNvmeInterface.request"

    def _set_args(self, args=None):
        module_args = self.REQUIRED_PARAMS.copy()
        if args is not None:
            module_args.update(args)
        set_module_args(module_args)

    def test_valid_options_pass(self):
        """Verify valid options."""
        valid_option_list = [{"state": "enabled", "config_method": "static", "address": "192.168.1.100", "subnet_mask": "255.255.255.0",
                              "gateway": "192.168.1.1", "mtu": 1500},
                             {"address": "192.168.1.100"},
                             {"state": "enabled", "config_method": "dhcp", "mtu": 1500},
                             {"state": "disabled"}]

        for option in valid_option_list:
            self._set_args(option)
            nvme = NetAppESeriesNvmeInterface()

    def test_invalid_options_fail(self):
        """Verify invalid options throw expected exceptions."""
        invalid_option_list = [{"state": "enabled", "config_method": "static", "address": "1920.168.1.100", "subnet_mask": "255.255.255.0",
                                "gateway": "192.168.1.1", "mtu": 1500},
                               {"state": "enabled", "config_method": "static", "address": "192.168.1.100", "subnet_mask": "255.2550.255.0",
                                "gateway": "192.168.1.1", "mtu": 1500},
                               {"state": "enabled", "config_method": "static", "address": "192.168.1.100", "subnet_mask": "255.255.255.0",
                                "gateway": "192.168..100", "mtu": 1500},
                               {"state": "enabled", "config_method": "static", "address": "192.168.1.100", "subnet_mask": "2550.255.255.0",
                                "gateway": "192.168.1.1000", "mtu": 1500}]

        for option in invalid_option_list:
            self._set_args(option)
            with self.assertRaises(AnsibleFailJson):
                nvme = NetAppESeriesNvmeInterface()

    def test_get_nvmeof_interfaces_pass(self):
        """Verify get_nvmeof_interfaces method returns the expected list of interface values."""
        options = {"address": "192.168.1.100"}
        response = [{"controllerRef": "070000000000000000000001", "interfaceRef": "2201020000000000000000000000000000000000",
                     "ioInterfaceTypeData": {"interfaceType": "ib",
                                             "ib": {"interfaceRef": "2201020000000000000000000000000000000000", "channel": 1, "linkState": "up"}},
                     "commandProtocolPropertiesList": {"commandProtocolProperties": [
                         {"commandProtocol": "nvme", "nvmeProperties": {"commandSet": "nvmeof", "nvmeofProperties": {
                             "provider": "providerInfiniband", "ibProperties": {"ipAddressData": {
                                 "addressType": "ipv4", "ipv4Data": {"configState": "configured", "ipv4Address": "192.168.1.100"}}}}}}]}}]
        self._set_args(options)
        nvme = NetAppESeriesNvmeInterface()
        with mock.patch(self.REQ_FUNC, return_value=(200, response)):
            self.assertEquals(nvme.get_nvmeof_interfaces(), [
                {"properties": {"provider": "providerInfiniband", "ibProperties": {
                    "ipAddressData": {"addressType": "ipv4",
                                      "ipv4Data": {"configState": "configured", "ipv4Address": "192.168.1.100"}}}},
                 "reference": "2201020000000000000000000000000000000000", "channel": 1, "interface_type": "ib",
                 "interface": {"interfaceRef": "2201020000000000000000000000000000000000", "channel": 1,
                               "linkState": "up"}, "controller_id": "070000000000000000000001",
                 "link_status": "up"}])

    def test_get_nvmeof_interfaces_fail(self):
        """Verify get_nvmeof_interfaces method throws the expected exceptions."""
        options = {"address": "192.168.1.100"}
        self._set_args(options)
        nvme = NetAppESeriesNvmeInterface()
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to retrieve defined host interfaces."):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                nvme.get_nvmeof_interfaces()

    def test_get_target_interface_pass(self):
        """Verify get_target_interface returns the expected interface."""
        # options = {"state": "enabled", "config_method": "static", "address": "192.168.1.100", "subnet_mask": "255.255.255.0",
        #            "gateway": "192.168.1.1", "mtu": 1500}
        options = {"address": "192.168.1.200"}
        self._set_args(options)
        nvme = NetAppESeriesNvmeInterface()
        nvme.get_nvmeof_interfaces = lambda: [
            {"properties": {"provider": "providerInfiniband", "ibProperties": {
                "ipAddressData": {"addressType": "ipv4",
                                  "ipv4Data": {"configState": "configured", "ipv4Address": "192.168.1.100"}}}},
             "reference": "2201020000000000000000000000000000000000", "channel": 5,
             "interface_type": {"interfaceRef": "2201020000000000000000000000000000000000", "channel": 5,
                                "linkState": "up"}, "controller_id": "070000000000000000000001",
             "link_status": "up"},
            {"properties": {"provider": "providerInfiniband", "ibProperties": {
                "ipAddressData": {"addressType": "ipv4",
                                  "ipv4Data": {"configState": "configured", "ipv4Address": "192.168.2.100"}}}},
             "reference": "2201030000000000000000000000000000000000", "channel": 4,
             "interface_type": {"interfaceRef": "2201030000000000000000000000000000000000", "channel": 4,
                                "linkState": "up"}, "controller_id": "070000000000000000000001",
             "link_status": "up"},
            {"properties": {"provider": "providerInfiniband", "ibProperties": {
                "ipAddressData": {"addressType": "ipv4",
                                  "ipv4Data": {"configState": "configured", "ipv4Address": "192.168.3.100"}}}},
             "reference": "2201040000000000000000000000000000000000", "channel": 6,
             "interface_type": {"interfaceRef": "2201040000000000000000000000000000000000", "channel": 6,
                                "linkState": "down"}, "controller_id": "070000000000000000000001",
             "link_status": "up"}]
        nvme.get_controllers = lambda: {"A": "070000000000000000000001", "B": "070000000000000000000002"}
        self.assertEqual(nvme.get_target_interface(), {
            "properties": {"provider": "providerInfiniband", "ibProperties": {
                "ipAddressData": {"addressType": "ipv4",
                                  "ipv4Data": {"configState": "configured", "ipv4Address": "192.168.2.100"}}}},
            "reference": "2201030000000000000000000000000000000000", "channel": 4,
            "interface_type": {"interfaceRef": "2201030000000000000000000000000000000000", "channel": 4,
                               "linkState": "up"}, "controller_id": "070000000000000000000001",
            "link_status": "up"})

    def test_get_target_interface_fail(self):
        """Verify get_target_interface method throws the expected exceptions."""
        options = {"address": "192.168.1.200", "channel": "0"}
        self._set_args(options)
        nvme = NetAppESeriesNvmeInterface()
        nvme.get_nvmeof_interfaces = lambda: [
            {"properties": {"provider": "providerInfiniband", "ibProperties": {
                "ipAddressData": {"addressType": "ipv4",
                                  "ipv4Data": {"configState": "configured", "ipv4Address": "192.168.1.100"}}}},
             "reference": "2201020000000000000000000000000000000000", "channel": 5,
             "interface_type": {"interfaceRef": "2201020000000000000000000000000000000000", "channel": 5,
                                "linkState": "up"}, "controller_id": "070000000000000000000001",
             "link_status": "up"}]
        nvme.get_controllers = lambda: {"A": "070000000000000000000001", "B": "070000000000000000000002"}
        with self.assertRaisesRegexp(AnsibleFailJson, "Invalid controller .*? NVMe channel."):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                nvme.get_target_interface()

        options = {"address": "192.168.1.200", "channel": "2"}
        self._set_args(options)
        nvme = NetAppESeriesNvmeInterface()
        nvme.get_nvmeof_interfaces = lambda: [
            {"properties": {"provider": "providerInfiniband", "ibProperties": {
                "ipAddressData": {"addressType": "ipv4",
                                  "ipv4Data": {"configState": "configured", "ipv4Address": "192.168.1.100"}}}},
             "reference": "2201020000000000000000000000000000000000", "channel": 5,
             "interface_type": {"interfaceRef": "2201020000000000000000000000000000000000", "channel": 5,
                                "linkState": "up"}, "controller_id": "070000000000000000000001",
             "link_status": "up"}]
        nvme.get_controllers = lambda: {"A": "070000000000000000000001", "B": "070000000000000000000002"}
        with self.assertRaisesRegexp(AnsibleFailJson, "Invalid controller .*? NVMe channel."):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                nvme.get_target_interface()

    def test_update_pass(self):
        """Verify update successfully completes"""
        # options = {"state": "enabled", "config_method": "static", "address": "192.168.1.100", "subnet_mask": "255.255.255.0",
        #            "gateway": "192.168.1.1", "mtu": 1500}
        options = {"address": "192.168.1.200"}
        iface = {"properties": {"provider": "providerInfiniband",
                                "ibProperties": {"ipAddressData": {"addressType": "ipv4",
                                                                   "ipv4Data": {"configState": "configured", "ipv4Address": "192.168.1.100"}}}},
                 "reference": "2201020000000000000000000000000000000000", "channel": 5, "interface_type": "ib", "controllerRef": "070000000000000000000001",
                 "link_status": "up"}
        self._set_args(options)
        nvme = NetAppESeriesNvmeInterface()
        nvme.get_target_interface = lambda: iface
        with self.assertRaisesRegexp(AnsibleExitJson, "NVMeoF interface settings have been updated."):
            with mock.patch(self.REQ_FUNC, return_value=(200, None)):
                nvme.update()

        options = {"address": "192.168.1.200"}
        iface = {"properties": {"provider": "providerInfiniband",
                                "ibProperties": {"ipAddressData": {"addressType": "ipv4",
                                                                   "ipv4Data": {"configState": "configured", "ipv4Address": "192.168.1.100"}}}},
                 "reference": "2201020000000000000000000000000000000000", "channel": 5, "interface_type": "ib", "controllerRef": "070000000000000000000001",
                 "link_status": "up"}
        self._set_args(options)
        nvme = NetAppESeriesNvmeInterface()
        nvme.module.check_mode = True
        nvme.get_target_interface = lambda: iface
        with self.assertRaisesRegexp(AnsibleExitJson, "No changes have been made."):
            with mock.patch(self.REQ_FUNC, return_value=(200, None)):
                nvme.update()

        options = {"address": "192.168.1.100"}
        iface = {"properties": {"provider": "providerInfiniband",
                                "ibProperties": {"ipAddressData": {"addressType": "ipv4",
                                                                   "ipv4Data": {"configState": "configured", "ipv4Address": "192.168.1.100"}}}},
                 "reference": "2201020000000000000000000000000000000000", "channel": 5, "interface_type": "ib", "controllerRef": "070000000000000000000001",
                 "link_status": "up"}
        self._set_args(options)
        nvme = NetAppESeriesNvmeInterface()
        nvme.get_target_interface = lambda: iface

        with self.assertRaisesRegexp(AnsibleExitJson, "No changes have been made."):
            with mock.patch(self.REQ_FUNC, return_value=(200, None)):
                nvme.update()

    def test_update_fail(self):
        """Verify update throws expected exception."""
        # options = {"state": "enabled", "config_method": "static", "address": "192.168.1.100", "subnet_mask": "255.255.255.0",
        #            "gateway": "192.168.1.1", "mtu": 1500}
        options = {"address": "192.168.1.200"}
        iface = {"properties": {"provider": "providerInfiniband",
                                "ibProperties": {"ipAddressData": {"addressType": "ipv4",
                                                                   "ipv4Data": {"configState": "configured", "ipv4Address": "192.168.1.100"}}}},
                 "reference": "2201020000000000000000000000000000000000", "channel": 5, "interface_type": "ib", "controllerRef": "070000000000000000000001",
                 "link_status": "up"}
        self._set_args(options)
        nvme = NetAppESeriesNvmeInterface()
        nvme.get_target_interface = lambda: iface
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to configure interface."):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                nvme.update()
