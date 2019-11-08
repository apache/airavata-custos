#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

class Groups(object):

    def createGroup(self, userGroup):
        pass

    def updateGroup(self, userGroup):
        pass

    def check_if_exists(self, tenant_id, group_id):
        pass

    def delete_group(self, tenant_id, group_id):
        pass

    def get_group(self, tenant_id, group_id):
        pass

    def get_groups(self, tenant_id, offset=0, limit=-1):
        pass

    def add_users_to_group(self, tenant_id, group_id, user_ids):
        pass

    def remove_users_from_group(self, tenant_id, group_id, user_ids):
        pass

    def transfer_group_ownership(self, tenant_id, group_id, owner_id):
        pass

    def add_group_admins(self, tenant_id, group_id, admin_ids):
        pass

    def remove_group_admins(self, tenant_id, group_id, admin_ids):
        pass

    def has_admin_access(self, tenant_id, group_id, admin_id):
        pass

    def has_owner_access(self, tenant_id, group_id, owner_id):
        pass

    def get_group_members(self, tenant_id, group_id):
        pass

    def get_all_member_groups_for_user(self, tenant_id, user_id):
        pass