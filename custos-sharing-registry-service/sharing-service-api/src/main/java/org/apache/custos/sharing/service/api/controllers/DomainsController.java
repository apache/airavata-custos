package org.apache.custos.sharing.service.api.controllers;

import org.apache.custos.sharing.service.core.models.Domain;
import org.apache.custos.sharing.service.core.exceptions.SharingRegistryException;
import org.apache.custos.sharing.service.core.service.SharingRegistryService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import javax.validation.Valid;
import java.util.List;

@RestController
@RequestMapping("/domains")
public class DomainsController {

    @Autowired
    SharingRegistryService sharingRegistryService;

    /**
     <p>API method to create a new domain</p>
     */
    @RequestMapping(method= RequestMethod.POST)
    public @ResponseBody Domain createDomain(@Valid @RequestBody Domain domain){
        Domain createdDomain = sharingRegistryService.createDomain(domain);
        if(createdDomain != null) return createdDomain;
        else throw new SharingRegistryException("Domain creation failed");
    }

    /**
     <p>API method to update a domain</p>
     */
    @RequestMapping(method = RequestMethod.PUT)
    public ResponseEntity<String> updateDomain(@Valid @RequestBody Domain domain){
        boolean response = sharingRegistryService.updateDomain(domain);
        if(response) return new ResponseEntity<>(HttpStatus.OK);
        else throw new SharingRegistryException("Domain update failed");
    }

    /**
     <p>API method to check Domain Exists</p>
     */
    @RequestMapping(value = "/{domainId}", method = RequestMethod.HEAD)
    public ResponseEntity<String> isDomainExists(@PathVariable("domainId") String domainId) {
        boolean response = sharingRegistryService.isDomainExists(domainId);
        HttpStatus status = HttpStatus.NOT_FOUND;
        if(response){
            status = HttpStatus.OK;
        }
        return new ResponseEntity<>(status);
    }

    /**
     <p>API method to retrieve a domain</p>
     */
    @RequestMapping(value = "/{domainId}",method = RequestMethod.GET)
    public ResponseEntity<Domain> getDomain(@PathVariable("domainId") String domainId){
        Domain domain = sharingRegistryService.getDomain(domainId);
        if(domain != null) return new ResponseEntity<>(domain, HttpStatus.OK);
        else return new ResponseEntity<>(HttpStatus.NOT_FOUND);
    }

    /**
     <p>API method to delete domain</p>
     */
    @RequestMapping(value = "/{domainId}", method = RequestMethod.DELETE)
    public ResponseEntity<String> deleteDomain(@PathVariable("domainId") String domainId){
        boolean response = sharingRegistryService.deleteDomain(domainId);
        if(response){
            return new ResponseEntity<>(HttpStatus.OK);
        }else{
            throw new SharingRegistryException("Could not delete the domain");
        }
    }

    /**
     <p>API method to get all domain.</p>
     */
    @RequestMapping(method = RequestMethod.GET)
    public @ResponseBody List<Domain> getDomains(@RequestParam(value = "offset") int offset, @RequestParam(value = "limit") int limit){
        return sharingRegistryService.getDomains(offset, limit);
    }

}
