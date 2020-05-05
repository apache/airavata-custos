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
# This file contains the environment variables for the service. Set them here and they should
# get picked up by each script as needed (this assumes everything is being run from the current
# directory).

# Tip: If you are working with several different clients, you may want to comment out the
# setting REGISTRATION_URI so it does not get set to what is here.


export SERVER=https://custos.scigap.org:32036/tenant-management/v1.0.0/oauth2/tenant
export ACTIVATION_ENDPOINT=https://custos.scigap.org:32036/tenant-management/v1.0.0/status
export ADMIN_ID=custos/In634rvBEjIWtUHUsjF9/10000307
export ADMIN_SECRET=GgDv0GiXWsMZeqqQINMiTUtzLX6zpdIMpmkWpsoc
export REGISTRATION_URI=https://custos.scigap.org:32036/tenant-management/v1.0.0/oauth2/tenant?client_id=custos/h9RWrJZTxFoewlVtZf0x/10000308

# We set the bearer token here so it is available subsequently. This is the least problematic way to
# do this since it is easy to get the escaping wrong.

#export BEARER_TOKEN=$(echo -n $ADMIN_ID:$ADMIN_SECRET | base64 -w 0)
export BEARER_TOKEN=Y3VzdG9zL0luNjM0cnZCRWpJV3RVSFVzakY5LzEwMDAwMzA3OkdnRHYwR2lYV3NNWmVxcVFJTk1pVFV0ekxYNnpwZElNcG1rV3Bzb2M=