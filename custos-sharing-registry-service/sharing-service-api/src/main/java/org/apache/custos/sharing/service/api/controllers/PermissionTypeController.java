package org.apache.custos.sharing.service.api.controllers;

import org.apache.custos.sharing.service.core.models.PermissionType;
import org.apache.custos.sharing.service.core.exceptions.ResourceNotFoundException;
import org.apache.custos.sharing.service.core.exceptions.SharingRegistryException;
import org.apache.custos.sharing.service.core.service.SharingRegistryService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.Collections;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("domain/{domainId}/permissionTypes")
public class PermissionTypeController {

    @Autowired
    SharingRegistryService sharingRegistryService;
    /**
     <p>API method to create permission type</p>
     */
    @RequestMapping(method = RequestMethod.POST)
    public @ResponseBody PermissionType createPermissionType(@Valid @RequestBody PermissionType permissionType){
        PermissionType createdPermissionType = sharingRegistryService.createPermissionType(permissionType);
        if(createdPermissionType != null){
            return createdPermissionType;
        }else{
            throw new SharingRegistryException("Could not create the permission type");
        }
    }

    /**
     <p>API method to update permission type</p>
     */
    @RequestMapping(method = RequestMethod.PUT)
    public ResponseEntity<String> updatePermissionType(@Valid @RequestBody PermissionType permissionType){

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
    @RequestMapping(value= "/{permissionId}", method = RequestMethod.HEAD)
    public ResponseEntity<String> isPermissionExists(@PathVariable("permissionId") String permissionId, @PathVariable("domainId") String domainId){
        boolean response = sharingRegistryService.isPermissionExists(domainId, permissionId);
        HttpStatus status = HttpStatus.NOT_FOUND;
        if(response){
            status = HttpStatus.OK;
        }
        return new ResponseEntity<>(status);
    }

    /**
     <p>API method to delete permission type</p>
     */
    @RequestMapping(value= "/{permissionTypeId}", method = RequestMethod.DELETE)
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
    @RequestMapping(value= "/{permissionTypeId}", method = RequestMethod.GET)
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
    @RequestMapping(method = RequestMethod.GET)
    public @ResponseBody
    List<PermissionType> getPermissionTypes(@PathVariable("domainId") String domainId, @RequestParam("offset") int offset, @RequestParam("limit") int limit){
        return sharingRegistryService.getPermissionTypes(domainId, offset, limit);
    }

}
