# Ansible Network Collection for Dell EMC OS9

## Collection contents

This collection includes the Ansible modules, plugins and roles needed to provision and manage Dell EMC PowerSwitch platforms running Dell EMC OS9. Sample playbooks and documentation are also included to show how the collection can be used.

### Collection core modules

- **os9_command.py** — Run commands on devices running OS9

- **os9_config.py** — Manage configuration sections on devices running OS9
  
- **os9_facts.py** — Collect facts from devices running OS9

### Collection roles

These roles facilitate provisioning and administration of devices running Dell EMC OS9. There are over 22 roles available that provide a comprehensive coverage of most OS9 resources, including os9_aaa, os9_bgp and os9_ecmp. The documentation for each role is at [OS9 roles](https://github.com/ansible-collections/dellemc.os9/blob/master/docs/roles.rst).

### Sample use case playbooks

This collection includes the following sample playbooks that illustrate end to end use cases:

- [CLOS Fabric](https://github.com/ansible-collections/dellemc.os9/blob/master/playbooks/clos_fabric_ebgp/README.md) — Example playbook to build a Layer 3 Leaf-Spine fabric

## Installation

Use this command to install the latest version of the OS9 collection from Ansible Galaxy:

```
    ansible-galaxy collection install dellemc.os9

```

To install a specific version, a version range identifier must be specified. For example, to install the most recent version that is greater than or equal to 1.0.0 and less than 2.0.0:

```
    ansible-galaxy collection install 'dellemc.os9:>=1.0.0,<2.0.0'

```

## Version compatibility

* Ansible version 2.10 or higher
* Python 2.7 or higher and Python 3.5 or higher

> **NOTE**: For Ansible versions lower than 2.10, use the legacy [dellos9 modules](https://ansible-dellos-docs.readthedocs.io/en/latest/modules.html#os9-modules) and [dellos roles](https://ansible-dellos-docs.readthedocs.io/en/latest/roles.html).
       

## Sample playbook

**playbook.yaml**

```
- hosts: os9_switches
  connection: network_cli
  collections:
    - dellemc.os9
  roles:
    - os9_vlan
```

**host_vars/os9_sw1.yaml**

```
hostname: os9_sw1
# Parameters for connection type network_cli
ansible_ssh_user: xxxx
ansible_ssh_pass: xxxx
ansible_network_os: dellemc.os9.os9

# Create vlan100 and delete vlan888
os9_vlan:
    vlan 100:
      description: "Blue"
      state: present
    vlan 888:
      state: absent

```

**inventory.yaml**

```
[os9_sw1]
os9_sw1 ansible_host=100.104.28.119

[os9_sw2]
os9_sw2 ansible_host=100.104.28.118
    
[os9_switches:children]
os9_sw1
os9_sw2

```

(c) 2017-2020 Dell Inc. or its subsidiaries. All rights reserved.
