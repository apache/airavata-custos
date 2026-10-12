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

// `next dev` compiles a route, its client chunks and the auth route on first request; specs racing
// that compile under parallel workers miss their first expectation. One signed-in visit to every
// page compiles them before any spec starts.
import { readdirSync } from "node:fs";
import path from "node:path";
import { type FullConfig, chromium } from "@playwright/test";
import { signInAs } from "./fixtures/auth";

const appDir = path.join(__dirname, "..", "src", "app");

const routes = (readdirSync(appDir, { recursive: true }) as string[])
  .filter((file) => path.basename(file) === "page.tsx")
  .map((file) =>
    `/${path.dirname(file)}`
      .split(path.sep)
      .filter((segment) => !/^\(.*\)$/.test(segment))
      .map((segment) => segment.replace(/^\[.*\]$/, "warm-up"))
      .join("/")
      .replace(/^\/\.?$|^$/, "/"),
  );

export default async function warmUp(config: FullConfig) {
  const browser = await chromium.launch();
  const page = await browser.newPage({ baseURL: config.projects[0]?.use.baseURL });
  await signInAs(page, "admin");
  for (const route of routes) await page.goto(route, { waitUntil: "load" });
  await browser.close();
}
