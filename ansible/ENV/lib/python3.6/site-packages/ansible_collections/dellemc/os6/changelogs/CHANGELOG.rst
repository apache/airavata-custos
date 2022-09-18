======================================================================
Ansible Network Collection for Dell EMC OS6 Release Notes
======================================================================

.. contents:: Topics

v1.0.7
======

Release Summary
---------------

- Fixed sanity error found during the sanity tst of automation hub upload
- os6 interface role readme updated

v1.0.6
======

Bugfixes
---------------

- module utils fix for exit handling in multilevel parent commands 
- config module fix to handle multiline banner
- terminal plugin fix to handle error reported by management access lists

v1.0.5
======

Bugfixes
---------------

- config module fix to handle issues faced while parsing running config and fixing idempotency issue for banner config
- command module change to keep similar changes across all dell networking OSs
- terminal plugin fix to send "terminal length 0" command

v1.0.4
======

Bugfixes
---------------

- Fix issue in using list of strings for `commands` argument for `os6_command` module
- Fix issue in using "os6_facts" module for non-legacy n-series platofrms

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

- os6_command - Run commands on devices running Dell EMC os6.
- os6_config - Manage configuration on devices running os6.
- os6_facts - Collect facts from devices running os6.

New Roles
---------

- os6_aaa - Facilitates the configuration of Authentication Authorization and Accounting (AAA), TACACS and RADIUS server.
- os6_acl - Facilitates the configuration of Access Control lists.
- os6_bgp - Facilitates the configuration of border gateway protocol (BGP) attributes.
- os6_interface - Facilitates the configuration of interface attributes.
- os6_lag - Facilitates the configuration of link aggregation group (LAG) attributes.
- os6_lldp - Facilitates the configuration of link layer discovery protocol (LLDP) attributes at global and interface level.
- os6_logging - Facilitates the configuration of global logging attributes and logging servers.
- os6_ntp - Facilitates the configuration of network time protocol (NTP) attributes.
- os6_qos - Facilitates the configuration of quality of service attributes including policy-map and class-map.
- os6_snmp - Facilitates the configuration of  global SNMP attributes.
- os6_system - Facilitates the configuration of hostname and hashing algorithm.
- os6_users - Facilitates the configuration of global system user attributes.
- os6_vlan - Facilitates the configuration of virtual LAN (VLAN) attributes.
- os6_vrrp - Facilitates the configuration of virtual router redundancy protocol (VRRP) attributes.
- os6_xstp - Facilitates the configuration of xSTP attributes.

\(c) 2020 Dell Inc. or its subsidiaries. All Rights Reserved.
