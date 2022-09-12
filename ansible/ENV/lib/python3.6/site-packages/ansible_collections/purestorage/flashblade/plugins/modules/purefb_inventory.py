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
module: purefb_inventory
version_added: '1.0.0'
short_description: Collect information from Pure Storage FlashBlade
description:
  - Collect information from a Pure Storage FlashBlade running the
    Purity//FB operating system. By default, the module will collect basic
    information including hosts, host groups, protection
    groups and volume counts. Additional information can be collected
    based on the configured set of arguements.
author:
  - Pure Storage ansible Team (@sdodsley) <pure-ansible-team@purestorage.com>
extends_documentation_fragment:
  - purestorage.flashblade.purestorage.fb
"""

EXAMPLES = r"""
- name: collect FlashBlade invenroty
  purefa_inventory:
    fa_url: 10.10.10.2
    api_token: e31060a7-21fc-e277-6240-25983c6c4592
- name: show default information
  debug:
    msg: "{{ array_info['purefb_info'] }}"

"""

RETURN = r"""
purefb_inventory:
  description: Returns the inventory information for the FlashArray
  returned: always
  type: complex
  sample: {
        "admins": {
            "pureuser": {
                "role": "array_admin",
                "type": "local"
            }
        },
        "apps": {
            "offload": {
                "description": "Snapshot offload to NFS or Amazon S3",
                "status": "healthy",
                "version": "5.2.1"
            }
        }
    }
"""


from ansible.module_utils.basic import AnsibleModule
from ansible_collections.purestorage.flashblade.plugins.module_utils.purefb import (
    get_blade,
    purefb_argument_spec,
)


def generate_hardware_dict(blade):
    hw_info = {
        "fans": {},
        "controllers": {},
        "blades": {},
        "chassis": {},
        "ethernet": {},
        "modules": {},
        "power": {},
        "switch": {},
    }
    components = blade.hardware.list_hardware(filter="type='fm'")
    for component in range(0, len(components.items)):
        component_name = components.items[component].name
        hw_info["modules"][component_name] = {
            "slot": components.items[component].slot,
            "status": components.items[component].status,
            "serial": components.items[component].serial,
            "model": components.items[component].model,
        }
    components = blade.hardware.list_hardware(filter="type='eth'")
    for component in range(0, len(components.items)):
        component_name = components.items[component].name
        hw_info["ethernet"][component_name] = {
            "slot": components.items[component].slot,
            "status": components.items[component].status,
            "serial": components.items[component].serial,
            "model": components.items[component].model,
            "speed": components.items[component].speed,
        }
    components = blade.hardware.list_hardware(filter="type='fan'")
    for component in range(0, len(components.items)):
        component_name = components.items[component].name
        hw_info["fans"][component_name] = {
            "slot": components.items[component].slot,
            "status": components.items[component].status,
        }
    components = blade.hardware.list_hardware(filter="type='fb'")
    for component in range(0, len(components.items)):
        component_name = components.items[component].name
        hw_info["blades"][component_name] = {
            "slot": components.items[component].slot,
            "status": components.items[component].status,
            "serial": components.items[component].serial,
            "model": components.items[component].model,
        }
    components = blade.hardware.list_hardware(filter="type='pwr'")
    for component in range(0, len(components.items)):
        component_name = components.items[component].name
        hw_info["power"][component_name] = {
            "slot": components.items[component].slot,
            "status": components.items[component].status,
            "serial": components.items[component].serial,
            "model": components.items[component].model,
        }
    components = blade.hardware.list_hardware(filter="type='xfm'")
    for component in range(0, len(components.items)):
        component_name = components.items[component].name
        hw_info["switch"][component_name] = {
            "slot": components.items[component].slot,
            "status": components.items[component].status,
            "serial": components.items[component].serial,
            "model": components.items[component].model,
        }
    components = blade.hardware.list_hardware(filter="type='ch'")
    for component in range(0, len(components.items)):
        component_name = components.items[component].name
        hw_info["chassis"][component_name] = {
            "slot": components.items[component].slot,
            "index": components.items[component].index,
            "status": components.items[component].status,
            "serial": components.items[component].serial,
            "model": components.items[component].model,
        }

    return hw_info


def main():
    argument_spec = purefb_argument_spec()

    module = AnsibleModule(argument_spec, supports_check_mode=True)
    blade = get_blade(module)

    module.exit_json(changed=False, purefb_info=generate_hardware_dict(blade))


if __name__ == "__main__":
    main()
