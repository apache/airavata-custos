AAA role
=======

This role facilitates the configuration of authentication, authorization, and acccounting (AAA), and supports the configuration of RADIUS and TACACS servers. This role is abstracted for Dell EMC PowerSwitch platforms running Dell EMC OS6. 

The AAA role requires an SSH connection for connectivity to Dell EMC OS6. You can use any of the built-in OS connection variables.

Role variables
--------------

- Role is abstracted using the `ansible_network_os` variable that can take `dellemc.os6.os6` as the value
- If `os6_cfg_generate` is set to true, the variable generates the role configuration commands in a file
- Any role variable with a corresponding state variable set to absent negates the configuration of that variable
- Setting an empty value for any variable negates the corresponding configuration
- Variables and values are case-sensitive

**os6_aaa keys**

| Key        | Type                      | Description                                             | Support               |
|------------|---------------------------|---------------------------------------------------------|-----------------------|
| ``radius_server``            | dictionary        | Configures the RADIUS server (see ``radius_server.*``) | os6 |
| ``radius_server.key``        | string (required): 0,7,LINE | Configures the authentication key for the RADIUS server | os6 |
| ``radius_server.key_string`` | string            | Configures the user key string; variable takes the hidden user key string if value is 7; variable takes the unencrypted user key (clear-text) if value is 0; variable supported only if *radius_server.key* is 7 or 0 | os6 |
| ``radius_server.retransmit`` | integer           | Configures the number of retransmissions; field to be left blank to remove the retransimission configuration for RADIUS server authentication | os6  |
| ``radius_server.timeout``    | integer           | Configures the timeout for retransmissions, timeout must be an integer 1 and 30; field needs to be left blank to remove the timeout configurations for RADIUS server authentication | os6  |
| ``radius_server.host``       | dictionary        | Configures the RADIUS server host (see ``host.*``) | os6  |
| ``host.ip``                  | string            | Configures the RADIUS server host address | os6  |
| ``host.key``                 | string (required); 0,7,LINE           | Configures the authentication key | os6  |
| ``host.key_string``          | string            | Configures the user key string; variable takes the hidden user key string if value is 7; variable takes the unencrypted user key (clear-text) if value is 0; variable supported only if *host.key* is 7 or 0 | os6 |
| ``host.retransmit``          | integer           | Configures the number of retransmissions | os6 |
| ``host.auth_port``           | integer           | Configures the authentication port (0 to 65535)  | os6  |
| ``host.timeout``             | integer           | Configures timeout for retransmissions | os6 |
| ``host.state``               | string: present,absent         | Removes the RADIUS server host if set to absent | os6  |
| ``radius_server.acct``       | dictionary        | Configures the RADIUS server acct (see ``host.*``) | os6  |
| ``acct.ip``                  | string            | Configures the RADIUS server acct address | os6  |
| ``acct.key``                 | string (required); 0,7,LINE           | Configures the authentication key | os6  |
| ``acct.key_string``          | string            | Configures the user key string; variable takes the hidden user key string if value is 7; variable takes the unencrypted user key (clear-text) if value is 0; variable supported only if *host.key* is 7 or 0 | os6 |
| ``acct.auth_port`` | integer           | Configures the authentication port (0 to 65535)  | os6  |
| ``acct.state``               | string: present,absent         | Removes the RADIUS server acct if set to absent | os6  |
| ``radius_server.auth``       | dictionary        | Configures the RADIUS server auth (see ``auth.*``) | os6 |
| ``auth.ip``                  | string            | Configures the RADIUS server host address | os6  |
| ``auth.key``                 | string (required); 0,7,LINE           | Configures the authentication key | os6  |
| ``auth.key_string``          | string            | Configures the user key string; variable takes the hidden user key string if value is 7; variable takes the unencrypted user key (clear-text) if value is 0; variable supported only if *host.key* is 7 or 0 | os6 |
| ``auth.name``                | string (required) | Configures the auth name of the RADIUS servers | os6  |
| ``auth.usage``               | string (required) | Configures the usage type of the RADIUS servers | os6  |
| ``auth.priority``            | integer           | Configures the number of priority | os6 |
| ``auth.retransmit``          | integer           | Configures the number of retransmissions | os6 |
| ``auth.auth_port``           | integer           | Configures the authentication port (0 to 65535)  | os6  |
| ``auth.timeout``             | integer           | Configures timeout for retransmissions | os6 |
| ``auth.deadtime``            | integer           | Configures the number of deadtime | os6 |
| ``auth.attribute``           | dictionary        | Configures the RADIUS server auth (see ``attribute.*``) | os6 |
| ``attribute.id``             | integer           | Configures the RADIUS server attribute ID (see ``attribute.*``) | os6  |
| ``attribute.type``           | integer            | Configures the RADIUS server attribute type based on ID | os6  |
| ``attribute.state``          | string: present,absent         | Removes the RADIUS server attribute if set to absent | os6  |
| ``auth.state``               | string: present,absent         | Removes the radius server auth if set to absent | os6  |
| ``radius_server.attribute``  | dictionary        | Configures the RADIUS server auth (see ``attribute.*``) | os6 |
| ``attribute.id``             | integer            | Configures the RADIUS server attribute ID (see ``attribute.*``) | os6  |
| ``attribute.type``           | integer            | Configures the RADIUS server attribute type based on ID | os6  |
| ``attribute.state``          | string: present,absent         | Removes the RADIUS server attribute if set to absent | os6  |
| ``tacacs_server``            | dictionary        | Configures the TACACS server (see ``tacacs_server.*``)| os6 |
| ``tacacs_server.key``        | string (required): 0,7,LINE           | Configures the authentication key for TACACS server | os6 |
| ``tacacs_server.key_string`` | string            | Configures the user key string; variable takes the hidden user key string if value is 7; variable takes the unencrypted user key (clear-text) if value is 0; variable supported only if *tacacs_server.key* is 7 or 0 | os6 |
| ``tacacs_server.host``       | dictionary        | Configures the TACACS server host (see ``host.*``) | os6 |
| ``host.ip``                  | string            | Configures the TACACS sever host address | os6 |
| ``host.key``                 | string (required): 0,7,LINE           | Configures the authentication key | os6 |
| ``host.key_string``          | string            | Configures the user key string; variable takes the hidden user key string if value is 7; variable takes the unencrypted user key (clear-text) if value is 0; variable supported only if *host.key* is 7 or 0 | os6  |
| ``host.auth_port``           | integer           | Configures the authentication port (0 to 65535) | os6 |
| ``host.timeout``             | integer           | Configures the timeout for retransmissions | os6 |
| ``host.state``               | string: present,absent         | Removes the TACACS server host if set to absent | os6 |
| ``aaa_accounting``           | dictionary        | Configures accounting parameters (see ``aaa_accounting.*``) | os6 |
| ``aaa_accounting.dot1x``     | string: none,start-stop,stop-only,wait-start        | Configures accounting for dot1x events | os6 |
| ``aaa_authorization``        | dictionary        | Configures authorization parameters (see ``aaa_authorization.*``) | os6  |
| ``aaa_authorization.exec``   | list              | Configures authorization for EXEC (shell) commands (see ``exec.*``) | os6 |
| ``exec.authorization_list_name`` | string        | Configures named authorization list for EXEC commands | os6  |
| ``exec.authorization_method`` | string: none         | Configures no authorization of EXEC commands | os6  |
| ``exec.use_data``            | string: local,tacacs, radius        | Configures data used for authorization | os6 |
| ``exec.state``               | string: present,absent         | Removes the named authorization list for the EXEC commands if set to absent | os6  |
| ``aaa_authorization.network``     | string: none,radius,ias        | Configures authorization for network events | os6 |
| ``aaa_authentication.auth_list`` | list        | Configures named authentication list for hosts (see ``host.*``) | os6 |
| ``auth_list.name``           | string         | Configures named authentication list | os6 |
| ``auth_list.login_or_enable`` | string: enable,login         | Configures authentication list for login or enable | os6  |
| ``auth_list.server``         | string: radius,tacacs         | Configures AAA to use this list of all server hosts | os6 |
| ``auth_list.use_password``   | string: line,local,enable,none         | Configures password to use for authentication | os6 |
| ``auth_list.state``          | string: present,absent         | Removes the named authentication list if set to absent | os6 |
| ``aaa_authentication.dot1x``     | string: none,radius,ias        | Configures authentication for dot1x events | os6 |
| ``aaa_server``               | dictionary        | Configures the AAA server (see ``aaa_server.*``) | os6 |
| ``radius``                   | dictionary        | Configures the RADIUS server (see ``radius.*``) | os6 |
| ``dynamic_author``           | dictionary        | Configures the RADIUS server (see ``dynamic_author.*``) | os6 |
| ``dynamic_author.auth_type`` | string         | Configures the authentication type for the radius server | os6 |
| ``dynamic_author.client``    | list     | Configures the client for the RADIUS server | os6 |
| ``client.ip``                | string         | Configures the client IP for the radius server | os6 |
| ``client.key``               | string (required): 0,7,LINE | Configures the authentication key for the RADIUS server | os6 |
| ``client.key_string``        | string            | Configures the user key string; variable takes the hidden user key string if value is 7; variable takes the unencrypted user key (clear-text) if value is 0; variable supported only if *radius_server.key* is 7 or 0 | os6 |
| ``client.state``             | string: present,absent         | Removes the accounting of client if set to absent | os6 |
| ``dynamic_author.state``     | string: present,absent         | Removes the accounting of client if set to absent | os6 |
> **NOTE**: Asterisk (*) denotes the default value if none is specified.

Connection variables
--------------------

Ansible Dell EMC Networking roles require connection information to establish communication with the nodes in your inventory. This information can exist in the Ansible *group_vars* or *host_vars* directories or inventory, or in the playbook itself.

| Key         | Required | Choices    | Description                                         |
|-------------|----------|------------|-----------------------------------------------------|
| ``ansible_host`` | yes      |            | Specifies the hostname or address for connecting to the remote device over the specified transport |
| ``ansible_port`` | no       |            | Specifies the port used to build the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_PORT` option is used; it defaults to 22 |
| ``ansible_ssh_user`` | no       |            | Specifies the username that authenticates the CLI login for the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_USER` environment variable value is used  |
| ``ansible_ssh_pass`` | no       |            | Specifies the password that authenticates the connection to the remote device.  |
| ``ansible_become`` | no       | yes, no\*   | Instructs the module to enter privileged mode on the remote device before sending any commands; if value is unspecified, the `ANSIBLE_BECOME` environment variable value is used, and the device attempts to execute all commands in non-privileged mode |
| ``ansible_become_method`` | no       | enable, sudo\*   | Instructs the module to allow the become method to be specified for handling privilege escalation; if value is unspecified, the `ANSIBLE_BECOME_METHOD` environment variable value is used. |
| ``ansible_become_pass`` | no       |            | Specifies the password to use if required to enter privileged mode on the remote device; if ``ansible_become`` is set to no this key is not applicable. |
| ``ansible_network_os`` | yes      | os6, null\*  | This value is used to load the correct terminal and cliconf plugins to communicate with the remote device. |

> **NOTE**: Asterisk (*) denotes the default value if none is specified.

Example playbook
----------------

This example uses the *os6_aaa* role to configure AAA for RADIUS and TACACS servers. It creates a *hosts* file with the switch details and corresponding variables. The hosts file should define the `ansible_network_os` variable with the corresponding Dell EMC OS6 name. 

When `os6_cfg_generate` is set to true, the variable generates the configuration commands as a .part file in the *build_dir* path. By default, it is set to false and it writes a simple playbook that only references the *os6_aaa* role.

**Sample hosts file**

    switch1 ansible_host= <ip_address> 

**Sample host_vars/switch1**

    hostname: switch1
    ansible_become: yes
    ansible_become_method: enable
    ansible_become_pass: xxxxx
    ansible_ssh_user: xxxxx
    ansible_ssh_pass: xxxxx
    ansible_network_os: dellemc.os6.os6
    build_dir: ../temp/temp_os6

    os6_aaa:
      radius_server:
            key: 7
            key_string: 9ea8ec421c2e2e5bec757f44205015f6d81e83a4f0aa52fb
            retransmit: 5
            timeout: 25
            host:
              - ip: 10.0.0.1
                key: 0
                key_string: aaa
                retransmit: 5
                auth_port: 3
                timeout: 2
                state: present
      tacacs_server:
            key: 7
            key_string: 9ea8ec421c2e2e5bec757f44205015f6d81e83a4f0aa52fa
            host:
              - ip: 10.0.0.50
                key: 0
                key_string: aaa
                auth_port: 3
                timeout: 2
                state: present
      aaa_accounting:
          dot1x: none
      aaa_authorization:
          exec:
            - authorization_list_name: aaa
              authorization_method: none
              use_data: local
              state: present
          network: radius
      aaa_authentication:
          auth_list:
            - name: default
              login_or_enable: login
              server: radius
              use_password: local
              state: present
            - name: console
              server: tacacs
              login_or_enable: login
              use_password: local
              state: present
      aaa_server:
          radius: 
            dynamic_author:
              auth_type: 
              client: 
                - ip: 10.0.0.1
                  key: 0
                  key_string: aaskjsksdkjsdda
                  state: present
                - ip: 10.0.0.2
                  key: 
                  key_string: aaskjsksdkjsdda
                  state: present
              state: present
            


**Simple playbook to setup system — switch1.yaml**

    - hosts: switch1
      roles:
         - dellemc.os6.os6_aaa

**Run**

    ansible-playbook -i hosts switch1.yaml

(c) 2017-2020 Dell Inc. or its subsidiaries. All rights reserved.
