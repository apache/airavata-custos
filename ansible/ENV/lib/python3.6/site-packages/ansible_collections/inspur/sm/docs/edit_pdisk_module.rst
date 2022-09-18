.. Document meta

:orphan:

.. Anchors

.. _ansible_collections.inspur.sm.edit_pdisk_module:

.. Anchors: short name for ansible.builtin

.. Anchors: aliases



.. Title

inspur.sm.edit_pdisk -- Set physical disk.
++++++++++++++++++++++++++++++++++++++++++

.. Collection note

.. note::
    This plugin is part of the `inspur.sm collection <https://galaxy.ansible.com/inspur/sm>`_.

    To install it use: :code:`ansible-galaxy collection install inspur.sm`.

    To use it in a playbook, specify: :code:`inspur.sm.edit_pdisk`.

.. version_added

.. versionadded:: 0.1.0 of inspur.sm

.. contents::
   :local:
   :depth: 1

.. Deprecated


Synopsis
--------

.. Description

- Set physical disk on Inspur server.

.. Aliases


.. Requirements


.. Options

Parameters
----------

.. raw:: html

    <table  border=0 cellpadding=0 class="documentation-table">
        <tr>
            <th colspan="2">Parameter</th>
            <th>Choices/<font color="blue">Defaults</font></th>
                        <th width="100%">Comments</th>
        </tr>
                    <tr>
                                                                <td colspan="2">
                    <div class="ansibleOptionAnchor" id="parameter-ctrl_id"></div>
                    <b>ctrl_id</b>
                    <a class="ansibleOptionLink" href="#parameter-ctrl_id" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">integer</span>
                                                                    </div>
                                                        </td>
                                <td>
                                                                                                                                                            </td>
                                                                <td>
                                            <div>Raid controller ID.</div>
                                            <div>Required when <em>Info=None</em>.</div>
                                                        </td>
            </tr>
                                <tr>
                                                                <td colspan="2">
                    <div class="ansibleOptionAnchor" id="parameter-device_id"></div>
                    <b>device_id</b>
                    <a class="ansibleOptionLink" href="#parameter-device_id" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">integer</span>
                                                                    </div>
                                                        </td>
                                <td>
                                                                                                                                                            </td>
                                                                <td>
                                            <div>physical drive id.</div>
                                            <div>Required when <em>Info=None</em>.</div>
                                                        </td>
            </tr>
                                <tr>
                                                                <td colspan="2">
                    <div class="ansibleOptionAnchor" id="parameter-host"></div>
                    <b>host</b>
                    <a class="ansibleOptionLink" href="#parameter-host" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                                                                    </div>
                                                        </td>
                                <td>
                                                                                                                                                            </td>
                                                                <td>
                                            <div>Specifies the DNS host name or address for connecting to the remote device over the specified transport.  The value of host is used as the destination address for the transport.</div>
                                                        </td>
            </tr>
                                <tr>
                                                                <td colspan="2">
                    <div class="ansibleOptionAnchor" id="parameter-info"></div>
                    <b>info</b>
                    <a class="ansibleOptionLink" href="#parameter-info" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                                                                    </div>
                                                        </td>
                                <td>
                                                                                                                            <ul style="margin: 0; padding: 0"><b>Choices:</b>
                                                                                                                                                                <li>show</li>
                                                                                    </ul>
                                                                            </td>
                                                                <td>
                                            <div>Show controller and pdisk info.</div>
                                                        </td>
            </tr>
                                <tr>
                                                                <td colspan="2">
                    <div class="ansibleOptionAnchor" id="parameter-option"></div>
                    <b>option</b>
                    <a class="ansibleOptionLink" href="#parameter-option" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                                                                    </div>
                                                        </td>
                                <td>
                                                                                                                            <ul style="margin: 0; padding: 0"><b>Choices:</b>
                                                                                                                                                                <li>UG</li>
                                                                                                                                                                                                <li>UB</li>
                                                                                                                                                                                                <li>OFF</li>
                                                                                                                                                                                                <li>FAIL</li>
                                                                                                                                                                                                <li>RBD</li>
                                                                                                                                                                                                <li>ON</li>
                                                                                                                                                                                                <li>JB</li>
                                                                                                                                                                                                <li>ES</li>
                                                                                                                                                                                                <li>EM</li>
                                                                                                                                                                                                <li>EN</li>
                                                                                                                                                                                                <li>ET</li>
                                                                                                                                                                                                <li>LOC</li>
                                                                                                                                                                                                <li>STL</li>
                                                                                    </ul>
                                                                            </td>
                                                                <td>
                                            <div>Set operation options fo physical disk,</div>
                                            <div>UG is Unconfigured Good,UB is Unconfigured Bad,</div>
                                            <div>OFF is offline,FAIL is Failed,RBD is Rebuild,</div>
                                            <div>ON is Online,JB is JBOD,ES is Drive Erase stop,</div>
                                            <div>EM is Drive Erase Simple,EN is Drive Erase Normal,</div>
                                            <div>ET is Drive Erase Through,LOC is Locate,STL is Stop Locate.</div>
                                            <div>Required when <em>Info=None</em>.</div>
                                                        </td>
            </tr>
                                <tr>
                                                                <td colspan="2">
                    <div class="ansibleOptionAnchor" id="parameter-password"></div>
                    <b>password</b>
                    <a class="ansibleOptionLink" href="#parameter-password" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                                                                    </div>
                                                        </td>
                                <td>
                                                                                                                                                            </td>
                                                                <td>
                                            <div>Specifies the password to use to authenticate the connection to the remote device. If the value is not specified in the task, the value of environment variable <code>ANSIBLE_NET_PASSWORD</code> will be used instead.</div>
                                                        </td>
            </tr>
                                <tr>
                                                                <td colspan="2">
                    <div class="ansibleOptionAnchor" id="parameter-provider"></div>
                    <b>provider</b>
                    <a class="ansibleOptionLink" href="#parameter-provider" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">dictionary</span>
                                                                    </div>
                                                        </td>
                                <td>
                                                                                                                                                            </td>
                                                                <td>
                                            <div>A dict object containing connection details.</div>
                                                        </td>
            </tr>
                                        <tr>
                                                    <td class="elbow-placeholder"></td>
                                                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="parameter-provider/host"></div>
                    <b>host</b>
                    <a class="ansibleOptionLink" href="#parameter-provider/host" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                                                                    </div>
                                                        </td>
                                <td>
                                                                                                                                                            </td>
                                                                <td>
                                            <div>Specifies the DNS host name or address for connecting to the remote device over the specified transport.  The value of host is used as the destination address for the transport.</div>
                                                        </td>
            </tr>
                                <tr>
                                                    <td class="elbow-placeholder"></td>
                                                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="parameter-provider/password"></div>
                    <b>password</b>
                    <a class="ansibleOptionLink" href="#parameter-provider/password" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                                                                    </div>
                                                        </td>
                                <td>
                                                                                                                                                            </td>
                                                                <td>
                                            <div>Specifies the password to use to authenticate the connection to the remote device. If the value is not specified in the task, the value of environment variable <code>ANSIBLE_NET_PASSWORD</code> will be used instead.</div>
                                                        </td>
            </tr>
                                <tr>
                                                    <td class="elbow-placeholder"></td>
                                                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="parameter-provider/username"></div>
                    <b>username</b>
                    <a class="ansibleOptionLink" href="#parameter-provider/username" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                                                                    </div>
                                                        </td>
                                <td>
                                                                                                                                                            </td>
                                                                <td>
                                            <div>Configures the username to use to authenticate the connection to the remote device. If the value is not specified in the task, the value of environment variable <code>ANSIBLE_NET_USERNAME</code> will be used instead.</div>
                                                        </td>
            </tr>
                    
                                <tr>
                                                                <td colspan="2">
                    <div class="ansibleOptionAnchor" id="parameter-username"></div>
                    <b>username</b>
                    <a class="ansibleOptionLink" href="#parameter-username" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                                                                    </div>
                                                        </td>
                                <td>
                                                                                                                                                            </td>
                                                                <td>
                                            <div>Configures the username to use to authenticate the connection to the remote device. If the value is not specified in the task, the value of environment variable <code>ANSIBLE_NET_USERNAME</code> will be used instead.</div>
                                                        </td>
            </tr>
                        </table>
    <br/>

.. Notes


.. Seealso


.. Examples

Examples
--------

.. code-block:: yaml+jinja

    
    - name: Edit pdisk test
      hosts: ism
      connection: local
      gather_facts: no
      vars:
        ism:
          host: "{{ ansible_ssh_host }}"
          username: "{{ username }}"
          password: "{{ password }}"

      tasks:

      - name: "Show pdisk information"
        inspur.sm.edit_pdisk:
          info: "show"
          provider: "{{ ism }}"

      - name: "Edit pdisk"
        inspur.sm.edit_pdisk:
          ctrl_id: 0
          device_id: 1
          option: "LOC"
          provider: "{{ ism }}"




.. Facts


.. Return values

Return Values
-------------
Common return values are documented :ref:`here <common_return_values>`, the following are the fields unique to this module:

.. raw:: html

    <table border=0 cellpadding=0 class="documentation-table">
        <tr>
            <th colspan="1">Key</th>
            <th>Returned</th>
            <th width="100%">Description</th>
        </tr>
                    <tr>
                                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="return-changed"></div>
                    <b>changed</b>
                    <a class="ansibleOptionLink" href="#return-changed" title="Permalink to this return value"></a>
                    <div style="font-size: small">
                      <span style="color: purple">boolean</span>
                                          </div>
                                    </td>
                <td>always</td>
                <td>
                                            <div>Check to see if a change was made on the device.</div>
                                        <br/>
                                    </td>
            </tr>
                                <tr>
                                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="return-message"></div>
                    <b>message</b>
                    <a class="ansibleOptionLink" href="#return-message" title="Permalink to this return value"></a>
                    <div style="font-size: small">
                      <span style="color: purple">string</span>
                                          </div>
                                    </td>
                <td>always</td>
                <td>
                                            <div>Messages returned after module execution.</div>
                                        <br/>
                                    </td>
            </tr>
                                <tr>
                                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="return-state"></div>
                    <b>state</b>
                    <a class="ansibleOptionLink" href="#return-state" title="Permalink to this return value"></a>
                    <div style="font-size: small">
                      <span style="color: purple">string</span>
                                          </div>
                                    </td>
                <td>always</td>
                <td>
                                            <div>Status after module execution.</div>
                                        <br/>
                                    </td>
            </tr>
                        </table>
    <br/><br/>

..  Status (Presently only deprecated)


.. Authors

Authors
~~~~~~~

- WangBaoshan (@ISIB-group)



.. Parsing errors

