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

const url = "http://localhost:3218/.well-known/custos-extension.json";
for (let attempt = 0; attempt < 60; attempt += 1) {
  try {
    const response = await fetch(url);
    if (response.ok) process.exit(0);
  } catch {
    // The fixture process may still be binding its port.
  }
  await new Promise((resolve) => setTimeout(resolve, 250));
}
throw new Error(`Timed out waiting for ${url}`);
