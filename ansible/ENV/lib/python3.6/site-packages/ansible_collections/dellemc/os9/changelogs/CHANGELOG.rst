======================================================================
Ansible Network Collection for Dell EMC OS9 Release Notes
======================================================================

.. contents:: Topics

v1.0.4
======

Release Summary
---------------

- Fixed sanity error found during the sanity tst of automation hub upload
- Fix issue in using list of strings for commands argument for os10_command module (https://github.com/ansible-collections/dellemc.os9/issues/15)

v1.0.3
======

Release Summary
---------------

Added bug fixes for bugs found during System Test.

v1.0.2
======

Release Summary
---------------

Added changelogs.

v1.0.1
======

Release Summary
---------------

Updated documentation review comments.

v1.0.0
======

New Modules
-----------

- os9_command - Run commands on devices running Dell EMC os9.
- os9_config - Manage configuration on devices running os9.
- os9_facts - Collect facts from devices running os9.

New Roles
---------

- os9_aaa - Facilitates the configuration of Authentication Authorization and Accounting (AAA), TACACS and RADIUS server.
- os9_acl - Facilitates the configuration of Access Control lists.
- os9_bgp - Facilitates the configuration of border gateway protocol (BGP) attributes.
- os9_copy_config - This role pushes the backup running configuration into a os9 device.
- os9_dcb - Facilitates the configuration of data center bridging (DCB).
- os9_dns - Facilitates the configuration of domain name service (DNS).
- os9_ecmp - Facilitates the configuration of equal cost multi-path (ECMP) for IPv4.
- os9_interface - Facilitates the configuration of interface attributes.
- os9_lag - Facilitates the configuration of link aggregation group (LAG) attributes.
- os9_lldp - Facilitates the configuration of link layer discovery protocol (LLDP) attributes at global and interface level.
- os9_logging - Facilitates the configuration of global logging attributes and logging servers.
- os9_ntp - Facilitates the configuration of network time protocol (NTP) attributes.
- os9_prefix_list - Facilitates the configuration of IP prefix-list.
- os9_sflow - Facilitates the configuration of global and interface level sFlow attributes.
- os9_snmp - Facilitates the configuration of  global SNMP attributes.
- os9_system - Facilitates the configuration of hostname and hashing algorithm.
- os9_users - Facilitates the configuration of global system user attributes.
- os9_vlan - Facilitates the configuration of virtual LAN (VLAN) attributes.
- os9_vlt - Facilitates the configuration of virtual link trunking (VLT).
- os9_vrf - Facilitates the configuration of virtual routing and forwarding (VRF).
- os9_vrrp - Facilitates the configuration of virtual router redundancy protocol (VRRP) attributes.
- os9_xstp - Facilitates the configuration of xSTP attributes.

\(c) 2020 Dell Inc. or its subsidiaries. All Rights Reserved.
