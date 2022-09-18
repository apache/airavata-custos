AAA role
========

This role facilitates the configuration of authentication, authorization, and acccounting (AAA), and supports the configuration of RADIUS and TACACS servers. This role is abstracted for Dell EMC PowerSwitch platforms running Dell EMC OS9.

The AAA role requires an SSH connection for connectivity to a Dell EMC OS9 device. You can use any of the built-in OS connection variables.


Role variables
--------------

- Role is abstracted using the `ansible_network_os` variable that can take `dellemc.os9.os9` as the value
- If `os9_cfg_generate` is set to true, the variable generates the role configuration commands in a file
- Any role variable with a corresponding state variable set to absent negates the configuration of that variable
- Setting an empty value for any variable negates the corresponding configuration
- Variables and values are case-sensitive

**os9_aaa keys**

| Key        | Type                      | Description                                             | Support               |
|------------|---------------------------|---------------------------------------------------------|-----------------------|
| ``radius_server``            | dictionary        | Configures the RADIUS server (see ``radius_server.*``) | os9 |
| ``radius_server.key``        | string (required): 0,7,LINE | Configures the authentication key for the RADIUS server | os9 |
| ``radius_server.key_string`` | string            | Configures the user key string; variable takes the hidden user key string if value is 7; variable takes the unencrypted user key (clear-text) if value is 0; variable supported only if *radius_server.key* is 7 or 0 | os9 |
| ``radius_server.retransmit`` | integer           | Configures the number of retransmissions | os9  |
| ``radius_server.timeout``    | integer           | Configures the timeout for retransmissions | os9  |
| ``radius_server.deadtime``   | integer           | Configures the server dead time | os9  |
| ``radius_server.group``      | dictionary        | Configures the RADIUS servers group (see ``group.*``) | os9 |  
| ``group.name``               | string (required) | Configures the group name of the RADIUS servers | os9  |
| ``group.host``               | dictionary        | Configures the RADIUS server host in the group (see ``host.*``) | os9 |
| ``host.ip``                  | string            | Configures the RADIUS server host address in the group | os9  |
| ``host.key``                 | string (required): 0,7,LINE | Configures the authentication key | os9  |
| ``host.key_string``          | string: 7,0            | Configures the user key string; variable takes the hidden user key string if value is 7; variable takes the unencrypted user key (clear-text) if value is 0; variable supported only if *host.key* is 7 or 0 | os9 |
| ``host.retransmit``          | integer           | Configures the number of retransmissions | os9 |
| ``host.auth_port``           | integer           | Configures the authentication port (0 to 65535) | os9  |
| ``host.timeout``             | integer           | Configures the timeout for retransmissions | os9  |
| ``host.state``               | string: present,absent         | Removes the host from group of RADIUS server if set to absent | os9 |
| ``group.vrf``                | dictionary        | Configures the VRF for RADIUS servers in the group (see ``vrf.*``) | os9 |  
| ``vrf.vrf_name``             | string (required) | Configures the name of VRF for the RADIUS server group | os9 |
| ``vrf.source_intf``          | integer           | Configures the source interface for outgoing packets from servers in the group | os9 |
| ``vrf.state``                | string: present,absent         | Removes the VRF from group of RADIUS servers if set to absent | os9  |
| ``group.state``              | string: present,absent         | Removes the RADIUS server group if set to absent | os9  |
| ``radius_server.host``       | dictionary        | Configures the RADIUS server host (see ``host.*``) | os9  |
| ``host.ip``                  | string            | Configures the RADIUS server host address | os9  |
| ``host.key``                 | string (required); 0,7,LINE           | Configures the authentication key | os9  |
| ``host.key_string``          | string            | Configures the user key string; variable takes the hidden user key string if value is 7; variable takes the unencrypted user key (clear-text) if value is 0; variable supported only if *host.key* is 7 or 0 | os9 |
| ``host.retransmit``          | integer           | Configures the number of retransmissions | os9 |
| ``host.auth_port``           | integer           | Configures the authentication port (0 to 65535)  | os9  |
| ``host.timeout``             | integer           | Configures timeout for retransmissions | os9 |
| ``host.state``               | string: present,absent         | Removes the RADIUS server host if set to absent | os9  |
| ``auth.key``                 | string (required); 0,7,LINE           | Configures the authentication key | os9  |
| ``tacacs_server``            | dictionary        | Configures the TACACS server (see ``tacacs_server.*``)| os9 |
| ``tacacs_server.key``        | string (required): 0,7,LINE           | Configures the authentication key for TACACS server | os9 |
| ``tacacs_server.key_string`` | string            | Configures the user key string; variable takes the hidden user key string if value is 7; variable takes the unencrypted user key (clear-text) if value is 0; variable supported only if *tacacs_server.key* is 7 or 0 | os9 |
| ``tacacs_server.group``      | dictionary        | Configures the group of TACACS servers (see ``group.*``) | os9 |
| ``group.name``               | string (required)  | Configures the group name of the TACACS servers | os9  |
| ``group.host``               | dictionary        | Configures the TACACS server host in the group (see ``host.*``) | os9  |
| ``host.ip``                  | string            | Configures the TACACS server host address in the group | os9  |
| ``host.key``                 | string (required): 0,7,LINE           | Configures the authentication key of the TACACS server host | os9 |
| ``host.key_string``          | string            | Configures the user key string; variable takes the hidden user key string if value is 7; variable takes the unencrypted user key (clear-text) if value is 0; variable supported only *host.key* is 7 or 0 | os9 |
| ``host.retransmit``          | integer           | Configures the number of retransmissions | os9 |
| ``host.auth_port``           | integer           | Configures the authentication port (0 to 65535) | os9  |
| ``host.timeout``             | integer           | Configures timeout for retransmissions | os9  |
| ``host.state``               | string: present,absent         | Removes the host from group of TACACS server if set to absent | os9 |
| ``group.vrf``                | dictionary        | Configures VRF for TACACS servers in the group (see ``vrf.*``) | os9  |
| ``vrf.vrf_name``             | string (required) | Configures the name of VRF for TACACS server group | os9 |
| ``vrf.source_intf``          | integer           | Configures source interface for outgoing packets from servers in the group | os9 |
| ``vrf.state``                | string: present,absent         | Removes the VRF from group of TACACS server if set to absent | os9 |
| ``group.state``              | string: present,absent         | Removes the TACACS server group if set to absent | os9 |
| ``tacacs_server.host``       | dictionary        | Configures the TACACS server host (see ``host.*``) | os9 |
| ``host.ip``                  | string            | Configures the TACACS sever host address | os9 |
| ``host.key``                 | string (required): 0,7,LINE           | Configures the authentication key | os9 |
| ``host.key_string``          | string            | Configures the user key string; variable takes the hidden user key string if value is 7; variable takes the unencrypted user key (clear-text) if value is 0; variable supported only if *host.key* is 7 or 0 | os9  |
| ``host.retransmit``          | integer           | Configures the number of retransmissions | os9 |
| ``host.auth_port``           | integer           | Configures the authentication port (0 to 65535) | os9 |
| ``host.timeout``             | integer           | Configures the timeout for retransmissions | os9 |
| ``host.state``               | string: present,absent         | Removes the TACACS server host if set to absent | os9 |
| ``aaa_accounting``           | dictionary        | Configures accounting parameters (see ``aaa_accounting.*``) | os9 |
| ``aaa_accounting.commands``  | list              | Configures accounting for EXEC (shell) and config commands (see ``commands.*``) | os9 |
| ``commands.enable_level``    | integer           | Configures enable level for accounting of commands | os9 |
| ``commands.role_name``       | string            | Configures user role for accounting of commands; variable is mutually exclusive with ``enable_level`` | os9 |
| ``commands.accounting_list_name`` | integer      | Configures named accounting list for commands | os9  |
| ``commands.no_accounting``   | boolean           | Configures no accounting of commands | os9 |
| ``commands.record_option``   | string: start-stop,stop-only,wait-start        | Configures options to record data | os9 |
| ``commands.state``           | string: present,absent         | Removes the named accounting list for the commands if set to absent | os9 |
| ``aaa_accounting.exec``      | list              | Configures accounting for EXEC (shell) commands (see ``exec.*``) | os9 |
| ``exec.accounting_list_name`` | string           | Configures named accounting list for EXEC (shell) commands | os9 |
| ``exec.no_accounting``       | boolean           | Configures no accounting of EXEC (shell) commands | os9  |
| ``exec.record_option``       | string: start-stop,stop-only,wait-start      | Configures options to record data | os9 |
| ``exec.state``               | string: present,absent         | Removes the named accounting list for the EXEC (shell) commands if set to absent | os9 |
| ``aaa_accounting.suppress``  | boolean           | Suppresses accounting for users with NULL username | os9|
| ``aaa_accounting.dot1x``     | string: none,start-stop,stop-only,wait-start        | Configures accounting for dot1x events | os9 |
| ``aaa_accounting.rest``      | string:none,start-stop,stop-only,wait-start        | Configures accounting for REST interface events | os9  |
| ``aaa_authorization``        | dictionary        | Configures authorization parameters (see ``aaa_authorization.*``) | os9  |
| ``aaa_authorization.commands`` | list            | Configures authorization for EXEC (shell) and config commands (see ``commands.*``)| os9 |
| ``commands.enable_level``    | integer           | Configures enable level for authorization of commands | os9  |
| ``commands.role_name``       | string            | Configures user role for authorization of commands; mutually exclusive with ``enable_level`` | os9  |
| ``commands.authorization_list_name`` | string         | Configures named authorization list for commands | os9 |
| ``commands.authorization_method`` | string: none         | Configures no authorization of commands | os9  |
| ``commands.use_data``        | string: local,tacacs+        | Configures data used for authorization | os9  |
| ``commands.state``           | string: present,absent         | Removes the named authorization list for the commands if set to absent | os9 |
| ``aaa_authorization.config_commands`` | boolean        | Configures authorization for configuration mode commands | os9 |
| ``aaa_authorization.role_only`` | boolean        | Configures validation of authentication mode for user role | os9 |
| ``aaa_authorization.exec``   | list              | Configures authorization for EXEC (shell) commands (see ``exec.*``) | os9 |
| ``exec.authorization_list_name`` | string        | Configures named authorization list for EXEC (shell) commands | os9  |
| ``exec.authorization_method`` | string: none         | Configures no authorization of EXEC (shell) commands | os9  |
| ``exec.use_data``            | string: local,tacacs+        | Configures data used for authorization | os9 |
| ``exec.state``               | string: present,absent         | Removes the named authorization list for the EXEC (shell) commands if set to absent | os9  |
| ``aaa_authorization.network``     | string: none,radius,ias        | Configures authorization for network events | os9 |
| ``aaa_authentication``       | dictionary        | Configures authentication parameters (see ``aaa_authentication.*``) | os9  |
| ``aaa_radius``                | dictionary       | Configures AAA for RADIUS group of servers (see ``aaa_radius.*``) | os9 |
| ``aaa_radius.group``          | string           | Configures name of the RADIUS group of servers for AAA | os9  |
| ``aaa_radius.auth_method``    | string: pap,mschapv2    | Configures authentication method of RADIUS group of servers for AAA | os9 |
| ``aaa_tacacs``                | dictionary       | Configures AAA for TACACS group of servers (see ``aaa_tacacs.*``) | os9 |
| ``aaa_tacacs.group``          | string           | Configures name of the TACACS group of servers for AAA | os9  |
| ``aaa_authentication.auth_list`` | list        | Configures named authentication list for hosts (see ``host.*``) | os9 |
| ``auth_list.name``           | string         | Configures named authentication list | os9 |
| ``auth_list.login_or_enable`` | string: enable,login         | Configures authentication list for login or enable | os9  |
| ``auth_list.server``         | string: radius,tacacs+         | Configures AAA to use this list of all server hosts | os9 |
| ``auth_list.use_password``   | string: line,local,enable,none         | Configures password to use for authentication | os9 |
| ``auth_list.state``          | string: present,absent         | Removes the named authentication list if set to absent | os9 |
| ``aaa_authentication.dot1x``     | string: none,radius,ias        | Configures authentication for dot1x events | os9 |
| ``line_terminal``            | dictionary | Configures the terminal line (see ``line_terminal.*``) | os9 |
| ``line_terminal.<terminal>`` | dictionary | Configures the primary or virtual terminal line (see ``<terminal>.*``); value can be console <line_number>, vty <line_number> | os9 |
| ``<terminal>.authorization`` | dictionary | Configures authorization parameters of line terminal (see ``authorization.*``) | os9 |
| ``authorization.commands``   | list        | Configures authorization for EXEC (shell) and config commands (see ``commands.*``) | os9  |
| ``commands.enable_level``    | integer         | Configures enable level for authorization of commands at line terminal | os9  |
| ``commands.role_name``       | string         | Configures user role for authorization of commands at line terminal; mutually exclusive with `enable_level` |  os9 |
| ``commands.authorization_list_name`` | string         | Configures named authorization list for commands | os9 |
| ``commands.state``           | string: present,absent         | Removes the authorization of commands from line terminal if set to absent | os9 |
| ``authorization.exec``       | list        | Configures authorization for EXEC (shell) commands at line terminal (see ``exec.*``) | os9 |
| ``exec.authorization_list_name`` | string         | Configures named authorization list for EXEC (shell) commands | os9 |
| ``exec.state``               | string: present,absent         | Removes the authorization of EXEC (shell) from line terminal if set to absent | os9  |
| ``<terminal>.accounting`` | dictionary | Configures accounting parameters of line terminal (see ``accounting.*``) | os9  |
| ``accounting.commands``      | list        | Configures accounting for EXEC (shell) and config commands (see ``commands.*``) | os9 |
| ``commands.enable_level``    | integer         | Configures enable level for accounting of commands at line terminal | os9|
| ``commands.role_name``       | string         | Configures user role for accounting of commands at line terminal; mutually exclusive with ``enable_level`` | os9 |
| ``commands.accounting_list_name`` | string         | Configures named accounting list for commands | os9 |
| ``commands.state``           | string: present,absent         | Removes the accounting of commands from line terminal if set to absent | os9|
| ``accounting.exec``          | list        | Configures accounting for EXEC (shell) commands at line terminal (see ``exec.*``) | os9  |
| ``exec.accounting_list_name`` | string         | Configures named accounting list for EXEC (shell) commands | os9 |
| ``exec.state``               | string: present,absent         | Removes the accounting of EXEC (shell) from line terminal if set to absent | os9 |
| ``<terminal>.authentication`` | dictionary | Configures authentication parameters of line terminal (see ``authentication.*``) | os9 |
| ``authentication.enable``    | string         | Configures the authentication list for privilege-level password authentication | os9 |
| ``authentication.login``     | string         | Configures the authentication list for password checking | os9  |
| ``client.ip``                | string         | Configures the client IP for the radius server | os9 |
| ``client.key``               | string (required): 0,7,LINE | Configures the authentication key for the RADIUS server | os9 |
> **NOTE**: Asterisk (*) denotes the default value if none is specified.

Connection variables
--------------------

Ansible Dell EMC network roles require connection information to establish communication with the nodes in your inventory. This information can exist in the Ansible *group_vars* or *host_vars* directories or inventory, or in the playbook itself.

| Key         | Required | Choices    | Description                                         |
|-------------|----------|------------|-----------------------------------------------------|
| ``ansible_host`` | yes      |            | Specifies the hostname or address for connecting to the remote device over the specified transport |
| ``ansible_port`` | no       |            | Specifies the port used to build the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_PORT` option is used; it defaults to 22 |
| ``ansible_ssh_user`` | no       |            | Specifies the username that authenticates the CLI login for the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_USER` environment variable value is used  |
| ``ansible_ssh_pass`` | no       |            | Specifies the password that authenticates the connection to the remote device.  |
| ``ansible_become`` | no       | yes, no\*   | Instructs the module to enter privileged mode on the remote device before sending any commands; if value is unspecified, the `ANSIBLE_BECOME` environment variable value is used, and the device attempts to execute all commands in non-privileged mode |
| ``ansible_become_method`` | no       | enable, sudo\*   | Instructs the module to allow the become method to be specified for handling privilege escalation; if value is unspecified, the `ANSIBLE_BECOME_METHOD` environment variable value is used |
| ``ansible_become_pass`` | no       |            | Specifies the password to use if required to enter privileged mode on the remote device; if ``ansible_become`` is set to no this key is not applicable |
| ``ansible_network_os`` | yes      | os9, null\*  | Loads the correct terminal and cliconf plugins to communicate with the remote device |

> **NOTE**: Asterisk (*) denotes the default value if none is specified.

Example playbook
----------------

This example uses the *os9_aaa* role to configure AAA for RADIUS and TACACS servers. It creates a *hosts* file with the switch details and corresponding variables. The hosts file should define the `ansible_network_os` variable with the corresponding Dell EMC OS0 name. 

When `os9_cfg_generate` is set to true, the variable generates the configuration commands as a .part file in the *build_dir* path. By default, it is set to false and it writes a simple playbook that only references the *os9_aaa* role.

**Sample hosts file**

    leaf1 ansible_host= <ip_address> 

**Sample host_vars/leaf1**

    hostname: leaf1
    ansible_become: yes
    ansible_become_method: xxxxx
    ansible_become_pass: xxxxx
    ansible_ssh_user: xxxxx
    ansible_ssh_pass: xxxxx
    ansible_network_os: dellemc.os9.os9
    build_dir: ../temp/os9

    os9_aaa:
      radius_server:
          key: radius
          retransmit: 5
          timeout: 40
          deadtime: 2300
          group:
            - name: RADIUS
              host:
                - ip: 2001:4898:f0:f09b::1002
                  key: 0
                  key_string: aaaa
                  retransmit: 5
                  auth_port: 3
                  timeout: 2
                  state: present
              vrf:
                vrf_name: test
                source_intf: fortyGigE 1/2
                state: absent
              state: present
          host:
            - ip: 2001:4898:f0:f09b::1002
              key: xxx
              retransmit: 5
              auth_port: 3
              timeout: 2
              state: present
      tacacs_server:
          key: 7
          key_string: 9ea8ec421c2e2e5bec757f44205015f6d81e83a4f0aa52fa
          group:
            - name: TACACS
              host:
                - ip: 2001:4898:f0:f09b::1000
                  key: 0
                  key_string: aaa
                  retransmit: 6
                  auth_port: 3
                  timeout: 2
                  state: present
              vrf:
                vrf_name: tes
                source_intf: fortyGigE 1/3
                state: present
              state: present
          host:
            - ip: 2001:4898:f0:f09b::1000
              key: 0
              key_string: aa
              retransmit: 5
              auth_port: 3
              timeout: 2
              state: present
      aaa_accounting:
          commands:
              - enable_level: 2
                accounting_list_name: aa
                record_option: start-stop
                state: present
              - role_name: netadmin
                accounting_list_name: aa
                no_accounting: none
          suppress: True
          exec:
              - accounting_list_name: aaa
                no_accounting: true
                state: present
          dot1x: none
          rest: none
      aaa_authorization:
          commands:
              - enable_level: 2
                authorization_list_name: aa
                use_data: local
                state: present
              - role_name: netadmin
                authorization_list_name: aa
                authorization_method: none
                use_data: local
          config_commands: True
          role_only:
          exec:
              - authorization_list_name: aaa
                authorization_method: if-authenticated
                use_data: local
                state: present
      aaa_authentication:
          auth_list:
            - name: default
              login_or_enable: login
              server: radius
              use_password: local
              state: present
            - name: console
              server: tacacs+
              login_or_enable: login
              use_password: local
      aaa_radius:
          group: RADIUS
          auth_method: pap
      aaa_tacacs:
          group: TACACS
      line_terminal:
        vty 0:
          authorization:
            commands:
               - enable_level: 2
                 authorization_list_name: aa
                 state: present
               - role_name: netadmin
                 authorization_list_name: aa
                 state: present
            exec:
               - authorization_list_name: aa
                 state: present
          accounting:
            commands:
               - enable_level: 2
                 accounting_list_name: aa
                 state: present
               - role_name: netadmin
                 accounting_list_name: aa
                 state: absent
            exec:
               accounting_list_name: aa
               state: present
          authentication:
            enable:
            login: console

**Simple playbook to setup system — leaf.yaml**

    - hosts: leaf1
      roles:
         - dellemc.os9.os9_aaa

**Run**

    ansible-playbook -i hosts leaf.yaml

(c) 2017-2020 Dell Inc. or its subsidiaries. All rights reserved.
