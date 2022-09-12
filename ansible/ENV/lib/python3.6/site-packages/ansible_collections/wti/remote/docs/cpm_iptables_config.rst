
cpm_iptables_config -- Set network IPTables parameters in WTI OOB and PDU devices
=================================================================================

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
    The protocol that the iptables entry should be applied. 0 = ipv4, 1 = ipv6.


  index (False, list, None)
    Index in which command should be inserted. If not defined entry will start at position one.


  command (True, list, None)
    Actual iptables command to send to the WTI device.


  clear (False, int, None)
    Removes all the iptables for the protocol being defined before setting the newly defined entry.





Notes
-----

.. note::
   - Use ``groups/cpm`` in ``module_defaults`` to set common options used between CPM modules.




Examples
--------

.. code-block:: yaml+jinja

    
    # Set Network IPTables Parameters
    - name: Set the an IPTables Parameter for a WTI device
      cpm_iptables_config:
        cpm_url: "nonexist.wti.com"
        cpm_username: "super"
        cpm_password: "super"
        use_https: true
        validate_certs: false
        command: "iptables -A INPUT -p tcp -m state --state NEW -m tcp --dport 443 -j ACCEPT"

    # Sets multiple Network IPTables Parameters
    - name: Set the IPTables Parameters a WTI device
      cpm_iptables_config:
        cpm_url: "nonexist.wti.com"
        cpm_username: "super"
        cpm_password: "super"
        use_https: true
        validate_certs: false
        index:
          - 1
          - 2
        command:
          - "iptables -A INPUT -p tcp -m state --state NEW -m tcp --dport 443 -j ACCEPT"
          - "iptables -A INPUT -p tcp -m state --state NEW -m tcp --dport 22 -j ACCEPT"



Return Values
-------------

  data (always, complex, )
    The output JSON returned from the commands sent

    iptables (always, dict, [{'eth0': {'ietf-ipv4': {'clear': 1, 'entries': [{'entry': 'iptables -A INPUT -p tcp -m state --state NEW -m tcp --dport 443 -j ACCEPT', 'index': '1'}, {'entry': 'iptables -A INPUT -p tcp -m state --state NEW -m tcp --dport 22 -j ACCEPT', 'index': '2'}]}}}])
      Current k/v pairs of interface info for the WTI device after module execution.





Status
------




- This  is not guaranteed to have a backwards compatible interface. *[preview]*


- This  is maintained by community.



Authors
~~~~~~~

- Western Telematic Inc. (@wtinetworkgear)

