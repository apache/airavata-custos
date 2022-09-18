# (c) 2020, NetApp, Inc
# BSD-3 Clause (see COPYING or https://opensource.org/licenses/BSD-3-Clause)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

from ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_lun_mapping import NetAppESeriesLunMapping
from units.modules.utils import AnsibleExitJson, AnsibleFailJson, ModuleTestCase, set_module_args
from units.compat import mock


class NetAppLunMappingTest(ModuleTestCase):
    REQUIRED_PARAMS = {"api_username": "rw",
                       "api_password": "password",
                       "api_url": "http://localhost",
                       "ssid": "1"}

    REQ_FUNC = "ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_lun_mapping.NetAppESeriesLunMapping.request"
    GRAPH_RESPONSE = {"storagePoolBundle": {"host": [{"name": "host1", "hostRef": "1"},
                                                     {"name": "host2", "hostRef": "2"},
                                                     {"name": "host3", "hostRef": "3"}],
                                            "cluster": [{"name": "hostgroup1", "clusterRef": "10"},
                                                        {"name": "hostgroup2", "clusterRef": "20"},
                                                        {"name": "hostgroup3", "clusterRef": "30"}],
                                            "lunMapping": [{"volumeRef": "100", "mapRef": "1", "lunMappingRef": "100001", "lun": 5},
                                                           {"volumeRef": "200", "mapRef": "2", "lunMappingRef": "200001", "lun": 3},
                                                           {"volumeRef": "1000", "mapRef": "10", "lunMappingRef": "300001", "lun": 6},
                                                           {"volumeRef": "2000", "mapRef": "20", "lunMappingRef": "400001", "lun": 4}]},
                      "volume": [{"name": "volume1", "volumeRef": "100", "listOfMappings": [{"lun": 5}]},
                                 {"name": "volume2", "volumeRef": "200", "listOfMappings": [{"lun": 3}]},
                                 {"name": "volume3", "volumeRef": "300", "listOfMappings": []}],
                      "highLevelVolBundle": {"thinVolume": [{"name": "thin_volume1", "volumeRef": "1000", "listOfMappings": [{"lun": 6}]},
                                                            {"name": "thin_volume2", "volumeRef": "2000", "listOfMappings": [{"lun": 4}]},
                                                            {"name": "thin_volume3", "volumeRef": "3000", "listOfMappings": []}]},
                      "sa": {"accessVolume": {"name": "access_volume", "accessVolumeRef": "10000"}}}
    MAPPING_INFO = {"lun_mapping": [{"volume_reference": "100", "map_reference": "1", "lun_mapping_reference": "100001", "lun": 5},
                                    {"volume_reference": "200", "map_reference": "2", "lun_mapping_reference": "200001", "lun": 3},
                                    {"volume_reference": "1000", "map_reference": "10", "lun_mapping_reference": "300001", "lun": 6},
                                    {"volume_reference": "2000", "map_reference": "20", "lun_mapping_reference": "400001", "lun": 4}],
                    "volume_by_reference": {"100": "volume1", "200": "volume2", "300": "volume3", "1000": "thin_volume1", "2000": "thin_volume2",
                                            "3000": "thin_volume3", "10000": "access_volume"},
                    "volume_by_name": {"volume1": "100", "volume2": "200", "volume3": "300", "thin_volume1": "1000", "thin_volume2": "2000",
                                       "thin_volume3": "3000", "access_volume": "10000"},
                    "lun_by_name": {"volume1": 5, "volume2": 3, "thin_volume1": 6, "thin_volume2": 4},
                    "target_by_reference": {"1": "host1", "2": "host2", "3": "host3", "10": "hostgroup1", "20": "hostgroup2", "30": "hostgroup3",
                                            "0000000000000000000000000000000000000000": "DEFAULT_HOSTGROUP"},
                    "target_by_name": {"host1": "1", "host2": "2", "host3": "3", "hostgroup1": "10", "hostgroup2": "20", "hostgroup3": "30",
                                       "DEFAULT_HOSTGROUP": "0000000000000000000000000000000000000000"},
                    "target_type_by_name": {"host1": "host", "host2": "host", "host3": "host", "hostgroup1": "group", "hostgroup2": "group",
                                            "hostgroup3": "group", "DEFAULT_HOSTGROUP": "group"}}

    def _set_args(self, args=None):
        module_args = self.REQUIRED_PARAMS.copy()
        if args is not None:
            module_args.update(args)
        set_module_args(module_args)

    def test_update_mapping_info_pass(self):
        """Verify update_mapping_info method creates the correct data structure."""
        options = {"target": "host1", "volume": "volume1"}
        self._set_args(options)
        mapping = NetAppESeriesLunMapping()
        with mock.patch(self.REQ_FUNC, return_value=(200, self.GRAPH_RESPONSE)):
            mapping.update_mapping_info()
            print("%s" % mapping.mapping_info)
            self.assertEquals(mapping.mapping_info, self.MAPPING_INFO)

    def test_update_mapping_info_fail(self):
        """Verify update_mapping_info throws the expected exceptions."""
        response = {"storagePoolBundle": {"host": [{"name": "host1", "hostRef": "1"},
                                                   {"name": "host2", "hostRef": "2"},
                                                   {"name": "host3", "hostRef": "3"}],
                                          "cluster": [{"name": "host1", "clusterRef": "10"},
                                                      {"name": "hostgroup2", "clusterRef": "20"},
                                                      {"name": "hostgroup3", "clusterRef": "30"}]}}
        options = {"target": "host1", "volume": "volume1"}
        self._set_args(options)
        mapping = NetAppESeriesLunMapping()
        with mock.patch(self.REQ_FUNC, return_value=(200, response)):
            with self.assertRaisesRegexp(AnsibleFailJson, "Ambiguous target type: target name is used for both host and group targets!"):
                mapping.update_mapping_info()

    def test_get_lun_mapping_pass(self):
        """Verify get_lun_mapping method creates the correct data structure."""
        options = {"target": "host1", "volume": "volume1"}
        self._set_args(options)
        mapping = NetAppESeriesLunMapping()
        mapping.update_mapping_info = lambda: None
        mapping.mapping_info = self.MAPPING_INFO
        self.assertEquals(mapping.get_lun_mapping(), (True, "100001", 5))

        options = {"target": "host1", "volume": "volume1", "lun": 5}
        self._set_args(options)
        mapping = NetAppESeriesLunMapping()
        mapping.update_mapping_info = lambda: None
        mapping.mapping_info = self.MAPPING_INFO
        self.assertEquals(mapping.get_lun_mapping(), (True, "100001", 5))

        options = {"target": "host1", "volume": "volume3", "lun": 10}
        self._set_args(options)
        mapping = NetAppESeriesLunMapping()
        mapping.update_mapping_info = lambda: None
        mapping.mapping_info = self.MAPPING_INFO
        self.assertEquals(mapping.get_lun_mapping(), (False, None, None))

    def test_get_lun_mapping_fail(self):
        """Verify get_lun_mapping throws the expected exceptions."""
        options = {"target": "host1", "volume": "volume3", "lun": 5}
        self._set_args(options)
        mapping = NetAppESeriesLunMapping()
        mapping.update_mapping_info = lambda: None
        mapping.mapping_info = self.MAPPING_INFO
        with self.assertRaisesRegexp(AnsibleFailJson, "Option lun value is already in use for target!"):
            mapping.get_lun_mapping()

        options = {"target": "host10", "volume": "volume3"}
        self._set_args(options)
        mapping = NetAppESeriesLunMapping()
        mapping.update_mapping_info = lambda: None
        mapping.mapping_info = self.MAPPING_INFO
        with self.assertRaisesRegexp(AnsibleFailJson, "Target does not exist."):
            mapping.get_lun_mapping()

        options = {"target": "host1", "volume": "volume10"}
        self._set_args(options)
        mapping = NetAppESeriesLunMapping()
        mapping.update_mapping_info = lambda: None
        mapping.mapping_info = self.MAPPING_INFO
        with self.assertRaisesRegexp(AnsibleFailJson, "Volume does not exist."):
            mapping.get_lun_mapping()

    def test_update_pass(self):
        """Verify update method creates the correct data structure."""
        options = {"target": "host1", "volume": "volume1"}
        self._set_args(options)
        mapping = NetAppESeriesLunMapping()
        mapping.update_mapping_info = lambda: None
        mapping.mapping_info = self.MAPPING_INFO
        with mock.patch(self.REQ_FUNC, return_value=(200, None)):
            with self.assertRaises(AnsibleExitJson):
                mapping.update()

        options = {"target": "host1", "volume": "volume1", "lun": 5}
        self._set_args(options)
        mapping = NetAppESeriesLunMapping()
        mapping.update_mapping_info = lambda: None
        mapping.mapping_info = self.MAPPING_INFO
        with mock.patch(self.REQ_FUNC, return_value=(200, None)):
            with self.assertRaises(AnsibleExitJson):
                mapping.update()

        options = {"target": "host1", "volume": "volume3", "lun": 10}
        self._set_args(options)
        mapping = NetAppESeriesLunMapping()
        mapping.update_mapping_info = lambda: None
        mapping.mapping_info = self.MAPPING_INFO
        with mock.patch(self.REQ_FUNC, return_value=(200, None)):
            with self.assertRaises(AnsibleExitJson):
                mapping.update()

        options = {"target": "host1", "volume": "volume1", "lun": 10}
        self._set_args(options)
        mapping = NetAppESeriesLunMapping()
        mapping.update_mapping_info = lambda: None
        mapping.mapping_info = self.MAPPING_INFO
        with mock.patch(self.REQ_FUNC, return_value=Exception()):
            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to update storage array lun mapping."):
                mapping.update()

    def test_update_fail(self):
        """Verify update throws the expected exceptions."""
        options = {"target": "host3", "volume": "volume3"}
        self._set_args(options)
        mapping = NetAppESeriesLunMapping()
        mapping.update_mapping_info = lambda: None
        mapping.mapping_info = self.MAPPING_INFO
        with mock.patch(self.REQ_FUNC, return_value=Exception()):
            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to update storage array lun mapping."):
                mapping.update()

        options = {"state": "absent", "target": "host1", "volume": "volume1"}
        self._set_args(options)
        mapping = NetAppESeriesLunMapping()
        mapping.update_mapping_info = lambda: None
        mapping.mapping_info = self.MAPPING_INFO
        with mock.patch(self.REQ_FUNC, return_value=Exception()):
            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to update storage array lun mapping."):
                mapping.update()

        options = {"target": "host3", "volume": "volume3", "lun": 15}
        self._set_args(options)
        mapping = NetAppESeriesLunMapping()
        mapping.update_mapping_info = lambda: None
        mapping.mapping_info = self.MAPPING_INFO
        with mock.patch(self.REQ_FUNC, return_value=Exception()):
            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to update storage array lun mapping."):
                mapping.update()
