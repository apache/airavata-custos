QoS role
========

This role facilitates the configuration of quality of service (QoS) attributes like policy-map and class-map. This role is abstracted for Dell EMC PowerSwitch platforms running Dell EMC OS6.

The QoS role requires an SSH connection for connectivity to a Dell EMC OS6 device. You can use any of the built-in OS connection variables.


Role variables
--------------

- Role is abstracted using the `ansible_network_os` variable that can take a `dellemc.os6.os6` as a value
- If `os6_cfg_generate` is set to true, the variable generates the role configuration commands in a file
- Any role variable with a corresponding state variable set to absent negates the configuration of that variable 
- Setting an empty value for any variable negates the corresponding configuration
- Variables and values are case-sensitive

**os6_qos keys**

| Key        | Type                      | Description                                             | Support               |
|------------|---------------------------|---------------------------------------------------------|-----------------------|
| ``policy_map`` | list | Configures the policy-map (see ``policy_map.*``) | os6 |
| ``policy_map.name`` | string (required)        | Configures the policy-map name  | os6 |
| ``policy_map.type`` | string: in, out in os6   | Configures the policy-map type  | os6 |
| ``policy_map.class_instances`` | list | Specifies the class instances for the policy | os6 |
| ``class_instances.name`` | string | Specifies name of class instance | os6 |
| ``class_instances.policy`` | list | Specifies list of associated policies for the class | os6 |
| ``policy_map.state`` | string: present\*,absent   | Deletes the policy-map if set to absent  | os6 |
| ``class_map`` | list | Configures the class-map (see ``class_map.*``) | os6 |
| ``class_map.name`` | string (required)        | Configures the class-map name  | os6 |
| ``class_map.type`` | string: match-all, match-any in os6    | Configures the class-map type  | os6 |
| ``class-map.match_condition`` | list | Specifies the type of match-conditions required for the class | os6 |
| ``class_map.state`` | string: present\*,absent   | Deletes the class-map if set to absent  | os6 |

> **NOTE**: Asterisk (\*) denotes the default value if none is specified. 

Connection variables
--------------------

Ansible Dell EMC Networking roles require connection information to establish communication with the nodes in your inventory. This information can exist in the Ansible *group_vars* or *host_vars* directories, or inventory or in the playbook itself.

| Key         | Required | Choices    | Description                                         |
|-------------|----------|------------|-----------------------------------------------------|
| ``ansible_host`` | yes      |            | Specifies the hostname or address for connecting to the remote device over the specified transport |
| ``ansible_port`` | no       |            | Specifies the port used to build the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_PORT` option is used; it defaults to 22 |
| ``ansible_ssh_user`` | no       |            | Specifies the username that authenticates the CLI login for the connection to the remote device; if value is unspecified, the `ANSIBLE_REMOTE_USER` environment variable value is used  |
| ``ansible_ssh_pass`` | no       |            | Specifies the password that authenticates the connection to the remote device.  |
| ``ansible_become`` | no       | yes, no\*   | Instructs the module to enter privileged mode on the remote device before sending any commands; if value is unspecified, the `ANSIBLE_BECOME` environment variable value is used, and the device attempts to execute all commands in non-privileged mode |
| ``ansible_become_method`` | no       | enable, sudo\*   | Instructs the module to allow the become method to be specified for handling privilege escalation; if value is unspecified, the `ANSIBLE_BECOME_METHOD` environment variable value is used |
| ``ansible_become_pass`` | no       |            | Specifies the password to use if required to enter privileged mode on the remote device; if ``ansible_become`` is set to no this key is not applicable |
| ``ansible_network_os`` | yes      | os6, null\*  | Loads the correct terminal and cliconf plugins to communicate with the remote device |

> **NOTE**: Asterisk (\*) denotes the default value if none is specified.

Example playbook
----------------

This example uses the *os6_qos* role to configure the policy-map class-map. It creates a *hosts* file with the switch details and corresponding variables. The hosts file should define the `ansible_network_os` variable with corresponding Dell EMC OS6 name. 

When `os6_cfg_generate` is set to true, the variable generates the configuration commands as a .part file in *build_dir* path. By default, the variable is set to false. It writes a simple playbook that only references the *os6_qos* role. By including the role, you automatically get access to all of the tasks to configure QoS features. 

**Sample hosts file**
 
    switch1 ansible_host= <ip_address> 

**Sample host_vars/switch1**

    hostname: switch1
    ansible_become: yes
    ansible_become_method: enable
    ansible_ssh_user: xxxxx
    ansible_ssh_pass: xxxxx
    ansible_network_os: dellemc.os6.os6
    build_dir: ../temp/temp_os6
	  
    os6_qos:
      policy_map:
        - name: testpolicy
          type: qos
          class_instances:
            - name: video
              policy:
                - assign-queue 1
          state: present
      class_map:
        - name: testclass
          type: application
          match_condition:
            - ip dscp 26
          state: present
     
**Simple playbook to setup qos — switch1.yaml**

    - hosts: switch1
      roles:
         - dellemc.os6.os6_qos

**Run**

    ansible-playbook -i hosts switch1.yaml

(c) 2017-2020 Dell Inc. or its subsidiaries. All rights reserved.
