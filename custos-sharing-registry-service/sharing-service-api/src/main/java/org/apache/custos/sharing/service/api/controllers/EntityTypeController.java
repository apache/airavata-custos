package org.apache.custos.sharing.service.api.controllers;

import org.apache.custos.sharing.service.core.models.EntityType;
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
@RequestMapping("/domains/{domainId}/entityTypes")
public class EntityTypeController {

    @Autowired
    SharingRegistryService sharingRegistryService;
    /**
     <p>API method to create a new entity type</p>
     */
    @RequestMapping(method = RequestMethod.POST)
    public EntityType createEntityType(@Valid @RequestBody EntityType entityType){
        EntityType createdEntityType = sharingRegistryService.createEntityType(entityType);
        if(createdEntityType != null){
            return createdEntityType;
        }
        else{
            throw new SharingRegistryException("Failed to create the entity type");
        }
    }

    /**
     <p>API method to update entity type</p>
     */
    @RequestMapping(method = RequestMethod.PUT)
    public ResponseEntity<String> updateEntityType(@Valid @RequestBody EntityType entityType){
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
    @RequestMapping(value = "/{entityTypeId}", method = RequestMethod.HEAD)
    public ResponseEntity<String> isEntityTypeExists(@PathVariable("entityTypeId") String entityTypeId, @PathVariable("domainId") String domainId){
        boolean response = sharingRegistryService.isEntityTypeExists(domainId, entityTypeId);
        HttpStatus status = HttpStatus.NOT_FOUND;
        if(response){
            status = HttpStatus.OK;
        }
        return new ResponseEntity<>(status);
    }

    /**
     <p>API method to delete entity type</p>
     */
    @RequestMapping(value = "/{entityTypeId}", method = RequestMethod.DELETE)
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
    @RequestMapping(value = "/{entityTypeId}", method = RequestMethod.GET)
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
    @RequestMapping(method = RequestMethod.GET)
    public @ResponseBody List<EntityType> getEntityTypes(@PathVariable("domainId") String domainId, @RequestParam("offset") int offset, @RequestParam("limit") int limit){
       List<EntityType> entityTypes = sharingRegistryService.getEntityTypes(domainId, offset, limit);
       return entityTypes;
    }
}
