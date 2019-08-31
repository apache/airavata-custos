package org.apache.custos.sharing.service.api.controllers;

import org.apache.custos.sharing.service.core.models.GroupAdmin;
import org.apache.custos.sharing.service.core.models.User;
import org.apache.custos.sharing.service.core.models.UserGroup;
import org.apache.custos.sharing.service.core.exceptions.ResourceNotFoundException;
import org.apache.custos.sharing.service.core.exceptions.SharingRegistryException;
import org.apache.custos.sharing.service.core.service.SharingRegistryService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Collections;
import java.util.List;
import java.util.Map;

@RestController
public class GroupController {

    @Autowired
    private static SharingRegistryService sharingRegistryService;

    /**
     <p>API method to create a new group</p>
     */
    @RequestMapping(value = "/domains/{domainId}/groups", method = RequestMethod.POST)
    public ResponseEntity<String> createGroup(@RequestBody UserGroup group){
        String createdGroupId = sharingRegistryService.createGroup(group);
        if(createdGroupId.equals(group.getGroupId())){
            return new ResponseEntity<>(HttpStatus.OK);
        }else{
            throw new SharingRegistryException("Failed to create group");
        }
    }

    /**
     <p>API method to update a group</p>
     */
    @RequestMapping(value = "/domains/{domainId}/groups", method = RequestMethod.PUT)
    public ResponseEntity<String> updateGroup(@RequestBody UserGroup group){
        boolean response = sharingRegistryService.updateGroup(group);
        if(response){
            return new ResponseEntity<>(HttpStatus.OK);
        }
        else{
            throw new SharingRegistryException("Failed to update the group");
        }
    }

    /**
     <p>API method to check Group Exists</p>
     */
    @RequestMapping(value = "/domains/{domainId}/groups/{groupId}", method = RequestMethod.HEAD)
    public ResponseEntity<String> isGroupExists(@PathVariable("domainId") String domainId, @PathVariable("groupId") String groupId){
        boolean response = sharingRegistryService.isGroupExists(domainId, groupId);
        HttpStatus status = HttpStatus.NOT_FOUND;
        if(response){
            status = HttpStatus.OK;
        }
        return new ResponseEntity<>(status);
    }

    /**
     <p>API method to delete a group</p>
     */
    @RequestMapping(value = "/domains/{domainId}/groups/{groupId}", method = RequestMethod.DELETE)
    public ResponseEntity<String> deleteGroup(@PathVariable("domainId") String domainId, @PathVariable("groupId") String groupId){
        boolean response = sharingRegistryService.deleteGroup(domainId, groupId);
        if(response){
            return new ResponseEntity<>(HttpStatus.OK);
        }
        else{
            throw new SharingRegistryException("Failed to delete the group");
        }
    }

    /**
     <p>API method to get a group</p>
     */
    @RequestMapping(value = "/domains/{domainId}/groups/{groupId}", method = RequestMethod.GET)
    public @ResponseBody UserGroup getGroup(@PathVariable("groupId") String groupId, @PathVariable("domainId") String domainId){
        UserGroup userGroup = sharingRegistryService.getGroup(domainId,groupId);
        if(userGroup != null){
            return userGroup;
        }else{
            throw new ResourceNotFoundException("Could not find the user group with groupId: "+ groupId + " in domainId: "+ domainId);
        }
    }

    /**
     <p>API method to get groups in a domainId.</p>
     */
    @RequestMapping(value = "/domains/{domainId}/groups",method = RequestMethod.GET)
    public @ResponseBody List<UserGroup> getGroups(@PathVariable("domainId") String domainId, @RequestParam int offset, @RequestParam int limit){
        return sharingRegistryService.getGroups(domainId, offset, limit);
    }

    /**
     <p>API method to add list of users to a group</p>
     */
    @RequestMapping(value = "/domains/{domainId}/groups/{groupId}/users", method = RequestMethod.POST)
    public ResponseEntity<String> addUsersToGroup(@PathVariable("groupId") String groupId, @PathVariable("domainId") String domainId, @RequestBody List<String> userIds){
        boolean response = sharingRegistryService.addUsersToGroup(domainId, userIds, groupId);
        if(response){
            return new ResponseEntity<>(HttpStatus.OK);
        }
        else{
            throw new SharingRegistryException("Failed to add users to the group");
        }
    }

    /**
     <p>API method to remove users from a group</p>
     */
    @RequestMapping(value = "/domains/{domainId}/groups/{groupId}/users", method = RequestMethod.DELETE)
    public ResponseEntity<String> removeUsersFromGroup(@PathVariable("groupId") String groupId, @PathVariable("domainId") String domainId, @RequestBody List<String> userIds){
        boolean response = sharingRegistryService.removeUsersFromGroup(domainId, userIds, groupId);
        if(response){
            return new ResponseEntity<>(HttpStatus.OK);
        }
        else{
            throw new SharingRegistryException("Failed to remove users from the group");
        }
    }

    /**
     <p>API method to transfer group ownership</p>
     */
    @RequestMapping(value = "/domains/{domainId}/groups/{groupId}/owners/{ownerId}", method = RequestMethod.PUT)
    public ResponseEntity<String> transferGroupOwnership(@PathVariable("domainId") String domainId,@PathVariable("groupId") String groupId, @PathVariable("ownerId") String newOwnerId){
        boolean response = sharingRegistryService.transferGroupOwnership(domainId,groupId, newOwnerId);
        if(response){
            return new ResponseEntity<>(HttpStatus.OK);
        }
        else{
            throw new SharingRegistryException("Failed to change group owner");
        }
    }

    /**
     <p>API method to add Admin for a group</p>
     */
    @RequestMapping(value = "/domains/{domainId}/groups/{groupId}/admins", method = RequestMethod.POST)
    public ResponseEntity<String> addGroupAdmins(@PathVariable("domainId") String domainId,@PathVariable("groupId") String groupId, @RequestBody List<GroupAdmin> admins){
        boolean response = sharingRegistryService.addGroupAdmins(domainId,groupId, admins);
        if(response){
            return new ResponseEntity<>(HttpStatus.OK);
        }
        else{
            throw new SharingRegistryException("Failed to add group admins");
        }
    }

    /**
     <p>API method to remove Admin for a group</p>
     */
    @RequestMapping(value = "/domains/{domainId}/groups/{groupId}/admins", method = RequestMethod.DELETE)
    public ResponseEntity<String> removeGroupAdmins(@PathVariable("domainId") String domainId,@PathVariable("groupId") String groupId, @RequestBody List<String> adminIds){
        boolean response = sharingRegistryService.removeGroupAdmins(domainId,groupId, adminIds);
        if(response){
            return new ResponseEntity<>(HttpStatus.OK);
        }
        else{
            throw new SharingRegistryException("Failed to remove group admins");
        }
    }

    /**
     <p>API method to check whether the user has Admin access for the group</p>
     */
    @RequestMapping(value = "/domains/{domainId}/groups/{groupId}/admin/{adminId}", method = RequestMethod.HEAD)
    public ResponseEntity<String> hasAdminAccess(@PathVariable("domainId") String domainId, @PathVariable("groupId") String groupId, @PathVariable("adminId") String adminId){
        boolean response = sharingRegistryService.hasAdminAccess(domainId, groupId, adminId);
        HttpStatus status = HttpStatus.NOT_FOUND;
        if(response){
            status = HttpStatus.OK;
        }
        return new ResponseEntity<>(status);
    }

    /**
     <p>API method to check whether the user has Owner access for the group</p>
     */
    @RequestMapping(value = "/domains/{domainId}/groups/{groupId}/owner/{ownerId}", method = RequestMethod.HEAD)
    public ResponseEntity<String> hasOwnerAccess(@PathVariable("domainId") String domainId, @PathVariable("groupId") String groupId, @PathVariable("ownerId") String ownerId){
        boolean response = sharingRegistryService.hasOwnerAccess(domainId, groupId, ownerId);
        HttpStatus status = HttpStatus.NOT_FOUND;
        if(response){
            status = HttpStatus.OK;
        }
        return new ResponseEntity<>(status);
    }

    /**
     <p>API method to get list of child users in a group. Only the direct members will be returned.</p>
     */
    @RequestMapping(value = "/domains/{domainId}/groups/{groupId}/users", method = RequestMethod.GET)
    public List<User> getGroupMembersOfTypeUser(@PathVariable("domainId") String domainId, @PathVariable("groupId") String groupId, @RequestParam("offset") int offset, @RequestParam("limit") int limit){
        return sharingRegistryService.getGroupMembersOfTypeUser(domainId, groupId, offset, limit);
    }

    /**
     <p>API method to get list of child groups in a group. Only the direct members will be returned.</p>
     */
    @RequestMapping(value = "/domains/{domainId}/groups/{groupId}/subgroups", method = RequestMethod.GET)
    public List<UserGroup> getGroupMembersOfTypeGroup(@PathVariable("domainId") String domainId, @PathVariable("groupId") String groupId, @RequestParam("offset") int offset, @RequestParam("limit") int limit){
        return sharingRegistryService.getGroupMembersOfTypeGroup(domainId, groupId, offset, limit);
    }

    /**
     <p>API method to add a child group to a parent group.</p>
     */
    @RequestMapping(value = "/domains/{domainId}/groups/{groupId}/subgroups", method = RequestMethod.POST)
    public ResponseEntity<String> addChildGroupsToParentGroup(@PathVariable("domainId") String domainId, @PathVariable("groupId") String groupId, @RequestBody List<String> childIds){
        boolean response = sharingRegistryService.addChildGroupsToParentGroup(domainId, childIds, groupId);
        if(response){
            return new ResponseEntity<>(HttpStatus.OK);
        }else{
            throw new SharingRegistryException("Failed to add child groups to parent group");
        }
    }

    /**
     <p>API method to remove a child group from parent group.</p>
     */
    @RequestMapping(value = "/domains/{domainId}/groups/{groupId}/subgroups", method = RequestMethod.DELETE)
    public ResponseEntity<String> removeChildGroupFromParentGroup(@PathVariable("domainId") String domainId, @PathVariable("groupId") String groupId, @PathVariable("childGroupId") String childGroupId){
        boolean response = sharingRegistryService.removeChildGroupFromParentGroup(domainId, childGroupId, groupId);
        if(response){
            return new ResponseEntity<>(HttpStatus.OK);
        }else{
            throw new SharingRegistryException("Failed to remove child groups from parent group");
        }
    }

    /**
     <p>API method to get all groups user is a member of.</p>
     */
    @RequestMapping(value = "/domains/{domainId}/users/{userId}/groups", method = RequestMethod.PUT)
    public @ResponseBody List<UserGroup> getAllMemberGroupsForUser(@PathVariable("domainId") String domainId, @PathVariable("userId") String userId){
        return sharingRegistryService.getAllMemberGroupsForUser(domainId, userId);
    }
}
