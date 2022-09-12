
cpm_user -- Get various status and parameters from WTI OOB and PDU devices
==========================================================================

.. contents::
   :local:
   :depth: 1


Synopsis
--------

Get/Add/Edit Delete Users from WTI OOB and PDU devices






Parameters
----------

  cpm_action (True, str, None)
    This is the Action to send the module.


  cpm_url (True, str, None)
    This is the URL of the WTI device to send the module.


  cpm_username (True, str, None)
    This is the Basic Authentication Username of the WTI device to send the module.


  cpm_password (True, str, None)
    This is the Basic Authentication Password of the WTI device to send the module.


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


  user_name (True, str, None)
    This is the User Name that needs to be create/modified/deleted


  user_pass (False, str, None)
    This is the User Password that needs to be create/modified/deleted

    If the user is being Created this parameter is required


  user_accesslevel (False, int, None)
    This is the access level that needs to be create/modified/deleted

    0 View, 1 User, 2 SuperUser, 3 Administrator


  user_accessssh (False, int, None)
    If the user has access to the WTI device via SSH

    0 No , 1 Yes


  user_accessserial (False, int, None)
    If the user has access to the WTI device via Serial ports

    0 No , 1 Yes


  user_accessweb (False, int, None)
    If the user has access to the WTI device via Web

    0 No , 1 Yes


  user_accessapi (False, int, None)
    If the user has access to the WTI device via RESTful APIs

    0 No , 1 Yes


  user_accessmonitor (False, int, None)
    If the user has ability to monitor connection sessions

    0 No , 1 Yes


  user_accessoutbound (False, int, None)
    If the user has ability to initiate Outbound connection

    0 No , 1 Yes


  user_portaccess (False, str, None)
    If AccessLevel is lower than Administrator, which ports the user has access


  user_plugaccess (False, str, None)
    If AccessLevel is lower than Administrator, which plugs the user has access


  user_groupaccess (False, str, None)
    If AccessLevel is lower than Administrator, which Groups the user has access


  user_callbackphone (False, str, None)
    This is the Call Back phone number used for POTS modem connections









Examples
--------

.. code-block:: yaml+jinja

    
    # Get User Parameters
    - name: Get the User Parameters for the given user of a WTI device
      cpm_user:
        cpm_action: "getuser"
        cpm_url: "rest.wti.com"
        cpm_username: "restuser"
        cpm_password: "restfuluserpass12"
        use_https: true
        validate_certs: true
        user_name: "usernumberone"

    # Create User
    - name: Create a User on a given WTI device
      cpm_user:
        cpm_action: "adduser"
        cpm_url: "rest.wti.com"
        cpm_username: "restuser"
        cpm_password: "restfuluserpass12"
        use_https: true
        validate_certs: false
        user_name: "usernumberone"
        user_pass: "complicatedpassword"
        user_accesslevel: 2
        user_accessssh: 1
        user_accessserial: 1
        user_accessweb: 0
        user_accessapi: 1
        user_accessmonitor: 0
        user_accessoutbound: 0
        user_portaccess: "10011111"
        user_plugaccess: "00000111"
        user_groupaccess: "00000000"

    # Edit User
    - name: Edit a User on a given WTI device
      cpm_user:
        cpm_action: "edituser"
        cpm_url: "rest.wti.com"
        cpm_username: "restuser"
        cpm_password: "restfuluserpass12"
        use_https: true
        validate_certs: false
        user_name: "usernumberone"
        user_pass: "newpasswordcomplicatedpassword"

    # Delete User
    - name: Delete a User from a given WTI device
      cpm_user:
        cpm_action: "deleteuser"
        cpm_url: "rest.wti.com"
        cpm_username: "restuser"
        cpm_password: "restfuluserpass12"
        use_https: true
        validate_certs: true
        user_name: "usernumberone"



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

