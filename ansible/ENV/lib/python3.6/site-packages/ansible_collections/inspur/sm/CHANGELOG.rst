=======================
Inspur.sm Release Notes
=======================

.. contents:: Topics


v1.1.3
======

Bugfixes
--------

- Add ansible 2.11 test.
- Add the no_log=true attribute to some modules.

v1.1.2
======

Bugfixes
--------

- Update 'supports_check_mode=False' to 'supports_check_mode=True' for all modules ending in '_info'.

v1.1.1
======

Minor Changes
-------------

- Modified version information to 1.1.1 in galaxy.yml.

Bugfixes
--------

- Update version_added field in ad_group, ldap_group, user, and user_group modules to match the collection version they were first introduced in.

v1.1.0
======

Minor Changes
-------------

- Add CODE_OF_CONDUCT.md file.
- Add a meta/runtime.yml file.
- Add the code of conduct to the README.md file.
- Delete the Collections imported in the adapter_info.py.
- Delete the Collections imported in the module.
- Documentation, examples, and return use FQCNs to M(..).
- Modify ansible_test.yml to add push trigger rule.
- Modify ansibled-test. yml file, add timing execution script, run environment only keep Python 3.8 version.
- Modify inspur_sm_sdk in README.md to inspursmsdk.
- Modify paybooks,Using FQCN.
- Modify the README.md file to add Ansible Code of Conduct (COC).
- Modify the README.md file to add content for releasing, versioning and deprecation(https://github.com/ISIB-Group/inspur.sm/issues/27).
- Modify the README.md file to change the supported Anible version to 2.10.0
- Modify the ansible-test.yml file to Remove the Python Version from the Run sanity tests.
- Modify the ansible-test.yml file to add Ansible and Python versions.
- Modify the description of Ansible in README.md.
- Modify the format of DOCUMENTATION on Required in the module.
- Modify the github repository path referenced in galaxy.yml.
- Modify the module_utils/ism.py file to add check mode processing.
- Modify the state of chenged in the module when the operation changes.
- Modify the value of supports_check_mode in the module to False.
- Regenerate the.rst file.

v1.0.3
======

Release Summary
---------------

Modify the content format of 'readme.md'.

v1.0.2
======

Release Summary
---------------

Modify the generated.RST file style.
