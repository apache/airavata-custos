Logging role
============

This role facilitates the configuration of global logging attributes, and it supports the configuration of logging servers. This role is abstracted for Dell EMC PowerSwitch platforms running Dell EMC OS9.

The Logging role requires an SSH connection for connectivity to a Dell EMC OS9 device. You can use any of the built-in OS connection variables.


Role variables
--------------

- Role is abstracted using the `ansible_network_os` variable that can take `dellemc.os9.os9` as the value
- If the `os9_cfg_generate` variable is set to true, it generates the role configuration commands in a file
- Any role variable with a corresponding state variable set to absent negates the configuration of that variable
- Setting an empty value for any variable negates the corresponding configuration
- Variables and values are case-sensitive

**os9_logging keys**

| Key        | Type                      | Description                                             | Support               |
|------------|---------------------------|---------------------------------------------------------|-----------------------|
| ``logging`` | list | Configures the logging server (see ``logging.*``) | os9 |
| ``logging.ip`` | string (required)         | Configures the IPv4 address for the logging server (A.B.C.D format) | os9 |
| ``logging.secure_port`` | integer | Specifies log messages over the TLS port, CA certificates must be installed to specify the log messages over TLS port  | os9 |
| ``logging.tcp_port`` | integer | Specifies log messages over the TCP port if *secure_port* is not defined | os9 |
| ``logging.udp_port`` | integer | Specifies log messages over the UDP port if both TCP and the secure port key are not defined | os9 |
| ``logging.vrf`` | dict | Specifies a VRF instance to be used to reach the host | os9 |
| ``logging.vrf.name`` | string | Specifies the VRF name | os9 |
| ``logging.vrf.secure_port`` | integer | Specifies log messages over the TLS port,  CA certificates must be installed to specify the log messages over TLS port | os9 |
| ``logging.vrf.tcp_port`` | integer | Specifies log messages over the TCP port if *secure_port key* is not defined | os9 |
| ``logging.vrf.udp_port`` | integer | Specifies log messages over the UDP port if both TCP and *secure_port_key* is not defined | os9 |
| ``logging.vrf.state`` | string: absent,present\* | Deletes VRF instance of the logging server if set to absent | os9 |
| ``logging.state`` | string: absent,present\*     | Deletes the logging server if set to absent   | os9 |
| ``buffer`` | integer | Specifies the buffered logging severity level (0 to 7) | os9 |
| ``console_level`` | integer | Configures the console logging level (0 to 7) | os9  |
| ``trap_level`` | integer | Configures the syslog server severity level (0 to 7) | os9|
| ``syslog_version`` | integer | Configures the syslog version (0/1) | os9 |
| ``monitor`` | integer | Configures the terminal line logging level (0 to 7) | os9|
| ``history`` | integer | Configures the syslog history table (0 to 7) | os9 |
| ``history_size`` | integer | Specifies the history table size | os9  |
| ``on`` | boolean |  Enables logging to all supported destinations if set to true | os9 |
| ``extended`` | boolean | Enables extended logging if set to true | os9  |
| ``coredump`` | dict | Configures coredump logging | os9  |
| ``coredump.server`` | dict | Specifies all server details | os9 |
| ``coredump.server.server_ip`` | string (required) | Specifies the IPv4/IPv6 address of the logging server | os9 |
| ``coredump.server.username`` | string | Specifies the username to be configured | os9 |
| ``coredump.server.password`` | string | Specifies the password to be configured | os9 |
| ``coredump.server.state`` | string: present,absent\* | Deletes the coredump server if set to absent | os9 |
| ``coredump.stackunit`` |dict | Specifies details for enabling a coredump on the stack-unit | os9 |
| ``coredump.stackunit.all`` | boolean | Enables a coredump on all stack-units  | os9 |
| ``coredump.stackunit.unit_num`` | integer | Specifies the stack-unit number (0 to 5) | os9 |
| ``coredump.stackunit.state`` | string: present,absent\*| Deletes the stack-unit coredump if set to absent | os9 |
| ``source_interface`` | string | Configures the source interface for logging | os9 |

> **NOTE**: Asterisk (_*_) denotes the default value if none is specified. 

Connection variables
--------------------

Ansible Dell EMC network roles require connection information to establish communication with the nodes in your inventory. This information can exist in the Ansible *group_vars* or *host_vars* directories or inventory, or in the playbook itself.

| Key         | Required | Choices    | Description                                         |
|-------------|----------|------------|-----------------------------------------------------|
| ``ansible_host`` | yes      |            | Specifies the hostname or address for connecting to the remote device over the specified transport |
| ``ansible_port`` | no       |            | Specifies the port used to build the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_PORT` option is used; it defaults to 22 |
| ``ansible_ssh_user`` | no       |            | Specifies the username that authenticates the CLI login for the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_USE`R environment variable value is used  |
| ``ansible_ssh_pass`` | no       |            | Specifies the password that authenticates the connection to the remote device |
| ``ansible_become`` | no       | yes, no\*   | Instructs the module to enter privileged mode on the remote device before sending any commands; if value is unspecified, the `ANSIBLE_BECOME` environment variable value is used, and the device attempts to execute all commands in non-privileged mode |
| ``ansible_become_method`` | no       | enable, sudo\*   | Instructs the module to allow the become method to be specified for handling privilege escalation; if value is unspecified, the `ANSIBLE_BECOME_METHOD` environment variable value is used |
| ``ansible_become_pass`` | no       |            | Specifies the password to use if required to enter privileged mode on the remote device; if ``ansible_become`` is set to no this key is not applicable |
| ``ansible_network_os`` | yes      | os9, null\*  | Loads the correct terminal and cliconf plugins to communicate with the remote device |

> **NOTE**: Asterisk (\*) denotes the default value if none is specified.

Example playbook
----------------

This example uses the *os9_logging* role to completely set up logging servers. It creates a *hosts* file with the switch details and corresponding variables. The hosts file should define the `ansible_network_os` variable with corresponding Dell EMC OS9 name. 

When `os9_cfg_generate` is set to true, the variable generates the configuration commands as a .part file in *build_dir* path. By default, the variable is set to false.

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
	  
    os9_logging:
      logging:
       - ip : 1.1.1.1
         state: present
       - ip: 2.2.2.2
         secure_port: 1025
         tcp_port: 1024
         udp_port: 2000
         state: present
       - ip: 3.3.3.3
         vrf:
           name: test
           secure_port: 1024
           tcp_port: 1025
           udp_port: 2000
           state: present
         secure_port: 1025
         tcp_port: 2000
         udp_port: 1025
         state: present
     buffer: 5
     console: 7
     trap: 5
     version: 5
     history: 4
     history_size: 3
     monitor: 5
     on: true
     extended: true
     coredump:
       server:
         server_ip: 2.2.2.2
         username: u1
         password: pwd
         state: present
       stackunit:
          all: true
          unit_num: 5
          state: present
     source_interface: "fortyGigE 1/9"

**Simple playbook to setup logging — leaf.yaml**

    - hosts: leaf1
      roles:
         - dellemc.os9.os9_logging

**Run**

    ansible-playbook -i hosts leaf.yaml

(c) 2020 Dell Inc. or its subsidiaries. All rights reserved.
