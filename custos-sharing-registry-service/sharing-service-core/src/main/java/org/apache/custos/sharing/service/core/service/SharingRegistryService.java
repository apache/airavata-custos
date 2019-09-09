/**
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
package org.apache.custos.sharing.service.core.service;

import org.apache.commons.lang.exception.ExceptionUtils;
import org.apache.custos.commons.utils.DBInitializer;
import org.apache.custos.sharing.service.core.exceptions.InvalidRequestException;
import org.apache.custos.sharing.service.core.exceptions.ResourceNotFoundException;
import org.apache.custos.sharing.service.core.models.*;
import org.apache.custos.sharing.service.core.db.entities.*;
import org.apache.custos.sharing.service.core.db.repositories.*;
import org.apache.custos.sharing.service.core.db.utils.DBConstants;
import org.apache.custos.sharing.service.core.db.utils.SharingRegistryDBInitConfig;
import org.apache.custos.sharing.service.core.exceptions.DuplicateEntryException;
import org.apache.custos.sharing.service.core.exceptions.SharingRegistryException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import java.lang.reflect.Field;
import java.lang.reflect.Modifier;
import java.util.*;
import java.util.stream.Collectors;

@Service
public class SharingRegistryService {
    private final static Logger logger = LoggerFactory.getLogger(SharingRegistryService.class);

    public static String OWNER_PERMISSION_NAME = "OWNER";
    public static String SHARING_CPI_VERSION = "1.0";

    public SharingRegistryService(SharingRegistryDBInitConfig sharingRegistryDBInitConfig) {
        DBInitializer.initializeDB(sharingRegistryDBInitConfig);
    }

    public String getAPIVersion() {
        return SHARING_CPI_VERSION;
    }

    /**
     * * Domain Operations
     * *
     */

    public Domain createDomain(Domain domain) throws SharingRegistryException, DuplicateEntryException {
        try{
            if((new DomainRepository()).get(domain.getDomainId()) != null)
                throw new DuplicateEntryException("There exist domain with given domain id");

            domain.setCreatedTime(System.currentTimeMillis());
            domain.setUpdatedTime(System.currentTimeMillis());
            Domain createdDomain = (new DomainRepository()).create(domain);

            //create the global permission for the domain
            PermissionType permissionType = new PermissionType();
            permissionType.setPermissionTypeId(domain.getDomainId() + ":" + OWNER_PERMISSION_NAME);
            permissionType.setDomainId(domain.getDomainId());
            permissionType.setName(OWNER_PERMISSION_NAME);
            permissionType.setDescription("GLOBAL permission to " + domain.getDomainId());
            permissionType.setCreatedTime(System.currentTimeMillis());
            permissionType.setUpdatedTime(System.currentTimeMillis());
            (new PermissionTypeRepository()).create(permissionType);

            return createdDomain;
        }catch(DuplicateEntryException ex){
            throw ex;
        }catch (Throwable ex){
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean updateDomain(Domain domain) throws SharingRegistryException, ResourceNotFoundException {
        try{
            Domain oldDomain = (new DomainRepository()).get(domain.getDomainId());
            if(oldDomain == null){
                throw new ResourceNotFoundException("Could not find the domain with domainId: "+ domain.getDomainId());
            }
            domain.setCreatedTime(oldDomain.getCreatedTime());
            domain.setUpdatedTime(System.currentTimeMillis());
            domain = getUpdatedObject(oldDomain, domain);
            Domain updatedDomain = (new DomainRepository()).update(domain);
            if(updatedDomain != null && updatedDomain.getDomainId().equals(oldDomain.getDomainId())){
                return true;
            }
            throw new SharingRegistryException("Could not update the domain with domainId: "+ oldDomain.getDomainId());
        }catch(ResourceNotFoundException ex) {
            throw ex;
        }catch(Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    /**
     * <p>API method to check Domain Exists</p>
     *
     * @param domainId
     */
    public boolean isDomainExists(String domainId) throws SharingRegistryException {
        try{
            return (new DomainRepository()).isExists(domainId);
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean deleteDomain(String domainId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isDomainExists(domainId)){
                throw new ResourceNotFoundException("Could not find domain with domainId: "+ domainId);
            }
            boolean deleted = (new DomainRepository()).delete(domainId);
            if(deleted) return true;
            else throw new SharingRegistryException("Could not delete the domain with domainId: "+ domainId);
        }catch (ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public Domain getDomain(String domainId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            Domain domain = (new DomainRepository()).get(domainId);
            if(domain == null){
                throw new ResourceNotFoundException("Could not find domain with domainId:" + domainId);
            }
            return domain;
        }catch (ResourceNotFoundException ex){
            throw ex;
        } catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public List<Domain> getDomains(int offset, int limit) throws SharingRegistryException {
        try{
            return (new DomainRepository()).select(new HashMap<>(), offset, limit);
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    /**
     * * User Operations
     * *
     */
    public User createUser(User user) throws SharingRegistryException, DuplicateEntryException {
        try{
            UserPK userPK = new UserPK();
            userPK.setUserId(user.getUserId());
            userPK.setDomainId(user.getDomainId());
            if((new UserRepository()).get(userPK) != null)
                throw new DuplicateEntryException("There exist user with given user id");

            user.setCreatedTime(System.currentTimeMillis());
            user.setUpdatedTime(System.currentTimeMillis());
            User createdUser = (new UserRepository()).create(user);

            UserGroup userGroup = new UserGroup();
            userGroup.setGroupId(user.getUserId());
            userGroup.setDomainId(user.getDomainId());
            userGroup.setName(user.getUserName());
            userGroup.setDescription("user " + user.getUserName() + " group");
            userGroup.setOwnerId(user.getUserId());
            userGroup.setGroupType(GroupType.USER_LEVEL_GROUP);
            userGroup.setGroupCardinality(GroupCardinality.SINGLE_USER);
            userGroup.setCreatedTime(System.currentTimeMillis());
            userGroup.setUpdatedTime(System.currentTimeMillis());
            (new UserGroupRepository()).create(userGroup);

            return createdUser;
        }catch(DuplicateEntryException ex){
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean updatedUser(User user) throws SharingRegistryException {
        try{
            if(!isUserExists(user.getDomainId(), user.getUserId())){
                throw new ResourceNotFoundException("Could not find user with userId: " + user.getUserId());
            }
            UserPK userPK = new UserPK();
            userPK.setUserId(user.getUserId());
            userPK.setDomainId(user.getDomainId());
            User oldUser = (new UserRepository()).get(userPK);
            user.setCreatedTime(oldUser.getCreatedTime());
            user.setUpdatedTime(System.currentTimeMillis());
            user = getUpdatedObject(oldUser, user);
            User updatedUser = (new UserRepository()).update(user);

            UserGroupPK userGroupPK = new UserGroupPK();
            userGroupPK.setGroupId(user.getUserId());
            userGroupPK.setDomainId(user.getDomainId());
            UserGroup userGroup = (new UserGroupRepository()).get(userGroupPK);
            userGroup.setName(user.getUserName());
            userGroup.setDescription("user " + user.getUserName() + " group");
            updateGroup(userGroup);
            if(updatedUser != null && updatedUser.getUserId().equals(user.getUserId())){
                return true;
            }
            throw new SharingRegistryException("Could not update user with userId: " + user.getUserId());
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    /**
     * <p>API method to check User Exists</p>
     *
     * @param userId
     */
    public boolean isUserExists(String domainId, String userId) throws SharingRegistryException {
        try{
            UserPK userPK = new UserPK();
            userPK.setDomainId(domainId);
            userPK.setUserId(userId);
            return (new UserRepository()).isExists(userPK);
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean deleteUser(String domainId, String userId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isUserExists(domainId, userId)) throw new ResourceNotFoundException("Could not find userId: "+userId +" in domainId:" + domainId);
            UserPK userPK = new UserPK();
            userPK.setUserId(userId);
            userPK.setDomainId(domainId);
            boolean deleteuser = (new UserRepository()).delete(userPK);

            UserGroupPK userGroupPK = new UserGroupPK();
            userGroupPK.setGroupId(userId);
            userGroupPK.setDomainId(domainId);
            boolean deleteusergroup = (new UserGroupRepository()).delete(userGroupPK);
            if(deleteuser && deleteusergroup) return true;
            throw new SharingRegistryException("Could not delete user: "+ userId);
        }catch (ResourceNotFoundException ex){
            throw ex;
        }
        catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public User getUser(String domainId, String userId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isUserExists(domainId, userId)) throw new ResourceNotFoundException("Could not find userId: "+userId +" in domainId:" + domainId);
            UserPK userPK = new UserPK();
            userPK.setUserId(userId);
            userPK.setDomainId(domainId);
            return (new UserRepository()).get(userPK);
        }catch (ResourceNotFoundException ex){
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public List<User> getUsers(String domain, int offset, int limit) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isDomainExists(domain)) throw new ResourceNotFoundException("Could not find domain with domainId: "+ domain);
            HashMap<String, String> filters = new HashMap<>();
            filters.put(DBConstants.UserTable.DOMAIN_ID, domain);
            return (new UserRepository()).select(filters, offset, limit);
        }catch (ResourceNotFoundException ex){
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    /**
     * * Group Operations
     * *
     */
    public UserGroup createGroup(UserGroup group) throws DuplicateEntryException, SharingRegistryException{
        try {
            UserGroupPK userGroupPK = new UserGroupPK();
            userGroupPK.setGroupId(group.getGroupId());
            userGroupPK.setDomainId(group.getDomainId());
            if ((new UserGroupRepository()).get(userGroupPK) != null)
                throw new DuplicateEntryException("There exists group with given group id");
            //Client created groups are always of type MULTI_USER
            group.setGroupCardinality(GroupCardinality.MULTI_USER);
            group.setCreatedTime(System.currentTimeMillis());
            group.setUpdatedTime(System.currentTimeMillis());
            //Add group admins once the group is created
            //group.unsetGroupAdmins();
            UserGroup createdUserGroup = (new UserGroupRepository()).create(group);

            addUsersToGroup(group.getDomainId(), Arrays.asList(group.getOwnerId()), group.getGroupId());
            addGroupAdmins(group.getDomainId(), group.getGroupId(), group.getGroupAdmins().stream().map(y->y.getAdminId()).collect(Collectors.toList()));
            return createdUserGroup;
        } catch (DuplicateEntryException ex) {
            throw ex;
        } catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean updateGroup(UserGroup group) throws SharingRegistryException, ResourceNotFoundException, InvalidRequestException {
        try {
            if (isGroupExists(group.getDomainId(), group.getGroupId())) {
                throw new ResourceNotFoundException("Could not find the group with groupId: " + group.getGroupId());
            }
            group.setUpdatedTime(System.currentTimeMillis());
            UserGroupPK userGroupPK = new UserGroupPK();
            userGroupPK.setGroupId(group.getGroupId());
            userGroupPK.setDomainId(group.getDomainId());
            UserGroup oldGroup = (new UserGroupRepository()).get(userGroupPK);
            if(!group.getOwnerId().equals(oldGroup.getOwnerId()))
                throw new InvalidRequestException("Group owner cannot be changed");
            group.setGroupCardinality(oldGroup.getGroupCardinality());
            group.setCreatedTime(oldGroup.getCreatedTime());
            group = getUpdatedObject(oldGroup, group);
            UserGroup updatedGroup = (new UserGroupRepository()).update(group);
            if (!updatedGroup.getGroupId().equals(oldGroup.getGroupId()))
                throw new SharingRegistryException("Group could not be updated");
            return true;
        }catch (InvalidRequestException | ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    /**
     * API method to check Group Exists
     * @param domainId
     * @param groupId
     * @return
     * @throws SharingRegistryException
     */

    public boolean isGroupExists(String domainId, String groupId) throws SharingRegistryException {
        try{
            if (isGroupExists(domainId, groupId)) {
                throw new ResourceNotFoundException("Could not find the group with groupId: " + groupId);
            }
            UserGroupPK userGroupPK = new UserGroupPK();
            userGroupPK.setDomainId(domainId);
            userGroupPK.setGroupId(groupId);
            return (new UserGroupRepository()).isExists(userGroupPK);
        }catch (ResourceNotFoundException ex) {
            throw ex;
        }
        catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean deleteGroup(String domainId, String groupId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if (isGroupExists(domainId, groupId)) {
                throw new ResourceNotFoundException("Could not find the group with groupId: " + groupId);
            }
            UserGroupPK userGroupPK = new UserGroupPK();
            userGroupPK.setGroupId(groupId);
            userGroupPK.setDomainId(domainId);
            boolean deleted = (new UserGroupRepository()).delete(userGroupPK);
            if(!deleted) throw new SharingRegistryException("Could not deleted the group with groupId: " + groupId);
            return true;
        }catch (ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public UserGroup getGroup(String domainId, String groupId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            UserGroupPK userGroupPK = new UserGroupPK();
            userGroupPK.setGroupId(groupId);
            userGroupPK.setDomainId(domainId);
            UserGroup userGroup = (new UserGroupRepository()).get(userGroupPK);
            if(userGroup != null) return userGroup;
            throw new ResourceNotFoundException("Could not find the group with groupId: " + groupId);
        }catch (ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public List<UserGroup> getGroups(String domain, int offset, int limit) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isDomainExists(domain)) throw new ResourceNotFoundException("Could not find the domain with domainId: "+domain);
            HashMap<String, String> filters = new HashMap<>();
            filters.put(DBConstants.UserGroupTable.DOMAIN_ID, domain);
            // Only return groups with MULTI_USER cardinality which is the only type of cardinality allowed for client created groups
            filters.put(DBConstants.UserGroupTable.GROUP_CARDINALITY, GroupCardinality.MULTI_USER.name());
            return (new UserGroupRepository()).select(filters, offset, limit);
        }catch (ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean addUsersToGroup(String domainId, List<String> userIds, String groupId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isDomainExists(domainId)) throw new ResourceNotFoundException("Could not find the domain with domainId: "+domainId);
            for(int i=0; i < userIds.size(); i++){
                GroupMembership groupMembership = new GroupMembership();
                groupMembership.setParentId(groupId);
                groupMembership.setChildId(userIds.get(i));
                groupMembership.setChildType(GroupChildType.USER);
                groupMembership.setDomainId(domainId);
                groupMembership.setCreatedTime(System.currentTimeMillis());
                groupMembership.setUpdatedTime(System.currentTimeMillis());
                (new GroupMembershipRepository()).create(groupMembership);
            }
            return true;
        }catch (ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean removeUsersFromGroup(String domainId, List<String> userIds, String groupId) throws SharingRegistryException, ResourceNotFoundException, InvalidRequestException{
        try{
            if(!isDomainExists(domainId)) throw new ResourceNotFoundException("Could not find the domain with domainId: "+domainId);
            for (String userId: userIds) {
                if (hasOwnerAccess(domainId, groupId, userId)) {
                    throw new InvalidRequestException("List of User Ids contains Owner Id. Cannot remove owner from the group");
                }
            }

            for(int i=0; i < userIds.size(); i++){
                GroupMembershipPK groupMembershipPK = new GroupMembershipPK();
                groupMembershipPK.setParentId(groupId);
                groupMembershipPK.setChildId(userIds.get(i));
                groupMembershipPK.setDomainId(domainId);
                (new GroupMembershipRepository()).delete(groupMembershipPK);
            }
            return true;
        }catch (ResourceNotFoundException | InvalidRequestException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean transferGroupOwnership(String domainId, String groupId, String newOwnerId) throws SharingRegistryException {
        try {
            if(!isDomainExists(domainId) || !isGroupExists(domainId, groupId) || !isUserExists(domainId, newOwnerId)) throw new ResourceNotFoundException("Could not find resource");
            List<User> groupUser = getGroupMembersOfTypeUser(domainId, groupId, 0, -1);
            if (!isUserBelongsToGroup(groupUser, newOwnerId)) {
                throw new InvalidRequestException("New group owner is not part of the group");
            }

            if (hasOwnerAccess(domainId, groupId, newOwnerId)) {
                throw new InvalidRequestException("User already the current owner of the group");
            }
            // remove the new owner as Admin if present
            if (hasAdminAccess(domainId, groupId, newOwnerId)) {
                removeGroupAdmins(domainId, groupId, Arrays.asList(newOwnerId));
            }

            UserGroupPK userGroupPK = new UserGroupPK();
            userGroupPK.setGroupId(groupId);
            userGroupPK.setDomainId(domainId);
            UserGroup userGroup = (new UserGroupRepository()).get(userGroupPK);
            UserGroup newUserGroup = new UserGroup();
            newUserGroup.setUpdatedTime(System.currentTimeMillis());
            newUserGroup.setOwnerId(newOwnerId);
            newUserGroup.setGroupCardinality(GroupCardinality.MULTI_USER);
            newUserGroup.setCreatedTime(userGroup.getCreatedTime());
            newUserGroup = getUpdatedObject(userGroup, newUserGroup);

            (new UserGroupRepository()).update(newUserGroup);

            return true;
        }catch (ResourceNotFoundException | InvalidRequestException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    private boolean isUserBelongsToGroup(List<User> groupUser, String newOwnerId) {
        for (User user: groupUser) {
            if (user.getUserId().equals(newOwnerId)) {
                return true;
            }
        }
        return false;
    }

    public boolean addGroupAdmins(String domainId, String groupId, List<String> admins) throws SharingRegistryException, ResourceNotFoundException,DuplicateEntryException, InvalidRequestException  {
        try{
            if(!isDomainExists(domainId) || !isGroupExists(domainId, groupId)) throw new ResourceNotFoundException("Could not find resource");
            List<User> groupUser = getGroupMembersOfTypeUser(domainId, groupId, 0, -1);

            for (String adminId: admins) {
                if (! isUserBelongsToGroup(groupUser, adminId)) {
                    throw new InvalidRequestException("Admin not the user of the group. GroupId : "+ groupId + ", AdminId : "+ adminId);
                }
                GroupAdminPK groupAdminPK = new GroupAdminPK();
                groupAdminPK.setGroupId(groupId);
                groupAdminPK.setAdminId(adminId);
                groupAdminPK.setDomainId(domainId);
                GroupAdmin groupAdmin = new GroupAdmin();
                groupAdmin.setAdminId(adminId);
                groupAdmin.setDomainId(domainId);
                groupAdmin.setGroupId(groupId);
                if((new GroupAdminRepository()).get(groupAdminPK) != null)
                    throw new DuplicateEntryException("User already an admin for the group");

                (new GroupAdminRepository()).create(groupAdmin);
            }
            return true;
        }catch (ResourceNotFoundException | DuplicateEntryException | InvalidRequestException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean removeGroupAdmins(String domainId, String groupId, List<String> adminIds) throws SharingRegistryException {
        try {
            for (String adminId: adminIds) {
                GroupAdminPK groupAdminPK = new GroupAdminPK();
                groupAdminPK.setAdminId(adminId);
                groupAdminPK.setDomainId(domainId);
                groupAdminPK.setGroupId(groupId);
                (new GroupAdminRepository()).delete(groupAdminPK);
            }
            return true;
        }
        catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean hasAdminAccess(String domainId, String groupId, String adminId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isDomainExists(domainId) || !isGroupExists(domainId, groupId) || !isUserExists(domainId, adminId)) throw new ResourceNotFoundException("Could not find resource");
            GroupAdminPK groupAdminPK = new GroupAdminPK();
            groupAdminPK.setGroupId(groupId);
            groupAdminPK.setAdminId(adminId);
            groupAdminPK.setDomainId(domainId);

            if((new GroupAdminRepository()).get(groupAdminPK) != null)
                return true;
            return false;
        }catch (ResourceNotFoundException ex){
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean hasOwnerAccess(String domainId, String groupId, String ownerId) throws SharingRegistryException {
        try {
            if(!isDomainExists(domainId) || !isGroupExists(domainId, groupId) || !isUserExists(domainId, ownerId)) throw new ResourceNotFoundException("Could not find resource");
            UserGroupPK userGroupPK = new UserGroupPK();
            userGroupPK.setGroupId(groupId);
            userGroupPK.setDomainId(domainId);
            UserGroup getGroup = (new UserGroupRepository()).get(userGroupPK);

            if(getGroup.getOwnerId().equals(ownerId))
                return true;
            return false;
        }catch (ResourceNotFoundException ex){
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public List<User> getGroupMembersOfTypeUser(String domainId, String groupId, int offset, int limit) throws SharingRegistryException {
        try{
            if(!isDomainExists(domainId) || !isGroupExists(domainId, groupId)) throw new ResourceNotFoundException("Could not find resource");
            //TODO limit offset
            List<User> groupMemberUsers = (new GroupMembershipRepository()).getAllChildUsers(domainId, groupId);
            return groupMemberUsers;
        }catch (ResourceNotFoundException ex){
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public List<UserGroup> getGroupMembersOfTypeGroup(String domainId, String groupId, int offset, int limit) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isDomainExists(domainId) || !isGroupExists(domainId, groupId)) throw new ResourceNotFoundException("Could not find resource");
            //TODO limit offset
            List<UserGroup> groupMemberGroups = (new GroupMembershipRepository()).getAllChildGroups(domainId, groupId);
            return groupMemberGroups;
        }catch (ResourceNotFoundException ex){
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean addChildGroupsToParentGroup(String domainId, List<String> childIds, String groupId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isDomainExists(domainId) || !isGroupExists(domainId, groupId)) throw new ResourceNotFoundException("Could not find resource");
            for(String childId : childIds) {
                //Todo check for cyclic dependencies
                if(!isGroupExists(domainId, childId)) throw new ResourceNotFoundException("Could not find resource");
                GroupMembership groupMembership = new GroupMembership();
                groupMembership.setParentId(groupId);
                groupMembership.setChildId(childId);
                groupMembership.setChildType(GroupChildType.GROUP);
                groupMembership.setDomainId(domainId);
                groupMembership.setCreatedTime(System.currentTimeMillis());
                groupMembership.setUpdatedTime(System.currentTimeMillis());
                (new GroupMembershipRepository()).create(groupMembership);
            }
            return true;
        }catch (ResourceNotFoundException ex){
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean removeChildGroupFromParentGroup(String domainId, String childId, String groupId) throws SharingRegistryException {
        try{
            if(!isDomainExists(domainId) || !isGroupExists(domainId, groupId) || !isGroupExists(domainId, childId)) throw new ResourceNotFoundException("Could not find resource");
            GroupMembershipPK groupMembershipPK = new GroupMembershipPK();
            groupMembershipPK.setParentId(groupId);
            groupMembershipPK.setChildId(childId);
            groupMembershipPK.setDomainId(domainId);
            (new GroupMembershipRepository()).delete(groupMembershipPK);
            return true;
        }catch (ResourceNotFoundException ex){
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public List<UserGroup> getAllMemberGroupsForUser(String domainId, String userId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isDomainExists(domainId) || !isUserExists(domainId, userId)) throw new ResourceNotFoundException("Could not find resource");
            GroupMembershipRepository groupMembershipRepository = new GroupMembershipRepository();
            return groupMembershipRepository.getAllMemberGroupsForUser(domainId, userId);
        }catch (ResourceNotFoundException ex){
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    /**
     * * EntityType Operations
     * *
     */
    public EntityType createEntityType(EntityType entityType) throws SharingRegistryException, DuplicateEntryException {
        try{
            EntityTypePK entityTypePK = new EntityTypePK();
            entityTypePK.setDomainId(entityType.getDomainId());
            entityTypePK.setEntityTypeId(entityType.getEntityTypeId());
            if((new EntityTypeRepository()).get(entityTypePK) != null)
                throw new DuplicateEntryException("There exist EntityType with given EntityType id");

            entityType.setCreatedTime(System.currentTimeMillis());
            entityType.setUpdatedTime(System.currentTimeMillis());
            return (new EntityTypeRepository()).create(entityType);
        }catch (DuplicateEntryException ex) {
            throw ex;
        }catch(Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean updateEntityType(EntityType entityType) throws SharingRegistryException, ResourceNotFoundException{
        try{
            if(!isEntityTypeExists(entityType.getDomainId(), entityType.getEntityTypeId())) throw new ResourceNotFoundException("Could not find entity type with id: " + entityType.getEntityTypeId());
            entityType.setUpdatedTime(System.currentTimeMillis());
            EntityTypePK entityTypePK = new EntityTypePK();
            entityTypePK.setDomainId(entityType.getDomainId());
            entityTypePK.setEntityTypeId(entityType.getEntityTypeId());
            EntityType oldEntityType = (new EntityTypeRepository()).get(entityTypePK);
            entityType.setCreatedTime(oldEntityType.getCreatedTime());
            entityType = getUpdatedObject(oldEntityType, entityType);
            EntityType updatedEntityType = (new EntityTypeRepository()).update(entityType);
            if(updatedEntityType != null && updatedEntityType.getEntityTypeId().equals(entityType.getEntityTypeId())) return true;
            throw new SharingRegistryException("Could not update the entity type");
        }catch (ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    /**
     * <p>API method to check EntityType Exists</p>
     *
     * @param entityTypeId
     */
    public boolean isEntityTypeExists(String domainId, String entityTypeId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isEntityTypeExists(domainId, entityTypeId)) throw new ResourceNotFoundException("Could not find entity type with id: " + entityTypeId);
            EntityTypePK entityTypePK = new EntityTypePK();
            entityTypePK.setDomainId(domainId);
            entityTypePK.setEntityTypeId(entityTypeId);
            return (new EntityTypeRepository()).isExists(entityTypePK);
        }catch (ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean deleteEntityType(String domainId, String entityTypeId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isEntityTypeExists(domainId, entityTypeId)) throw new ResourceNotFoundException("Could not find entity type with id: " + entityTypeId);
            EntityTypePK entityTypePK = new EntityTypePK();
            entityTypePK.setDomainId(domainId);
            entityTypePK.setEntityTypeId(entityTypeId);
            (new EntityTypeRepository()).delete(entityTypePK);
            return true;
        }catch (ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public EntityType getEntityType(String domainId, String entityTypeId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isEntityTypeExists(domainId, entityTypeId)) throw new ResourceNotFoundException("Could not find entity type with id: " + entityTypeId);
            EntityTypePK entityTypePK = new EntityTypePK();
            entityTypePK.setDomainId(domainId);
            entityTypePK.setEntityTypeId(entityTypeId);
            return (new EntityTypeRepository()).get(entityTypePK);
        }catch (ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public List<EntityType> getEntityTypes(String domain, int offset, int limit) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isDomainExists(domain)) throw new ResourceNotFoundException("Could not find domain with id: " + domain);
            HashMap<String, String> filters = new HashMap<>();
            filters.put(DBConstants.EntityTypeTable.DOMAIN_ID, domain);
            return (new EntityTypeRepository()).select(filters, offset, limit);
        }catch (ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    /**
     * * Permission Operations
     * *
     */
    public PermissionType createPermissionType(PermissionType permissionType) throws SharingRegistryException, DuplicateEntryException {
        try{
            PermissionTypePK permissionTypePK =  new PermissionTypePK();
            permissionTypePK.setDomainId(permissionType.getDomainId());
            permissionTypePK.setPermissionTypeId(permissionType.getPermissionTypeId());
            if((new PermissionTypeRepository()).get(permissionTypePK) != null)
                throw new DuplicateEntryException("There exist PermissionType with given PermissionType id");
            permissionType.setCreatedTime(System.currentTimeMillis());
            permissionType.setUpdatedTime(System.currentTimeMillis());
            (new PermissionTypeRepository()).create(permissionType);
            return permissionType;
        }catch (DuplicateEntryException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean updatePermissionType(PermissionType permissionType) throws SharingRegistryException, ResourceNotFoundException{
        try{
            if(!isPermissionExists(permissionType.getDomainId(), permissionType.getPermissionTypeId())) throw new ResourceNotFoundException("Permission Type not found");
            permissionType.setUpdatedTime(System.currentTimeMillis());
            PermissionTypePK permissionTypePK =  new PermissionTypePK();
            permissionTypePK.setDomainId(permissionType.getDomainId());
            permissionTypePK.setPermissionTypeId(permissionType.getPermissionTypeId());
            PermissionType oldPermissionType = (new PermissionTypeRepository()).get(permissionTypePK);
            permissionType = getUpdatedObject(oldPermissionType, permissionType);
            PermissionType updatedPermissionType = (new PermissionTypeRepository()).update(permissionType);
            if(updatedPermissionType != null && updatedPermissionType.getPermissionTypeId().equals(permissionType.getPermissionTypeId())) return true;
            throw new SharingRegistryException("Could not update the permission type");
        }catch (ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    /**
     * <p>API method to check Permission Exists</p>
     *
     * @param permissionId
     */
    public boolean isPermissionExists(String domainId, String permissionId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isPermissionExists(domainId, permissionId)) throw new ResourceNotFoundException("Permission Type not found");
            PermissionTypePK permissionTypePK = new PermissionTypePK();
            permissionTypePK.setDomainId(domainId);
            permissionTypePK.setPermissionTypeId(permissionId);
            return (new PermissionTypeRepository()).isExists(permissionTypePK);
        }catch (ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean deletePermissionType(String domainId, String permissionTypeId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isPermissionExists(domainId, permissionTypeId)) throw new ResourceNotFoundException("Permission Type not found");
            PermissionTypePK permissionTypePK =  new PermissionTypePK();
            permissionTypePK.setDomainId(domainId);
            permissionTypePK.setPermissionTypeId(permissionTypeId);
            (new PermissionTypeRepository()).delete(permissionTypePK);
            return true;
        }catch (ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public PermissionType getPermissionType(String domainId, String permissionTypeId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isPermissionExists(domainId, permissionTypeId)) throw new ResourceNotFoundException("Permission Type not found");
            PermissionTypePK permissionTypePK =  new PermissionTypePK();
            permissionTypePK.setDomainId(domainId);
            permissionTypePK.setPermissionTypeId(permissionTypeId);
            return (new PermissionTypeRepository()).get(permissionTypePK);
        }catch (ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public List<PermissionType> getPermissionTypes(String domain, int offset, int limit) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isDomainExists(domain)) throw new ResourceNotFoundException("Could not find the domain");
            HashMap<String, String> filters = new HashMap<>();
            filters.put(DBConstants.PermissionTypeTable.DOMAIN_ID, domain);
            return (new PermissionTypeRepository()).select(filters, offset, limit);
        }catch (ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    /**
     * * Entity Operations
     * *
     */
    public Entity createEntity(Entity entity) throws SharingRegistryException, DuplicateEntryException {
        try{
            EntityPK entityPK = new EntityPK();
            entityPK.setDomainId(entity.getDomainId());
            entityPK.setEntityId(entity.getEntityId());
            if((new EntityRepository()).get(entityPK) != null)
                throw new DuplicateEntryException("There exist Entity with given Entity id");

            UserPK userPK = new UserPK();
            userPK.setDomainId(entity.getDomainId());
            userPK.setUserId(entity.getOwnerId());
            if(!(new UserRepository()).isExists(userPK)){
                //Todo this is for Airavata easy integration. Proper thing is to throw an exception here
                User user = new User();
                user.setUserId(entity.getOwnerId());
                user.setDomainId(entity.getDomainId());
                user.setUserName(user.getUserId().split("@")[0]);

                createUser(user);
            }
            entity.setCreatedTime(System.currentTimeMillis());
            entity.setUpdatedTime(System.currentTimeMillis());

            if(entity.getOriginalEntityCreationTime()==0){
                entity.setOriginalEntityCreationTime(entity.getCreatedTime());
            }
            Entity createdEntity = (new EntityRepository()).create(entity);

            //Assigning global permission for the owner
            Sharing newSharing = new Sharing();
            newSharing.setPermissionTypeId((new PermissionTypeRepository()).getOwnerPermissionTypeIdForDomain(entity.getDomainId()));
            newSharing.setEntityId(entity.getEntityId());
            newSharing.setGroupId(entity.getOwnerId());
            newSharing.setSharingType(SharingType.DIRECT_CASCADING);
            newSharing.setInheritedParentId(entity.getEntityId());
            newSharing.setDomainId(entity.getDomainId());
            newSharing.setCreatedTime(System.currentTimeMillis());
            newSharing.setUpdatedTime(System.currentTimeMillis());

            (new SharingRepository()).create(newSharing);

            // creating records for inherited permissions
            if (entity.getParentEntityId() != null && entity.getParentEntityId() != "") {
                addCascadingPermissionsForEntity(entity);
            }

            return createdEntity;
        }catch(DuplicateEntryException ex) {
            throw ex;
        }catch(Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    private void addCascadingPermissionsForEntity(Entity entity) throws SharingRegistryException {
        Sharing newSharing;
        List<Sharing> sharings = (new SharingRepository()).getCascadingPermissionsForEntity(entity.getDomainId(),
                entity.getParentEntityId());
        for (Sharing sharing : sharings) {
                    newSharing = new Sharing();
                    newSharing.setPermissionTypeId(sharing.getPermissionTypeId());
                    newSharing.setEntityId(entity.getEntityId());
                    newSharing.setGroupId(sharing.getGroupId());
                    newSharing.setInheritedParentId(sharing.getInheritedParentId());
                    newSharing.setSharingType(SharingType.INDIRECT_CASCADING);
                    newSharing.setDomainId(entity.getDomainId());
                    newSharing.setCreatedTime(System.currentTimeMillis());
                    newSharing.setUpdatedTime(System.currentTimeMillis());

                    (new SharingRepository()).create(newSharing);
                }
            }

    public boolean updateEntity(Entity entity) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isEntityExists(entity.getDomainId(), entity.getEntityId())) throw new ResourceNotFoundException("Could not find entity with entityId: " + entity.getEntityId());
            //TODO Check for permission changes
            entity.setUpdatedTime(System.currentTimeMillis());
            EntityPK entityPK = new EntityPK();
            entityPK.setDomainId(entity.getDomainId());
            entityPK.setEntityId(entity.getEntityId());
            Entity oldEntity = (new EntityRepository()).get(entityPK);
            entity.setCreatedTime(oldEntity.getCreatedTime());
            // check if parent entity changed and re-add inherited permissions
            if (!Objects.equals(oldEntity.getParentEntityId(), entity.getParentEntityId())) {
                logger.debug("Parent entity changed for {}, updating inherited permissions", entity.getEntityId());
                if (oldEntity.getParentEntityId() != null && oldEntity.getParentEntityId() != "") {
                    logger.debug("Removing inherited permissions from {} that were inherited from parent {}", entity.getEntityId(), oldEntity.getParentEntityId());
                    (new SharingRepository()).removeAllIndirectCascadingPermissionsForEntity(entity.getDomainId(), entity.getEntityId());
                }
                if (entity.getParentEntityId() != null && entity.getParentEntityId() != "") {
                    // re-add INDIRECT_CASCADING permissions
                    logger.debug("Adding inherited permissions to {} that are inherited from parent {}", entity.getEntityId(), entity.getParentEntityId());
                    addCascadingPermissionsForEntity(entity);
                }
            }
            entity = getUpdatedObject(oldEntity, entity);
            entity.setSharedCount((new SharingRepository()).getSharedCount(entity.getDomainId(), entity.getEntityId()));
            Entity updatedEntity = (new EntityRepository()).update(entity);
            if(updatedEntity != null && updatedEntity.getEntityId().equals(entity.getEntityId())) return true;
            throw new SharingRegistryException("Could not update the entity");
        }catch(ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    /**
     * <p>API method to check Entity Exists</p>
     *
     * @param entityId
     */
    public boolean isEntityExists(String domainId, String entityId) throws SharingRegistryException {
        try{
            EntityPK entityPK = new EntityPK();
            entityPK.setDomainId(domainId);
            entityPK.setEntityId(entityId);
            return (new EntityRepository()).isExists(entityPK);
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean deleteEntity(String domainId, String entityId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isEntityExists(domainId, entityId)) throw new ResourceNotFoundException("Could not find entity with entityId:" + entityId + " in domain:" + domainId);
            //TODO Check for permission changes
            EntityPK entityPK = new EntityPK();
            entityPK.setDomainId(domainId);
            entityPK.setEntityId(entityId);
            (new EntityRepository()).delete(entityPK);
            return true;
        }catch(ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public Entity getEntity(String domainId, String entityId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isEntityExists(domainId, entityId)) throw new ResourceNotFoundException("Could not find entity with entityId:" + entityId + " in domain:" + domainId);
            EntityPK entityPK = new EntityPK();
            entityPK.setDomainId(domainId);
            entityPK.setEntityId(entityId);
            return (new EntityRepository()).get(entityPK);
        }catch(ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public List<Entity> searchEntities(String domainId, String userId, List<SearchCriteria> filters,
                                       int offset, int limit) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isUserExists(domainId, userId)) throw new ResourceNotFoundException("Could not find user:" + userId + " in domain:" + domainId);
            List<String> groupIds = new ArrayList<>();
            groupIds.add(userId);
            (new GroupMembershipRepository()).getAllParentMembershipsForChild(domainId, userId).stream().forEach(gm -> groupIds.add(gm.getParentId()));
            return (new EntityRepository()).searchEntities(domainId, groupIds, filters, offset, limit);
        }catch(ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public List<User> getListOfSharedUsers(String domainId, String entityId, String permissionTypeId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isEntityExists(domainId, entityId) || !isPermissionExists(domainId, permissionTypeId)) throw new ResourceNotFoundException("Could not find the resource");
            return (new UserRepository()).getAccessibleUsers(domainId, entityId, permissionTypeId);
        }catch(ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public List<User> getListOfDirectlySharedUsers(String domainId, String entityId, String permissionTypeId)
            throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isEntityExists(domainId, entityId) || !isPermissionExists(domainId, permissionTypeId)) throw new ResourceNotFoundException("Could not find the resource");
            return (new UserRepository()).getDirectlyAccessibleUsers(domainId, entityId, permissionTypeId);
        }catch(ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public List<UserGroup> getListOfSharedGroups(String domainId, String entityId, String permissionTypeId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isEntityExists(domainId, entityId) || !isPermissionExists(domainId, permissionTypeId)) throw new ResourceNotFoundException("Could not find the resource");
            return (new UserGroupRepository()).getAccessibleGroups(domainId, entityId, permissionTypeId);
        }catch(ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public List<UserGroup> getListOfDirectlySharedGroups(String domainId, String entityId, String permissionTypeId)
            throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isEntityExists(domainId, entityId) || !isPermissionExists(domainId, permissionTypeId)) throw new ResourceNotFoundException("Could not find the resource");
            return (new UserGroupRepository()).getDirectlyAccessibleGroups(domainId, entityId, permissionTypeId);
        }catch(ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    /**
     * Sharing Entity with Users and Groups
     * @param domainId
     * @param entityId
     * @param userList
     * @param permissionTypeId
     * @param cascadePermission
     * @return
     * @throws SharingRegistryException
     * @throws ResourceNotFoundException
     */

    public boolean shareEntityWithUsers(String domainId, String entityId, List<String> userList, String permissionTypeId, boolean cascadePermission) throws SharingRegistryException, ResourceNotFoundException {
        try{
            for(String user: userList){
                if(!isUserExists(domainId, user)) throw new ResourceNotFoundException("Could not find the user");
            }
            if(!isDomainExists(domainId) || !isEntityExists(domainId, entityId) || !isPermissionExists(domainId, permissionTypeId)) throw new ResourceNotFoundException("Could not find the user");
            return shareEntity(domainId, entityId, userList, permissionTypeId, cascadePermission);
        }catch(ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean shareEntityWithGroups(String domainId, String entityId, List<String> groupList, String permissionTypeId, boolean cascadePermission) throws SharingRegistryException, ResourceNotFoundException {
        try{
            for(String group: groupList){
                if(!isGroupExists(domainId, group)) throw new ResourceNotFoundException("Could not find the group");
            }
            if(!isDomainExists(domainId) || !isEntityExists(domainId, entityId) || !isPermissionExists(domainId, permissionTypeId)) throw new ResourceNotFoundException("Could not find the user");
            return shareEntity(domainId, entityId, groupList, permissionTypeId, cascadePermission);
        }catch(ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    private boolean shareEntity(String domainId, String entityId, List<String> groupOrUserList, String permissionTypeId, boolean cascadePermission)  throws SharingRegistryException, ResourceNotFoundException {
        try{
            for(String groupOrUser: groupOrUserList){
                if(!isUserExists(domainId, groupOrUser) && !isGroupExists(domainId, groupOrUser)) throw new ResourceNotFoundException("Could not find the user");
            }
            if(!isDomainExists(domainId) || !isEntityExists(domainId, entityId) || !isPermissionExists(domainId, permissionTypeId)) throw new ResourceNotFoundException("Could not find the user");
            if(permissionTypeId.equals((new PermissionTypeRepository()).getOwnerPermissionTypeIdForDomain(domainId))){
                throw new SharingRegistryException(OWNER_PERMISSION_NAME + " permission cannot be assigned or removed");
            }

            List<Sharing> sharings = new ArrayList<>();

            //Adding permission for the specified users/groups for the specified entity
            LinkedList<Entity> temp = new LinkedList<>();
            for(String userId : groupOrUserList){
                Sharing sharing = new Sharing();
                sharing.setPermissionTypeId(permissionTypeId);
                sharing.setEntityId(entityId);
                sharing.setGroupId(userId);
                sharing.setInheritedParentId(entityId);
                sharing.setDomainId(domainId);
                if(cascadePermission) {
                    sharing.setSharingType(SharingType.DIRECT_CASCADING);
                }else {
                    sharing.setSharingType(SharingType.DIRECT_NON_CASCADING);
                }
                sharing.setCreatedTime(System.currentTimeMillis());
                sharing.setUpdatedTime(System.currentTimeMillis());

                sharings.add(sharing);
            }

            if(cascadePermission){
                //Adding permission for the specified users/groups for all child entities
                (new EntityRepository()).getChildEntities(domainId, entityId).stream().forEach(e -> temp.addLast(e));
                while(temp.size() > 0){
                    Entity entity = temp.pop();
                    String childEntityId = entity.getEntityId();
                    for(String userId : groupOrUserList){
                        Sharing sharing = new Sharing();
                        sharing.setPermissionTypeId(permissionTypeId);
                        sharing.setEntityId(childEntityId);
                        sharing.setGroupId(userId);
                        sharing.setInheritedParentId(entityId);
                        sharing.setSharingType(SharingType.INDIRECT_CASCADING);
                        sharing.setInheritedParentId(entityId);
                        sharing.setDomainId(domainId);
                        sharing.setCreatedTime(System.currentTimeMillis());
                        sharing.setUpdatedTime(System.currentTimeMillis());
                        sharings.add(sharing);
                        (new EntityRepository()).getChildEntities(domainId, childEntityId).stream().forEach(e -> temp.addLast(e));
                    }
                }
            }
            (new SharingRepository()).create(sharings);

            EntityPK entityPK = new EntityPK();
            entityPK.setDomainId(domainId);
            entityPK.setEntityId(entityId);
            Entity entity = (new EntityRepository()).get(entityPK);
            entity.setSharedCount((new SharingRepository()).getSharedCount(domainId, entityId));
            (new EntityRepository()).update(entity);
            return true;
        }catch(ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean revokeEntitySharingFromUsers(String domainId, String entityId, List<String> userList, String permissionTypeId) throws SharingRegistryException, ResourceNotFoundException, InvalidRequestException {
        try{
            for(String user: userList){
                if(!isUserExists(domainId, user)) throw new ResourceNotFoundException("Could not find the user");
            }
            if(!isDomainExists(domainId) || !isEntityExists(domainId, entityId) || !isPermissionExists(domainId, permissionTypeId)) throw new ResourceNotFoundException("Could not find the user");
            if(permissionTypeId.equals((new PermissionTypeRepository()).getOwnerPermissionTypeIdForDomain(domainId))){
                throw new InvalidRequestException(OWNER_PERMISSION_NAME + " permission cannot be assigned or removed");
            }
            return revokeEntitySharing(domainId, entityId, userList, permissionTypeId);
        }catch (InvalidRequestException | ResourceNotFoundException ex) {
            throw ex;
        }catch(Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }


    public boolean revokeEntitySharingFromGroups(String domainId, String entityId, List<String> groupList, String permissionTypeId) throws SharingRegistryException, ResourceNotFoundException, InvalidRequestException {
        try{
            for(String group: groupList){
                if(!isGroupExists(domainId, group)) throw new ResourceNotFoundException("Could not find the group");
            }
            if(!isDomainExists(domainId) || !isEntityExists(domainId, entityId) || !isPermissionExists(domainId, permissionTypeId)) throw new ResourceNotFoundException("Could not find the user");
            if(permissionTypeId.equals((new PermissionTypeRepository()).getOwnerPermissionTypeIdForDomain(domainId))){
                throw new InvalidRequestException(OWNER_PERMISSION_NAME + " permission cannot be assigned or removed");
            }
            return revokeEntitySharing(domainId, entityId, groupList, permissionTypeId);
        }catch (InvalidRequestException | ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }


    public boolean userHasAccess(String domainId, String userId, String entityId, String permissionTypeId) throws SharingRegistryException, ResourceNotFoundException {
        try{
            if(!isDomainExists(domainId) || !isEntityExists(domainId, entityId) || !isPermissionExists(domainId, permissionTypeId) || !isUserExists(domainId, userId)) throw new ResourceNotFoundException("Could not find the resource");
            //check whether the user has permission directly or indirectly
            List<GroupMembership> parentMemberships = (new GroupMembershipRepository()).getAllParentMembershipsForChild(domainId, userId);
            List<String> groupIds = new ArrayList<>();
            parentMemberships.stream().forEach(pm->groupIds.add(pm.getParentId()));
            groupIds.add(userId);
            return (new SharingRepository()).hasAccess(domainId, entityId, groupIds, Arrays.asList(permissionTypeId,
                    (new PermissionTypeRepository()).getOwnerPermissionTypeIdForDomain(domainId)));
        }catch (ResourceNotFoundException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }
    }

    public boolean revokeEntitySharing(String domainId, String entityId, List<String> groupOrUserList, String permissionTypeId) throws SharingRegistryException, InvalidRequestException, ResourceNotFoundException {
        try{
            for(String groupOrUser: groupOrUserList){
                if(!isUserExists(domainId, groupOrUser) && !isGroupExists(domainId, groupOrUser)) throw new ResourceNotFoundException("Could not find the user");
            }
            if(!isDomainExists(domainId) || !isEntityExists(domainId, entityId) || !isPermissionExists(domainId, permissionTypeId)) throw new ResourceNotFoundException("Could not find the user");
            if(permissionTypeId.equals((new PermissionTypeRepository()).getOwnerPermissionTypeIdForDomain(domainId))){
                throw new InvalidRequestException(OWNER_PERMISSION_NAME + " permission cannot be removed");
            }

            //revoking permission for the entity
            for(String groupId : groupOrUserList){
                SharingPK sharingPK = new SharingPK();
                sharingPK.setEntityId(entityId);
                sharingPK.setGroupId(groupId);
                sharingPK.setPermissionTypeId(permissionTypeId);
                sharingPK.setInheritedParentId(entityId);
                sharingPK.setDomainId(domainId);

                (new SharingRepository()).delete(sharingPK);
            }

            //revoking permission from inheritance
            List<Sharing> temp = new ArrayList<>();
            (new SharingRepository()).getIndirectSharedChildren(domainId, entityId, permissionTypeId).stream().forEach(s -> temp.add(s));
            for(Sharing sharing : temp){
                String childEntityId = sharing.getEntityId();
                for(String groupId : groupOrUserList){
                    SharingPK sharingPK = new SharingPK();
                    sharingPK.setEntityId(childEntityId);
                    sharingPK.setGroupId(groupId);
                    sharingPK.setPermissionTypeId(permissionTypeId);
                    sharingPK.setInheritedParentId(entityId);
                    sharingPK.setDomainId(domainId);

                    (new SharingRepository()).delete(sharingPK);
                }
            }

            EntityPK entityPK = new EntityPK();
            entityPK.setDomainId(domainId);
            entityPK.setEntityId(entityId);
            Entity entity = (new EntityRepository()).get(entityPK);
            entity.setSharedCount((new SharingRepository()).getSharedCount(domainId, entityId));
            (new EntityRepository()).update(entity);
            return true;
        }catch (ResourceNotFoundException | InvalidRequestException ex) {
            throw ex;
        }catch (Throwable ex) {
            logger.error(ex.getMessage(), ex);
            throw new SharingRegistryException(ex.getMessage() + " Stack trace:" + ExceptionUtils.getStackTrace(ex));
        }

    }



    private <T> T getUpdatedObject(T oldEntity, T newEntity) throws SharingRegistryException {
        Field[] newEntityFields = newEntity.getClass().getDeclaredFields();
        Hashtable newHT = fieldsToHT(newEntityFields, newEntity);

        Class oldEntityClass = oldEntity.getClass();
        Field[] oldEntityFields = oldEntityClass.getDeclaredFields();

        for (Field field : oldEntityFields){
            if (!Modifier.isFinal(field.getModifiers())) {
                field.setAccessible(true);
                Object o = newHT.get(field.getName());
                if (o != null) {
                    Field f = null;
                    try {
                        f = oldEntityClass.getDeclaredField(field.getName());
                        f.setAccessible(true);
                        logger.debug("setting " + f.getName());
                        f.set(oldEntity, o);
                    } catch (Exception e) {
                        throw new SharingRegistryException(e.getMessage());
                    }
                }
            }
        }
        return oldEntity;
    }

    private static Hashtable<String, Object> fieldsToHT(Field[] fields, Object obj){
        Hashtable<String,Object> hashtable = new Hashtable<>();
        for (Field field: fields){
            field.setAccessible(true);
            try {
                Object retrievedObject = field.get(obj);
                if (retrievedObject != null){
                    logger.debug("scanning " + field.getName());
                    hashtable.put(field.getName(), field.get(obj));
                }
            } catch (IllegalAccessException e) {
                e.printStackTrace();
            }
        }
        return hashtable;
    }

}
