#!/usr/bin/python
# -*- coding:utf-8 -*-

# Copyright(C) 2020 Inspur Inc. All Rights Reserved.
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import (absolute_import, division, print_function)

__metaclass__ = type

DOCUMENTATION = '''
---
module: user
version_added: "1.1.0"
author:
    - WangBaoshan (@ISIB-group)
short_description: Manage user.
description:
   - Manage user on Inspur server.
options:
    state:
        description:
            - Whether the user should exist or not, taking action if the state is different from what is stated.
        choices: ['present', 'absent']
        default: present
        type: str
    uname:
        description:
            - User name.
        type: str
        required: true
    upass:
        description:
            - User password.
        type: str
    role_id:
        description:
            - user group, default user group,'Administrator', 'Operator', 'Commonuser','OEM','NoAccess',
            - use command C(user_group_info) can get all group information.
        type: str
    priv:
        description:
            - User access, select one or more from None/KVM/VMM/SOL.
        choices: ['kvm', 'vmm', 'sol', 'none']
        type: list
        elements: str
extends_documentation_fragment:
    - inspur.sm.ism
'''

EXAMPLES = '''
- name: User test
  hosts: ism
  no_log: true
  connection: local
  gather_facts: no
  vars:
    ism:
      host: "{{ ansible_ssh_host }}"
      username: "{{ username }}"
      password: "{{ password }}"

  tasks:

  - name: "Add user"
    inspur.sm.add_user:
      state: "present"
      uname: "wbs"
      upass: "admin"
      role_id: "Administrator"
      priv: "kvm,sol"
      provider: "{{ ism }}"

  - name: "Set user"
    inspur.sm.add_user:
      state: "present"
      uname: "wbs"
      upass: "12345678"
      role_id: "user"
      priv: "kvm,sol"
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


class User(object):
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
        self.module.params['subcommand'] = 'edituser'
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
        uname=dict(type='str', required=True),
        upass=dict(type='str', required=False, no_log=True),
        role_id=dict(type='str', required=False),
        priv=dict(type='list', elements='str', required=False, choices=['kvm', 'vmm', 'sol', 'none']),
    )
    argument_spec.update(ism_argument_spec)
    user_obj = User(argument_spec)
    user_obj.work()


if __name__ == '__main__':
    main()
