.. _splunk.es.correlation_search_info_module:


*********************************
splunk.es.correlation_search_info
*********************************

**Manage Splunk Enterprise Security Correlation Searches**


Version added: 1.0.0

.. contents::
   :local:
   :depth: 1


Synopsis
--------
- This module allows for the query of Splunk Enterprise Security Correlation Searches




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
                    <b>name</b>
                    <a class="ansibleOptionLink" href="#parameter-" title="Permalink to this option"></a>
                    <div style="font-size: small">
                        <span style="color: purple">string</span>
                    </div>
                </td>
                <td>
                </td>
                <td>
                        <div>Name of coorelation search</div>
                </td>
            </tr>
    </table>
    <br/>




Examples
--------

.. code-block:: yaml+jinja

    - name: Example usage of splunk.es.correlation_search_info
      splunk.es.correlation_search_info:
        name: "Name of correlation search"
      register: scorrelation_search_info

    - name: debug display information gathered
      debug:
        var: scorrelation_search_info




Status
------


Authors
~~~~~~~

- Ansible Security Automation Team (@maxamillion) <https://github.com/ansible-security>
