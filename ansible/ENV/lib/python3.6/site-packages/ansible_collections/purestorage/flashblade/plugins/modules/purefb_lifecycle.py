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
module: purefb_lifecycle
version_added: '1.4.0'
short_description: Manage FlashBlade object lifecycles
description:
- Manage lifecycles for object buckets
author:
- Pure Storage Ansible Team (@sdodsley) <pure-ansible-team@purestorage.com>
options:
  state:
    description:
    - Create or delete lifecycle rule
    default: present
    type: str
    choices: [ absent, present ]
  bucket:
    description:
    - Bucket the lifecycle rule applies to
    type: str
    required: true
  name:
    description:
    - Name of the lifecycle rule
    type: str
    required: true
  enabled:
    description:
    - State of lifecycle rule
    type: bool
    default: True
  keep_for:
    description:
    - Time after which previous versions will be marked expired.
    - Enter as days (d) or weeks (w). Range is 1 - 2147483647 days.
    type: str
  prefix:
    description:
    - Object key prefix identifying one or more objects in the bucket
    type: str
extends_documentation_fragment:
- purestorage.flashblade.purestorage.fb
"""

EXAMPLES = r"""
- name: Create a lifecycle rule called bar for bucket foo
  purefb_lifecycle:
    name: bar
    bucket: foo
    keep_for: 2d
    prefix: test
    fb_url: 10.10.10.2
    api_token: T-9f276a18-50ab-446e-8a0c-666a3529a1b6
- name: Delete lifecycle rule foo from bucket foo
  purefb_lifecycle:
    name: foo
    bucket: bar
    state: absent
    fb_url: 10.10.10.2
    api_token: T-9f276a18-50ab-446e-8a0c-666a3529a1b6
"""

RETURN = r"""
"""

HAS_PURITYFB = True
try:
    from purity_fb import LifecycleRulePost, LifecycleRulePatch, Reference
except ImportError:
    HAS_PURITYFB = False


from ansible.module_utils.basic import AnsibleModule
from ansible_collections.purestorage.flashblade.plugins.module_utils.purefb import (
    get_blade,
    purefb_argument_spec,
)


MIN_REQUIRED_API_VERSION = "1.10"


def _get_bucket(module, blade):
    s3bucket = None
    buckets = blade.buckets.list_buckets()
    for bucket in range(0, len(buckets.items)):
        if buckets.items[bucket].name == module.params["bucket"]:
            s3bucket = buckets.items[bucket]
    return s3bucket


def _convert_to_millisecs(day):
    if day[-1:].lower() == "w":
        return int(day[:-1]) * 7 * 86400000
    elif day[-1:].lower() == "d":
        return int(day[:-1]) * 86400000
    return 0


def _findstr(text, match):
    for line in text.splitlines():
        if match in line:
            found = line
    return found


def delete_rule(module, blade):
    """Delete lifecycle rule"""
    changed = True
    if not module.check_mode:
        try:
            blade.lifecycle_rules.delete_lifecycle_rules(
                names=[module.params["bucket"] + "/" + module.params["name"]]
            )
        except Exception:
            module.fail_json(
                msg="Failed to delete lifecycle rule {0} for bucket {1}.".format(
                    module.params["name"], module.params["bucket"]
                )
            )
    module.exit_json(changed=changed)


def create_rule(module, blade):
    """Create lifecycle policy"""
    changed = True
    if not module.check_mode:
        if not module.params["keep_for"]:
            module.fail_json(
                msg="'keep_for' is required to create a new lifecycle rule"
            )
        if not module.params["keep_for"][-1:].lower() in ["w", "d"]:
            module.fail_json(msg="'keep_for' format incorrect - specify as 'd' or 'w'")
        try:
            attr = LifecycleRulePost(
                bucket=Reference(name=module.params["bucket"]),
                rule_id=module.params["name"],
                keep_previous_version_for=_convert_to_millisecs(
                    module.params["keep_for"]
                ),
                prefix=module.params["prefix"],
            )
            blade.lifecycle_rules.create_lifecycle_rules(rule=attr)
            if not module.params["enabled"]:
                attr = LifecycleRulePatch()
                attr.enabled = False
                blade.lifecycle_rules.update_lifecycle_rules(
                    name=[module.params["bucket"] + "/" + module.params["name"]],
                    rule=attr,
                )
        except Exception:
            module.fail_json(
                msg="Failed to create lifecycle rule {0} for bucket {1}.".format(
                    module.params["name"], module.params["bucket"]
                )
            )
    module.exit_json(changed=changed)


def update_rule(module, blade, rule):
    """Update snapshot policy"""
    changed = False
    current_rule = {
        "prefix": rule.prefix,
        "keep_previous_version_for": rule.keep_previous_version_for,
        "enabled": rule.enabled,
    }
    if not module.params["prefix"]:
        prefix = current_rule["prefix"]
    else:
        prefix = module.params["prefix"]
    if not module.params["keep_for"]:
        keep_for = current_rule["keep_previous_version_for"]
    else:
        keep_for = _convert_to_millisecs(module.params["keep_for"])
    new_rule = {
        "prefix": prefix,
        "keep_previous_version_for": keep_for,
        "enabled": module.params["enabled"],
    }

    if current_rule != new_rule:
        changed = True
        if not module.check_mode:
            try:
                attr = LifecycleRulePatch(
                    keep_previous_version_for=new_rule["keep_previous_version_for"],
                    prefix=new_rule["prefix"],
                )
                attr.enabled = module.params["enabled"]
                blade.lifecycle_rules.update_lifecycle_rules(
                    names=[module.params["bucket"] + "/" + module.params["name"]],
                    rule=attr,
                )
                changed = True
            except Exception:
                module.fail_json(
                    msg="Failed to update lifecycle rule {0} for bucket {1}.".format(
                        module.params["name"], module.params["bucket"]
                    )
                )
    module.exit_json(changed=changed)


def main():
    argument_spec = purefb_argument_spec()
    argument_spec.update(
        dict(
            state=dict(type="str", default="present", choices=["absent", "present"]),
            enabled=dict(type="bool", default=True),
            bucket=dict(type="str", required=True),
            name=dict(type="str", required=True),
            prefix=dict(
                type="str",
            ),
            keep_for=dict(type="str"),
        )
    )

    module = AnsibleModule(argument_spec, supports_check_mode=True)

    if not HAS_PURITYFB:
        module.fail_json(msg="purity_fb sdk is required for this module")

    state = module.params["state"]
    blade = get_blade(module)
    versions = blade.api_version.list_versions().versions

    if MIN_REQUIRED_API_VERSION not in versions:
        module.fail_json(
            msg="Minimum FlashBlade REST version required: {0}".format(
                MIN_REQUIRED_API_VERSION
            )
        )

    if not _get_bucket(module, blade):
        module.fail_json(
            msg="Specified bucket {0} does not exist".format(module.params["bucket"])
        )

    try:
        rule = blade.lifecycle_rules.list_lifecycle_rules(
            names=[module.params["bucket"] + "/" + module.params["name"]]
        )
    except Exception:
        rule = None

    if rule and state == "present":
        update_rule(module, blade, rule.items[0])
    elif state == "present" and not rule:
        create_rule(module, blade)
    elif state == "absent" and rule:
        delete_rule(module, blade)

    module.exit_json(changed=False)


if __name__ == "__main__":
    main()
