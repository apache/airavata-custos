# (c) 2020, NetApp, Inc
# BSD-3 Clause (see COPYING or https://opensource.org/licenses/BSD-3-Clause)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

import datetime
import os
from ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_client_certificate import NetAppESeriesClientCertificate
from units.modules.utils import AnsibleFailJson, AnsibleExitJson, ModuleTestCase, set_module_args
from units.compat import mock


class NetAppESeriesClientCertificateTest(ModuleTestCase):

    REQUIRED_PARAMS = {"api_username": "username",
                       "api_password": "password",
                       "api_url": "https://localhost:8443/devmgr/v2",
                       "ssid": "1", "validate_certs": "no"}

    REQUEST_FUNC = "ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_client_certificate.NetAppESeriesClientCertificate.request"
    LOAD_PEM_X509_CERTIFICATE = "ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_client_certificate.x509.load_pem_x509_certificate"
    LOAD_DER_X509_CERTIFICATE = "ansible_collections.netapp_eseries.santricity.plugins.modules.na_santricity_client_certificate.x509.load_der_x509_certificate"
    BASE_REQUEST_FUNC = "ansible_collections.netapp_eseries.santricity.plugins.module_utils.santricity.request"

    CERTIFICATE_PATH = "certificate.crt"
    CERTIFICATE_CONTENT = """Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number: 1 (0x1)
    Signature Algorithm: sha256WithRSAEncryption
        Issuer: C=AU, ST=Florida, L=Palm City, O=Internet Widgits Pty Ltd
        Validity
            Not Before: Apr  1 19:30:07 2019 GMT
            Not After : Mar 29 19:30:07 2029 GMT
        Subject: C=AU, ST=Florida, O=Internet Widgits Pty Ltd, CN=test.example.com
        Subject Public Key Info:
            Public Key Algorithm: rsaEncryption
                Public-Key: (2048 bit)
                Modulus:
                    00:ad:64:b5:4c:40:bb:0f:03:e8:2d:a3:76:af:14:
                    49:b8:06:4a:f9:48:9b:ad:f2:69:55:42:b0:49:de:
                    cd:10:c3:37:71:1a:f8:e1:5e:88:61:b3:c3:0f:7a:
                    3b:3e:eb:47:d3:7b:02:f9:40:6d:11:e9:c6:d0:05:
                    3c:ab:d2:51:97:a3:c9:5d:e4:31:89:85:28:dd:96:
                    75:c7:18:87:0e:a4:26:cb:bc:6d:2f:47:74:89:10:
                    a0:40:5c:39:4e:c2:52:bc:72:25:6c:30:48:dc:50:
                    4e:c7:10:68:7f:96:ef:14:78:05:b3:53:5a:91:2a:
                    8f:b0:5d:75:f0:85:b7:34:6f:78:43:44:a6:3c:4d:
                    87:56:d0:fb:cf:53:de:50:f8:a7:70:89:68:52:83:
                    87:32:70:da:cc:3f:d5:ae:f8:b4:8f:d9:de:40:b7:
                    9a:15:c3:83:4b:62:73:d3:a9:e6:fe:2e:4a:33:7f:
                    13:76:10:d5:d4:04:18:44:9c:b7:a8:17:3f:fe:4b:
                    5d:d4:92:5e:9f:95:64:77:ef:1c:01:09:6a:a3:29:
                    33:08:10:fa:5b:1c:ab:45:16:9d:ee:93:0b:90:d4:
                    ea:cf:0e:13:c8:73:d2:29:00:fa:c1:10:ed:20:66:
                    4f:f5:a5:cf:8d:4e:2a:8e:4a:f2:8e:59:f1:a5:b6:
                    f5:87
                Exponent: 65537 (0x10001)
        X509v3 extensions:
            X509v3 Basic Constraints:
                CA:FALSE
            Netscape Comment:
                OpenSSL Generated Certificate
            X509v3 Subject Key Identifier:
                08:21:10:B9:3E:A5:AF:63:02:88:F3:9D:77:74:FC:BB:AE:A0:BE:6F
            X509v3 Authority Key Identifier:
                keyid:B8:CC:D9:8C:03:C6:06:C3:C4:22:DD:04:64:70:79:0C:93:3F:5C:E8

    Signature Algorithm: sha256WithRSAEncryption
         5b:9f:d8:f5:74:e0:66:56:99:62:d8:6f:c0:15:d9:fc:4f:8b:
         3d:ab:7a:a5:e0:55:49:62:fc:1f:d3:d1:71:4a:55:e9:a2:03:
         7b:57:8f:f2:e4:5b:9c:17:9e:e9:fe:4e:20:a7:48:87:e9:e8:
         80:e9:89:3c:4a:94:a2:68:6d:6d:b0:53:e3:9f:a5:dc:b9:cb:
         21:c3:b0:9f:1b:e1:32:8b:e3:cb:df:ba:32:bb:f4:fd:ef:83:
         9e:64:be:c4:37:4e:c2:90:65:60:3e:19:17:57:7f:59:9c:3d:
         8a:4b:4d:c6:42:ad:c4:98:d3:e1:88:74:3d:67:8b:6e:fd:85:
         1a:d0:ba:52:bc:24:bd:9e:74:82:d6:5f:8f:c7:2d:d8:04:b9:
         fa:bd:e7:ef:5b:cf:d4:28:bf:c0:9a:6b:0c:7b:b7:3a:95:91:
         1c:f3:ad:5b:ce:48:cf:fa:c1:6e:82:f2:df:bd:ba:51:8e:00:
         fb:86:b1:a6:a9:6a:5e:e4:e4:17:a2:35:b5:3c:fa:b1:4f:8d:
         b7:24:53:0f:63:ac:16:f5:91:a0:15:e9:59:cd:59:55:28:a3:
         d9:c0:70:74:30:5b:01:2a:e4:25:44:36:dd:74:f1:4a:3c:c3:
         ad:52:51:c1:c7:79:7a:d7:21:23:a0:b6:55:c4:0d:27:40:10:
         4f:9c:db:04:f8:37:5a:4b:a1:9b:f2:78:b3:63:1a:c5:e3:6a:
         a8:6d:c9:d5:73:41:91:c0:49:2c:72:32:43:73:f2:15:3e:c1:
         31:5d:91:b9:04:c1:78:a8:4e:cf:34:90:ee:05:f9:e5:ee:21:
         4c:1b:ae:55:fd:d8:c9:39:91:4c:5e:61:d9:72:10:a4:24:6a:
         20:c6:ad:44:0c:81:7a:ca:d5:fc:1c:6a:bf:52:9d:87:13:47:
         dd:79:9e:6f:6e:03:be:06:7a:87:c9:5f:2d:f8:9f:c6:44:e6:
         05:c0:cd:28:17:2c:09:28:50:2b:12:39:ff:86:85:71:6b:f0:
         cd:0f:4d:54:89:de:88:ee:fb:e8:e3:ba:45:97:9e:67:d6:ae:
         38:54:86:79:ca:fe:99:b4:20:25:d2:30:aa:3a:62:95:0f:dd:
         42:00:18:88:c7:1f:42:07:1d:dd:9c:42:c4:2f:56:c5:50:b1:
         cd:6d:b9:36:df:9f:5d:f5:77:b3:cd:e4:b8:62:ed:2b:50:d0:
         0b:a2:31:0c:ae:20:8c:b4:0a:83:1f:20:3f:6c:d6:c7:bc:b6:
         84:ae:60:6e:69:2b:cb:01:22:55:a4:e5:3e:62:34:bd:20:f8:
         12:13:6f:25:8d:49:88:74:ba:61:51:bc:bc:8a:c6:fb:02:31:
         ce:5b:85:df:55:d0:55:9b
-----BEGIN CERTIFICATE-----
MIIEqTCCApGgAwIBAgIBATANBgkqhkiG9w0BAQsFADBWMQswCQYDVQQGEwJBVTEQ
MA4GA1UECAwHRmxvcmlkYTESMBAGA1UEBwwJUGFsbSBDaXR5MSEwHwYDVQQKDBhJ
bnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwHhcNMTkwNDAxMTkzMDA3WhcNMjkwMzI5
MTkzMDA3WjBdMQswCQYDVQQGEwJBVTEQMA4GA1UECAwHRmxvcmlkYTEhMB8GA1UE
CgwYSW50ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMRkwFwYDVQQDDBB0ZXN0LmV4YW1w
bGUuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArWS1TEC7DwPo
LaN2rxRJuAZK+UibrfJpVUKwSd7NEMM3cRr44V6IYbPDD3o7PutH03sC+UBtEenG
0AU8q9JRl6PJXeQxiYUo3ZZ1xxiHDqQmy7xtL0d0iRCgQFw5TsJSvHIlbDBI3FBO
xxBof5bvFHgFs1NakSqPsF118IW3NG94Q0SmPE2HVtD7z1PeUPincIloUoOHMnDa
zD/Vrvi0j9neQLeaFcODS2Jz06nm/i5KM38TdhDV1AQYRJy3qBc//ktd1JJen5Vk
d+8cAQlqoykzCBD6WxyrRRad7pMLkNTqzw4TyHPSKQD6wRDtIGZP9aXPjU4qjkry
jlnxpbb1hwIDAQABo3sweTAJBgNVHRMEAjAAMCwGCWCGSAGG+EIBDQQfFh1PcGVu
U1NMIEdlbmVyYXRlZCBDZXJ0aWZpY2F0ZTAdBgNVHQ4EFgQUCCEQuT6lr2MCiPOd
d3T8u66gvm8wHwYDVR0jBBgwFoAUuMzZjAPGBsPEIt0EZHB5DJM/XOgwDQYJKoZI
hvcNAQELBQADggIBAFuf2PV04GZWmWLYb8AV2fxPiz2reqXgVUli/B/T0XFKVemi
A3tXj/LkW5wXnun+TiCnSIfp6IDpiTxKlKJobW2wU+Ofpdy5yyHDsJ8b4TKL48vf
ujK79P3vg55kvsQ3TsKQZWA+GRdXf1mcPYpLTcZCrcSY0+GIdD1ni279hRrQulK8
JL2edILWX4/HLdgEufq95+9bz9Qov8Caawx7tzqVkRzzrVvOSM/6wW6C8t+9ulGO
APuGsaapal7k5BeiNbU8+rFPjbckUw9jrBb1kaAV6VnNWVUoo9nAcHQwWwEq5CVE
Nt108Uo8w61SUcHHeXrXISOgtlXEDSdAEE+c2wT4N1pLoZvyeLNjGsXjaqhtydVz
QZHASSxyMkNz8hU+wTFdkbkEwXioTs80kO4F+eXuIUwbrlX92Mk5kUxeYdlyEKQk
aiDGrUQMgXrK1fwcar9SnYcTR915nm9uA74GeofJXy34n8ZE5gXAzSgXLAkoUCsS
Of+GhXFr8M0PTVSJ3oju++jjukWXnmfWrjhUhnnK/pm0ICXSMKo6YpUP3UIAGIjH
H0IHHd2cQsQvVsVQsc1tuTbfn131d7PN5Lhi7StQ0AuiMQyuIIy0CoMfID9s1se8
toSuYG5pK8sBIlWk5T5iNL0g+BITbyWNSYh0umFRvLyKxvsCMc5bhd9V0FWb
-----END CERTIFICATE-----"""
    #
    # {'expire_date': datetime.datetime(2029, 3, 29, 19, 30, 7),
    #  'issuer_dn': [u'AU', u'Florida', u'Palm City', u'Internet Widgits Pty Ltd'],
    #  'start_date': datetime.datetime(2019, 4, 1, 19, 30, 7),
    #  'subject_dn': [u'AU', u'Florida', u'Internet Widgits Pty Ltd', u'test.example.com']})
    #
    CERTIFICATE_FINGERPRINT = b"4cb68a8039a54b2f5fbe4c55dabb92464a0149a9fce64eb779fd3211c482e44e"
    GET_CERTIFICATE_RESPONSE_OLD = [
        {"alias": "f869e886-4262-42de-87a6-8f99fc3e6272",
         "subjectDN": "CN=test.example.com, O=Internet Widgits Pty Ltd, ST=Florida, C=AU",
         "issuerDN": "O=Internet Widgits Pty Ltd, L=Palm City, ST=Florida, C=AU",
         "start": "2019-04-01T19:30:07.000+0000", "expire": "2029-03-29T19:30:07.000+0000", "isUserInstalled": True},
        {"alias": "ca2", "subjectDN": "sdn2", "issuerDN": "idn2",
         "start": "2019-04-02T13:07:30.516Z", "expire": "2019-04-02T13:07:30.516Z", "isUserInstalled": False},
        {"alias": "ca3", "subjectDN": "sdn3", "issuerDN": "idn3",
         "start": "2019-04-02T13:07:30.516Z", "expire": "2019-04-02T13:07:30.516Z", "isUserInstalled": False},
        {"alias": "ca4", "subjectDN": "sdn4", "issuerDN": "idn4",
         "start": "2019-04-02T13:07:30.516Z", "expire": "2019-04-02T13:07:30.516Z", "isUserInstalled": False}]
    GET_CERTIFICATE_RESPONSE = [
        {'alias': 'alias1', 'expire': '2019-04-02T13:46:04.285Z', 'isKeyEntry': True, 'isUserInstalled': True,
         'issuerDN': 'string', 'issuerRdns': [{'attributes': [{'name': 'string', 'value': 'string'}]}],
         'sha256Fingerprint': b'4cb68a8039a54b2f5fbe4c55dabb92464a0149a9fce64eb779fd3211c482e44e',
         'shaFingerprint': b'4cb68a8039a54b2f5fbe4c55dabb92464a0149a9fce64eb779fd3211c482e44e',
         'start': '2019-04-02T13:46:04.285Z', 'status': 'trusted', 'subjectDN': 'string',
         'subjectRdns': [{'attributes': [{'name': 'string', 'value': 'string'}]}], 'truststore': True, 'type': 'selfSigned'},
        {"alias": "alias1", "shaFingerprint": CERTIFICATE_FINGERPRINT, "sha256Fingerprint": CERTIFICATE_FINGERPRINT,
         "subjectDN": "string", "subjectRdns": [{"attributes": [{"name": "string", "value": "string"}]}],
         "issuerDN": "string", "issuerRdns": [{"attributes": [{"name": "string", "value": "string"}]}],
         "start": "2019-04-02T13:46:04.285Z", "expire": "2019-04-02T13:46:04.285Z", "status": "trusted",
         "truststore": True, "isUserInstalled": True, "isKeyEntry": True, "type": "selfSigned"},
        {"alias": "alias1", "shaFingerprint": "123412341234", "sha256Fingerprint": "4567345673456",
         "subjectDN": "string", "subjectRdns": [{"attributes": [{"name": "string", "value": "string"}]}],
         "issuerDN": "string", "issuerRdns": [{"attributes": [{"name": "string", "value": "string"}]}],
         "start": "2019-04-02T13:46:04.285Z", "expire": "2019-04-02T13:46:04.285Z", "status": "trusted",
         "truststore": True, "isUserInstalled": True, "isKeyEntry": True, "type": "selfSigned"}
    ]

    def _set_args(self, args=None):
        module_args = self.REQUIRED_PARAMS.copy()
        if args is not None:
            module_args.update(args)
        set_module_args(module_args)

        if not os.path.exists(self.CERTIFICATE_PATH):
            with open(self.CERTIFICATE_PATH, "w") as fh:
                fh.write(self.CERTIFICATE_CONTENT)

    def test_init_url_path_prefix(self):
        """Verify url path prefix for both embedded and proxy scenarios."""
        self._set_args({"certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": False})]):
            certificate = NetAppESeriesClientCertificate()
            self.assertEquals(certificate.url_path_prefix, "")

        self._set_args({"certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": True})]):
            certificate = NetAppESeriesClientCertificate()
            self.assertEquals(certificate.url_path_prefix, "storage-systems/1/forward/devmgr/v2/")

        self._set_args({"ssid": "0", "certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": True})]):
            certificate = NetAppESeriesClientCertificate()
            self.assertEquals(certificate.url_path_prefix, "")

        self._set_args({"ssid": "PROXY", "certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": True})]):
            certificate = NetAppESeriesClientCertificate()
            self.assertEquals(certificate.url_path_prefix, "")

    def test_certificate_info_pass(self):
        """Determine whether certificate_info returns expected results."""
        self._set_args({"certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": False})]):
            certificate = NetAppESeriesClientCertificate()
            self.assertEquals(certificate.certificate_info(self.CERTIFICATE_PATH),
                              {"start_date": datetime.datetime(2019, 4, 1, 19, 30, 7),
                               "expire_date": datetime.datetime(2029, 3, 29, 19, 30, 7),
                               "subject_dn": ["AU", "Florida", "Internet Widgits Pty Ltd", "test.example.com"],
                               "issuer_dn": ["AU", "Florida", "Palm City", "Internet Widgits Pty Ltd"]})

    def test_certificate_info_fail(self):
        """Determine wehther certificate_info throws expected exceptions."""
        self._set_args({"certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": False})]):
            certificate = NetAppESeriesClientCertificate()
            with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to load certificate."):
                with mock.patch(self.LOAD_PEM_X509_CERTIFICATE, side_effect=Exception()):
                    with mock.patch(self.LOAD_DER_X509_CERTIFICATE, side_effect=Exception()):
                        certificate.certificate_info(self.CERTIFICATE_PATH)

        self._set_args({"certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": False})]):
            certificate = NetAppESeriesClientCertificate()
            with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to open certificate file or invalid certificate object type."):
                with mock.patch(self.LOAD_PEM_X509_CERTIFICATE, return_value=None):
                    certificate.certificate_info(self.CERTIFICATE_PATH)

    def test_certificate_fingerprint_pass(self):
        """Determine whether certificate_fingerprint returns expected results."""
        self._set_args({"certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": False})]):
            certificate = NetAppESeriesClientCertificate()
            self.assertEquals(certificate.certificate_fingerprint(self.CERTIFICATE_PATH), "4cb68a8039a54b2f5fbe4c55dabb92464a0149a9fce64eb779fd3211c482e44e")

    def test_certificate_fingerprint_fail(self):
        """Determine whether certificate_fingerprint throws expected exceptions."""
        self._set_args({"certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": False})]):
            certificate = NetAppESeriesClientCertificate()
            with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to determine certificate fingerprint."):
                with mock.patch(self.LOAD_PEM_X509_CERTIFICATE, side_effect=Exception()):
                    with mock.patch(self.LOAD_DER_X509_CERTIFICATE, side_effect=Exception()):
                        certificate.certificate_fingerprint(self.CERTIFICATE_PATH)

    def test_determine_changes_pass(self):
        """Determine whether determine_changes successful return expected results."""
        self._set_args({"certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": False})]):
            certificate = NetAppESeriesClientCertificate()
            with mock.patch(self.REQUEST_FUNC, return_value=(200, self.GET_CERTIFICATE_RESPONSE)):
                certificate.determine_changes()
                self.assertEquals(certificate.add_certificates, ["certificate.crt"])
                # self.assertEquals(certificate.remove_certificates, [])

        self._set_args({"certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": False})]):
            certificate = NetAppESeriesClientCertificate()
            with mock.patch(self.REQUEST_FUNC, side_effect=[(404, None), (200, self.GET_CERTIFICATE_RESPONSE_OLD)]):
                certificate.determine_changes()
                self.assertEquals(certificate.add_certificates, [])
                # self.assertEquals(certificate.remove_certificates, [])

        self._set_args({"certificates": []})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": False})]):
            certificate = NetAppESeriesClientCertificate()
            with mock.patch(self.REQUEST_FUNC, side_effect=[(404, None), (200, self.GET_CERTIFICATE_RESPONSE_OLD)]):
                certificate.determine_changes()
                self.assertEquals(certificate.add_certificates, [])
                self.assertEquals(certificate.remove_certificates, [self.GET_CERTIFICATE_RESPONSE_OLD[0]])

    def test_determine_changes_fail(self):
        """Determine whether determine_changes throws expected exceptions."""
        self._set_args({"certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": False})]):
            certificate = NetAppESeriesClientCertificate()
            with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to retrieve remote server certificates."):
                with mock.patch(self.REQUEST_FUNC, return_value=(300, [])):
                    certificate.determine_changes()

        self._set_args({"certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": False})]):
            certificate = NetAppESeriesClientCertificate()
            with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to retrieve remote server certificates."):
                with mock.patch(self.REQUEST_FUNC, side_effect=[(404, None), (300, [])]):
                    certificate.determine_changes()

    def test_upload_certificate_pass(self):
        """Validate upload_certificate successfully completes"""
        self._set_args({"certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": False})]):
            certificate = NetAppESeriesClientCertificate()
            with mock.patch(self.REQUEST_FUNC, return_value=(200, [])):
                certificate.upload_certificate(self.CERTIFICATE_PATH)

        self._set_args({"certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": False})]):
            certificate = NetAppESeriesClientCertificate()
            with mock.patch(self.REQUEST_FUNC, side_effect=[(404, None), (200, [])]):
                certificate.upload_certificate(self.CERTIFICATE_PATH)

    def test_upload_certificate_fail(self):
        """Validate upload_certificate successfully completes"""
        self._set_args({"certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": False})]):
            certificate = NetAppESeriesClientCertificate()
            with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to upload certificate."):
                with mock.patch(self.REQUEST_FUNC, return_value=(300, [])):
                    certificate.upload_certificate(self.CERTIFICATE_PATH)

        self._set_args({"certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": False})]):
            certificate = NetAppESeriesClientCertificate()
            with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to upload certificate."):
                with mock.patch(self.REQUEST_FUNC, side_effect=[(404, None), (300, [])]):
                    certificate.upload_certificate(self.CERTIFICATE_PATH)

    def test_delete_certificate_pass(self):
        """Validate delete_certificate successfully completes"""
        self._set_args({"certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": False})]):
            certificate = NetAppESeriesClientCertificate()
            with mock.patch(self.REQUEST_FUNC, return_value=(200, [])):
                certificate.delete_certificate({"alias": "alias1"})

        self._set_args({"certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": False})]):
            certificate = NetAppESeriesClientCertificate()
            with mock.patch(self.REQUEST_FUNC, side_effect=[(404, None), (200, [])]):
                certificate.delete_certificate({"alias": "alias1"})

    def test_delete_certificate_fail(self):
        """Validate delete_certificate successfully completes"""
        self._set_args({"certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": False})]):
            certificate = NetAppESeriesClientCertificate()
            with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to delete certificate."):
                with mock.patch(self.REQUEST_FUNC, return_value=(300, [])):
                    certificate.delete_certificate({"alias": "alias1"})

        self._set_args({"certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": False})]):
            certificate = NetAppESeriesClientCertificate()
            with self.assertRaisesRegexp(AnsibleFailJson, r"Failed to delete certificate."):
                with mock.patch(self.REQUEST_FUNC, side_effect=[(404, None), (300, [])]):
                    certificate.delete_certificate({"alias": "alias1"})

    def test_apply_pass(self):
        """Verify apply functions as expected."""
        self._set_args({"certificates": [self.CERTIFICATE_PATH]})
        with mock.patch(self.BASE_REQUEST_FUNC, side_effect=[(200, {"version": "03.00.0000.0000"}), (200, {"runningAsProxy": False})]):
            certificate = NetAppESeriesClientCertificate()
            certificate.determine_changes = lambda: None
            certificate.delete_certificate = lambda x: None
            certificate.upload_certificate = lambda x: None

            certificate.remove_certificates = []
            certificate.add_certificates = []
            certificate.module.check_mode = False
            with self.assertRaises(AnsibleExitJson):
                certificate.apply()

            certificate.remove_certificates = []
            certificate.add_certificates = []
            certificate.module.check_mode = True
            with self.assertRaises(AnsibleExitJson):
                certificate.apply()

            certificate.remove_certificates = [True]
            certificate.add_certificates = []
            certificate.module.check_mode = False
            with self.assertRaises(AnsibleExitJson):
                certificate.apply()

            certificate.remove_certificates = []
            certificate.add_certificates = [True]
            certificate.module.check_mode = False
            with self.assertRaises(AnsibleExitJson):
                certificate.apply()
