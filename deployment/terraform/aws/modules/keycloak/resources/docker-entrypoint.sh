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

# ECS 1.3
if [ -n "${ECS_CONTAINER_METADATA_URI}" ]; then
  EXTERNAL_ADDR=$(curl -fs "${ECS_CONTAINER_METADATA_URI}" \
    | jq -r '.Networks[0].IPv4Addresses[0]')
fi

# ECS 1.4
if [ -n "${ECS_CONTAINER_METADATA_URI_V4}" ]; then
  EXTERNAL_ADDR=$(curl -fs "${ECS_CONTAINER_METADATA_URI_V4}" \
    | jq -r '.Networks[0].IPv4Addresses[0]')
fi

if [ -z "${EXTERNAL_ADDR}" ]; then
  EXTERNAL_ADDR=127.0.0.1
fi
export EXTERNAL_ADDR


if [ -z "${HOSTNAME}" ]; then
  HOSTNAME="localhost"
fi

exec /opt/keycloak/bin/kc.sh start --optimized "$@"
exit $?
