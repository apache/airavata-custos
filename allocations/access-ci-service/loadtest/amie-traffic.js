/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import http from 'k6/http';
import {check, sleep} from 'k6';
import {Counter} from 'k6/metrics';
import {randomIntBetween} from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const BASE_URL = __ENV.AMIE_BASE_URL || 'https://a3mdev.xsede.org/amie-api-test';
const SITE = __ENV.AMIE_SITE || 'GaTech';
const API_KEY = __ENV.AMIE_API_KEY;

if (!API_KEY) {
    throw new Error('AMIE_API_KEY environment variable is required');
}

const scenariosCreated = new Counter('amie_scenarios_created');

const HEADERS = {
    'XA-SITE': SITE,
    'XA-API-KEY': API_KEY,
};

// Mock server uses mixed type to get both success and failure packets.
// Standard test scenarios with the real AMIE test server.
const USE_MOCK = (BASE_URL.includes('localhost') || BASE_URL.includes('127.0.0.1'));

// Weighted scenario types for real AMIE (cumulative thresholds)
const REAL_SCENARIO_TYPES = [
    {type: 'request_project_reactivate', weight: 0.30},
    {type: 'request_account_reactivate', weight: 0.60},
    {type: 'request_person_merge', weight: 0.80},
    {type: 'request_user_modify', weight: 1.00},
];

const MOCK_SCENARIO_TYPES = [
    {type: 'mixed', weight: 0.50},
    {type: 'success_only', weight: 0.70},
    {type: 'failures_only', weight: 0.85},
    {type: 'heavy', weight: 1.00},
];

const SCENARIO_TYPES = USE_MOCK ? MOCK_SCENARIO_TYPES : REAL_SCENARIO_TYPES;

function pickScenarioType() {
    const r = Math.random();
    for (const s of SCENARIO_TYPES) {
        if (r <= s.weight) return s.type;
    }
    return SCENARIO_TYPES[SCENARIO_TYPES.length - 1].type;
}

// Traffic stages: warm-up → ramp → peak → cool down → steady → quiet → spike → recovery
export const options = {
    stages: [
        {duration: '1m', target: 1},   // warm-up
        {duration: '2m', target: 5},   // ramp up
        {duration: '5m', target: 5},   // peak
        {duration: '2m', target: 2},   // cool down
        {duration: '5m', target: 2},   // steady
        {duration: '3m', target: 1},   // quiet
        {duration: '30s', target: 8},   // spike
        {duration: '2m', target: 1},   // recovery
    ],
    thresholds: {
        http_req_failed: ['rate<0.05'],
        http_req_duration: ['p(95)<5000'],
    },
};

export function setup() {
    // Reset AMIE test server before the run
    const res = http.post(`${BASE_URL}/test/${SITE}/reset`, null, {headers: HEADERS});
    check(res, {'reset succeeded': (r) => r.status === 200});
    console.log(`AMIE test server reset for site ${SITE}`);
    sleep(2);
}

export default function () {
    const scenarioType = pickScenarioType();
    const url = `${BASE_URL}/test/${SITE}/scenarios?type=${scenarioType}`;

    const res = http.post(url, null, {
        headers: HEADERS,
        tags: {scenario_type: scenarioType},
    });

    check(res, {
        'scenario created (200)': (r) => r.status === 200,
    });

    if (res.status === 200) {
        scenariosCreated.add(1, {type: scenarioType});
    }

    // Randomized pause between scenario creation (5-15s)
    sleep(randomIntBetween(5, 15));
}

export function teardown() {
    console.log('Load test complete. Check Grafana for processing metrics.');
}
