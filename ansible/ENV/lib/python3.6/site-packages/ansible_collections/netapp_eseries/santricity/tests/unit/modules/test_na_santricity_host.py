# (c) 2020, NetApp, Inc
# BSD-3 Clause (see COPYING or https://opensource.org/licenses/BSD-3-Clause)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

from ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_host import NetAppESeriesHost
from units.modules.utils import AnsibleExitJson, AnsibleFailJson, ModuleTestCase, set_module_args
from units.compat import mock


class HostTest(ModuleTestCase):
    REQUIRED_PARAMS = {
        'api_username': 'rw',
        'api_password': 'password',
        'api_url': 'http://localhost',
        'ssid': '1',
        'name': '1',
    }
    HOST = {
        'name': '1',
        'hostRef': '123',
        'label': '1',
        'id': '0' * 30,
        'clusterRef': 40 * '0',
        'hostTypeIndex': 28,
        'hostSidePorts': [],
        'initiators': [],
        'ports': [],
    }
    HOST_ALT = {
        'name': '2',
        'label': '2',
        'id': '1' * 30,
        'clusterRef': '1',
        'hostSidePorts': [],
        'initiators': [],
        'ports': [],
    }
    EXISTING_HOSTS = [
        {"hostRef": "84000000600A098000A4B28D00303D065D430118", "clusterRef": "0000000000000000000000000000000000000000", "label": "Beegfs_storage1",
         "hostTypeIndex": 28, "ports": [], "initiators": [{"initiatorRef": "89000000600A098000A4B28D00303CF55D4300E3",
                                                           "nodeName": {"ioInterfaceType": "iscsi",
                                                                        "iscsiNodeName": "iqn.1993-08.org.debian.beegfs-storage1:01:b0621126818",
                                                                        "remoteNodeWWN": None, "nvmeNodeName": None},
                                                           "alias": {"ioInterfaceType": "iscsi", "iscsiAlias": ""}, "label": "beegfs_storage1_iscsi_0",
                                                           "hostRef": "84000000600A098000A4B28D00303D065D430118",
                                                           "id": "89000000600A098000A4B28D00303CF55D4300E3"}],
         "hostSidePorts": [{"type": "iscsi", "address": "iqn.1993-08.org.debian.beegfs-storage1:01:b0621126818", "label": "beegfs_storage1_iscsi_0"}],
         "id": "84000000600A098000A4B28D00303D065D430118", "name": "beegfs_storage1"},
        {"hostRef": "84000000600A098000A4B9D10030370B5D430109", "clusterRef": "0000000000000000000000000000000000000000", "label": "beegfs_metadata1",
         "hostTypeIndex": 28, "ports": [], "initiators": [{"initiatorRef": "89000000600A098000A4B28D00303CFC5D4300F7",
                                                           "nodeName": {"ioInterfaceType": "iscsi",
                                                                        "iscsiNodeName": "iqn.1993-08.org.debian.beegfs-metadata:01:69e4efdf30b8",
                                                                        "remoteNodeWWN": None, "nvmeNodeName": None},
                                                           "alias": {"ioInterfaceType": "iscsi", "iscsiAlias": ""}, "label": "beegfs_metadata1_iscsi_0",
                                                           "hostRef": "84000000600A098000A4B9D10030370B5D430109",
                                                           "id": "89000000600A098000A4B28D00303CFC5D4300F7"}],
         "hostSidePorts": [{"type": "iscsi", "address": "iqn.1993-08.org.debian.beegfs-metadata:01:69e4efdf30b8", "label": "beegfs_metadata1_iscsi_0"}],
         "id": "84000000600A098000A4B9D10030370B5D430109", "name": "beegfs_metadata1"},
        {"hostRef": "84000000600A098000A4B9D10030370B5D430109", "clusterRef": "85000000600A098000A4B9D1003637135D483DEB", "label": "beegfs_metadata2",
         "hostTypeIndex": 28, "ports": [], "initiators": [{"initiatorRef": "89000000600A098000A4B28D00303CFC5D4300F7",
                                                           "nodeName": {"ioInterfaceType": "iscsi",
                                                                        "iscsiNodeName": "iqn.used_elsewhere",
                                                                        "remoteNodeWWN": None, "nvmeNodeName": None},
                                                           "alias": {"ioInterfaceType": "iscsi", "iscsiAlias": ""}, "label": "beegfs_metadata2_iscsi_0",
                                                           "hostRef": "84000000600A098000A4B9D10030370B5D430109",
                                                           "id": "89000000600A098000A4B28D00303CFC5D4300F7"}],
         "hostSidePorts": [{"type": "iscsi", "address": "iqn.used_elsewhere", "label": "beegfs_metadata2_iscsi_0"}],
         "id": "84000000600A098000A4B9D10030370B5D430120", "name": "beegfs_metadata2"}]
    HOST_GROUPS = [{"clusterRef": "85000000600A098000A4B9D1003637135D483DEB", "label": "test_group", "isSAControlled": False,
                    "confirmLUNMappingCreation": False, "protectionInformationCapableAccessMethod": True, "isLun0Restricted": False,
                    "id": "85000000600A098000A4B9D1003637135D483DEB", "name": "test_group"}]
    HOST_TYPES = [{"name": "FactoryDefault", "index": 0, "code": "FactoryDefault"},
                  {"name": "Windows 2000/Server 2003/Server 2008 Non-Clustered", "index": 1, "code": "W2KNETNCL"},
                  {"name": "Solaris", "index": 2, "code": "SOL"},
                  {"name": "Linux", "index": 6, "code": "LNX"},
                  {"name": "LnxALUA", "index": 7, "code": "LnxALUA"},
                  {"name": "Windows 2000/Server 2003/Server 2008 Clustered", "index": 8, "code": "W2KNETCL"},
                  {"name": "LnxTPGSALUA_SF", "index": 27, "code": "LnxTPGSALUA_SF"},
                  {"name": "LnxDHALUA", "index": 28, "code": "LnxDHALUA"}]
    REQ_FUNC = 'ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_host.NetAppESeriesHost.request'

    def _set_args(self, args):
        module_args = self.REQUIRED_PARAMS.copy()
        module_args.update(args)
        set_module_args(module_args)

    def test_host_exists_pass(self):
        """Verify host_exists produces expected results."""
        with mock.patch(self.REQ_FUNC, return_value=(200, self.EXISTING_HOSTS)):
            self._set_args({'state': 'present', 'name': 'new_host', 'host_type': 'linux dm-mp', 'force_port': False,
                            'ports': [{'label': 'new_host_port_1', 'type': 'fc', 'port': '0x08ef08ef08ef08ef'}]})
            host = NetAppESeriesHost()
            self.assertFalse(host.host_exists)

            self._set_args({'state': 'present', 'name': 'does_not_exist', 'host_type': 'linux dm-mp',
                            'ports': [{'label': 'beegfs_storage1_iscsi_0', 'type': 'iscsi',
                                       'port': 'iqn.1993-08.org.debian.beegfs-storage1:01:b0621126818'}]})
            host = NetAppESeriesHost()
            self.assertFalse(host.host_exists)

            self._set_args({'state': 'present', 'name': 'beegfs_storage1', 'host_type': 'linux dm-mp',
                            'ports': [{'label': 'beegfs_storage1_iscsi_0', 'type': 'iscsi', 'port': 'iqn.differentiqn.org'}]})
            host = NetAppESeriesHost()
            self.assertTrue(host.host_exists)

            with mock.patch(self.REQ_FUNC, return_value=(200, self.EXISTING_HOSTS)):
                self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'linux dm-mp', 'force_port': True,
                                'ports': [{'label': 'beegfs_metadata1_iscsi_0', 'type': 'iscsi',
                                           'port': 'iqn.1993-08.org.debian.beegfs-storage1:01:b0621126818'}]})
                host = NetAppESeriesHost()
                self.assertTrue(host.host_exists)

    def test_host_exists_fail(self):
        """Verify host_exists produces expected exceptions."""
        self._set_args({'state': 'present', 'host_type': 'linux dm-mp', 'ports': [{'label': 'abc', 'type': 'iscsi', 'port': 'iqn:0'}]})
        host = NetAppESeriesHost()
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to determine host existence."):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                exists = host.host_exists

    def test_needs_update_pass(self):
        """Verify needs_update produces expected results."""
        # No changes
        with mock.patch(self.REQ_FUNC, return_value=(200, self.EXISTING_HOSTS)):
            self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'linux dm-mp',
                            'ports': [{'label': 'beegfs_metadata1_iscsi_0', 'type': 'iscsi',
                                       'port': 'iqn.1993-08.org.debian.beegfs-metadata:01:69e4efdf30b8'}]})
            host = NetAppESeriesHost()
            exists = host.host_exists
            self.assertFalse(host.needs_update)

        # Change host type
        with mock.patch(self.REQ_FUNC, return_value=(200, self.EXISTING_HOSTS)):
            self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'windows', 'force_port': False,
                            'ports': [{'label': 'beegfs_metadata1_iscsi_1', 'type': 'iscsi', 'port': 'iqn.not_used'}]})
            host = NetAppESeriesHost()
            exists = host.host_exists
            self.assertTrue(host.needs_update)

        # Add port to host
        with mock.patch(self.REQ_FUNC, return_value=(200, self.EXISTING_HOSTS)):
            self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'linux dm-mp', 'force_port': False,
                            'ports': [{'label': 'beegfs_metadata1_iscsi_1', 'type': 'iscsi', 'port': 'iqn.not_used'}]})
            host = NetAppESeriesHost()
            exists = host.host_exists
            self.assertTrue(host.needs_update)

        # Change port name
        with mock.patch(self.REQ_FUNC, return_value=(200, self.EXISTING_HOSTS)):
            self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'linux dm-mp', 'force_port': False,
                            'ports': [{'label': 'beegfs_metadata1_iscsi_2', 'type': 'iscsi',
                                       'port': 'iqn.1993-08.org.debian.beegfs-metadata:01:69e4efdf30b8'}]})
            host = NetAppESeriesHost()
            exists = host.host_exists
            self.assertTrue(host.needs_update)

        # take port from another host by force
        with mock.patch(self.REQ_FUNC, return_value=(200, self.EXISTING_HOSTS)):
            self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'linux dm-mp', 'force_port': True,
                            'ports': [{'label': 'beegfs_metadata2_iscsi_0', 'type': 'iscsi',
                                       'port': 'iqn.1993-08.org.debian.beegfs-metadata:01:69e4efdf30b8'}]})
            host = NetAppESeriesHost()
            exists = host.host_exists
            self.assertTrue(host.needs_update)

    def test_needs_update_fail(self):
        """Verify needs_update produces expected exceptions."""
        with self.assertRaisesRegexp(AnsibleFailJson, "is associated with a different host."):
            with mock.patch(self.REQ_FUNC, return_value=(200, self.EXISTING_HOSTS)):
                self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'linux dm-mp', 'force_port': False,
                                'ports': [{'label': 'beegfs_metadata2_iscsi_0', 'type': 'iscsi',
                                           'port': 'iqn.1993-08.org.debian.beegfs-metadata:01:69e4efdf30b8'}]})
                host = NetAppESeriesHost()
                exists = host.host_exists
                host.needs_update

    def test_valid_host_type_pass(self):
        """Validate the available host types."""
        with mock.patch(self.REQ_FUNC, return_value=(200, self.HOST_TYPES)):
            self._set_args({'state': 'present', 'host_type': '0'})
            host = NetAppESeriesHost()
            self.assertTrue(host.valid_host_type)
            self._set_args({'state': 'present', 'host_type': '28'})
            host = NetAppESeriesHost()
            self.assertTrue(host.valid_host_type)
            self._set_args({'state': 'present', 'host_type': 'windows'})
            host = NetAppESeriesHost()
            self.assertTrue(host.valid_host_type)
            self._set_args({'state': 'present', 'host_type': 'linux dm-mp'})
            host = NetAppESeriesHost()
            self.assertTrue(host.valid_host_type)

    def test_valid_host_type_fail(self):
        """Validate the available host types."""
        with self.assertRaisesRegexp(AnsibleFailJson, "host_type must be either a host type name or host type index found integer the documentation"):
            self._set_args({'state': 'present', 'host_type': 'non-host-type'})
            host = NetAppESeriesHost()

        with mock.patch(self.REQ_FUNC, return_value=(200, self.HOST_TYPES)):
            with self.assertRaisesRegexp(AnsibleFailJson, "There is no host type with index"):
                self._set_args({'state': 'present', 'host_type': '4'})
                host = NetAppESeriesHost()
                valid = host.valid_host_type

        with mock.patch(self.REQ_FUNC, return_value=Exception()):
            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to get host types."):
                self._set_args({'state': 'present', 'host_type': '4'})
                host = NetAppESeriesHost()
                valid = host.valid_host_type

    def test_assigned_host_ports_pass(self):
        """Verify assigned_host_ports gives expected results."""

        # Add an unused port to host
        with mock.patch(self.REQ_FUNC, return_value=(200, self.EXISTING_HOSTS)):
            self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'linux dm-mp', 'force_port': False,
                            'ports': [{'label': 'beegfs_metadata1_iscsi_1', 'type': 'iscsi', 'port': 'iqn.not_used'}]})
            host = NetAppESeriesHost()
            exists = host.host_exists
            self.assertTrue(host.needs_update)
            self.assertEquals(host.assigned_host_ports(), {})

        # Change port name (force)
        with mock.patch(self.REQ_FUNC, return_value=(200, self.EXISTING_HOSTS)):
            self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'linux dm-mp', 'force_port': True,
                            'ports': [{'label': 'beegfs_metadata1_iscsi_2', 'type': 'iscsi',
                                       'port': 'iqn.1993-08.org.debian.beegfs-metadata:01:69e4efdf30b8'}]})
            host = NetAppESeriesHost()
            exists = host.host_exists
            self.assertTrue(host.needs_update)
            self.assertEquals(host.assigned_host_ports(), {'84000000600A098000A4B9D10030370B5D430109': ['89000000600A098000A4B28D00303CFC5D4300F7']})

        # Change port type
        with mock.patch(self.REQ_FUNC, return_value=(200, self.EXISTING_HOSTS)):
            self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'linux dm-mp', 'force_port': True,
                            'ports': [{'label': 'beegfs_metadata1_iscsi_1', 'type': 'fc', 'port': '08:ef:7e:24:52:a0'}]})
            host = NetAppESeriesHost()
            exists = host.host_exists
            self.assertTrue(host.needs_update)
            self.assertEquals(host.assigned_host_ports(), {})

        # take port from another host by force
        with mock.patch(self.REQ_FUNC, return_value=(200, self.EXISTING_HOSTS)):
            self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'linux dm-mp', 'force_port': True,
                            'ports': [{'label': 'beegfs_metadata2_iscsi_0', 'type': 'iscsi', 'port': 'iqn.used_elsewhere'}]})
            host = NetAppESeriesHost()
            exists = host.host_exists
            self.assertTrue(host.needs_update)
            self.assertEquals(host.assigned_host_ports(), {'84000000600A098000A4B9D10030370B5D430109': ['89000000600A098000A4B28D00303CFC5D4300F7']})

        # take port from another host by force
        with mock.patch(self.REQ_FUNC, side_effect=[(200, self.EXISTING_HOSTS), (200, {})]):
            self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'linux dm-mp', 'force_port': True,
                            'ports': [{'label': 'beegfs_metadata2_iscsi_0', 'type': 'iscsi', 'port': 'iqn.used_elsewhere'}]})
            host = NetAppESeriesHost()
            exists = host.host_exists
            self.assertTrue(host.needs_update)
            self.assertEquals(host.assigned_host_ports(apply_unassigning=True),
                              {'84000000600A098000A4B9D10030370B5D430109': ['89000000600A098000A4B28D00303CFC5D4300F7']})

    def test_assigned_host_ports_fail(self):
        """Verify assigned_host_ports gives expected exceptions."""
        # take port from another
        with self.assertRaisesRegexp(AnsibleFailJson, "There are no host ports available OR there are not enough unassigned host ports"):
            with mock.patch(self.REQ_FUNC, side_effect=[(200, self.EXISTING_HOSTS)]):
                self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'linux dm-mp', 'force_port': False,
                                'ports': [{'label': 'beegfs_metadata1_iscsi_2', 'type': 'iscsi',
                                           'port': 'iqn.1993-08.org.debian.beegfs-metadata:01:69e4efdf30b8'}]})
                host = NetAppESeriesHost()
                exists = host.host_exists
                self.assertTrue(host.needs_update)
                host.assigned_host_ports(apply_unassigning=True)

        # take port from another host and fail because force == False
        with self.assertRaisesRegexp(AnsibleFailJson, "There are no host ports available OR there are not enough unassigned host ports"):
            with mock.patch(self.REQ_FUNC, side_effect=[(200, self.EXISTING_HOSTS)]):
                self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'linux dm-mp', 'force_port': False,
                                'ports': [{'label': 'beegfs_metadata2_iscsi_0', 'type': 'iscsi', 'port': 'iqn.used_elsewhere'}]})
                host = NetAppESeriesHost()
                exists = host.host_exists
                self.assertTrue(host.needs_update)
                host.assigned_host_ports(apply_unassigning=True)

        # take port from another host and fail because force == False
        with self.assertRaisesRegexp(AnsibleFailJson, "There are no host ports available OR there are not enough unassigned host ports"):
            with mock.patch(self.REQ_FUNC, side_effect=[(200, self.EXISTING_HOSTS)]):
                self._set_args({'state': 'present', 'name': 'beegfs_metadata3', 'host_type': 'linux dm-mp', 'force_port': False,
                                'ports': [{'label': 'beegfs_metadata2_iscsi_0', 'type': 'iscsi', 'port': 'iqn.used_elsewhere'}]})
                host = NetAppESeriesHost()
                exists = host.host_exists
                host.assigned_host_ports(apply_unassigning=True)

        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to unassign host port."):
            with mock.patch(self.REQ_FUNC, side_effect=[(200, self.EXISTING_HOSTS), Exception()]):
                self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'linux dm-mp', 'force_port': True,
                                'ports': [{'label': 'beegfs_metadata2_iscsi_0', 'type': 'iscsi', 'port': 'iqn.used_elsewhere'}]})
                host = NetAppESeriesHost()
                exists = host.host_exists
                self.assertTrue(host.needs_update)
                host.assigned_host_ports(apply_unassigning=True)

    def test_update_host_pass(self):
        """Verify update_host produces expected results."""
        # Change host type
        with self.assertRaises(AnsibleExitJson):
            with mock.patch(self.REQ_FUNC, return_value=(200, self.EXISTING_HOSTS)):
                self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'windows', 'force_port': True,
                                'ports': [{'label': 'beegfs_metadata1_iscsi_1', 'type': 'iscsi',
                                           'port': 'iqn.1993-08.org.debian.beegfs-storage1:01:b0621126818'}]})
                host = NetAppESeriesHost()
                host.build_success_payload = lambda x: {}
                exists = host.host_exists
                self.assertTrue(host.needs_update)
                host.update_host()

        # Change port iqn
        with self.assertRaises(AnsibleExitJson):
            with mock.patch(self.REQ_FUNC, return_value=(200, self.EXISTING_HOSTS)):
                self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'linux dm-mp', 'force_port': False,
                                'ports': [{'label': 'beegfs_metadata1_iscsi_1', 'type': 'iscsi', 'port': 'iqn.not_used'}]})
                host = NetAppESeriesHost()
                host.build_success_payload = lambda x: {}
                exists = host.host_exists
                self.assertTrue(host.needs_update)
                host.update_host()

        # Change port type to fc
        with self.assertRaises(AnsibleExitJson):
            with mock.patch(self.REQ_FUNC, return_value=(200, self.EXISTING_HOSTS)):
                self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'linux dm-mp', 'force_port': False,
                                'ports': [{'label': 'beegfs_metadata1_iscsi_1', 'type': 'fc', 'port': '0x08ef08ef08ef08ef'}]})
                host = NetAppESeriesHost()
                host.build_success_payload = lambda x: {}
                exists = host.host_exists
                self.assertTrue(host.needs_update)
                host.update_host()

        # Change port name
        with self.assertRaises(AnsibleExitJson):
            with mock.patch(self.REQ_FUNC, return_value=(200, self.EXISTING_HOSTS)):
                self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'windows', 'force_port': True,
                                'ports': [{'label': 'beegfs_metadata1_iscsi_12', 'type': 'iscsi',
                                           'port': 'iqn.1993-08.org.debian.beegfs-metadata:01:69e4efdf30b8'}]})
                host = NetAppESeriesHost()
                host.build_success_payload = lambda x: {}
                exists = host.host_exists
                self.assertTrue(host.needs_update)
                host.update_host()

    def test_update_host_fail(self):
        """Verify update_host produces expected exceptions."""
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to update host."):
            with mock.patch(self.REQ_FUNC, side_effect=[(200, self.EXISTING_HOSTS), Exception()]):
                self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'windows', 'force_port': False,
                                'ports': [{'label': 'beegfs_metadata1_iscsi_0', 'type': 'iscsi',
                                           'port': 'iqn.1993-08.org.debian.beegfs-metadata:01:69e4efdf30b8'}]})
                host = NetAppESeriesHost()
                host.build_success_payload = lambda x: {}
                exists = host.host_exists
                self.assertTrue(host.needs_update)
                host.update_host()

    def test_create_host_pass(self):
        """Verify create_host produces expected results."""
        def _assigned_host_ports(apply_unassigning=False):
            return None

        with self.assertRaises(AnsibleExitJson):
            with mock.patch(self.REQ_FUNC, return_value=(200, {'id': '84000000600A098000A4B9D10030370B5D430109'})):
                self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'windows', 'force_port': True,
                                'ports': [{'label': 'beegfs_metadata1_iscsi_1', 'type': 'iscsi',
                                           'port': 'iqn.1993-08.org.debian.beegfs-storage1:01:b0621126818'}]})
                host = NetAppESeriesHost()
                with mock.patch(self.REQ_FUNC, return_value=(200, [])):
                    host.assigned_host_ports = _assigned_host_ports
                    host.build_success_payload = lambda x: {}
                    host.create_host()

    def test_create_host_fail(self):
        """Verify create_host produces expected exceptions."""
        def _assigned_host_ports(apply_unassigning=False):
            return None

        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to create host."):
            with mock.patch(self.REQ_FUNC, side_effect=[(200, []), Exception()]):
                self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'windows', 'force_port': True,
                                'ports': [{'label': 'beegfs_metadata1_iscsi_1', 'type': 'iscsi',
                                           'port': 'iqn.1993-08.org.debian.beegfs-storage1:01:b0621126818'}]})
                host = NetAppESeriesHost()
                host.assigned_host_ports = _assigned_host_ports
                host.build_success_payload = lambda x: {}
                host.create_host()

        with self.assertRaisesRegexp(AnsibleExitJson, "Host already exists."):
            with mock.patch(self.REQ_FUNC, side_effect=[(200, self.EXISTING_HOSTS)]):
                self._set_args({'state': 'present', 'name': 'beegfs_storage1', 'host_type': 'linux dm-mp', 'force_port': True,
                                'ports': [{'label': 'beegfs_storage1_iscsi_0', 'type': 'iscsi',
                                           'port': 'iqn.1993-08.org.debian.beegfs-storage1:01:b0621126818'}]})
                host = NetAppESeriesHost()
                host.assigned_host_ports = _assigned_host_ports
                host.build_success_payload = lambda x: {}
                host.create_host()

    def test_remove_host_pass(self):
        """Verify remove_host produces expected results."""
        with mock.patch(self.REQ_FUNC, return_value=(200, None)):
            self._set_args({'state': 'absent', 'name': 'beegfs_metadata1', 'host_type': 'linux dm-mp', 'force_port': False,
                            'ports': [{'label': 'beegfs_metadata1_iscsi_0', 'type': 'iscsi',
                                       'port': 'iqn.1993-08.org.debian.beegfs-metadata:01:69e4efdf30b8'}]})
            host = NetAppESeriesHost()
            host.host_obj = {"id": "84000000600A098000A4B9D10030370B5D430109"}
            host.remove_host()

    def test_remove_host_fail(self):
        """Verify remove_host produces expected exceptions."""
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to remove host."):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                self._set_args({'state': 'absent', 'name': 'beegfs_metadata1', 'host_type': 'linux dm-mp', 'force_port': False,
                                'ports': [{'label': 'beegfs_metadata1_iscsi_0', 'type': 'iscsi',
                                           'port': 'iqn.1993-08.org.debian.beegfs-metadata:01:69e4efdf30b8'}]})
                host = NetAppESeriesHost()
                host.host_obj = {"id": "84000000600A098000A4B9D10030370B5D430109"}
                host.remove_host()

    def test_build_success_payload(self):
        """Validate success payload."""
        def _assigned_host_ports(apply_unassigning=False):
            return None

        self._set_args({'state': 'present', 'name': 'beegfs_metadata1', 'host_type': 'windows', 'force_port': True,
                        'ports': [{'label': 'beegfs_metadata1_iscsi_1', 'type': 'iscsi', 'port': 'iqn.1993-08.org.debian.beegfs-storage1:01:b0621126818'}]})
        host = NetAppESeriesHost()
        self.assertEquals(host.build_success_payload(), {'api_url': 'http://localhost/', 'ssid': '1'})
