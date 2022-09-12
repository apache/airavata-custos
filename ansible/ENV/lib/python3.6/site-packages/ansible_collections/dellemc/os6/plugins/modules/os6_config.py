#!/usr/bin/python
#
# (c) 2020 Peter Sprygada, <psprygada@ansible.com>
# Copyright (c) 2020 Dell Inc.
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function
__metaclass__ = type


ANSIBLE_METADATA = {'metadata_version': '1.1',
                    'status': ['preview'],
                    'supported_by': 'community'}

DOCUMENTATION = """
---
module: os6_config
author: "Abirami N (@abirami-n)"
short_description: Manage Dell EMC OS6 configuration sections
description:
  - OS6 configurations use a simple block indent file syntax
    for segmenting configuration into sections.  This module provides
    an implementation for working with OS6 configuration sections in
    a deterministic way.
extends_documentation_fragment: dellemc.os6.os6
options:
  lines:
    description:
      - The ordered set of commands that should be configured in the
        section.  The commands must be the exact same commands as found
        in the device running-config. Be sure to note the configuration
        command syntax as some commands are automatically modified by the
        device config parser. This argument is mutually exclusive with I(src).
    type: list
    aliases: ['commands']
  parents:
    description:
      - The ordered set of parents that uniquely identify the section or hierarchy
        the commands should be checked against.  If the parents argument
        is omitted, the commands are checked against the set of top
        level or global commands.
    type: list
  src:
    description:
      - Specifies the source path to the file that contains the configuration
        or configuration template to load.  The path to the source file can
        either be the full path on the Ansible control host or a relative
        path from the playbook or role root directory. This argument is
        mutually exclusive with I(lines).
    type: path
  before:
    description:
      - The ordered set of commands to push on to the command stack if
        a change needs to be made.  This allows the playbook designer
        the opportunity to perform configuration commands prior to pushing
        any changes without affecting how the set of commands are matched
        against the system.
    type: list
  after:
    description:
      - The ordered set of commands to append to the end of the command
        stack if a change needs to be made.  Just like with I(before) this
        allows the playbook designer to append a set of commands to be
        executed after the command set.
    type: list
  match:
    description:
      - Instructs the module on the way to perform the matching of
        the set of commands against the current device config.  If
        match is set to I(line), commands are matched line by line.  If
        match is set to I(strict), command lines are matched with respect
        to position.  If match is set to I(exact), command lines
        must be an equal match.  Finally, if match is set to I(none), the
        module will not attempt to compare the source configuration with
        the running configuration on the remote device.
    type: str
    default: line
    choices: ['line', 'strict', 'exact', 'none']
  replace:
    description:
      - Instructs the module on the way to perform the configuration
        on the device.  If the replace argument is set to I(line) then
        the modified lines are pushed to the device in configuration
        mode.  If the replace argument is set to I(block) then the entire
        command block is pushed to the device in configuration mode if any
        line is not correct.
    type: str
    default: line
    choices: ['line', 'block']
  update:
    description:
      - The I(update) argument controls how the configuration statements
        are processed on the remote device.  Valid choices for the I(update)
        argument are I(merge) and I(check).  When you set this argument to
        I(merge), the configuration changes merge with the current
        device running configuration.  When you set this argument to I(check)
        the configuration updates are determined but not actually configured
        on the remote device.
    type: str
    default: merge
    choices: ['merge', 'check']
  save:
    description:
      - The C(save) argument instructs the module to save the running-
        config to the startup-config at the conclusion of the module
        running.  If check mode is specified, this argument is ignored.
    type: bool
    default: 'no'
  config:
    description:
      - The module, by default, will connect to the remote device and
        retrieve the current running-config to use as a base for comparing
        against the contents of source.  There are times when it is not
        desirable to have the task get the current running-config for
        every task in a playbook.  The I(config) argument allows the
        implementer to pass in the configuration to use as the base
        config for comparison.
    type: str
  backup:
    description:
      - This argument will cause the module to create a full backup of
        the current C(running-config) from the remote device before any
        changes are made. If the C(backup_options) value is not given,
        the backup file is written to the C(backup) folder in the playbook
        root directory. If the directory does not exist, it is created.
    type: bool
    default: 'no'
  backup_options:
    description:
      - This is a dict object containing configurable options related to backup file path.
        The value of this option is read only when C(backup) is set to I(yes), if C(backup) is set
        to I(no) this option will be silently ignored.
    suboptions:
      filename:
        description:
          - The filename to be used to store the backup configuration. If the the filename
            is not given it will be generated based on the hostname, current time and date
            in format defined by <hostname>_config.<current-date>@<current-time>
        type: str
      dir_path:
        description:
          - This option provides the path ending with directory name in which the backup
            configuration file will be stored. If the directory does not exist it will be first
            created and the filename is either the value of C(filename) or default filename
            as described in C(filename) options description. If the path value is not given
            in that case a I(backup) directory will be created in the current working directory
            and backup configuration will be copied in C(filename) within I(backup) directory.
        type: path
    type: dict
"""

EXAMPLES = """
- os6_config:
    lines: ['hostname {{ inventory_hostname }}']
- os6_config:
    lines:
      - 10 permit ip 1.1.1.1 any log
      - 20 permit ip 2.2.2.2 any log
      - 30 permit ip 3.3.3.3 any log
      - 40 permit ip 4.4.4.4 any log
      - 50 permit ip  5.5.5.5 any log
    parents: ['ip access-list test']
    before: ['no ip access-list test']
    match: exact
- os6_config:
    lines:
      - 10 permit ip 1.1.1.1 any log
      - 20 permit ip 2.2.2.2 any log
      - 30 permit ip 3.3.3.3 any log
      - 40 permit ip 4.4.4.4 any log
    parents: ['ip access-list test']
    before: ['no ip access-list test']
    replace: block
- os6_config:
    lines: ['hostname {{ inventory_hostname }}']
    backup: yes
    backup_options:
      filename: backup.cfg
      dir_path: /home/user
"""

RETURN = """
updates:
  description: The set of commands that will be pushed to the remote device.
  returned: always
  type: list
  sample: ['interface Te1/0/1', 'no shutdown', 'exit']
commands:
  description: The set of commands that will be pushed to the remote device
  returned: always
  type: list
  sample: ['interface Te1/0/1', 'no shutdown', 'exit']
saved:
  description: Returns whether the configuration is saved to the startup
               configuration or not.
  returned: When not check_mode.
  type: bool
  sample: True
backup_path:
  description: The full path to the backup file
  returned: when backup is yes
  type: str
  sample: /playbooks/ansible/backup/os6_config.2017-07-16@22:28:34
"""
from ansible.module_utils.basic import AnsibleModule
from ansible_collections.dellemc.os6.plugins.module_utils.network.os6 import get_config, get_sublevel_config, NetworkConfig
from ansible_collections.dellemc.os6.plugins.module_utils.network.os6 import os6_argument_spec, check_args
from ansible_collections.dellemc.os6.plugins.module_utils.network.os6 import load_config, run_commands
from ansible_collections.dellemc.os6.plugins.module_utils.network.os6 import WARNING_PROMPTS_RE
from ansible_collections.ansible.netcommon.plugins.module_utils.network.common.config import dumps
import re
from ansible.module_utils.six import iteritems
from ansible.module_utils.connection import exec_command
from ansible.module_utils._text import to_bytes


def get_candidate(module):
    candidate = NetworkConfig(indent=0)
    banners = {}
    if module.params['src']:
        src, banners = extract_banners(module.params['src'])
        candidate.load(src)
    elif module.params['lines']:
        parents = module.params['parents'] or list()
        commands = module.params['lines'][0]
        if (isinstance(commands, dict)) and (isinstance(commands['command'], list)):
            candidate.add(commands['command'], parents=parents)
        elif (isinstance(commands, dict)) and (isinstance(commands['command'], str)):
            candidate.add([commands['command']], parents=parents)
        else:
            lines, banners = extract_banners(module.params['lines'])
            candidate.add(lines, parents=parents)
    return candidate, banners


def extract_banners(config):
    flag = False
    if isinstance(config, list):
        str1 = "\n"
        config = str1.join(config)
        flag = True
    banners = {}
    banner_cmds = re.findall(r'^banner (\w+)', config, re.M)
    for cmd in banner_cmds:
        regex = r'banner %s \"(.+?)\".*' % cmd
        match = re.search(regex, config, re.S)
        if match:
            key = 'banner %s' % cmd
            banners[key] = match.group(1).strip()

    for cmd in banner_cmds:
        regex = r'banner %s \"(.+?)\".*' % cmd
        match = re.search(regex, config, re.S)
        if match:
            config = config.replace(str(match.group(1)), '')
    config = re.sub(r'banner \w+ \"\"', '', config)
    if flag:
        config = config.split("\n")
    return (config, banners)


def diff_banners(want, have):
    candidate = {}
    for key, value in iteritems(want):
        if value != have.get(key):
            candidate[key] = value
    return candidate


def get_running_config(module):
    contents = module.params['config']
    if not contents:
        contents = get_config(module)
    contents, banners = extract_banners(contents)
    return contents, banners


def load_banners(module, banners):
    result_banners = []
    exec_command(module, 'configure terminal')
    for each in banners:
        delimiter = '"'
        cmdline = ""
        for key, value in each.items():
            cmdline = key + " " + delimiter + value + delimiter
            for cmd in cmdline.split("\n"):
                rc, out, err = exec_command(module, module.jsonify({'command': cmd, 'sendonly': True}))
            result_banners.append(cmdline)
    exec_command(module, 'end')
    return result_banners


def main():
    backup_spec = dict(
        filename=dict(),
        dir_path=dict(type='path')
    )
    argument_spec = dict(
        lines=dict(aliases=['commands'], type='list'),
        parents=dict(type='list'),

        src=dict(type='path'),

        before=dict(type='list'),
        after=dict(type='list'),

        match=dict(default='line',
                   choices=['line', 'strict', 'exact', 'none']),
        replace=dict(default='line', choices=['line', 'block']),

        update=dict(choices=['merge', 'check'], default='merge'),
        save=dict(type='bool', default=False),
        config=dict(),
        backup=dict(type='bool', default=False),
        backup_options=dict(type='dict', options=backup_spec)
    )

    argument_spec.update(os6_argument_spec)
    mutually_exclusive = [('lines', 'src'),
                          ('parents', 'src')]

    module = AnsibleModule(argument_spec=argument_spec,
                           mutually_exclusive=mutually_exclusive,
                           supports_check_mode=True)

    parents = module.params['parents'] or list()

    match = module.params['match']
    replace = module.params['replace']

    warnings = list()
    check_args(module, warnings)
    result = dict(changed=False, saved=False, warnings=warnings)

    candidate, want_banners = get_candidate(module)
    if module.params['backup']:
        if not module.check_mode:
            result['__backup__'] = get_config(module)

    commands = list()
    if any((module.params['lines'], module.params['src'])):
        if match != 'none':
            config, have_banners = get_running_config(module)
            config = NetworkConfig(contents=config, indent=0)
            if parents:
                config = get_sublevel_config(config, module)
            configobjs = candidate.difference(config, match=match, replace=replace)
        else:
            configobjs = candidate.items
            have_banners = {}
        diffbanners = diff_banners(want_banners, have_banners)
        banners = list()
        if diffbanners:
            banners.append(diffbanners)
        if configobjs or banners:
            commands = dumps(configobjs, 'commands')
            if ((isinstance(module.params['lines'], list)) and
                    (isinstance(module.params['lines'][0], dict)) and
                    set(['prompt', 'answer']).issubset(module.params['lines'][0])):
                cmd = {'command': commands,
                       'prompt': module.params['lines'][0]['prompt'],
                       'answer': module.params['lines'][0]['answer']}
                commands = [module.jsonify(cmd)]
            else:
                if commands:
                    commands = commands.split('\n')

            if module.params['before']:
                commands[:0], before_banners = extract_banners(module.params['before'])
                if before_banners:
                    banners.insert(0, before_banners)

            if module.params['after']:
                commands_after, after_banners = extract_banners(module.params['after'])
                commands.extend(commands_after)
                if after_banners:
                    banners.insert(len(banners), after_banners)

            if not module.check_mode and module.params['update'] == 'merge':
                if commands:
                    load_config(module, commands)
                if banners:
                    result_banners = load_banners(module, banners)
                else:
                    result_banners = []

            result['changed'] = True
            result['commands'] = commands
            result['updates'] = commands if commands else []
            result['banners'] = result_banners
            if result['banners']:
                result['updates'].extend(result_banners)

    if module.params['save']:
        result['changed'] = True
        if not module.check_mode:
            cmd = {'command': 'copy running-config startup-config',
                   'prompt': r'\(y/n\)\s?$', 'answer': 'y'}
            run_commands(module, [cmd])
            result['saved'] = True
        else:
            module.warn('Skipping command `copy running-config startup-config`'
                        'due to check_mode.  Configuration not copied to '
                        'non-volatile storage')

    module.exit_json(**result)


if __name__ == '__main__':
    main()
