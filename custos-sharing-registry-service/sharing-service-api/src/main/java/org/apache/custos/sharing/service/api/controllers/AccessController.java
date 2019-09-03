package org.apache.custos.sharing.service.api.controllers;

import org.apache.custos.sharing.service.core.exceptions.SharingRegistryException;
import org.apache.custos.sharing.service.core.service.SharingRegistryService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import java.util.List;

@RestController
@RequestMapping("/domains/{domainId}/entities/{entityId}/permissionTypes/{permissionTypeId}")
public class AccessController {

    @Autowired
    SharingRegistryService sharingRegistryService;

    /**
     <p>API method to share an entity with users</p>
     */
    @RequestMapping(value = "/users", method = RequestMethod.POST)
    public ResponseEntity<String> shareEntityWithUsers(@RequestParam("ids") List<String> userList, @PathVariable("domainId") String domainId, @PathVariable("entityId") String entityId, @PathVariable("permissionTypeId") String permissionTypeId, @RequestParam("cascadePermission") boolean cascadePermission){

        boolean response = sharingRegistryService.shareEntityWithUsers(domainId, entityId, userList, permissionTypeId, cascadePermission);
        if(response) return new ResponseEntity<>(HttpStatus.OK);
        else throw new SharingRegistryException("Sharing entity with users failed");

    }

    /**
     <p>API method to revoke sharing from a list of users</p>
     */
    @RequestMapping(value = "/users", method = RequestMethod.DELETE)
    public ResponseEntity<String> revokeEntitySharingFromUsers(@RequestParam("ids") List<String> userList, @PathVariable("domainId") String domainId, @PathVariable("entityId") String entityId, @PathVariable("permissionTypeId") String permissionTypeId){
        boolean response = sharingRegistryService.revokeEntitySharingFromUsers(domainId, entityId, userList, permissionTypeId);
        if(response) return new ResponseEntity<>(HttpStatus.OK);
        else throw new SharingRegistryException("Revoking entity sharing with users failed");
    }

    /**
     <p>API method to share an entity with list of groups</p>
     */
    @RequestMapping(value = "/groups", method = RequestMethod.POST)
    public ResponseEntity<String> shareEntityWithGroups(@RequestParam("groupIds") List<String> groupList, @PathVariable("domainId") String domainId, @PathVariable("entityId") String entityId, @PathVariable("permissionTypeId") String permissionTypeId, @RequestParam("cascadePermission") boolean cascadePermission){
       boolean response = sharingRegistryService.shareEntityWithGroups(domainId,entityId, groupList, permissionTypeId, cascadePermission);
        if(response) return new ResponseEntity<>(HttpStatus.OK);
        else throw new SharingRegistryException("Sharing entity with groups failed");
    }

    /**
     <p>API method to revoke sharing from list of users</p>
     */
    @RequestMapping(value = "/groups", method = RequestMethod.DELETE)
    public ResponseEntity<String> revokeEntitySharingFromGroups(@RequestParam("groupIds") List<String> groupList, @PathVariable("domainId") String domainId, @PathVariable("entityId") String entityId, @PathVariable("permissionTypeId") String permissionTypeId){
        boolean response = sharingRegistryService.revokeEntitySharingFromGroups(domainId, entityId, groupList, permissionTypeId);
        if(response) return new ResponseEntity<>(HttpStatus.OK);
        else throw new SharingRegistryException("Revoking entity sharing with groups failed");
    }

    /**
     <p>API method to check whether a user has access to a specific entity</p>
     */
    @RequestMapping(value = "/users/{userId}", method = RequestMethod.HEAD)
    public ResponseEntity<String> userHasAccess(@PathVariable("domainId") String domainId, @PathVariable("userId") String userId, @PathVariable("entityId") String entityId, @PathVariable("permissionTypeId") String permissionTypeId){
        boolean response = sharingRegistryService.userHasAccess(domainId, userId, entityId, permissionTypeId);
        HttpStatus status = HttpStatus.NOT_FOUND;
        if(response){
            status = HttpStatus.OK;
        }
        return new ResponseEntity<>(status);
    }
}
