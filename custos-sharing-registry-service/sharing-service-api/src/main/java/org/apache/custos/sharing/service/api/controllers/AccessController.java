package org.apache.custos.sharing.service.api.controllers;

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
public class AccessController {

    @Autowired
    SharingRegistryService sharingRegistryService;

    /**
     <p>API method to share an entity with users</p>
     */
    @RequestMapping(value = "share/users/domain/{domainId}/entity/{entityId}/permissionType/{permissionTypeId}/cascadePermission/{cascadePermission}", method = RequestMethod.PUT)
    public ResponseEntity<String> shareEntityWithUsers(@RequestBody List<String> userList, @PathVariable("domainId") String domainId, @PathVariable("entityId") String entityId, @PathVariable("permissionType") String permissionTypeId, @PathVariable("cascadePermission") boolean cascadePermission){

        boolean response = sharingRegistryService.shareEntityWithUsers(domainId, entityId, userList, permissionTypeId, cascadePermission);
        if(response) return new ResponseEntity<>(HttpStatus.OK);
        else throw new SharingRegistryException("Sharing entity with users failed");

    }

    /**
     <p>API method to revoke sharing from a list of users</p>
     */
    @RequestMapping(value = "revoke/users/domain/{domainId}/entity/{entityId}/permissionType/{permissionTypeId}", method = RequestMethod.PUT)
    public ResponseEntity<String> revokeEntitySharingFromUsers(@RequestBody List<String> userList, @PathVariable("domainId") String domainId, @PathVariable("entityId") String entityId, @PathVariable("permissionType") String permissionTypeId){
        boolean response = sharingRegistryService.revokeEntitySharingFromUsers(domainId, entityId, userList, permissionTypeId);
        if(response) return new ResponseEntity<>(HttpStatus.OK);
        else throw new SharingRegistryException("Revoking entity sharing with users failed");
    }

    /**
     <p>API method to share an entity with list of groups</p>
     */
    @RequestMapping(value = "share/groups/domain/{domainId}/entity/{entityId}/permissionType/{permissionTypeId}/cascadePermission/{cascadePermission}", method = RequestMethod.PUT)
    public ResponseEntity<String> shareEntityWithGroups(@RequestBody List<String> groupList, @PathVariable("domainId") String domainId, @PathVariable("entityId") String entityId, @PathVariable("permissionType") String permissionTypeId, @PathVariable("cascadePermission") boolean cascadePermission){
       boolean response = sharingRegistryService.shareEntityWithGroups(domainId,entityId, groupList, permissionTypeId, cascadePermission);
        if(response) return new ResponseEntity<>(HttpStatus.OK);
        else throw new SharingRegistryException("Sharing entity with groups failed");
    }

    /**
     <p>API method to revoke sharing from list of users</p>
     */
    @RequestMapping(value = "revoke/groups/domain/{domainId}/entity/{entityId}/permissionType/{permissionTypeId}", method = RequestMethod.PUT)
    public ResponseEntity<String> revokeEntitySharingFromGroups(@RequestBody List<String> groupList, @PathVariable("domainId") String domainId, @PathVariable("entityId") String entityId, @PathVariable("permissionType") String permissionTypeId){
        boolean response = sharingRegistryService.revokeEntitySharingFromGroups(domainId, entityId, groupList, permissionTypeId);
        if(response) return new ResponseEntity<>(HttpStatus.OK);
        else throw new SharingRegistryException("Revoking entity sharing with groups failed");
    }

    /**
     <p>API method to check whether a user has access to a specific entity</p>
     */
    @RequestMapping(value = "hasaccess/domain/{domainId}/user/{userId}/entity/{entityId}/permissionType/{permissionTypeId}", method = RequestMethod.PUT)
    public @ResponseBody
    Map<String, Boolean> userHasAccess(@PathVariable("domainId") String domainId, @PathVariable("userId") String userId, @PathVariable("entityId") String entityId, @PathVariable("permissionType") String permissionTypeId){
        boolean response = sharingRegistryService.userHasAccess(domainId, userId, entityId, permissionTypeId);
        return Collections.singletonMap("access", response);
    }
}
