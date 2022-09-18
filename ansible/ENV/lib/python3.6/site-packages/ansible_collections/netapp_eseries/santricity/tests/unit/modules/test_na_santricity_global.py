# (c) 2020, NetApp, Inc
# BSD-3 Clause (see COPYING or https://opensource.org/licenses/BSD-3-Clause)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

from ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_global import NetAppESeriesGlobalSettings
from units.modules.utils import AnsibleExitJson, AnsibleFailJson, ModuleTestCase, set_module_args
from units.compat.mock import patch, mock_open


class GlobalSettingsTest(ModuleTestCase):
    REQUIRED_PARAMS = {
        'api_username': 'rw',
        'api_password': 'password',
        'api_url': 'http://localhost',
        'ssid': '1',
    }
    REQ_FUNC = 'ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_global.NetAppESeriesGlobalSettings.request'

    def _set_args(self, args=None):
        module_args = self.REQUIRED_PARAMS.copy()
        if args is not None:
            module_args.update(args)
        set_module_args(module_args)

    def test_init_pass(self):
        """Verify module instantiates successfully."""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 80, "default_host_type": "linux dm-mp", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()

        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 80, "default_host_type": "linux dm-mp", "automatic_load_balancing": "disabled",
                        "host_connectivity_reporting": "disabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()

        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 80, "default_host_type": "linux dm-mp", "automatic_load_balancing": "disabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()

    def test_init_fail(self):
        """Verify module fails when autoload is enabled but host connectivity reporting is not."""
        self._set_args({"automatic_load_balancing": "enabled", "host_connectivity_reporting": "disabled"})
        with self.assertRaisesRegexp(AnsibleFailJson, r"Option automatic_load_balancing requires host_connectivity_reporting to be enabled."):
            instance = NetAppESeriesGlobalSettings()

    def test_get_current_configuration_pass(self):
        """Ensure get_current_configuration method succeeds."""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 80, "default_host_type": "linux dm-mp", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        with patch(self.REQ_FUNC, side_effect=[(200, {"productCapabilities": [], "featureParameters": {"cacheBlockSizes": []}}), (200, []),
                                               (200, [{"defaultHostTypeIndex": 28, "cache": {"cacheBlkSize": 32768, "demandFlushThreshold": 90}}]),
                                               (200, {"autoLoadBalancingEnabled": True, "hostConnectivityReportingEnabled": True, "name": "array1"})]):
            self.assertEqual(instance.get_current_configuration(), {"autoload_capable": False, "autoload_enabled": True, "cache_block_size_options": [],
                                                                    "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 90},
                                                                    "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                                    "host_type_options": {}, "name": 'array1'})

    def test_get_current_configuration_fail(self):
        """Ensure exceptions are thrown when current configuration requests fail."""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 80, "default_host_type": "linux dm-mp", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to retrieve storage array capabilities."):
            with patch(self.REQ_FUNC, side_effect=[Exception()]):
                instance.get_current_configuration()

        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 80, "default_host_type": "linux dm-mp", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to retrieve storage array host options."):
            with patch(self.REQ_FUNC, side_effect=[(200, {"productCapabilities": [], "featureParameters": {"cacheBlockSizes": []}}), Exception()]):
                instance.get_current_configuration()

        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 80, "default_host_type": "linux dm-mp", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to retrieve cache settings."):
            with patch(self.REQ_FUNC, side_effect=[(200, {"productCapabilities": [], "featureParameters": {"cacheBlockSizes": []}}), (200, []), Exception()]):
                instance.get_current_configuration()

        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 80, "default_host_type": "linux dm-mp", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to determine current configuration."):
            with patch(self.REQ_FUNC, side_effect=[(200, {"productCapabilities": [], "featureParameters": {"cacheBlockSizes": []}}), (200, []),
                                                   (200, [{"defaultHostTypeIndex": 28, "cache": {"cacheBlkSize": 32768, "demandFlushThreshold": 90}}]),
                                                   Exception()]):
                instance.get_current_configuration()

    def test_cache_block_size_pass(self):
        """Verify cache_block_size passes successfully."""
        self._set_args({"cache_flush_threshold": 80, "default_host_type": "linux dm-mp", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": False, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 90},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {}, "name": 'array1'}
        self.assertFalse(instance.change_cache_block_size_required())

        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 80, "default_host_type": "linux dm-mp", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": False, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 90},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {}, "name": 'array1'}
        self.assertFalse(instance.change_cache_block_size_required())

        self._set_args({"cache_block_size": 16384, "cache_flush_threshold": 80, "default_host_type": "linux dm-mp", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": False, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 90},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {}, "name": 'array1'}
        self.assertTrue(instance.change_cache_block_size_required())

    def test_cache_block_size_fail(self):
        """Verify cache_block_size throws expected exceptions."""
        self._set_args({"cache_block_size": 16384, "cache_flush_threshold": 80, "default_host_type": "linux dm-mp", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": False, "autoload_enabled": True, "cache_block_size_options": [32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 90},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {}, "name": 'array1'}
        with self.assertRaisesRegexp(AnsibleFailJson, r"Invalid cache block size."):
            self.assertTrue(instance.change_cache_block_size_required())

    def test_change_cache_flush_threshold_required_pass(self):
        """Verify change_cache_block_size_required passes successfully."""
        self._set_args({"cache_block_size": 32768, "default_host_type": "linux dm-mp", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": False, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {}, "name": 'array1'}
        self.assertFalse(instance.change_cache_flush_threshold_required())

        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 80, "default_host_type": "linux dm-mp", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": False, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {}, "name": 'array1'}
        self.assertFalse(instance.change_cache_flush_threshold_required())

        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "linux dm-mp", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": False, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {}, "name": 'array1'}
        self.assertTrue(instance.change_cache_flush_threshold_required())

    def test_change_cache_flush_threshold_required_fail(self):
        """Verify change_cache_block_size_required throws expected exceptions."""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 100, "default_host_type": "linux dm-mp", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": False, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {}, "name": 'array1'}

        with self.assertRaisesRegexp(AnsibleFailJson, r"Invalid cache flushing threshold, it must be equal to or between 0 and 100."):
            instance.change_cache_flush_threshold_required()

        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 0, "default_host_type": "linux dm-mp", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": False, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {}, "name": 'array1'}

        with self.assertRaisesRegexp(AnsibleFailJson, r"Invalid cache flushing threshold, it must be equal to or between 0 and 100."):
            instance.change_cache_flush_threshold_required()

    def test_change_host_type_required_pass(self):
        """Verify change_host_type_required passes successfully."""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": False, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        self.assertFalse(instance.change_host_type_required())

        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Linux DM-MP", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": False, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        self.assertFalse(instance.change_host_type_required())

        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": False, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        self.assertTrue(instance.change_host_type_required())

    def test_change_host_type_required_fail(self):
        """Verify change_host_type_required throws expected exceptions"""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "NotAHostType", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": False, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        with self.assertRaisesRegexp(AnsibleFailJson, r"Invalid host type index!"):
            self.assertTrue(instance.change_host_type_required())

    def test_change_autoload_enabled_required_pass(self):
        """Verify change_autoload_enabled_required passes successfully."""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": True, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        self.assertFalse(instance.change_autoload_enabled_required())

        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": True, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        self.assertFalse(instance.change_autoload_enabled_required())

        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "disabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": True, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        self.assertTrue(instance.change_autoload_enabled_required())

    def test_change_autoload_enabled_required_fail(self):
        """Verify change_autoload_enabled_required throws expected exceptions"""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "NotAHostType", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": False, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        with self.assertRaisesRegexp(AnsibleFailJson, r"Automatic load balancing is not available."):
            self.assertTrue(instance.change_autoload_enabled_required())

    def test_change_host_connectivity_reporting_enabled_required_pass(self):
        """Verify change_host_connectivity_reporting_enabled_required passes successfully."""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "disabled",
                        "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": True, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        self.assertFalse(instance.change_host_connectivity_reporting_enabled_required())

        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "disabled",
                        "host_connectivity_reporting": "enabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": True, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        self.assertFalse(instance.change_host_connectivity_reporting_enabled_required())

        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "disabled",
                        "host_connectivity_reporting": "disabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": True, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        self.assertTrue(instance.change_host_connectivity_reporting_enabled_required())

    def test_change_name_required_pass(self):
        """Verify change_name_required passes successfully."""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "disabled",
                        "host_connectivity_reporting": "disabled"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": True, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        self.assertFalse(instance.change_name_required())

        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "disabled",
                        "host_connectivity_reporting": "disabled", "name": "array1"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": True, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        self.assertFalse(instance.change_name_required())

        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "disabled",
                        "host_connectivity_reporting": "disabled", "name": "array2"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": True, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        self.assertTrue(instance.change_name_required())

    def test_change_name_required_fail(self):
        """Verify change_name_required throws expected exceptions"""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "NotAHostType", "automatic_load_balancing": "enabled",
                        "host_connectivity_reporting": "enabled", "name": "A" * 31})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": False, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        with self.assertRaisesRegexp(AnsibleFailJson, r"The provided name is invalid, it must be less than or equal to 30 characters in length."):
            self.assertTrue(instance.change_name_required())

    def test_update_cache_settings_pass(self):
        """Verify update_cache_settings passes successfully."""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "disabled",
                        "host_connectivity_reporting": "disabled", "name": "array2"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": True, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        with patch(self.REQ_FUNC, return_value=(200, None)):
            instance.update_cache_settings()

    def test_update_cache_settings_fail(self):
        """Verify update_cache_settings throws expected exceptions"""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "disabled",
                        "host_connectivity_reporting": "disabled", "name": "array2"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": True, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to set cache settings."):
            with patch(self.REQ_FUNC, return_value=Exception()):
                instance.update_cache_settings()

    def test_update_host_type_pass(self):
        """Verify update_host_type passes successfully."""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "disabled",
                        "host_connectivity_reporting": "disabled", "name": "array2"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": True, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        with patch(self.REQ_FUNC, return_value=(200, None)):
            instance.update_host_type()

    def test_update_host_type_fail(self):
        """Verify update_host_type throws expected exceptions"""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "disabled",
                        "host_connectivity_reporting": "disabled", "name": "array2"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": True, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to set default host type."):
            with patch(self.REQ_FUNC, return_value=Exception()):
                instance.update_host_type()

    def test_update_autoload_pass(self):
        """Verify update_autoload passes successfully."""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "disabled",
                        "host_connectivity_reporting": "disabled", "name": "array2"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": True, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        with patch(self.REQ_FUNC, return_value=(200, None)):
            instance.update_autoload()

    def test_update_autoload_fail(self):
        """Verify update_autoload throws expected exceptions"""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "disabled",
                        "host_connectivity_reporting": "disabled", "name": "array2"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": True, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to set automatic load balancing state."):
            with patch(self.REQ_FUNC, return_value=Exception()):
                instance.update_autoload()

    def test_update_host_connectivity_reporting_enabled_pass(self):
        """Verify update_host_connectivity_reporting_enabled passes successfully."""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "disabled",
                        "host_connectivity_reporting": "disabled", "name": "array2"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": True, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        with patch(self.REQ_FUNC, return_value=(200, None)):
            instance.update_host_connectivity_reporting_enabled()

    def test_update_host_connectivity_reporting_enabled_fail(self):
        """Verify update_host_connectivity_reporting_enabled throws expected exceptions"""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "disabled",
                        "host_connectivity_reporting": "disabled", "name": "array2"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": True, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to enable host connectivity reporting."):
            with patch(self.REQ_FUNC, return_value=Exception()):
                instance.update_host_connectivity_reporting_enabled()

    def test_update_name_pass(self):
        """Verify update_name passes successfully."""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "disabled",
                        "host_connectivity_reporting": "disabled", "name": "array2"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": True, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        with patch(self.REQ_FUNC, return_value=(200, None)):
            instance.update_name()

    def test_update_name_fail(self):
        """Verify update_name throws expected exceptions"""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "disabled",
                        "host_connectivity_reporting": "disabled", "name": "array2"})
        instance = NetAppESeriesGlobalSettings()
        instance.get_current_configuration = lambda: {"autoload_capable": True, "autoload_enabled": True, "cache_block_size_options": [16384, 32768],
                                                      "cache_settings": {"cache_block_size": 32768, "cache_flush_threshold": 80},
                                                      "default_host_type_index": 28, "host_connectivity_reporting_enabled": True,
                                                      "host_type_options": {"windows": 1, "linux": 28}, "name": 'array1'}
        with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to set the storage array name!"):
            with patch(self.REQ_FUNC, return_value=Exception()):
                instance.update_name()

    def test_update_pass(self):
        """Verify update passes successfully."""
        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "disabled",
                        "host_connectivity_reporting": "disabled", "name": "array2"})
        instance = NetAppESeriesGlobalSettings()

        instance.change_autoload_enabled_required = lambda: False
        instance.change_cache_block_size_required = lambda: False
        instance.change_cache_flush_threshold_required = lambda: False
        instance.change_host_type_required = lambda: False
        instance.change_name_required = lambda: False
        instance.change_host_connectivity_reporting_enabled_required = lambda: False
        with self.assertRaisesRegexp(AnsibleExitJson, r"'changed': False"):
            with patch(self.REQ_FUNC, side_effect=[(200, {"productCapabilities": [], "featureParameters": {"cacheBlockSizes": []}}), (200, []),
                                                   (200, [{"defaultHostTypeIndex": 28, "cache": {"cacheBlkSize": 32768, "demandFlushThreshold": 90}}]),
                                                   (200, {"autoLoadBalancingEnabled": True, "hostConnectivityReportingEnabled": True, "name": "array1"})] * 2):
                instance.update()

        self._set_args({"cache_block_size": 32768, "cache_flush_threshold": 90, "default_host_type": "Windows", "automatic_load_balancing": "disabled",
                        "host_connectivity_reporting": "disabled", "name": "array2"})
        instance = NetAppESeriesGlobalSettings()
        instance.change_autoload_enabled_required = lambda: True
        instance.change_cache_block_size_required = lambda: False
        instance.change_cache_flush_threshold_required = lambda: False
        instance.change_host_type_required = lambda: False
        instance.change_name_required = lambda: False
        instance.change_host_connectivity_reporting_enabled_required = lambda: False
        instance.update_autoload = lambda: None
        with self.assertRaisesRegexp(AnsibleExitJson, r"'changed': True"):
            with patch(self.REQ_FUNC, side_effect=[(200, {"productCapabilities": [], "featureParameters": {"cacheBlockSizes": []}}), (200, []),
                                                   (200, [{"defaultHostTypeIndex": 28, "cache": {"cacheBlkSize": 32768, "demandFlushThreshold": 90}}]),
                                                   (200, {"autoLoadBalancingEnabled": True, "hostConnectivityReportingEnabled": True, "name": "array1"})] * 2):
                instance.update()
