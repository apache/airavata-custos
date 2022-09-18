BGP role
========

This role facilitates the configuration of border gateway protocol (BGP) attributes. It supports the configuration of router ID, networks, neighbors, and maximum path. This role is abstracted for Dell EMC PowerSwitch platforms running Dell EMC OS6.

The BGP role requires an SSH connection for connectivity to a Dell EMC OS6 device. You can use any of the built-in OS connection variables.


Role variables
--------------
 
- Role is abstracted using the `ansible_network_os` variable that can take `dellemc.os6.os6` as a value
- If variable `os6_cfg_generate` is set to true, it generates the role configuration commands in a file
- Any role variable with a corresponding state variable setting to absent negates the configuration of that variable
- Setting an empty value for any variable negates the corresponding configuration
- Variables and values are case-sensitive

> **NOTE**: IP routing needs to be enabled on the switch prior to configuring BGP via the *os6_bgp* role.

**os6_bgp keys**

| Key        | Type                      | Description                                             | Support               |
|------------|---------------------------|---------------------------------------------------------|-----------------------|
| ``asn`` | string (required) | Configures the autonomous system (AS) number of the local BGP instance | os6 |
| ``router_id`` | string | Configures the IP address of the local BGP router instance | os6 |
| ``maxpath_ibgp`` | integer | Configures the maximum number of paths to forward packets through iBGP (1 to 64; default 1) | os6 |
| ``maxpath_ebgp`` | integer | Configures the maximum number of paths to forward packets through eBGP (1 to 64; default 1) | os6 |
| ``ipv4_network`` | list | Configures an IPv4 BGP networks (see ``ipv4_network.*``) | os6 |
| ``ipv4_network.address`` | string (required)         | Configures the IPv4 address of the BGP network (A.B.C.D/E format)   | os6 |
| ``ipv4_network.state`` | string: absent,present\* | Deletes an IPv4 BGP network if set to absent | os6 |
| ``ipv6_network`` | list | Configures an IPv6 BGP network (see ``ipv6_network.*``) | os6 |
| ``ipv6_network.address`` | string (required)         | Configures the IPv6 address of the BGP network (2001:4898:5808:ffa2::1/126 format)  | os6 |
| ``ipv6_network.state`` | string: absent,present\* | Deletes an IPv6 BGP network if set to absent | os6 |
| ``neighbor`` | list | Configures IPv4 BGP neighbors (see ``neighbor.*``) | os6 |
| ``neighbor.ip`` | string (required)         | Configures the IPv4 address of the BGP neighbor (10.1.1.1)  | os6 |
| ``neighbor.name`` | string (required)         | Configures the BGP peer-group with this name; supported only when the neighbor is a peer group; mutually exclusive with *neighbor.ip* | os6 |
| ``neighbor.type`` | string (required): ipv4,ipv6,peergroup       | Specifies the BGP neighbor type   | os6 |
| ``neighbor.remote_asn`` | string (required)         | Configures the remote AS number of the BGP neighbor  | os6 |
| ``neighbor.remote_asn_state`` | string: absent,present\* | Deletes the remote AS number from the peer group if set to absent; supported only when *neighbor.type* is "peergroup" | os6 |
| ``neighbor.timer`` | string          | Configures neighbor timers (<int> <int>); 5 10, where 5 is the keepalive interval and 10 is the holdtime, field needs to be left blank to remove the timer configurations | os6 |
| ``neighbor.default_originate`` | boolean: true, false\*     | Configures default originate routes to the BGP neighbor, field needs to be left blank to remove the default  originate routes | os6 | 
| ``neighbor.peergroup`` | string          | Configures neighbor to BGP peer-group (configured peer-group name) | os6 |
| ``neighbor.peergroup_state`` | string: absent,present\* | Deletes the IPv4 BGP neighbor from the peer-group if set to absent | os6 |
| ``neighbor.admin`` | string: up,down       | Configures the administrative state of the neighbor  | os6 |
| ``neighbor.src_loopback`` | integer         | Configures the source loopback interface for routing packets | os6  |
| ``neighbor.src_loopback_state`` | string: absent,present\* | Deletes the source for routing packets if set to absent                 | os6 |
| ``neighbor.ebgp_multihop`` | integer | Configures the maximum-hop count value allowed in eBGP neighbors that are not directly connected (default 255), field needs to be left blank to remove the maximum hop count value | os6 |               
| ``neighbor.subnet`` | string (required)         | Configures the passive BGP neighbor to this subnet | os6 |
| ``neighbor.subnet_state`` | string: absent,present\* | Deletes the subnet range set for dynamic IPv4 BGP neighbor if set to absent            | os6 |
| ``neighbor.state`` | string: absent,present\* | Deletes the IPv4 BGP neighbor if set to absent | os6 |
| ``redistribute`` | list | Configures the redistribute list to get information from other routing protocols (see ``redistribute.*``) | os6 |
| ``redistribute.route_type`` | string (required): static,connected        | Configures the name of the routing protocol to redistribute | os6 |
| ``redistribute.address_type`` | string (required): ipv4,ipv6                  | Configures the address type of IPv4 or IPv6 routes | os6 |
| ``redistribute.state`` | string: absent,present\* | Deletes the redistribution information if set to absent | os6 |
| ``state`` |  string: absent,present\*    | Deletes the local router BGP instance if set to absent      | os6 |

> **NOTE**: Asterisk (\*) denotes the default value if none is specified.

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
| ``ansible_become_method`` | no       | enable, sudo\*   | Instructs the module to allow the become method to be specified for handling privilege escalation; if value is unspecified, the `ANSIBLE_BECOME_METHOD` environment variable value is used. |
| ``ansible_become_pass`` | no       |            | Specifies the password to use if required to enter privileged mode on the remote device; if ``ansible_become`` is set to no this key is not applicable. |
| ``ansible_network_os`` | yes      | os6, null\*  | This value is used to load the correct terminal and cliconf plugins to communicate with the remote device. |

> **NOTE**: Asterisk (\*) denotes the default value if none is specified.

Example playbook
----------------

This example uses the *os6_bgp* role to configure the BGP network and neighbors. It creates a *hosts* file with the switch details, a *host_vars* file with connection variables and the corresponding role variables.

When `os6_cfg_generate` is set to true, the variable generates the configuration commands as a .part file in *build_dir* path. By default, the variable is set to false. This example writes a simple playbook that only references the *os6_bgp* role. 

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
	  
    os6_bgp:
        asn: 11
        router_id: 192.168.3.100
        maxpath_ibgp: 2
        maxpath_ebgp: 2
        ipv4_network:
          - address: 102.1.1.0 255.255.255.255
            state: present
        ipv6_network:
          - address: "2001:4898:5808:ffa0::/126"
            state: present
        neighbor:
          - ip: 192.168.10.2
            type: ipv4
            remote_asn: 12
            timer: 5 10
            default_originate: False
            peergroup: per
            admin: up
            state: present
          - ip: 2001:4898:5808:ffa2::1
            type: ipv6
            remote_asn: 14
            peergroup: per
            state: present
          - name: peer1
            type: peergroup
            remote_asn: 14
            ebgp_multihop: 4
            subnet: 10.128.5.192/27
            state: present
          - ip: 172.20.12.1
            type: ipv4
            remote_asn: 64640
            timer: 3 9
        redistribute:
          - route_type: static
            address_type: ipv4
            state: present
          - route_type: connected
            address_type: ipv6
            state: present
        state: present

**Simple playbook to configure BGP — switch1.yaml**

    - hosts: switch1
      roles:
         - dellemc.os6.os6_bgp

**Run**

    ansible-playbook -i hosts switch1.yaml

(c) 2017-2020 Dell Inc. or its subsidiaries. All rights reserved.
