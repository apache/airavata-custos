
cpm_plugcontrol -- Get and Set Plug actions on WTI OOB and PDU power devices
============================================================================

.. contents::
   :local:
   :depth: 1


Synopsis
--------

Get and Set Plug actions on WTI OOB and PDU devices






Parameters
----------

  cpm_action (True, str, None)
    This is the Action to send the module.


  cpm_url (True, str, None)
    This is the URL of the WTI device  to send the module.


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
    This is the plug number or the plug name that is to be manipulated

    For the plugget command, the plug_id 'all' will return the status of all the plugs the

    user has rights to access.


  plug_state (False, str, None)
    This is what action to take on the plug.









Examples
--------

.. code-block:: yaml+jinja

    
    # Get Plug status for all ports
    - name: Get the Plug status for ALL ports of a WTI device
      cpm_plugcontrol:
        cpm_action: "getplugcontrol"
        cpm_url: "rest.wti.com"
        cpm_username: "restpower"
        cpm_password: "restfulpowerpass12"
        use_https: true
        validate_certs: true
        plug_id: "all"

    # Get Plug status for port 2
    - name: Get the Plug status for the given port of a WTI device
      cpm_plugcontrol:
        cpm_action: "getplugcontrol"
        cpm_url: "rest.wti.com"
        cpm_username: "restpower"
        cpm_password: "restfulpowerpass12"
        use_https: true
        validate_certs: false
        plug_id: "2"

    # Reboot plug 5
    - name: Reboot Plug 5 on a given WTI device
      cpm_plugcontrol:
        cpm_action: "setplugcontrol"
        cpm_url: "rest.wti.com"
        cpm_username: "restpower"
        cpm_password: "restfulpowerpass12"
        use_https: true
        plug_id: "5"
        plug_state: "boot"



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

