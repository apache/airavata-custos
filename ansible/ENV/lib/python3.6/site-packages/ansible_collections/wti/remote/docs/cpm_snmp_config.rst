
cpm_snmp_config -- Set network IPTables parameters in WTI OOB and PDU devices
=============================================================================

.. contents::
   :local:
   :depth: 1


Synopsis
--------

Set network IPTables parameters in WTI OOB and PDU devices






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


  protocol (False, int, None)
    The protocol that the SNMP entry should be applied. 0 = ipv4, 1 = ipv6.


  enable (False, int, None)
    The activates SNMP polling for the specified interface and protocol.


  interface (False, str, None)
    The ethernet port for the SNMP we are defining.


  readonly (False, int, None)
    Controls the ability to change configuration parameters with SNMP.


  version (False, int, None)
    Defined which version of SNMP the device will respond to 0 = V1/V2 Only, 1 = V3 Only, 2 = V1/V2/V3.


  contact (False, str, None)
    The name of the administrator responsible for SNMP issues.


  location (False, str, None)
    The location of the SNMP Server.


  systemname (False, str, None)
    The hostname of the WTI Device.


  rocommunity (False, str, None)
    Read Only Community Password, not used for SNMP V3.


  rwcommunity (False, str, None)
    Read/Write Community Password, not used for SNMP V3.


  clear (False, int, None)
    Removes all the users for the protocol being defined before setting the newly defined entries.


  index (False, list, None)
    Index of the user being modified (V3 only).


  username (False, list, None)
    Sets the User Name for SNMPv3 access (V3 only).


  authpriv (False, list, None)
    Configures the Authentication and Privacy features for SNMPv3 communication, 0 = Auth/NoPriv, 1 = Auth/Priv (V3 only).


  authpass (False, list, None)
    Sets the Authentication Password for SNMPv3 (V3 only).


  authproto (False, list, None)
    Which authentication protocol will be used, 0 = MD5, 1 = SHA1 (V3 only).


  privpass (False, list, None)
    Sets the Privacy Password for SNMPv3 (V3 only) (V3 only).


  privproto (False, list, None)
    Which privacy protocol will be used, 0 = DES, 1 = AES128 (V3 only).





Notes
-----

.. note::
   - Use ``groups/cpm`` in ``module_defaults`` to set common options used between CPM modules.




Examples
--------

.. code-block:: yaml+jinja

    
    # Sets the device SNMP Parameters
    - name: Set the an SNMP Parameter for a WTI device
      cpm_iptables_config:
        cpm_url: "nonexist.wti.com"
        cpm_username: "super"
        cpm_password: "super"
        interface: "eth0"
        use_https: true
        validate_certs: false
        protocol: 0
        clear: 1
        enable: 1
        readonly: 0
        version: 0
        rocommunity: "ropassword"
        rwcommunity: "rwpassword"

    # Sets the device SNMP Parameters
    - name: Set the SNMP Parameters a WTI device
      cpm_iptables_config:
        cpm_url: "nonexist.wti.com"
        cpm_username: "super"
        cpm_password: "super"
        use_https: true
        validate_certs: false
        version: 1
        index:
          - 1
          - 2
        username:
          - "username1"
          - "username2"
        authpriv:
          - 1
          - 1
        authpass:
          - "authpass1"
          - "uthpass2"
        authproto:
          - 1
          - 1
        privpass:
          - "authpass1"
          - "uthpass2"
        privproto:
          - 1
          - 1



Return Values
-------------

  data (always, complex, )
    The output JSON returned from the commands sent

    snmpaccess (always, dict, [{'eth0': {'ietf-ipv4': {'clear': 1, 'enable': 0, 'readonly': 0, 'version': 0, 'users': [{'username': 'username1', 'authpass': 'testpass', 'authpriv': '1', 'authproto': '0', 'privpass': 'privpass1', 'privproto': '0', 'index': '1'}]}}}])
      Current k/v pairs of interface info for the WTI device after module execution.





Status
------




- This  is not guaranteed to have a backwards compatible interface. *[preview]*


- This  is maintained by community.



Authors
~~~~~~~

- Western Telematic Inc. (@wtinetworkgear)

