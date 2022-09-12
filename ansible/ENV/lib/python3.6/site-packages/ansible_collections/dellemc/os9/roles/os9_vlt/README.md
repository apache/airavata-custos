VLT role
========

This role facilitates the configuration of the basics of virtual link trunking (VLT) to provide a loop-free topology. This role is abstracted for Dell EMC PowerSwitch platforms running Dell EMC OS9.

The VLT role requires an SSH connection for connectivity to your Dell EMC OS9 device. You can use any of the built-in OS connection variables.

Role variables
--------------

- Role is abstracted using the `ansible_network_os` variable  that can take `dellemc.os9.os9` as the value  
- If `os9_cfg_generate` is set to true, the variable generates the role configuration commands in a file
- Any role variable with a corresponding state variable set to absent negates the configuration of that variable
- Setting an empty value for any variable negates the corresponding configuration
- Variables and values are case-sensitive

**os9_vlt keys**

| Key        | Type                      | Description                                             | Support               |
|------------|---------------------------|---------------------------------------------------------|-----------------------|
| ``domain``       | integer (required)        | Configures the VLT domain identification number (1 to 1000) | os9 |
| ``backup_destination`` | string    | Configures an IPv4 address for the VLT backup link (A.B.C.D format or X:X:X:X::X format) | os9  |
| ``destination_type``  | string    | Configures the backup destination based on this destination type (IPv4 or IPv6)| os9 |
| ``backup_destination_vrf``      | string         | Configures the virtual routing and forwarding (VRF) instance through which the backup destination IP is reachable (*vrfname* must be present)  | os9|
| ``VLTi``    | integer        | Configures the peer link port-channel ID for the VLT domain (1 to 4096) | os9 |
| ``peer_routing``     | boolean        | Configures VLT peer routing | os9 |
| ``peer_routing_timeout``     | integer        | Configures the timeout for peer routing (1 to 65535)| os9 |
| ``multicast_peer_routing_timeout``     | integer        | Configures the timeout for multicast peer routing (1 to 1200) | os9 |
| ``priority``     | integer        | Configures the primary priority to the corresponding channel ID | os9 |
| ``unit_id``       | integer        | Configures the system unit ID for VLT (either 0 or 1) | os9 |
| ``vlt_peers``     | dictionary       | Contains objects to configure the VLT peer port-channel (see ``vlt_peers.*``)  | os9 |
| ``vlt_peers.<portchannelid>``    | dictionary       | Configures the VLT peer port-channel (`Po <portchannelid> value`) | os9 |
| ``vlt_peers.<portchannelid>.peer_lag``     | integer       | Configures the port-channel ID of the VLT peer lag  |  os9 |
| ``system_mac``     | string        | Configures the system MAC address for VLT  | os9 |
| ``delay_restore``     | integer       | Configures the delay in bringing up VLT ports after reload or peer-link restoration (default 90)| os9 |
| ``delay_restore_abort_threshold``     | integer        | Configures the wait interval for VLT delay-restore timer to abort (default 60) | os9 |
| ``proxy_gateway``     | dictionary       | Contains objects to configure the VLT proxy gateway (see ``proxy_gateway.*``)  | os9 |
| ``proxy_gateway.static``     | dictionary       | Contains objects to configure the static VLT proxy gateway (see ``static.*``)  | os9 |
| ``static.remote_mac``     | list      | Configures the remote MAC for static VLT proxy gateway (see ``remote_mac.*``)  | os9 |
| ``remote_mac.address``     | string        | Configures the remote MAC address for the static VLT proxy gateway  | os9 |
| ``remote_mac.exclude_vlan_range``     | string        | Configures the exclude VLAN for the static VLT proxy gateway  | os9 |
| ``remote_mac.state``        | string: absent,present     | Deletes the remote MAC address or exclude VLAN configured on the proxy gateway if set to absent | os9 |
| ``static.proxy_static_state``        | string: absent,present     | Deletes the static VLT proxy gateway if set to absent | os9 |
| ``proxy_gateway.lldp``     | dictionary       | Contains objects to configure LLDP VLT proxy gateway (see ``lldp.*`` for each item); mutually exclusive with *proxy_gateway.static* | os9 |
| ``lldp.peer_domain_link``     | list      | Configures the VLT proxy gateway interface (see ``peer_domain_link.*``)  | os9 |
| ``peer_domain_link.port_channel_id``     | integer        | Configures the port-channel for the VLT proxy gateway  | os9 |
| ``peer_domain_link.exclude_vlan_range``     | string        | Configures to exclude VLAN for LLDP VLT proxy gateway  | os9 |
| ``peer_domain_link.state``        | string: absent,present     | Deletes the port-channel or exclude VLAN configured on the proxy gateway if set to absent | os9 |
| ``lldp.proxy_lldp_state``        | string: absent,present     | Deletes the LLDP VLT proxy gateway if set to absent | os9 |
| ``lldp.vlt_peer_mac``     | boolean        | Configures the proxy gateway transmit for square VLT | os9 |
| ``lldp.peer_timeout``     | integer        | Configures the proxy gateway restore timer (1 to 65535) | os9 |
| ``state``        | string: absent,present     | Deletes the VLT instance if set to absent | os9 |

> **NOTE**: Asterisk (\*) denotes the default value if none is specified.

Connection variables
--------------------

Ansible Dell EMC network OS roles require connection information to establish communication with the nodes in your inventory. This information can exist in the Ansible *group_vars* or *host_vars* directories or inventory, or in the playbook itself.

| Key         | Required | Choices    | Description                                         |
|-------------|----------|------------|-----------------------------------------------------|
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

This example uses the *os9_vlt* role to setup a VLT-domain. It creates a *hosts* file with the switch details and corresponding variables.The hosts file should define the `ansible_network_os` variable with corresponding Dell EMC OS9 name.

When `os9_cfg_generate` is set to true, the variable generates the configuration commands as a .part file in *build_dir* path. By default, the variable is set to false. It writes a simple playbook that only references the *os9_vlt* role.

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

    os9_vlt:
        domain: 1
        backup_destination: 192.168.1.1
        destination_type: "ipv4"
        priority: 1
        VLTi: 101
        backup_destination_vrf: VLTi-KEEPALIVE
        peer_routing: true
        peer_routing_timeout: 200
        multicast_peer_routing_timeout: 250
        unit_id: 0
        system_mac: aa:aa:aa:aa:aa:aa
        delay_restore: 100
        delay_restore_abort_threshold: 110
        proxy_gateway:
          static:
            remote_mac:
              - address: aa:aa:aa:aa:aa:aa
                exclude_vlan_range: 2
                state: present
            proxy_static_state: present
        vlt_peers:
          Po 12:
            peer_lag: 13
        state: present

**Simple playbook to setup system — leaf.yaml**

    - hosts: leaf1
      roles:
         - dellemc.os9.os9_vlt

**Run**

    ansible-playbook -i hosts leaf.yaml

(c) 2017-2020 Dell Inc. or its subsidiaries. All rights reserved.
