
cpm_current_info -- Get the Current Information of a WTI device
===============================================================

.. contents::
   :local:
   :depth: 1


Synopsis
--------

Get the Current Information of a WTI device






Parameters
----------

  cpm_url (True, str, None)
    This is the URL of the WTI device to send the module.


  cpm_username (True, str, None)
    This is the Username of the WTI device to send the module.


  cpm_password (True, str, None)
    This is the Password of the WTI device to send the module.


  cpm_startdate (False, str, None)
    Start date of the range to look for current data


  cpm_enddate (False, str, None)
    End date of the range to look for current data


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

    
    - name: Get the Current Information of a WTI device
      cpm_current_info:
        cpm_url: "nonexist.wti.com"
        cpm_username: "super"
        cpm_password: "super"
        use_https: true
        validate_certs: false

    - name: Get the Current Information of a WTI device
      cpm_current_info:
        cpm_url: "nonexist.wti.com"
        cpm_username: "super"
        cpm_password: "super"
        use_https: false
        validate_certs: false
        startdate: 01-12-2020"
        enddate: 02-16-2020"



Return Values
-------------

  data (always, complex, )
    The output JSON returned from the commands sent

    timestamp (success, str, 2020-02-24T20:54:03+00:00)
      Current timestamp of the WTI device after module execution.

    powerunit (success, str, 1)
      Identifies if the WTI device is a power type device.

    outletmetering (success, str, 1)
      Identifies if the WTI device has Poiwer Outlet metering.

    ats (success, str, 1)
      Identifies if the WTI device is an ATS type of power device.

    plugcount (success, str, 8)
      Current outlet plug count of the WTI device after module execution.

    powerfactor (success, str, 100)
      Power factor of the WTI device after module execution.

    powereff (success, str, 100)
      Power efficiency of the WTI device after module execution.

    powerdatacount (success, str, 1)
      Total powerdata samples returned after module execution.

    powerdata (success, dict, [{'timestamp': '2020-02-24T23:29:31+00:00', 'temperature': '90', 'format': 'F', 'branch1': [{'voltage1': '118.00', 'current1': '0.00'}]}])
      Power data of the WTI device after module execution.

    status (always, dict, {'code': '0', 'text': 'OK'})
      Return status after module completion





Status
------




- This  is not guaranteed to have a backwards compatible interface. *[preview]*


- This  is maintained by community.



Authors
~~~~~~~

- Western Telematic Inc. (@wtinetworkgear)

