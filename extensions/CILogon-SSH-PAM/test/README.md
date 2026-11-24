<!--
    Licensed to the Apache Software Foundation (ASF) under one
    or more contributor license agreements.  See the NOTICE file
    distributed with this work for additional information
    regarding copyright ownership.  The ASF licenses this file
    to you under the Apache License, Version 2.0 (the
    "License"); you may not use this file except in compliance
    with the License.  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing,
    software distributed under the License is distributed on an
    "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
    KIND, either express or implied.  See the License for the
    specific language governing permissions and limitations
    under the License.
-->

# PAM module tests

The tests are Google Test programs (`test_config`, `unit`, `test_pam_oauth2_device`) built against `../src`.
A mock server stands in for the OIDC provider on port 8042.

1. Install Google Test. The build expects a system install under `/usr`.
2. Start the mock server: `./mock_server.py`.
3. In another terminal, run `make` here. It builds and runs each test.

`make test` in the module directory does the same. `data/` holds sample configurations and expected QR codes.
