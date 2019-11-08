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

class Access(object):

    def share_entity_with_users(self, tenant_id, entity_id, user_ids, permission):
        pass

    def share_entity_with_groups(self, tenant_id, entity_id, group_ids, permission):
        pass

    def revoke_entity_sharing_with_users(self, tenant_id, entity_id, user_ids, permission):
        pass

    def revoke_entity_sharing_with_groups(self, tenant_id, entity_id, group_ids, permission):
        pass

    def check_user_access(self, tenant_id, entity_id, user_id,permission):
        pass

    def check_group_access(self, tenant_id, entity_id, group_id, permission):
        pass