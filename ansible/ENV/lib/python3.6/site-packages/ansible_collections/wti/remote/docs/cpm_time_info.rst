
cpm_time_info -- Get Time/Date parameters in WTI OOB and PDU devices
====================================================================

.. contents::
   :local:
   :depth: 1


Synopsis
--------

Get Time/Date and NTP parameters from WTI OOB and PDU devices






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

    
    - name: Get the Time/Date Parameters for a WTI device
      cpm_time_info:
        cpm_url: "nonexist.wti.com"
        cpm_username: "super"
        cpm_password: "super"
        use_https: true
        validate_certs: false

    - name: Get the Time/Date Parameters for a WTI device
      cpm_time_info:
        cpm_url: "nonexist.wti.com"
        cpm_username: "super"
        cpm_password: "super"
        use_https: false
        validate_certs: false



Return Values
-------------

  data (always, complex, )
    The output JSON returned from the commands sent

    date (success, str, 11/14/2019)
      Current Date of the WTI device after module execution.

    time (success, str, 12:12:00)
      Current Time of the WTI device after module execution.

    timezone (success, int, 5)
      Current Timezone of the WTI device after module execution.

    ntp (always, dict, {'enable': '0', 'ietf-ipv4': {'address': [{'primary': '192.168.0.169', 'secondary': '12.34.56.78'}]}, 'ietf-ipv6': {'address': [{'primary': '', 'secondary': ''}]}, 'timeout': '4'})
      Current k/v pairs of ntp info of the WTI device after module execution.





Status
------




- This  is not guaranteed to have a backwards compatible interface. *[preview]*


- This  is maintained by community.



Authors
~~~~~~~

- Western Telematic Inc. (@wtinetworkgear)

