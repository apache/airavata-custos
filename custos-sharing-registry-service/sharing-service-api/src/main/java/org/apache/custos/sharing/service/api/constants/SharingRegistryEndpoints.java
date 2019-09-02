package org.apache.custos.sharing.service.api.constants;

import org.springframework.stereotype.Component;

@Component
public class SharingRegistryEndpoints {

    public String BASE_URL = "http://localhost:7070";

    //Group APIs getters
    public String getCreateGroupApi(String domainId) {
        return String.format("/domains/%s/groups", domainId);
    }

    public String getUpdateGroupApi(String domainId) {
        return String.format("/domains/%s/groups", domainId);
    }

    public String getIsGroupExistsApi(String domainId, String groupId) {
        return String.format("/domains/%s/groups/%s", domainId, groupId);
    }

    public String getDeleteGroupApi(String domainId, String groupId) {
        return String.format("/domains/%s/groups/%s", domainId, groupId);
    }

    public String getGetGroupApi(String domainId, String groupId) {
        return String.format("/domains/%s/groups/%s", domainId, groupId);
    }

    public String getGetGroupsApi(String domainId, int offset, int limit) {
        return String.format("/domains/%s/groups?offset=%d&limit=%d", domainId,offset, limit);
    }

    public String getAddUsersToGroupApi(String domainId, String groupId) {
        return String.format("/domains/%s/groups/%s/users", domainId, groupId);
    }

    public String getRemoveUsersFromGroupApi(String domainId, String groupId) {
        return String.format("/domains/%s/groups/%s/users", domainId, groupId);
    }

    public String getTransferGroupOwnershipApi(String domainId, String groupId, String ownerId) {
        return String.format("/domains/%s/groups/%s/owners/%s", domainId, groupId, ownerId);
    }

    public String getAddGroupAdminsApi(String domainId, String groupId) {
        return String.format("/domains/%s/groups/%s/admins", domainId, groupId);
    }

    public String getRemoveGroupAdminsApi(String domainId, String groupId) {
        return String.format("/domains/%s/groups/%s/admins", domainId, groupId);
    }

    public String getHasAdminAccessApi(String domainId, String groupId, String adminId) {
        return String.format("/domains/%s/groups/%s/admins/%s", domainId, groupId, adminId);
    }

    public String getHasOwnerAccessApi(String domainId, String groupId, String ownerId) {
        return String.format("/domains/%s/groups/%s/owners/%s", domainId, groupId, ownerId);
    }

    public String getGroupMemebersOfTypeUserApi(String domainId, String groupId, int offset, int limit) {
        return String.format("/domains/%s/groups/%s/users?offset=%d&limit=%d", domainId, groupId, offset, limit);
    }

    public String getGroupMembersOfTypeGroupApi(String domainId, String groupId, int offset, int limit) {
        return String.format("/domains/%s/groups/%s/subgroups?offset=%d&limit=%d", domainId, groupId, offset, limit);
    }

    public String getAddGroupsApi(String domainId, String groupId) {
        return String.format("/domains/%s/groups/%s/subgroups", domainId, groupId);
    }

    public String getRemoveGroupsApi(String domainId, String groupId) {
        return String.format("/domains/%s/groups/%s/subgroups", domainId, groupId);
    }

    public String getGetGroupsForUserApi(String domainId, String userId) {
        return String.format("/domains/%s/groups/users/%s", domainId, userId);
    }

    public String getBaseUrl() {
        return BASE_URL;
    }


}
