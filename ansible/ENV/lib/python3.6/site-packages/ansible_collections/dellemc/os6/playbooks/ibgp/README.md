# Provision OS6 Switch Stack using the Ansible Network Collection for Dell EMC OS6

This example describes how to use Ansible to configure Dell EMC PowerSwitch platforms running Dell EMC OS6. The sample topology contains two OS6 switches connected with each other. This example configures iBGP between two routers using the same AS.

## Create a simple Ansible playbook

**1**. Create an inventory file called `inventory.yaml`, then specify the device IP addresses under use in the inventory.

**2**. Create a group variable file called `group_vars/all`, then define credentials common to all hosts.

**3**. Create a host variable file called `host_vars/switch1.yaml`, then define credentials, hostname for switch1.

**4**. Create a host variable file called `host_vars/switch2.yaml`, then define credentials and hostname for switch2.

**5**. Create a playbook called `os6switch.yaml`.

**6**. Run the playbook.

    ansible-playbook  -i  inventory.yaml  os6switch.yaml
    
(c) 2017-2020 Dell Inc. or its subsidiaries. All rights reserved.
