LAG role
========

This role facilitates the configuration of link aggregation group (LAG) attributes, and supports the creation and deletion of a LAG and its member ports. It also supports the configuration of an interface type as a static or dynamic LAG, hash scheme in os6 devices, and minimum required link. This role is abstracted for Dell EMC PowerSwitch platforms running Dell EMC OS6.

The LAG role requires an SSH connection for connectivity to a Dell EMC OS6 device. You can use any of the built-in OS connection variables.


Role variables
--------------

- Role is abstracted using the `ansible_network_os` variable that can take `dellemc.os6.os6` as a value
- Object drives the tasks in this role
- `os6_lag` (dictionary) contains the hostname (dictionary)
- Hostname is the value of the *hostname* variable that corresponds to the name of the OS device
- Any role variable with a corresponding state variable setting to absent negates the configuration of that variable
- Setting an empty value to any variable negates the corresponding configuration
- `os6_lag` (dictionary) holds a dictionary with the port-channel ID key in `Po <ID>` format (1 to 128 for os6)
- Variables and values are case-sensitive

**port-channel ID keys**

| Key        | Type                      | Description                                             | Support               |
|------------|---------------------------|---------------------------------------------------------|-----------------------|
| ``type``      | string: static,dynamic      | Configures the interface either as a static or dynamic LAG           | os6 |
| ``min_links`` | integer                       | Configures the minimum number of links in the LAG that must be in *operup* status (1 to 8), field needs to be left blank to remove the minimum number of links | os6 |
| ``hash`` | integer | Configures the hash value for OS6 devices (1 to 7), field needs to be left blank to remove the hash value | os6 |
| ``channel_members``  | list  | Specifies the list of port members to be associated to the port channel (see ``channel_members.*``) | os6 |
| ``channel_members.port`` | string  | Specifies valid OS6 interface names to be configured as port channel members | os6 |
| ``channel_members.state`` | string: absent,present | Deletes the port member association if set to absent | os6 |
| ``state``  | string: absent,present\*           | Deletes the LAG corresponding to the port channel ID if set to absent | os6 |

> **NOTE**: Asterisk (\*) denotes the default value if none is specified.

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
| ``ansible_become_method`` | no       | enable, sudo\*   | Instructs the module to allow the become method to be specified for handling privilege escalation; if value is unspecified, the `ANSIBLE_BECOME_METHOD` environment variable value is used. |
| ``ansible_become_pass`` | no       |            | Specifies the password to use if required to enter privileged mode on the remote device; if ``ansible_become`` is set to no this key is not applicable. |
| ``ansible_network_os`` | yes      | os6, null\*  | Loads the correct terminal and cliconf plugins to communicate with the remote device. |
> **NOTE**: Asterisk (\*) denotes the default value if none is specified.

Example playbook
----------------

This example uses the *os6_lag* role to setup port channel ID and description, and configures hash algorithm and minimum links for the LAG. Channel members can be configured for the port-channel either in static or dynamic mode. You can also delete the LAG with the port channel ID or delete the members associated to it. This example creates a *hosts* file with the switch details and corresponding variables. The hosts file should define the `ansible_network_os` variable with corresponding Dell EMC OS6 name.

When `os6_cfg_generate` is set to true, the variable generates the configuration commands as a .part file in *build_dir* path. By default, the variable is set to false. It writes a simple playbook that only references the *os6-lag* role.

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

    os6_lag:
        Po 127:
          type: static
          hash: 7
          min_links: 3
          channel_members:
            - port: Fo4/0/1
              state: present
            - port: Fo4/0/1
              state: present
          state: present

**Simple playbook to setup system — switch1.yaml**

    - hosts: switch1
      roles:
         - dellemc.os6.os6_lag

**Run**

    ansible-playbook -i hosts switch1.yaml

(c) 2017-2020 Dell Inc. or its subsidiaries. All rights reserved.
