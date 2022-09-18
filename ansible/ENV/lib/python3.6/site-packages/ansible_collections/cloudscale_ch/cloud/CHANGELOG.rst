==============================================
Ansible Collection cloudscale.ch Release Notes
==============================================

.. contents:: Topics


v2.1.0
======

Minor Changes
-------------

- Add interface parameter to server module (https://github.com/cloudscale-ch/ansible-collection-cloudscale/pull/54).
- Rename server_uuids parameter to servers in volume module (https://github.com/cloudscale-ch/ansible-collection-cloudscale/pull/54).

Deprecated Features
-------------------

- The aliases ``server_uuids`` and ``server_uuid`` of the servers parameter in the volume module will be removed in version 3.0.0.

v2.0.0
======

Breaking Changes / Porting Guide
--------------------------------

- floating_ip - ``name`` is required for assigning a new floating IP.

v1.3.1
======

Minor Changes
-------------

- Implemented identical naming support of the same resource type per zone (https://github.com/cloudscale-ch/ansible-collection-cloudscale/pull/46).

Bugfixes
--------

- Fix inventory plugin failing to launch (https://github.com/cloudscale-ch/ansible-collection-cloudscale/issues/49).

v1.3.0
======

Minor Changes
-------------

- floating_ip - Added an optional name parameter to gain idempotency. The parameter will be required for assigning a new floating IP with release of version 2.0.0 (https://github.com/cloudscale-ch/ansible-collection-cloudscale/pull/43/).
- floating_ip - Allow to reserve an IP without assignment to a server (https://github.com/cloudscale-ch/ansible-collection-cloudscale/pull/31/).

New Modules
-----------

- subnet - Manages subnets on the cloudscale.ch IaaS service

v1.2.0
======

Minor Changes
-------------

- server_group - The module has been refactored and the code simplifed (https://github.com/cloudscale-ch/ansible-collection-cloudscale/pull/23).
- volume - The module has been refactored and the code simplifed (https://github.com/cloudscale-ch/ansible-collection-cloudscale/pull/24).

New Modules
-----------

- network - Manages networks on the cloudscale.ch IaaS service

v1.1.0
======

Minor Changes
-------------

- floating_ip - added tags support (https://github.com/cloudscale-ch/ansible-collection-cloudscale/pull/16)

New Modules
-----------

- objects_user - Manages objects users on the cloudscale.ch IaaS service
