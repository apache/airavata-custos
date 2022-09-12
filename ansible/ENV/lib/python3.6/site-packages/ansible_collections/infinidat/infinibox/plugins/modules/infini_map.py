#!/usr/bin/python
# -*- coding: utf-8 -*-
# Copyright: (c) 2020, Infinidat <info@infinidat.com>
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function
from infinisdk.core.exceptions import APICommandFailed
__metaclass__ = type


ANSIBLE_METADATA = {'metadata_version': '1.1',
                    'status': ['preview'],
                    'supported_by': 'community'}


DOCUMENTATION = r'''
---
module: infini_map
version_added: '2.10'
short_description: Create and Delete mapping of a volume to a host on Infinibox
description:
    - This module creates or deletes mappings of volumes to hosts on
      Infinibox. infini_map is implemented separately from infini_host to allow
      ansible plays to remove, or make absent, a mapping without removing the host.
author: David Ohlemacher (@ohlemacher)
options:
  host:
    description:
      - Host Name
    required: true
  volume:
    description:
      - Volume name to map to the host
    required: true
  state:
    description:
      - Creates mapping when present or removes when absent, or provides
        details of a mapping when stat.
    required: false
    default: present
    choices: [ "stat", "present", "absent" ]
extends_documentation_fragment:
    - infinibox
'''

EXAMPLES = r'''
- name: Map a volume to an existing host
  infini_map:
    host: foo.example.com
    volume: bar
    state: present  # Default
    user: admin
    password: secret
    system: ibox001

- name: Unmap volume bar from host foo.example.com
  infini_map:
    host: foo.example.com
    volume: bar
    state: absent
    system: ibox01
    user: admin
    password: secret

- name: Stat mapping of volume bar to host foo.example.com
  infini_map:
    host: foo.example.com
    volume: bar
    state: stat
    system: ibox01
    user: admin
    password: secret
'''

# RETURN = r''' # '''

from ansible.module_utils.basic import AnsibleModule, missing_required_lib
from ansible.module_utils.infinibox import \
    HAS_INFINISDK, api_wrapper, infinibox_argument_spec, \
    get_pool, get_system, get_volume, get_host, merge_two_dicts


def vol_is_mapped_to_host(volume, host):
    volume_fields = volume.get_fields()
    volume_id = volume_fields.get('id')
    host_luns = host.get_luns()
    #print('volume id: {0}'.format(volume_id))
    #print('host luns: {0}'.format(str(host_luns)))
    for lun in host_luns:
        if lun.volume == volume:
            #print('found mapped volume: {0}'.format(volume))
            return True
    return False


def find_lun_use(module, host, volume):
    check_result = {'lun_used': False, 'lun_volume_matches': False}
    desired_lun = module.params['lun']

    if desired_lun:
        for host_lun in host.get_luns():
            if desired_lun == host_lun.lun:
                if host_lun.volume == volume:
                    check_result = {'lun_used': True, 'lun_volume_matches': True}
                else:
                    check_result = {'lun_used': True, 'lun_volume_matches': False}

    return check_result


def find_lun(host, volume):
    found_lun = None
    luns = host.get_luns()

    for lun in luns:
        if lun.volume == volume:
            found_lun = lun.lun
    return found_lun


@api_wrapper
def create_mapping(module, system):
    """
    Create mapping of volume to host. If already mapped, exit_json with changed False.
    """
    changed = False

    host = system.hosts.get(name=module.params['host'])
    volume = get_volume(module, system)

    lun_use = find_lun_use(module, host, volume)
    if lun_use['lun_used']:
        #assert not lun_use['lun_volume_matches'], "Cannot have matching lun and volume in create_mapping()"
        msg = "Cannot create mapping of volume '{}' to host '{}' using lun '{}'. Lun in use.".format(
            volume.get_name(),
            host.get_name(),
            module.params['lun'])
        module.fail_json(msg=msg)

    try:
        desired_lun = module.params['lun']
        if not module.check_mode:
            host.map_volume(volume, lun=desired_lun)
        changed = True
    except APICommandFailed as err:
        if "is already mapped" not in str(err):
            module.fail_json('Cannot map volume {0} to host {1}: {2}'.format(
                module.params['volume'],
                module.params['host'],
                str(err)))

    return changed


@api_wrapper
def update_mapping(module, system):
    host = system.hosts.get(name=module.params['host'])
    volume = get_volume(module, system)
    desired_lun = module.params['lun']

    assert vol_is_mapped_to_host(volume, host)

    if desired_lun:
        found_lun = find_lun(host, volume)
        if found_lun != desired_lun:
            msg = "Cannot change the lun from '{}' to '{}' for existing mapping of volume '{}' to host '{}'".format(
                found_lun,
                desired_lun,
                volume.get_name(),
                host.get_name())
            module.fail_json(msg=msg)

    changed = False
    return changed


@api_wrapper
def delete_mapping(module, system):
    """
    Remove mapping of volume from host. If the either the volume or host
    do not exist, then there should be no mapping to unmap. If unmapping
    generates a key error with 'has no logical units' in its message, then
    the volume is not mapped.  Either case, return changed=False.
    """
    changed = False
    msg = ""

    if not module.check_mode:
        volume = get_volume(module, system)
        host = system.hosts.get(name=module.params['host'])

        if volume and host:
            try:
                existing_lun = find_lun(host, volume)
                host.unmap_volume(volume)
                changed = True
                msg = "Volume '{0}' was unmapped from host '{1}' freeing lun '{2}'".format(
                    module.params['volume'],
                    module.params['host'],
                    existing_lun,
                )
            except KeyError as err:
                if 'has no logical units' not in str(err):
                    module.fail_json('Cannot unmap volume {0} from host {1}: {2}'.format(
                        module.params['volume'],
                        module.params['host'],
                        str(err)))
                else:
                    msg = "Volume {0} was not mapped to host {1} and so unmapping was not executed".format(
                        module.params['volume'],
                        module.params['host'],
                    )
        else:
            msg = "Either volume {0} or host {1} does not exist. Unmapping was not executed".format(
                module.params['volume'],
                module.params['host'],
            )
    else:  # check_mode
        changed = True

    module.exit_json(msg=msg, changed=changed)


def get_sys_vol_host(module):
    system = get_system(module)
    volume = get_volume(module, system)
    host = get_host(module, system)
    return (system, volume, host)


def get_mapping_fields(volume, host):
    luns = host.get_luns()
    for lun in luns:
        if volume.get_name() == lun.volume.get_name():
            field_dict = dict(
                id=lun.id,
            )
            return field_dict
    assert False, 'Failed to find lun details from volume {0} and host {1}'.format(
        volume.get_name(), host.get_name())


def handle_stat(module):
    system, volume, host = get_sys_vol_host(module)
    volume_name = module.params['volume']
    host_name = module.params['host']
    if not volume:
        module.fail_json(msg='Volume {0} not found'.format(volume_name))
    if not host:
        module.fail_json(msg='Host {0} not found'.format(host_name))
    if not vol_is_mapped_to_host(volume, host):
        msg = 'Volume {0} is not mapped to host {1}'.format(volume_name, host_name)
        module.fail_json(msg=msg)

    found_lun = find_lun(host, volume)

    field_dict = get_mapping_fields(volume, host)
    result = dict(
        changed=False,
        volume_lun=found_lun,
        msg = 'Volume {0} is mapped to host {1}'.format(volume_name, host_name),
    )
    result = merge_two_dicts(result, field_dict)
    module.exit_json(**result)


def handle_present(module):
    system, volume, host= get_sys_vol_host(module)
    if not volume:
        module.fail_json(changed=False, msg='Volume {0} not found'.format(
            module.params['volume']))
    if not host:
        module.fail_json(changed=False, msg='Host {0} not found'.format(
            module.params['host']))
    if not vol_is_mapped_to_host(volume, host):
        changed = create_mapping(module, system)
        # TODO: Why is find_lun() returning None after creating the mapping?
        #       host.get_luns() returns an empty list, why?
        # existing_lun = find_lun(host, volume)
        # msg = "Volume '{0}' map to host '{1}' created using lun '{2}'".format(
        #     volume.get_name(),
        #     host.get_name(),
        #     existing_lun,
        # )
        msg = "Volume '{0}' map to host '{1}' created".format(
            volume.get_name(),
            host.get_name()
        )
    else:
        changed = update_mapping(module, system)
        existing_lun = find_lun(host, volume)
        msg = "Volume '{0}' map to host '{1}' already exists using lun '{2}'".format(
            volume.get_name(),
            host.get_name(),
            existing_lun,
        )

    result = dict(
        changed=changed,
        msg=msg,
    )
    module.exit_json(**result)


def handle_absent(module):
    system, volume, host = get_sys_vol_host(module)
    if not volume or not host:
        module.exit_json(changed=False, msg='Mapping of volume {0} to host {1} already absent'.format(
            module.params['volume'],
            module.params['host']))
    else:
        changed = delete_mapping(module, system)
        module.exit_json(changed=changed, msg="File system removed")


def execute_state(module):
    state = module.params['state']
    try:
        if state == 'stat':
            handle_stat(module)
        elif state == 'present':
            handle_present(module)
        elif state == 'absent':
            handle_absent(module)
        else:
            module.fail_json(msg='Internal handler error. Invalid state: {0}'.format(state))
    finally:
        system = get_system(module)
        system.logout()


def main():
    """
    Gather auguments and manage mapping of vols to hosts.
    """
    argument_spec = infinibox_argument_spec()
    argument_spec.update(
        dict(
            host=dict(required=True),
            state=dict(default='present', choices=['stat', 'present', 'absent']),
            volume=dict(required=True),
            lun=dict(type=int),
        )
    )

    module = AnsibleModule(argument_spec, supports_check_mode=True)

    if not HAS_INFINISDK:
        module.fail_json(msg=missing_required_lib('infinisdk'))

    execute_state(module)


if __name__ == '__main__':
    main()
