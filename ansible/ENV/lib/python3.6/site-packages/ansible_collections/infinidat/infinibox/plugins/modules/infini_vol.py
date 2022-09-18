#!/usr/bin/python
# -*- coding: utf-8 -*-
# Copyright: (c) 2020, Infinidat <info@infinidat.com>
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function
__metaclass__ = type


ANSIBLE_METADATA = {'metadata_version': '1.1',
                    'status': ['preview'],
                    'supported_by': 'community'}


DOCUMENTATION = r'''
---
module: infini_vol
version_added: 2.3
short_description:  Create, Delete or Modify volumes on Infinibox
description:
    - This module creates, deletes or modifies a volume on Infinibox.
author: Gregory Shulov (@GR360RY)
options:
  name:
    description:
      - Volume Name
    required: true
  state:
    description:
      - Creates/Modifies master volume or snapshot when present or removes when absent.
    required: false
    default: present
    choices: [ "stat", "present", "absent" ]
  size:
    description:
      - Volume size in MB, GB or TB units.  Required for creating a master volume, but not a snapshot
    required: false
  snapshot_lock_expires_at:
    description:
      - This will cause a snapshot to be locked at the specified date-time.
        Uses python's datetime format YYYY-mm-dd HH:MM:SS.ffffff, e.g. 2020-02-13 16:21:59.699700
    required: false
  snapshot_lock_only:
    description:
      - This will lock an existing snapshot but will suppress refreshing the snapshot.
    type: bool
    required: false
    default: false
  thin_provision:
    description:
      - Whether the master volume should be thin provisioned.  Required for creating a master volume, but not a snapshot.
    type: bool
    required: false
    default: true
    version_added: '2.8'
  pool:
    description:
      - Pool that master volume will reside within.  Required for creating a master volume, but not a snapshot.
    required: false
  volume_type:
    description:
      - Specifies the volume type, regular volume or snapshot.
    required: false
    default: master
    choices: [ "master", "snapshot" ]
  parent_volume_name:
    description:
      - Specify a volume name. This is the volume parent for creating a snapshot. Required if volume_type is snapshot.
    required: false
extends_documentation_fragment:
    - infinibox
requirements:
    - capacity
'''

EXAMPLES = r'''
- name: Create new volume named foo under pool named bar
  infini_vol:
    name: foo
    # volume_type: master  # Default
    size: 1TB
    thin_provision: yes
    pool: bar
    state: present
    user: admin
    password: secret
    system: ibox001
- name: Create snapshot named foo_snap from volume named foo
  infini_vol:
    name: foo_snap
    volume_type: snapshot
    parent_volume_name: foo
    state: present
    user: admin
    password: secret
    system: ibox001
- name: Stat snapshot, also a volume, named foo_snap
  infini_vol:
    name: foo_snap
    state: present
    user: admin
    password: secret
    system: ibox001
- name: Remove snapshot, also a volume, named foo_snap
  infini_vol:
    name: foo_snap
    state: absent
    user: admin
    password: secret
    system: ibox001
'''

# RETURN = r''' # '''

try:
    from capacity import KiB, Capacity
    HAS_CAPACITY = True
except ImportError:
    HAS_CAPACITY = False

import arrow
from ansible.module_utils.basic import AnsibleModule, missing_required_lib
from ansible.module_utils.infinibox import \
    HAS_INFINISDK, api_wrapper, infinibox_argument_spec, ObjectNotFound, \
    get_pool, get_system, get_volume


@api_wrapper
def create_volume(module, system):
    """Create Volume"""
    if not module.check_mode:
        if module.params['thin_provision']:
            prov_type = 'THIN'
        else:
            prov_type = 'THICK'
        pool = get_pool(module, system)
        volume = system.volumes.create(name=module.params['name'], provtype=prov_type, pool=pool)

        if module.params['size']:
            size = Capacity(module.params['size']).roundup(64 * KiB)
            volume.update_size(size)
    changed = True
    return changed


@api_wrapper
def update_volume(module, volume):
    """Update Volume"""
    changed = False
    if module.params['size']:
        size = Capacity(module.params['size']).roundup(64 * KiB)
        if volume.get_size() != size:
            if not module.check_mode:
                volume.update_size(size)
            changed = True
    if module.params['thin_provision'] is not None:
        type = str(volume.get_provisioning())
        if type == 'THICK' and module.params['thin_provision']:
            if not module.check_mode:
                volume.update_provisioning('THIN')
            changed = True
        if type == 'THIN' and not module.params['thin_provision']:
            if not module.check_mode:
                volume.update_provisioning('THICK')
            changed = True
    return changed


@api_wrapper
def delete_volume(module, volume):
    """ Delete Volume. Volume could be a snapshot."""
    if not module.check_mode:
        volume.delete()
    changed = True
    return True


@api_wrapper
def create_snapshot(module, system):
    """Create Snapshot from parent volume"""
    snapshot_name = module.params['name']
    parent_volume_name = module.params['parent_volume_name']
    try:
        parent_volume = system.volumes.get(name=parent_volume_name)
    except ObjectNotFound as err:
        msg = 'Cannot create snapshot {}. Parent volume {} not found'.format(
            snapshot_name,
            parent_volume_name)
        module.fail_json(msg=msg)
    if not parent_volume:
        msg = "Cannot find new snapshot's parent volume named {}".format(parent_volume_name)
        module.fail_json(msg=msg)
    if not module.check_mode:
        if module.params['snapshot_lock_only']:
            msg = "Snapshot does not exist. Cannot comply with 'snapshot_lock_only: true'."
            module.fail_json(msg=msg)
        check_snapshot_lock_options(module)
        snapshot = parent_volume.create_snapshot(name=snapshot_name)
    manage_snapshot_locks(module, snapshot)
    changed = True
    return changed


@api_wrapper
def update_snapshot(module, snapshot):
    """
    Update/refresh snapshot. May also lock it.
    """
    refresh_changed = False
    if not module.params['snapshot_lock_only']:
        snap_is_locked = snapshot.get_lock_state() == "LOCKED"
        if not snap_is_locked:
            if not module.check_mode:
                snapshot.refresh_snapshot()
            refresh_changed = True
        else:
            msg = "Snapshot is locked and may not be refreshed"
            module.fail_json(msg=msg)

    check_snapshot_lock_options(module)
    lock_changed = manage_snapshot_locks(module, snapshot)

    return refresh_changed or lock_changed


def get_sys_pool_vol_parname(module):
    system = get_system(module)
    pool = get_pool(module, system)
    volume = get_volume(module, system)
    parname = module.params['parent_volume_name']
    return (system, pool, volume, parname)


def check_snapshot_lock_options(module):
    """
    Check if specified options are feasible for a snapshot.

    Prevent very long lock times.
    max_delta_minutes limits locks to 30 days (43200 minutes).

    This functionality is broken out from manage_snapshot_locks() to allow
    it to be called by create_snapshot() before the snapshot is actually
    created.
    """
    snapshot_lock_expires_at = module.params['snapshot_lock_expires_at']

    if snapshot_lock_expires_at:  # Then user has specified wish to lock snap
        lock_expires_at = arrow.get(snapshot_lock_expires_at)

        # Check for lock in the past
        now = arrow.utcnow()
        if lock_expires_at <= now:
            msg =  "Cannot lock snapshot with a snapshot_lock_expires_at "
            msg += "of '{}' from the past".format(snapshot_lock_expires_at)
            module.fail_json(msg=msg)

        # Check for lock later than max lock, i.e. too far in future.
        max_delta_minutes = 43200  # 30 days in minutes
        max_lock_expires_at = now.shift(minutes=max_delta_minutes)
        if lock_expires_at >= max_lock_expires_at:
            msg = "snapshot_lock_expires_at exceeds {} days in the future".format(
                max_delta_minutes//24//60)
            module.fail_json(msg=msg)


def manage_snapshot_locks(module, snapshot):
    """
    Manage the locking of a snapshot. Check for bad lock times.
    See check_snapshot_lock_options() which has additional checks.
    """
    name = module.params["name"]
    snapshot_lock_expires_at = module.params['snapshot_lock_expires_at']
    snap_is_locked = snapshot.get_lock_state() == "LOCKED"
    current_lock_expires_at = snapshot.get_lock_expires_at()
    changed = False

    check_snapshot_lock_options(module)

    if snapshot_lock_expires_at:  # Then user has specified wish to lock snap
        lock_expires_at = arrow.get(snapshot_lock_expires_at)
        if snap_is_locked and lock_expires_at < current_lock_expires_at:
            # Lock earlier than current lock
            msg = "snapshot_lock_expires_at '{}' preceeds the current lock time of '{}'".format(
                lock_expires_at,
                current_lock_expires_at)
            module.fail_json(msg=msg)
        elif snap_is_locked and lock_expires_at == current_lock_expires_at:
            # Lock already set to correct time
            pass
        else:
            # Set lock
            if not module.check_mode:
                snapshot.update_lock_expires_at(lock_expires_at)
            changed = True
    return changed


def handle_stat(module):
    system, pool, volume, parname = get_sys_pool_vol_parname(module)
    if not volume:
        msg = "Volume {} not found. Cannot stat.".format(module.params['name'])
        module.fail_json(msg=msg)
    fields = volume.get_fields() #from_cache=True, raw_value=True)
    created_at = str(fields.get('created_at', None))
    has_children = fields.get('has_children', None)
    lock_expires_at= str(volume.get_lock_expires_at())
    lock_state = volume.get_lock_state()
    mapped = str(fields.get('mapped', None))
    name = fields.get('name', None)
    parent_id = fields.get('parent_id', None)
    size = str(volume.get_size())
    updated_at = str(fields.get('updated_at', None))
    used = str(fields.get('used_size', None))
    volume_id = fields.get('id', None)
    volume_type = fields.get('type', None)
    if volume_type == 'SNAPSHOT':
        msg = 'Snapshot stat found'
    else:
        msg = 'Volume stat found'

    result = dict(
        changed=False,
        created_at=created_at,
        has_children=has_children,
        lock_expires_at=lock_expires_at,
        lock_state=lock_state,
        mapped=mapped,
        msg=msg,
        parent_id=parent_id,
        size=size,
        updated_at=updated_at,
        used=used,
        volume_id=volume_id,
        volume_type=volume_type,
    )
    module.exit_json(**result)


def handle_present(module):
    system, pool, volume, parname = get_sys_pool_vol_parname(module)
    volume_type = module.params['volume_type']
    if volume_type == 'master':
        if not volume:
            changed = create_volume(module, system)
            module.exit_json(changed=changed, msg="Volume created")
        else:
            changed = update_volume(module, volume)
            module.exit_json(changed=changed, msg="Volume updated")
    elif volume_type == 'snapshot':
        snapshot = volume
        if not snapshot:
            changed = create_snapshot(module, system)
            module.exit_json(changed=changed, msg="Snapshot created")
        else:
            changed = update_snapshot(module, snapshot)
            module.exit_json(changed=changed, msg="Snapshot updated")
    else:
        module.fail_json(msg='A programming error has occurred')


def handle_absent(module):
    system, pool, volume, parname = get_sys_pool_vol_parname(module)
    volume_type = module.params['volume_type']

    if volume and volume.get_lock_state() == "LOCKED":
        msg = "Cannot delete snapshot. Locked."
        module.fail_json(msg=msg)

    if volume_type == 'master':
        if not volume:
            module.exit_json(changed=False, msg="Volume already absent")
        else:
            changed = delete_volume(module, volume)
            module.exit_json(changed=changed, msg="Volume removed")
    elif volume_type == 'snapshot':
        if not volume:
            module.exit_json(changed=False, msg="Snapshot already absent")
        else:
            snapshot = volume
            changed = delete_volume(module, snapshot)
            module.exit_json(changed=changed, msg="Snapshot removed")
    else:
        module.fail_json(msg='A programming error has occured')


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


def check_options(module):
    """Verify module options are sane"""
    state               = module.params['state']
    size                = module.params['size']
    pool                = module.params['pool']
    volume_type         = module.params['volume_type']
    parent_volume_name  = module.params['parent_volume_name']
    if state == 'present':
        if volume_type == 'master':
            if state == 'present':
                if parent_volume_name:
                    msg =  "parent_volume_name should not be specified "
                    msg += "if volume_type is 'volume'. Snapshots only."
                    module.fail_json(msg=msg)
                if not size:
                    msg = "Size is required to create a volume"
                    module.fail_json(msg=msg)
        elif volume_type == "snapshot":
            if size or pool:
                msg =  "Neither pool nor size should not be specified "
                msg += "for volume_type snapshot"
                module.fail_json(msg=msg)
            if state == "present":
                if not parent_volume_name:
                    msg =  "For state 'present' and volume_type 'snapshot', "
                    msg += "parent_volume_name is required"
                    module.fail_json(msg=msg)
        else:
            msg = "A programming error has occurred"
            module.fail_json(msg=msg)


def main():
    argument_spec = infinibox_argument_spec()
    argument_spec.update(
        dict(
            name=dict(required=True),
            parent_volume_name=dict(required=False),
            pool=dict(required=False),
            size=dict(),
            snapshot_lock_expires_at=dict(),
            snapshot_lock_only=dict(type='bool', default=False),
            state=dict(default='present', choices=['stat', 'present', 'absent']),
            thin_provision=dict(type='bool', default=True),
            volume_type=dict(default='master', choices=['master', 'snapshot']),
        )
    )

    module = AnsibleModule(argument_spec, supports_check_mode=True)

    if not HAS_INFINISDK:
        module.fail_json(msg=missing_required_lib('infinisdk'))

    if module.params['size']:
        try:
            Capacity(module.params['size'])
        except Exception:
            module.fail_json(msg='size (Physical Capacity) should be defined in MB, GB, TB or PB units')

    check_options(module)
    execute_state(module)


if __name__ == '__main__':
    main()
