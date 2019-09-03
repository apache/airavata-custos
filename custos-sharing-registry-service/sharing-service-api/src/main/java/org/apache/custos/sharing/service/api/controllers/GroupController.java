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
    public @ResponseBody UserGroup createGroup(@Valid @RequestBody UserGroup group){
        UserGroup createdUserGroup = sharingRegistryService.createGroup(group);
        if(createdUserGroup != null){
            return createdUserGroup;
        }else{
            throw new SharingRegistryException("Failed to create group");
        }
    }

    /**
     <p>API method to update a group</p>
     */
    @RequestMapping(method = RequestMethod.PUT)
    public ResponseEntity<String> updateGroup(@Valid @RequestBody UserGroup group){
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
    @RequestMapping(value = "/{groupId}", method = RequestMethod.GET)
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
    @RequestMapping(value = "/groups",method = RequestMethod.GET)
    public @ResponseBody List<UserGroup> getGroups(@PathVariable("domainId") String domainId, @RequestParam int offset, @RequestParam int limit){
        return sharingRegistryService.getGroups(domainId, offset, limit);
    }

    /**
     <p>API method to add list of users to a group</p>
     */
    @RequestMapping(value = "/{groupId}/users", method = RequestMethod.POST)
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
    @RequestMapping(value = "/{groupId}/users", method = RequestMethod.DELETE)
    public ResponseEntity<String> removeUsersFromGroup(@PathVariable("groupId") String groupId, @PathVariable("domainId") String domainId, @RequestParam("ids") List<String> userIds){
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
    @RequestMapping(value = "/{groupId}/owners/{ownerId}", method = RequestMethod.PUT)
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
    @RequestMapping(value = "/{groupId}/admins", method = RequestMethod.POST)
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
    @RequestMapping(value = "/{groupId}/admins", method = RequestMethod.DELETE)
    public ResponseEntity<String> removeGroupAdmins(@PathVariable("domainId") String domainId,@PathVariable("groupId") String groupId, @RequestParam("ids") List<String> adminIds){
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
    public @ResponseBody List<User> getGroupMembersOfTypeUser(@PathVariable("domainId") String domainId, @PathVariable("groupId") String groupId, @RequestParam("offset") int offset, @RequestParam("limit") int limit){
        return sharingRegistryService.getGroupMembersOfTypeUser(domainId, groupId, offset, limit);
    }

    /**
     <p>API method to get list of child groups in a group. Only the direct members will be returned.</p>
     */
    @RequestMapping(value = "/{groupId}/subgroups", method = RequestMethod.GET)
    public @ResponseBody List<UserGroup> getGroupMembersOfTypeGroup(@PathVariable("domainId") String domainId, @PathVariable("groupId") String groupId, @RequestParam("offset") int offset, @RequestParam("limit") int limit){
        return sharingRegistryService.getGroupMembersOfTypeGroup(domainId, groupId, offset, limit);
    }

    /**
     <p>API method to add a child group to a parent group.</p>
     */
    @RequestMapping(value = "/{groupId}/subgroups", method = RequestMethod.POST)
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
    @RequestMapping(value = "/{groupId}/subgroups/{childGroupId}", method = RequestMethod.DELETE)
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
    @RequestMapping(value = "/users/{userId}", method = RequestMethod.GET)
    public @ResponseBody List<UserGroup> getAllMemberGroupsForUser(@PathVariable("domainId") String domainId, @PathVariable("userId") String userId){
        return sharingRegistryService.getAllMemberGroupsForUser(domainId, userId);
    }

//    public static void main(String[] args) throws Exception{
//        String url = "http://localhost:7070";
//        String r = "/domains/default/groups";
//        UserGroup test = new UserGroup();
//        test.setGroupId("teqq1");
//        test.setDomainId("default");
//        test.setOwnerId("default-admin@default");
//        test.setName("");
//        test.setGroupType(GroupType.USER_LEVEL_GROUP);
//        ServiceRequest serviceRequest = new ServiceRequest();
//        HttpResponse response = serviceRequest.httpPost(url, r, test);
//        System.out.println(response);
//        if(response.getStatusLine().getStatusCode() == org.apache.http.HttpStatus.SC_OK) {
//            BufferedReader rd = new BufferedReader(
//                    new InputStreamReader(response.getEntity().getContent()));
//            StringBuffer result = new StringBuffer();
//            String line = "";
//            while ((line = rd.readLine()) != null) {
//                result.append(line);
//            }
//            JSONObject jsonObject = new JSONObject(result.toString());
//            System.out.println(jsonObject.get("groupId"));
//        }
//    }
}
