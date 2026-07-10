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

# tls-debug

Small utility using curl to debug connecting to a remote server using SSL/TLS.  It could be https as when
connecting to an IAM server, or it could be ldaps if connecting to a secure LDAP endpoint.  Or indeed any
other protocol for which curl supports TLS.

## Build

The only dependency is libcurl (including the headers (dev) package for building the utility), but of course
curl itself will have further dependencies.

```
gcc -Wall -o tls-debug tls-debug.c -lcurl
```

## Run

```
./tls-debug trustanchors.pem https://iam-host.example.com/
```

Here `trustanchors.pem` can be a file (as in this example) with multiple trust anchor (certification
authority) certificates concatenated together, or it can be a directory as used by OpenSSL.  If it is a
directory,  the connection will work only if curl uses OpenSSL.

## Options

The utility understands no options whatsoever.  Debug is what you get and debug is all you get.

## TODO

The next obvious step is to add client credentials as the third command line parameter.
