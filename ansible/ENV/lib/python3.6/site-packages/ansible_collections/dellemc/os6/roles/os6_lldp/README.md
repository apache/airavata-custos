LLDP role
=========

This role facilitates the configuration of link layer discovery protocol (LLDP) attributes at a global and interface level. It supports the configuration of hello, mode, multiplier, advertise TLVs, management interface, FCoE, and iSCSI at global and interface level. This role is abstracted for Dell EMC PowerSwitch platforms running Dell EMC OS6.

The LLDP role requires an SSH connection for connectivity to a Dell EMC OS6 device. You can use any of the built-in OS connection variables.

Role variables
--------------

- If `os6_cfg_generate` is set to true, the variable generates the role configuration commands in a file
- Any role variable with a corresponding state variable set to absent negates the configuration of that variable
- Setting an empty value for any variable negates the corresponding configuration
- Variables and values are case-sensitive

**os6_lldp keys**

| Key        | Type                      | Description                                             | Support               |
|------------|---------------------------|---------------------------------------------------------|-----------------------|
| ``timers`` | dictionary | Configures the LLDP global timer value | os6 |
| ``timers.interval`` | integer | Configures the interval in seconds to transmit local LLDP data (5 to 32768), field needs to be left blank to remove the interval | os6 |
| ``timers.hold`` | integer | Configures the interval multiplier to set local LLDP data TTL (2 to 10), field needs to be left blank to remove the interval multiplier | os6 |
| ``timers.reinit`` | integer | Configures the reinit value (1 to 10), field needs to be left blank to remove the reinit value | os6 |
| ``notification_interval`` | integer | Configures the minimum interval to send remote data change notifications (5 to 3600), field needs to be left blank to remove the minimum interval | os6 |
| ``advertise`` | dictionary     | Configures LLDP-MED and TLV advertisement at the global level (see ``advertise.*``) | os6 |
| ``advertise.med`` | dictionary     | Configures MED TLVs advertisement (see ``med_tlv.*``) | os6 |
| ``med.global_med`` | boolean     | Configures global MED TLVs advertisement | os6 |
| ``med.fast_start_repeat_count`` | integer | Configures MED fast start repeat count value (1 to 10), field needs to be left blank to remove the value | os6 |
| ``med.config_notification`` | boolean | Configure all the ports to send the topology change notification | os6 | 
| ``local_interface`` | dictionary     | Configures LLDP at the interface level (see ``local_interface.*``) | os6 |
| ``local_interface.<interface name>`` | dictionary     | Configures LLDP at the interface level (see ``<interface name>.*``)     | os6 |
| ``<interface name>.mode``  | dictionary: rx,tx   | Configures LLDP mode configuration at the interface level | os6 |
| ``<interface name>.mode.tx``  | boolean | Enables/disables LLDP transmit capability at interface level | os6 |
| ``<interface name>.mode.rx``  | boolean | Enables/disables LLDP receive capability at interface level | os6 |
| ``<interface name>.notification``  | boolean | Enables/disables LLDP remote data change notifications at interface level | os6 |
| ``<interface name>.advertise`` | dictionary     | Configures LLDP-MED TLV advertisement at the interface level (see ``advertise.*``)     | os6 |
| ``advertise.med`` | dictionary     | Configures MED TLVs advertisement at the interface level (see ``med_tlv.*``) | os6 |
| ``med.enable`` | boolean     | Enables interface level MED capabilities | os6 |
| ``med.config_notification`` | boolean     | Configures sending the topology change notification |os6 |


Connection variables
--------------------

Ansible Dell EMC Networking roles require connection information to establish communication with the nodes in your inventory. This information can exist in the Ansible *group_vars* or *host_vars* directories or inventory, or in the playbook itself.

| Key         | Required | Choices    | Description                                         |
|-------------|----------|------------|-----------------------------------------------------|
| ``ansible_host`` | yes      |            | Specifies the hostname or address for connecting to the remote device over the specified transport |
| ``ansible_port`` | no       |            | Specifies the port used to build the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_PORT` option is used; it defaults to 22 |
| ``ansible_ssh_user`` | no       |            | Specifies the username that authenticates the CLI login for the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_USER` environment variable value is used  |
| ``ansible_ssh_pass`` | no       |            | Specifies the password that authenticates the connection to the remote device |
| ``ansible_become`` | no       | yes, no\*   | Instructs the module to enter privileged mode on the remote device before sending any commands; if value is unspecified, the `ANSIBLE_BECOME` environment variable value is used, and the device attempts to execute all commands in non-privileged mode |
| ``ansible_become_method`` | no       | enable, sudo\*   | Instructs the module to allow the become method to be specified for handling privilege escalation; if value is unspecified, the `ANSIBLE_BECOME_METHOD` environment variable value is used |
| ``ansible_become_pass`` | no       |            | Specifies the password to use if required to enter privileged mode on the remote device; if ``ansible_become`` is set to no this key is not applicable |
| ``ansible_network_os`` | yes      | os6, null\*  | Loads the correct terminal and cliconf plugins to communicate with the remote device |

> **NOTE**: Asterisk (\*) denotes the default value if none is specified.

Example playbook
----------------

This example uses the *os6_lldp* role to configure protocol lldp. It creates a *hosts* file with the switch details and corresponding variables. The hosts file should define the `ansible_network_os` variable with corresponding Dell EMC OS6 name. 

When `os6_cfg_generate` is set to true, the variable generates the configuration commands as a .part file in *build_dir* path. By default, the variable is set to false. It writes a simple playbook that only references the *os6_lldp* role.
 
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
    os6_lldp:
      timers:
        reinit: 2
        interval: 5
        hold: 5
      notification_interval: 5
      advertise:
        med:
          global_med: true
          fast_start_repeat_count: 4
          config_notification: true
      local_interface:
        Gi1/0/1:
          mode:
            tx: true
            rx: false
          notification: true
          advertise:
          med:
            config_notification: true
            enable: true


**Simple playbook to setup system — switch1.yaml**

    - hosts: switch1
      roles:
         - dellemc.os6.os6_lldp

**Run**

    ansible-playbook -i hosts switch1.yaml

(c) 2017-2020 Dell Inc. or its subsidiaries. All rights reserved.
