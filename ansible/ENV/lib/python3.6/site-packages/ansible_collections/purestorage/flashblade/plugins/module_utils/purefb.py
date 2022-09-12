# -*- coding: utf-8 -*-

# This code is part of Ansible, but is an independent component.
# This particular file snippet, and this file snippet only, is BSD licensed.
# Modules you write using this snippet, which is embedded dynamically by Ansible
# still belong to the author of the module, and may assign their own license
# to the complete work.
#
# Copyright (c), Simon Dodsley <simon@purestorage.com>,2017
# All rights reserved.
#
# Redistribution and use in source and binary forms, with or without modification,
# are permitted provided that the following conditions are met:
#
#    * Redistributions of source code must retain the above copyright
#      notice, this list of conditions and the following disclaimer.
#    * Redistributions in binary form must reproduce the above copyright notice,
#      this list of conditions and the following disclaimer in the documentation
#      and/or other materials provided with the distribution.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
# ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
# WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
# IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT,
# INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
# PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
# INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
# LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE
# USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

from __future__ import absolute_import, division, print_function

__metaclass__ = type

HAS_PURITY_FB = True
try:
    from purity_fb import PurityFb
except ImportError:
    HAS_PURITY_FB = False

HAS_PYPURECLIENT = True
try:
    from pypureclient import flashblade
except ImportError:
    PYPURECLIENT = False

from os import environ
import platform

VERSION = "1.4"
USER_AGENT_BASE = "Ansible"
API_AGENT_VERSION = "1.5"


def get_blade(module):
    """Return System Object or Fail"""
    user_agent = "%(base)s %(class)s/%(version)s (%(platform)s)" % {
        "base": USER_AGENT_BASE,
        "class": __name__,
        "version": VERSION,
        "platform": platform.platform(),
    }
    blade_name = module.params["fb_url"]
    api = module.params["api_token"]

    if HAS_PURITY_FB:
        if blade_name and api:
            blade = PurityFb(blade_name)
            blade.disable_verify_ssl()
            try:
                blade.login(api)
                versions = blade.api_version.list_versions().versions
                if API_AGENT_VERSION in versions:
                    blade._api_client.user_agent = user_agent
            except Exception:
                module.fail_json(
                    msg="Pure Storage FlashBlade authentication failed. Check your credentials"
                )
        elif environ.get("PUREFB_URL") and environ.get("PUREFB_API"):
            blade = PurityFb(environ.get("PUREFB_URL"))
            blade.disable_verify_ssl()
            try:
                blade.login(environ.get("PUREFB_API"))
                versions = blade.api_version.list_versions().versions
                if API_AGENT_VERSION in versions:
                    blade._api_client.user_agent = user_agent
            except Exception:
                module.fail_json(
                    msg="Pure Storage FlashBlade authentication failed. Check your credentials"
                )
        else:
            module.fail_json(
                msg="You must set PUREFB_URL and PUREFB_API environment variables "
                "or the fb_url and api_token module arguments"
            )
    else:
        module.fail_json(msg="purity_fb SDK not installed.")
    return blade


def get_system(module):
    """Return System Object or Fail"""
    user_agent = "%(base)s %(class)s/%(version)s (%(platform)s)" % {
        "base": USER_AGENT_BASE,
        "class": __name__,
        "version": VERSION,
        "platform": platform.platform(),
    }
    blade_name = module.params["fb_url"]
    api = module.params["api_token"]

    if HAS_PYPURECLIENT:
        if blade_name and api:
            # TODO:(SD) when the page has been added to Purity//FB
            #            versions = requests.get(
            #                "https://" + blade_name + "/api/api_version", verify=False
            #            )
            #            api_version = versions.json()["version"][-1]
            system = flashblade.Client(
                target=blade_name,
                api_token=api,
                user_agent=user_agent,
                #                version=api_version,
            )
        elif environ.get("PUREFB_URL") and environ.get("PUREFB_API"):
            # TODO:(SD) when the page has been added to Purity//FB
            #            versions = requests.get(
            #                "https://" + environ.get("PUREFB_URL") + "/api/api_version", verify=False
            #            )
            #            api_version = versions.json()["version"][-1]
            system = flashblade.Client(
                target=(environ.get("PUREFB_URL")),
                api_token=(environ.get("PUREFB_API")),
                user_agent=user_agent,
                #                version=api_version,
            )
        else:
            module.fail_json(
                msg="You must set PUREFB_URL and PUREFB_API environment variables "
                "or the fb_url and api_token module arguments"
            )
        try:
            system.get_hardware()
        except Exception:
            module.fail_json(
                msg="Pure Storage FlashBlade authentication failed. Check your credentials"
            )
    else:
        module.fail_json(msg="pypureclient SDK not installed.")
    return system


def purefb_argument_spec():
    """Return standard base dictionary used for the argument_spec argument in AnsibleModule"""

    return dict(
        fb_url=dict(),
        api_token=dict(no_log=True),
    )
