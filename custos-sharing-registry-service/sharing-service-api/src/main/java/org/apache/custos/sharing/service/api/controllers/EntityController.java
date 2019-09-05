package org.apache.custos.sharing.service.api.controllers;

import org.apache.custos.sharing.service.core.exceptions.ResourceNotFoundException;
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

import javax.validation.Valid;
import java.util.List;

@RestController
@RequestMapping("/domains/{domainId}/entities")
public class EntityController {

    @Autowired
    private SharingRegistryService sharingRegistryService;

    /**
     <p>API method to register new entity</p>
     */
    @RequestMapping(method = RequestMethod.POST)
    public ResponseEntity<Entity> createEntity(@Valid @RequestBody Entity entity){
        Entity createdEntity = sharingRegistryService.createEntity(entity);
        return new ResponseEntity<>(createdEntity, HttpStatus.CREATED);
    }

    /**
     <p>API method to update entity</p>
     */
    @RequestMapping(method = RequestMethod.PUT)
    public ResponseEntity<String> updateEntity(@Valid @RequestBody Entity entity){
        sharingRegistryService.updateEntity(entity);
        return new ResponseEntity<>(HttpStatus.OK);
    }

    /**
     <p>API method to check Entity Exists</p>
     */
    @RequestMapping(value = "/{entityId}", method = RequestMethod.HEAD)
    public ResponseEntity<String> isEntityExists(@PathVariable("entityId") String entityId, @PathVariable("domainId") String domainId){
        boolean response = sharingRegistryService.isEntityExists(entityId, domainId);
        HttpStatus status = HttpStatus.NOT_FOUND;
        if(response){
            status = HttpStatus.OK;
        }
        return new ResponseEntity<>(status);
    }

    /**
     <p>API method to delete entity</p>
     */
    @RequestMapping(value = "/{entityId}", method = RequestMethod.DELETE)
    public ResponseEntity<String> deleteEntity(@PathVariable("entityId") String entityId, @PathVariable("domainId") String domainId){
        sharingRegistryService.deleteEntity(domainId, entityId);
        return new ResponseEntity<>(HttpStatus.OK);
    }

    /**
     <p>API method to get entity</p>
     */
    @RequestMapping(value = "/{entityId}", method = RequestMethod.GET)
    public ResponseEntity<Entity> getEntity(@PathVariable("entityId") String entityId, @PathVariable("domainId") String domainId){
        Entity entity = sharingRegistryService.getEntity(domainId, entityId);
        return new ResponseEntity<>(entity, HttpStatus.OK);
    }

    /**
     <p>API method to search entities</p>
     */
    @RequestMapping(value = "/users/{userId}", method = RequestMethod.GET)
    public @ResponseBody List<Entity> searchEntities(@RequestBody List<SearchCriteria> filters, @PathVariable("domainId") String domainId, @PathVariable("userId") String userId, @RequestParam(name = "offset", defaultValue = "0") int offset, @RequestParam(name = "limit", defaultValue = "-1") int limit){
        return sharingRegistryService.searchEntities(domainId, userId, filters, offset, limit);
    }

    /**
     <p>API method to get a list of shared users given the entity id</p>
     */
    @RequestMapping(value = "/{entityId}/permissionType/{permissionTypeId}/users", method = RequestMethod.GET)
    public @ResponseBody List<User> getListOfSharedUsers(@PathVariable("entityId") String entityId, @PathVariable("domainId") String domainId, @PathVariable("permissionTypeId") String permissionTypeId){
        return sharingRegistryService.getListOfSharedUsers(domainId, entityId, permissionTypeId);
    }

    /**
     <p>API method to get a list of shared users given the entity id where the sharing type is directly applied</p>
     */
    @RequestMapping(value = "/{entityId}/permissionType/{permissionTypeId}/directUsers/", method = RequestMethod.GET)
    public @ResponseBody List<User> getListOfDirectlySharedUsers(@PathVariable("entityId") String entityId, @PathVariable("domainId") String domainId, @PathVariable("permissionTypeId") String permissionTypeId){
        return sharingRegistryService.getListOfDirectlySharedUsers(domainId, entityId, permissionTypeId);
    }

    /**
     <p>API method to get a list of shared groups given the entity id</p>
     */
    @RequestMapping(value = "/{entityId}/permissionType/{permissionTypeId}/groups", method = RequestMethod.GET)
    public @ResponseBody List<UserGroup> getListOfSharedGroups(@PathVariable("entityId") String entityId, @PathVariable("domainId") String domainId, @PathVariable("permissionTypeId") String permissionTypeId){
        return sharingRegistryService.getListOfSharedGroups(domainId, entityId, permissionTypeId);
    }

    /**
     <p>API method to get a list of directly shared groups given the entity id where the sharing type is directly applied</p>
     */
    @RequestMapping(value = "/{entityId}/permissionType/{permissionTypeId}/directGroups", method = RequestMethod.GET)
    public @ResponseBody List<UserGroup> getListOfDirectlySharedGroups(@PathVariable("entityId") String entityId, @PathVariable("domainId") String domainId, @PathVariable("permissionTypeId") String permissionTypeId){
        return sharingRegistryService.getListOfDirectlySharedGroups(domainId, entityId, permissionTypeId);
    }
}
