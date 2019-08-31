package org.apache.group.profile.service.api.controllers;

import org.apache.custos.profile.service.core.models.GroupModel;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/group")
public class GroupServiceController {

    @RequestMapping(method = RequestMethod.POST)
    public ResponseEntity<String> createGroup(@RequestBody GroupModel groupModel){
        return null;
    }

    @RequestMapping(method = RequestMethod.PUT)
    public ResponseEntity<String> updateGroup(@RequestBody GroupModel groupModel){
        return null;
    }

    @RequestMapping(value = "/id/{groupId}/owner/{ownerId}", method = RequestMethod.DELETE)
    public ResponseEntity<String> deleteGroup(@PathVariable("groupId") String groupId, @PathVariable("ownerId") String ownerId){
        return null;
    }

    @RequestMapping(value = "/id/{groupId}", method = RequestMethod.GET)
    public ResponseEntity<String> getGroup(@PathVariable("groupId") String groupId){
        return null;
    }

    @RequestMapping(value = "/all", method = RequestMethod.GET)
    public @ResponseBody
    List<GroupModel> getGroups(){
        return null;
    }

    @RequestMapping(value = "/username/{username}", method = RequestMethod.GET)
    public @ResponseBody
    List<GroupModel> getAllGroupsUserBelongs(@PathVariable("username") String username){
        return null;
    }

    @RequestMapping(value = "/id/{groupId}", method = RequestMethod.PUT)
    public ResponseEntity<String> addUsersToGroup(@PathVariable("groupId") String groupId, @RequestBody List<String> userIds){
        return null;
    }

    @RequestMapping(value = "/id/{groupId}", method = RequestMethod.PUT)
    public ResponseEntity<String> removeUsersFromGroup(@PathVariable("groupId") String groupId, @RequestBody List<String> userIds){
        return null;
    }

    @RequestMapping(value = "/id/{groupId}/owner/{ownerId}", method = RequestMethod.PUT)
    public ResponseEntity<String> transferGroupOwnership(@PathVariable("groupId") String groupId, @PathVariable("ownerId") String ownerId){
        return null;
    }

    @RequestMapping(value = "addAdmins/id/{groupId}", method = RequestMethod.PUT)
    public ResponseEntity<String> addGroupAdmins(@PathVariable("groupId") String groupId, @RequestBody List<String> adminIds){
        return null;
    }

    @RequestMapping(value = "removeAdmins/id/{groupId}", method = RequestMethod.PUT)
    public ResponseEntity<String> removeGroupAdmins(@PathVariable("groupId") String groupId, @RequestBody List<String> adminIds){
        return null;
    }

    @RequestMapping(value = "/id/{groupId}/admin/{adminId}", method = RequestMethod.GET)
    public @ResponseBody
    Map<String, Boolean> hasAdminAccess(@PathVariable("groupId") String groupId, @PathVariable("adminId") String adminId){
        return null;
    }

    @RequestMapping(value = "/id/{groupId}/owner/{ownerId}", method = RequestMethod.GET)
    public @ResponseBody
    Map<String, Boolean> hasOwnerAccess(@PathVariable("groupId") String groupId, @PathVariable("ownerId") String ownerId){
        return null;
    }
}
