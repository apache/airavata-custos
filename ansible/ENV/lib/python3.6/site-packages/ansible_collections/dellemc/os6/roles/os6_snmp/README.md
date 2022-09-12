SNMP role
=========

This role facilitates the configuration of global SNMP attributes. It supports the configuration of SNMP server attributes including users, group, community, location, and traps. This role is abstracted for Dell EMC PowerSwitch platforms running Dell EMC OS6.

The SNMP role requires an SSH connection for connectivity to a Dell EMC OS6 device. You can use any of the built-in OS connection variables.


Role variables
--------------

- Role is abstracted using the `ansible_network_os` variable that can take `dellemc.os6.os6` as a value
- If `os6_cfg_generate` is set to true, the variable generates the role configuration commands in a file
- Any role variable with a corresponding state variable set to absent negates the configuration of that variable
- Setting an empty value for any variable negates the corresponding configuration
- Variables and values are case-sensitive

**os6_snmp keys**

| Key        | Type                      | Description                                             | Support               |
|------------|---------------------------|---------------------------------------------------------|-----------------------|
| ``snmp_contact`` | string | Configures SNMP contact information, field needs to be left blank to remove the contact information  | os6 |
| ``snmp_location`` | string | Configures SNMP location information, field needs to be left blank to remove the location | os6 |
| ``snmp_community`` | list | Configures SNMP community information (see ``snmp_community.*``) | os6 |
| ``snmp_community.name`` | string (required)         | Configures the SNMP community string | os6 |
| ``snmp_community.access_mode`` | string: ro,rw           | Configures access-mode for the community | os6 |
| ``snmp_community.state`` | string: absent,present\*   | Deletes the SNMP community information if set to absent | os6 |
| ``snmp_host`` | list | Configures SNMP hosts to receive SNMP traps (see ``snmp_host.*``) | os6 |
| ``snmp_host.ip`` | string | Configures the IP address of the SNMP trap host | os6  |
| ``snmp_host.communitystring`` | string | Configures the SNMP community string of the trap host | os6 |
| ``snmp_host.udpport`` | string | Configures the UDP number of the SNMP trap host (0 to 65535) | os6 |
| ``snmp_host.state`` | string: absent,present\* | Deletes the SNMP trap host if set to absent | os6 |
| ``snmp_traps`` | list | Configures SNMP traps (see ``snmp_traps.*``) | os6  |
| ``snmp_traps.name`` | string | Enables SNMP traps   | os6 |
| ``snmp_traps.state`` | string: absent,present\* | Deletes the SNMP trap if set to absent | os6 |

> **NOTE**: Asterisk (\*) denotes the default value if none is specified. 

Connection variables
--------------------

Ansible Dell EMC Networking roles require connection information to establish communication with the nodes in your inventory. This information can exist in the Ansible *group_vars* or *host_vars* directories, or inventory or in the playbook itself.

| Key         | Required | Choices    | Description                                         |
|-------------|----------|------------|-----------------------------------------------------|
| ``ansible_host`` | yes      |            | Specifies the hostname or address for connecting to the remote device over the specified transport |
| ``ansible_port`` | no       |            | Specifies the port used to build the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_PORT` option is used; it defaults to 22 |
| ``ansible_ssh_user`` | no       |            | Specifies the username that authenticates the CLI login for the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_USER` environment variable value is used  |
| ``ansible_ssh_pass`` | no       |            | Specifies the password that authenticates the connection to the remote device.  |
| ``ansible_become`` | no       | yes, no\*   | Instructs the module to enter privileged mode on the remote device before sending any commands; if value is unspecified, the `ANSIBLE_BECOME` environment variable value is used, and the device attempts to execute all commands in non-privileged mode |
| ``ansible_become_method`` | no       | enable, sudo\*   | Instructs the module to allow the become method to be specified for handling privilege escalation; if value is unspecified, the `ANSIBLE_BECOME_METHOD` environment variable value is used |
| ``ansible_become_pass`` | no       |            | Specifies the password to use if required to enter privileged mode on the remote device; if ``ansible_become`` is set to no this key is not applicable |
| ``ansible_network_os`` | yes      | os6, null\*  | Loads the correct terminal and cliconf plugins to communicate with the remote device |

> **NOTE**: Asterisk (\*) denotes the default value if none is specified.

Example playbook
----------------

This example uses the *os6_snmp* role to completely set up the SNMP server attributes. It creates a *hosts* file with the switch details and corresponding variables. The hosts file should define the `ansible_network_os` variable with corresponding Dell EMC OS6 name. 

When `os6_cfg_generate` is set to true, the variable generates the configuration commands as a .part file in *build_dir* path. By default, the variable is set to false. It writes a simple playbook that only references the *os6_snmp* role. By including the role, you automatically get access to all of the tasks to configure SNMP features. 

**Sample hosts file**
 
    switch1 ansible_host= <ip_address> 

**Sample host_vars/switch1**

    hostname: switch1
    ansible_become: yes
    ansible_become_method: enable
    ansible_become_pass: xxxxx
    ansible_ssh_user: xxxxx
    ansible_ssh_pass: xxxxx
    ansible_network_os: dellemc.os6.os6
    build_dir: ../temp/temp_os6
	  
    os6_snmp:
      snmp_contact:  test
      snmp_location: Santa Clara
      snmp_community:
        - name: public
          access_mode: ro
          state: present
        - name: private
          access_mode: rw
          state: present
      snmp_host:
        - ip: 10.0.0.1
          communitystring: public
          udpport: 1
          state: absent
      snmp_traps:
        - name: config
          state: present

**Simple playbook to setup snmp — switch1.yaml**

    - hosts: switch1
      roles:
         - dellemc.os6.os6_snmp

**Run**

    ansible-playbook -i hosts switch1.yaml

(c) 2017-2020 Dell Inc. or its subsidiaries. All rights reserved.
