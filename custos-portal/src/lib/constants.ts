/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing,
 *  software distributed under the License is distributed on an
 *  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *  KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

import packageJson from '../../package.json';

export const PORTAL_VERSION = packageJson.version;
export const CLIENT_ID = 'custos-gcq8jxkwpvs2gcudzmfn-10000000';;
export const BACKEND_URL = 'http://localhost:8081';
export const APP_URL = "http://localhost:5173";

export const APP_REDIRECT_URI = `${APP_URL}/callback/`;
export const TENANT_ID = '10000000';