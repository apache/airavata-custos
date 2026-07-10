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

# Packaging pam_oauth2_device

## Building a deb package

(Tested on Ubuntu 18.04)

1. Update package metadata in the `debian` directory. Specifically, update the
   `changelog` file. Update pam_oauth2_device version in `deb/build.sh` and
   `deb/Dockerfile`.
2. Follow the commands in `deb/build.sh` script to build the package.
   Alternatively, build the package in a docker container `deb/build.sh`
   (signing is currently not supported).

```bash
docker build -t pamoauth2device-deb-build .
docker run --rm -v ${PWD}:/data pamoauth2device-deb-build bash -c 'cp *.deb /data'
```

## Building a rpm package

1. Update pam_oauth2_device version in `rpm/pamoauth2device.spec` and
`rpm/Dockerfile` files. Update change log in `rpm/pamoauth2device.spec`.
2. In the `rpm` directory, build the container and extract the rpm file.

```bash
docker build -t pamoauth2device-rpm-build .
docker run --rm -v ${PWD}:/data pamoauth2device-rpm-build cp -r 'rpmbuild/RPMS /data'
```
