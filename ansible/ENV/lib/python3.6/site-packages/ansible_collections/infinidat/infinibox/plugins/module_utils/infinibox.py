# -*- coding: utf-8 -*-
# This code is part of Ansible, but is an independent component.
# This particular file snippet, and this file snippet only, is BSD licensed.
# Modules you write using this snippet, which is embedded dynamically by Ansible
# still belong to the author of the module, and may assign their own license
# to the complete work.
#
# Copyright: (c) 2020, Infinidat <info@infinidat.com>
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



HAS_INFINISDK = True
try:
    from infinisdk import InfiniBox, core
except ImportError:
    HAS_INFINISDK = False

from functools import wraps
from os import environ
from os import path
from datetime import datetime
from infinisdk.core.exceptions import ObjectNotFound


def unixMillisecondsToDate(unix_ms):
    return (datetime.utcfromtimestamp(unix_ms/1000.), 'UTC')


def api_wrapper(func):
    """ Catch API Errors Decorator"""
    @wraps(func)
    def __wrapper(*args, **kwargs):
        module = args[0]
        try:
            return func(*args, **kwargs)
        except core.exceptions.APICommandException as e:
            module.fail_json(msg=e.message)
        except core.exceptions.SystemNotFoundException as e:
            module.fail_json(msg=e.message)
        except Exception:
            raise
    return __wrapper


@api_wrapper
def get_system(module):
    """Return System Object or Fail"""
    box = module.params['system']
    user = module.params.get('user', None)
    password = module.params.get('password', None)

    if user and password:
        system = InfiniBox(box, auth=(user, password), use_ssl=True)
    elif environ.get('INFINIBOX_USER') and environ.get('INFINIBOX_PASSWORD'):
        system = InfiniBox(box, \
                           auth=(environ.get('INFINIBOX_USER'), \
                                 environ.get('INFINIBOX_PASSWORD')), \
                           use_ssl=True)
    elif path.isfile(path.expanduser('~') + '/.infinidat/infinisdk.ini'):
        system = InfiniBox(box, use_ssl=True)
    else:
        module.fail_json(msg="You must set INFINIBOX_USER and INFINIBOX_PASSWORD environment variables or set username/password module arguments")

    try:
        system.login()
    except Exception:
        module.fail_json(msg="Infinibox authentication failed. Check your credentials")
    return system


def infinibox_argument_spec():
    """Return standard base dictionary used for the argument_spec argument in AnsibleModule"""
    return dict(
        system=dict(required=True),
        user=dict(),
        password=dict(no_log=True),
    )


def infinibox_required_together():
    """Return the default list used for the required_together argument to AnsibleModule"""
    return [['user', 'password']]


def merge_two_dicts(dict1, dict2):
    """
    Merge two dicts into one and return.
    result = {**dict1, **dict2} only works in py3.5+.
    """
    result = dict1.copy()
    result.update(dict2)
    return result


@api_wrapper
def get_pool(module, system):
    """
    Return Pool. Try key look up using 'pool', or if that fails, 'name'.
    If the pool is not found, return None.
    """
    try:
        try:
            name = module.params['pool']
        except KeyError:
            name = module.params['name']
        return system.pools.get(name=name)
    except Exception:
        return None


@api_wrapper
def get_filesystem(module, system):
    """Return Filesystem or None"""
    try:
        try:
            filesystem = system.filesystems.get(name=module.params['filesystem'])
        except KeyError:
            filesystem = system.filesystems.get(name=module.params['name'])
        return filesystem
    except Exception:
        return None


@api_wrapper
def get_export(module, system):
    """Return export if found or None if not found"""
    try:
        try:
            export_name = module.params['export']
        except KeyError:
            export_name = module.params['name']

        export = system.exports.get(export_path=export_name)
    except ObjectNotFound as err:
        return None

    return export


@api_wrapper
def get_volume(module, system):
    """Return Volume or None"""
    try:
        try:
            volume = system.volumes.get(name=module.params['name'])
        except KeyError:
            volume = system.volumes.get(name=module.params['volume'])
        return volume
    except Exception:
        return None


@api_wrapper
def get_host(module, system):
    """Find a host by the name specified in the module"""
    host = None

    for a_host in system.hosts.to_list():
        a_host_name = a_host.get_name()
        try:
            host_param = module.params['name']
        except KeyError:
            host_param = module.params['host']

        if a_host_name == host_param:
            host = a_host
            break
    return host


@api_wrapper
def get_cluster(module, system):
    """Find a cluster by the name specified in the module"""
    cluster = None
    #print("dir:", dir(system))

    for a_cluster in system.host_clusters.to_list():
        a_cluster_name = a_cluster.get_name()
        cluster_param = module.params['name']

        if a_cluster_name == cluster_param:
            cluster = a_cluster
            break
    return cluster


@api_wrapper
def get_user(module, system):
    """Find a user by the user_name specified in the module"""
    user = None
    user_name = module.params['user_name']
    try:
        user = system.users.get(name=user_name)
    except ObjectNotFound:
        pass
    return user
