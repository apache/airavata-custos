# Ansible Network Collection for Dell EMC OS6

## Collection contents

This collection includes the Ansible modules, plugins and roles needed to privision and manage Dell EMC PowerSwitch platforms running Dell EMC OS6. Sample playbooks and documentation are also included to show how the collection can be used.

### Collection core modules

- **os6_command.py** — Run commands on devices running OS6

- **os6_config.py** — Manage configuration on devices running OS6

- **os6_facts.py** — Collect facts from devices running OS6

### Collection roles

These roles facilitate provisioning and administration of devices running Dell EMC OS6. There are over 15 roles available that provide a comprehensive coverage of most OS6 resources, including os6_interface, os6_aaa, os6_bgp, and os6_xstp. The documentation for each role is at [OS6 roles](https://github.com/ansible-collections/dellemc.os6/blob/master/docs/roles.rst).

### Sample use case playbooks

This collection inlcudes the following sample playbook that illustrate end to end use cases:

  - [iBGP](https://github.com/ansible-collections/dellemc.os6/blob/master/playbooks/ibgp/README.md) — Example playbook to configure iBGP between two routers

## Installation

Use this command to install the latest version of the OS6 collection from Ansible Galaxy:

```
ansible-galaxy collection install dellemc.os6

```
To install a specific version, a version range identifier must be specified. For example, to install the most recent version that is greater than or equal to 1.0.0 and less than 2.0.0:

```
ansible-galaxy collection install 'dellemc.os6:>=1.0.0,<2.0.0'

```

## Version compatibility

* Ansible version 2.10 or higher
* Python 2.7 or higher and Python 3.5 or higher

> **NOTE**: For Ansible versions lower than 2.10, use the legacy [dellos6 modules](https://ansible-dellos-docs.readthedocs.io/en/latest/modules.html#os6-modules) and [dellos roles](https://ansible-dellos-docs.readthedocs.io/en/latest/roles.html).

## Sample playbook

**playbook.yaml**

```
- hosts: os6_switches
  connection: network_cli
  collections:
    - dellemc.os6
  roles:
    - os6_vlan

```

**host_vars/os6_sw1.yaml**

```
hostname: os6_sw1
# Parameters for connection type network_cli
ansible_ssh_user: xxxx
ansible_ssh_pass: xxxx
ansible_become: yes
ansible_become_method: enable
ansible_network_os: dellemc.os6.os6

# Create vlan100 and delete vlan888
os6_vlan:
    vlan 100:
      name: "Blue"
      state: present
    vlan 888:
      state: absent


```

**inventory.yaml**

```
[os6_sw1]
os6_sw1 ansible_host= 100.94.51.40

[os6_sw2]
os6_sw2 ansible_host= 100.94.52.38

[os6_switches:children]
os6_sw1
os6_sw2

```

(c) 2017-2020 Dell Inc. or its subsidiaries. All rights reserved.
