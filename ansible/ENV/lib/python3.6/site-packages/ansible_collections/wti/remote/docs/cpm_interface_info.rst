
cpm_interface_info -- Get network interface parameters from WTI OOB and PDU devices
===================================================================================

.. contents::
   :local:
   :depth: 1


Synopsis
--------

Get network interface parameters from WTI OOB and PDU devices






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


  interface (False, list, None)
    This is the ethernet port name that is getting retrieved. It can include a single ethernet

    port name, multiple ethernet port names separated by commas or not defined for all ports.





Notes
-----

.. note::
   - Use ``groups/cpm`` in ``module_defaults`` to set common options used between CPM modules.)




Examples
--------

.. code-block:: yaml+jinja

    
    - name: Get the network interface Parameters for a WTI device for all interfaces
      cpm_interface_info:
        cpm_url: "nonexist.wti.com"
        cpm_username: "super"
        cpm_password: "super"
        use_https: true
        validate_certs: false

    - name: Get the network interface Parameters for a WTI device for a specific interface
      cpm_interface_info:
        cpm_url: "nonexist.wti.com"
        cpm_username: "super"
        cpm_password: "super"
        interface: "eth0,eth1"
        use_https: false
        validate_certs: false



Return Values
-------------

  data (always, complex, )
    The output JSON returned from the commands sent

    totalports (success, int, 1)
      Total ethernet ports requested of the WTI device.

    interface (always, dict, {'name': 'eth1', 'type': '0', 'mac_address': '00-09-9b-02-45-db', 'is_up': '0', 'is_gig': '1', 'speed': '10', 'negotiation': '0', 'ietf-ipv4': {'address': [{'ip': '10.10.10.2', 'netmask': '255.255.255.0', 'gateway': ''}], 'dhcpclient': [{'enable': 0, 'hostname': '', 'lease': -1, 'obdns': 1, 'updns': 1}]}, 'ietf-ipv6': {'address': [{'ip': '', 'netmask': '', 'gateway': ''}]}})
      Current k/v pairs of interface info for the WTI device after module execution.





Status
------




- This  is not guaranteed to have a backwards compatible interface. *[preview]*


- This  is maintained by community.



Authors
~~~~~~~

- Western Telematic Inc. (@wtinetworkgear)

