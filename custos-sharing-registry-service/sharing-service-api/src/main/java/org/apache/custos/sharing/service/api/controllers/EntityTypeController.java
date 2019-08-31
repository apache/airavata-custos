package org.apache.custos.sharing.service.api.controllers;

import org.apache.custos.sharing.service.core.models.EntityType;
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
@RequestMapping("/entityType")
public class EntityTypeController {

    @Autowired
    private static SharingRegistryService sharingRegistryService;
    /**
     <p>API method to create a new entity type</p>
     */
    @RequestMapping(value = "/create", method = RequestMethod.POST)
    public ResponseEntity<String> createEntityType(@RequestBody EntityType entityType){
        String createdEntityTypeId = sharingRegistryService.createEntityType(entityType);
        if(createdEntityTypeId.equals(entityType.getEntityTypeId())){
            return new ResponseEntity<>(HttpStatus.OK);
        }
        else{
            throw new SharingRegistryException("Failed to create the entity type");
        }
    }

    /**
     <p>API method to update entity type</p>
     */
    @RequestMapping(value = "/update", method = RequestMethod.PUT)
    public ResponseEntity<String> updateEntityType(@RequestBody EntityType entityType){
        boolean response = sharingRegistryService.updateEntityType(entityType);
        if(response){
            return new ResponseEntity<>(HttpStatus.OK);
        }
        else{
            throw new SharingRegistryException("Failed to updated the entity type");
        }
    }

    /**
     <p>API method to check EntityType Exists</p>
     */
    @RequestMapping(value = "/exists/id/{entityTypeId/domain/{domainId}", method = RequestMethod.GET)
    public @ResponseBody Map<String, Boolean> isEntityTypeExists(@PathVariable("entityTypeId") String entityTypeId, @PathVariable("domainId") String domainId){
        boolean response = sharingRegistryService.isEntityTypeExists(domainId, entityTypeId);
        return Collections.singletonMap("exists", response);
    }

    /**
     <p>API method to delete entity type</p>
     */
    @RequestMapping(value = "/id/{entityTypeId}/domain/{domainId}", method = RequestMethod.DELETE)
    public ResponseEntity<String> deleteEntityType(@PathVariable("entityTypeId") String entityTypeId, @PathVariable("domainId") String domainId){
        boolean response = sharingRegistryService.deleteEntityType(domainId, entityTypeId);
        if(response){
            return new ResponseEntity<>(HttpStatus.OK);
        }
        else{
            throw new SharingRegistryException("Failed to delete the entity type");
        }
    }

    /**
     <p>API method to get an entity type</p>
     */
    @RequestMapping(value = "/id/{entityTypeId}/domain/{domainId}", method = RequestMethod.GET)
    public @ResponseBody EntityType getEntityType(@PathVariable("entityTypeId") String entityTypeId, @PathVariable("domainId") String domainId){
        EntityType entityType = sharingRegistryService.getEntityType(domainId, entityTypeId);
        if(entityType != null){
            return entityType;
        }
        else{
            throw new ResourceNotFoundException("Could not find the entity type with entityTypeId: " + entityTypeId + " in domainId: "+ domainId);
        }
    }

    /**
     <p>API method to get entity types in a domainId.</p>
     */
    @RequestMapping(value = "/domain/{domainId}/offset/{offset}/limit/{limit}", method = RequestMethod.GET)
    public @ResponseBody List<EntityType> getEntityTypes(@PathVariable("domainId") String domainId, @PathVariable("offset") int offset, @PathVariable("limit") int limit){
       List<EntityType> entityTypes = sharingRegistryService.getEntityTypes(domainId, offset, limit);
       return entityTypes;
    }
}
