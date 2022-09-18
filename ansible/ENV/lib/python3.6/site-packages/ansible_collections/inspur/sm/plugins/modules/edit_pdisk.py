#!/usr/bin/python
# -*- coding:utf-8 -*-

# Copyright (C) 2020 Inspur Inc. All Rights Reserved.
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import (absolute_import, division, print_function)

__metaclass__ = type

DOCUMENTATION = '''
---
module: edit_pdisk
version_added: "0.1.0"
author:
    - WangBaoshan (@ISIB-group)
short_description: Set physical disk.
description:
   - Set physical disk on Inspur server.
options:
    info:
        description:
            - Show controller and pdisk info.
        choices: ['show']
        type: str
    ctrl_id:
        description:
            - Raid controller ID.
            - Required when I(Info=None).
        type: int
    device_id:
        description:
            - physical drive id.
            - Required when I(Info=None).
        type: int
    option:
        description:
            - Set operation options fo physical disk,
            - UG is Unconfigured Good,UB is Unconfigured Bad,
            - OFF is offline,FAIL is Failed,RBD is Rebuild,
            - ON is Online,JB is JBOD,ES is Drive Erase stop,
            - EM is Drive Erase Simple,EN is Drive Erase Normal,
            - ET is Drive Erase Through,LOC is Locate,STL is Stop Locate.
            - Required when I(Info=None).
        choices: ['UG', 'UB', 'OFF', 'FAIL', 'RBD', 'ON', 'JB', 'ES', 'EM', 'EN', 'ET', 'LOC', 'STL']
        type: str
extends_documentation_fragment:
    - inspur.sm.ism
'''

EXAMPLES = '''
- name: Edit pdisk test
  hosts: ism
  connection: local
  gather_facts: no
  vars:
    ism:
      host: "{{ ansible_ssh_host }}"
      username: "{{ username }}"
      password: "{{ password }}"

  tasks:

  - name: "Show pdisk information"
    inspur.sm.edit_pdisk:
      info: "show"
      provider: "{{ ism }}"

  - name: "Edit pdisk"
    inspur.sm.edit_pdisk:
      ctrl_id: 0
      device_id: 1
      option: "LOC"
      provider: "{{ ism }}"
'''

RETURN = '''
message:
    description: Messages returned after module execution.
    returned: always
    type: str
state:
    description: Status after module execution.
    returned: always
    type: str
changed:
    description: Check to see if a change was made on the device.
    returned: always
    type: bool
'''

from ansible.module_utils.basic import AnsibleModule
from ansible_collections.inspur.sm.plugins.module_utils.ism import (ism_argument_spec, get_connection)


class Disk(object):
    def __init__(self, argument_spec):
        self.spec = argument_spec
        self.module = None
        self.init_module()
        self.results = dict()

    def init_module(self):
        """Init module object"""

        self.module = AnsibleModule(
            argument_spec=self.spec, supports_check_mode=False)

    def run_command(self):
        self.module.params['subcommand'] = 'setpdisk'
        self.results = get_connection(self.module)
        if self.results['State'] == 'Success':
            self.results['changed'] = True

    def show_result(self):
        """Show result"""
        self.module.exit_json(**self.results)

    def work(self):
        """Worker"""
        self.run_command()
        self.show_result()


def main():
    argument_spec = dict(
        info=dict(type='str', required=False, choices=['show']),
        ctrl_id=dict(type='int', required=False),
        device_id=dict(type='int', required=False),
        option=dict(type='str', required=False, choices=['UG', 'UB', 'OFF', 'FAIL', 'RBD', 'ON', 'JB', 'ES', 'EM', 'EN', 'ET', 'LOC', 'STL']),
    )
    argument_spec.update(ism_argument_spec)
    disk_obj = Disk(argument_spec)
    disk_obj.work()


if __name__ == '__main__':
    main()
