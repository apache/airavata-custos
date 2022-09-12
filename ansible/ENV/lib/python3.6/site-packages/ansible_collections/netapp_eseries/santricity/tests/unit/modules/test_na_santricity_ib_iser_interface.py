# (c) 2020, NetApp, Inc
# BSD-3 Clause (see COPYING or https://opensource.org/licenses/BSD-3-Clause)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

from ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_ib_iser_interface import NetAppESeriesIbIserInterface
from units.modules.utils import AnsibleExitJson, AnsibleFailJson, ModuleTestCase, set_module_args
from units.compat import mock


class NvmeInterfaceTest(ModuleTestCase):
    REQUIRED_PARAMS = {"api_username": "rw",
                       "api_password": "password",
                       "api_url": "http://localhost",
                       "ssid": "1",
                       "controller": "A",
                       "channel": 1}

    REQ_FUNC = "ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_ib_iser_interface.NetAppESeriesIbIserInterface.request"

    def _set_args(self, args=None):
        module_args = self.REQUIRED_PARAMS.copy()
        if args is not None:
            module_args.update(args)
        set_module_args(module_args)

    def test_invalid_options_fail(self):
        """Verify invalid options fail."""
        options_list = [{"address": "nonaddress@somewhere.com"},
                        {"address": "192.168.100.1000"},
                        {"address": "1192.168.100.100"}]

        for options in options_list:
            self._set_args(options)
            with self.assertRaisesRegexp(AnsibleFailJson, "An invalid ip address was provided for address."):
                iface = NetAppESeriesIbIserInterface()

    def test_get_interfaces_pass(self):
        """Verify get_interfaces method passes."""
        self._set_args({"address": "192.168.100.100"})
        iface = NetAppESeriesIbIserInterface()
        with mock.patch(self.REQ_FUNC, return_value=(200, [{"interfaceType": "iscsi", "iscsi": {"interfaceData": {"type": "infiniband",
                                                                                                                  "infinibandData": {"isIser": True}}}},
                                                           {"interfaceType": "iscsi", "iscsi": {"interfaceData": {"type": "infiniband",
                                                                                                                  "infinibandData": {"isIser": True}}}},
                                                           {"interfaceType": "fc", "fc": {}}])):
            self.assertEquals(iface.get_interfaces(),
                              [{'interfaceType': 'iscsi', 'iscsi': {'interfaceData': {'type': 'infiniband', 'infinibandData': {'isIser': True}}}},
                               {'interfaceType': 'iscsi', 'iscsi': {'interfaceData': {'type': 'infiniband', 'infinibandData': {'isIser': True}}}}])

    def test_get_interfaces_fails(self):
        """Verify get_interfaces method throws expected exceptions."""
        self._set_args({"address": "192.168.100.100"})
        iface = NetAppESeriesIbIserInterface()
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to retrieve defined host interfaces."):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                iface.get_interfaces()

        self._set_args({"address": "192.168.100.100"})
        iface = NetAppESeriesIbIserInterface()
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to detect any InfiniBand iSER interfaces!"):
            with mock.patch(self.REQ_FUNC, return_value=(200, [{"interfaceType": "eth", "eth": {"interfaceData": {"type": "ethernet",
                                                                                                                  "infinibandData": {"isIser": False}}}},
                                                               {"interfaceType": "iscsi", "iscsi": {"interfaceData": {"type": "infiniband",
                                                                                                                      "infinibandData": {"isIser": False}}}},
                                                               {"interfaceType": "fc", "fc": {}}])):
                iface.get_interfaces()

    def test_get_ib_link_status_pass(self):
        """Verify expected data structure."""
        self._set_args({"address": "192.168.100.100"})
        iface = NetAppESeriesIbIserInterface()
        with mock.patch(self.REQ_FUNC, return_value=(200, {"ibPorts": [{"channelPortRef": 1, "linkState": "active"},
                                                                       {"channelPortRef": 2, "linkState": "down"},
                                                                       {"channelPortRef": 3, "linkState": "down"},
                                                                       {"channelPortRef": 4, "linkState": "active"}]})):
            self.assertEquals(iface.get_ib_link_status(), {1: 'active', 2: 'down', 3: 'down', 4: 'active'})

    def test_get_ib_link_status_fail(self):
        """Verify expected exception is thrown."""
        self._set_args({"address": "192.168.100.100"})
        iface = NetAppESeriesIbIserInterface()
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to retrieve ib link status information!"):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                statuses = iface.get_ib_link_status()

    def test_is_change_required_pass(self):
        """Verify is_change_required method returns expected values."""
        self._set_args({"address": "192.168.100.100"})
        iface = NetAppESeriesIbIserInterface()
        iface.get_target_interface = lambda: {"iscsi": {"ipv4Data": {"ipv4AddressData": {"ipv4Address": "192.168.1.1"}}}}
        self.assertTrue(iface.is_change_required())

        self._set_args({"address": "192.168.100.100"})
        iface = NetAppESeriesIbIserInterface()
        iface.get_target_interface = lambda: {"iscsi": {"ipv4Data": {"ipv4AddressData": {"ipv4Address": "192.168.100.100"}}}}
        self.assertFalse(iface.is_change_required())

    def test_make_request_body_pass(self):
        """Verify expected request body."""
        self._set_args({"address": "192.168.100.100"})
        iface = NetAppESeriesIbIserInterface()
        iface.get_target_interface = lambda: {"iscsi": {"id": "1234", "ipv4Data": {"ipv4AddressData": {"ipv4Address": "192.168.1.1"}}}}
        self.assertEquals(iface.make_request_body(), {"iscsiInterface": "1234",
                                                      "settings": {"tcpListenPort": [],
                                                                   "ipv4Address": ["192.168.100.100"],
                                                                   "ipv4SubnetMask": [],
                                                                   "ipv4GatewayAddress": [],
                                                                   "ipv4AddressConfigMethod": [],
                                                                   "maximumFramePayloadSize": [],
                                                                   "ipv4VlanId": [],
                                                                   "ipv4OutboundPacketPriority": [],
                                                                   "ipv4Enabled": [],
                                                                   "ipv6Enabled": [],
                                                                   "ipv6LocalAddresses": [],
                                                                   "ipv6RoutableAddresses": [],
                                                                   "ipv6PortRouterAddress": [],
                                                                   "ipv6AddressConfigMethod": [],
                                                                   "ipv6OutboundPacketPriority": [],
                                                                   "ipv6VlanId": [],
                                                                   "ipv6HopLimit": [],
                                                                   "ipv6NdReachableTime": [],
                                                                   "ipv6NdRetransmitTime": [],
                                                                   "ipv6NdStaleTimeout": [],
                                                                   "ipv6DuplicateAddressDetectionAttempts": [],
                                                                   "maximumInterfaceSpeed": []}})

    def test_update_pass(self):
        """Verify update method behavior."""
        self._set_args({"address": "192.168.100.100"})
        iface = NetAppESeriesIbIserInterface()
        iface.is_change_required = lambda: False
        with self.assertRaisesRegexp(AnsibleExitJson, "No changes were required."):
            iface.update()

        self._set_args({"address": "192.168.100.100"})
        iface = NetAppESeriesIbIserInterface()
        iface.is_change_required = lambda: True
        iface.check_mode = True
        with self.assertRaisesRegexp(AnsibleExitJson, "No changes were required."):
            iface.update()

        self._set_args({"address": "192.168.100.100"})
        iface = NetAppESeriesIbIserInterface()
        iface.is_change_required = lambda: True
        iface.make_request_body = lambda: {}
        with self.assertRaisesRegexp(AnsibleExitJson, "The interface settings have been updated."):
            with mock.patch(self.REQ_FUNC, return_value=(200, None)):
                iface.update()

    def test_update_fail(self):
        """Verify exceptions are thrown."""
        self._set_args({"address": "192.168.100.100"})
        iface = NetAppESeriesIbIserInterface()
        iface.is_change_required = lambda: True
        iface.make_request_body = lambda: {}
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to modify the interface!"):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                iface.update()
