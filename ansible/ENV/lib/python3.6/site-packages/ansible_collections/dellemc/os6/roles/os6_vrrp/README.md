VRRP role
=========

This role facilitates configuring virtual router redundancy protocol (VRRP) attributes. It supports the creation of VRRP groups for interfaces and setting the VRRP group attributes. This role is abstracted for OS6.

The VRRP role requires an SSH connection for connectivity to a Dell EMC OS6 device. You can use any of the built-in OS connection variables.

Role variables
--------------

- Role is abstracted using the `ansible_network_os` variable that can take `dellemc.os6.os6` as a value
- If `os6_cfg_generate` is set to true, the variable generates the role configuration commands in a file
- Any role variable with a corresponding state variable set to absent negates the configuration of that variable
- Setting an empty value for any variable negates the corresponding configuration
- `os6_vrrp` (dictionary) holds a dictionary with the interface name key
- Interface name can correspond to any of the valid os6 interface with a unique interface identifier name
- Physical interfaces names must be in *<interfacename> <tuple>* format (for example *Fo1/0/1*)
- Variables and values are case-sensitive

| Key        | Type                      | Description                                             | Support               |
|------------|---------------------------|---------------------------------------------------------|-----------------------|
| ``vrrp_group_id``    | integer (required)  | Configures the ID for the VRRP group (1 to 255) | os6 |
| ``description``      | string          | Configures a single line description for the VRRP group | os6 |
| ``virtual_address``  | string          | Configures a virtual address to the VRRP group (A.B.C.D format) | os6 |
| ``enable``      | boolean: true,false        | Enables/disables the VRRP group at the interface  | os6 |
| ``preempt``      | boolean: true\*,false          | Configures preempt mode on the VRRP group | os6 |
| ``priority``      |integer          | Configures priority for the VRRP group (1 to 255; default 100), field needs to be left blank to remove the priority  | os6 |
| ``state``       | string: present\*,absent          | Deletes the VRRP group from the interface if set to absent; VRRP group needs to be disabled to delete the VRRP group from the interface | os6 |
                                                                                                 
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

This example uses the *os6_vrrp* role to configure VRRP commands at the interfaces. It creates a *hosts* file with the switch details and corresponding variables. The hosts file should define the `ansible_network_os` variable with corresponding Dell EMC OS6 name. 

When `os6_cfg_generate` is set to true, the variable generates the configuration commands as a .part file in *build_dir* path. By default, the variable is set to false. It writes a simple playbook that only references the *os6_vrrp* role.

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
    build_dir: ../temp/os6
    os6_vrrp:
        vlan 4:
          - vrrp_group_id: 4
            state: present
            description: "Interface-vrrp4"
            virtual_address: 10.2.0.1
            enable: true
            priority: 120
            preempt: false

          
**Simple playbook to setup system — switch1.yaml**

    - hosts: switch1
      roles:
         - dellemc.os6.os6_vrrp
                
**Run**

    ansible-playbook -i hosts leaf.yaml

(c) 2017-2020 Dell Inc. or its subsidiaries. All rights reserved.
