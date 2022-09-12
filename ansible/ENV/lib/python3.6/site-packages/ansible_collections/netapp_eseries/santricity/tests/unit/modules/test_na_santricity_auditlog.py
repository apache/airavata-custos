# (c) 2020, NetApp, Inc
# BSD-3 Clause (see COPYING or https://opensource.org/licenses/BSD-3-Clause)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

from ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_auditlog import NetAppESeriesAuditLog
from units.modules.utils import AnsibleFailJson, ModuleTestCase, set_module_args
from units.compat import mock


class NetAppESeriesAuditLogTests(ModuleTestCase):
    REQUIRED_PARAMS = {'api_username': 'rw',
                       'api_password': 'password',
                       'api_url': 'http://localhost',
                       'ssid': '1'}
    REQ_FUNC = 'ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_auditlog.NetAppESeriesAuditLog.request'
    BASE_REQ_FUNC = 'ansible_collections.netapp_eseries.santricity.plugins.module_utils.santricity.request'
    MAX_RECORDS_MAXIMUM = 50000
    MAX_RECORDS_MINIMUM = 100

    def _set_args(self, **kwargs):
        module_args = self.REQUIRED_PARAMS.copy()
        if kwargs is not None:
            module_args.update(kwargs)
        set_module_args(module_args)

    def test_max_records_argument_pass(self):
        """Verify NetAppESeriesAuditLog argument's max_records and threshold upper and lower boundaries."""
        initial = {"max_records": 1000,
                   "log_level": "writeOnly",
                   "full_policy": "overWrite",
                   "threshold": 90}
        max_records_set = (self.MAX_RECORDS_MINIMUM, 25000, self.MAX_RECORDS_MAXIMUM)

        for max_records in max_records_set:
            initial["max_records"] = max_records
            self._set_args(**initial)
            with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
                audit_log = NetAppESeriesAuditLog()
                self.assertTrue(audit_log.max_records == max_records)

    def test_max_records_argument_fail(self):
        """Verify NetAppESeriesAuditLog arument's max_records and threshold upper and lower boundaries."""
        initial = {"max_records": 1000,
                   "log_level": "writeOnly",
                   "full_policy": "overWrite",
                   "threshold": 90}
        max_records_set = (self.MAX_RECORDS_MINIMUM - 1, self.MAX_RECORDS_MAXIMUM + 1)

        for max_records in max_records_set:
            with self.assertRaisesRegexp(AnsibleFailJson, r"Audit-log max_records count must be between 100 and 50000"):
                initial["max_records"] = max_records
                self._set_args(**initial)
                NetAppESeriesAuditLog()

    def test_threshold_argument_pass(self):
        """Verify NetAppESeriesAuditLog argument's max_records and threshold upper and lower boundaries."""
        initial = {"max_records": 1000,
                   "log_level": "writeOnly",
                   "full_policy": "overWrite",
                   "threshold": 90}
        threshold_set = (60, 75, 90)

        for threshold in threshold_set:
            initial["threshold"] = threshold
            self._set_args(**initial)
            with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
                audit_log = NetAppESeriesAuditLog()
                self.assertTrue(audit_log.threshold == threshold)

    def test_threshold_argument_fail(self):
        """Verify NetAppESeriesAuditLog arument's max_records and threshold upper and lower boundaries."""
        initial = {"max_records": 1000,
                   "log_level": "writeOnly",
                   "full_policy": "overWrite",
                   "threshold": 90}
        threshold_set = (59, 91)

        for threshold in threshold_set:
            with self.assertRaisesRegexp(AnsibleFailJson, r"Audit-log percent threshold must be between 60 and 90"):
                initial["threshold"] = threshold
                self._set_args(**initial)
                with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
                    NetAppESeriesAuditLog()

    def test_get_configuration_pass(self):
        """Validate get configuration does not throw exception when normal request is returned."""
        initial = {"max_records": 1000,
                   "log_level": "writeOnly",
                   "full_policy": "overWrite",
                   "threshold": 90}
        expected = {"auditLogMaxRecords": 1000,
                    "auditLogLevel": "writeOnly",
                    "auditLogFullPolicy": "overWrite",
                    "auditLogWarningThresholdPct": 90}

        self._set_args(**initial)
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            audit_log = NetAppESeriesAuditLog()

        with mock.patch(self.REQ_FUNC, return_value=(200, expected)):
            body = audit_log.get_configuration()
            self.assertTrue(body == expected)

    def test_get_configuration_fail(self):
        """Verify AnsibleJsonFail exception is thrown."""
        initial = {"max_records": 1000,
                   "log_level": "writeOnly",
                   "full_policy": "overWrite",
                   "threshold": 90}

        self._set_args(**initial)
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            audit_log = NetAppESeriesAuditLog()

        with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to retrieve the audit-log configuration!"):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                audit_log.get_configuration()

    def test_build_configuration_pass(self):
        """Validate configuration changes will force an update."""
        response = {"auditLogMaxRecords": 1000,
                    "auditLogLevel": "writeOnly",
                    "auditLogFullPolicy": "overWrite",
                    "auditLogWarningThresholdPct": 90}
        initial = {"max_records": 1000,
                   "log_level": "writeOnly",
                   "full_policy": "overWrite",
                   "threshold": 90}
        changes = [{"max_records": 50000},
                   {"log_level": "all"},
                   {"full_policy": "preventSystemAccess"},
                   {"threshold": 75}]

        for change in changes:
            initial_with_changes = initial.copy()
            initial_with_changes.update(change)
            self._set_args(**initial_with_changes)
            with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
                audit_log = NetAppESeriesAuditLog()

            with mock.patch(self.REQ_FUNC, return_value=(200, response)):
                update = audit_log.build_configuration()
                self.assertTrue(update)

    def test_delete_log_messages_fail(self):
        """Verify AnsibleJsonFail exception is thrown."""
        initial = {"max_records": 1000,
                   "log_level": "writeOnly",
                   "full_policy": "overWrite",
                   "threshold": 90}

        self._set_args(**initial)
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            audit_log = NetAppESeriesAuditLog()

        with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to delete audit-log messages!"):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                audit_log.delete_log_messages()

    def test_update_configuration_delete_pass(self):
        """Verify 422 and force successfully returns True."""
        body = {"auditLogMaxRecords": 1000,
                "auditLogLevel": "writeOnly",
                "auditLogFullPolicy": "overWrite",
                "auditLogWarningThresholdPct": 90}
        initial = {"max_records": 2000,
                   "log_level": "writeOnly",
                   "full_policy": "overWrite",
                   "threshold": 90,
                   "force": True}

        self._set_args(**initial)
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            audit_log = NetAppESeriesAuditLog()
            with mock.patch(self.REQ_FUNC, side_effect=[(200, body),
                                                        (422, {u"invalidFieldsIfKnown": None,
                                                               u"errorMessage": u"Configuration change...",
                                                               u"localizedMessage": u"Configuration change...",
                                                               u"retcode": u"auditLogImmediateFullCondition",
                                                               u"codeType": u"devicemgrerror"}),
                                                        (200, None),
                                                        (200, None)]):
                self.assertTrue(audit_log.update_configuration())

    def test_update_configuration_delete_skip_fail(self):
        """Verify 422 and no force results in AnsibleJsonFail exception."""
        body = {"auditLogMaxRecords": 1000,
                "auditLogLevel": "writeOnly",
                "auditLogFullPolicy": "overWrite",
                "auditLogWarningThresholdPct": 90}
        initial = {"max_records": 2000,
                   "log_level": "writeOnly",
                   "full_policy": "overWrite",
                   "threshold": 90,
                   "force": False}

        self._set_args(**initial)
        with mock.patch(self.BASE_REQ_FUNC, side_effect=[(200, {"version": "04.00.00.00"}), (200, {"runningAsProxy": False})]):
            audit_log = NetAppESeriesAuditLog()

        with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to update audit-log configuration!"):
            with mock.patch(self.REQ_FUNC, side_effect=[(200, body), Exception(422, {"errorMessage": "error"}),
                                                        (200, None), (200, None)]):
                audit_log.update_configuration()
