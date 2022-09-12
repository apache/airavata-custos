# (c) 2020, NetApp, Inc
# BSD-3 Clause (see COPYING or https://opensource.org/licenses/BSD-3-Clause)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

from ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_alerts_syslog import NetAppESeriesAlertsSyslog
from units.modules.utils import AnsibleFailJson, AnsibleExitJson, ModuleTestCase, set_module_args
from units.compat import mock


class NetAppESeriesAlertSyslogTest(ModuleTestCase):
    REQUIRED_PARAMS = {
        "api_username": "rw",
        "api_password": "password",
        "api_url": "http://localhost",
    }
    REQ_FUNC = 'ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_alerts_syslog.NetAppESeriesAlertsSyslog.request'
    BASE_REQ_FUNC = 'ansible_collections.netapp_eseries.santricity.plugins.module_utils.santricity.request'

    def _set_args(self, args=None):
        module_args = self.REQUIRED_PARAMS.copy()
        if args is not None:
            module_args.update(args)
        set_module_args(module_args)

    def test_valid_options_pass(self):
        """Validate valid options."""
        options_list = [{"servers": []},
                        {"servers": [{"address": "192.168.1.100"}]},
                        {"servers": [{"address": "192.168.1.100", "port": 1000}]},
                        {"servers": [{"address": "192.168.1.100"}, {"address": "192.168.1.200", "port": 1000}, {"address": "192.168.1.300", "port": 2000}]},
                        {"servers": [{"address": "192.168.1.101"}, {"address": "192.168.1.102"}, {"address": "192.168.1.103"},
                                     {"address": "192.168.1.104"}, {"address": "192.168.1.105"}]}]

        for options in options_list:
            self._set_args(options)
            with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
                syslog = NetAppESeriesAlertsSyslog()
        for options in options_list:
            self._set_args(options)
            with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": True})]):
                syslog = NetAppESeriesAlertsSyslog()

    def test_invalid_options_fail(self):
        """Validate exceptions are thrown when invalid options are provided."""
        options_list = [{"servers": [{"address": "192.168.1.100"}, {"address": "192.168.1.200"}, {"address": "192.168.1.300"},
                                     {"address": "192.168.1.101"}, {"address": "192.168.1.102"}, {"address": "192.168.1.103"}]}]

        for options in options_list:
            self._set_args(options)
            with self.assertRaisesRegexp(AnsibleFailJson, "Maximum number of syslog servers is 5!"):
                with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
                    syslog = NetAppESeriesAlertsSyslog()

    def test_change_required_pass(self):
        """Validate is_change_required properly reports true."""
        options_list = [{"servers": []},
                        {"servers": [{"address": "192.168.1.100"}]},
                        {"servers": [{"address": "192.168.1.100", "port": 1000}]},
                        {"servers": [{"address": "192.168.1.100"}, {"address": "192.168.1.200", "port": 1000}, {"address": "192.168.1.300", "port": 2000}]},
                        {"servers": [{"address": "192.168.1.101"}, {"address": "192.168.1.102"}, {"address": "192.168.1.103"},
                                     {"address": "192.168.1.104"}, {"address": "192.168.1.105"}]}]
        current_config_list = [{"syslogReceivers": [{"serverName": "192.168.1.100", "portNumber": 514}]},
                               {"syslogReceivers": [{"serverName": "192.168.1.100", "portNumber": 1000}]},
                               {"syslogReceivers": [{"serverName": "192.168.1.101", "portNumber": 1000}]},
                               {"syslogReceivers": [{"serverName": "192.168.1.100", "portNumber": 514}]},
                               {"syslogReceivers": [{"serverName": "192.168.1.100", "portNumber": 514}]}]

        for index in range(5):
            self._set_args(options_list[index])
            with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
                syslog = NetAppESeriesAlertsSyslog()
                syslog.get_current_configuration = lambda: current_config_list[index]
                self.assertTrue(syslog.is_change_required())

    def test_get_current_configuration_fail(self):
        """Verify get_current_configuration throws expected exception."""
        self._set_args({"servers": []})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            syslog = NetAppESeriesAlertsSyslog()

            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to retrieve syslog configuration!"):
                with mock.patch(self.REQ_FUNC, return_value=Exception()):
                    syslog.get_current_configuration()

    def test_no_change_required_pass(self):
        """Validate is_change_required properly reports false."""
        options_list = [{"servers": []},
                        {"servers": [{"address": "192.168.1.100"}]},
                        {"servers": [{"address": "192.168.1.101", "port": 1000}, {"address": "192.168.1.100", "port": 514}]}]
        current_config_list = [{"syslogReceivers": []},
                               {"syslogReceivers": [{"serverName": "192.168.1.100", "portNumber": 514}]},
                               {"syslogReceivers": [{"serverName": "192.168.1.100", "portNumber": 514}, {"serverName": "192.168.1.101", "portNumber": 1000}]}]

        for index in range(3):
            self._set_args(options_list[index])
            with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
                syslog = NetAppESeriesAlertsSyslog()
                syslog.get_current_configuration = lambda: current_config_list[index]
                self.assertFalse(syslog.is_change_required())

    def test_request_body_pass(self):
        """Verify request body is properly formatted."""
        options_list = [{"servers": []},
                        {"servers": [{"address": "192.168.1.100"}]},
                        {"servers": [{"address": "192.168.1.101", "port": 1000}, {"address": "192.168.1.100", "port": 514}]}]
        expected_config_list = [{"syslogReceivers": [], "defaultFacility": 3, "defaultTag": "StorageArray"},
                                {"syslogReceivers": [{"serverName": "192.168.1.100", "portNumber": 514}], "defaultFacility": 3, "defaultTag": "StorageArray"},
                                {"syslogReceivers": [{"serverName": "192.168.1.101", "portNumber": 1000}, {"serverName": "192.168.1.100", "portNumber": 514}],
                                 "defaultFacility": 3, "defaultTag": "StorageArray"}]

        for index in range(3):
            self._set_args(options_list[index])
            with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
                syslog = NetAppESeriesAlertsSyslog()
                self.assertEqual(syslog.make_request_body(), expected_config_list[index])

    def test_test_configuration_fail(self):
        """Verify get_current_configuration throws expected exception."""
        self._set_args({"servers": []})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            syslog = NetAppESeriesAlertsSyslog()

            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to send test message!"):
                with mock.patch(self.REQ_FUNC, return_value=Exception()):
                    syslog.test_configuration()

    def test_update_pass(self):
        """Verify update method successfully completes."""
        self._set_args({"test": True, "servers": [{"address": "192.168.1.100"}]})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            syslog = NetAppESeriesAlertsSyslog()
            syslog.is_change_required = lambda: True
            syslog.make_request_body = lambda: {}
            self.test_configuration = lambda: None

            with self.assertRaises(AnsibleExitJson):
                with mock.patch(self.REQ_FUNC, return_value=(200, None)):
                    syslog.update()

    def tests_update_fail(self):
        """Verify update method throws expected exceptions."""
        self._set_args({"servers": []})
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            syslog = NetAppESeriesAlertsSyslog()
            syslog.is_change_required = lambda: True
            syslog.make_request_body = lambda: {}

            with self.assertRaisesRegexp(AnsibleFailJson, "Failed to add syslog server!"):
                with mock.patch(self.REQ_FUNC, return_value=Exception()):
                    syslog.update()
