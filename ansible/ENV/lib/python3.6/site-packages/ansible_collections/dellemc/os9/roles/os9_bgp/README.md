BGP role
========

This role facilitates the configuration of border gateway protocol (BGP) attributes. It supports the configuration of router ID, networks, neighbors, and maximum path. This role is abstracted for Dell EMC PowerSwitch platforms running Dell EMC OS9.

The BGP role requires an SSH connection for connectivity to a Dell EMC OS9 device. You can use any of the built-in OS connection variables.


Role variables
--------------
 
- Role is abstracted using the `ansible_network_os` variable that can take `dellemc.os9.os9` as the value
- If `os9_cfg_generate` is set to true, it generates the role configuration commands in a file
- Any role variable with a corresponding state variable setting to absent negates the configuration of that variable
- Setting an empty value for any variable negates the corresponding configuration
- Variables and values are case-sensitive

**os9_bgp keys**

| Key        | Type                      | Description                                             | Support               |
|------------|---------------------------|---------------------------------------------------------|-----------------------|
| ``asn`` | string (required) | Configures the autonomous system (AS) number of the local BGP instance | os9 |
| ``router_id`` | string | Configures the IP address of the local BGP router instance | os9 |
| ``graceful_restart`` | boolean | Configures graceful restart capability | os9 |
| ``graceful_restart.state`` | string: absent,present\* | Removes graceful restart capability if set to absent | os9 |
| ``maxpath_ibgp`` | integer | Configures the maximum number of paths to forward packets through iBGP (1 to 64; default 1) | os9 |
| ``maxpath_ebgp`` | integer | Configures the maximum number of paths to forward packets through eBGP (1 to 64; default 1) | os9 |
| ``best_path`` | list | Configures the default best-path selection (see ``best_path.*``) | os9 |
| ``best_path.as_path`` | string (required): ignore,multipath-relax     | Configures the AS path used for the best-path computation   | os9 |
| ``best_path.as_path_state`` | string: absent,present\*     | Deletes the AS path configuration if set to absent  | os9 |
| ``best_path.ignore_router_id`` | boolean: true,false | Ignores the router identifier in best-path computation if set to true | os9 |
| ``best_path.med`` | list | Configures the MED attribute (see ``med.*``) | os9 |
| ``med.attribute`` | string (required): confed,missing-as-best     | Configures the MED attribute used for the best-path computation   | os9 |
| ``med.state`` | string: absent,present\* | Deletes the MED attribute if set to absent | os9,  |
| ``ipv4_network`` | list | Configures an IPv4 BGP networks (see ``ipv4_network.*``) | , os9,  |
| ``ipv4_network.address`` | string (required)         | Configures the IPv4 address of the BGP network (A.B.C.D/E format)   | os9 |
| ``ipv4_network.state`` | string: absent,present\* | Deletes an IPv4 BGP network if set to absent | os9 |
| ``ipv6_network`` | list | Configures an IPv6 BGP network (see ``ipv6_network.*``) | os9 |
| ``ipv6_network.address`` | string (required)         | Configures the IPv6 address of the BGP network (2001:4898:5808:ffa2::1/126 format)  | os9 |
| ``ipv6_network.state`` | string: absent,present\* | Deletes an IPv6 BGP network if set to absent | os9 |
| ``neighbor`` | list | Configures IPv4 BGP neighbors (see ``neighbor.*``) | os9 |
| ``neighbor.ip`` | string (required)         | Configures the IPv4 address of the BGP neighbor (10.1.1.1)  | os9 |
| ``neighbor.interface`` | string      | Configures the BGP neighbor interface details |   |
| ``neighbor.name`` | string (required)         | Configures the BGP peer-group with this name; supported only when the neighbor is a peer group; mutually exclusive with *neighbor.ip* | os9 |
| ``neighbor.type`` | string (required): ipv4,ipv6,peergroup       | Specifies the BGP neighbor type   | os9 |
| ``neighbor.remote_asn`` | string (required)         | Configures the remote AS number of the BGP neighbor  | os9 |
| ``neighbor.remote_asn_state`` | string: absent,present\* | Deletes the remote AS number from the peer group if set to absent; supported only when *neighbor.type* is "peergroup" | os9 |
| ``neighbor.timer`` | string          | Configures neighbor timers (<int> <int>); 5 10, where 5 is the keepalive interval and 10 is the holdtime | os9 |
| ``neighbor.default_originate`` | boolean: true, false\*     | Configures default originate routes to the BGP neighbor | os9 | 
| ``neighbor.peergroup`` | string          | Configures neighbor to BGP peer-group (configured peer-group name) | os9 |
| ``neighbor.peergroup_state`` | string: absent,present\* | Deletes the IPv4 BGP neighbor from the peer-group if set to absent | os9 |
| ``neighbor.distribute_list`` | list | Configures the distribute list to filter networks from routing updates (see ``distribute_list.*``) | os9 |
| ``distribute_list.in`` | string       | Configures the name of the prefix-list to filter incoming packets  | os9 |
| ``distribute_list.in_state`` | string: absent,present\* | Deletes the filter at incoming packets if set to absent           | os9 |
| ``distribute_list.out`` | string       | Configures the name of the prefix-list to filter outgoing packets   | os9 |
| ``distribute_list.out_state`` | string: absent,present\* | Deletes the filter at outgoing packets if set to absent          | os9 |
| ``neighbor.admin`` | string: up,down       | Configures the administrative state of the neighbor  | os9 |
| ``neighbor.adv_interval`` | integer       | Configures the advertisement interval of the neighbor  | os9 |
| ``neighbor.fall_over`` | string: absent,present       | Configures the session fall on peer-route loss  | os9 |
| ``neighbor.sender_loop_detect`` | boolean: true,false         | Enables/disables the sender-side loop detect for neighbors | os9 |
| ``neighbor.src_loopback`` | integer         | Configures the source loopback interface for routing packets | os9 |
| ``neighbor.src_loopback_state`` | string: absent,present\* | Deletes the source for routing packets if set to absent                 | os9 |
| ``neighbor.ebgp_multihop`` | integer | Configures the maximum-hop count value allowed in eBGP neighbors that are not directly connected (default 255) | os9 |
| ``neighbor.passive`` | boolean: true,false\*     | Configures the passive BGP peer group; supported only when neighbor is a peer-group | os9 |                 
| ``neighbor.subnet`` | string (required)         | Configures the passive BGP neighbor to this subnet; required together with the *neighbor.passive* key for os9 devices | , os9,  |
| ``neighbor.subnet_state`` | string: absent,present\* | Deletes the subnet range set for dynamic IPv4 BGP neighbor if set to absent            | os9 |
| ``neighbor.limit`` | integer    | Configures maximum dynamic peers count (key is required together with ``neighbor.subnet``) |  |
| ``neighbor.bfd`` | boolean | Enables BDF for neighbor |  |
| ``neighbor.state`` | string: absent,present\* | Deletes the IPv4 BGP neighbor if set to absent | os9 |
| ``redistribute`` | list | Configures the redistribute list to get information from other routing protocols (see ``redistribute.*``) | os9 |
| ``redistribute.route_type`` | string (required): static,connected        | Configures the name of the routing protocol to redistribute | os9 |
| ``redistribute.route_map_name`` | string        | Configures the route-map to redistribute | os9 |
| ``redistribute.route_map`` |  string: absent,present\*    | Deletes the route-map to redistribute if set to absent        | os9 |
| ``redistribute.address_type`` | string (required): ipv4,ipv6                  | Configures the address type of IPv4 or IPv6 routes | os9 |
| ``redistribute.state`` | string: absent,present\* | Deletes the redistribution information if set to absent | os9 |
| ``state`` |  string: absent,present\*    | Deletes the local router BGP instance if set to absent      | os9 |

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
| ``ansible_become_method`` | no       | enable, sudo\*   | Instructs the module to allow the become method to be specified for handling privilege escalation; if value is unspecified, the `ANSIBLE_BECOME_METHOD` environment variable value is used |
| ``ansible_become_pass`` | no       |            | Specifies the password to use if required to enter privileged mode on the remote device; if ``ansible_become`` is set to no this key is not applicable |
| ``ansible_network_os`` | yes      | os9, null\*  | Loads the correct terminal and cliconf plugins to communicate with the remote device |

> **NOTE**: Asterisk (\*) denotes the default value if none is specified.

Example playbook
----------------

This example uses the *os9_bgp* role to configure the BGP network and neighbors. The example creates a hosts file with the switch details and corresponding variables. The hosts file should define the `ansible_network_os` variable with the corresponding Dell EMC OS9 name.

When `os9_cfg_generate` is set to true, the variable generates the configuration commands as a .part file in *build_dir* path. By default, the variable is set to false. This example writes a simple playbook that only references the *os9_bgp* role. The sample host_vars given below is for os9. 

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
	  
    os9_bgp:
        asn: 11
        router_id: 192.168.3.100
        maxpath_ibgp: 2
        maxpath_ebgp: 2
        graceful_restart: true
        best_path:
           as_path: ignore
           ignore_router_id: true
           med:
            - attribute: confed
              state: present
            - attribute: missing-as-best
              state: present
        ipv4_network:
          - address: 102.1.1.0/30
            state: present
        ipv6_network:
          - address: "2001:4898:5808:ffa0::/126"
            state: present
        neighbor:
          - ip: 192.168.10.2
            type: ipv4
            remote_asn: 12
            timer: 5 10
            adv_interval: 40
            fall_over: present
            default_originate: False
            peergroup: per
            peergroup_state: present
            sender_loop_detect: false
            src_loopback: 1
            src_loopback_state: present
            distribute_list:
               in: aa
               in_state: present
            ebgp_multihop: 25
            admin: up
            state: present
          - ip: 2001:4898:5808:ffa2::1
            type: ipv6
            remote_asn: 14
            peergroup: per
            peergroup_state: present
            distribute_list:
               in: aa
               in_state: present
            src_loopback: 0
            src_loopback_state: present
            ebgp_multihop: 255
            admin: up
            state: present
          - name: peer1
            type: peergroup
            remote_asn: 14
            distribute_list:
               in: an
               in_state: present
               out: bb
               out_state: present
            passive: True
            subnet: 10.128.4.192/27
            subnet_state: present
            state: present
          - ip: 172.20.12.1
            description: O_site2-spine1
            type: ipv4
            remote_asn: 64640
            fall_over: present
            ebgp_multihop: 4
            src_loopback: 1
            adv_interval: 1
            timer: 3 9
            send_community:
              - type: extended
            address_family:
              - type: ipv4
                activate: falsesrc_loopback
                state: present
              - type: l2vpn
                activate: true
                state: present
            admin: up
            state: present
        redistribute:
          - route_type: static
            route_map_name: aa
            state: present
            address_type: ipv4
          - route_type: connected
            address_type: ipv6
            state: present
        state: present

**Simple playbook to configure BGP — leaf.yaml**

    - hosts: leaf1
      roles:
         - dellemc.os9.os9_bgp

**Run**

    ansible-playbook -i hosts leaf.yaml

(c) 2017-2020 Dell Inc. or its subsidiaries. All rights reserved.
