# (c) 2020, NetApp, Inc
# BSD-3 Clause (see COPYING or https://opensource.org/licenses/BSD-3-Clause)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

from units.modules.utils import AnsibleExitJson, AnsibleFailJson, ModuleTestCase, set_module_args
from ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_proxy_drive_firmware_upload import NetAppESeriesProxyDriveFirmwareUpload
from units.compat.mock import patch, mock_open


class StoragePoolTest(ModuleTestCase):
    REQUIRED_PARAMS = {"api_username": "username",
                       "api_password": "password",
                       "api_url": "http://localhost/devmgr/v2",
                       "validate_certs": "no"}

    REQUEST_FUNC = "ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_proxy_drive_firmware_upload." \
                   "NetAppESeriesProxyDriveFirmwareUpload.request"
    CREATE_MULTIPART_FORMDATA_FUNC = "ansible_collections.netapp_eseries.santricity.plugins.modules." \
                                     "na_santricity_proxy_drive_firmware_upload.create_multipart_formdata"
    OS_PATH_EXISTS_FUNC = "os.path.exists"
    OS_PATH_ISDIR_FUNC = "os.path.isdir"
    OS_LISTDIR_FUNC = "os.listdir"

    def _set_args(self, args=None):
        module_args = self.REQUIRED_PARAMS.copy()
        if args is not None:
            module_args.update(args)
        set_module_args(module_args)

    def test_determine_file_paths_pass(self):
        """Ensure determine_file_paths method succeeds when all files exist."""
        self._set_args({"firmware": ["/path/to/firmware1.dlp", "/path/to/firmware/directory"]})
        firmware = NetAppESeriesProxyDriveFirmwareUpload()

        with patch(self.OS_PATH_EXISTS_FUNC, return_value=True):
            with patch(self.OS_PATH_ISDIR_FUNC, side_effect=[False, True]):
                with patch(self.OS_LISTDIR_FUNC, return_value=["firmware2.dlp", "firmware3.dlp"]):
                    firmware.determine_file_paths()
                    self.assertEqual(firmware.files, {"firmware1.dlp": "/path/to/firmware1.dlp",
                                                      "firmware2.dlp": "/path/to/firmware/directory/firmware2.dlp",
                                                      "firmware3.dlp": "/path/to/firmware/directory/firmware3.dlp"})

    def test_determine_file_paths_fail(self):
        """Ensure determine_file_paths method throws expected exception."""
        self._set_args({"firmware": ["/path/to/firmware1.dlp", "/path/to/firmware/directory"]})
        firmware = NetAppESeriesProxyDriveFirmwareUpload()

        with self.assertRaisesRegexp(AnsibleFailJson, r"Drive firmware file does not exist!"):
            with patch(self.OS_PATH_EXISTS_FUNC, side_effect=[True, False]):
                firmware.determine_file_paths()

    def test_determine_changes_pass(self):
        """Determine whether determine_changes returns expected results."""
        self._set_args({"firmware": ["/path/to/firmware1.dlp", "/path/to/firmware/directory"]})
        firmware = NetAppESeriesProxyDriveFirmwareUpload()
        firmware.files = {"firmware1.dlp": "/path/to/firmware1.dlp",
                          "firmware2.dlp": "/path/to/firmware/directory/firmware2.dlp",
                          "firmware3.dlp": "/path/to/firmware/directory/firmware3.dlp"}

        with patch(self.REQUEST_FUNC, return_value=(200, [{"fileName": "firmware1.dlp"}, {"fileName": "firmware3.dlp"}, {"fileName": "firmware4.dlp"}])):
            firmware.determine_changes()

            self.assertEqual(firmware.add_files, ["firmware2.dlp"])
            self.assertEqual(firmware.remove_files, ["firmware4.dlp"])

    def test_determine_changes_fail(self):
        """Ensure class constructor fails when file does not exist."""
        self._set_args({"firmware": ["/path/to/firmware1.dlp", "/path/to/firmware/directory"]})
        firmware = NetAppESeriesProxyDriveFirmwareUpload()

        with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to retrieve proxy drive firmware file list."):
            with patch(self.REQUEST_FUNC, return_value=Exception()):
                firmware.determine_changes()

    def test_upload_files_pass(self):
        """Ensure upload_files method successfully passes."""
        self._set_args({"firmware": ["/path/to/firmware1.dlp", "/path/to/firmware/directory"]})
        firmware = NetAppESeriesProxyDriveFirmwareUpload()
        firmware.files = {"firmware1.dlp": "/path/to/firmware1.dlp",
                          "firmware2.dlp": "/path/to/firmware/directory/firmware2.dlp",
                          "firmware3.dlp": "/path/to/firmware/directory/firmware3.dlp"}
        firmware.add_files = ["firmware1.dlp", "firmware2.dlp"]

        with patch(self.CREATE_MULTIPART_FORMDATA_FUNC, return_value=(None, None)):
            with patch(self.REQUEST_FUNC, return_value=(200, None)):
                firmware.upload_files()

    def test_delete_files_pass(self):
        """Ensure delete_files completes as expected."""
        self._set_args({"firmware": ["/path/to/firmware1.dlp", "/path/to/firmware/directory"]})
        firmware = NetAppESeriesProxyDriveFirmwareUpload()
        firmware.remove_files = ["firmware1.dlp", "firmware2.dlp"]

        with patch(self.REQUEST_FUNC, return_value=(204, None)):
            firmware.delete_files()

    def test_apply_pass(self):
        """Ensure that the apply method behaves as expected."""
        self._set_args({"firmware": ["/path/to/firmware1.dlp", "/path/to/firmware/directory"]})
        firmware = NetAppESeriesProxyDriveFirmwareUpload()
        firmware.files = {"firmware1.dlp": "/path/to/firmware1.dlp",
                          "firmware2.dlp": "/path/to/firmware/directory/firmware2.dlp",
                          "firmware3.dlp": "/path/to/firmware/directory/firmware3.dlp"}
        firmware.module.check_mode = True
        firmware.is_proxy = lambda: True
        firmware.determine_file_paths = lambda: None
        firmware.determine_changes = lambda: None

        firmware.add_files = ["firmware1.dlp", "firmware2.dlp"]
        firmware.remove_files = ["firmware3.dlp", "firmware4.dlp"]
        with self.assertRaisesRegexp(AnsibleExitJson, r"'changed': True"):
            firmware.apply()

        firmware.add_files = ["firmware1.dlp", "firmware2.dlp"]
        firmware.remove_files = []
        with self.assertRaisesRegexp(AnsibleExitJson, r"'changed': True"):
            firmware.apply()

        firmware.add_files = []
        firmware.remove_files = ["firmware3.dlp", "firmware4.dlp"]
        with self.assertRaisesRegexp(AnsibleExitJson, r"'changed': True"):
            firmware.apply()

        firmware.add_files = []
        firmware.remove_files = []
        with self.assertRaisesRegexp(AnsibleExitJson, r"'changed': False"):
            firmware.apply()

    def test_apply_fail(self):
        """Ensure that the apply method fails when not executing against the proxy."""
        self._set_args({"firmware": ["/path/to/firmware1.dlp", "/path/to/firmware/directory"]})
        firmware = NetAppESeriesProxyDriveFirmwareUpload()
        firmware.is_proxy = lambda: False

        with self.assertRaisesRegexp(AnsibleFailJson, r"Module can only be executed against SANtricity Web Services Proxy."):
            firmware.apply()
