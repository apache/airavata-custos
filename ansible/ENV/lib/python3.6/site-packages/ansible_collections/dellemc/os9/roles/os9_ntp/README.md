NTP role
========

This role facilitates the configuration of network time protocol (NTP) attributes, and it specifically enables configuration of NTP server. This role is abstracted for Dell EMC PowerSwitch platforms running Dell EMC OS9.

The NTP role requires an SSH connection for connectivity to a Dell EMC OS9 device. You can use any of the built-in OS connection variables.


Role variables
--------------

- Role is abstracted using the `ansible_network_os` variable that can take `dellemc.os9.os9` as the value.
- If `os9_cfg_generate` is set to true, the variable generates the role configuration commands in a file
- Any role variable with a corresponding state variable set to absent negates the configuration of that variable
- Setting an empty value for any variable negates the corresponding configuration
- Variables and values are case-sensitive

**os9_ntp keys**

| Key        | Type                      | Description                                             | Support               |
|------------|---------------------------|---------------------------------------------------------|-----------------------|
| ``server`` | list | Configures the NTP server (see ``server.*``) | os9 |
| ``server.ip`` | string (required)         | Configures an IPv4 address for the NTP server (A.B.C.D format) | os9 |
| ``server.vrf`` | list | Configures the NTP server for VRF instance; list item contains the names of the VRF instance | os9 |
| ``server.state`` | string: absent,present\*     | Deletes the NTP server if set to absent                   | os9 |

> **NOTE**: Asterisk (\*) denotes the default value if none is specified. 

Connection variables
--------------------

Ansible Dell EMC network roles require connection information to establish communication with the nodes in your inventory. This information can exist in the Ansible *group_vars* or *host_vars* directories or inventory or in the playbook itself.

| Key         | Required | Choices    | Description                                           |
|-------------|----------|------------|-------------------------------------------------------|
| ``ansible_host`` | yes      |            | Specifies the hostname or address for connecting to the remote device over the specified transport |
| ``ansible_port`` | no       |            | Specifies the port used to build the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_PORT` option is used; it defaults to 22 |
| ``ansible_ssh_user`` | no       |            | Specifies the username that authenticates the CLI login for the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_USER` environment variable value is used  |
| ``ansible_ssh_pass`` | no       |            | Specifies the password that authenticates the connection to the remote device  |
| ``ansible_become`` | no       | yes, no\*   | Instructs the module to enter privileged mode on the remote device before sending any commands; if value is unspecified, the `ANSIBLE_BECOME` environment variable value is used, and the device attempts to execute all commands in non-privileged mode |
| ``ansible_become_method`` | no       | enable, sudo\*   | Instructs the module to allow the become method to be specified for handling privilege escalation; if value is unspecified, the `ANSIBLE_BECOME_METHOD` environment variable value is used |
| ``ansible_become_pass`` | no       |            | Specifies the password to use if required to enter privileged mode on the remote device; if ``ansible_become`` is set to no this key is not applicable |
| ``ansible_network_os`` | yes      | os9, null\*  | Loads the correct terminal and cliconf plugins to communicate with the remote device |

> **NOTE**: Asterisk (\*) denotes the default value if none is specified.

Example playbook
----------------

This example uses the *os9_ntp* role to set the NTP server, source ip, authentication and broadcast service. It creates a *hosts* file with the switch details and corresponding variables. The hosts file should define the `ansible_network_os` variable with corresponding Dell EMC OS9 name. 

When the `os9_cfg_generate` variable is set to true, it generates the configuration commands as a .part file in *build_dir* path. By default it is set to false. The example writes a simple playbook that only references the *os9_ntp* role. By including the role, you automatically get access to all of the tasks to configure NTP attributes.

**Sample hosts file**
 
    leaf1 ansible_host= <ip_address> 

**Sample host_vars/leaf1**

    host: leaf1
    ansible_ssh_user: xxxxx
    ansible_ssh_pass: xxxxx
    ansible_network_os: dellemc.os9.os9
    build_dir: ../temp/os9
	  
    os9_ntp:
      source: ethernet 1/1/2
      master: 5
      authenticate: true
      authentication_key:
        - key_num: 123
          key_string_type: 7
          key_string: test
          state: present
      trusted_key:
        - key_num: 1323
          state: present
      server:
        - ip: 2.2.2.2
          key: 345
          prefer: true
          state: present
      intf:
        ethernet 1/1/2:
          disable: true
          broadcast: true
 
**Simple playbook to setup NTP — leaf.yaml**

    - hosts: leaf1
      roles:
         - dellemc.os9.os9_ntp

**Run**

    ansible-playbook -i hosts leaf.yaml

(c) 2017-2020 Dell Inc. or its subsidiaries. All rights reserved.
