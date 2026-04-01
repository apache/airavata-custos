#!/usr/bin/env python3
#
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements. See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership. The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License. You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied. See the License for the
# specific language governing permissions and limitations
# under the License.
#
# Mock AMIE server for testing the ACCESS CI service with both
# success and failure scenarios. Simulates the AMIE API endpoints
# that the service polls.
#
# Usage:
#   python3 mock-amie-server.py
#
# Then point the service at: access.amie.base-url=http://localhost:8180

import json
import random
import time
import uuid
from flask import Flask, jsonify, request
from pathlib import Path

app = Flask(__name__)

pending_packets = []
replied_packets = {}
packet_counter = 900000  # starting ID
stats = {"created": 0, "fetched": 0, "replied": 0}

SCENARIOS_DIR = Path(__file__).parent / "scenarios"


def next_id():
    global packet_counter
    packet_counter += 1
    return packet_counter


def make_packet(packet_type, body, transaction_id=None):
    rec_id = next_id()
    return {
        "header": {
            "packet_rec_id": rec_id,
            "transaction_id": transaction_id or f"TXN-{uuid.uuid4().hex[:8]}",
            "packet_type": packet_type,
            "type_id": 1,
            "local_site_name": "MockSite",
            "remote_site_name": "TGCDB",
            "packet_state": "in_progress",
            "packet_timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        },
        "type": packet_type,
        "body": body,
    }


# Scenario generators

def gen_valid_project_create():
    gid = str(random.randint(100000, 999999))
    grant = f"TST{random.randint(100000, 999999)}"
    return make_packet("request_project_create", {
        "GrantNumber": grant,
        "PfosNumber": f"PFOS-{grant}",
        "ProjectTitle": f"Mock project {grant}",
        "PiGlobalID": gid,
        "PiFirstName": random.choice(["Alice", "Bob", "Carol", "Dave", "Eve"]),
        "PiLastName": random.choice(["Smith", "Johnson", "Williams", "Brown", "Jones"]),
        "PiEmail": f"user{gid}@example.edu",
        "PiOrganization": "Mock University",
        "PiOrgCode": "MOCK",
        "NsfStatusCode": "AC",
        "PiDnList": [f"/C=US/O=Mock University/CN=User {gid}"],
        "ServiceUnitsAllocated": str(random.randint(1000, 50000)),
        "StartDate": "2026-01-01",
        "EndDate": "2026-12-31",
        "ResourceList": "mock-cluster.example.edu",
    })


def gen_valid_account_create():
    gid = str(random.randint(100000, 999999))
    return make_packet("request_account_create", {
        "ProjectID": f"PRJ-MOCK{random.randint(1000, 9999)}",
        "GrantNumber": f"TST{random.randint(100000, 999999)}",
        "UserGlobalID": gid,
        "UserFirstName": random.choice(["Frank", "Grace", "Heidi", "Ivan", "Judy"]),
        "UserLastName": random.choice(["Davis", "Garcia", "Rodriguez", "Wilson", "Martinez"]),
        "UserEmail": f"user{gid}@example.edu",
        "UserOrganization": "Mock Institute",
        "UserOrgCode": "MI",
        "NsfStatusCode": "GR",
        "UserDnList": [f"/C=US/O=Mock Institute/CN=User {gid}"],
        "ResourceList": "mock-cluster.example.edu",
    })


def gen_valid_user_modify():
    gid = str(random.randint(100000, 999999))
    return make_packet("request_user_modify", {
        "PersonID": f"person-mock-{uuid.uuid4().hex[:8]}",
        "ActionType": "replace",
        "UserFirstName": "Updated",
        "UserLastName": "Name",
        "UserEmail": f"updated{gid}@example.edu",
        "UserOrganization": "Updated University",
    })


def gen_inform_transaction_complete():
    return make_packet("inform_transaction_complete", {
        "StatusCode": "Success",
        "StatusMessage": "OK",
        "DetailCode": "1",
    })


# Failure scenarios

def gen_missing_global_id():
    """request_account_create with no UserGlobalID — will cause IllegalArgumentException."""
    return make_packet("request_account_create", {
        "ProjectID": f"PRJ-FAIL{random.randint(1000, 9999)}",
        "GrantNumber": f"FAIL{random.randint(100000, 999999)}",
        # UserGlobalID MISSING — required field
        "UserFirstName": "NoGlobalID",
        "UserLastName": "User",
        "UserEmail": "noglobal@example.edu",
        "ResourceList": "mock-cluster.example.edu",
    })


def gen_missing_grant_number():
    """request_project_create with no GrantNumber — will cause IllegalArgumentException."""
    return make_packet("request_project_create", {
        # GrantNumber MISSING — required field
        "PiGlobalID": str(random.randint(100000, 999999)),
        "PiFirstName": "NoGrant",
        "PiLastName": "User",
        "PiEmail": "nogrant@example.edu",
        "NsfStatusCode": "AC",
        "ResourceList": "mock-cluster.example.edu",
    })


def gen_missing_email():
    """request_project_create with no PiEmail — known optional field."""
    gid = str(random.randint(100000, 999999))
    return make_packet("request_project_create", {
        "GrantNumber": f"NOEMAIL{random.randint(100000, 999999)}",
        "PiGlobalID": gid,
        "PiFirstName": "NoEmail",
        "PiLastName": "Person",
        # PiEmail MISSING — optional per protocol, but we need it
        "PiOrganization": "No Email University",
        "NsfStatusCode": "AC",
        "PiDnList": [],
        "ResourceList": "mock-cluster.example.edu",
    })


def gen_missing_pi_name():
    """request_project_create with no PiFirstName — will cause IllegalArgumentException."""
    return make_packet("request_project_create", {
        "GrantNumber": f"NONAME{random.randint(100000, 999999)}",
        "PiGlobalID": str(random.randint(100000, 999999)),
        # PiFirstName MISSING — asserted as required
        "PiLastName": "OnlyLast",
        "PiEmail": "noname@example.edu",
        "NsfStatusCode": "AC",
        "ResourceList": "mock-cluster.example.edu",
    })


def gen_invalid_person_modify():
    """request_user_modify for a person that doesn't exist in the DB."""
    return make_packet("request_user_modify", {
        "PersonID": "nonexistent-person-id-12345",
        "ActionType": "replace",
        "UserFirstName": "Ghost",
        "UserLastName": "User",
        "UserEmail": "ghost@example.edu",
    })


def gen_unknown_packet_type():
    """Packet with an unrecognized type — should hit NoOpHandler."""
    return make_packet("request_something_unknown", {
        "SomeField": "SomeValue",
    })


def gen_empty_body():
    """Packet with an empty body — should cause NPE or validation failure."""
    return make_packet("request_project_create", {})


# Scenario mix

SUCCESS_GENERATORS = [
    (gen_valid_project_create, 3),
    (gen_valid_account_create, 3),
    (gen_valid_user_modify, 1),
    (gen_inform_transaction_complete, 2),
]

FAILURE_GENERATORS = [
    (gen_missing_global_id, 2),
    (gen_missing_grant_number, 1),
    (gen_missing_email, 2),
    (gen_missing_pi_name, 1),
    (gen_invalid_person_modify, 2),
    (gen_unknown_packet_type, 1),
    (gen_empty_body, 1),
]


def generate_batch(success_count=6, failure_count=4):
    """Generate a mixed batch of success and failure packets."""
    packets = []

    success_pool = []
    for gen, weight in SUCCESS_GENERATORS:
        success_pool.extend([gen] * weight)

    failure_pool = []
    for gen, weight in FAILURE_GENERATORS:
        failure_pool.extend([gen] * weight)

    for _ in range(success_count):
        gen = random.choice(success_pool)
        packets.append(gen())

    for _ in range(failure_count):
        gen = random.choice(failure_pool)
        packets.append(gen())

    random.shuffle(packets)
    return packets


# API endpoints

@app.route("/packets/<site>", methods=["GET"])
def get_packets(site):
    """Return pending packets (mimics AMIE API poll).
    Returns a JSON array at the top level, matching the format
    that AmieClient.parsePacketsFromResponse() expects."""
    if not pending_packets:
        return jsonify([])

    batch = list(pending_packets)
    pending_packets.clear()
    stats["fetched"] += len(batch)

    app.logger.info(f"Serving {len(batch)} packets to site {site}")
    return jsonify(batch)


@app.route("/packets/<site>/<int:packet_rec_id>/reply", methods=["POST"])
def reply_to_packet(site, packet_rec_id):
    """Accept a reply from the service."""
    replied_packets[packet_rec_id] = request.get_json(silent=True)
    stats["replied"] += 1
    return jsonify({"message": "Reply accepted"}), 200


@app.route("/test/<site>/reset", methods=["POST"])
def reset(site):
    """Reset all state."""
    pending_packets.clear()
    replied_packets.clear()
    stats["created"] = 0
    stats["fetched"] = 0
    stats["replied"] = 0
    app.logger.info(f"Reset state for site {site}")
    return jsonify({"message": f"Reset site data for {site}", "result": None})


@app.route("/test/<site>/scenarios", methods=["POST"])
def create_scenario(site):
    """Generate a batch of mixed success/failure scenarios."""
    scenario_type = request.args.get("type", "mixed")

    if scenario_type == "mixed":
        packets = generate_batch(success_count=6, failure_count=4)
    elif scenario_type == "failures_only":
        packets = generate_batch(success_count=0, failure_count=8)
    elif scenario_type == "success_only":
        packets = generate_batch(success_count=8, failure_count=0)
    elif scenario_type == "heavy":
        packets = generate_batch(success_count=15, failure_count=10)
    else:
        packets = generate_batch(success_count=3, failure_count=2)

    pending_packets.extend(packets)
    stats["created"] += len(packets)

    app.logger.info(f"Created {len(packets)} packets ({scenario_type}) for site {site}")
    return jsonify({"message": "Test scenario initiated", "result": None})


@app.route("/stats", methods=["GET"])
def get_stats():
    """Stats endpoint for debugging."""
    return jsonify({
        "pending": len(pending_packets),
        "replied": len(replied_packets),
        "stats": stats,
    })


if __name__ == "__main__":
    print("Mock AMIE Server starting on http://localhost:8180")
    print("Point your service at: access.amie.base-url=http://localhost:8180")
    print("")
    print("Scenario types:")
    print("  POST /test/{site}/scenarios?type=mixed          — 6 success + 4 failure")
    print("  POST /test/{site}/scenarios?type=failures_only  — 8 failures")
    print("  POST /test/{site}/scenarios?type=success_only   — 8 successes")
    print("  POST /test/{site}/scenarios?type=heavy          — 15 success + 10 failure")
    print("")
    app.run(host="0.0.0.0", port=8180, debug=False)
