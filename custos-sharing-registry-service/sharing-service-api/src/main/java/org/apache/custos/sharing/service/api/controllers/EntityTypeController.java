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
    public ResponseEntity<EntityType> createEntityType(@Valid @RequestBody EntityType entityType){
        EntityType createdEntityType = sharingRegistryService.createEntityType(entityType);
        return new ResponseEntity<>(createdEntityType, HttpStatus.CREATED);
    }

    /**
     <p>API method to update entity type</p>
     */
    @RequestMapping(method = RequestMethod.PUT)
    public ResponseEntity<String> updateEntityType(@Valid @RequestBody EntityType entityType){
        sharingRegistryService.updateEntityType(entityType);
        return new ResponseEntity<>(HttpStatus.OK);
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
        sharingRegistryService.deleteEntityType(domainId, entityTypeId);
        return new ResponseEntity<>(HttpStatus.OK);
    }

    /**
     <p>API method to get an entity type</p>
     */
    @RequestMapping(value = "/{entityTypeId}", method = RequestMethod.GET)
    public ResponseEntity<EntityType> getEntityType(@PathVariable("entityTypeId") String entityTypeId, @PathVariable("domainId") String domainId){
        EntityType entityType = sharingRegistryService.getEntityType(domainId, entityTypeId);
        return new ResponseEntity<>(entityType, HttpStatus.OK);
    }

    /**
     <p>API method to get entity types in a domainId.</p>
     */
    @RequestMapping(method = RequestMethod.GET)
    public @ResponseBody List<EntityType> getEntityTypes(@PathVariable("domainId") String domainId, @RequestParam(name = "offset", defaultValue = "0") int offset, @RequestParam(name = "limit", defaultValue = "-1") int limit){
       return sharingRegistryService.getEntityTypes(domainId, offset, limit);
    }
}
