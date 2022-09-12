ECMP role
=========

This role facilitates the configuration of equal cost multi-path (ECMP), and it supports the configuration of ECMP for IPv4. This role is abstracted for Dell EMC PowerSwitch platforms running Dell EMC OS9. 

The ECMP role requires an SSH connection for connectivity to a Dell EMC OS9 device. You can use any of the built-in OS connection variables.

Role variables
--------------

- Role is abstracted using the `ansible_network_os` variable that can take the dellemc.os9.os9 as a value
- If `os9_cfg_generate` is set to true, the variable generates the role configuration commands in a file
- Any role variable with a corresponding state variable set to absent negates the configuration of that variable
- Setting an empty value for any variable negates the corresponding configuration
- Variables and values are case-sensitive

**os9_ecmp keys**

| Key        | Type                      | Description                                             | Support               |
|------------|---------------------------|---------------------------------------------------------|-----------------------|
| ``weighted_ecmp`` | boolean: true,false           | Configures weighted ECMP | os9         |
| ``ecmp_group_max_paths`` | integer        | Configures the number of maximum-paths per ecmp-group                 | os9 |
| ``ecmp_group_path_fallback`` | boolean: true,false          | Configures ECMP group path management | os9  |
| ``ecmp <group id>`` | dictionary          | Configures ECMP group (see ``ecmp <group id>.*``) | os9 |
| ``ecmp <group id>.interface`` | list           | Configures interface into an ECMP group                        | os9 |
| ``ecmp <group id>.link_bundle_monitor`` | boolean: true,false           | Configures link-bundle monitoring   | os9 |
| ``ecmp <group id>.state`` | string: present\*,absent           | Deletes the ECMP group if set to absent           |  os9 |

Connection variables
--------------------

Ansible Dell EMC network roles require connection information to establish communication with the nodes in your inventory. This information can exist in the Ansible *group_vars* or *host_vars* directories or inventory, or in the playbook itself.

| Key         | Required | Choices    | Description                                         |
|-------------|----------|------------|-----------------------------------------------------|
| ``ansible_host`` | yes      |            | Specifies the hostname or address for connecting to the remote device over the specified transport |
| ``ansible_port`` | no       |            | Specifies the port used to build the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_PORT` option is used; it defaults to 22 |
| ``ansible_ssh_user`` | no       |            | Specifies the username that authenticates the CLI login for the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_USER` environment variable value is used  |
| ``ansible_ssh_pass`` | no       |            | Specifies the password that authenticates the connection to the remote device.  |
| ``ansible_become`` | no       | yes, no\*   | Instructs the module to enter privileged mode on the remote device before sending any commands; if value is unspecified, the `ANSIBLE_BECOME` environment variable value is used, and the device attempts to execute all commands in non-privileged mode |
| ``ansible_become_method`` | no       | enable, sudo\*   | Instructs the module to allow the become method to be specified for handling privilege escalation; if value is unspecified, the `ANSIBLE_BECOME_METHOD` environment variable value is used |
| ``ansible_become_pass`` | no       |            | Specifies the password to use if required to enter privileged mode on the remote device; if ``ansible_become`` is set to no this key is not applicable |
| ``ansible_network_os`` | yes      | os9, null\*  | Loads the correct terminal and cliconf plugins to communicate with the remote device |

> **NOTE**: Asterisk (\*) denotes the default value if none is specified.

Example playbook
----------------

This example uses the *os9_ecmp* role to configure ECMP for IPv4. The example creates a *hosts* file with the switch details and corresponding variables. The hosts file should define the `ansible_network_os` variable with the corresponding Dell EMC OS9 name.

When `os9_cfg_generate` is set to true, the variable generates the configuration commands as a .part file in *build_dir* path. By default, the variable is set to false. The example writes a simple playbook that only references the *os9_ecmp* role. The sample *host_vars* is provided for OS9 only.

**Sample hosts file**

    leaf1 ansible_host= <ip_address> 

**Sample host_vars/leaf1**

    hostname: leaf1
    ansible_become: yes
    ansible_become_method: xxxxx
    ansible_become_pass: xxxxx
    ansible_ssh_user: xxxxx
    ansible_ssh_pass: xxxxx
    ansible_network_os: dellemc.os9.os9
    build_dir: ../temp/os9
    os9_ecmp:
      ecmp 1:
        interface:
          - fortyGigE 1/49
          - fortyGigE 1/51
        link_bundle_monitor: true
        state: present
      weighted_ecmp: true
      ecmp_group_max_paths: 3
      ecmp_group_path_fallback: true
            
**Simple playbook to setup system — leaf.yaml**

    - hosts: leaf1
      roles:
         - dellemc.os9.os9_ecmp

**Run**

    ansible-playbook -i hosts leaf.yaml
    
(c) 2020 Dell Inc. or its subsidiaries. All rights reserved.
