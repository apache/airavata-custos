
cpm_plugconfig -- Get and Set Plug Parameters on WTI OOB and PDU power devices
==============================================================================

.. contents::
   :local:
   :depth: 1


Synopsis
--------

Get and Set Plug Parameters on WTI OOB and PDU devices






Parameters
----------

  cpm_action (True, str, None)
    This is the Action to send the module.


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
    F

    l

    a

    g

     

    t

    o

     

    c

    o

    n

    t

    r

    o

    l

     

    i

    f

     

    t

    h

    e

     

    l

    o

    o

    k

    u

    p

     

    w

    i

    l

    l

     

    o

    b

    s

    e

    r

    v

    e

     

    H

    T

    T

    P

     

    p

    r

    o

    x

    y

     

    e

    n

    v

    i

    r

    o

    n

    m

    e

    n

    t

     

    v

    a

    r

    i

    a

    b

    l

    e

    s

     

    w

    h

    e

    n

     

    p

    r

    e

    s

    e

    n

    t

    .


  plug_id (True, str, None)
    This is the plug number that is to be manipulated

    For the getplugconfig command, the plug_id 'all' will return the status of all the plugs the

    user has rights to access.


  plug_name (False, str, None)
    The new name of the Plug.


  plug_bootdelay (False, int, None)
    On a reboot command, this is the time when a plug will turn on power, after it has been turned off.

    0='0.5 Secs', 1='1 Sec', 2='2 Sec', 3='5 Sec', 4='15 Sec', 5='30 Sec', 6='1 Min', 7='2 Mins', - 8='3 Mins', 9='5 Mins'.


  plug_default (False, int, None)
    What the Plugs default state is when the device starts. 0 - Off, 1 - On.


  plug_bootpriority (False, int, None)
    Prioritizes which plug gets its state changed first. The lower the number the higher the priority.

    Valid value can from 1 to the maximum number of plugs of the WTI unit.









Examples
--------

.. code-block:: yaml+jinja

    
    # Get Plug parameters for all ports
    - name: Get the Plug parameters for ALL ports of a WTI Power device
      cpm_plugconfig:
        cpm_action: "getplugconfig"
        cpm_url: "rest.wti.com"
        cpm_username: "restpower"
        cpm_password: "restfulpowerpass12"
        use_https: true
        validate_certs: true
        plug_id: "all"

    # Get Plug parameters for port 2
    - name: Get the Plug parameters for the given port of a WTI Power device
      cpm_plugconfig:
        cpm_action: "getplugconfig"
        cpm_url: "rest.wti.com"
        cpm_username: "restpower"
        cpm_password: "restfulpowerpass12"
        use_https: true
        validate_certs: false
        plug_id: "2"

    # Configure plug 5
    - name: Configure parameters for Plug 5 on a given WTI Power device
      cpm_plugconfig:
        cpm_action: "setplugconfig"
        cpm_url: "rest.wti.com"
        cpm_username: "restpower"
        cpm_password: "restfulpowerpass12"
        use_https: true
        plug_id: "5"
        plug_name: "NewPlugNameFive"
        plug_bootdelay: "3"
        plug_default: "0"
        plug_bootpriority: "1"



Return Values
-------------

  data (always, str, )
    The output JSON returned from the commands sent




Status
------




- This  is not guaranteed to have a backwards compatible interface. *[preview]*


- This  is maintained by community.



Authors
~~~~~~~

- W
- e
- s
- t
- e
- r
- n
-  
- T
- e
- l
- e
- m
- a
- t
- i
- c
-  
- I
- n
- c
- .
-  
- (
- @
- w
- t
- i
- n
- e
- t
- w
- o
- r
- k
- g
- e
- a
- r
- )

