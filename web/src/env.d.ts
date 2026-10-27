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

// The environment variables the portal, its build and its Playwright suites read by name.
declare namespace NodeJS {
  interface ProcessEnv {
    NEXT_PUBLIC_PORTAL_USE_MSW?: string;
    NEXT_PUBLIC_PORTAL_BUILD_SHA?: string;
    NEXT_PUBLIC_PORTAL_NAME?: string;
    NEXT_PUBLIC_PORTAL_LOGO?: string;
    NEXT_PUBLIC_PORTAL_FAVICON?: string;
    NEXTAUTH_SECRET?: string;
    PORT?: string;
    CI?: string;
    LIVE_PORTAL_URL?: string;
    ANALYTICS_SHOT_DIR?: string;
  }
}
