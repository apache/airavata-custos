package org.apache.custos.sharing.service.api.controllers;

import org.apache.custos.sharing.service.core.models.PermissionType;
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
@RequestMapping("/permissionType")
public class PermissionTypeController {

    @Autowired
    private static SharingRegistryService sharingRegistryService;
    /**
     <p>API method to create permission type</p>
     */
    @RequestMapping(value= "/create", method = RequestMethod.POST)
    public ResponseEntity<String> createPermissionType(@RequestBody PermissionType permissionType){
        String permissionTypeId = sharingRegistryService.createPermissionType(permissionType);
        if(permissionTypeId.equals(permissionType.getPermissionTypeId())){
            return new ResponseEntity<>(HttpStatus.OK);
        }else{
            throw new SharingRegistryException("Could not create the permission type");
        }
    }

    /**
     <p>API method to update permission type</p>
     */
    @RequestMapping(value= "/update", method = RequestMethod.PUT)
    public ResponseEntity<String> updatePermissionType(@RequestBody PermissionType permissionType){

        boolean response = sharingRegistryService.updatePermissionType(permissionType);
        if(response){
            return new ResponseEntity<>(HttpStatus.OK);
        }else{
            throw new SharingRegistryException("Could not update the permission type");
        }
    }

    /**
     <p>API method to check Permission Exists</p>
     */
    @RequestMapping(value= "/exists/id/{permissionId}/domain/{domainId}", method = RequestMethod.GET)
    public @ResponseBody
    Map<String, Boolean> isPermissionExists(@PathVariable("permissionId") String permissionId, @PathVariable("domainId") String domainId){
        boolean response = sharingRegistryService.isPermissionExists(domainId, permissionId);
        return Collections.singletonMap("exists", response);
    }

    /**
     <p>API method to delete permission type</p>
     */
    @RequestMapping(value= "/id/{permissionTypeId}/domain/{domainId}", method = RequestMethod.DELETE)
    public ResponseEntity<String> deletePermissionType(@PathVariable("permissionTypeId") String permissionTypeId, @PathVariable("domainId") String domainId){
        boolean response = sharingRegistryService.deletePermissionType(domainId,permissionTypeId);
        if(response){
            return new ResponseEntity<>(HttpStatus.OK);
        }else{
            throw new SharingRegistryException("Could not delete the permission type");
        }
    }

    /**
     <p>API method to get permission type</p>
     */
    @RequestMapping(value= "/id/{permissionTypeId}/domain/{domainId}", method = RequestMethod.GET)
    public @ResponseBody PermissionType getPermissionType(@PathVariable("permissionTypeId") String permissionTypeId, @PathVariable("domainId") String domainId){
        PermissionType permissionType = sharingRegistryService.getPermissionType(domainId, permissionTypeId);
        if(permissionType != null){
            return permissionType;
        }else{
            throw new ResourceNotFoundException("Could not find the permission type with permission Id: "+ permissionTypeId + " in domainId: " + domainId);
        }
    }

    /**
     <p>API method to get list of permission types in a given domainId.</p>
     */
    @RequestMapping(value= "/domain/{domainId}/offset/{offset}/limit/{limit}", method = RequestMethod.GET)
    public @ResponseBody
    List<PermissionType> getPermissionTypes(@PathVariable("domainId") String domainId, @PathVariable("offset") int offset, @PathVariable("limit") int limit){
        return sharingRegistryService.getPermissionTypes(domainId, offset, limit);
    }

}
