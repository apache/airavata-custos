##############################################################
Ansible Network Collection Roles for Dell EMC OS9
##############################################################

The roles facilitate provisioning of devices running Dell EMC OS9. This document describes each of the roles.

AAA role
--------

The `os9_aaa <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_aaa/README.md>`_ role facilitates the configuration of authentication, authorization, and acccounting (AAA). It supports the configuration of TACACS and RADIUS server and AAA.


ACL role
--------

The `os9_acl <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_acl/README.md>`_ role facilitates the configuration of an Access Control list (ACL). It supports the configuration of different types of ACLs (standard and extended) for both IPv4 and IPv6, and assigns the access-class to line terminals.


BGP role
--------

The `os9_bgp <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_bgp/README.md>`_ role facilitates the configuration of border gateway protocol (BGP) attributes. It supports the configuration of router ID, networks, neighbors, and maximum path.


Copy configuration role
-----------------------

The `os9_copy_config <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_copy_config/README.md>`_ role pushes the backup running configuration into a device. This role merges the configuration in the template file with the running configuration of the Dell EMC Networking OS9 device.


DCB role
--------

The `os9_dcb <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_dcb/README.md>`_ role facilitates the configuration of data center bridging (DCB). It supports the configuration of the DCB map and the DCB buffer, and assigns them to interfaces.


DNS role
--------

The `os9_dns <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_dns/README.md>`_ role facilitates the configuration of domain name service (DNS).


ECMP role
---------

The `os9_ecmp <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_ecmp/README.md>`_ role facilitates the configuration of equal cost multi-path (ECMP). It supports the configuration of ECMP for IPv4.


Interface role
--------------

The `os9_interface <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_interface/README.md>`_ role facilitates the configuration of interface attributes. It supports the configuration of administrative state, description, MTU, IP address, IP helper, suppress_ra and port mode. 


LAG role
--------

The `os9_lag <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_lag/README.md>`_ role facilitates the configuration of link aggregation group (LAG) attributes, and supports the creation and deletion of a LAG and its member ports. It also supports the configuration of an interface type (static/dynamic) and minimum required link.


LLDP role
---------

The `os9_lldp <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_lldp/README.md>`_ role facilitates the configuration of link layer discovery protocol (LLDP) attributes at global and interface level. This role supports the configuration of hello, mode, multiplier, advertise tlvs, management interface, fcoe, iscsi at global and interface levels.


Logging role
------------

The `os9_logging <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_logging/README.md>`_ role facilitates the configuration of global logging attributes, and supports the configuration of logging servers.


NTP role
--------

The `os9_ntp <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_ntp/README.md>`_ role facilitates the configuration of network time protocol attributes.


Prefix-list role
----------------

The `os9_prefix_list <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_prefix_list/README.md>`_ role facilitates the configuration of a prefix-list, supports the configuration of IP prefix-list, and assigns the prefix-list to line terminals.


sFlow role
----------

The `os9_sflow <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_sflow/README.md>`_ role facilitates the configuration of global and interface level sFlow attributes. It supports the configuration of sFlow collectors at the global level, enable/disable, and specification of sFlow polling-interval, sample-rate, max-datagram size, and so on are supported at the interface and global level.


SNMP role
---------

The `os9_snmp <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_snmp/README.md>`_ role facilitates the configuration of global snmp attributes. It supports the configuration of SNMP server attributes like users, group, community, location, and traps.


System role
-----------

The `os9_system <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_system/README.md>`_ role facilitates the configuration of global system attributes. This role specifically enables configuration of hostname and enable password for os9. It also supports the configuration of management route, hash alogrithm, clock, line terminal, banner, and reload type.


Users role
----------

The `os9_users <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_users/README.md>`_ role facilitates the configuration of global system user attributes. This role supports the configuration of CLI users.


VLAN role
---------

The `os9_vlan <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_vlan/README.md>`_ role facilitates configuring virtual LAN (VLAN) attributes. This role supports the creation and deletion of a VLAN and its member ports.


VLT role
--------

The `os9_vlt <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_vlt/README.md>`_ role facilitates the configuration of the basics of virtual link trunking (VLT) to provide a loop-free topology.


VRF role
--------

The `os9_vrf <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_vrf/README.md>`_ role facilitates the configuration of basic virtual routing and forwarding (VRF) that helps in the partition of physical routers to multiple virtual routers.


VRRP role
---------

The `os9_vrrp <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_vrrp/README.md>`_ role facilitates configuration of virtual router redundancy protocol (VRRP) attributes. This role supports the creation of VRRP groups for interfaces, and setting the VRRP group attributes.


xSTP role
---------

The `os9_xstp <https://github.com/ansible-collections/dellemc.os9/blob/master/roles/os9_xstp/README.md>`_ role facilitates the configuration of xSTP attributes. This role supports multiple version of spanning-tree protocol (STP), rapid spanning-tree (RSTP) protocol, multiple spanning-tree (MST), and per-VLAN spanning-tree (PVST). This role supports the configuration of bridge priority, enabling and disabling spanning-tree, creating and deleting instances, and mapping virtual LAN (VLAN) to instances.
