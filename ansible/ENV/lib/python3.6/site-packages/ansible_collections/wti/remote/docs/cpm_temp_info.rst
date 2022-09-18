
cpm_temp_info -- Get temperature information from WTI OOB and PDU devices
=========================================================================

.. contents::
   :local:
   :depth: 1


Synopsis
--------

Get temperature information from WTI OOB and PDU devices






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





Notes
-----

.. note::
   - Use ``groups/cpm`` in ``module_defaults`` to set common options used between CPM modules.)




Examples
--------

.. code-block:: yaml+jinja

    
    - name: Get the Temperature Information of a WTI device
      cpm_temp_info:
        cpm_url: "nonexist.wti.com"
        cpm_username: "super"
        cpm_password: "super"
        use_https: true
        validate_certs: false

    - name: Get the Temperature Information of a WTI device
      cpm_temp_info:
        cpm_url: "nonexist.wti.com"
        cpm_username: "super"
        cpm_password: "super"
        use_https: false
        validate_certs: false



Return Values
-------------

  data (always, complex, )
    The output JSON returned from the commands sent

    temperature (success, str, 76)
      Current Temperature of the WTI device after module execution.

    format (success, str, F)
      Current Temperature format (Celsius or Fahrenheit) of the WTI device after module execution.

    timestamp (success, str, 2020-02-24T20:54:03+00:00)
      Current timestamp of the WTI device after module execution.

    status (always, dict, {'code': '0', 'text': 'OK'})
      Return status after module completion





Status
------




- This  is not guaranteed to have a backwards compatible interface. *[preview]*


- This  is maintained by community.



Authors
~~~~~~~

- Western Telematic Inc. (@wtinetworkgear)

