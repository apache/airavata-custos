#!/usr/bin/python
# -*- coding:utf-8 -*-

# Copyright(C) 2020 Inspur Inc. All Rights Reserved.
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import (absolute_import, division, print_function)

__metaclass__ = type

DOCUMENTATION = '''
---
module: user_group
version_added: "1.1.0"
author:
    - WangBaoshan (@ISIB-group)
short_description: Manage user group.
description:
   - Manage user group on Inspur server.
options:
    state:
        description:
            - Whether the user group should exist or not, taking action if the state is different from what is stated.
        choices: ['present', 'absent']
        default: present
        type: str
    name:
        description:
            - Group name.
        required: true
        type: str
    pri:
        description:
            - Group privilege.
            - Required when I(state=present).
        choices: ['administrator', 'operator', 'user', 'oem', 'none']
        type: str
extends_documentation_fragment:
    - inspur.sm.ism
'''

EXAMPLES = '''
- name: User group test
  hosts: ism
  connection: local
  gather_facts: no
  vars:
    ism:
      host: "{{ ansible_ssh_host }}"
      username: "{{ username }}"
      password: "{{ password }}"

  tasks:

  - name: "Add user group"
    inspur.sm.user_group:
      state: "present"
      name: "test"
      pri: "administrator"
      provider: "{{ ism }}"

  - name: "Set user group"
    inspur.sm.user_group:
      state: "present"
      name: "test"
      pri: "user"
      provider: "{{ ism }}"

  - name: "Delete user group"
    inspur.sm.user_group:
      state: "absent"
      name: "test"
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

from ansible_collections.inspur.sm.plugins.module_utils.ism import (ism_argument_spec, get_connection)
from ansible.module_utils.basic import AnsibleModule


class UserGroup(object):
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
        self.module.params['subcommand'] = 'editusergroup'
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
        state=dict(type='str', choices=['present', 'absent'], default='present'),
        name=dict(type='str', required=True),
        pri=dict(type='str', required=False, choices=['administrator', 'operator', 'user', 'oem', 'none']),
    )
    argument_spec.update(ism_argument_spec)
    usergroup_obj = UserGroup(argument_spec)
    usergroup_obj.work()


if __name__ == '__main__':
    main()
