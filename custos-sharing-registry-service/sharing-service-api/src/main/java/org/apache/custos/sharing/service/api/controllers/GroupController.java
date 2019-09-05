package org.apache.custos.sharing.service.api.controllers;

import org.apache.custos.sharing.service.core.models.User;
import org.apache.custos.sharing.service.core.models.UserGroup;
import org.apache.custos.sharing.service.core.service.SharingRegistryService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.List;

@RestController
@RequestMapping("/domains/{domainId}/groups")
public class GroupController {

    @Autowired
    SharingRegistryService sharingRegistryService;

    /**
     <p>API method to create a new group</p>
     */
    @RequestMapping(method = RequestMethod.POST)
    public ResponseEntity<UserGroup> createGroup(@Valid @RequestBody UserGroup group){
        UserGroup createdUserGroup = sharingRegistryService.createGroup(group);
        return new ResponseEntity<>(createdUserGroup, HttpStatus.CREATED);
    }

    /**
     <p>API method to update a group</p>
     */
    @RequestMapping(method = RequestMethod.PUT)
    public ResponseEntity<String> updateGroup(@Valid @RequestBody UserGroup group){
        sharingRegistryService.updateGroup(group);
        return new ResponseEntity<>(HttpStatus.OK);
    }

    /**
     <p>API method to check Group Exists</p>
     */
    @RequestMapping(value = "/{groupId}", method = RequestMethod.HEAD)
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
    @RequestMapping(value = "/{groupId}", method = RequestMethod.DELETE)
    public ResponseEntity<String> deleteGroup(@PathVariable("domainId") String domainId, @PathVariable("groupId") String groupId){
        sharingRegistryService.deleteGroup(domainId, groupId);
        return new ResponseEntity<>(HttpStatus.OK);
    }

    /**
     <p>API method to get a group</p>
     */
    @RequestMapping(value = "/{groupId}", method = RequestMethod.GET)
    public ResponseEntity<UserGroup> getGroup(@PathVariable("groupId") String groupId, @PathVariable("domainId") String domainId){
        UserGroup userGroup = sharingRegistryService.getGroup(domainId,groupId);
        return new ResponseEntity<>(userGroup,HttpStatus.OK);
    }

    /**
     <p>API method to get groups in a domainId.</p>
     */
    @RequestMapping(value = "/groups",method = RequestMethod.GET)
    public @ResponseBody List<UserGroup> getGroups(@PathVariable("domainId") String domainId, @RequestParam int offset, @RequestParam int limit){
        return sharingRegistryService.getGroups(domainId, offset, limit);
    }

    /**
     <p>API method to add list of users to a group</p>
     */
    @RequestMapping(value = "/{groupId}/users", method = RequestMethod.POST)
    public ResponseEntity<String> addUsersToGroup(@PathVariable("groupId") String groupId, @PathVariable("domainId") String domainId, @RequestParam("ids") List<String> userIds){
        sharingRegistryService.addUsersToGroup(domainId, userIds, groupId);
        return new ResponseEntity<>(HttpStatus.OK);
    }

    /**
     <p>API method to remove users from a group</p>
     */
    @RequestMapping(value = "/{groupId}/users", method = RequestMethod.DELETE)
    public ResponseEntity<String> removeUsersFromGroup(@PathVariable("groupId") String groupId, @PathVariable("domainId") String domainId, @RequestParam("ids") List<String> userIds){
        sharingRegistryService.removeUsersFromGroup(domainId, userIds, groupId);
        return new ResponseEntity<>(HttpStatus.OK);
    }

    /**
     <p>API method to transfer group ownership</p>
     */
    @RequestMapping(value = "/{groupId}/owners/{ownerId}", method = RequestMethod.PUT)
    public ResponseEntity<String> transferGroupOwnership(@PathVariable("domainId") String domainId,@PathVariable("groupId") String groupId, @PathVariable("ownerId") String newOwnerId){
        sharingRegistryService.transferGroupOwnership(domainId,groupId, newOwnerId);
        return new ResponseEntity<>(HttpStatus.OK);
    }

    /**
     <p>API method to add Admin for a group</p>
     */
    @RequestMapping(value = "/{groupId}/admins", method = RequestMethod.POST)
    public ResponseEntity<String> addGroupAdmins(@PathVariable("domainId") String domainId,@PathVariable("groupId") String groupId, @RequestParam("adminIds") List<String> admins){
        sharingRegistryService.addGroupAdmins(domainId,groupId, admins);
        return new ResponseEntity<>(HttpStatus.OK);
    }

    /**
     <p>API method to remove Admin for a group</p>
     */
    @RequestMapping(value = "/{groupId}/admins", method = RequestMethod.DELETE)
    public ResponseEntity<String> removeGroupAdmins(@PathVariable("domainId") String domainId,@PathVariable("groupId") String groupId, @RequestParam("adminIds") List<String> adminIds){
        sharingRegistryService.removeGroupAdmins(domainId,groupId, adminIds);
        return new ResponseEntity<>(HttpStatus.OK);
    }

    /**
     <p>API method to check whether the user has Admin access for the group</p>
     */
    @RequestMapping(value = "/{groupId}/admins/{adminId}", method = RequestMethod.HEAD)
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
    @RequestMapping(value = "/{groupId}/owners/{ownerId}", method = RequestMethod.HEAD)
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
    @RequestMapping(value = "/{groupId}/users", method = RequestMethod.GET)
    public @ResponseBody List<User> getGroupMembersOfTypeUser(@PathVariable("domainId") String domainId, @PathVariable("groupId") String groupId, @RequestParam(name = "offset", defaultValue = "0") int offset, @RequestParam(value = "limit", defaultValue = "-1") int limit){
        return sharingRegistryService.getGroupMembersOfTypeUser(domainId, groupId, offset, limit);
    }

    /**
     <p>API method to get list of child groups in a group. Only the direct members will be returned.</p>
     */
    @RequestMapping(value = "/{groupId}/subgroups", method = RequestMethod.GET)
    public @ResponseBody List<UserGroup> getGroupMembersOfTypeGroup(@PathVariable("domainId") String domainId, @PathVariable("groupId") String groupId, @RequestParam(name= "offset", defaultValue = "0") int offset, @RequestParam(value = "limit", defaultValue = "-1") int limit){
        return sharingRegistryService.getGroupMembersOfTypeGroup(domainId, groupId, offset, limit);
    }

    /**
     <p>API method to add a child group to a parent group.</p>
     */
    @RequestMapping(value = "/{groupId}/subgroups", method = RequestMethod.POST)
    public ResponseEntity<String> addChildGroupsToParentGroup(@PathVariable("domainId") String domainId, @PathVariable("groupId") String groupId, @RequestParam("subgroupIds") List<String> childIds){
        sharingRegistryService.addChildGroupsToParentGroup(domainId, childIds, groupId);
        return new ResponseEntity<>(HttpStatus.OK);
    }

    /**
     <p>API method to remove a child group from parent group.</p>
     */
    @RequestMapping(value = "/{groupId}/subgroups/{childGroupId}", method = RequestMethod.DELETE)
    public ResponseEntity<String> removeChildGroupFromParentGroup(@PathVariable("domainId") String domainId, @PathVariable("groupId") String groupId, @PathVariable("childGroupId") String childGroupId){
        sharingRegistryService.removeChildGroupFromParentGroup(domainId, childGroupId, groupId);
        return new ResponseEntity<>(HttpStatus.OK);
    }

    /**
     <p>API method to get all groups user is a member of.</p>
     */
    @RequestMapping(value = "/users/{userId}", method = RequestMethod.GET)
    public @ResponseBody List<UserGroup> getAllMemberGroupsForUser(@PathVariable("domainId") String domainId, @PathVariable("userId") String userId){
        return sharingRegistryService.getAllMemberGroupsForUser(domainId, userId);
    }
}
