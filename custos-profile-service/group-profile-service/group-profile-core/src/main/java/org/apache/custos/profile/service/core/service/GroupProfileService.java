//package org.apache.custos.profile.service.core.service;
//
//import org.apache.custos.commons.utils.ServiceRequest;
//import org.apache.custos.profile.user.core.repositories.UserProfileRepository;
//import org.apache.custos.sharing.service.api.constants.SharingRegistryEndpoints;
//import org.apache.custos.sharing.service.core.models.GroupCardinality;
//import org.apache.custos.sharing.service.core.models.GroupType;
//import org.apache.custos.sharing.service.core.models.UserGroup;
//import org.apache.custos.profile.service.core.models.*;
//import org.apache.custos.profile.service.core.exceptions.*;
//import org.slf4j.Logger;
//import org.slf4j.LoggerFactory;
//import org.springframework.beans.factory.annotation.Autowired;
//
//import java.util.ArrayList;
//import java.util.List;
//import java.util.UUID;
//import java.util.stream.Collectors;
//
//public class GroupProfileService {
//
//    @Autowired
//    private static ServiceRequest serviceRequest;
//
//    @Autowired
//    private static SharingRegistryEndpoints SharingRegistryEndpoints;
//
//    private static final Logger logger = LoggerFactory.getLogger(GroupProfileService.class);
//    private static final String GROUP_MANAGER_CPI_VERSION = "1.0";
//    private UserProfileRepository userProfileRepository = new UserProfileRepository();
//    private static final String SHARING_SERVICE_BASE_URL = SharingRegistryEndpoints.BASE_URL;
//
//    public String getAPIVersion() {
//        return GROUP_MANAGER_CPI_VERSION;
//    }
//
//    public String createGroup(GroupModel groupModel, String gatewayId, String userId) throws GroupManagerServiceException{
//        try {
//
//            UserGroup sharingUserGroup = new UserGroup();
//            sharingUserGroup.setGroupId(UUID.randomUUID().toString());
//            sharingUserGroup.setName(groupModel.getName());
//            sharingUserGroup.setDescription(groupModel.getDescription());
//            sharingUserGroup.setGroupType(GroupType.USER_LEVEL_GROUP);
//            sharingUserGroup.setGroupCardinality(GroupCardinality.MULTI_USER);
//            sharingUserGroup.setDomainId(gatewayId);
//            sharingUserGroup.setOwnerId(userId);
//
//            String groupId = serviceRequest.httpPost(SHARING_SERVICE_BASE_URL, SharingRegistryEndpoints.CREATE_GROUP, );
//            //TODO: should happen in one rest call
//            //internalAddUsersToGroup(sharingClient, gatewayId, groupModel.getMembers(), groupId);
//            //addGroupAdmins(authzToken,groupId,groupModel.getAdmins());
//            return groupId;
//        }
//        catch (Exception e) {
//            String msg = "Error Creating Group" ;
//            logger.error(msg, e);
//            GroupManagerServiceException exception = new GroupManagerServiceException(msg + " More info : " + e.getMessage());
//            throw exception;
//        }
//    }
//
//    public boolean updateGroup(GroupModel groupModel, String gatewayId, String userId) throws GroupManagerServiceException{
//        try {
//            if (!(sharingClient.hasOwnerAccess(domainId, groupModel.getId(), userId)
//                    || sharingClient.hasAdminAccess(domainId, groupModel.getId(), userId))) {
//                throw new GroupManagerServiceException("User does not have permission to update group");
//            }
//
//            UserGroup sharingUserGroup = new UserGroup();
//            sharingUserGroup.setGroupId(groupModel.getId());
//            sharingUserGroup.setName(groupModel.getName());
//            sharingUserGroup.setDescription(groupModel.getDescription());
//            sharingUserGroup.setGroupType(GroupType.USER_LEVEL_GROUP);
//            sharingUserGroup.setDomainId(gatewayId);
//
//            //adding and removal of users should be handle separately??
//            String response = serviceRequest.httpPut(SHARING_SERVICE_BASE_URL, SharingRegistryEndpoints.UPDATE_GROUP);
//        }
//        catch (Exception e) {
//            String msg = "Error Updating Group" ;
//            logger.error(msg, e);
//            GroupManagerServiceException exception = new GroupManagerServiceException(msg + " More info : " + e.getMessage());
//            throw exception;
//        }
//    }
//
//    public boolean deleteGroup(AuthzToken authzToken, String groupId, String ownerId) throws GroupManagerServiceException, AuthorizationException, TException {
//        try {
//            SharingRegistryService.Client sharingClient = getSharingRegistryServiceClient();
//            String userId = getUserId(authzToken);
//            String domainId = getDomainId(authzToken);
//            if (!(sharingClient.hasOwnerAccess(domainId, groupId, userId))) {
//                throw new GroupManagerServiceException("User does not have permission to delete group");
//            }
//
//            sharingClient.deleteGroup(getDomainId(authzToken), groupId);
//            return true;
//        }
//        catch (Exception e) {
//            String msg = "Error Deleting Group. Group ID: " + groupId ;
//            logger.error(msg, e);
//            GroupManagerServiceException exception = new GroupManagerServiceException();
//            exception.setMessage(msg + " More info : " + e.getMessage());
//            throw exception;
//        }
//    }
//
//    @Override
//    @SecurityCheck
//    public GroupModel getGroup(AuthzToken authzToken, String groupId) throws GroupManagerServiceException, AuthorizationException, TException {
//        try {
//            SharingRegistryService.Client sharingClient = getSharingRegistryServiceClient();
//            final String domainId = getDomainId(authzToken);
//            UserGroup userGroup = sharingClient.getGroup(domainId, groupId);
//
//            GroupModel groupModel = convertToGroupModel(userGroup, sharingClient);
//
//            return groupModel;
//        }
//        catch (Exception e) {
//            String msg = "Error Retreiving Group. Group ID: " + groupId ;
//            logger.error(msg, e);
//            GroupManagerServiceException exception = new GroupManagerServiceException();
//            exception.setMessage(msg + " More info : " + e.getMessage());
//            throw exception;
//        }
//    }
//
//    @Override
//    @SecurityCheck
//    public List<GroupModel> getGroups(AuthzToken authzToken) throws GroupManagerServiceException, AuthorizationException, TException {
//        final String domainId = getDomainId(authzToken);
//        SharingRegistryService.Client sharingClient = null;
//        try {
//            sharingClient = getSharingRegistryServiceClient();
//            List<UserGroup> userGroups = sharingClient.getGroups(domainId, 0, -1);
//
//            return convertToGroupModels(userGroups, sharingClient);
//        }
//        catch (Exception e) {
//            String msg = "Error Retrieving Groups. Domain ID: " + domainId;
//            logger.error(msg, e);
//            GroupManagerServiceException exception = new GroupManagerServiceException();
//            exception.setMessage(msg + " More info : " + e.getMessage());
//            throw exception;
//        } finally {
//            closeSharingClient(sharingClient);
//        }
//    }
//
//    @Override
//    @SecurityCheck
//    public List<GroupModel> getAllGroupsUserBelongs(AuthzToken authzToken, String userName) throws GroupManagerServiceException, AuthorizationException, TException {
//        try {
//            SharingRegistryService.Client sharingClient = getSharingRegistryServiceClient();
//            List<GroupModel> groupModels = new ArrayList<GroupModel>();
//            final String domainId = getDomainId(authzToken);
//            List<UserGroup> userGroups = sharingClient.getAllMemberGroupsForUser(domainId, userName);
//
//            return convertToGroupModels(userGroups, sharingClient);
//        }
//        catch (Exception e) {
//            String msg = "Error Retreiving All Groups for User. User ID: " + userName ;
//            logger.error(msg, e);
//            GroupManagerServiceException exception = new GroupManagerServiceException();
//            exception.setMessage(msg + " More info : " + e.getMessage());
//            throw exception;
//        }
//    }
//
//    @Override
//    public boolean addUsersToGroup(AuthzToken authzToken, List<String> userIds, String groupId) throws GroupManagerServiceException, AuthorizationException, TException {
//        try {
//            SharingRegistryService.Client sharingClient = getSharingRegistryServiceClient();
//            String userId = getUserId(authzToken);
//            String domainId = getDomainId(authzToken);
//            if (!(sharingClient.hasOwnerAccess(domainId, groupId, userId)
//                    || sharingClient.hasAdminAccess(domainId, groupId, userId))) {
//                throw new GroupManagerServiceException("User does not have access to add users to the group");
//            }
//            return internalAddUsersToGroup(sharingClient, domainId, userIds, groupId);
//
//        } catch (Exception e) {
//            String msg = "Error adding users to group. Group ID: " + groupId ;
//            logger.error(msg, e);
//            GroupManagerServiceException exception = new GroupManagerServiceException();
//            exception.setMessage(msg + " More info : " + e.getMessage());
//            throw exception;
//        }
//    }
//
//    @Override
//    public boolean removeUsersFromGroup(AuthzToken authzToken, List<String> userIds, String groupId) throws GroupManagerServiceException, AuthorizationException, TException {
//        try {
//            SharingRegistryService.Client sharingClient = getSharingRegistryServiceClient();
//            String userId = getUserId(authzToken);
//            String domainId = getDomainId(authzToken);
//            if (!(sharingClient.hasOwnerAccess(domainId, groupId, userId)
//                    || sharingClient.hasAdminAccess(domainId, groupId, userId))) {
//                throw new GroupManagerServiceException("User does not have access to remove users to the group");
//            }
//            return sharingClient.removeUsersFromGroup(domainId, userIds, groupId);
//        } catch (Exception e) {
//            String msg = "Error remove users to group. Group ID: " + groupId ;
//            logger.error(msg, e);
//            GroupManagerServiceException exception = new GroupManagerServiceException();
//            exception.setMessage(msg + " More info : " + e.getMessage());
//            throw exception;
//        }
//    }
//
//    @Override
//    @SecurityCheck
//    public boolean transferGroupOwnership(AuthzToken authzToken, String groupId, String newOwnerId) throws GroupManagerServiceException, AuthorizationException, TException {
//        try{
//            SharingRegistryService.Client sharingClient = getSharingRegistryServiceClient();
//            String userId = getUserId(authzToken);
//            String domainId = getDomainId(authzToken);
//            if (!(sharingClient.hasOwnerAccess(domainId, groupId, userId))) {
//                throw new GroupManagerServiceException("User does not have Owner permission to transfer group ownership");
//            }
//            return sharingClient.transferGroupOwnership(getDomainId(authzToken), groupId, newOwnerId);
//        }
//        catch (Exception e) {
//            String msg = "Error Transferring Group Ownership";
//            logger.error(msg, e);
//            GroupManagerServiceException exception = new GroupManagerServiceException();
//            exception.setMessage(msg + " More info : " + e.getMessage());
//            throw exception;
//        }
//
//    }
//
//    @Override
//    @SecurityCheck
//    public boolean addGroupAdmins(AuthzToken authzToken, String groupId, List<String> adminIds) throws GroupManagerServiceException, AuthorizationException, TException {
//        try {
//            SharingRegistryService.Client sharingClient = getSharingRegistryServiceClient();
//            String userId = getUserId(authzToken);
//            String domainId = getDomainId(authzToken);
//            if (!(sharingClient.hasOwnerAccess(domainId, groupId, userId))) {
//                throw new GroupManagerServiceException("User does not have Owner permission to add group admins");
//            }
//            return sharingClient.addGroupAdmins(getDomainId(authzToken), groupId, adminIds);
//        }
//        catch (Exception e) {
//            String msg = "Error Adding Admins to Group. Group ID: " + groupId;
//            logger.error(msg, e);
//            GroupManagerServiceException exception = new GroupManagerServiceException();
//            exception.setMessage(msg + " More info : " + e.getMessage());
//            throw exception;
//        }
//    }
//
//    @Override
//    @SecurityCheck
//    public boolean removeGroupAdmins(AuthzToken authzToken, String groupId, List<String> adminIds) throws GroupManagerServiceException, AuthorizationException, TException {
//        try {
//            SharingRegistryService.Client sharingClient = getSharingRegistryServiceClient();
//            String userId = getUserId(authzToken);
//            String domainId = getDomainId(authzToken);
//            if (!(sharingClient.hasOwnerAccess(domainId, groupId, userId))) {
//                throw new GroupManagerServiceException("User does not have Owner permission to remove group admins");
//            }
//            return sharingClient.removeGroupAdmins(getDomainId(authzToken), groupId, adminIds);
//        }
//        catch (Exception e) {
//            String msg = "Error Removing Admins from the Group. Group ID: " + groupId;
//            logger.error(msg, e);
//            GroupManagerServiceException exception = new GroupManagerServiceException();
//            exception.setMessage(msg + " More info : " + e.getMessage());
//            throw exception;
//        }
//    }
//
//    @Override
//    @SecurityCheck
//    public boolean hasAdminAccess(AuthzToken authzToken, String groupId, String adminId) throws GroupManagerServiceException, AuthorizationException, TException {
//        try {
//            SharingRegistryService.Client sharingClient = getSharingRegistryServiceClient();
//            return sharingClient.hasAdminAccess(getDomainId(authzToken), groupId, adminId);
//        }
//        catch (Exception e) {
//            String msg = "Error Checking Admin Access for the Group. Group ID: " + groupId + " Admin ID: " + adminId;
//            logger.error(msg, e);
//            GroupManagerServiceException exception = new GroupManagerServiceException();
//            exception.setMessage(msg + " More info : " + e.getMessage());
//            throw exception;
//        }
//    }
//
//    @Override
//    @SecurityCheck
//    public boolean hasOwnerAccess(AuthzToken authzToken, String groupId, String ownerId) throws GroupManagerServiceException, AuthorizationException, TException {
//        try {
//            SharingRegistryService.Client sharingClient = getSharingRegistryServiceClient();
//            return sharingClient.hasOwnerAccess(getDomainId(authzToken), groupId, ownerId);
//        }
//        catch (Exception e) {
//            String msg = "Error Checking Owner Access for the Group. Group ID: " + groupId + " Owner ID: " + ownerId;
//            logger.error(msg, e);
//            GroupManagerServiceException exception = new GroupManagerServiceException();
//            exception.setMessage(msg + " More info : " + e.getMessage());
//            throw exception;
//        }
//    }
//
//    // TODO: replace these methods with ThriftClientPool (see AIRAVATA-2607)
//    private SharingRegistryService.Client getSharingRegistryServiceClient() throws TException, ApplicationSettingsException {
//        final int serverPort = Integer.parseInt(ServerSettings.getSharingRegistryPort());
//        final String serverHost = ServerSettings.getSharingRegistryHost();
//        try {
//            return SharingRegistryServiceClientFactory.createSharingRegistryClient(serverHost, serverPort);
//        } catch (SharingRegistryException e) {
//            throw new TException("Unable to create sharing registry client...", e);
//        }
//    }
//
//    private String getDomainId(AuthzToken authzToken) {
//        return authzToken.getClaimsMap().get(Constants.GATEWAY_ID);
//    }
//
//    private String getUserId(AuthzToken authzToken) {
//        return authzToken.getClaimsMap().get(Constants.USER_NAME) + "@" + getDomainId(authzToken);
//    }
//
//    private List<GroupModel> convertToGroupModels(List<UserGroup> userGroups, SharingRegistryService.Client sharingClient) throws TException {
//
//        List<GroupModel> groupModels = new ArrayList<>();
//
//        for (UserGroup userGroup: userGroups) {
//            GroupModel groupModel = convertToGroupModel(userGroup, sharingClient);
//
//            groupModels.add(groupModel);
//        }
//        return groupModels;
//    }
//
//    private GroupModel convertToGroupModel(UserGroup userGroup, SharingRegistryService.Client sharingClient) throws TException {
//        GroupModel groupModel = new GroupModel();
//        groupModel.setId(userGroup.getGroupId());
//        groupModel.setName(userGroup.getName());
//        groupModel.setDescription(userGroup.getDescription());
//        groupModel.setOwnerId(userGroup.getOwnerId());
//        final List<String> admins = userGroup.getGroupAdmins().stream()
//                .map(groupAdmin -> groupAdmin.getAdminId())
//                .collect(Collectors.toList());
//        groupModel.setAdmins(admins);
//
//        sharingClient.getGroupMembersOfTypeUser(userGroup.getDomainId(), userGroup.getGroupId(), 0, -1).stream().forEach(user->
//                groupModel.addToMembers(user.getUserId())
//        );
//        return groupModel;
//    }
//
//    private void closeSharingClient(SharingRegistryService.Client sharingClient) {
//        if (sharingClient != null) {
//            if (sharingClient.getInputProtocol().getTransport().isOpen()) {
//                sharingClient.getInputProtocol().getTransport().close();
//            }
//            if (sharingClient.getOutputProtocol().getTransport().isOpen()) {
//                sharingClient.getOutputProtocol().getTransport().close();
//            }
//        }
//    }
//
//    private boolean internalAddUsersToGroup(SharingRegistryService.Client sharingClient, String domainId, List<String> userIds, String groupId) throws SharingRegistryException, TException {
//
//        // FIXME: workaround for UserProfiles that failed to sync to the sharing
//        // registry: create any missing users in the sharing registry
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
//        return sharingClient.addUsersToGroup(domainId, userIds, groupId);
//    }
//}
//
//}
