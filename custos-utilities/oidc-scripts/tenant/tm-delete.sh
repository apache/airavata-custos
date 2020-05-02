#!/bin/bash
#
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
#
# A script that performs a cURL call to the custos server that has the client management API enabled on it.
# This will issue a POST to the endpoint (as per RFC 7591) and will create a new client on the server
# from the given JSON object.
# This register an admin client in the Custos and the Custos admin will manually accept the client, once accepted
# you MUST use the clientId and clientSecret issued in the response for subsequent tenant creations.

source ./setenv.sh

curl -k -X DELETE  -H "Authorization: Bearer $BEARER_TOKEN"  $REGISTRATION_URI > output_delete.json
cat output_admin.json