# xSTP role

This role facilitates the configuration of xSTP attributes. It supports multiple version of spanning-tree protocol (STP), rapid spanning-tree (RSTP), rapid per-VLAN spanning-tree (Rapid PVST+), multiple spanning-tree (MST), and per-VLAN spanning-tree (PVST). It supports the configuration of bridge priority, enabling and disabling spanning-tree, creating and deleting instances, and mapping virtual LAN (VLAN) to instances. This role is abstracted for Dell EMC PowerSwitch platforms running Dell EMC OS6.

The xSTP role requires an SSH connection for connectivity to a Dell EMC OS6 device. You can use any of the built-in OS connection variables.


Role variables
--------------

- Role is abstracted using the `ansible_network_os` variable that can take `dellemc.os6.os6` as a value
- `os6_xstp` (dictionary) contains the hostname (dictionary)
- Hostname is the value of the *hostname* variable that corresponds to the name of the OS device
- Any role variable with a corresponding state variable set to absent negates the configuration of that variable
- Setting an empty value to any variable negates the corresponding configuration
- Variables and values are case-sensitive

**hostname keys**

| Key        | Type                      | Description                                             | Support              |
|------------|---------------------------|---------------------------------------------------------|----------------------|
| ``type``       | string (required) | Configures the type of spanning-tree mode specified that can vary according to the OS device; include RSTP, rapid-PVST, and MST | os6 |
| ``enable``  | boolean: true,false             | Enables/disables the spanning-tree protocol specified in the type variable | os6 |
| ``stp``  | dictionary             | Configures simple spanning-tree protocol (see ``stp.* keys``) | os6 |
| ``stp.bridge_priority`` | integer | Configures bridge-priority for the spanning-tree (0 to 61440 in multiples of 4096) | os6 |
| ``rstp``  | dictionary             | Configures rapid spanning-tree (see ``rstp.*``)  | os6 |
| ``rstp.bridge_priority`` | integer | Configures bridge-priority for the spanning-tree (0 to 61440 in multiples of 4096) | os6 |
| ``pvst``  | dictionary     | Configures per-VLAN spanning-tree protocol (see ``pvst.*``) | os6 |
| ``pvst.vlan`` | list | Configures the VLAN for PVST (see ``vlan.*``)  | os6 |
| ``vlan.range_or_id``  | string             | Configures a VLAN/range of VLANs for the per-VLAN spanning-tree protocol | os6 |
| ``vlan.bridge_priority`` | integer | Configures bridge-priority for the per-VLAN spanning-tree (0 to 61440 in multiples of 4096); mutually exclusive with *vlan.root* | os6 |
| ``vlan.state`` | string: absent, present\* | Deletes the configured PVST VLAN with ID if set to absent | os6 |
| ``mstp``  | dictionary     | Configures multiple spanning-tree protocol (see ``mstp.*``)  | os6 |
| ``mstp.mstp_instances`` | list | Configures a MSTP instance (see ``mstp_instances.*``)  | os6 |
| ``mstp_instances.number``     | integer                   | Configures the multiple spanning-tree instance number | os6 |
| ``mstp_instances.vlans``      | string     | Configures a VLAN/range of VLANs by mapping it to the instance number  | os6 |
| ``mstp_instances.bridge_priority`` | integer | Configures the bridge-priority for the spanning-tree (0 to 61440 in multiples of 4096); mutually exclusive with *mstp_instances.root* | os6 |
| ``mstp_instances.vlans_state`` | string: absent,present\* | Deletes a set of VLANs mapped to the spanning-tree instance if set to absent | os6 |
| ``intf`` | list | Configures multiple spanning-tree in an interface (see ``intf.*``)  | os6 |
| ``intf <interface name>``| dictionary | Configures the interface name (see ``intf.<interface name>.*``) | os6 |
| ``intf.<interface name>.edge_port`` | boolean: true,false | Enables port fast at the interface level if set to true | os6 |

> **NOTE**: Asterisk (_*_) denotes the default value if none is specified.

Connection variables
--------------------

Ansible Dell EMC network roles require connection information to establish communication with the nodes in your inventory. This information can exist in the Ansible *group_vars* or *host_vars* directories or inventory or in the playbook itself.

| Key         | Required | Choices    | Description                                           |
|-------------|----------|------------|-------------------------------------------------------|
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

This example uses the *os6_xstp* role to configure different variants of spanning-tree. Based on the type of STP and defined objects, VLANs are associated and bridge priorities are assigned. It creates a *hosts* file with the switch details, and a *host_vars* file with connection variables. The corresponding role variables are defined in the *vars/main.yml* file at the role path. 
It writes a simple playbook that only references the *os6_xstp* role. By including the role, you automatically get access to all of the tasks to configure xSTP. 

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


**Sample vars/main.yml**

    os6_xstp:
        type: stp
        enable: true
        stp:
          bridge_priority: 4096
        pvst:
          vlan:
            - range_or_id: 10
              bridge_priority: 4096
              state: present
        mstp:
          mstp_instances:
            - number: 1
              vlans: 10,12
              bridge_priority: 4096
              vlans_state: present
        intf:
          Fo4/0/1:
            edge_port: true

**Simple playbook to setup system — switch1.yml**

    - hosts: switch1
      roles:
         - dellemc.os6.os6_xstp
 
**Run**

    ansible-playbook -i hosts switch1.yml

(c) 2017-2020 Dell Inc. or its subsidiaries. All rights reserved.
