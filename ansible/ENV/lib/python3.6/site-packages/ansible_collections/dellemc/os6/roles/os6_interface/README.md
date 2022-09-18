Interface role
==============

This role facilitates the configuration of interface attributes. It supports the configuration of admin state, description, MTU, IP address, IP helper, suppress_ra, and port mode. This role is abstracted for Dell EMC PowerSwitch platforms running Dell EMC OS6.

The interface role requires an SSH connection for connectivity to a Dell EMC OS6 device. You can use any of the built-in OS connection variables.


Role variables
--------------

- Role is abstracted using the `ansible_network_os` variable that can take `dellemc.os6.os6` as a value
- If `os6_cfg_generate` is set to true, the variable generates the role configuration commands in a file
- Any role variable with a corresponding state variable setting to absent negates the configuration of that variable
- Setting an empty value for any variable negates the corresponding configuration
- `os6_interface` (dictionary) holds a dictionary with the interface name; interface name can correspond to any of the valid OS interfaces with the unique interface identifier name
- For physical interfaces, the interface name must be in *<interfacename> <tuple>* format; for logical interfaces, the interface must be in *<logical_interfacename> <id>* format; physical interface name can be *Te1/0/1* for os6 devices
- For interface ranges, the interface name must be in *range <interface_type> <node/slot/port[:subport]-node/slot/port[:subport]>* format
- Logical interface names can be *vlan 1* or *port-channel 1*
- Variables and values are case-sensitive

> **NOTE**: Only define supported variables for the interface type. For example, do not define the *switchport* variable for a logical interface, and do not define an IP address for physical interfaces in OS6 devices.

**os6_interface name keys**

| Key        | Type                      | Description                                             | Support               |
|------------|---------------------------|---------------------------------------------------------|-----------------------|
| ``desc``  | string         | Configures a single line interface description  | os6 |
| ``portmode`` | string | Configures port-mode according to the device type | os6 (access and trunk)  |
| ``admin``      | string: up,down\*              | Configures the administrative state for the interface; configuring the value as administratively "up" enables the interface; configuring the value as administratively "down" disables the interface | os6 |
| ``suppress_ra`` | string; present,absent     | Configures IPv6 router advertisements if set to present | os6 |
| ``ip_type_dynamic`` | boolean: true,false           | Configures IP address DHCP if set to true (*ip_and_mask* is ignored if set to true) | os6 |
| ``ip_and_mask`` | string | configures the specified IP address to the interface VLAN on os6 devices (192.168.11.1 255.255.255.0 format) | os6 |
| ``ipv6_and_mask`` | string | configures a specified IP address to the interface VLAN on os6 devices (2001:4898:5808:ffa2::1/126 format) | os6 |
| ``ipv6_reachabletime``       | integer                       | Configures the reachability time for IPv6 neighbor discovery (0 to 3600000), field needs to be left blank to remove the reachability time | os6 |
| ``ip_helper`` | list | Configures DHCP server address objects (see ``ip_helper.*``) | os6 |
| ``ip_helper.ip`` | string (required)         | Configures the IPv4 address of the DHCP server (A.B.C.D format)  | os6 |
| ``ip_helper.state`` | string: absent,present\* | Deletes the IP helper address if set to absent           | os6 |

> **NOTE**: Asterisk (*) denotes the default value if none is specified.

Connection variables
--------------------

Ansible Dell EMC Networking roles require connection information to establish communication with the nodes in your inventory. This information can exist in the Ansible *group_vars* or *host_vars* directories or inventory, or in the playbook itself.

| Key         | Required | Choices    | Description                                         |
|-------------|----------|------------|-----------------------------------------------------|
| ``ansible_host`` | yes      |            | Specifies the hostname or address for connecting to the remote device over the specified transport |
| ``ansible_port`` | no       |            | Specifies the port used to build the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_PORT` option is used; it defaults to 22 |
| ``ansible_ssh_user`` | no       |            | Specifies the username that authenticates the CLI login for the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_USER` environment variable value is used  |
| ``ansible_ssh_pass`` | no       |            | Specifies the password that authenticates the connection to the remote device.  |
| ``ansible_become`` | no       | yes, no\*   | Instructs the module to enter privileged mode on the remote device before sending any commands; if value is unspecified, the `ANSIBLE_BECOME` environment variable value is used, and the device attempts to execute all commands in non-privileged mode |
| ``ansible_become_method`` | no       | enable, sudo\*   | Instructs the module to allow the become method to be specified for handling privilege escalation; if value is unspecified, the `ANSIBLE_BECOME_METHOD` environment variable value is used. |
| ``ansible_become_pass`` | no       |            | Specifies the password to use if required to enter privileged mode on the remote device; if ``ansible_become`` is set to no this key is not applicable. |
| ``ansible_network_os`` | yes      | os6, null\*  | This value is used to load the correct terminal and cliconf plugins to communicate with the remote device. |

> **NOTE**: Asterisk (*) denotes the default value if none is specified.

Example playbook
----------------

This example uses the *os6-interface* role to set up description, MTU, admin status, portmode, and switchport details for an interface. The example creates a *hosts* file with the switch details and orresponding variables. The hosts file should define the `ansible_network_os` variable with corresponding Dell EMC OS6 name.

When `os6_cfg_generate` is set to true, the variable generates the configuration commands as a .part file in *build_dir* path. By default, this variable is set to false. The example writes a simple playbook that only references the *os6-interface* role.

**Sample hosts file**

    switch1 ansible_host= <ip_address> 

**Sample host_vars/switch1**

    hostname: "switch1"
    ansible_become: yes
    ansible_become_method: enable
    ansible_become_pass: xxxxx
    ansible_ssh_user: xxxxx
    ansible_ssh_pass: xxxxx
    ansible_network_os: dellemc.os6.os6
    build_dir: ../temp/temp_os6

    os6_interface:
        Te1/0/8:
                desc: "Connected to Spine1"
                portmode: trunk
                admin: up
        vlan 100:
                admin: down
                ip_and_mask: 3.3.3.3 255.255.255.0
                ipv6_and_mask: 2002:4898:5408:faaf::1/64
                suppress_ra: present
                ip_helper:
                  - ip: 10.0.0.36
                    state: absent
                ipv6_reachabletime: 600000
        vlan 20:
                suppress_ra: absent
                admin: up

**Simple playbook to setup system — switch1.yaml**

    - hosts: switch1
      roles:
         - dellemc.os6.os6_interface
 
**Run**

    ansible-playbook -i hosts switch1.yaml

(c) 2017-2020 Dell Inc. or its subsidiaries. All rights reserved.
