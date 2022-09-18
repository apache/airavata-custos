#!/usr/bin/python
# -*- coding: utf-8 -*-

# (c) 2020, Simon Dodsley (simon@purestorage.com)
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function

__metaclass__ = type

ANSIBLE_METADATA = {
    "metadata_version": "1.1",
    "status": ["preview"],
    "supported_by": "community",
}

DOCUMENTATION = r"""
---
module: purefb_policy
version_added: '1.0.0'
short_description: Manage FlashBlade policies
description:
- Manage policies for filesystem and file replica links
author:
- Pure Storage Ansible Team (@sdodsley) <pure-ansible-team@purestorage.com>
options:
  state:
    description:
    - Create or delete policy
    default: present
    type: str
    choices: [ absent, present ]
  name:
    description:
    - Name of the policy
    type: str
  enabled:
    description:
    - State of policy
    type: bool
    default: True
  every:
    description:
    - Interval between snapshots in seconds
    - Range available 300 - 31536000 (equates to 5m to 365d)
    type: int
  keep_for:
    description:
    - How long to keep snapshots for
    - Range available 300 - 31536000 (equates to 5m to 365d)
    - Must not be set less than I(every)
    type: int
  at:
    description:
    - Provide a time in 12-hour AM/PM format, eg. 11AM
    type: str
  timezone:
    description:
    - Time Zone used for the I(at) parameter
    - If not provided, the module will attempt to get the current local timezone from the server
    type: str
  filesystem:
    description:
    - List of filesystems to add to a policy on creation
    - To amend policy members use the I(purefb_fs) module
    type: list
    elements: str
  replica_link:
    description:
    - List of filesystem replica links to add to a policy on creation
    - To amend policy members use the I(purefb_fs_replica) module
    type: list
    elements: str
extends_documentation_fragment:
- purestorage.flashblade.purestorage.fb
"""

EXAMPLES = r"""
- name: Create a simple policy with no rules
  purefb_policy:
    name: test_policy
    fb_url: 10.10.10.2
    api_token: T-9f276a18-50ab-446e-8a0c-666a3529a1b6
- name: Create a policy and connect to existing filesystems and filesystem replica links
  purefb_policy:
    name: test_policy_with_members
    filesystem:
    - fs1
    - fs2
    replica_link:
    - rl1
    - rl2
    fb_url: 10.10.10.2
    api_token: T-9f276a18-50ab-446e-8a0c-666a3529a1b6
- name: Create a policy with rules
  purefb_policy:
    name: test_policy2
    at: 11AM
    keep_for: 86400
    every: 86400
    timezone: Asia/Shanghai
    fb_url: 10.10.10.2
    api_token: T-9f276a18-50ab-446e-8a0c-666a3529a1b6
- name: Delete a policy
  purefb_policy:
    name: test_policy
    state: absent
    fb_url: 10.10.10.2
    api_token: T-9f276a18-50ab-446e-8a0c-666a3529a1b6
"""

RETURN = r"""
"""

HAS_PURITYFB = True
try:
    from purity_fb import Policy, PolicyRule, PolicyPatch
except ImportError:
    HAS_PURITYFB = False

HAS_PYTZ = True
try:
    import pytz
except ImportError:
    HAS_PYTX = False

import os
import re
import platform

from ansible.module_utils.common.process import get_bin_path
from ansible.module_utils.facts.utils import get_file_content
from ansible.module_utils.basic import AnsibleModule
from ansible_collections.purestorage.flashblade.plugins.module_utils.purefb import (
    get_blade,
    purefb_argument_spec,
)


MIN_REQUIRED_API_VERSION = "1.9"


def _convert_to_millisecs(hour):
    if hour[-2:] == "AM" and hour[:2] == "12":
        return 0
    elif hour[-2:] == "AM":
        return int(hour[:-2]) * 3600000
    elif hour[-2:] == "PM" and hour[:2] == "12":
        return 43200000
    return (int(hour[:-2]) + 12) * 3600000


def _findstr(text, match):
    for line in text.splitlines():
        if match in line:
            found = line
    return found


def _get_local_tz(module, timezone="UTC"):
    """
    We will attempt to get the local timezone of the server running the module and use that.
    If we can't get the timezone then we will set the default to be UTC

    Linnux has been tested and other opersting systems should be OK.
    Failures cause assumption of UTC

    Windows is not supported and will assume UTC
    """
    if platform.system() == "Linux":
        timedatectl = get_bin_path("timedatectl")
        if timedatectl is not None:
            rcode, stdout, stderr = module.run_command(timedatectl)
            if rcode == 0 and stdout:
                line = _findstr(stdout, "Time zone")
                full_tz = line.split(":", 1)[1].rstrip()
                timezone = full_tz.split()[0]
                return timezone
            else:
                module.warn("Incorrect timedatectl output. Timezone will be set to UTC")
        else:
            if os.path.exists("/etc/timezone"):
                timezone = get_file_content("/etc/timezone")
            else:
                module.warn("Could not find /etc/timezone. Assuming UTC")

    elif platform.system() == "SunOS":
        if os.path.exists("/etc/default/init"):
            for line in get_file_content("/etc/default/init", "").splitlines():
                if line.startswith("TZ="):
                    timezone = line.split("=", 1)[1]
                    return timezone
        else:
            module.warn("Could not find /etc/default/init. Assuming UTC")

    elif re.match("^Darwin", platform.platform()):
        systemsetup = get_bin_path("systemsetup")
        if systemsetup is not None:
            rcode, stdout, stderr = module.execute(systemsetup, "-gettimezone")
            if rcode == 0 and stdout:
                timezone = stdout.split(":", 1)[1].lstrip()
            else:
                module.warn("Could not run systemsetup. Assuming UTC")
        else:
            module.warn("Could not find systemsetup. Assuming UTC")

    elif re.match("^(Free|Net|Open)BSD", platform.platform()):
        if os.path.exists("/etc/timezone"):
            timezone = get_file_content("/etc/timezone")
        else:
            module.warn("Could not find /etc/timezone. Assuming UTC")

    elif platform.system() == "AIX":
        aix_oslevel = int(platform.version() + platform.release())
        if aix_oslevel >= 61:
            if os.path.exists("/etc/environment"):
                for line in get_file_content("/etc/environment", "").splitlines():
                    if line.startswith("TZ="):
                        timezone = line.split("=", 1)[1]
                        return timezone
            else:
                module.warn("Could not find /etc/environment. Assuming UTC")
        else:
            module.warn(
                "Cannot determine timezone when AIX os level < 61. Assuming UTC"
            )

    else:
        module.warn("Could not find /etc/timezone. Assuming UTC")

    return timezone


def delete_policy(module, blade):
    """Delete policy"""
    changed = True
    if not module.check_mode:
        try:
            blade.policies.delete_policies(names=[module.params["name"]])
        except Exception:
            module.fail_json(
                msg="Failed to delete policy {0}.".format(module.params["name"])
            )
    module.exit_json(changed=changed)


def create_policy(module, blade):
    """Create snapshot policy"""
    changed = True
    if not module.check_mode:
        try:
            if module.params["at"] and module.params["every"]:
                if not module.params["every"] % 86400 == 0:
                    module.fail_json(
                        msg="At time can only be set if every value is a multiple of 86400"
                    )
                if not module.params["timezone"]:
                    module.params["timezone"] = _get_local_tz(module)
                    if module.params["timezone"] not in pytz.all_timezones_set:
                        module.fail_json(
                            msg="Timezone {0} is not valid".format(
                                module.params["timezone"]
                            )
                        )
            if not module.params["keep_for"]:
                module.params["keep_for"] = 0
            if not module.params["every"]:
                module.params["every"] = 0
            if module.params["keep_for"] < module.params["every"]:
                module.fail_json(
                    msg="Retention period cannot be less than snapshot interval."
                )
            if module.params["at"] and not module.params["timezone"]:
                module.params["timezone"] = _get_local_tz(module)
                if module.params["timezone"] not in set(pytz.all_timezones_set):
                    module.fail_json(
                        msg="Timezone {0} is not valid".format(
                            module.params["timezone"]
                        )
                    )

            if module.params["keep_for"]:
                if not 300 <= module.params["keep_for"] <= 34560000:
                    module.fail_json(
                        msg="keep_for parameter is out of range (300 to 34560000)"
                    )
                if not 300 <= module.params["every"] <= 34560000:
                    module.fail_json(
                        msg="every parameter is out of range (300 to 34560000)"
                    )
                if module.params["at"]:
                    attr = Policy(
                        enabled=module.params["enabled"],
                        rules=[
                            PolicyRule(
                                keep_for=module.params["keep_for"] * 1000,
                                every=module.params["every"] * 1000,
                                at=_convert_to_millisecs(module.params["at"]),
                                time_zone=module.params["timezone"],
                            )
                        ],
                    )
                else:
                    attr = Policy(
                        enabled=module.params["enabled"],
                        rules=[
                            PolicyRule(
                                keep_for=module.params["keep_for"] * 1000,
                                every=module.params["every"] * 1000,
                            )
                        ],
                    )
            else:
                attr = Policy(enabled=module.params["enabled"])
            blade.policies.create_policies(names=[module.params["name"]], policy=attr)
        except Exception:
            module.fail_json(
                msg="Failed to create policy {0}.".format(module.params["name"])
            )
        if module.params["filesystem"]:
            try:
                blade.file_systems.list_file_systems(names=module.params["filesystem"])
                blade.policies.create_policy_filesystems(
                    policy_names=[module.params["name"]],
                    member_names=module.params["filesystem"],
                )
            except Exception:
                blade.policies.delete_policies(names=[module.params["name"]])
                module.fail_json(
                    msg="Failed to connect filesystems to policy {0}, "
                    "or one of {1} doesn't exist.".format(
                        module.params["name"], module.params["filesystem"]
                    )
                )
        if module.params["replica_link"]:
            for link in module.params["replica_link"]:
                remote_array = (
                    blade.file_system_replica_links.list_file_system_replica_links(
                        local_file_system_names=[link]
                    )
                )
                try:
                    blade.policies.create_policy_file_system_replica_links(
                        policy_names=[module.params["name"]],
                        member_names=[link],
                        remote_names=[remote_array.items[0].remote.name],
                    )
                except Exception:
                    blade.policies.delete_policies(names=[module.params["name"]])
                    module.fail_json(
                        msg="Failed to connect filesystem replicsa link {0} to policy {1}. "
                        "Replica Link {0} does not exist.".format(
                            link, module.params["name"]
                        )
                    )
    module.exit_json(changed=changed)


def update_policy(module, blade, policy):
    """Update snapshot policy"""
    changed = False
    if not policy.rules:
        current_policy = {
            "time_zone": None,
            "every": 0,
            "keep_for": 0,
            "at": 0,
            "enabled": policy.enabled,
        }
    else:
        if policy.rules[0].keep_for != 0:
            policy.rules[0].keep_for = int(policy.rules[0].keep_for / 1000)
        if policy.rules[0].every != 0:
            policy.rules[0].every = int(policy.rules[0].every / 1000)

        current_policy = {
            "time_zone": policy.rules[0].time_zone,
            "every": policy.rules[0].every,
            "keep_for": policy.rules[0].keep_for,
            "at": policy.rules[0].at,
            "enabled": policy.enabled,
        }
    if not module.params["every"]:
        every = 0
    else:
        every = module.params["every"]
    if not module.params["keep_for"]:
        keep_for = 0
    else:
        keep_for = module.params["keep_for"]
    if module.params["at"]:
        at_time = _convert_to_millisecs(module.params["at"])
    else:
        at_time = None
    if not module.params["timezone"]:
        timezone = _get_local_tz(module)
    else:
        timezone = module.params["timezone"]
    if at_time:
        new_policy = {
            "time_zone": timezone,
            "every": every,
            "keep_for": keep_for,
            "at": at_time,
            "enabled": module.params["enabled"],
        }
    else:
        new_policy = {
            "time_zone": None,
            "every": every,
            "keep_for": keep_for,
            "at": None,
            "enabled": module.params["enabled"],
        }
    if (
        new_policy["time_zone"]
        and new_policy["time_zone"] not in pytz.all_timezones_set
    ):
        module.fail_json(
            msg="Timezone {0} is not valid".format(module.params["timezone"])
        )

    if current_policy != new_policy:
        if not module.params["at"]:
            module.params["at"] = current_policy["at"]
        if not module.params["keep_for"]:
            module.params["keep_for"] = current_policy["keep_for"]
        if not module.params["every"]:
            module.params["every"] = current_policy["every"]
        if module.params["at"] and module.params["every"]:
            if not module.params["every"] % 86400 == 0:
                module.fail_json(
                    msg="At time can only be set if every value is a multiple of 86400"
                )
        if module.params["keep_for"] < module.params["every"]:
            module.fail_json(
                msg="Retention period cannot be less than snapshot interval."
            )
        if module.params["at"] and not module.params["timezone"]:
            module.params["timezone"] = _get_local_tz(module)
            if module.params["timezone"] not in set(pytz.all_timezones_set):
                module.fail_json(
                    msg="Timezone {0} is not valid".format(module.params["timezone"])
                )

        changed = True
        if not module.check_mode:
            try:
                attr = PolicyPatch()
                attr.enabled = module.params["enabled"]
                if at_time:
                    attr.add_rules = [
                        PolicyRule(
                            keep_for=module.params["keep_for"] * 1000,
                            every=module.params["every"] * 1000,
                            at=at_time,
                            time_zone=timezone,
                        )
                    ]
                else:
                    attr.add_rules = [
                        PolicyRule(
                            keep_for=module.params["keep_for"] * 1000,
                            every=module.params["every"] * 1000,
                        )
                    ]
                attr.remove_rules = [
                    PolicyRule(
                        keep_for=current_policy["keep_for"] * 1000,
                        every=current_policy["every"] * 1000,
                        at=current_policy["at"],
                        time_zone=current_policy["time_zone"],
                    )
                ]
                blade.policies.update_policies(
                    names=[module.params["name"]], policy_patch=attr
                )
            except Exception:
                module.fail_json(
                    msg="Failed to update policy {0}.".format(module.params["name"])
                )
    module.exit_json(changed=changed)


def main():
    argument_spec = purefb_argument_spec()
    argument_spec.update(
        dict(
            state=dict(type="str", default="present", choices=["absent", "present"]),
            enabled=dict(type="bool", default=True),
            timezone=dict(type="str"),
            name=dict(type="str"),
            at=dict(type="str"),
            every=dict(type="int"),
            keep_for=dict(type="int"),
            filesystem=dict(type="list", elements="str"),
            replica_link=dict(type="list", elements="str"),
        )
    )

    required_together = [["keep_for", "every"]]

    module = AnsibleModule(
        argument_spec, required_together=required_together, supports_check_mode=True
    )

    if not HAS_PURITYFB:
        module.fail_json(msg="purity_fb sdk is required for this module")
    if not HAS_PYTZ:
        module.fail_json(msg="pytz is required for this module")

    state = module.params["state"]
    blade = get_blade(module)
    versions = blade.api_version.list_versions().versions

    if MIN_REQUIRED_API_VERSION not in versions:
        module.fail_json(
            msg="Minimum FlashBlade REST version required: {0}".format(
                MIN_REQUIRED_API_VERSION
            )
        )

    try:
        policy = blade.policies.list_policies(names=[module.params["name"]])
    except Exception:
        policy = None

    if policy and state == "present":
        update_policy(module, blade, policy.items[0])
    elif state == "present" and not policy:
        create_policy(module, blade)
    elif state == "absent" and policy:
        delete_policy(module, blade)

    module.exit_json(changed=False)


if __name__ == "__main__":
    main()
