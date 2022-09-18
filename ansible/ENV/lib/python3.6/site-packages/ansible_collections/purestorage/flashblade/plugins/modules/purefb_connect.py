#!/usr/bin/python
# -*- coding: utf-8 -*-

# (c) 2020, Simon Dodsley (simon@purestorage.com)
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function

__metaclass__ = type

ANSIBLE_METADATA = {
    "metadata_version": "1.1",
    "status": ["preview"],
    "supported_by": "community",
}

DOCUMENTATION = r"""
---
module: purefb_connect
version_added: '1.0.0'
short_description: Manage replication connections between two FlashBlades
description:
- Manage replication connections to specified remote FlashBlade system
author:
- Pure Storage Ansible Team (@sdodsley) <pure-ansible-team@purestorage.com>
options:
  state:
    description:
    - Create or delete replication connection
    default: present
    type: str
    choices: [ absent, present ]
  encrypted:
    description:
    - Define if replication connection is encrypted
    type: bool
    default: False
  target_url:
    description:
    - Management IP address of target FlashBlade system
    type: str
    required: true
  target_api:
    description:
    - API token for target FlashBlade system
    type: str
extends_documentation_fragment:
- purestorage.flashblade.purestorage.fb
"""

EXAMPLES = r"""
- name: Create a connection to remote FlashBlade system
  purefb_connect:
    target_url: 10.10.10.20
    target_api: 9c0b56bc-f941-f7a6-9f85-dcc3e9a8f7d6
    fb_url: 10.10.10.2
    api_token: e31060a7-21fc-e277-6240-25983c6c4592
- name: Delete connection to target FlashBlade system
  purefb_connect:
    state: absent
    target_url: 10.10.10.20
    target_api: 9c0b56bc-f941-f7a6-9f85-dcc3e9a8f7d6
    fb_url: 10.10.10.2
    api_token: e31060a7-21fc-e277-6240-25983c6c4592
"""

RETURN = r"""
"""

HAS_PURITYFB = True
try:
    from purity_fb import PurityFb, ArrayConnection, ArrayConnectionPost
except ImportError:
    HAS_PURITYFB = False

from ansible.module_utils.basic import AnsibleModule
from ansible_collections.purestorage.flashblade.plugins.module_utils.purefb import (
    get_blade,
    purefb_argument_spec,
)


MIN_REQUIRED_API_VERSION = "1.9"


def _check_connected(module, blade):
    connected_blades = blade.array_connections.list_array_connections()
    for target in range(0, len(connected_blades.items)):
        if connected_blades.items[target].management_address is None:
            try:
                remote_system = PurityFb(module.params["target_url"])
                remote_system.login(module.params["target_api"])
                remote_array = remote_system.arrays.list_arrays().items[0].name
                if connected_blades.items[target].remote.name == remote_array:
                    return connected_blades.items[target]
            except Exception:
                module.fail_json(
                    msg="Failed to connect to remote array {0}.".format(
                        module.params["target_url"]
                    )
                )
        if connected_blades.items[target].management_address == module.params[
            "target_url"
        ] and connected_blades.items[target].status in [
            "connected",
            "connecting",
            "partially_connected",
        ]:
            return connected_blades.items[target]
    return None


def break_connection(module, blade, target_blade):
    """Break connection between arrays"""
    changed = True
    if not module.check_mode:
        source_blade = blade.arrays.list_arrays().items[0].name
        try:
            if target_blade.management_address is None:
                module.fail_json(
                    msg="Disconnect can only happen from the array that formed the connection"
                )
            blade.array_connections.delete_array_connections(
                remote_names=[target_blade.remote.name]
            )
        except Exception:
            module.fail_json(
                msg="Failed to disconnect {0} from {1}.".format(
                    target_blade.remote.name, source_blade
                )
            )
    module.exit_json(changed=changed)


def create_connection(module, blade):
    """Create connection between arrays"""
    changed = True
    if not module.check_mode:
        remote_array = module.params["target_url"]
        try:
            remote_system = PurityFb(module.params["target_url"])
            remote_system.login(module.params["target_api"])
            remote_array = remote_system.arrays.list_arrays().items[0].name
            remote_conn_cnt = (
                remote_system.array_connections.list_array_connections().pagination_info.total_item_count
            )
            # TODO: SD - Update with new max when fan-in/fan-out is enabled for FB
            if remote_conn_cnt == 1:
                module.fail_json(
                    msg="Remote array {0} already connected to another array. Fan-In not supported".format(
                        remote_array
                    )
                )
            connection_key = (
                remote_system.array_connections.create_array_connections_connection_keys()
                .items[0]
                .connection_key
            )
            remote_array = remote_system.arrays.list_arrays().items[0].name
            connection_info = ArrayConnectionPost(
                management_address=module.params["target_url"],
                encrypted=module.params["encrypted"],
                connection_key=connection_key,
            )
            blade.array_connections.create_array_connections(
                array_connection=connection_info
            )
        except Exception:
            module.fail_json(
                msg="Failed to connect to remote array {0}.".format(remote_array)
            )
    module.exit_json(changed=changed)


def update_connection(module, blade, target_blade):
    """Update array connection - only encryption currently"""
    changed = True
    if not module.check_mode:
        if target_blade.management_address is None:
            module.fail_json(
                msg="Update can only happen from the array that formed the connection"
            )
        if module.params["encrypted"] != target_blade.encrypted:
            if (
                module.params["encrypted"]
                and blade.file_system_replica_links.list_file_system_replica_links().pagination_info.total_item_count
                != 0
            ):
                module.fail_json(
                    msg="Cannot turn array connection encryption on if file system replica links exist"
                )
            new_attr = ArrayConnection(encrypted=module.params["encrypted"])
            changed = True
            if not module.check_mode:
                try:
                    blade.array_connections.update_array_connections(
                        remote_names=[target_blade.remote.name],
                        array_connection=new_attr,
                    )
                except Exception:
                    module.fail_json(
                        msg="Failed to change encryption setting for array connection."
                    )
        else:
            changed = False
    module.exit_json(changed=changed)


def main():
    argument_spec = purefb_argument_spec()
    argument_spec.update(
        dict(
            state=dict(type="str", default="present", choices=["absent", "present"]),
            encrypted=dict(type="bool", default=False),
            target_url=dict(type="str", required=True),
            target_api=dict(type="str", no_log=True),
        )
    )

    required_if = [("state", "present", ["target_api"])]

    module = AnsibleModule(
        argument_spec, required_if=required_if, supports_check_mode=True
    )

    if not HAS_PURITYFB:
        module.fail_json(msg="purity_fb sdk is required for this module")

    state = module.params["state"]
    blade = get_blade(module)
    versions = blade.api_version.list_versions().versions

    if MIN_REQUIRED_API_VERSION not in versions:
        module.fail_json(
            msg="Minimum FlashBlade REST version required: {0}".format(
                MIN_REQUIRED_API_VERSION
            )
        )

    target_blade = _check_connected(module, blade)
    if state == "present" and not target_blade:
        # TODO: SD - Update with new max when fan-out is supported
        if (
            blade.array_connections.list_array_connections().pagination_info.total_item_count
            == 1
        ):
            module.fail_json(
                msg="Source FlashBlade already connected to another array. Fan-Out not supported"
            )
        create_connection(module, blade)
    elif state == "present" and target_blade:
        update_connection(module, blade, target_blade)
    elif state == "absent" and target_blade:
        break_connection(module, blade, target_blade)

    module.exit_json(changed=False)


if __name__ == "__main__":
    main()
