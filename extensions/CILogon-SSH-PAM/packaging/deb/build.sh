#!/bin/bash

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


NAME=pamoauth2device
VERSION=0.1.1
URL_REPO=https://github.com/jsurkont/pam_oauth2_device
BUILD_DIR=${NAME}-${VERSION}

curl -L ${URL_REPO}/archive/v${VERSION}.tar.gz -o ${NAME}_${VERSION}.orig.tar.gz
mkdir ${BUILD_DIR}
tar -xzf ${NAME}_${VERSION}.orig.tar.gz -C ${BUILD_DIR} --strip-components 1
cp -r debian ${BUILD_DIR}
cd ${BUILD_DIR}
debuild --force-sign
