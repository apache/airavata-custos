# (c) 2020, NetApp, Inc
# BSD-3 Clause (see COPYING or https://opensource.org/licenses/BSD-3-Clause)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

import time
from units.compat import mock
from ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_asup import NetAppESeriesAsup
from units.modules.utils import AnsibleExitJson, AnsibleFailJson, ModuleTestCase, set_module_args


class AsupTest(ModuleTestCase):
    REQUIRED_PARAMS = {
        "api_username": "rw",
        "api_password": "password",
        "api_url": "http://localhost",
        "ssid": "1",
    }

    REQ_FUNC = "ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_asup.NetAppESeriesAsup.request"
    BASE_REQ_FUNC = 'ansible_collections.netapp_eseries.santricity.plugins.module_utils.santricity.request'
    TIME_FUNC = "ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_asup.time.time"

    def _set_args(self, args=None):
        module_args = self.REQUIRED_PARAMS.copy()
        if args is not None:
            module_args.update(args)
        set_module_args(module_args)

    def test_valid_options_pass(self):
        """Validate valid options."""
        options_list = [
            {"state": "disabled", "active": False},
            {"state": "enabled", "active": False, "start": 20, "end": 24, "days": ["saturday", "sunday"],
             "method": "email", "email": {"server": "192.168.1.100", "sender": "noreply@netapp.com"}},
            {"state": "enabled", "active": False, "start": 20, "end": 24, "days": ["saturday", "sunday"],
             "method": "https", "routing_type": "direct"},
            {"state": "enabled", "active": False, "start": 20, "end": 24, "days": ["saturday", "sunday"],
             "method": "https", "routing_type": "proxy", "proxy": {"host": "192.168.1.100", "port": 1234}},
            {"state": "enabled", "active": False, "start": 20, "end": 24, "days": ["saturday", "sunday"],
             "method": "https", "routing_type": "script", "proxy": {"script": "/path/to/proxy/script.sh"}},
            {"state": "maintenance_enabled", "maintenance_duration": 24, "maintenance_emails": ["janey@netapp.com", "joe@netapp.com"]},
            {"state": "maintenance_disabled"}
        ]

        for options in options_list:
            self._set_args(options)
            with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
                asup = NetAppESeriesAsup()
        for options in options_list:
            self._set_args(options)
            with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": True})]):
                asup = NetAppESeriesAsup()

    def test_invalid_options_fail(self):
        """Verify invalid options throw expected exceptions."""
        options_list = [
            {"state": "enabled", "active": False, "start": 24, "end": 23, "days": ["saturday", "sunday"],
             "method": "email", "email": {"server": "192.168.1.100", "sender": "noreply@netapp.com"}},
            {"state": "enabled", "active": False, "start": -1, "end": 23, "days": ["saturday", "sunday"],
             "method": "email", "email": {"server": "192.168.1.100", "sender": "noreply@netapp.com"}},
            {"state": "enabled", "active": False, "start": 20, "end": 25, "days": ["saturday", "sunday"],
             "method": "email", "email": {"server": "192.168.1.100", "sender": "noreply@netapp.com"}},
            {"state": "enabled", "active": False, "start": 20, "end": 24, "days": ["not_a_day", "sunday"],
             "method": "https", "routing_type": "direct"},
            {"state": "maintenance_enabled", "maintenance_duration": 0, "maintenance_emails": ["janey@netapp.com", "joe@netapp.com"]},
            {"state": "maintenance_enabled", "maintenance_duration": 73, "maintenance_emails": ["janey@netapp.com", "joe@netapp.com"]},
        ]

        for options in options_list:
            self._set_args(options)
            with self.assertRaises(AnsibleFailJson):
                with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
                    asup = NetAppESeriesAsup()

    def test_get_configuration_fail(self):
        """Verify get_configuration method throws expected exceptions."""
        self._set_args({"state": "disabled", "active": False})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            asup = NetAppESeriesAsup()
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                with self.assertRaisesRegexp(AnsibleFailJson, "Failed to retrieve ASUP configuration!"):
                    asup.get_configuration()
        self._set_args({"state": "disabled", "active": False})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            asup = NetAppESeriesAsup()
            with mock.patch(self.REQ_FUNC, return_value=(200, {"asupCapable": False, "onDemandCapable": True})):
                with self.assertRaisesRegexp(AnsibleFailJson, "Failed to retrieve ASUP configuration!"):
                    asup.get_configuration()
        self._set_args({"state": "disabled", "active": False})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            asup = NetAppESeriesAsup()
            with mock.patch(self.REQ_FUNC, return_value=(200, {"asupCapable": True, "onDemandCapable": False})):
                with self.assertRaisesRegexp(AnsibleFailJson, "Failed to retrieve ASUP configuration!"):
                    asup.get_configuration()
        self._set_args({"state": "disabled", "active": False})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            asup = NetAppESeriesAsup()
            with mock.patch(self.REQ_FUNC, return_value=(200, {"asupCapable": False, "onDemandCapable": False})):
                with self.assertRaisesRegexp(AnsibleFailJson, "Failed to retrieve ASUP configuration!"):
                    asup.get_configuration()

    def test_in_maintenance_mode_pass(self):
        """Verify whether asup is in maintenance mode successful."""
        self._set_args({"state": "disabled", "active": False})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            asup = NetAppESeriesAsup()
            with mock.patch(self.REQ_FUNC, return_value=(200, [{"key": "ansible_asup_maintenance_stop_time", "value": str(time.time() + 10000)}])):
                self.assertTrue(asup.in_maintenance_mode())

        self._set_args({"state": "disabled", "active": False})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            asup = NetAppESeriesAsup()
            with mock.patch(self.REQ_FUNC, return_value=(200, [{"key": "ansible_asup_maintenance_email_list", "value": "janey@netapp.com,joe@netapp.com"},
                                                               {"key": "ansible_asup_maintenance_stop_time", "value": str(time.time() - 1)}])):
                self.assertFalse(asup.in_maintenance_mode())

    def test_in_maintenance_mode_fail(self):
        """Verify that in_maintenance mode throws expected exceptions."""
        self._set_args({"state": "disabled", "active": False})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            asup = NetAppESeriesAsup()
            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to retrieve maintenance windows information!"):
                with mock.patch(self.REQ_FUNC, return_value=Exception()):
                    asup.in_maintenance_mode()

    def test_update_configuration_pass(self):
        """Verify that update_configuration completes successfully."""
        asup_config = [{"asupCapable": True,
                        "onDemandCapable": True,
                        "asupEnabled": True,
                        "onDemandEnabled": True,
                        "remoteDiagsEnabled": True,
                        "delivery": {"method": "smtp",
                                     "routingType": "none",
                                     "proxyHost": None,
                                     "proxyPort": 0,
                                     "proxyUserName": None,
                                     "proxyPassword": None,
                                     "proxyScript": None,
                                     "mailRelayServer": "server@example.com",
                                     "mailSenderAddress": "noreply@example.com"},
                        "destinationAddress": "autosupport@netapp.com",
                        "schedule": {"dailyMinTime": 0,
                                     "dailyMaxTime": 1439,
                                     "weeklyMinTime": 0,
                                     "weeklyMaxTime": 1439,
                                     "daysOfWeek": ["sunday", "monday", "tuesday"]}},
                       {"asupCapable": True,
                        "onDemandCapable": True,
                        "asupEnabled": True,
                        "onDemandEnabled": False,
                        "remoteDiagsEnabled": False,
                        "delivery": {
                            "method": "https",
                            "routingType": "proxyServer",
                            "proxyHost": "192.168.1.100",
                            "proxyPort": 1234,
                            "proxyUserName": None,
                            "proxyPassword": None,
                            "proxyScript": None,
                            "mailRelayServer": None,
                            "mailSenderAddress": None
                        },
                        "destinationAddress": "https://support.netapp.com/put/AsupPut/",
                        "schedule": {
                            "dailyMinTime": 1200,
                            "dailyMaxTime": 1439,
                            "weeklyMinTime": 0,
                            "weeklyMaxTime": 1439,
                            "daysOfWeek": ["sunday", "saturday"]}},
                       {"asupCapable": True,
                        "onDemandCapable": True,
                        "asupEnabled": True,
                        "onDemandEnabled": False,
                        "remoteDiagsEnabled": False,
                        "delivery": {
                            "method": "https",
                            "routingType": "proxyScript",
                            "proxyHost": None,
                            "proxyPort": 0,
                            "proxyUserName": None,
                            "proxyPassword": None,
                            "proxyScript": "/home/user/path/to/script.sh",
                            "mailRelayServer": None,
                            "mailSenderAddress": None
                        },
                        "destinationAddress": "https://support.netapp.com/put/AsupPut/",
                        "schedule": {
                            "dailyMinTime": 0,
                            "dailyMaxTime": 420,
                            "weeklyMinTime": 0,
                            "weeklyMaxTime": 1439,
                            "daysOfWeek": ["monday", "tuesday", "wednesday", "thursday", "friday"]}}]
        options_list = [{"state": "disabled", "active": False},
                        {"state": "enabled", "active": False, "start": 20, "end": 24, "days": ["saturday"],
                         "method": "email", "email": {"server": "192.168.1.100", "sender": "noreply@netapp.com"}},
                        {"state": "enabled", "active": False, "start": 20, "end": 24, "days": ["sunday"],
                         "method": "https", "routing_type": "direct"},
                        {"state": "enabled", "active": False, "start": 20, "end": 24, "days": ["saturday", "sunday"],
                         "method": "https", "routing_type": "proxy", "proxy": {"host": "192.168.1.100", "port": 1234}},
                        {"state": "enabled", "active": False, "start": 20, "end": 24, "days": ["saturday", "sunday"],
                         "method": "https", "routing_type": "script", "proxy": {"script": "/path/to/proxy/script.sh"}},
                        {"state": "maintenance_enabled", "maintenance_duration": 24, "maintenance_emails": ["janey@netapp.com", "joe@netapp.com"]},
                        {"state": "maintenance_disabled"}]

        for index, options in enumerate(options_list):
            self._set_args(options)
            with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
                asup = NetAppESeriesAsup()
                asup.get_configuration = lambda: asup_config[index % 3]
                asup.in_maintenance_mode = lambda: False

                with mock.patch(self.REQ_FUNC, return_value=(200, None)):
                    asup.update_configuration()

    def test_update_configuration_fail(self):
        """Verify that update_configuration throws expected exceptions."""
        asup_config = {"asupCapable": True,
                       "onDemandCapable": True,
                       "asupEnabled": True,
                       "onDemandEnabled": True,
                       "remoteDiagsEnabled": True,
                       "delivery": {"method": "smtp",
                                    "routingType": "none",
                                    "proxyHost": None,
                                    "proxyPort": 0,
                                    "proxyUserName": None,
                                    "proxyPassword": None,
                                    "proxyScript": None,
                                    "mailRelayServer": "server@example.com",
                                    "mailSenderAddress": "noreply@example.com"},
                       "destinationAddress": "autosupport@netapp.com",
                       "schedule": {"dailyMinTime": 0,
                                    "dailyMaxTime": 1439,
                                    "weeklyMinTime": 0,
                                    "weeklyMaxTime": 1439,
                                    "daysOfWeek": ["sunday", "monday", "tuesday"]}}

        # Exceptions for state=="enabled" or state=="disabled"
        self._set_args({"state": "enabled", "active": False, "start": 20, "end": 24, "days": ["saturday"],
                        "method": "email", "email": {"server": "192.168.1.100", "sender": "noreply@netapp.com"}})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            asup = NetAppESeriesAsup()
            asup.get_configuration = lambda: asup_config
            asup.in_maintenance_mode = lambda: False
            asup.validate = lambda: True
            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to validate ASUP configuration!"):
                with mock.patch(self.REQ_FUNC, return_value=Exception()):
                    asup.update_configuration()
        self._set_args({"state": "disabled", "active": False})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            asup = NetAppESeriesAsup()
            asup.get_configuration = lambda: asup_config
            asup.in_maintenance_mode = lambda: False
            asup.validate = lambda: False
            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to change ASUP configuration!"):
                with mock.patch(self.REQ_FUNC, return_value=Exception()):
                    asup.update_configuration()

        # Exceptions for state=="maintenance enabled"
        self._set_args({"state": "maintenance_enabled", "maintenance_duration": 24, "maintenance_emails": ["janey@netapp.com", "joe@netapp.com"]})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            asup = NetAppESeriesAsup()
            asup.get_configuration = lambda: {"asupEnabled": False}
            asup.in_maintenance_mode = lambda: False
            with self.assertRaisesRegexp(AnsibleFailJson, "AutoSupport must be enabled before enabling or disabling maintenance mode."):
                asup.update_configuration()
        self._set_args({"state": "maintenance_enabled", "maintenance_duration": 24, "maintenance_emails": ["janey@netapp.com", "joe@netapp.com"]})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            asup = NetAppESeriesAsup()
            asup.get_configuration = lambda: {"asupEnabled": True}
            asup.in_maintenance_mode = lambda: False
            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to enabled ASUP maintenance window."):
                with mock.patch(self.REQ_FUNC, return_value=Exception()):
                    asup.update_configuration()
        self._set_args({"state": "maintenance_enabled", "maintenance_duration": 24, "maintenance_emails": ["janey@netapp.com", "joe@netapp.com"]})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            asup = NetAppESeriesAsup()
            asup.get_configuration = lambda: {"asupEnabled": True}
            asup.in_maintenance_mode = lambda: False
            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to store maintenance information."):
                with mock.patch(self.REQ_FUNC, side_effect=[(200, None), Exception()]):
                    asup.update_configuration()
        self._set_args({"state": "maintenance_enabled", "maintenance_duration": 24, "maintenance_emails": ["janey@netapp.com", "joe@netapp.com"]})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            asup = NetAppESeriesAsup()
            asup.get_configuration = lambda: {"asupEnabled": True}
            asup.in_maintenance_mode = lambda: False
            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to store maintenance information."):
                with mock.patch(self.REQ_FUNC, side_effect=[(200, None), (200, None), Exception()]):
                    asup.update_configuration()

        # Exceptions for state=="maintenance disabled"
        self._set_args({"state": "maintenance_disabled"})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            asup = NetAppESeriesAsup()
            asup.get_configuration = lambda: {"asupEnabled": True}
            asup.in_maintenance_mode = lambda: True
            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to disable ASUP maintenance window."):
                with mock.patch(self.REQ_FUNC, return_value=Exception()):
                    asup.update_configuration()
        self._set_args({"state": "maintenance_disabled"})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            asup = NetAppESeriesAsup()
            asup.get_configuration = lambda: {"asupEnabled": True}
            asup.in_maintenance_mode = lambda: True
            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to store maintenance information."):
                with mock.patch(self.REQ_FUNC, side_effect=[(200, None), Exception()]):
                    asup.update_configuration()
        self._set_args({"state": "maintenance_disabled"})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            asup = NetAppESeriesAsup()
            asup.get_configuration = lambda: {"asupEnabled": True}
            asup.in_maintenance_mode = lambda: True
            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to store maintenance information."):
                with mock.patch(self.REQ_FUNC, side_effect=[(200, None), (200, None), Exception()]):
                    asup.update_configuration()
