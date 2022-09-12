# -*- coding: utf-8 -*-
# Copyright: (c) 2020, Infinidat <info@infinidat.com>
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from ansible.errors import AnsibleError
import datetime

def delta_time(dt, **kwargs):
    """
    Add to the time.
    Ref: https://docs.python.org/3.6/library/datetime.html#timedelta-objects
    """
    return dt + datetime.timedelta(**kwargs)

class FilterModule(object):
    """
    A filter look up class for custom filter plugins.
    Ref: https://www.dasblinkenlichten.com/creating-ansible-filter-plugins/
    """
    def filters(self):
        """
        Lookup the filter function by name and execute it.
        """
        return self.filter_map

    filter_map = {
        'delta_time': delta_time,
    }
