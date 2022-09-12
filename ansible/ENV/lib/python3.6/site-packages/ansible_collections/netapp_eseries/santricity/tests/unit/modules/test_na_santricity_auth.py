# (c) 2020, NetApp, Inc
# BSD-3 Clause (see COPYING or https://opensource.org/licenses/BSD-3-Clause)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

from ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_auth import NetAppESeriesAuth
from units.modules.utils import AnsibleExitJson, AnsibleFailJson, ModuleTestCase, set_module_args
from units.compat import mock


class AuthTest(ModuleTestCase):
    REQUIRED_PARAMS = {"api_username": "admin", "api_password": "password", "api_url": "http://localhost", "ssid": "1"}
    REQ_FUNC = "ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_auth.NetAppESeriesAuth.request"
    SLEEP_FUNC = "ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_auth.sleep"

    def _set_args(self, args=None):
        module_args = self.REQUIRED_PARAMS.copy()
        if args is not None:
            module_args.update(args)
        set_module_args(module_args)

    def test_minimum_password_length_change_required_pass(self):
        """Verify minimum_password_length_change_required returns expected values."""
        self._set_args({"ssid": "Proxy", "user": "admin", "password": "adminpass", "minimum_password_length": 8})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        with mock.patch(self.REQ_FUNC, return_value=(200, {"adminPasswordSet": False, "minimumPasswordLength": 8})):
            self.assertFalse(auth.minimum_password_length_change_required())
        self._set_args({"ssid": "Proxy", "user": "admin", "password": "adminpass", "minimum_password_length": 7})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        with mock.patch(self.REQ_FUNC, return_value=(200, {"adminPasswordSet": False, "minimumPasswordLength": 8})):
            self.assertTrue(auth.minimum_password_length_change_required())

        self._set_args({"ssid": "10", "user": "admin", "password": "adminpass", "minimum_password_length": 8})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        with mock.patch(self.REQ_FUNC, return_value=(200, {"adminPasswordSet": False, "minimumPasswordLength": 8})):
            self.assertFalse(auth.minimum_password_length_change_required())

        self._set_args({"ssid": "10", "user": "admin", "password": "adminpass", "minimum_password_length": 8})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: True
        with mock.patch(self.REQ_FUNC, return_value=(200, {"adminPasswordSet": False, "minimumPasswordLength": 8})):
            self.assertFalse(auth.minimum_password_length_change_required())
        self._set_args({"ssid": "10", "user": "admin", "password": "adminpass", "minimum_password_length": 7})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: True
        with mock.patch(self.REQ_FUNC, return_value=(200, {"adminPasswordSet": False, "minimumPasswordLength": 8})):
            self.assertTrue(auth.minimum_password_length_change_required())

        self._set_args({"ssid": "1", "user": "admin", "password": "adminpass", "minimum_password_length": 8})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: False
        auth.is_embedded_available = lambda: True
        with mock.patch(self.REQ_FUNC, return_value=(200, {"adminPasswordSet": False, "minimumPasswordLength": 8})):
            self.assertFalse(auth.minimum_password_length_change_required())
        self._set_args({"ssid": "1", "user": "admin", "password": "adminpass", "minimum_password_length": 7})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: False
        auth.is_embedded_available = lambda: True
        with mock.patch(self.REQ_FUNC, return_value=(200, {"adminPasswordSet": False, "minimumPasswordLength": 8})):
            self.assertTrue(auth.minimum_password_length_change_required())

    def test_minimum_password_length_change_required_fail(self):
        """Verify minimum_password_length_change_required throws expected exceptions."""
        self._set_args({"ssid": "Proxy", "user": "admin", "password": "adminpass", "minimum_password_length": 10})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: False
        with self.assertRaisesRegexp(AnsibleFailJson, "Password does not meet the length requirement"):
            with mock.patch(self.REQ_FUNC, return_value=(200, {"adminPasswordSet": False, "minimumPasswordLength": 8})):
                auth.minimum_password_length_change_required()

        self._set_args({"ssid": "Proxy", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: False
        with self.assertRaisesRegexp(AnsibleFailJson, "Password does not meet the length requirement"):
            with mock.patch(self.REQ_FUNC, return_value=(200, {"adminPasswordSet": True, "minimumPasswordLength": 10})):
                auth.minimum_password_length_change_required()

    def test_update_minimum_password_length_pass(self):
        """Verify update_minimum_password_length returns expected values."""
        self._set_args({"ssid": "Proxy", "user": "admin", "password": "adminpass", "minimum_password_length": 8})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        auth.is_admin_password_set = True
        with mock.patch(self.REQ_FUNC, return_value=(200, None)):
            auth.update_minimum_password_length()
        self._set_args({"ssid": "Proxy", "user": "admin", "password": "adminpass", "minimum_password_length": 8})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        auth.is_admin_password_set = False
        with mock.patch(self.REQ_FUNC, return_value=(200, None)):
            auth.update_minimum_password_length()
        self._set_args({"ssid": "Proxy", "user": "admin", "password": "adminpass", "minimum_password_length": 8})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        auth.is_admin_password_set = False
        with mock.patch(self.REQ_FUNC, side_effect=[Exception(), (200, None)]):
            auth.update_minimum_password_length()

        self._set_args({"ssid": "10", "user": "admin", "password": "adminpass", "minimum_password_length": 8})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: True
        auth.is_admin_password_set = True
        with mock.patch(self.REQ_FUNC, return_value=(200, None)):
            auth.update_minimum_password_length()
        self._set_args({"ssid": "10", "user": "admin", "password": "adminpass", "minimum_password_length": 8})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: True
        auth.is_admin_password_set = False
        with mock.patch(self.REQ_FUNC, return_value=(200, None)):
            auth.update_minimum_password_length()

        self._set_args({"ssid": "1", "user": "admin", "password": "adminpass", "minimum_password_length": 8})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: False
        auth.is_embedded_available = lambda: True
        auth.is_admin_password_set = True
        with mock.patch(self.REQ_FUNC, return_value=(200, None)):
            auth.update_minimum_password_length()
        self._set_args({"ssid": "1", "user": "admin", "password": "adminpass", "minimum_password_length": 8})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: False
        auth.is_embedded_available = lambda: True
        auth.is_admin_password_set = False
        with mock.patch(self.REQ_FUNC, return_value=(200, None)):
            auth.update_minimum_password_length()

    def test_update_minimum_password_length_fail(self):
        """Verify update_minimum_password_length throws expected exceptions."""
        self._set_args({"ssid": "Proxy", "user": "admin", "password": "adminpass", "minimum_password_length": 8})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        auth.is_admin_password_set = False
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to set minimum password length."):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                auth.update_minimum_password_length()

        self._set_args({"ssid": "10", "user": "admin", "password": "adminpass", "minimum_password_length": 8})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: True
        auth.is_admin_password_set = False
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to set minimum password length."):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                auth.update_minimum_password_length()

        self._set_args({"ssid": "1", "user": "admin", "password": "adminpass", "minimum_password_length": 8})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: False
        auth.is_embedded_available = lambda: True
        auth.is_admin_password_set = False
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to set minimum password length."):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                auth.update_minimum_password_length()

    def test_logout_system_pass(self):
        """Verify logout_system returns expected values."""
        self._set_args({"ssid": "Proxy", "user": "admin", "password": "adminpass", "minimum_password_length": 8})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        with mock.patch(self.REQ_FUNC, return_value=(204, None)):
            auth.logout_system()
        self._set_args({"ssid": "10", "user": "admin", "password": "adminpass", "minimum_password_length": 8})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        with mock.patch(self.REQ_FUNC, return_value=(204, None)):
            auth.logout_system()
        self._set_args({"ssid": "Proxy", "user": "admin", "password": "adminpass", "minimum_password_length": 8})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: True
        with mock.patch(self.REQ_FUNC, return_value=(204, None)):
            auth.logout_system()
        self._set_args({"ssid": "Proxy", "user": "admin", "password": "adminpass", "minimum_password_length": 8})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: False
        auth.is_embedded_available = lambda: True
        with mock.patch(self.REQ_FUNC, return_value=(204, None)):
            auth.logout_system()

    def test_password_change_required_pass(self):
        """Verify password_change_required returns expected values."""
        self._set_args({"ssid": "Proxy", "user": "admin"})
        auth = NetAppESeriesAuth()
        self.assertFalse(auth.password_change_required())

        self._set_args({"ssid": "Proxy", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        with mock.patch(self.REQ_FUNC, side_effect=[(200, {"minimumPasswordLength": 8, "adminPasswordSet": False})]):
            self.assertTrue(auth.password_change_required())
        self._set_args({"ssid": "10", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: True
        with mock.patch(self.REQ_FUNC, side_effect=[(200, {"minimumPasswordLength": 8, "adminPasswordSet": False})]):
            self.assertTrue(auth.password_change_required())
        self._set_args({"ssid": "10", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        with mock.patch(self.REQ_FUNC, side_effect=[(200, {"minimumPasswordLength": 8, "adminPasswordSet": False})]):
            self.assertTrue(auth.password_change_required())
        self._set_args({"ssid": "10", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: False
        auth.is_embedded_available = lambda: True
        with mock.patch(self.REQ_FUNC, side_effect=[(200, {"minimumPasswordLength": 8, "adminPasswordSet": False})]):
            self.assertTrue(auth.password_change_required())

        self._set_args({"ssid": "Proxy", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        auth.logout_system = lambda: None
        with mock.patch(self.REQ_FUNC, side_effect=[(200, {"minimumPasswordLength": 8, "adminPasswordSet": True}), (200, None)]):
            self.assertFalse(auth.password_change_required())
        with mock.patch(self.REQ_FUNC, side_effect=[(200, {"minimumPasswordLength": 8, "adminPasswordSet": True}), (401, None)]):
            self.assertTrue(auth.password_change_required())

        self._set_args({"ssid": "10", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: True
        auth.logout_system = lambda: None
        with mock.patch(self.REQ_FUNC, side_effect=[(200, {"minimumPasswordLength": 8, "adminPasswordSet": True}), (200, None)]):
            self.assertFalse(auth.password_change_required())
        with mock.patch(self.REQ_FUNC, side_effect=[(200, {"minimumPasswordLength": 8, "adminPasswordSet": True}), (401, None)]):
            self.assertTrue(auth.password_change_required())

        self._set_args({"ssid": "10", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        auth.logout_system = lambda: None
        with mock.patch(self.REQ_FUNC, side_effect=[(200, {"minimumPasswordLength": 8, "adminPasswordSet": True}), (200, {"isValidPassword": True})]):
            self.assertFalse(auth.password_change_required())
        with mock.patch(self.REQ_FUNC, side_effect=[(200, {"minimumPasswordLength": 8, "adminPasswordSet": True}), (200, {"isValidPassword": False})]):
            self.assertTrue(auth.password_change_required())

        self._set_args({"ssid": "10", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: False
        auth.is_embedded_available = lambda: True
        auth.logout_system = lambda: None
        with mock.patch(self.REQ_FUNC, side_effect=[(200, {"minimumPasswordLength": 8, "adminPasswordSet": True}), (200, None)]):
            self.assertFalse(auth.password_change_required())
        with mock.patch(self.REQ_FUNC, side_effect=[(200, {"minimumPasswordLength": 8, "adminPasswordSet": True}), (401, None)]):
            self.assertTrue(auth.password_change_required())

    def test_password_change_required_fail(self):
        """Verify password_change_required throws expected exceptions."""
        self._set_args({"ssid": "Proxy", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        auth.logout_system = lambda: None
        with self.assertRaisesRegexp(AnsibleFailJson, "SAML enabled! SAML disables default role based login."):
            with mock.patch(self.REQ_FUNC, side_effect=[(200, {"minimumPasswordLength": 8, "adminPasswordSet": True}), (422, None)]):
                auth.password_change_required()

        self._set_args({"ssid": "10", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        auth.logout_system = lambda: None
        auth.is_web_services_version_met = lambda x: True
        with self.assertRaisesRegexp(AnsibleFailJson, "For platforms before E2800 use SANtricity Web Services Proxy 4.1 or later!"):
            with mock.patch(self.REQ_FUNC, side_effect=[(200, {"minimumPasswordLength": 8, "adminPasswordSet": True}), (404, None)]):
                self.assertFalse(auth.password_change_required())
        auth.is_web_services_version_met = lambda x: False
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to validate stored password!"):
            with mock.patch(self.REQ_FUNC, side_effect=[(200, {"minimumPasswordLength": 8, "adminPasswordSet": True}), (404, None)]):
                self.assertFalse(auth.password_change_required())
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to validate stored password!"):
            with mock.patch(self.REQ_FUNC, side_effect=[(200, {"minimumPasswordLength": 8, "adminPasswordSet": True}), (422, None)]):
                self.assertFalse(auth.password_change_required())

        self._set_args({"ssid": "10", "user": "monitor", "password": "monitorpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        auth.logout_system = lambda: None
        auth.is_web_services_version_met = lambda x: True
        with self.assertRaisesRegexp(AnsibleFailJson, "Role based login not available! Only storage system password can be set for storage systems prior to"):
            with mock.patch(self.REQ_FUNC, side_effect=[(200, {"minimumPasswordLength": 8, "adminPasswordSet": True})]):
                self.assertFalse(auth.password_change_required())

    def test_set_array_admin_password_pass(self):
        """Verify set_array_admin_password results."""
        self._set_args({"ssid": "Proxy", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        with mock.patch(self.REQ_FUNC, side_effect=[(200, None)]):
            auth.set_array_admin_password()
        with mock.patch(self.REQ_FUNC, side_effect=[Exception(), (200, None)]):
            auth.set_array_admin_password()

        self._set_args({"ssid": "10", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        with mock.patch(self.REQ_FUNC, side_effect=[(200, None)]):
            auth.set_array_admin_password()
        auth.is_embedded_available = lambda: True
        with mock.patch(self.REQ_FUNC, side_effect=[(200, None)]):
            auth.set_array_admin_password()

        self._set_args({"ssid": "1", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: False
        auth.is_embedded_available = lambda: True
        with mock.patch(self.REQ_FUNC, return_value=(200, None)):
            auth.set_array_admin_password()

    def test_set_array_admin_password_fail(self):
        """Verify set_array_admin_password throws expected exceptions."""
        self._set_args({"ssid": "Proxy", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to set proxy's admin password."):
            with mock.patch(self.REQ_FUNC, side_effect=[Exception(), Exception()]):
                auth.set_array_admin_password()

        self._set_args({"ssid": "10", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to set storage system's admin password."):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                auth.set_array_admin_password()

        self._set_args({"ssid": "1", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: False
        auth.is_embedded_available = lambda: True
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to set embedded storage system's admin password."):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                auth.set_array_admin_password()

    def test_set_array_password_pass(self):
        """Verify set_array_password results."""
        self._set_args({"ssid": "Proxy", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        auth.is_admin_password_set = True
        with mock.patch(self.REQ_FUNC, return_value=(200, None)):
            auth.set_array_password()

        self._set_args({"ssid": "10", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: True
        auth.is_admin_password_set = True
        with mock.patch(self.REQ_FUNC, return_value=(200, None)):
            auth.set_array_password()

        self._set_args({"ssid": "1", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: False
        auth.is_embedded_available = lambda: True
        auth.is_admin_password_set = True
        with mock.patch(self.REQ_FUNC, return_value=(200, None)):
            auth.set_array_password()

    def test_set_array_password_fail(self):
        """Verify set_array_password throws expected exceptions."""
        self._set_args({"ssid": "Proxy", "user": "monitor", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        auth.is_admin_password_set = False
        with self.assertRaisesRegexp(AnsibleFailJson, "Admin password not set! Set admin password before changing non-admin user passwords."):
            auth.set_array_password()

        self._set_args({"ssid": "Proxy", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: False
        auth.is_admin_password_set = True
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to set proxy password."):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                auth.set_array_password()

        self._set_args({"ssid": "10", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: True
        auth.is_embedded_available = lambda: True
        auth.is_admin_password_set = True
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to set embedded user password."):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                auth.set_array_password()

        self._set_args({"ssid": "1", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_proxy = lambda: False
        auth.is_embedded_available = lambda: True
        auth.is_admin_password_set = True
        with self.assertRaisesRegexp(AnsibleFailJson, "Failed to set embedded user password."):
            with mock.patch(self.REQ_FUNC, return_value=Exception()):
                auth.set_array_password()

    def test_apply_pass(self):
        """Verify apply results."""
        self._set_args({"ssid": "1", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_admin_password_set = True
        auth.password_change_required = lambda: True
        auth.minimum_password_length_change_required = lambda: True
        auth.update_minimum_password_length = lambda: None
        auth.set_array_admin_password = lambda: None
        auth.set_array_password = lambda: None
        with self.assertRaisesRegexp(AnsibleExitJson, "'admin' password and required password length has been changed."):
            auth.apply()

        self._set_args({"ssid": "1", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_admin_password_set = False
        auth.password_change_required = lambda: True
        auth.minimum_password_length_change_required = lambda: True
        auth.update_minimum_password_length = lambda: None
        auth.set_array_admin_password = lambda: None
        auth.set_array_password = lambda: None
        with self.assertRaisesRegexp(AnsibleExitJson, "'admin' password and required password length has been changed."):
            auth.apply()

        self._set_args({"ssid": "1", "user": "monitor", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_admin_password_set = True
        auth.password_change_required = lambda: True
        auth.minimum_password_length_change_required = lambda: True
        auth.update_minimum_password_length = lambda: None
        auth.set_array_admin_password = lambda: None
        auth.set_array_password = lambda: None
        with self.assertRaisesRegexp(AnsibleExitJson, "'monitor' password and required password length has been changed."):
            auth.apply()

        self._set_args({"ssid": "1", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_admin_password_set = True
        auth.password_change_required = lambda: True
        auth.minimum_password_length_change_required = lambda: False
        auth.update_minimum_password_length = lambda: None
        auth.set_array_admin_password = lambda: None
        auth.set_array_password = lambda: None
        with self.assertRaisesRegexp(AnsibleExitJson, "'admin' password has been changed."):
            auth.apply()

        self._set_args({"ssid": "1", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_admin_password_set = True
        auth.password_change_required = lambda: False
        auth.minimum_password_length_change_required = lambda: True
        auth.update_minimum_password_length = lambda: None
        auth.set_array_admin_password = lambda: None
        auth.set_array_password = lambda: None
        with self.assertRaisesRegexp(AnsibleExitJson, "Required password length has been changed."):
            auth.apply()

        self._set_args({"ssid": "1", "user": "admin", "password": "adminpass"})
        auth = NetAppESeriesAuth()
        auth.is_admin_password_set = True
        auth.password_change_required = lambda: False
        auth.minimum_password_length_change_required = lambda: False
        auth.update_minimum_password_length = lambda: None
        auth.set_array_admin_password = lambda: None
        auth.set_array_password = lambda: None
        with self.assertRaisesRegexp(AnsibleExitJson, "No changes have been made."):
            auth.apply()
