#  Licensed to the Apache Software Foundation (ASF) under one or more
#  contributor license agreements.  See the NOTICE file distributed with
#  this work for additional information regarding copyright ownership.
#  The ASF licenses this file to You under the Apache License, Version 2.0
#  (the "License"); you may not use this file except in compliance with
#  the License.  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#

import logging

import OpenSSL
import requests
import os
import datetime
from urllib3.exceptions import InsecureRequestWarning
import warnings

requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)
from custos.transport.settings import CustosServerClientSettings
import custos.clients.utils.utilities as utl

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)


class CertificateFetchingRestClient(object):

    def __init__(self, custos_server_setting):
        self.custos_settings = custos_server_setting
        self.target = self.custos_settings.CUSTOS_SERVER_HOST + ":" + str(self.custos_settings.CUSTOS_SERVER_PORT)
        self.url = "https://" + self.target + "/resource-secret-management/v1.0.0/secret"
        self.ownertype = "CUSTOS"
        self.resource_type = "SERVER_CERTIFICATE"
        self.params = {
            'metadata.owner_type': self.ownertype,
            'metadata.resource_type': self.resource_type
        }

        self.rootdir = os.path.abspath(os.curdir)
        encodedStr = utl.get_token(self.custos_settings)
        self.header = {'Authorization': 'Bearer {}'.format(encodedStr)}

    def load_certificate(self):
        if not self.__is_certificate_valid():
            self.__download_certificate()

    def __download_certificate(self):
        r = requests.get(url=self.url, params=self.params, headers=self.header, stream=True, timeout=60, verify=False)
        value = r.json()['value']
        path = self.custos_settings.CUSTOS_CERT_PATH
        f = open(path, "w")
        f.write(value)

        try:
            with warnings.catch_warnings():
                warnings.simplefilter('ignore', InsecureRequestWarning)
                yield
        finally:
            f.close()

    def __is_certificate_valid(self):
        if os.path.isfile(self.custos_settings.CUSTOS_CERT_PATH):
            file = open(self.custos_settings.CUSTOS_CERT_PATH)
            x509 = OpenSSL.crypto.load_certificate(OpenSSL.crypto.FILETYPE_PEM,
                                                   file.read())
            expires = datetime.datetime.strptime(x509.get_notAfter().decode('ascii'), '%Y%m%d%H%M%SZ')
            now = datetime.datetime.now()

            if now > expires:
                return False
            else:
                return True
        else:
            return False
