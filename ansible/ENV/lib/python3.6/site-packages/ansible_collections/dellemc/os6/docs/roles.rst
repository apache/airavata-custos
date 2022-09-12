##############################################################
Ansible Network Collection Roles for Dell EMC OS6
##############################################################

The roles facilitate provisioning of Dell EMC PowerSwitch platforms running Dell EMC OS6. 

AAA role
********

The `os6_aaa <https://github.com/ansible-collections/dellemc.os6/blob/master/roles/os6_aaa/README.md>`_ role facilitates the configuration of authentication, authorization, and acccounting (AAA). It supports the configuration of TACACS and RADIUS server, and AAA.


ACL role
********

The `os6_acl <https://github.com/ansible-collections/dellemc.os6/blob/master/roles/os6_acl/README.md>`_ role facilitates the configuration of an access-control list (ACL). It supports the configuration of different types of ACLs (standard and extended) for both IPv4 and IPv6, and assigns the access-class to line terminals.


BGP role
********

The `os6_bgp <https://github.com/ansible-collections/dellemc.os6/blob/master/roles/os6_bgp/README.md>`_ role facilitates the configuration of border gateway protocol (BGP) attributes. It supports the configuration of router ID, networks, neighbors, and maximum path.


Interface role
**************

The `os6_interface <https://github.com/ansible-collections/dellemc.os6/blob/master/roles/os6_interface/README.md>`_ role facilitates the configuration of interface attributes. It supports the configuration of administrative state, description, MTU, IP address, IP helper, and port mode. 


LAG role
********

The `os6_lag <https://github.com/ansible-collections/dellemc.os6/blob/master/roles/os6_lag/README.md>`_ role facilitates the configuration of link aggregation group (LAG) attributes, and supports the creation and deletion of a LAG and its member ports. It also supports the configuration of type (static/dynamic), hash scheme, and minimum required link.


LLDP role
*********

The `os6_lldp <https://github.com/ansible-collections/dellemc.os6/blob/master/roles/os6_lldp/README.md>`_ role facilitates the configuration of link layer discovery protocol (LLDP) attributes at global and interface level. This role supports the configuration of hello, mode, multiplier, advertise tlvs, management interface, fcoe, iscsi at global and interface levels.


Logging role
************

The `os6_logging <https://github.com/ansible-collections/dellemc.os6/blob/master/roles/os6_logging/README.md>`_ role facilitates the configuration of global logging attributes, and supports the configuration of logging servers.


NTP role
********

The `os6_ntp <https://github.com/ansible-collections/dellemc.os6/blob/master/roles/os6_ntp/README.md>`_ role facilitates the configuration of network time protocol (NTP) attributes.


QoS role
********

The `os6_qos <https://github.com/ansible-collections/dellemc.os6/blob/master/roles/os6_qos/README.md>`_ role facilitates the configuration of quality of service (QoS) attributes including policy-map and class-map.


SNMP role
*********

The `os6_snmp <https://github.com/ansible-collections/dellemc.os6/blob/master/roles/os6_snmp/README.md>`_ role facilitates the configuration of global simple network management protocol (SNMP) attributes. It supports the configuration of SNMP server attributes like users, group, community, location, and traps.


System role
***********

The `os6_system <https://github.com/ansible-collections/dellemc.os6/blob/master/roles/os6_system/README.md>`_ role facilitates the configuration of global system attributes. This role specifically enables configuration of hostname and enable password for OS6.


Users role
**********

The `os6_users <https://github.com/ansible-collections/dellemc.os6/blob/master/roles/os6_users/README.md>`_ role facilitates the configuration of global system user attributes. This role supports the configuration of CLI users.


VLAN role
*********

The `os6_vlan <https://github.com/ansible-collections/dellemc.os6/blob/master/roles/os6_vlan/README.md>`_ role facilitates configuring virtual LAN (VLAN) attributes. This role supports the creation and deletion of a VLAN and its member ports.


VRRP role
*********

The `os6_vrrp <https://github.com/ansible-collections/dellemc.os6/blob/master/roles/os6_vrrp/README.md>`_ role facilitates configuration of virtual router redundancy protocol (VRRP) attributes. This role supports the creation of VRRP groups for interfaces, and setting the VRRP group attributes.


xSTP role
*********

The `os6_xstp <https://github.com/ansible-collections/dellemc.os6/blob/master/roles/os6_xstp/README.md>`_ role facilitates the configuration of extended spanning-tree protocol (xSTP) attributes. This role supports multiple version of spanning-tree protocol (STP), rapid spanning-tree (RSTP) protocol, multiple spanning-tree (MST), and per-VLAN spanning-tree (PVST). This role supports the configuration of bridge priority, enabling and disabling spanning-tree, creating and deleting instances, and mapping virtual LAN (VLAN) to instances.