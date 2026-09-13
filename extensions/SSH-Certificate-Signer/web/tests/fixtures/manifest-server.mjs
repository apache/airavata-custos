// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

import { createServer } from "node:http";

const manifest = {
  schema_version: 1,
  id: "ssh-certificate-signer",
  name: "SSH Certificate Signer",
  base_path: "/signer",
  web_url: "http://localhost:3217",
  navigation: [
    {
      href: "/signer/certificates",
      label: "SSH Certificates",
      group: "admin",
      icon: "key-round",
      required_privilege: "signer:certificates:read",
    },
  ],
};

createServer((request, response) => {
  if (request.url !== "/.well-known/custos-extension.json") {
    response.writeHead(404).end();
    return;
  }
  response.writeHead(200, { "content-type": "application/json" });
  response.end(JSON.stringify(manifest));
}).listen(3218, "127.0.0.1");
