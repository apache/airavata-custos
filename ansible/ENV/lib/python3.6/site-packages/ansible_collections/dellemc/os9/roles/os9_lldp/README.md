LLDP role
=========

This role facilitates the configuration of link layer discovery protocol (LLDP) attributes at a global and interface level. It supports the configuration of hello, mode, multiplier, advertise TLVs, management interface, FCoE, iSCSI at global and interface level. This role is abstracted for Dell EMC PowerSwitch platforms running Dell EMC OS9.

The LLDP role requires an SSH connection for connectivity to a Dell EMC OS9 device. You can use any of the built-in OS connection variables.

Role variables
--------------

- Role is abstracted using the `ansible_network_os` variable that can take `dellemc.os9.os9` as the value
- If `os9_cfg_generate` is set to true, the variable generates the role configuration commands in a file
- Any role variable with a corresponding state variable set to absent negates the configuration of that variable
- Setting an empty value for any variable negates the corresponding configuration
- Variables and values are case-sensitive

**os9_lldp keys**

| Key        | Type                      | Description                                             | Support               |
|------------|---------------------------|---------------------------------------------------------|-----------------------|
| ``global_lldp_state`` | string: absent,present   | Deletes LLDP at a global level if set to absent | os9 |
| ``enable``  | boolean         | Enables or disables LLDP at a global level | os9 |
| ``hello`` | integer | Configures the global LLDP hello interval (5 to 180) | os9 |
| ``mode``  | string: rx,tx   | Configures global LLDP mode configuration | os9 |
| ``multiplier`` | integer | Configures the global LLDP multiplier (2 to 10) | os9 |
| ``fcoe_priority_bits`` | integer | Configures priority bits for FCoE traffic (1 to FF) |  os9 |
| ``iscsi_priority_bits`` | integer | Configures priority bits for iSCSI traffic (1 to FF) | os9 |
| ``dcbx`` | dictionary  | Configures DCBx parameters at the global level (see ``dcbx.*``)     |  os9 |
| ``dcbx.version`` | string     | Configures the DCBx version | os9 |
| ``advertise`` | dictionary     | Configures LLDP-MED and TLV advertisement at the global level (see ``advertise.*``) | os9 |
| ``advertise.dcbx_tlv`` | string     | Configures DCBx TLVs advertisements | os9 |
| ``advertise.dcbx_tlv_state`` | string: present,absent     | Deletes DCBx TLVs advertisement if set to absent | os9 |
| ``advertise.dcbx_appln_tlv`` | string     | Configures DCBx application priority TLVs advertisement | os9 |
| ``advertise.dcbx_appln_tlv_state`` | string: present,absent     | Deletes DCBx application priority TLVs advertisement if set to absent | os9 |
| ``advertise.dot1_tlv`` | dictionary     | Configures 802.1 TLVs advertisement (see ``dot1_tlv.*``) | os9 |
| ``dot1_tlv.port_tlv`` | dictionary     | Configures 802.1 TLVs advertisement (see ``port_tlv.*``) | os9 |
| ``port_tlv.protocol_vlan_id`` | boolean     | Configures 802.1 VLAN ID TLVs advertisement | os9 |
| ``port_tlv.port_vlan_id`` | boolean     | Configures 802.1 VLAN ID TLVs advertisement | os9 |
| ``dot1_tlv.vlan_tlv`` | dictionary     | Configures 802.1 VLAN TLVs advertisement (see ``vlan_tlv.*``) |  os9 |
| ``vlan_tlv.vlan_range`` | string     | Configures 802.1 VLAN name TLVs advertisement | os9 |
| ``advertise.dot3_tlv`` | dictionary     | Configures 802.3 TLVs advertisement (see ``dot3_tlv.*``) | os9 |
| ``dot3_tlv.max_frame_size`` | boolean     | Configures 802.3 maximum frame size TLVs advertisement | os9 |
| ``advertise.port_descriptor`` | boolean     | Configures global port descriptor advertisement | os9 |
| ``advertise.management_tlv`` | string     | Configures global management TLVs advertisement | os9 |
| ``advertise.management_tlv_state`` | string: absent,present     | Deletes global TLVs advertisement if set to absent | os9 |
| ``advertise.med`` | dictionary     | Configures MED TLVs advertisement (see ``med_tlv.*``) | , os9 |
| ``med.global_med`` | boolean     | Configures global MED TLVs advertisement | os9 |
| ``med.application`` | list     | Configures global MED TLVs advertisement for an application (see ``application.*``) | os9 |
| ``application.name`` | string     | Configures the application name for MED TLVs advertisement | os9 |
| ``application.vlan_id`` | integer     | Configures the VLAN ID for the application MED TLVs advertisement (1 to 4094) | os9 |
| ``application.priority_tagged`` | boolean     | Configures priority tagged for the application MED TLVs advertisement; mutually exclusive with *application.vlan_id* | os9 |
| ``application.l2_priority`` | integer     | Configures the L2 priority for the application MED TLVs advertisement (0 to 7) | os9 | 
| ``application.code_point_value`` | integer     | Configures differentiated services code point values for MED TLVs advertisement (0 to 63) | os9 |
| ``med.location_identification`` | list     | Configures MED location identification TLVs advertisement (see ``location_identification.*``) | os9 |
| ``location_identification.loc_info`` | string     | Configures location information for MED TLVs advertisement | os9 |
| ``location_identification.value`` | string     | Configures location information values | os9 |
| ``location_identification.state`` | string: absent,present   | Deletes the location information if set to absent | os9 |
| ``management_interface`` | dictionary     | Configures LLDP on the management interface (see ``management_interface.*``)     | os9 |
| ``management_interface.enable``  | boolean         | Enables/disables LLDP on the management interface | os9 |
| ``management_interface.hello`` | integer | Configures LLDP hello interval on the management interface (5 to 180) | os9 |
| ``management_interface.mode``  | string: rx,tx   | Configures LLDP mode on the management interface | os9 |
| ``management_interface.multiplier`` | integer | Configures LLDP multiplier on the management interface (2 to 10) | os9 |
| ``management_interface.advertise`` | dictionary     | Configures TLV advertisement on the management interface (see ``advertise.*``)     | os9 |
| ``advertise.port_descriptor`` | boolean     | Configures port descriptor advertisement on the management interface  | os9 |
| ``advertise.management_tlv`` | string     | Configures management TLVs advertisement  | os9 |
| ``advertise.management_tlv_state`` | string: absent,present     | Deletes management TLVs advertisement if set to absent | os9 |
| ``local_interface`` | dictionary     | Configures LLDP at the interface level (see ``local_interface.*``) | os9 |
| ``local_interface.<interface name>`` | dictionary     | Configures LLDP at the interface level (see ``interface name.*``)     | os9 |
| ``<interface name>.state`` | string: absent,present   | Deletes LLDP at the interface level if set to absent | os9 |
|  ``<interface name>.enable``  | boolean         | Enables or disables LLDP at the interface level | os9 |
| ``<interface name>.hello`` | integer | Configures LLDP hello interval at the interface level (5 to 180) | os9 |
| ``<interface name>.mode``  | string: rx,tx   | Configures LLDP mode configuration at the interface level | os9 |
| ``<interface name>.multiplier`` | integer | Configures LLDP multiplier at the interface level (2 to 10) |  os9 |
| ``<interface name>.dcbx`` | dictionary  | Configures DCBx parameters at the interface level (see ``dcbx.*``)     | os9 |
| ``dcbx.version`` | string     | Configures DCBx version at the interface level  | os9 |
| ``<interface name>.advertise`` | dictionary     | Configures LLDP-MED TLV advertisement at the interface level (see ``advertise.*``)     | os9 |
| ``advertise.dcbx_tlv`` | string     | Configures DCBx TLVs advertisement at the interface level | os9 |
| ``advertise.dcbx_tlv_state`` | string: present,absent     | Deletes interface level DCBx TLVs advertisement if set to absent | os9 |
| ``advertise.dcbx_appln_tlv`` | string     | Configures DCBx application priority TLVs advertisement at the interface level |  os9 |
| ``advertise.dcbx_appln_tlv_state`` | string: present,absent     | Deletes interface level DCBx application priority TLVs advertisement if set to absent | os9 |
| ``advertise.dot1_tlv`` | dictionary     | Configures 802.1 TLVs advertisement at the interface level (see ``dot1_tlv.*``) | os9 |
| ``dot1_tlv.port_tlv`` | dictionary     | Configures 802.1 TLVs advertisement at the interface level (see ``port_tlv.*``) | os9 |
| ``port_tlv.protocol_vlan_id`` | boolean     | Configures 802.1 VLAN ID TLVs advertisement at the interface level | os9 |
| ``port_tlv.port_vlan_id`` | boolean     | Configures 802.1 VLAN ID TLVs advertisement at the interface level | os9 | 
| ``dot1_tlv.vlan_tlv`` | dictionary     | Configures 802.1 VLAN TLVs advertisement at the interface level (see ``vlan_tlv.*``) | os9 |
| ``vlan_tlv.vlan_range`` | string     | Configures 802.1 VLAN name TLVs advertisement at the interface level | os9  |
| ``advertise.dot3_tlv`` | dictionary     | Configures 802.3 TLVs advertisement at the interface level (see ``dot3_tlv.*``) | os9 |
| ``dot3_tlv.max_frame_size`` | boolean   | Configures 802.3 maximum frame size TLVs advertisement at the interface level | os9 |
| ``advertise.port_descriptor`` | boolean     | Configures port descriptor advertisement at the interface level | os9 |
| ``advertise.management_tlv`` | string     | Configures TLVs advertisement at the interface level | os9 | 
| ``advertise.management_tlv_state`` | string: absent,present     | Deletes TLVs advertisement at the interface level if set to absent | os9 |
| ``advertise.med`` | dictionary     | Configures MED TLVs advertisement at the interface level (see ``med_tlv.*``) | os9 |
| ``med.global_med`` | boolean     | Configures MED TLVs advertisement at the interface level | os9 |
| ``med.application`` | list     | Configures MED TLVs advertisement for the application at the interface level (see ``application.*``) | os9 |
| ``application.name`` | string     | Configures the application name for MED TLVs advertisement | os9 |
| ``application.vlan_id`` | integer     | Configures the VLAN ID for the application MED TLVs advertisement at the interface level (1 to 4094) | os9 |
| ``application.priority_tagged`` | boolean     | Configures priority tagged for the application MED TLVs advertisement at the interface level; mutually exclusive with *application.vlan_id* | os9 |
| ``application.l2_priority`` | integer     | Configures the L2 priority for the application MED TLVs advertisement at the interface level (0 to 7) | os9 |
| ``application.code_point_value`` | integer     | Configures differentiated services code point value for MED TLVs advertisement at the interface level (0 to 63) | os9 |
| ``med.location_identification`` | list     | Configures MED location identification TLVs advertisement at the interface level (see ``location_identification.*``) | os9 |
| ``location_identification.loc_info`` | string     | Configures location information for MED TLVs advertisement at the interface level | os9 |
| ``location_identification.value`` | string     | Configures the location information value for MED TLVs advertisement at the interface level | os9 |
| ``location_identification.state`` | string: absent,present   | Deletes the interface level MED location information if set to absent | os9 |


Connection variables
--------------------

Ansible Dell EMC network roles require connection information to establish communication with the nodes in your inventory. This information can exist in the Ansible *group_vars* or *host_vars* directories or inventory, or in the playbook itself.

| Key         | Required | Choices    | Description                                         |
|-------------|----------|------------|-----------------------------------------------------|
| ``ansible_host`` | yes      |            | Specifies the hostname or address for connecting to the remote device over the specified transport |
| ``ansible_port`` | no       |            | Specifies the port used to build the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_PORT` option is used; it defaults to 22 |
| ``ansible_ssh_user`` | no       |            | Specifies the username that authenticates the CLI login for the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_USER` environment variable value is used  |
| ``ansible_ssh_pass`` | no       |            | Specifies the password that authenticates the connection to the remote device. |
| ``ansible_become`` | no       | yes, no\*   | Instructs the module to enter privileged mode on the remote device before sending any commands; if value is unspecified, the `ANSIBLE_BECOME` environment variable value is used, and the device attempts to execute all commands in non-privileged mode |
| ``ansible_become_method`` | no       | enable, sudo\*   | Instructs the module to allow the become method to be specified for handling privilege escalation; if value is unspecified, the `ANSIBLE_BECOME_METHOD` environment variable value is used |
| ``ansible_become_pass`` | no       |            | Specifies the password to use if required to enter privileged mode on the remote device; if ``ansible_become`` is set to no this key is not applicable |
| ``ansible_network_os`` | yes      | os9, null\*  | Loads the correct terminal and cliconf plugins to communicate with the remote device |

> **NOTE**: Asterisk (\*) denotes the default value if none is specified.

Example playbook
----------------

This example uses the *os9_lldp* role to configure protocol lldp. It creates a *hosts* file with the switch details and corresponding variables. The hosts file should define the `ansible_network_os` variable with corresponding Dell EMC OS9 name. 

When `os9_cfg_generate` is set to true, the variable generates the configuration commands as a .part file in *build_dir* path. By default, the variable is set to false. It writes a simple playbook that only references the *os9_lldp* role.
 
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
    os9_lldp:
      global_lldp_state: present
      enable: false
      mode: rx
      multiplier: 3
      fcoe_priority_bits: 3
      iscsi_priority_bits: 3
      hello: 6
      dcbx:
        version: auto
      management_interface:
        hello: 7
        multiplier: 3
        mode: tx
        enable: true
        advertise:
          port_descriptor: false
          management_tlv: management-address system-capabilities
          management_tlv_state: absent
      advertise:
        dcbx_tlv: pfc
        dcbx_tlv_state: absent
        dcbx_appln_tlv: fcoe
        dcbx_appln_tlv_state:
        dot1_tlv:
          port_tlv:
            protocol_vlan_id: true
            port_vlan_id: true
          vlan_tlv:
            vlan_range: 2-4
        dot3_tlv:
          max_frame_size: false
        port_descriptor: false
        management_tlv: management-address system-capabilities
        management_tlv_state: absent
        med:
          global_med: true
          application:
            - name: "guest-voice"
              vlan_id: 2
              l2_priority: 3
              code_point_value: 4
            - name: voice
              priority_tagged: true
              l2_priority: 3
              code_point_value: 4
          location_identification:
            - loc_info: ecs-elin
              value: 12345678911
              state: present
      local_interface:
        fortyGigE 1/3:
          lldp_state: present
          enable: false
          mode: rx
          multiplier: 3
          hello: 8
          dcbx:
            version: auto
          advertise:
            dcbx_tlv: pfc
            dcbx_tlv_state: present
            dcbx_appln_tlv: fcoe
            dcbx_appln_tlv_state: absent
            dot1_tlv:
              port_tlv:
                protocol_vlan_id: true
                port_vlan_id: true
              vlan_tlv:
                vlan_range: 2-4
                state: present
            dot3_tlv:
              max_frame_size: true
            port_descriptor: true
            management_tlv: management-address system-capabilities
            management_tlv_state: absent
            med:
              application:
                - name: guest-voice
                  vlan_id: 2
                  l2_priority: 3
                  code_point_value: 4
                - name: voice
                  priority_tagged: true
                  l2_priority: 3
                  code_point_value: 4
              location_identification:
                - loc_info: ecs-elin
                  value: 12345678911

**Simple playbook to setup system — leaf.yaml**

    - hosts: leaf1
      roles:
         - dellemc.os9.os9_lldp

**Run**

    ansible-playbook -i hosts leaf.yaml

(c) 2017-2020 Dell Inc. or its subsidiaries. All rights reserved.
