
cpm_status_info -- Get general status information from WTI OOB and PDU devices
==============================================================================

.. contents::
   :local:
   :depth: 1


Synopsis
--------

Get temperature general status from WTI OOB and PDU devices






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

    
    - name: Get the Status Information for a WTI device
      cpm_status_info:
        cpm_url: "nonexist.wti.com"
        cpm_username: "super"
        cpm_password: "super"
        use_https: true
        validate_certs: false

    - name: Get the Status Information for a WTI device
      cpm_status_info:
        cpm_url: "nonexist.wti.com"
        cpm_username: "super"
        cpm_password: "super"
        use_https: false
        validate_certs: false



Return Values
-------------

  data (always, complex, )
    The output JSON returned from the commands sent

    vendor (success, str, wti)
      Identifies WTI device as a WTI device.

    product (success, str, CPM-800-1-CA)
      Current Product Part Number of the WTI device.

    totalports (success, str, 9)
      Total serial ports of the WTI device.

    totalplugs (success, str, 8)
      Total Power Outlet plugs of the WTI device.

    option1/2 (success, str, WPO-STRT-CPM8 / W4G-VZW-CPM8)
      Various Identity options of the WTI.

    softwareversion (success, str, 6.60 19 Feb 2020)
      Expanded Firmware version of the WTI device.

    serialnumber (success, str, 12345678901234)
      Current Serial number of the WTI device.

    assettag (success, str, ARTE121)
      Current Asset Tag of the WTI device.

    siteid (success, str, GENEVARACK)
      Current Site-ID of the WTI device.

    analogmodemphonenumber (success, str, 9495869959)
      Current Analog Modem (if installed) Phone number of the WTI device.

    modeminstalled (success, str, Yes, 4G/LTE)
      Identifies if a modem is installed in the WTI device.

    modemmodel (success, str, MTSMC-LVW2)
      Identifies the modem model number (if installed) in the WTI device.

    gig_dualphy (success, str, Yes, Yes)
      Identifies dual ethernet port and gigabyte ethernet ports in the WTI device.

    cpu_boardprogramdate (success, str, ARM, 4-30-2019)
      Current Board and Program date of the WTI device.

    ram_flash (success, str, 512 MB, 128 MB)
      Total RAM and FLASH installed in the WTI device..

    lineinputcount_rating (success, str, 1 ,  20 Amps)
      Identifies total power inlets and their power rating.

    currentmonitor (success, str, Yes)
      Identifies if the unit has current monitoring capabilites.

    keylength (success, str, 2048)
      Current key length of the WTI device.

    opensslversion (success, str, 1.1.1d  10 Sep 2019)
      Current OpenSSL version running on the WTI device.

    opensshversion (success, str, 8.2p1)
      Current OpenSSH running on the WTI device.

    apacheversion (success, str, 2.4.41)
      Current Apache Web version running on the WTI device.

    apirelease (success, str, March 2020)
      Current Date of the API release of the WTI device.

    uptime (success, str, 259308.26)
      Current uptime of the WTI device.

    energywise (success, str, 1.2.0)
      Current Energywise version of the WTI device.

    restful (success, str, v1.0, v2 (Mar20))
      Current RESTful version of the WTI device.

    interface_list (success, str, eth0)
      Current ethernet ports of the WTI device.

    macaddresses (always, dict, {'mac': '00-09-9b-02-9a-26'})
      Current mac addresses of the WTI device.

    status (always, dict, {'code': '0', 'text': 'OK'})
      Return status after module completion





Status
------




- This  is not guaranteed to have a backwards compatible interface. *[preview]*


- This  is maintained by community.



Authors
~~~~~~~

- Western Telematic Inc. (@wtinetworkgear)

