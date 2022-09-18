Users role
==========

This role facilitates the configuration of global system user attributes, and it supports the configuration of CLI users. This role is abstracted for Dell EMC PowerSwitch platforms running Dell EMC OS9.

The users role requires an SSH connection for connectivity to a Dell EMC OS9 device. You can use any of the built-in OS connection variables.

Role variables
--------------

- Role is abstracted using the `ansible_network_os` variable that can take `dellemc.os9.os9` as the value
- If `os9_cfg_generate` is set to true, the variable generates the role configuration commands in a file
- Any role variable with a corresponding state variable set to absent negates the configuration of that variable
- Setting an empty value for any variable negates the corresponding configuration
- Variables and values are case-sensitive

**os9_users list keys**

| Key        | Type                      | Description                                             | Support               |
|------------|---------------------------|---------------------------------------------------------|-----------------------|
| ``userrole`` | stirng (required) | Configures the role name which can be configured for users | os9 |
| ``userrole_state`` | string: absent,present\* | Deletes the user role with specified name if set to absent | os9 |
| ``userrole_inherit`` | string: netadmin,netoperator,secadmin,sysadmin\* | Specifies the existing role name to inherit the permissions | os9 |
| ``username`` | string (required)         | Configures the username which must adhere to specific format guidelines (valid usernames begin with A-Z, a-z, or 0-9 and can also contain `@#$%^&*-_= +;<>,.~` characters) | os9 |
| ``password`` | string                    | Configures the password set for the username; | os9 |
| ``role`` | string                    | Configures the role assigned to the user | os9 |
| ``privilege`` | int                | Configures the privilege level for the user (0 to 15); if this key is ommitted, the default privilege is 1 for both os9 | os9  |
| ``access_class`` | string       | Configures the access-class for the user | os9 |
| ``pass_key`` | integer: 0\*,7 | Configures the password as encrypted if set to 7 in os9 devices | os9 |
| ``secret`` | string | Configures line password as secret in os9 devices | os9 |
| ``secret_key`` | integer: 0\*,5 | Configures the secret line password using md5 encrypted algorithm | os9 |
| ``state`` | string: absent,present\*     | Deletes a user account if set to absent  | os9 |

> **NOTE**: Asterisk (\*) denotes the default value if none is specified. 

Connection variables
--------------------

Ansible Dell EMC network roles require connection information to establish communication with the nodes in your inventory. This information can exist in the Ansible *group_vars* or *host_vars* directories, or inventory or in the playbook itself.

| Key         | Required | Choices    | Description                                         |
|-------------|----------|------------|-----------------------------------------------------|
| ``ansible_host`` | yes      |            | Specifies the hostname or address for connecting to the remote device over the specified transport |
| ``ansible_port`` | no       |            | Specifies the port used to build the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_PORT` option is used; it defaults to 22 |
| ``ansible_ssh_user`` | no       |            | Specifies the username that authenticates the CLI login for the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_USER` environment variable value is used  |
| ``ansible_ssh_pass`` | no       |            | Specifies the password that authenticates the connection to the remote device  |
| ``ansible_become`` | no       | yes, no\*   | Instructs the module to enter privileged mode on the remote device before sending any commands; if value is unspecified, the `ANSIBLE_BECOME` environment variable value is used, and the device attempts to execute all commands in non-privileged mode |
| ``ansible_become_method`` | no       | enable, sudo\*   | Instructs the module to allow the become method to be specified for handling privilege escalation; if value is unspecified, the `ANSIBLE_BECOME_METHOD` environment variable value is used |
| ``ansible_become_pass`` | no       |            | Specifies the password to use if required to enter privileged mode on the remote device; if ``ansible_become`` is set to no this key is not applicable |
| ``ansible_network_os`` | yes      | os9, null\*  | Loads the correct terminal and cliconf plugins to communicate with the remote device |

> **NOTE**: Asterisk (\*) denotes the default value if none is specified.

Example playbook
----------------

This example uses the *os9_users* role to configure global system user attributes. The example creates a hosts file with the switch details and corresponding variables. The hosts file should define the `ansible_network_os` variable with corresponding Dell EMC OS9 name.

If `os9_cfg_generate` is set to true, the variable generates the role configuration commands in a file. It writes a simple playbook that only references the *os9_users* role. By including the role, you automatically get access to all of the tasks to configure user features. 

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
	  
    os9_users:
       - userrole: role1
         userrole_state: present
         userrole_inherit: sysadmin
       - username: u1
         password: test
         role: sysadmin
         privilege: 0
         state: absent
       - username: u1
         password: false
         privilege: 1
         access_class: a1
         role: netadmin
         state: present
       - username: u2
         secret: test1
         secret_key : 0
         access_class: a2
         privilege: 3
         role: sysadmin
         state: present

**Simple playbook to setup users — leaf.yaml**

    - hosts: leaf1
      roles:
         - dellemc.os9.os9_users

**Run**

    ansible-playbook -i hosts leaf.yaml

(c) 2017-2020 Dell Inc. or its subsidiaries. All rights reserved.
