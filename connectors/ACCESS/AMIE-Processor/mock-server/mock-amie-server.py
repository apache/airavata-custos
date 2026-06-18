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
# Mock AMIE server for testing the ACCESS-AMIE connector. Generates
# packets that match the canonical AMIE protocol field names (see
# connectors/ACCESS/AMIE-Processor/testdata/*/incoming-request.json for the
# reference shapes). Field names are kept identical to what a real AMIE
# upstream emits so the connector can be exercised end-to-end without a
# decoder shim.

import os
import random
import time
import uuid
from flask import Flask, jsonify, request
from pathlib import Path

app = Flask(__name__)

pending_packets = []
replied_packets = {}
packet_counter = 900000
stats = {"created": 0, "fetched": 0, "replied": 0}

# Maps GrantNumber → Custos project UUID, populated as notify_project_create
# replies arrive. Packets carrying a "__GRANT__<GN>" placeholder for ProjectID
# are held back from /packets until the GrantNumber resolves here.
grant_to_project_id = {}

SCENARIOS_DIR = Path(__file__).parent / "scenarios"
DEV_EMAIL = os.getenv("DEV_EMAIL", "").strip()


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


# ----- Success generators (canonical AMIE protocol field names) -----

def gen_valid_project_create():
    """request_project_create — site is being asked to create a new project.
    Note: AMIE protocol does NOT include ProjectID here; the receiving site
    assigns one and returns it in notify_project_create."""
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
        "ResourceList": ["mock-cluster.example.edu"],
    })


def gen_valid_account_create():
    """request_account_create — user requests an account on an existing project."""
    gid = str(random.randint(100000, 999999))
    return make_packet("request_account_create", {
        "ProjectID": f"PRJ-MOCK{random.randint(1000, 9999)}",
        "GrantNumber": f"TST{random.randint(100000, 999999)}",
        "UserPersonID": f"person-{uuid.uuid4().hex[:8]}",
        "UserGlobalID": gid,
        "UserFirstName": random.choice(["Frank", "Grace", "Heidi", "Ivan", "Judy"]),
        "UserLastName": random.choice(["Davis", "Garcia", "Rodriguez", "Wilson", "Martinez"]),
        "UserEmail": f"user{gid}@example.edu",
        "UserOrganization": "Mock Institute",
        "UserOrgCode": "MI",
        "NsfStatusCode": "GR",
        "UserDnList": [f"/C=US/O=Mock Institute/CN=User {gid}"],
        "ResourceList": ["mock-cluster.example.edu"],
    })


def gen_valid_user_modify():
    """request_user_modify — replace user attributes for a known UserGlobalID."""
    gid = str(random.randint(100000, 999999))
    return make_packet("request_user_modify", {
        "ActionType": "replace",
        "ProjectID": f"PRJ-MOCK{random.randint(1000, 9999)}",
        "PersonID": f"person-{uuid.uuid4().hex[:8]}",
        "UserPersonID": f"user-person-{uuid.uuid4().hex[:8]}",
        "UserGlobalID": gid,
        "UserFirstName": "Updated",
        "UserLastName": "Name",
        "UserEmail": f"updated{gid}@example.edu",
        "UserOrganization": "Updated University",
        "UserOrgCode": "UPD",
        "NsfStatusCode": "AC",
    })


def gen_valid_user_modify_delete():
    """request_user_modify — delete a DN from a known user's DN list."""
    gid = str(random.randint(100000, 999999))
    return make_packet("request_user_modify", {
        "ActionType": "delete",
        "PersonID": f"person-{uuid.uuid4().hex[:8]}",
        "UserGlobalID": gid,
        "DnList": [
            f"/C=US/O=Mock Institute/CN=User {gid}",
        ],
    })


def gen_valid_data_account_create():
    """data_account_create — pass DNs for an existing user (looked up by GlobalID)."""
    gid = str(random.randint(100000, 999999))
    return make_packet("data_account_create", {
        "ProjectID": f"PRJ-MOCK{random.randint(1000, 9999)}",
        "PersonID": f"person-{uuid.uuid4().hex[:8]}",
        "GlobalID": gid,
        "DnList": [
            f"/C=US/O=Mock Institute/CN=User {gid}",
            f"/DC=EDU/CN=user{gid}",
        ],
    })


def gen_valid_data_project_create():
    """data_project_create — pass DNs for an existing PI (looked up by GlobalID)."""
    gid = str(random.randint(100000, 999999))
    return make_packet("data_project_create", {
        "ProjectID": f"PRJ-MOCK{random.randint(1000, 9999)}",
        "PersonID": f"person-{uuid.uuid4().hex[:8]}",
        "GlobalID": gid,
        "DnList": [
            f"/C=US/O=Mock University/CN=PI {gid}",
        ],
    })


def gen_valid_project_inactivate():
    return make_packet("request_project_inactivate", {
        "ProjectID": f"PRJ-MOCK{random.randint(1000, 9999)}",
        "PersonID": f"person-{uuid.uuid4().hex[:8]}",
        "ResourceList": ["mock-cluster.example.edu"],
        "GrantNumber": f"TST{random.randint(100000, 999999)}",
    })


def gen_valid_project_reactivate():
    return make_packet("request_project_reactivate", {
        "ProjectID": f"PRJ-MOCK{random.randint(1000, 9999)}",
        "PersonID": f"person-{uuid.uuid4().hex[:8]}",
        "ResourceList": ["mock-cluster.example.edu"],
        "GrantNumber": f"TST{random.randint(100000, 999999)}",
    })


def gen_valid_account_inactivate():
    return make_packet("request_account_inactivate", {
        "ProjectID": f"PRJ-MOCK{random.randint(1000, 9999)}",
        "PersonID": f"person-{uuid.uuid4().hex[:8]}",
        "ResourceList": ["mock-cluster.example.edu"],
    })


def gen_valid_account_reactivate():
    return make_packet("request_account_reactivate", {
        "ProjectID": f"PRJ-MOCK{random.randint(1000, 9999)}",
        "PersonID": f"person-{uuid.uuid4().hex[:8]}",
        "ResourceList": ["mock-cluster.example.edu"],
    })


def gen_valid_person_merge():
    primary = str(random.randint(100000, 999999))
    secondary = str(random.randint(100000, 999999))
    return make_packet("request_person_merge", {
        "KeepGlobalID": primary,
        "KeepPersonID": f"person-keep-{uuid.uuid4().hex[:8]}",
        "DeleteGlobalID": secondary,
        "DeletePersonID": f"person-delete-{uuid.uuid4().hex[:8]}",
        "MergeReason": "Duplicate person records",
    })


def gen_inform_transaction_complete():
    """Canonical fields: StatusCode, Message, DetailCode."""
    return make_packet("inform_transaction_complete", {
        "StatusCode": "Success",
        "Message": "OK",
        "DetailCode": "1",
    })


# ----- Failure generators (intentionally malformed) -----

def gen_missing_global_id():
    """request_account_create with no UserGlobalID."""
    return make_packet("request_account_create", {
        "ProjectID": f"PRJ-FAIL{random.randint(1000, 9999)}",
        "GrantNumber": f"FAIL{random.randint(100000, 999999)}",
        "UserFirstName": "NoGlobalID",
        "UserLastName": "User",
        "UserEmail": "noglobal@example.edu",
        "ResourceList": ["mock-cluster.example.edu"],
    })


def gen_missing_grant_number():
    """request_project_create with no GrantNumber."""
    return make_packet("request_project_create", {
        "PiGlobalID": str(random.randint(100000, 999999)),
        "PiFirstName": "NoGrant",
        "PiLastName": "User",
        "PiEmail": "nogrant@example.edu",
        "NsfStatusCode": "AC",
        "ResourceList": ["mock-cluster.example.edu"],
    })


def gen_missing_pi_global_id():
    """request_project_create with no PiGlobalID."""
    return make_packet("request_project_create", {
        "GrantNumber": f"NOPI{random.randint(100000, 999999)}",
        "PiFirstName": "NoGlobal",
        "PiLastName": "Person",
        "PiEmail": "nopi@example.edu",
        "NsfStatusCode": "AC",
        "ResourceList": ["mock-cluster.example.edu"],
    })


def gen_unknown_packet_type():
    """Packet with an unrecognized type — should hit NoOpHandler."""
    return make_packet("request_something_unknown", {"SomeField": "SomeValue"})


def gen_empty_body():
    return make_packet("request_project_create", {})


# ----- dev_email scripted scenario -----

DEV_USER_GID = "100001"
DEV_MEMBER_PROJECTS = [
    ("DEV-PROJ-002", "Climate Modeling Group", "Alice", "Smith", "alice.smith@bogus.example.edu", "100002", "CoPI"),
    ("DEV-PROJ-003", "Particle Physics Sim",   "Bob",   "Johnson", "bob.johnson@bogus.example.edu", "100003", "Allocation Manager"),
    ("DEV-PROJ-004", "Genomics Pipeline",      "Carol", "Williams", "carol.williams@bogus.example.edu", "100004", "User"),
]


def gen_dev_email_scenario():
    if not DEV_EMAIL:
        app.logger.warning("DEV_EMAIL is not set; dev_email scenario emits no packets")
        return []
    packets = []
    packets.append(make_packet("request_project_create", {
        "GrantNumber": "DEV-PROJ-001",
        "PfosNumber": "PFOS-DEV-PROJ-001",
        "ProjectTitle": "Dev's Own Project",
        "PiGlobalID": DEV_USER_GID,
        "PiFirstName": "Dev",
        "PiLastName": "User",
        "PiEmail": DEV_EMAIL,
        "PiOrganization": "Dev Lab",
        "PiOrgCode": "DEV",
        "NsfStatusCode": "AC",
        "PiDnList": ["/C=US/O=Dev Lab/CN=Dev User"],
        "ServiceUnitsAllocated": "100000",
        "StartDate": "2026-01-01",
        "EndDate": "2026-12-31",
        "ResourceList": ["mock-cluster.example.edu"],
    }))
    for grant, title, pi_first, pi_last, pi_email, pi_gid, dev_role in DEV_MEMBER_PROJECTS:
        packets.append(make_packet("request_project_create", {
            "GrantNumber": grant,
            "PfosNumber": f"PFOS-{grant}",
            "ProjectTitle": title,
            "PiGlobalID": pi_gid,
            "PiFirstName": pi_first,
            "PiLastName": pi_last,
            "PiEmail": pi_email,
            "PiOrganization": "Bogus Lab",
            "PiOrgCode": "BOGUS",
            "NsfStatusCode": "AC",
            "PiDnList": [f"/C=US/O=Bogus Lab/CN={pi_first} {pi_last}"],
            "ServiceUnitsAllocated": "50000",
            "StartDate": "2026-01-01",
            "EndDate": "2026-12-31",
            "ResourceList": ["mock-cluster.example.edu"],
        }))
        packets.append(make_packet("request_account_create", {
            "ProjectID": f"PRJ-{grant}",
            "GrantNumber": grant,
            "UserPersonID": f"person-{DEV_USER_GID}-{grant}",
            "UserGlobalID": DEV_USER_GID,
            "UserFirstName": "Dev",
            "UserLastName": "User",
            "UserEmail": DEV_EMAIL,
            "UserOrganization": "Dev Lab",
            "UserOrgCode": "DEV",
            "NsfStatusCode": "AC",
            "UserDnList": ["/C=US/O=Dev Lab/CN=Dev User"],
            "UserRole": dev_role,
            "ResourceList": ["mock-cluster.example.edu"],
        }))
    return packets


# Scenario mix

SUCCESS_GENERATORS = [
    (gen_valid_project_create, 2),
    (gen_valid_account_create, 2),
    (gen_valid_user_modify, 1),
    (gen_valid_user_modify_delete, 1),
    (gen_valid_data_account_create, 1),
    (gen_valid_data_project_create, 1),
    (gen_valid_project_inactivate, 1),
    (gen_valid_project_reactivate, 1),
    (gen_valid_account_inactivate, 1),
    (gen_valid_account_reactivate, 1),
    (gen_valid_person_merge, 1),
    (gen_inform_transaction_complete, 1),
]

FAILURE_GENERATORS = [
    (gen_missing_global_id, 2),
    (gen_missing_grant_number, 1),
    (gen_missing_pi_global_id, 1),
    (gen_unknown_packet_type, 1),
    (gen_empty_body, 1),
]


def generate_batch(success_count=6, failure_count=4):
    """Mixed batch of success and failure packets."""
    packets = []
    success_pool = []
    for gen, weight in SUCCESS_GENERATORS:
        success_pool.extend([gen] * weight)
    failure_pool = []
    for gen, weight in FAILURE_GENERATORS:
        failure_pool.extend([gen] * weight)
    for _ in range(success_count):
        packets.append(random.choice(success_pool)())
    for _ in range(failure_count):
        packets.append(random.choice(failure_pool)())
    random.shuffle(packets)
    return packets


def generate_all_handlers_once():
    """One packet per handler type. Useful for verifying every handler runs.

    For request_person_merge to actually exercise the merge path (rather than
    erroring on missing users), the scenario pre-creates two users via
    request_account_create with deterministic UserGlobalIDs before emitting
    the merge. Same trick for data_account_create / data_project_create:
    a prior request_account_create establishes the user that the DN-pass
    handler will look up by GlobalID."""
    primary_gid = f"merge-primary-{random.randint(100000, 999999)}"
    secondary_gid = f"merge-secondary-{random.randint(100000, 999999)}"
    data_user_gid = f"data-user-{random.randint(100000, 999999)}"
    data_pi_gid = f"data-pi-{random.randint(100000, 999999)}"

    def account_create_with_gid(gid):
        return make_packet("request_account_create", {
            "ProjectID": f"PRJ-MOCK{random.randint(1000, 9999)}",
            "GrantNumber": f"TST{random.randint(100000, 999999)}",
            "UserPersonID": f"person-{uuid.uuid4().hex[:8]}",
            "UserGlobalID": gid,
            "UserFirstName": "Test",
            "UserLastName": "User",
            "UserEmail": f"{gid}@example.edu",
            "UserOrganization": "Mock Institute",
            "UserOrgCode": "MI",
            "NsfStatusCode": "AC",
            "UserDnList": [],
            "ResourceList": ["mock-cluster.example.edu"],
        })

    return [
        # Set up: create the users that downstream handlers need.
        account_create_with_gid(primary_gid),
        account_create_with_gid(secondary_gid),
        account_create_with_gid(data_user_gid),
        account_create_with_gid(data_pi_gid),
        # One packet per handler type:
        gen_valid_project_create(),
        gen_valid_account_create(),
        gen_valid_user_modify(),
        gen_valid_user_modify_delete(),
        make_packet("data_account_create", {
            "ProjectID": f"PRJ-MOCK{random.randint(1000, 9999)}",
            "PersonID": f"person-{uuid.uuid4().hex[:8]}",
            "GlobalID": data_user_gid,
            "DnList": [f"/C=US/O=Mock/CN=User {data_user_gid}"],
        }),
        make_packet("data_project_create", {
            "ProjectID": f"PRJ-MOCK{random.randint(1000, 9999)}",
            "PersonID": f"person-{uuid.uuid4().hex[:8]}",
            "GlobalID": data_pi_gid,
            "DnList": [f"/C=US/O=Mock/CN=PI {data_pi_gid}"],
        }),
        gen_valid_project_inactivate(),
        gen_valid_project_reactivate(),
        gen_valid_account_inactivate(),
        gen_valid_account_reactivate(),
        make_packet("request_person_merge", {
            "KeepGlobalID": primary_gid,
            "KeepPersonID": f"person-keep-{uuid.uuid4().hex[:8]}",
            "DeleteGlobalID": secondary_gid,
            "DeletePersonID": f"person-delete-{uuid.uuid4().hex[:8]}",
            "MergeReason": "all_handlers test scenario",
        }),
        gen_inform_transaction_complete(),
    ]


# Deterministic baseline scenario. Skips handlers that need a Custos UUID
# (request_account_create, the inactivate/reactivate handlers).

def gen_baseline_scenario():
    return [
        make_packet("request_project_create", {
            "GrantNumber": "BL-001",
            "PfosNumber": "PFOS-BL-001",
            "ProjectTitle": "Baseline Project 1",
            "PiGlobalID": "bl-pi-001",
            "PiFirstName": "Pat",
            "PiLastName": "First",
            "PiEmail": "pat.first@baseline.example.edu",
            "PiOrganization": "Baseline Org",
            "PiOrgCode": "BASELINE",
            "NsfStatusCode": "AC",
            "PiDnList": ["/C=US/O=Baseline Org/CN=Pat First"],
            "ServiceUnitsAllocated": "10000",
            "StartDate": "2026-01-01",
            "EndDate": "2026-12-31",
            "ResourceList": ["baseline-cluster.example.edu"],
            "AllocationType": "new",
        }),
        make_packet("request_project_create", {
            "GrantNumber": "BL-002",
            "PfosNumber": "PFOS-BL-002",
            "ProjectTitle": "Baseline Project 2",
            "PiGlobalID": "bl-pi-002",
            "PiFirstName": "Sam",
            "PiLastName": "Second",
            "PiEmail": "sam.second@baseline.example.edu",
            "PiOrganization": "Baseline Org",
            "PiOrgCode": "BASELINE",
            "NsfStatusCode": "AC",
            "PiDnList": [],
            "ServiceUnitsAllocated": "20000",
            "StartDate": "2026-01-01",
            "EndDate": "2026-12-31",
            "ResourceList": ["baseline-cluster.example.edu"],
            "AllocationType": "new",
        }),
        # Re-delivery of BL-001 as a supplement; writes a compute_allocation_diffs row.
        make_packet("request_project_create", {
            "GrantNumber": "BL-001",
            "PfosNumber": "PFOS-BL-001",
            "ProjectTitle": "Baseline Project 1",
            "PiGlobalID": "bl-pi-001",
            "PiFirstName": "Pat",
            "PiLastName": "First",
            "PiEmail": "pat.first@baseline.example.edu",
            "PiOrganization": "Baseline Org",
            "PiOrgCode": "BASELINE",
            "NsfStatusCode": "AC",
            "ServiceUnitsAllocated": "5000",
            "StartDate": "2026-01-01",
            "EndDate": "2026-12-31",
            "ResourceList": ["baseline-cluster.example.edu"],
            "AllocationType": "supplement",
        }),
        make_packet("data_project_create", {
            "ProjectID": "BL-001",
            "PersonID": "bl-pi-001-person",
            "GlobalID": "bl-pi-001",
            "DnList": [
                "/C=US/O=Baseline Org/CN=Pat First Extra",
                "/DC=EDU/CN=patfirst",
            ],
        }),
        make_packet("data_account_create", {
            "ProjectID": "BL-002",
            "PersonID": "bl-pi-002-person",
            "GlobalID": "bl-pi-002",
            "DnList": [
                "/C=US/O=Baseline Org/CN=Sam Second",
            ],
        }),
        make_packet("request_user_modify", {
            "ActionType": "replace",
            "ProjectID": "BL-001",
            "PersonID": "bl-pi-001-person",
            "UserPersonID": "bl-pi-001-user",
            "UserGlobalID": "bl-pi-001",
            "UserFirstName": "Pat",
            "UserLastName": "Madison",
            "UserEmail": "pat.madison@baseline.example.edu",
            "UserOrganization": "Baseline Org",
            "UserOrgCode": "BASELINE",
            "NsfStatusCode": "AC",
        }),
        make_packet("request_person_merge", {
            "KeepGlobalID": "bl-pi-001",
            "KeepPersonID": "bl-keep-person",
            "DeleteGlobalID": "bl-pi-002",
            "DeletePersonID": "bl-delete-person",
            "MergeReason": "Duplicate person records",
        }),
        # Remove one of the survivor's DNs via the delete ActionType.
        make_packet("request_user_modify", {
            "ActionType": "delete",
            "PersonID": "bl-pi-001-person",
            "UserGlobalID": "bl-pi-001",
            "DnList": [
                "/DC=EDU/CN=patfirst",
            ],
        }),
        # request_account_create packets. ProjectID is a placeholder filled
        # in from the matching notify_project_create reply. PI memberships
        # come from request_project_create; these cover the non-PI roles.

        # BL-001 CO_PI
        make_packet("request_account_create", {
            "ProjectID": "__GRANT__BL-001",
            "GrantNumber": "BL-001",
            "UserPersonID": "bl-copi-001-person",
            "UserGlobalID": "bl-copi-001",
            "UserFirstName": "Casey",
            "UserLastName": "Collaborator",
            "UserEmail": "casey.collab@baseline.example.edu",
            "UserOrganization": "Baseline Org",
            "UserOrgCode": "BASELINE",
            "NsfStatusCode": "AC",
            "UserDnList": ["/C=US/O=Baseline Org/CN=Casey Collaborator"],
            "UserRole": "CO_PI",
            "ResourceList": ["baseline-cluster.example.edu"],
        }),
        # BL-001 MEMBER
        make_packet("request_account_create", {
            "ProjectID": "__GRANT__BL-001",
            "GrantNumber": "BL-001",
            "UserPersonID": "bl-user-001-person",
            "UserGlobalID": "bl-user-001",
            "UserFirstName": "Riley",
            "UserLastName": "Researcher",
            "UserEmail": "riley.research@baseline.example.edu",
            "UserOrganization": "Baseline Org",
            "UserOrgCode": "BASELINE",
            "NsfStatusCode": "AC",
            "UserDnList": ["/C=US/O=Baseline Org/CN=Riley Researcher"],
            "UserRole": "MEMBER",
            "ResourceList": ["baseline-cluster.example.edu"],
        }),
        # BL-002 ALLOCATION_MANAGER
        make_packet("request_account_create", {
            "ProjectID": "__GRANT__BL-002",
            "GrantNumber": "BL-002",
            "UserPersonID": "bl-mgr-002-person",
            "UserGlobalID": "bl-mgr-002",
            "UserFirstName": "Morgan",
            "UserLastName": "Manager",
            "UserEmail": "morgan.manager@baseline.example.edu",
            "UserOrganization": "Baseline Org",
            "UserOrgCode": "BASELINE",
            "NsfStatusCode": "AC",
            "UserDnList": ["/C=US/O=Baseline Org/CN=Morgan Manager"],
            "UserRole": "ALLOCATION_MANAGER",
            "ResourceList": ["baseline-cluster.example.edu"],
        }),
        make_packet("inform_transaction_complete", {
            "StatusCode": "Success",
            "Message": "Baseline complete",
            "DetailCode": "1",
        }),
    ]


# ----- API endpoints -----

@app.route("/packets/<site>", methods=["GET"])
def get_packets(site):
    if not pending_packets:
        return jsonify([])
    ready, holdback = [], []
    for pkt in pending_packets:
        pid = pkt.get("body", {}).get("ProjectID", "")
        if isinstance(pid, str) and pid.startswith("__GRANT__"):
            gn = pid[len("__GRANT__"):]
            resolved = grant_to_project_id.get(gn)
            if resolved:
                pkt["body"]["ProjectID"] = resolved
                ready.append(pkt)
            else:
                holdback.append(pkt)
        else:
            ready.append(pkt)
    pending_packets[:] = holdback
    if not ready:
        return jsonify([])
    stats["fetched"] += len(ready)
    app.logger.info(f"Serving {len(ready)} packets to site {site}")
    return jsonify(ready)


@app.route("/packets/<site>/<int:packet_rec_id>/reply", methods=["POST"])
def reply_to_packet(site, packet_rec_id):
    payload = request.get_json(silent=True) or {}
    replied_packets[packet_rec_id] = payload
    stats["replied"] += 1
    # Capture the Custos UUID from notify_project_create so the holdback
    # packets carrying "__GRANT__<GN>" can resolve.
    if payload.get("type") == "notify_project_create":
        body = payload.get("body", {}) or {}
        gn = body.get("GrantNumber")
        pid = body.get("ProjectID")
        if gn and pid:
            grant_to_project_id[gn] = pid
            app.logger.info(f"Captured GrantNumber {gn} -> ProjectID {pid}")
    return jsonify({"message": "Reply accepted"}), 200


@app.route("/test/<site>/reset", methods=["POST"])
def reset(site):
    pending_packets.clear()
    replied_packets.clear()
    grant_to_project_id.clear()
    stats["created"] = 0
    stats["fetched"] = 0
    stats["replied"] = 0
    app.logger.info(f"Reset state for site {site}")
    return jsonify({"message": f"Reset site data for {site}", "result": None})


@app.route("/test/<site>/scenarios", methods=["POST"])
def create_scenario(site):
    scenario_type = request.args.get("type", "mixed")
    if scenario_type == "mixed":
        packets = generate_batch(success_count=6, failure_count=4)
    elif scenario_type == "failures_only":
        packets = generate_batch(success_count=0, failure_count=8)
    elif scenario_type == "success_only":
        packets = generate_batch(success_count=8, failure_count=0)
    elif scenario_type == "heavy":
        packets = generate_batch(success_count=15, failure_count=10)
    elif scenario_type == "all_handlers":
        packets = generate_all_handlers_once()
    elif scenario_type == "dev_email":
        packets = gen_dev_email_scenario()
    elif scenario_type == "baseline":
        packets = gen_baseline_scenario()
    else:
        packets = generate_batch(success_count=3, failure_count=2)
    pending_packets.extend(packets)
    stats["created"] += len(packets)
    app.logger.info(f"Created {len(packets)} packets ({scenario_type}) for site {site}")
    return jsonify({"message": "Test scenario initiated", "result": None})


@app.route("/stats", methods=["GET"])
def get_stats():
    return jsonify({"pending": len(pending_packets), "replied": len(replied_packets), "stats": stats})


if __name__ == "__main__":
    print("Mock AMIE Server starting on http://localhost:8180")
    print("Point your connector at: AMIE_BASE_URL=http://localhost:8180")
    print()
    print("Scenario types:")
    print("  POST /test/{site}/scenarios?type=mixed          — 6 success + 4 failure")
    print("  POST /test/{site}/scenarios?type=failures_only  — 8 failures")
    print("  POST /test/{site}/scenarios?type=success_only   — 8 successes")
    print("  POST /test/{site}/scenarios?type=heavy          — 15 success + 10 failure")
    print("  POST /test/{site}/scenarios?type=all_handlers   — one packet per handler type")
    print("  POST /test/{site}/scenarios?type=dev_email      — scripted dev_email scenario")
    print()
    if DEV_EMAIL:
        print(f"DEV_EMAIL injection enabled: {DEV_EMAIL}")
    else:
        print("DEV_EMAIL unset; dev_email scenario will be empty")
    print()
    app.run(host="0.0.0.0", port=8180, debug=False)
