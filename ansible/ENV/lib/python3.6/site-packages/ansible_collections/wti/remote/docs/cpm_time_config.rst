
cpm_time_config -- Set Time/Date parameters in WTI OOB and PDU devices.
=======================================================================

.. contents::
   :local:
   :depth: 1


Synopsis
--------

Set Time/Date and NTP parameters parameters in WTI OOB and PDU devices






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


  date (False, str, None)
    Static date in the format of two digit month, two digit day, four digit year separated by a slash symbol.


  time (False, str, None)
    Static time in the format of two digit hour, two digit minute, two digit second separated by a colon symbol.


  timezone (False, int, None)
    This is timezone that is assigned to the WTI device.


  ntpenable (False, int, None)
    This enables or disables the NTP client service.


  ipv4address (False, str, None)
    Comma separated string of up to two addresses for a primary and secondary IPv4 base NTP server.


  ipv6address (False, str, None)
    Comma separated string of up to two addresses for a primary and secondary IPv6 base NTP server.


  timeout (False, int, None)
    Set the network timeout in seconds of contacting the NTP servers, valid options can be from 1-60.





Notes
-----

.. note::
   - Use ``groups/cpm`` in ``module_defaults`` to set common options used between CPM modules.




Examples
--------

.. code-block:: yaml+jinja

    
    # Set a static time/date and timezone of a WTI device
    - name: Set known fixed time/date of a WTI device
      cpm_time_config:
        cpm_url: "nonexist.wti.com"
        cpm_username: "super"
        cpm_password: "super"
        use_https: true
        validate_certs: false
        date: "12/12/2019"
        time: "09:23:46"
        timezone: 5

    # Enable NTP and set primary and seconday servers of a WTI device
    - name: Set NTP primary and seconday servers of a WTI device
      cpm_time_config:
        cpm_url: "nonexist.wti.com"
        cpm_username: "super"
        cpm_password: "super"
        use_https: true
        validate_certs: false
        timezone: 5
        ntpenable: 1
        ipv4address: "129.6.15.28.pool.ntp.org"
        timeout: 15



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

