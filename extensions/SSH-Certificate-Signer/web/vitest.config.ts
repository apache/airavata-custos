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

import path from "node:path";
import { fileURLToPath } from "node:url";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

const rootDir = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./vitest.setup.ts"],
    include: ["src/**/*.{test,spec}.{ts,tsx}"],
    exclude: ["**/node_modules/**", "**/.next/**", "tests/**"],
  },
  resolve: {
    alias: {
      "@": path.resolve(rootDir, "./src"),
      "@features": path.resolve(rootDir, "./src/features"),
      "@shared": path.resolve(rootDir, "./src/shared"),
      "@lib": path.resolve(rootDir, "./src/lib"),
      "@mocks": path.resolve(rootDir, "./src/mocks"),
      // Build-time barrier against client bundling; stubbed for unit tests.
      "server-only": path.resolve(rootDir, "./vitest.server-only-stub.ts"),
    },
  },
});
