==============================
Vultr Collection Release Notes
==============================

.. contents:: Topics


v1.1.0
======

Minor Changes
-------------

- vultr_block_storage - Included ability to resize, attach and detach Block Storage Volumes.

v1.0.0
======

v0.3.0
======

Minor Changes
-------------

- vultr_server_info, vultr_server - Improved handling of discontinued plans (https://github.com/ansible/ansible/issues/66707).

Bugfixes
--------

- vultr - Fixed the issue retry max delay param was ignored.

New Modules
-----------

- vultr_plan_baremetal_info - Gather information about the Vultr Bare Metal plans available.
- vultr_server_baremetal - Manages baremetal servers on Vultr.
