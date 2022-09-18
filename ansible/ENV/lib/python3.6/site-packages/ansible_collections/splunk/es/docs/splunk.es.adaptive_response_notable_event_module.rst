.. _splunk.es.adaptive_response_notable_event_module:


*****************************************
splunk.es.adaptive_response_notable_event
*****************************************

**Manage Splunk Enterprise Security Notable Event Adaptive Responses**


Version added: 1.0.0

.. contents::
   :local:
   :depth: 1


Synopsis
--------
- This module allows for creation, deletion, and modification of Splunk Enterprise Security Notable Event Adaptive Responses that are associated with a correlation search




Parameters
----------

.. raw:: html

    <table  border=0 cellpadding=0 class="documentation-table">
        <tr>
            <th colspan="1">Parameter</th>
            <th>Choices/<font color="blue">Defaults</font></th>
            <th width="100%">Comments</th>
        </tr>
            <tr>
                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="parameter-"></div>
                    <b>asset_extraction</b>
                    <a class="ansibleOptionLink" href="#parameter-" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">list</span>
                    </div>
                </td>
                <td>
                        <ul style="margin: 0; padding: 0"><b>Choices:</b>
                                    <li><div style="color: blue"><b>src</b>&nbsp;&larr;</div></li>
                                    <li><div style="color: blue"><b>dest</b>&nbsp;&larr;</div></li>
                                    <li><div style="color: blue"><b>dvc</b>&nbsp;&larr;</div></li>
                                    <li><div style="color: blue"><b>orig_host</b>&nbsp;&larr;</div></li>
                        </ul>
                        <b>Default:</b><br/><div style="color: blue">["src", "dest", "dvc", "orig_host"]</div>
                </td>
                <td>
                        <div>list of assets to extract, select any one or many of the available choices</div>
                        <div>defaults to all available choices</div>
                </td>
            </tr>
            <tr>
                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="parameter-"></div>
                    <b>correlation_search_name</b>
                    <a class="ansibleOptionLink" href="#parameter-" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                         / <span style="color: red">required</span>
                    </div>
                </td>
                <td>
                </td>
                <td>
                        <div>Name of correlation search to associate this notable event adaptive response with</div>
                </td>
            </tr>
            <tr>
                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="parameter-"></div>
                    <b>default_owner</b>
                    <a class="ansibleOptionLink" href="#parameter-" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                    </div>
                </td>
                <td>
                </td>
                <td>
                        <div>Default owner of the notable event, if unset it will default to Splunk System Defaults</div>
                </td>
            </tr>
            <tr>
                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="parameter-"></div>
                    <b>default_status</b>
                    <a class="ansibleOptionLink" href="#parameter-" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                    </div>
                </td>
                <td>
                        <ul style="margin: 0; padding: 0"><b>Choices:</b>
                                    <li>unassigned</li>
                                    <li>new</li>
                                    <li>in progress</li>
                                    <li>pending</li>
                                    <li>resolved</li>
                                    <li>closed</li>
                        </ul>
                </td>
                <td>
                        <div>Default status of the notable event, if unset it will default to Splunk System Defaults</div>
                </td>
            </tr>
            <tr>
                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="parameter-"></div>
                    <b>description</b>
                    <a class="ansibleOptionLink" href="#parameter-" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                         / <span style="color: red">required</span>
                    </div>
                </td>
                <td>
                </td>
                <td>
                        <div>Description of the notable event, this will populate the description field for the web console</div>
                </td>
            </tr>
            <tr>
                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="parameter-"></div>
                    <b>drill_down_earliest_offset</b>
                    <a class="ansibleOptionLink" href="#parameter-" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                    </div>
                </td>
                <td>
                        <b>Default:</b><br/><div style="color: blue">"$info_min_time$"</div>
                </td>
                <td>
                        <div>Set the amount of time before the triggering event to search for related events. For example, 2h. Use &quot;$info_min_time$&quot; to set the drill-down time to match the earliest time of the search</div>
                </td>
            </tr>
            <tr>
                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="parameter-"></div>
                    <b>drill_down_latest_offset</b>
                    <a class="ansibleOptionLink" href="#parameter-" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                    </div>
                </td>
                <td>
                        <b>Default:</b><br/><div style="color: blue">"$info_max_time$"</div>
                </td>
                <td>
                        <div>Set the amount of time after the triggering event to search for related events. For example, 1m. Use &quot;$info_max_time$&quot; to set the drill-down time to match the latest time of the search</div>
                </td>
            </tr>
            <tr>
                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="parameter-"></div>
                    <b>drill_down_name</b>
                    <a class="ansibleOptionLink" href="#parameter-" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                    </div>
                </td>
                <td>
                </td>
                <td>
                        <div>Name for drill down search, Supports variable substitution with fields from the matching event.</div>
                </td>
            </tr>
            <tr>
                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="parameter-"></div>
                    <b>drill_down_search</b>
                    <a class="ansibleOptionLink" href="#parameter-" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                    </div>
                </td>
                <td>
                </td>
                <td>
                        <div>Drill down search, Supports variable substitution with fields from the matching event.</div>
                </td>
            </tr>
            <tr>
                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="parameter-"></div>
                    <b>identity_extraction</b>
                    <a class="ansibleOptionLink" href="#parameter-" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">list</span>
                    </div>
                </td>
                <td>
                        <ul style="margin: 0; padding: 0"><b>Choices:</b>
                                    <li><div style="color: blue"><b>user</b>&nbsp;&larr;</div></li>
                                    <li><div style="color: blue"><b>src_user</b>&nbsp;&larr;</div></li>
                        </ul>
                        <b>Default:</b><br/><div style="color: blue">["user", "src_user"]</div>
                </td>
                <td>
                        <div>list of identity fields to extract, select any one or many of the available choices</div>
                        <div>defaults to all available choices</div>
                </td>
            </tr>
            <tr>
                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="parameter-"></div>
                    <b>investigation_profiles</b>
                    <a class="ansibleOptionLink" href="#parameter-" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                    </div>
                </td>
                <td>
                </td>
                <td>
                        <div>Investigation profile to assiciate the notable event with.</div>
                </td>
            </tr>
            <tr>
                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="parameter-"></div>
                    <b>name</b>
                    <a class="ansibleOptionLink" href="#parameter-" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                         / <span style="color: red">required</span>
                    </div>
                </td>
                <td>
                </td>
                <td>
                        <div>Name of notable event</div>
                </td>
            </tr>
            <tr>
                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="parameter-"></div>
                    <b>next_steps</b>
                    <a class="ansibleOptionLink" href="#parameter-" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">list</span>
                    </div>
                </td>
                <td>
                </td>
                <td>
                        <div>List of adaptive responses that should be run next</div>
                        <div>Describe next steps and response actions that an analyst could take to address this threat.</div>
                </td>
            </tr>
            <tr>
                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="parameter-"></div>
                    <b>recommended_actions</b>
                    <a class="ansibleOptionLink" href="#parameter-" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">list</span>
                    </div>
                </td>
                <td>
                </td>
                <td>
                        <div>List of adaptive responses that are recommended to be run next</div>
                        <div>Identifying Recommended Adaptive Responses will highlight those actions for the analyst when looking at the list of response actions available, making it easier to find them among the longer list of available actions.</div>
                </td>
            </tr>
            <tr>
                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="parameter-"></div>
                    <b>security_domain</b>
                    <a class="ansibleOptionLink" href="#parameter-" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                    </div>
                </td>
                <td>
                        <ul style="margin: 0; padding: 0"><b>Choices:</b>
                                    <li>access</li>
                                    <li>endpoint</li>
                                    <li>network</li>
                                    <li><div style="color: blue"><b>threat</b>&nbsp;&larr;</div></li>
                                    <li>identity</li>
                                    <li>audit</li>
                        </ul>
                </td>
                <td>
                        <div>Splunk Security Domain</div>
                </td>
            </tr>
            <tr>
                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="parameter-"></div>
                    <b>severity</b>
                    <a class="ansibleOptionLink" href="#parameter-" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                    </div>
                </td>
                <td>
                        <ul style="margin: 0; padding: 0"><b>Choices:</b>
                                    <li>informational</li>
                                    <li>low</li>
                                    <li>medium</li>
                                    <li><div style="color: blue"><b>high</b>&nbsp;&larr;</div></li>
                                    <li>critical</li>
                                    <li>unknown</li>
                        </ul>
                </td>
                <td>
                        <div>Severity rating</div>
                </td>
            </tr>
            <tr>
                <td colspan="1">
                    <div class="ansibleOptionAnchor" id="parameter-"></div>
                    <b>state</b>
                    <a class="ansibleOptionLink" href="#parameter-" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                         / <span style="color: red">required</span>
                    </div>
                </td>
                <td>
                        <ul style="margin: 0; padding: 0"><b>Choices:</b>
                                    <li>present</li>
                                    <li>absent</li>
                        </ul>
                </td>
                <td>
                        <div>Add or remove a data source.</div>
                </td>
            </tr>
    </table>
    <br/>




Examples
--------

.. code-block:: yaml+jinja

    - name: Example of using splunk.es.adaptive_response_notable_event module
      splunk.es.adaptive_response_notable_event:
        name: "Example notable event from Ansible"
        correlation_search_name: "Example Correlation Search From Ansible"
        description: "Example notable event from Ansible, description."
        state: "present"
        next_steps:
          - ping
          - nslookup
        recommended_actions:
          - script
          - ansiblesecurityautomation




Status
------


Authors
~~~~~~~

- Ansible Security Automation Team (@maxamillion) <https://github.com/ansible-security>
