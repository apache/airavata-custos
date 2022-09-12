#!/usr/bin/python
# -*- coding: utf-8 -*-

# (c) 2019, Simon Dodsley (simon@purestorage.com)
# GNU General Public License v3.0+ (see COPYING or
# https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function

__metaclass__ = type

ANSIBLE_METADATA = {
    "metadata_version": "1.1",
    "status": ["preview"],
    "supported_by": "community",
}

DOCUMENTATION = r"""
---
module: purefb_info
version_added: '1.0.0'
short_description: Collect information from Pure Storage FlashBlade
description:
  - Collect information from a Pure Storage FlashBlade running the
    Purity//FB operating system. By default, the module will collect basic
    information including hosts, host groups, protection
    groups and volume counts. Additional information can be collected
    based on the configured set of arguements.
author:
  - Pure Storage Ansible Team (@sdodsley) <pure-ansible-team@purestorage.com>
options:
  gather_subset:
    description:
      - When supplied, this argument will define the information to be collected.
        Possible values for this include all, minimum, config, performance,
        capacity, network, subnets, lags, filesystems, snapshots, buckets,
        replication, policies, arrays, accounts, admins, ad and kerberos.
    required: false
    type: list
    elements: str
    default: minimum
extends_documentation_fragment:
  - purestorage.flashblade.purestorage.fb
"""

EXAMPLES = r"""
- name: collect default set of info
  purefb_info:
    fb_url: 10.10.10.2
    api_token: T-55a68eb5-c785-4720-a2ca-8b03903bf641
  register: blade_info
- name: show default information
  debug:
    msg: "{{ blade_info['purefb_info']['default'] }}"

- name: collect configuration and capacity info
  purefb_info:
    gather_subset:
      - config
    fb_url: 10.10.10.2
    api_token: T-55a68eb5-c785-4720-a2ca-8b03903bf641
  register: blade_info
- name: show config information
  debug:
    msg: "{{ blade_info['purefb_info']['config'] }}"

- name: collect all info
  purefb_info:
    gather_subset:
      - all
    fb_url: 10.10.10.2
    api_token: T-55a68eb5-c785-4720-a2ca-8b03903bf641
  register: blade_info
- name: show all information
  debug:
    msg: "{{ blade_info['purefb_info'] }}"
"""

RETURN = r"""
purefb_info:
  description: Returns the information collected from the FlashBlade
  returned: always
  type: complex
  sample: {
        "admins": {
            "pureuser": {
                "api_token_timeout": null,
                "local": true,
                "public_key": null
            },
            "another_user": {
                "api_token_timeout": null,
                "local": false,
                "public_key": null
            },
        },
        "capacity": {
            "aggregate": {
                "data_reduction": 1.1179228,
                "snapshots": 0,
                "total_physical": 17519748439,
                "unique": 17519748439,
                "virtual": 19585726464
            },
            "file-system": {
                "data_reduction": 1.3642412,
                "snapshots": 0,
                "total_physical": 4748219708,
                "unique": 4748219708,
                "virtual": 6477716992
            },
            "object-store": {
                "data_reduction": 1.0263462,
                "snapshots": 0,
                "total_physical": 12771528731,
                "unique": 12771528731,
                "virtual": 6477716992
            },
            "total": 83359896948925
        },
        "config": {
            "alert_watchers": {
                "enabled": true,
                "name": "notify@acmestorage.com"
            },
            "array_management": {
                "base_dn": null,
                "bind_password": null,
                "bind_user": null,
                "enabled": false,
                "name": "management",
                "services": [
                    "management"
                ],
                "uris": []
            },
            "directory_service_roles": {
                "array_admin": {
                    "group": null,
                    "group_base": null
                },
                "ops_admin": {
                    "group": null,
                    "group_base": null
                },
                "readonly": {
                    "group": null,
                    "group_base": null
                },
                "storage_admin": {
                    "group": null,
                    "group_base": null
                }
            },
            "dns": {
                "domain": "demo.acmestorage.com",
                "name": "demo-fb-1",
                "nameservers": [
                    "8.8.8.8"
                ],
                "search": [
                    "demo.acmestorage.com"
                ]
            },
            "nfs_directory_service": {
                "base_dn": null,
                "bind_password": null,
                "bind_user": null,
                "enabled": false,
                "name": "nfs",
                "services": [
                    "nfs"
                ],
                "uris": []
            },
            "ntp": [
                "0.ntp.pool.org"
            ],
            "smb_directory_service": {
                "base_dn": null,
                "bind_password": null,
                "bind_user": null,
                "enabled": false,
                "name": "smb",
                "services": [
                    "smb"
                ],
                "uris": []
            },
            "smtp": {
                "name": "demo-fb-1",
                "relay_host": null,
                "sender_domain": "acmestorage.com"
            },
            "ssl_certs": {
                "certificate": "-----BEGIN CERTIFICATE-----\n\n-----END CERTIFICATE-----",
                "common_name": "Acme Storage",
                "country": "US",
                "email": null,
                "intermediate_certificate": null,
                "issued_by": "Acme Storage",
                "issued_to": "Acme Storage",
                "key_size": 4096,
                "locality": null,
                "name": "global",
                "organization": "Acme Storage",
                "organizational_unit": "Acme Storage",
                "passphrase": null,
                "private_key": null,
                "state": null,
                "status": "self-signed",
                "valid_from": "1508433967000",
                "valid_to": "2458833967000"
            }
        },
        "default": {
            "blades": 15,
            "buckets": 7,
            "filesystems": 2,
            "flashblade_name": "demo-fb-1",
            "object_store_accounts": 1,
            "object_store_users": 1,
            "purity_version": "2.2.0",
            "snapshots": 1,
            "total_capacity": 83359896948925
        },
        "filesystems": {
            "k8s-pvc-d24b1357-579e-11e8-811f-ecf4bbc88f54": {
                "destroyed": false,
                "fast_remove": false,
                "hard_limit": true,
                "nfs_rules": "10.21.255.0/24(rw,no_root_squash)",
                "provisioned": 21474836480,
                "snapshot_enabled": false
            },
            "z": {
                "destroyed": false,
                "fast_remove": false,
                "hard_limit": false,
                "provisioned": 1073741824,
                "snapshot_enabled": false
            }
        },
        "lag": {
            "uplink": {
                "lag_speed": 0,
                "port_speed": 40000000000,
                "ports": [
                    {
                        "name": "CH1.FM1.ETH1.1"
                    },
                    {
                        "name": "CH1.FM1.ETH1.2"
                    },
                ],
                "status": "healthy"
            }
        },
        "network": {
            "fm1.admin0": {
                "address": "10.10.100.6",
                "gateway": "10.10.100.1",
                "mtu": 1500,
                "netmask": "255.255.255.0",
                "services": [
                    "support"
                ],
                "type": "vip",
                "vlan": 2200
            },
            "fm2.admin0": {
                "address": "10.10.100.7",
                "gateway": "10.10.100.1",
                "mtu": 1500,
                "netmask": "255.255.255.0",
                "services": [
                    "support"
                ],
                "type": "vip",
                "vlan": 2200
            },
            "nfs1": {
                "address": "10.10.100.4",
                "gateway": "10.10.100.1",
                "mtu": 1500,
                "netmask": "255.255.255.0",
                "services": [
                    "data"
                ],
                "type": "vip",
                "vlan": 2200
            },
            "vir0": {
                "address": "10.10.100.5",
                "gateway": "10.10.100.1",
                "mtu": 1500,
                "netmask": "255.255.255.0",
                "services": [
                    "management"
                ],
                "type": "vip",
                "vlan": 2200
            }
        },
        "performance": {
            "aggregate": {
                "bytes_per_op": 0,
                "bytes_per_read": 0,
                "bytes_per_write": 0,
                "read_bytes_per_sec": 0,
                "reads_per_sec": 0,
                "usec_per_other_op": 0,
                "usec_per_read_op": 0,
                "usec_per_write_op": 0,
                "write_bytes_per_sec": 0,
                "writes_per_sec": 0
            },
            "http": {
                "bytes_per_op": 0,
                "bytes_per_read": 0,
                "bytes_per_write": 0,
                "read_bytes_per_sec": 0,
                "reads_per_sec": 0,
                "usec_per_other_op": 0,
                "usec_per_read_op": 0,
                "usec_per_write_op": 0,
                "write_bytes_per_sec": 0,
                "writes_per_sec": 0
            },
            "nfs": {
                "bytes_per_op": 0,
                "bytes_per_read": 0,
                "bytes_per_write": 0,
                "read_bytes_per_sec": 0,
                "reads_per_sec": 0,
                "usec_per_other_op": 0,
                "usec_per_read_op": 0,
                "usec_per_write_op": 0,
                "write_bytes_per_sec": 0,
                "writes_per_sec": 0
            },
            "s3": {
                "bytes_per_op": 0,
                "bytes_per_read": 0,
                "bytes_per_write": 0,
                "read_bytes_per_sec": 0,
                "reads_per_sec": 0,
                "usec_per_other_op": 0,
                "usec_per_read_op": 0,
                "usec_per_write_op": 0,
                "write_bytes_per_sec": 0,
                "writes_per_sec": 0
            }
        },
        "snapshots": {
            "z.188": {
                "destroyed": false,
                "source": "z",
                "source_destroyed": false,
                "suffix": "188"
            }
        },
        "subnet": {
            "new-mgmt": {
                "gateway": "10.10.100.1",
                "interfaces": [
                    {
                        "name": "fm1.admin0"
                    },
                    {
                        "name": "fm2.admin0"
                    },
                    {
                        "name": "nfs1"
                    },
                    {
                        "name": "vir0"
                    }
                ],
                "lag": "uplink",
                "mtu": 1500,
                "prefix": "10.10.100.0/24",
                "services": [
                    "data",
                    "management",
                    "support"
                ],
                "vlan": 2200
            }
        }
    }
"""


from ansible.module_utils.basic import AnsibleModule
from ansible_collections.purestorage.flashblade.plugins.module_utils.purefb import (
    get_blade,
    get_system,
    purefb_argument_spec,
)


MIN_REQUIRED_API_VERSION = "1.3"
HARD_LIMIT_API_VERSION = "1.4"
POLICIES_API_VERSION = "1.5"
CERT_GROUPS_API_VERSION = "1.8"
REPLICATION_API_VERSION = "1.9"
MULTIPROTOCOL_API_VERSION = "1.11"
MIN_32_API = "2.0"


def generate_default_dict(module, blade):
    default_info = {}
    defaults = blade.arrays.list_arrays().items[0]
    default_info["flashblade_name"] = defaults.name
    default_info["purity_version"] = defaults.version
    default_info["filesystems"] = len(blade.file_systems.list_file_systems().items)
    default_info["snapshots"] = len(
        blade.file_system_snapshots.list_file_system_snapshots().items
    )
    default_info["buckets"] = len(blade.buckets.list_buckets().items)
    default_info["object_store_users"] = len(
        blade.object_store_users.list_object_store_users().items
    )
    default_info["object_store_accounts"] = len(
        blade.object_store_accounts.list_object_store_accounts().items
    )
    default_info["blades"] = len(blade.blade.list_blades().items)
    default_info["certificates"] = len(blade.certificates.list_certificates().items)
    default_info["total_capacity"] = blade.arrays.list_arrays_space().items[0].capacity
    api_version = blade.api_version.list_versions().versions
    if POLICIES_API_VERSION in api_version:
        default_info["policies"] = len(blade.policies.list_policies().items)
    if CERT_GROUPS_API_VERSION in api_version:
        default_info["certificate_groups"] = len(
            blade.certificate_groups.list_certificate_groups().items
        )
    if REPLICATION_API_VERSION in api_version:
        default_info["fs_replicas"] = len(
            blade.file_system_replica_links.list_file_system_replica_links().items
        )
        default_info["remote_credentials"] = len(
            blade.object_store_remote_credentials.list_object_store_remote_credentials().items
        )
        default_info["bucket_replicas"] = len(
            blade.bucket_replica_links.list_bucket_replica_links().items
        )
        default_info["connected_arrays"] = len(
            blade.array_connections.list_array_connections().items
        )
        default_info["targets"] = len(blade.targets.list_targets().items)
        default_info["kerberos_keytabs"] = len(blade.keytabs.list_keytabs().items)
    # This section is just for REST 2.x features
    if MIN_32_API in api_version:
        blade = get_system(module)
        default_info["object_store_virtual_hosts"] = len(
            blade.get_object_store_virtual_hosts().items
        )
        default_info["api_clients"] = len(blade.get_api_clients().items)
        default_info["idle_timeout"] = int(
            list(blade.get_arrays().items)[0].idle_timeout / 60000
        )
        if list(blade.get_arrays_eula().items)[0].signature.accepted:
            default_info["EULA"] = "Signed"
        else:
            default_info["EULA"] = "Not Signed"
    return default_info


def generate_perf_dict(blade):
    perf_info = {}
    total_perf = blade.arrays.list_arrays_performance()
    http_perf = blade.arrays.list_arrays_performance(protocol="http")
    s3_perf = blade.arrays.list_arrays_performance(protocol="s3")
    nfs_perf = blade.arrays.list_arrays_performance(protocol="nfs")
    perf_info["aggregate"] = {
        "bytes_per_op": total_perf.items[0].bytes_per_op,
        "bytes_per_read": total_perf.items[0].bytes_per_read,
        "bytes_per_write": total_perf.items[0].bytes_per_write,
        "read_bytes_per_sec": total_perf.items[0].read_bytes_per_sec,
        "reads_per_sec": total_perf.items[0].reads_per_sec,
        "usec_per_other_op": total_perf.items[0].usec_per_other_op,
        "usec_per_read_op": total_perf.items[0].usec_per_read_op,
        "usec_per_write_op": total_perf.items[0].usec_per_write_op,
        "write_bytes_per_sec": total_perf.items[0].write_bytes_per_sec,
        "writes_per_sec": total_perf.items[0].writes_per_sec,
    }
    perf_info["http"] = {
        "bytes_per_op": http_perf.items[0].bytes_per_op,
        "bytes_per_read": http_perf.items[0].bytes_per_read,
        "bytes_per_write": http_perf.items[0].bytes_per_write,
        "read_bytes_per_sec": http_perf.items[0].read_bytes_per_sec,
        "reads_per_sec": http_perf.items[0].reads_per_sec,
        "usec_per_other_op": http_perf.items[0].usec_per_other_op,
        "usec_per_read_op": http_perf.items[0].usec_per_read_op,
        "usec_per_write_op": http_perf.items[0].usec_per_write_op,
        "write_bytes_per_sec": http_perf.items[0].write_bytes_per_sec,
        "writes_per_sec": http_perf.items[0].writes_per_sec,
    }
    perf_info["s3"] = {
        "bytes_per_op": s3_perf.items[0].bytes_per_op,
        "bytes_per_read": s3_perf.items[0].bytes_per_read,
        "bytes_per_write": s3_perf.items[0].bytes_per_write,
        "read_bytes_per_sec": s3_perf.items[0].read_bytes_per_sec,
        "reads_per_sec": s3_perf.items[0].reads_per_sec,
        "usec_per_other_op": s3_perf.items[0].usec_per_other_op,
        "usec_per_read_op": s3_perf.items[0].usec_per_read_op,
        "usec_per_write_op": s3_perf.items[0].usec_per_write_op,
        "write_bytes_per_sec": s3_perf.items[0].write_bytes_per_sec,
        "writes_per_sec": s3_perf.items[0].writes_per_sec,
    }
    perf_info["nfs"] = {
        "bytes_per_op": nfs_perf.items[0].bytes_per_op,
        "bytes_per_read": nfs_perf.items[0].bytes_per_read,
        "bytes_per_write": nfs_perf.items[0].bytes_per_write,
        "read_bytes_per_sec": nfs_perf.items[0].read_bytes_per_sec,
        "reads_per_sec": nfs_perf.items[0].reads_per_sec,
        "usec_per_other_op": nfs_perf.items[0].usec_per_other_op,
        "usec_per_read_op": nfs_perf.items[0].usec_per_read_op,
        "usec_per_write_op": nfs_perf.items[0].usec_per_write_op,
        "write_bytes_per_sec": nfs_perf.items[0].write_bytes_per_sec,
        "writes_per_sec": nfs_perf.items[0].writes_per_sec,
    }
    api_version = blade.api_version.list_versions().versions
    if REPLICATION_API_VERSION in api_version:
        file_repl_perf = (
            blade.array_connections.list_array_connections_performance_replication(
                type="file-system"
            )
        )
        obj_repl_perf = (
            blade.array_connections.list_array_connections_performance_replication(
                type="object-store"
            )
        )
        if len(file_repl_perf.total):
            perf_info["file_replication"] = {
                "received_bytes_per_sec": file_repl_perf.total[
                    0
                ].periodic.received_bytes_per_sec,
                "transmitted_bytes_per_sec": file_repl_perf.total[
                    0
                ].periodic.transmitted_bytes_per_sec,
            }
        if len(obj_repl_perf.total):
            perf_info["object_replication"] = {
                "received_bytes_per_sec": obj_repl_perf.total[
                    0
                ].periodic.received_bytes_per_sec,
                "transmitted_bytes_per_sec": obj_repl_perf.total[
                    0
                ].periodic.transmitted_bytes_per_sec,
            }
    return perf_info


def generate_config_dict(blade):
    config_info = {}
    config_info["dns"] = blade.dns.list_dns().items[0].to_dict()
    config_info["smtp"] = blade.smtp.list_smtp().items[0].to_dict()
    try:
        config_info["alert_watchers"] = (
            blade.alert_watchers.list_alert_watchers().items[0].to_dict()
        )
    except Exception:
        config_info["alert_watchers"] = ""
    api_version = blade.api_version.list_versions().versions
    if HARD_LIMIT_API_VERSION in api_version:
        config_info["array_management"] = (
            blade.directory_services.list_directory_services(names=["management"])
            .items[0]
            .to_dict()
        )
        config_info["directory_service_roles"] = {}
        roles = blade.directory_services.list_directory_services_roles()
        for role in range(0, len(roles.items)):
            role_name = roles.items[role].name
            config_info["directory_service_roles"][role_name] = {
                "group": roles.items[role].group,
                "group_base": roles.items[role].group_base,
            }
    config_info["nfs_directory_service"] = (
        blade.directory_services.list_directory_services(names=["nfs"])
        .items[0]
        .to_dict()
    )
    config_info["smb_directory_service"] = (
        blade.directory_services.list_directory_services(names=["smb"])
        .items[0]
        .to_dict()
    )
    config_info["ntp"] = blade.arrays.list_arrays().items[0].ntp_servers
    config_info["ssl_certs"] = blade.certificates.list_certificates().items[0].to_dict()
    api_version = blade.api_version.list_versions().versions
    if CERT_GROUPS_API_VERSION in api_version:
        try:
            config_info["certificate_groups"] = (
                blade.certificate_groups.list_certificate_groups().items[0].to_dict()
            )
        except Exception:
            config_info["certificate_groups"] = ""
    if REPLICATION_API_VERSION in api_version:
        config_info["snmp_agents"] = {}
        snmp_agents = blade.snmp_agents.list_snmp_agents()
        for agent in range(0, len(snmp_agents.items)):
            agent_name = snmp_agents.items[agent].name
            config_info["snmp_agents"][agent_name] = {
                "version": snmp_agents.items[agent].version,
                "engine_id": snmp_agents.items[agent].engine_id,
            }
            if config_info["snmp_agents"][agent_name]["version"] == "v3":
                config_info["snmp_agents"][agent_name][
                    "auth_protocol"
                ] = snmp_agents.items[agent].v3.auth_protocol
                config_info["snmp_agents"][agent_name][
                    "privacy_protocol"
                ] = snmp_agents.items[agent].v3.privacy_protocol
                config_info["snmp_agents"][agent_name]["user"] = snmp_agents.items[
                    agent
                ].v3.user
        config_info["snmp_managers"] = {}
        snmp_managers = blade.snmp_managers.list_snmp_managers()
        for manager in range(0, len(snmp_managers.items)):
            mgr_name = snmp_managers.items[manager].name
            config_info["snmp_managers"][mgr_name] = {
                "version": snmp_managers.items[manager].version,
                "host": snmp_managers.items[manager].host,
                "notification": snmp_managers.items[manager].notification,
            }
            if config_info["snmp_managers"][mgr_name]["version"] == "v3":
                config_info["snmp_managers"][mgr_name][
                    "auth_protocol"
                ] = snmp_managers.items[manager].v3.auth_protocol
                config_info["snmp_managers"][mgr_name][
                    "privacy_protocol"
                ] = snmp_managers.items[manager].v3.privacy_protocol
                config_info["snmp_managers"][mgr_name]["user"] = snmp_managers.items[
                    manager
                ].v3.user
    return config_info


def generate_subnet_dict(blade):
    sub_info = {}
    subnets = blade.subnets.list_subnets()
    for sub in range(0, len(subnets.items)):
        sub_name = subnets.items[sub].name
        if subnets.items[sub].enabled:
            sub_info[sub_name] = {
                "gateway": subnets.items[sub].gateway,
                "mtu": subnets.items[sub].mtu,
                "vlan": subnets.items[sub].vlan,
                "prefix": subnets.items[sub].prefix,
                "services": subnets.items[sub].services,
            }
            sub_info[sub_name]["lag"] = subnets.items[sub].link_aggregation_group.name
            sub_info[sub_name]["interfaces"] = []
            for iface in range(0, len(subnets.items[sub].interfaces)):
                sub_info[sub_name]["interfaces"].append(
                    {"name": subnets.items[sub].interfaces[iface].name}
                )
    return sub_info


def generate_lag_dict(blade):
    lag_info = {}
    groups = blade.link_aggregation_groups.list_link_aggregation_groups()
    for groupcnt in range(0, len(groups.items)):
        lag_name = groups.items[groupcnt].name
        lag_info[lag_name] = {
            "lag_speed": groups.items[groupcnt].lag_speed,
            "port_speed": groups.items[groupcnt].port_speed,
            "status": groups.items[groupcnt].status,
        }
        lag_info[lag_name]["ports"] = []
        for port in range(0, len(groups.items[groupcnt].ports)):
            lag_info[lag_name]["ports"].append(
                {"name": groups.items[groupcnt].ports[port].name}
            )
    return lag_info


def generate_admin_dict(blade):
    admin_info = {}
    api_version = blade.api_version.list_versions().versions
    if MULTIPROTOCOL_API_VERSION in api_version:
        admins = blade.admins.list_admins()
        for admin in range(0, len(admins.items)):
            admin_name = admins.items[admin].name
            admin_info[admin_name] = {
                "api_token_timeout": admins.items[admin].api_token_timeout,
                "public_key": admins.items[admin].public_key,
                "local": admins.items[admin].is_local,
            }
    return admin_info


def generate_targets_dict(blade):
    targets_info = {}
    targets = blade.targets.list_targets()
    for target in range(0, len(targets.items)):
        target_name = targets.items[target].name
        targets_info[target_name] = {
            "address": targets.items[target].address,
            "status": targets.items[target].status,
            "status_details": targets.items[target].status_details,
        }
    return targets_info


def generate_remote_creds_dict(blade):
    remote_creds_info = {}
    remote_creds = (
        blade.object_store_remote_credentials.list_object_store_remote_credentials()
    )
    for cred_cnt in range(0, len(remote_creds.items)):
        cred_name = remote_creds.items[cred_cnt].name
        remote_creds_info[cred_name] = {
            "access_key": remote_creds.items[cred_cnt].access_key_id,
            "remote_array": remote_creds.items[cred_cnt].remote.name,
        }
    return remote_creds_info


def generate_file_repl_dict(blade):
    file_repl_info = {}
    file_links = blade.file_system_replica_links.list_file_system_replica_links()
    for linkcnt in range(0, len(file_links.items)):
        fs_name = file_links.items[linkcnt].local_file_system.name
        file_repl_info[fs_name] = {
            "direction": file_links.items[linkcnt].direction,
            "lag": file_links.items[linkcnt].lag,
            "status": file_links.items[linkcnt].status,
            "remote_fs": file_links.items[linkcnt].remote.name
            + ":"
            + file_links.items[linkcnt].remote_file_system.name,
            "recovery_point": file_links.items[linkcnt].recovery_point,
        }
        file_repl_info[fs_name]["policies"] = []
        for policy_cnt in range(0, len(file_links.items[linkcnt].policies)):
            file_repl_info[fs_name]["policies"].append(
                file_links.items[linkcnt].policies[policy_cnt].display_name
            )
    return file_repl_info


def generate_bucket_repl_dict(blade):
    bucket_repl_info = {}
    bucket_links = blade.bucket_replica_links.list_bucket_replica_links()
    for linkcnt in range(0, len(bucket_links.items)):
        bucket_name = bucket_links.items[linkcnt].local_bucket.name
        bucket_repl_info[bucket_name] = {
            "direction": bucket_links.items[linkcnt].direction,
            "lag": bucket_links.items[linkcnt].lag,
            "paused": bucket_links.items[linkcnt].paused,
            "status": bucket_links.items[linkcnt].status,
            "remote_bucket": bucket_links.items[linkcnt].remote_bucket.name,
            "remote_credentials": bucket_links.items[linkcnt].remote_credentials.name,
            "recovery_point": bucket_links.items[linkcnt].recovery_point,
        }
    return bucket_repl_info


def generate_network_dict(blade):
    net_info = {}
    ports = blade.network_interfaces.list_network_interfaces()
    for portcnt in range(0, len(ports.items)):
        int_name = ports.items[portcnt].name
        if ports.items[portcnt].enabled:
            net_info[int_name] = {
                "type": ports.items[portcnt].type,
                "mtu": ports.items[portcnt].mtu,
                "vlan": ports.items[portcnt].vlan,
                "address": ports.items[portcnt].address,
                "services": ports.items[portcnt].services,
                "gateway": ports.items[portcnt].gateway,
                "netmask": ports.items[portcnt].netmask,
            }
    return net_info


def generate_capacity_dict(blade):
    capacity_info = {}
    total_cap = blade.arrays.list_arrays_space()
    file_cap = blade.arrays.list_arrays_space(type="file-system")
    object_cap = blade.arrays.list_arrays_space(type="object-store")
    capacity_info["total"] = total_cap.items[0].capacity
    capacity_info["aggregate"] = {
        "data_reduction": total_cap.items[0].space.data_reduction,
        "snapshots": total_cap.items[0].space.snapshots,
        "total_physical": total_cap.items[0].space.total_physical,
        "unique": total_cap.items[0].space.unique,
        "virtual": total_cap.items[0].space.virtual,
    }
    capacity_info["file-system"] = {
        "data_reduction": file_cap.items[0].space.data_reduction,
        "snapshots": file_cap.items[0].space.snapshots,
        "total_physical": file_cap.items[0].space.total_physical,
        "unique": file_cap.items[0].space.unique,
        "virtual": file_cap.items[0].space.virtual,
    }
    capacity_info["object-store"] = {
        "data_reduction": object_cap.items[0].space.data_reduction,
        "snapshots": object_cap.items[0].space.snapshots,
        "total_physical": object_cap.items[0].space.total_physical,
        "unique": object_cap.items[0].space.unique,
        "virtual": file_cap.items[0].space.virtual,
    }

    return capacity_info


def generate_snap_dict(blade):
    snap_info = {}
    snaps = blade.file_system_snapshots.list_file_system_snapshots()
    api_version = blade.api_version.list_versions().versions
    for snap in range(0, len(snaps.items)):
        snapshot = snaps.items[snap].name
        snap_info[snapshot] = {
            "destroyed": snaps.items[snap].destroyed,
            "source": snaps.items[snap].source,
            "suffix": snaps.items[snap].suffix,
            "source_destroyed": snaps.items[snap].source_destroyed,
        }
        if REPLICATION_API_VERSION in api_version:
            snap_info[snapshot]["owner"] = snaps.items[snap].owner.name
            snap_info[snapshot]["owner_destroyed"] = snaps.items[snap].owner_destroyed
            snap_info[snapshot]["source_display_name"] = snaps.items[
                snap
            ].source_display_name
            snap_info[snapshot]["source_is_local"] = snaps.items[snap].source_is_local
            snap_info[snapshot]["source_location"] = snaps.items[
                snap
            ].source_location.name
    return snap_info


def generate_snap_transfer_dict(blade):
    snap_transfer_info = {}
    snap_transfers = blade.file_system_snapshots.list_file_system_snapshots_transfer()
    for snap_transfer in range(0, len(snap_transfers.items)):
        transfer = snap_transfers.items[snap_transfer].name
        snap_transfer_info[transfer] = {
            "completed": snap_transfers.items[snap_transfer].completed,
            "data_transferred": snap_transfers.items[snap_transfer].data_transferred,
            "progress": snap_transfers.items[snap_transfer].progress,
            "direction": snap_transfers.items[snap_transfer].direction,
            "remote": snap_transfers.items[snap_transfer].remote.name,
            "remote_snapshot": snap_transfers.items[snap_transfer].remote_snapshot.name,
            "started": snap_transfers.items[snap_transfer].started,
            "status": snap_transfers.items[snap_transfer].status,
        }
    return snap_transfer_info


def generate_array_conn_dict(blade):
    array_conn_info = {}
    arrays = blade.array_connections.list_array_connections()
    for arraycnt in range(0, len(arrays.items)):
        array = arrays.items[arraycnt].remote.name
        array_conn_info[array] = {
            "encrypted": arrays.items[arraycnt].encrypted,
            "replication_addresses": arrays.items[arraycnt].replication_addresses,
            "management_address": arrays.items[arraycnt].management_address,
            "id": arrays.items[arraycnt].remote.id,
            "status": arrays.items[arraycnt].status,
            "version": arrays.items[arraycnt].version,
        }
        if arrays.items[arraycnt].encrypted:
            array_conn_info[array]["ca_certificate_group"] = arrays.items[
                arraycnt
            ].ca_certificate_group.name
    return array_conn_info


def generate_policies_dict(blade):
    policies_info = {}
    policies = blade.policies.list_policies()
    for policycnt in range(0, len(policies.items)):
        policy = policies.items[policycnt].name
        policies_info[policy] = {}
        policies_info[policy]["enabled"] = policies.items[policycnt].enabled
        if policies.items[policycnt].rules:
            policies_info[policy]["rules"] = (
                policies.items[policycnt].rules[0].to_dict()
            )
    return policies_info


def generate_bucket_dict(blade):
    bucket_info = {}
    buckets = blade.buckets.list_buckets()
    for bckt in range(0, len(buckets.items)):
        bucket = buckets.items[bckt].name
        bucket_info[bucket] = {
            "versioning": buckets.items[bckt].versioning,
            "object_count": buckets.items[bckt].object_count,
            "id": buckets.items[bckt].id,
            "account_name": buckets.items[bckt].account.name,
            "data_reduction": buckets.items[bckt].space.data_reduction,
            "snapshot_space": buckets.items[bckt].space.snapshots,
            "total_physical_space": buckets.items[bckt].space.total_physical,
            "unique_space": buckets.items[bckt].space.unique,
            "virtual_space": buckets.items[bckt].space.virtual,
            "created": buckets.items[bckt].created,
            "destroyed": buckets.items[bckt].destroyed,
            "time_remaining": buckets.items[bckt].time_remaining,
        }
    return bucket_info


def generate_kerb_dict(blade):
    kerb_info = {}
    keytabs = list(blade.get_keytabs().items)
    for ktab in range(0, len(keytabs)):
        keytab_name = keytabs[ktab].prefix
        kerb_info[keytab_name] = {}
        for key in range(0, len(keytabs)):
            if keytabs[key].prefix == keytab_name:
                kerb_info[keytab_name][keytabs[key].suffix] = {
                    "fqdn": keytabs[key].fqdn,
                    "kvno": keytabs[key].kvno,
                    "principal": keytabs[key].principal,
                    "realm": keytabs[key].realm,
                    "encryption_type": keytabs[key].encryption_type,
                }
    return kerb_info


def generate_ad_dict(blade):
    ad_info = {}
    active_directory = blade.get_active_directory()
    if active_directory.total_item_count != 0:
        ad_account = list(active_directory.items)[0]
        ad_info[ad_account.name] = {
            "computer": ad_account.computer_name,
            "domain": ad_account.domain,
            "directory_servers": ad_account.directory_servers,
            "kerberos_servers": ad_account.kerberos_servers,
            "service_principals": ad_account.service_principal_names,
            "join_ou": ad_account.join_ou,
            "encryption_types": ad_account.encryption_types,
        }
    return ad_info


def generate_object_store_accounts_dict(blade):
    account_info = {}
    accounts = list(blade.get_object_store_accounts().items)
    for account in range(0, len(accounts)):
        acc_name = accounts[account].name
        account_info[acc_name] = {
            "object_count": accounts[account].object_count,
            "data_reduction": accounts[account].space.data_reduction,
            "snapshots_space": accounts[account].space.snapshots,
            "total_physical_space": accounts[account].space.total_physical,
            "unique_space": accounts[account].space.unique,
            "virtual_space": accounts[account].space.virtual,
            "users": {},
        }
        acc_users = list(
            blade.get_object_store_users(filter='name="' + acc_name + '/*"').items
        )
        for acc_user in range(0, len(acc_users)):
            user_name = acc_users[acc_user].name.split("/")[1]
            account_info[acc_name]["users"][user_name] = {"keys": [], "policies": []}
            if (
                blade.get_object_store_access_keys(
                    filter='user.name="' + acc_users[acc_user].name + '"'
                ).total_item_count
                != 0
            ):
                access_keys = list(
                    blade.get_object_store_access_keys(
                        filter='user.name="' + acc_users[acc_user].name + '"'
                    ).items
                )
                for key in range(0, len(access_keys)):
                    account_info[acc_name]["users"][user_name]["keys"].append(
                        {
                            "name": access_keys[key].name,
                            "enabled": bool(access_keys[key].enabled),
                        }
                    )
            if (
                blade.get_object_store_access_policies_object_store_users(
                    member_names=[acc_users[acc_user].name]
                ).total_item_count
                != 0
            ):
                policies = list(
                    blade.get_object_store_access_policies_object_store_users(
                        member_names=[acc_users[acc_user].name]
                    ).items
                )
                for policy in range(0, len(policies)):
                    account_info[acc_name]["users"][user_name]["policies"].append(
                        policies[policy].policy.name
                    )
    return account_info


def generate_fs_dict(blade):
    fs_info = {}
    fsys = blade.file_systems.list_file_systems()
    for fsystem in range(0, len(fsys.items)):
        share = fsys.items[fsystem].name
        fs_info[share] = {
            "fast_remove": fsys.items[fsystem].fast_remove_directory_enabled,
            "snapshot_enabled": fsys.items[fsystem].snapshot_directory_enabled,
            "provisioned": fsys.items[fsystem].provisioned,
            "destroyed": fsys.items[fsystem].destroyed,
        }
        if fsys.items[fsystem].http.enabled:
            fs_info[share]["http"] = fsys.items[fsystem].http.enabled
        if fsys.items[fsystem].smb.enabled:
            fs_info[share]["smb_mode"] = fsys.items[fsystem].smb.acl_mode
        if fsys.items[fsystem].nfs.enabled:
            fs_info[share]["nfs_rules"] = fsys.items[fsystem].nfs.rules
        api_version = blade.api_version.list_versions().versions
        if MULTIPROTOCOL_API_VERSION in api_version:
            fs_info[share]["multi_protocol"] = {
                "safegaurd_acls": fsys.items[fsystem].multi_protocol.safeguard_acls,
                "access_control_style": fsys.items[
                    fsystem
                ].multi_protocol.access_control_style,
            }
        if HARD_LIMIT_API_VERSION in api_version:
            fs_info[share]["hard_limit"] = fsys.items[fsystem].hard_limit_enabled
        if REPLICATION_API_VERSION in api_version:
            fs_info[share]["promotion_status"] = fsys.items[fsystem].promotion_status
            fs_info[share]["requested_promotion_state"] = fsys.items[
                fsystem
            ].requested_promotion_state
            fs_info[share]["writable"] = fsys.items[fsystem].writable
            fs_info[share]["source"] = {
                "is_local": fsys.items[fsystem].source.is_local,
                "name": fsys.items[fsystem].source.name,
            }
    return fs_info


def main():
    argument_spec = purefb_argument_spec()
    argument_spec.update(
        dict(gather_subset=dict(default="minimum", type="list", elements="str"))
    )

    module = AnsibleModule(argument_spec, supports_check_mode=False)

    blade = get_blade(module)
    versions = blade.api_version.list_versions().versions

    if MIN_REQUIRED_API_VERSION not in versions:
        module.fail_json(
            msg="Minimum FlashBlade REST version required: {0}".format(
                MIN_REQUIRED_API_VERSION
            )
        )

    subset = [test.lower() for test in module.params["gather_subset"]]
    valid_subsets = (
        "all",
        "minimum",
        "config",
        "performance",
        "capacity",
        "network",
        "subnets",
        "lags",
        "filesystems",
        "snapshots",
        "buckets",
        "arrays",
        "replication",
        "policies",
        "accounts",
        "admins",
        "ad",
        "kerberos",
    )
    subset_test = (test in valid_subsets for test in subset)
    if not all(subset_test):
        module.fail_json(
            msg="value must gather_subset must be one or more of: %s, got: %s"
            % (",".join(valid_subsets), ",".join(subset))
        )

    info = {}

    if "minimum" in subset or "all" in subset:
        info["default"] = generate_default_dict(module, blade)
    if "performance" in subset or "all" in subset:
        info["performance"] = generate_perf_dict(blade)
    if "config" in subset or "all" in subset:
        info["config"] = generate_config_dict(blade)
    if "capacity" in subset or "all" in subset:
        info["capacity"] = generate_capacity_dict(blade)
    if "lags" in subset or "all" in subset:
        info["lag"] = generate_lag_dict(blade)
    if "network" in subset or "all" in subset:
        info["network"] = generate_network_dict(blade)
    if "subnets" in subset or "all" in subset:
        info["subnet"] = generate_subnet_dict(blade)
    if "filesystems" in subset or "all" in subset:
        info["filesystems"] = generate_fs_dict(blade)
    if "admins" in subset or "all" in subset:
        info["admins"] = generate_admin_dict(blade)
    if "snapshots" in subset or "all" in subset:
        info["snapshots"] = generate_snap_dict(blade)
    if "buckets" in subset or "all" in subset:
        info["buckets"] = generate_bucket_dict(blade)
    api_version = blade.api_version.list_versions().versions
    if POLICIES_API_VERSION in api_version:
        if "policies" in subset or "all" in subset:
            info["policies"] = generate_policies_dict(blade)
    if REPLICATION_API_VERSION in api_version:
        if "arrays" in subset or "all" in subset:
            info["arrays"] = generate_array_conn_dict(blade)
        if "replication" in subset or "all" in subset:
            info["file_replication"] = generate_file_repl_dict(blade)
            info["bucket_replication"] = generate_bucket_repl_dict(blade)
            info["snap_transfers"] = generate_snap_transfer_dict(blade)
            info["remote_credentials"] = generate_remote_creds_dict(blade)
            info["targets"] = generate_targets_dict(blade)
    if MIN_32_API in api_version:
        # Calls for data only available from Purity//FB 3.2 and higher
        blade = get_system(module)
        if "accounts" in subset or "all" in subset:
            info["accounts"] = generate_object_store_accounts_dict(blade)
        if "ad" in subset or "all" in subset:
            info["active_directory"] = generate_ad_dict(blade)
        if "kerberos" in subset or "all" in subset:
            info["kerberos"] = generate_kerb_dict(blade)

    module.exit_json(changed=False, purefb_info=info)


if __name__ == "__main__":
    main()
