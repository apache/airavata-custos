
cpm_snmp_info -- Get network SNMP parameters from WTI OOB and PDU devices
=========================================================================

.. contents::
   :local:
   :depth: 1


Synopsis
--------

Get network SNMP parameters from WTI OOB and PDU devices






Parameters
----------

  cpm_url (True, str, None)
    This is the URL of the WTI device to send the module.


  cpm_username (True, str, None)
    This is the Username of the WTI device to send the module.


  cpm_password (True, str, None)
    This is the Password of the WTI device to send the module.


  interface (False, str, None)
    This is the ethernet port name that is getting retrieved. It can include a single ethernet

    port name, multiple ethernet port names separated by commas or not defined for all ports.


  use_https (False, bool, True)
    Designates to use an https connection or http connection.


  validate_certs (False, bool, True)
    If false, SSL certificates will not be validated. This should only be used

    on personally controlled sites using self-signed certificates.


  use_proxy (False, bool, False)
    Flag to control if the lookup will observe HTTP proxy environment variables when present.





Notes
-----

.. note::
   - Use ``groups/cpm`` in ``module_defaults`` to set common options used between CPM modules.)




Examples
--------

.. code-block:: yaml+jinja

    
    - name: Get the network SNMP Parameters for all interfaces of a WTI device.
      cpm_interface_info:
        cpm_url: "nonexist.wti.com"
        cpm_username: "super"
        cpm_password: "super"
        use_https: true
        validate_certs: false


    - name: Get the network SNMP Parameters for eth0 of a WTI device.
      cpm_interface_info:
        cpm_url: "nonexist.wti.com"
        cpm_username: "super"
        cpm_password: "super"
        use_https: false
        validate_certs: false
        interface: "eth0"



Return Values
-------------

  data (always, complex, )
    The output JSON returned from the commands sent

    snmpaccess (always, dict, {'snmpaccess': [{'eth0': {'ietf-ipv4': [{'enable': 0, 'users': [{'index': '1', 'username': 'test10', 'authpriv': '1', 'authpass': 'testpass', 'authproto': '0', 'privpass': 'testpass', 'privproto': '1'}]}], 'ietf-ipv6': [{'enable': 0, 'users': [{'index': '1', 'username': 'test10', 'authpriv': '1', 'authpass': 'testpass', 'authproto': '0', 'privpass': 'testpass', 'privproto': '1'}]}]}}]})
      Current k/v pairs of SNMP info for the WTI device after module execution.





Status
------




- This  is not guaranteed to have a backwards compatible interface. *[preview]*


- This  is maintained by community.



Authors
~~~~~~~

- Western Telematic Inc. (@wtinetworkgear)

