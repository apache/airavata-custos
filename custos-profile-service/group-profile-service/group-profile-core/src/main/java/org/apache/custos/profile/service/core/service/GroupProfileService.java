package org.apache.custos.profile.service.core.service;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.google.gson.Gson;
import org.apache.custos.commons.exceptions.ServiceConnectionException;
import org.apache.custos.commons.utils.ServiceRequestClient;
import org.apache.custos.profile.commons.repositories.UserProfileRepository;
import org.apache.custos.sharing.service.api.constants.SharingRegistryEndpoints;
import org.apache.custos.sharing.service.core.exceptions.SharingRegistryException;
import org.apache.custos.sharing.service.core.models.GroupCardinality;
import org.apache.custos.sharing.service.core.models.GroupType;
import org.apache.custos.sharing.service.core.models.UserGroup;
import org.apache.custos.profile.service.core.models.*;
import org.apache.custos.profile.service.core.exceptions.*;
import org.apache.http.HttpResponse;
import org.apache.http.HttpStatus;
import org.json.HTTP;
import org.json.JSONObject;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

public class GroupProfileService {

    private static final Logger logger = LoggerFactory.getLogger(GroupProfileService.class);

    @Autowired
    ServiceRequestClient serviceRequestClient;

    @Autowired
    SharingRegistryEndpoints sharingRegistryEndpoints;

    private static final String GROUP_MANAGER_CPI_VERSION = "1.0";
    private UserProfileRepository userProfileRepository = new UserProfileRepository();
    private String SHARING_SERVICE_BASE_URL = sharingRegistryEndpoints.getBaseUrl();

    public String getAPIVersion() {
        return GROUP_MANAGER_CPI_VERSION;
    }

    private JSONObject getResponseBody(HttpResponse response) throws Exception {

        try {
            if (response.getStatusLine().getStatusCode() == HttpStatus.SC_OK) {
                BufferedReader rd = new BufferedReader(new InputStreamReader(response.getEntity().getContent()));
                StringBuffer result = new StringBuffer();
                String line = "";
                while ((line = rd.readLine()) != null) {
                    result.append(line);
                }
                JSONObject jsonObject = new JSONObject(result.toString());
                return jsonObject;
            } else {
                throw new Exception();
            }
        } catch (Exception ex) {
            //TODO: sharing service did not return correct response
            throw ex;
        }
    }

    private boolean validateHeadHttpRequest(HttpResponse response) throws Exception{
        if(response.getStatusLine().getStatusCode() == HttpStatus.SC_OK) {
            return true;
        }else if(response.getStatusLine().getStatusCode() == HttpStatus.SC_NOT_FOUND) {
            return false;
        }
        else {
            //TODO: this will be sharing server error
            throw new Exception();
        }
    }

    private boolean validatePutHttpRequest(HttpResponse response) throws Exception{
        if(response.getStatusLine().getStatusCode() == HttpStatus.SC_OK) {
            return true;
        } else {
            //TODO: this will be sharing server error
            throw new Exception();
        }
    }

    public String createGroup(GroupModel groupModel, String gatewayId, String userId) throws GroupManagerServiceException{
        try {

            UserGroup sharingUserGroup = new UserGroup();
            sharingUserGroup.setGroupId(groupModel.getId());
            sharingUserGroup.setName(groupModel.getName());
            sharingUserGroup.setDescription(groupModel.getDescription());
            sharingUserGroup.setGroupType(GroupType.USER_LEVEL_GROUP);
            sharingUserGroup.setGroupCardinality(GroupCardinality.MULTI_USER);
            sharingUserGroup.setDomainId(groupModel.getGatewayId());
            sharingUserGroup.setOwnerId(groupModel.getOwnerId());

            HttpResponse response = serviceRequestClient.httpPost(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getCreateGroupApi(groupModel.getGatewayId()), sharingUserGroup);

            //TODO: should happen in one rest call
            //internalAddUsersToGroup(sharingClient, gatewayId, groupModel.getMembers(), groupId);
            //addGroupAdmins(authzToken,groupId,groupModel.getAdmins());
            JSONObject responseEntity = getResponseBody(response);
            return responseEntity.get("groupId").toString();
        }catch (ServiceConnectionException ex) {
            //TODO: reason could not make connection request to sharing service
            throw ex;
        }
        catch (Exception e) {
            String msg = "Error Creating Group" ;
            logger.error(msg, e);
            GroupManagerServiceException exception = new GroupManagerServiceException(msg + " More info : " + e.getMessage());
            throw exception;
        }
    }

    public boolean updateGroup(GroupModel groupModel, String gatewayId, String userId) throws GroupManagerServiceException{
        try {
            boolean ownerAccess = validateHeadHttpRequest(serviceRequestClient.httpHead(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getHasOwnerAccessApi(gatewayId, groupModel.getId(), userId)));
            boolean adminAccess = validateHeadHttpRequest(serviceRequestClient.httpHead(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getHasAdminAccessApi(gatewayId, groupModel.getId(), userId)));
            if (!ownerAccess || ! adminAccess) {
                throw new GroupManagerServiceException("User does not have permission to update group");
            }

            UserGroup sharingUserGroup = new UserGroup();
            sharingUserGroup.setGroupId(groupModel.getId());
            sharingUserGroup.setName(groupModel.getName());
            sharingUserGroup.setDescription(groupModel.getDescription());
            sharingUserGroup.setGroupType(GroupType.USER_LEVEL_GROUP);
            sharingUserGroup.setDomainId(gatewayId);

            //adding and removal of users should be handle separately??
            HttpResponse response = serviceRequestClient.httpPut(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getUpdateGroupApi(groupModel.getGatewayId()), null);
            return validatePutHttpRequest(response);
        }
        catch (Exception e) {
            String msg = "Error Updating Group" ;
            logger.error(msg, e);
            GroupManagerServiceException exception = new GroupManagerServiceException(msg + " More info : " + e.getMessage());
            throw exception;
        }
    }

    public boolean deleteGroup(String groupId, String groupGatewayId, String domainId, String ownerId) throws GroupManagerServiceException{
        try {
            boolean hasOwnerAccess = validateHeadHttpRequest(serviceRequestClient.httpHead(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getHasOwnerAccessApi(domainId, groupId, ownerId)));
            if (!hasOwnerAccess) {
                throw new GroupManagerServiceException("User does not have permission to delete group");
            }


            serviceRequestClient.httpDelete(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getDeleteGroupApi(groupGatewayId,groupId));
            return true;
        }
        catch (Exception e) {
            String msg = "Error Deleting Group. Group ID: " + groupId ;
            logger.error(msg, e);
            GroupManagerServiceException exception = new GroupManagerServiceException(msg + " More info : " + e.getMessage());
            throw exception;
        }
    }

    public GroupModel getGroup(String groupId, String groupGatewayId) throws GroupManagerServiceException{
        try {
            JSONObject jsonObject = getResponseBody(serviceRequestClient.httpGet(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getGetGroupApi(groupGatewayId, groupId)));
            UserGroup userGroup = new Gson().fromJson(jsonObject.toString(),UserGroup.class);
            GroupModel groupModel = convertToGroupModel(userGroup);
            return groupModel;
        }
        catch (Exception e) {
            String msg = "Error Retreiving Group. Group ID: " + groupId ;
            logger.error(msg, e);
            GroupManagerServiceException exception = new GroupManagerServiceException(msg + " More info : " + e.getMessage());
            throw exception;
        }
    }

    public List<GroupModel> getGroups(String groupGatewayId) throws GroupManagerServiceException{
        try {
            JSONObject jsonObject = getResponseBody(serviceRequestClient.httpGet(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getGetGroupsApi(groupGatewayId, 0, -1)));
            //List<UserGroup> userGroups = new Gson().fromJson(jsonObject.toString(),UserGroup.class);
            List<UserGroup> userGroups = new ArrayList<>();
            return convertToGroupModels(userGroups);
        }
        catch (Exception e) {
            String msg = "Error Retrieving Groups. Domain ID: " + groupGatewayId;
            logger.error(msg, e);
            GroupManagerServiceException exception = new GroupManagerServiceException(msg + " More info : " + e.getMessage());
            throw exception;
        }
    }

    public List<GroupModel> getAllGroupsUserBelongs(String userId, String groupGatewayId) throws GroupManagerServiceException{
        try {
            List<GroupModel> groupModels = new ArrayList<GroupModel>();
            JSONObject jsonObject= getResponseBody(serviceRequestClient.httpGet(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getGetGroupsForUserApi(groupGatewayId, userId)));
            List<UserGroup> userGroups = new ArrayList<>();
            return convertToGroupModels(userGroups);
        }
        catch (Exception e) {
            String msg = "Error Retreiving All Groups for User. User ID: " + userId ;
            logger.error(msg, e);
            GroupManagerServiceException exception = new GroupManagerServiceException(msg + " More info : " + e.getMessage());
            throw exception;
        }
    }

    public boolean addUsersToGroup(List<String> userIds, String groupId, String groupGatewayId, String userId, String domainId) throws GroupManagerServiceException{
        try {
            boolean ownerAccess = validateHeadHttpRequest(serviceRequestClient.httpHead(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getHasOwnerAccessApi(domainId, groupId, userId)));
            boolean adminAccess = validateHeadHttpRequest(serviceRequestClient.httpHead(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getHasAdminAccessApi(domainId, groupId, userId)));
            if (! ownerAccess || !adminAccess) {
                throw new GroupManagerServiceException("User does not have access to add users to the group");
            }
            return internalAddUsersToGroup(domainId, userIds, groupId);

        } catch (Exception e) {
            String msg = "Error adding users to group. Group ID: " + groupId ;
            logger.error(msg, e);
            GroupManagerServiceException exception = new GroupManagerServiceException(msg + " More info : " + e.getMessage());
            throw exception;
        }
    }

    public boolean removeUsersFromGroup(List<String> userIds, String groupId, String groupGatewayId, String userId, String domainId) throws GroupManagerServiceException {
        try {
            boolean ownerAccess = validateHeadHttpRequest(serviceRequestClient.httpHead(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getHasOwnerAccessApi(domainId, groupId, userId)));
            boolean adminAccess = validateHeadHttpRequest(serviceRequestClient.httpHead(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getHasAdminAccessApi(domainId, groupId, userId)));
            if (! ownerAccess || !adminAccess)  {
                throw new GroupManagerServiceException("User does not have access to remove users to the group");
            }
            return validateHeadHttpRequest(serviceRequestClient.httpDelete(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getRemoveUsersFromGroupApi(groupGatewayId, groupId)));
        } catch (Exception e) {
            String msg = "Error remove users to group. Group ID: " + groupId ;
            logger.error(msg, e);
            GroupManagerServiceException exception = new GroupManagerServiceException(msg + " More info : " + e.getMessage());
            throw exception;
        }
    }

    public boolean transferGroupOwnership(String groupId, String groupGatewayId, String newOwnerId, String userId, String domainId) throws GroupManagerServiceException {
        try{
            boolean ownerAccess = validateHeadHttpRequest(serviceRequestClient.httpHead(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getHasOwnerAccessApi(domainId, groupId, userId)));
            if (!ownerAccess) {
                throw new GroupManagerServiceException("User does not have Owner permission to transfer group ownership");
            }
            return validatePutHttpRequest(serviceRequestClient.httpPut(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getTransferGroupOwnershipApi(groupGatewayId, groupId, newOwnerId), null));
        }
        catch (Exception e) {
            String msg = "Error Transferring Group Ownership";
            logger.error(msg, e);
            GroupManagerServiceException exception = new GroupManagerServiceException(msg + " More info : " + e.getMessage());
            throw exception;
        }

    }

    public boolean addGroupAdmins(String groupId,String groupGatewayId, List<String> adminIds, String userId, String domainId) throws GroupManagerServiceException {
        try {
            boolean ownerAccess = validateHeadHttpRequest(serviceRequestClient.httpHead(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getHasOwnerAccessApi(domainId, groupId, userId)));
            if (!ownerAccess) {
                throw new GroupManagerServiceException("User does not have Owner permission to add group admins");
            }
            return validatePutHttpRequest(serviceRequestClient.httpPost(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getAddGroupAdminsApi(groupGatewayId, groupId), adminIds));
        }
        catch (Exception e) {
            String msg = "Error Adding Admins to Group. Group ID: " + groupId;
            logger.error(msg, e);
            GroupManagerServiceException exception = new GroupManagerServiceException(msg + " More info : " + e.getMessage());
            throw exception;
        }
    }

    public boolean removeGroupAdmins(String groupId, String groupGatewayId,List<String> adminIds, String userId, String domainId) throws GroupManagerServiceException{
        try {
            boolean ownerAccess = validateHeadHttpRequest(serviceRequestClient.httpHead(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getHasOwnerAccessApi(domainId, groupId, userId)));
            if (!ownerAccess) {
                throw new GroupManagerServiceException("User does not have Owner permission to remove group admins");
            }
            return validatePutHttpRequest(serviceRequestClient.httpDelete(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getRemoveGroupAdminsApi(groupGatewayId, groupId)));
        }
        catch (Exception e) {
            String msg = "Error Removing Admins from the Group. Group ID: " + groupId;
            logger.error(msg, e);
            GroupManagerServiceException exception = new GroupManagerServiceException(msg + " More info : " + e.getMessage());
            throw exception;
        }
    }

    public boolean hasAdminAccess(String groupId, String adminId, String groupGatewayId) throws GroupManagerServiceException{
        try {
            boolean adminAccess = validateHeadHttpRequest(serviceRequestClient.httpHead(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getHasAdminAccessApi(groupGatewayId, groupId, adminId)));
            return adminAccess;
        }
        catch (Exception e) {
            String msg = "Error Checking Admin Access for the Group. Group ID: " + groupId + " Admin ID: " + adminId;
            logger.error(msg, e);
            GroupManagerServiceException exception = new GroupManagerServiceException(msg + " More info : " + e.getMessage());
            throw exception;
        }
    }

    public boolean hasOwnerAccess(String groupId, String groupGatewayId, String ownerId) throws GroupManagerServiceException{
        try {
            boolean ownerAccess = validateHeadHttpRequest(serviceRequestClient.httpHead(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getHasOwnerAccessApi(groupGatewayId, groupId, ownerId)));
            return ownerAccess;
        }
        catch (Exception e) {
            String msg = "Error Checking Owner Access for the Group. Group ID: " + groupId + " Owner ID: " + ownerId;
            logger.error(msg, e);
            GroupManagerServiceException exception = new GroupManagerServiceException(msg + " More info : " + e.getMessage());
            throw exception;
        }
    }


    private List<GroupModel> convertToGroupModels(List<UserGroup> userGroups){

        List<GroupModel> groupModels = new ArrayList<>();

        for (UserGroup userGroup: userGroups) {
            GroupModel groupModel = convertToGroupModel(userGroup);

            groupModels.add(groupModel);
        }
        return groupModels;
    }

    private GroupModel convertToGroupModel(UserGroup userGroup) {
        GroupModel groupModel = new GroupModel();
        groupModel.setId(userGroup.getGroupId());
        groupModel.setName(userGroup.getName());
        groupModel.setDescription(userGroup.getDescription());
        groupModel.setOwnerId(userGroup.getOwnerId());
        final List<String> admins = userGroup.getGroupAdmins().stream()
                .map(groupAdmin -> groupAdmin.getAdminId())
                .collect(Collectors.toList());
        groupModel.setAdmins(admins);

//        serviceRequestClient.httpGet(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getGroupMemebersOfTypeUserApi(userGroup.getDomainId(), userGroup.getGroupId(), 0, -1)).stream().forEach(user->
//                groupModel.getMembers().add((user.getUserId())
//        );
        return groupModel;
    }

    private boolean internalAddUsersToGroup(String domainId, List<String> userIds, String groupId) throws SharingRegistryException {

        // FIXME: workaround for UserProfiles that failed to sync to the sharing
        // registry: create any missing users in the sharing registry
//        for (String userId : userIds) {
//            if (!sharingClient.isUserExists(domainId, userId)) {
//                User user = new User();
//                user.setDomainId(domainId);
//                user.setUserId(userId);
//                UserProfile userProfile = userProfileRepository.get(userId);
//                user.setUserName(userProfile.getUserId());
//                user.setCreatedTime(userProfile.getCreationTime());
//                user.setEmail(userProfile.getEmailsSize() > 0 ? userProfile.getEmails().get(0) : null);
//                user.setFirstName(userProfile.getFirstName());
//                user.setLastName(userProfile.getLastName());
//                sharingClient.createUser(user);
//            }
//        }
//        return serviceRequestClient.httpPost(SHARING_SERVICE_BASE_URL, sharingRegistryEndpoints.getAddUsersToGroupApi(domainId, groupId), userIds);
        return true;
    }
}
