# xSTP role

This role facilitates the configuration of xSTP attributes. It supports multiple version of spanning-tree protocol (STP), rapid spanning-tree (RSTP), rapid per-VLAN spanning-tree (Rapid PVST+), multiple spanning-tree (MST), and per-VLAN spanning-tree (PVST). It supports the configuration of bridge priority, enabling and disabling spanning-tree, creating and deleting instances, and mapping virtual LAN (VLAN) to instances. This role is abstracted for Dell EMC PowerSwitch platforms running Dell EMC OS9.

The xSTP role requires an SSH connection for connectivity to a Dell EMC OS9 device. You can use any of the built-in OS connection variables.

Role variables
--------------

- Role is abstracted using the `ansible_network_os` variable that can take `dellemc.os9.os9` as the value
- If `os9_cfg_generate` is set to true, the variable generates the role configuration commands in a file
- `os9_xstp` (dictionary) contains the hostname (dictionary)
- Hostname is the value of the *hostname* variable that corresponds to the name of the OS device
- Any role variable with a corresponding state variable set to absent negates the configuration of that variable
- Setting an empty value to any variable negates the corresponding configuration
- Variables and values are case-sensitive

**hostname keys**

| Key        | Type                      | Description                                             | Support              |
|------------|---------------------------|---------------------------------------------------------|----------------------|
| ``type``       | string (required) | Configures the type of spanning-tree mode specified including STP, RSTP, PVST, and MSTP | os9 |
| ``enable``  | boolean: true,false             | Enables/disables the spanning-tree protocol specified in the type variable | os9 |
| ``stp``  | dictionary             | Configures simple spanning-tree protocol (see ``stp.* keys``) | os9 |
| ``stp.bridge_priority`` | integer | Configures bridge-priority for the spanning-tree (0 to 61440 in multiples of 4096) | os9 |
| ``stp.state`` | string: absent,present\* | Deletes the configured STP if set to absent | os9 |
| ``rstp``  | dictionary             | Configures rapid spanning-tree (see ``rstp.*``)  | os9 |
| ``rstp.bridge_priority`` | integer | Configures bridge-priority for the spanning-tree (0 to 61440 in multiples of 4096) | os9 |
| ``rstp.state ``| string: absent,present\* | Deletes the configured RSTP in os9 devices if set to absent | os9 |
| ``pvst``  | dictionary     | Configures per-VLAN spanning-tree protocol (see ``pvst.*``) | os9 |
| ``pvst.vlan`` | list | Configures the VLAN for PVST (see ``vlan.*``)  | os9 |
| ``vlan.range_or_id``  | string             | Configures a VLAN/range of VLANs for the per-VLAN spanning-tree protocol | os9 |
| ``vlan.bridge_priority`` | integer | Configures bridge-priority for the per-VLAN spanning-tree (0 to 61440 in multiples of 4096); mutually exclusive with *vlan.root* | os9 |
| ``pvst.state`` | string: absent,present\* | Deletes the configured PVST if set to absent | os9 |
| ``mstp``  | dictionary     | Configures multiple spanning-tree protocol (see ``mstp.*``)  | os9 |
| ``mstp.mstp_instances`` | list | Configures a MSTP instance (see ``mstp_instances.*``)  | os9 |
| ``mstp_instances.number``     | integer                   | Configures the multiple spanning-tree instance number | os9 |
| ``mstp_instances.vlans``      | string     | Configures a VLAN/range of VLANs by mapping it to the instance number in os9 devices | os9 |
| ``mstp_instances.bridge_priority`` | integer | Configures the bridge-priority for the spanning-tree (0 to 61440 in multiples of 4096); mutually exclusive with *mstp_instances.root* | os9 |
| ``mstp_instances.vlans_state`` | string: absent,present\* | Deletes a set of VLANs mapped to the spanning-tree instance if set to absent | os9 |
| ``mstp.state`` | string: absent,present\* | Deletes the configured MSTP if set to absent | os9 |
| ``intf`` | list | Configures multiple spanning-tree in an interface (see ``intf.*``)  | os9 |
| ``intf <interface name>``| dictionary | Configures the interface name (see ``intf.<interface name>.*``) | os9 |
| ``intf.<interface name>.stp_type`` | list: stp,mstp,pvst,rstp | Configures the list of spanning-tree in an interface | os9 |
| ``intf.<interface name>.edge_port`` | boolean: true,false | in os9 devices according to the stp_type EdgePort is configured;  | os9 |

> **NOTE**: Asterisk (_*_) denotes the default value if none is specified.

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

This example uses the *os9_xstp* role to configure different variants of spanning-tree. Based on the type of STP and defined objects, VLANs are associated and bridge priorities are assigned. It creates a *hosts* file with the switch details, and a *host_vars* file with connection variables. The corresponding role variables are defined in the *vars/main.yml* file at the role path. The hosts file should define the `ansible_network_os` variable with corresponding Dell EMC OS9 name.

It writes a simple playbook that only references the *os9_xstp* role. By including the role, you automatically get access to all of the tasks to configure xSTP. When `os9_cfg_generate` is set to true, the variable generates the configuration commands as a .part file in build_dir path. By default, this variable is set to false. The example writes a simple playbook that only references the *os9_xstp* role.

**Sample hosts file**

    spine1 ansible_host= <ip_address> 

**Sample host_vars/spine1**
    
    hostname: spine1
    ansible_become: yes
    ansible_become_method: xxxxx
    ansible_become_pass: xxxxx
    ansible_ssh_user: xxxxx
    ansible_ssh_pass: xxxxx
    ansible_network_os: dellemc.os9.os9
    build_dir: ../temp/os9


**Sample vars/main.yml**

     os9_xstp:
        type: rstp
        enable: true
        stp:
          bridge_priority: 4096
          state: present
        rstp:
          bridge_priority: 4096
        pvst:
          vlan:
            - range_or_id: 10
              bridge_priority: 4096
        mstp:
          mstp_instances:
            - number: 1
              vlans: 10,12
              bridge_priority: 4096
              vlans_state: present
        intf:
          fortyGigE 1/25:
            stp_type:
              - stp
              - mstp
            edge_port: true

**Simple playbook to setup system — spine.yml**

    - hosts: spine
      roles:
         - dellemc.os9.os9_xstp
 
**Run**

    ansible-playbook -i hosts spine.yml

(c) 2017-2020 Dell Inc. or its subsidiaries. All rights reserved.
