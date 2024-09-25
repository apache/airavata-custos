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

from base64 import b64encode


def get_token(custos_settings):
    tokenStr = custos_settings.CUSTOS_CLIENT_ID + ":" + custos_settings.CUSTOS_CLIENT_SEC
    tokenByte = tokenStr.encode('utf-8')
    encodedBytes = b64encode(tokenByte)
    return encodedBytes.decode('utf-8')
