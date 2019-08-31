package org.apache.custos.sharing.service.api.controllers;

import org.apache.custos.sharing.service.core.models.Entity;
import org.apache.custos.sharing.service.core.models.SearchCriteria;
import org.apache.custos.sharing.service.core.models.User;
import org.apache.custos.sharing.service.core.models.UserGroup;
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
@RequestMapping("/entity")
public class EntityController {

    @Autowired
    private SharingRegistryService sharingRegistryService;

    /**
     <p>API method to register new entity</p>
     */
    @RequestMapping(value = "/create", method = RequestMethod.POST)
    public ResponseEntity<String> createEntity(@RequestBody Entity entity){
        String createdEntityId = sharingRegistryService.createEntity(entity);
        if(createdEntityId.equals(entity.getEntityId())){
            return new ResponseEntity<>(HttpStatus.OK);
        }
        else throw new SharingRegistryException("Failed to create the entity");
    }

    /**
     <p>API method to update entity</p>
     */
    @RequestMapping(value = "/update", method = RequestMethod.PUT)
    public ResponseEntity<String> updateEntity(@RequestBody Entity entity){
        boolean response = sharingRegistryService.updateEntity(entity);
        if(response) return new ResponseEntity<>(HttpStatus.OK);
        else throw new SharingRegistryException("Failed to update the entity");
    }

    /**
     <p>API method to check Entity Exists</p>
     */
    @RequestMapping(value = "/exists/id/{entityId}/domain/{domainId}", method = RequestMethod.GET)
    public @ResponseBody Map<String, Boolean> isEntityExists(@PathVariable("entityId") String entityId, @PathVariable("domainId") String domainId){
        boolean response = sharingRegistryService.isEntityExists(entityId, domainId);
        return Collections.singletonMap("exists", response);
    }

    /**
     <p>API method to delete entity</p>
     */
    @RequestMapping(value = "/id/{entityId}/domain/{domainId}", method = RequestMethod.DELETE)
    public ResponseEntity<String> deleteEntity(@PathVariable("entityId") String entityId, @PathVariable("domainId") String domainId){
        boolean response = sharingRegistryService.deleteEntity(domainId, entityId);
        if(response) return new ResponseEntity<>(HttpStatus.OK);
        else throw new SharingRegistryException("Could not delete the entity");
    }

    /**
     <p>API method to get entity</p>
     */
    @RequestMapping(value = "/id/{entityId}/domain/{domainId}", method = RequestMethod.GET)
    public @ResponseBody Entity getEntity(@PathVariable("entityId") String entityId, @PathVariable("domainId") String domainId){
        Entity entity = sharingRegistryService.getEntity(domainId, entityId);
        if(entity != null) return entity;
        //TODO: not found exception should be thrown??
        else throw new SharingRegistryException("Could not retrieve the entity");
    }

    /**
     <p>API method to search entities</p>
     */
    @RequestMapping(value = "/search/domain/{domainId}/user/{userId}/offset/{offset}/limit/{limit}", method = RequestMethod.POST)
    public @ResponseBody List<Entity> searchEntities(@RequestBody List<SearchCriteria> filters, @PathVariable("domainId") String domainId, @PathVariable("userId") String userId, @PathVariable("offset") int offset, @PathVariable("limit") int limit){
        List<Entity> entities = sharingRegistryService.searchEntities(domainId, userId, filters, offset, limit);
        return entities;
    }

    /**
     <p>API method to get a list of shared users given the entity id</p>
     */
    @RequestMapping(value = "/sharedUsers/id/{entityId}/domain/{domainId}/permissionType/{permissionTypeId}", method = RequestMethod.GET)
    public @ResponseBody List<User> getListOfSharedUsers(@PathVariable("entityId") String entityId, @PathVariable("domainId") String domainId, @PathVariable("permissionTypeId") String permissionTypeId){
        return sharingRegistryService.getListOfSharedUsers(domainId, entityId, permissionTypeId);
    }

    /**
     <p>API method to get a list of shared users given the entity id where the sharing type is directly applied</p>
     */
    @RequestMapping(value = "/directlySharedUsers/id/{entityId}/domain/{domainId}/permissionType/{permissionTypeId}", method = RequestMethod.GET)
    public @ResponseBody List<User> getListOfDirectlySharedUsers(@PathVariable("entityId") String entityId, @PathVariable("domainId") String domainId, @PathVariable("permissionTypeId") String permissionTypeId){
        return sharingRegistryService.getListOfDirectlySharedUsers(domainId, entityId, permissionTypeId);
    }

    /**
     <p>API method to get a list of shared groups given the entity id</p>
     */
    @RequestMapping(value = "/sharedGroups/id/{entityId}/domain/{domainId}/permissionType/{permissionTypeId}", method = RequestMethod.GET)
    public @ResponseBody List<UserGroup> getListOfSharedGroups(@PathVariable("entityId") String entityId, @PathVariable("domainId") String domainId, @PathVariable("permissionTypeId") String permissionTypeId){
        return sharingRegistryService.getListOfSharedGroups(domainId, entityId, permissionTypeId);
    }

    /**
     <p>API method to get a list of directly shared groups given the entity id where the sharing type is directly applied</p>
     */
    @RequestMapping(value = "/directlySharedGroups/id/{entityId}/domain/{domainId}/permissionType/{permissionTypeId}", method = RequestMethod.GET)
    public @ResponseBody List<UserGroup> getListOfDirectlySharedGroups(@PathVariable("entityId") String entityId, @PathVariable("domainId") String domainId, @PathVariable("permissionTypeId") String permissionTypeId){
        return sharingRegistryService.getListOfDirectlySharedGroups(domainId, entityId, permissionTypeId);
    }
}
