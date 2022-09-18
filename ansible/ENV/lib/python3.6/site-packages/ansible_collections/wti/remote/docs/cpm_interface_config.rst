
cpm_interface_config -- Set network interface parameters in WTI OOB and PDU devices
===================================================================================

.. contents::
   :local:
   :depth: 1


Synopsis
--------

Set network interface parameters in WTI OOB and PDU devices






Parameters
----------

  cpm_url (True, str, None)
    This is the URL of the WTI device to send the module.


  cpm_username (True, str, None)
    This is the Username of the WTI device to send the module.


  cpm_password (True, str, None)
    This is the Password of the WTI device to send the module.


  use_https (False, bool, True)
    Designates to use an https connection or http connection.


  validate_certs (False, bool, True)
    If false, SSL certificates will not be validated. This should only be used

    on personally controlled sites using self-signed certificates.


  use_proxy (False, bool, False)
    Flag to control if the lookup will observe HTTP proxy environment variables when present.


  interface (False, str, None)
    This is the ethernet port name that is getting configured.


  negotiation (False, int, None)
    This is the speed of the interface port being configured.

    0=Auto, 1=10/half, 2=10/full, 3=100/half, 4=100/full, 5=1000/half, 6=1000/full


  ipv4address (False, str, None)
    IPv4 format IP address for the defined interface Port.


  ipv4netmask (False, str, None)
    IPv4 format Netmask for the defined interface Port.


  ipv4gateway (False, str, None)
    IPv4 format Gateway address for the defined interface Port.


  ipv4dhcpenable (False, int, None)
    Enable IPv4 DHCP request call to obtain confufuration information.


  ipv4dhcphostname (False, str, None)
    Define IPv4 DHCP Hostname.


  ipv4dhcplease (False, int, None)
    IPv4 DHCP Lease Time.


  ipv4dhcpobdns (False, int, None)
    IPv6 DHCP Obtain DNS addresses auto.


  ipv4dhcpupdns (False, int, None)
    IPv4 DHCP DNS Server Update.


  ipv4dhcpdefgateway (False, int, None)
    Enable or Disable this ports configuration as the default IPv4 route for the device.


  ipv6address (False, str, None)
    IPv6 format IP address for the defined interface Port.


  ipv6subnetprefix (False, str, None)
    IPv6 format Subnet Prefix for the defined interface Port.


  ipv6gateway (False, str, None)
    IPv6 format Gateway address for the defined interface Port.





Notes
-----

.. note::
   - Use ``groups/cpm`` in ``module_defaults`` to set common options used between CPM modules.




Examples
--------

.. code-block:: yaml+jinja

    
    # Set Network Interface Parameters
    - name: Set the Interface Parameters for port eth1 of a WTI device
      cpm_interface_config:
        cpm_url: "nonexist.wti.com"
        cpm_username: "super"
        cpm_password: "super"
        use_https: true
        validate_certs: false
        interface: "eth1"
        ipv4address: "192.168.0.14"
        ipv4netmask: "255.255.255.0"
        ipv4gateway: "192.168.0.1"
        negotiation: 0

    # Set Network Interface Parameters
    - name: Set the Interface Parameters for port eth1 to DHCP of a WTI device
      cpm_interface_config:
        cpm_url: "nonexist.wti.com"
        cpm_username: "super"
        cpm_password: "super"
        use_https: true
        validate_certs: false
        interface: "eth1"
        negotiation: 0
        ipv4dhcpenable: 1
        ipv4dhcphostname: ""
        ipv4dhcplease: -1
        ipv4dhcpobdns: 0
        ipv4dhcpupdns: 0
        ipv4dhcpdefgateway: 0



Return Values
-------------

  data (always, complex, )
    The output JSON returned from the commands sent

    totalports (success, int, 1)
      Total interface ports requested of the WTI device.

    interface (always, dict, {'name': 'eth1', 'type': '0', 'mac_address': '00-09-9b-02-45-db', 'is_up': '0', 'is_gig': '1', 'speed': '10', 'negotiation': '0', 'ietf-ipv4': {'address': [{'ip': '10.10.10.2', 'netmask': '255.255.255.0', 'gateway': ''}], 'dhcpclient': [{'enable': 0, 'hostname': '', 'lease': -1, 'obdns': 1, 'updns': 1}]}, 'ietf-ipv6': {'address': [{'ip': '', 'netmask': '', 'gateway': ''}]}})
      Current k/v pairs of interface info for the WTI device after module execution.





Status
------




- This  is not guaranteed to have a backwards compatible interface. *[preview]*


- This  is maintained by community.



Authors
~~~~~~~

- Western Telematic Inc. (@wtinetworkgear)

